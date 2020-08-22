package ddns

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/xsteadfastx/carson/internal/errs"
)

const zoneFile string = `$ORIGIN {{ .Zone.Name }}.
$TTL 300

@ IN SOA {{ .Nameserver }}. {{ .Admin | toDot }}. (
		{{ .Zone.Serial }}	; serial number
		86400	; refresh
		7200	; retry
		2419200	; expire
		3600	; min TTL
		)

		IN	NS	{{ .Nameserver }}.

{{ range $k, $v := .Zone.Records }}
{{ $k }}	IN	{{ $v.Type }}	{{ $v.Target }}
{{ end -}}
`

type now struct{}

func (n now) Now() string {
	return time.Now().Format("20060102")
}

type Refresher struct{}

// Refresh parses all zonefiles and adds them to the DDNS struct, creates Zonefiles.
func (ref Refresher) Refresh(ctx context.Context, ddns *DDNS, r Record, remoteAddr string) {
	ddns.Lock()
	defer ddns.Unlock()

	logger := log.WithFields(log.Fields{"uuid": GetUUID(ctx)})

	zones := []Zone{}

	logger.Info("parses zones...")

	for _, z := range ddns.Zones {
		zLogger := logger.WithFields(log.Fields{"zone": z.Name})
		nz, err := ddns.ParseZone(ctx, z.Name, now{})

		if err != nil {
			zLogger.Error(err)
			continue
		}

		zLogger.WithFields(log.Fields{"data": fmt.Sprintf("%+v", nz)}).Debug("adds new zone")

		zones = append(zones, nz)
	}

	logger.Info("stores zones...")

	ddns.Zones = zones

	logger.Info("add new record...")

	changed, err := ddns.AddRecord(ctx, r, remoteAddr)
	if err != nil {
		logger.Error(err)
		return
	}

	if changed || r == (Record{}) {
		logger.Info("writes zones...")

		for _, z := range ddns.Zones {
			zLogger := logger.WithFields(log.Fields{"zone": z.Name})
			err := ddns.CreateZone(ctx, z)

			if err != nil {
				zLogger.Error(err)
				continue
			}
		}
	}
}

// ParseZone will parse the zone file.
func (ddns *DDNS) ParseZone(ctx context.Context, name string, now Nower) (Zone, error) {
	logger := log.WithFields(log.Fields{"zone": name, "uuid": GetUUID(ctx)})

	p := path.Join(ddns.ZonesDir, fmt.Sprintf("%s.zone", name))

	_, err := os.Stat(p)

	if os.IsNotExist(err) {
		return Zone{Name: name, Serial: now.Now() + "01", Records: make(map[string]Record)}, nil
	}

	logger.Infof("read %s", p)
	raw, err := ioutil.ReadFile(p)

	if err != nil {
		return Zone{}, err
	}

	data := string(raw)

	// New Zone struct.
	zone := Zone{Name: name}

	// Handle serial number.
	reSerial := regexp.MustCompile(`\s(\d{10})\s+;\sserial number`)

	foundSerial := reSerial.FindAllStringSubmatch(data, 1)

	if len(foundSerial) < 1 || len(foundSerial[0]) < 2 {
		return Zone{}, errs.ErrNoSerialFound
	}

	// Raise serial if needed.
	oldSerial := foundSerial[0][1]
	newSerial, err := RaiseSerial(oldSerial, now)

	if err != nil {
		return Zone{}, err
	}

	zone.Serial = newSerial

	// Handle Records.
	reRecords := regexp.MustCompile(`(?U)([[:alnum:]]+)\s+IN\s+(A|CNAME)\s+(\d+\.\d+\.\d+\.\d+)`)
	rs := reRecords.FindAllStringSubmatch(data, -1)

	logger.WithFields(log.Fields{"records": rs}).Debug("found")

	zone.Records = make(map[string]Record)

	for _, r := range rs {
		zone.Records[r[1]] = Record{Target: r[3], Type: r[2]}
	}

	return zone, nil
}

// CreateZone will create a zone file.
func (ddns *DDNS) CreateZone(ctx context.Context, z Zone) error {
	zFile := path.Join(ddns.ZonesDir, fmt.Sprintf("%s.zone", z.Name))
	logger := log.WithFields(log.Fields{"zone_file": zFile, "uuid": GetUUID(ctx)})

	logger.Info("create")

	f, err := os.Create(zFile)

	if err != nil {
		return err
	}

	t, err := template.New("").Funcs(
		template.FuncMap{
			"toDot": func(s string) string {
				return strings.Replace(s, "@", ".", -1)
			},
		},
	).Parse(zoneFile)
	if err != nil {
		return err
	}

	wf := bufio.NewWriter(f)
	wt := tabwriter.NewWriter(wf, 4, 4, 4, '\t', 0)

	if err := t.Execute(wt, struct {
		Zone       Zone
		Nameserver string
		Admin      string
	}{Zone: z, Nameserver: ddns.Nameserver, Admin: ddns.Admin}); err != nil {
		return err
	}

	if err := wt.Flush(); err != nil {
		return err
	}

	if err := wf.Flush(); err != nil {
		return err
	}

	return nil
}

func (ddns *DDNS) AddRecord(ctx context.Context, r Record, remoteAddr string) (bool, error) { //nolint: funlen
	if r == (Record{}) {
		return false, nil
	}

	if r.Hostname == "" {
		return false, errs.ErrMissingRecordHostname
	}

	if r.Type == "" {
		return false, errs.ErrMissingRecordType
	}

	logger := log.WithFields(log.Fields{"uuid": GetUUID(ctx)})

	for _, z := range ddns.Zones {
		// r.Hostname is a FQDN.
		if strings.Contains(r.Hostname, z.Name) { //nolint: nestif
			zLogger := logger.WithFields(log.Fields{"zone": z.Name})

			hs := strings.Split(r.Hostname, "."+z.Name)

			if len(hs) != 2 { //nolint: gomnd
				ddns.Counters["errors"].With(
					prometheus.Labels{"error": errs.ErrCouldNotExtractHostname.Error()},
				).Inc()

				return false, errs.ErrCouldNotExtractHostname
			}

			h := hs[0]

			zLogger.WithFields(log.Fields{"hostname": h}).Info("extracted hostname")

			target := strings.Split(remoteAddr, ":")
			if len(target) != 2 { //nolint: gomnd
				ddns.Counters["errors"].With(
					prometheus.Labels{"error": errs.ErrCouldNotExtractTarget.Error()},
				).Inc()

				return false, errs.ErrCouldNotExtractTarget
			}

			t := target[0]

			rs, ok := z.Records[h]

			// If Record entry does not exists.
			if !ok {
				r.Target = t
				z.Records[h] = r

				return true, nil
			} else if ok {
				if rs.Target == t {
					zLogger.Info("same record, doing nothing")
					return false, nil
				}
				r.Target = t
				z.Records[h] = r
				ddns.Counters["update"].With(
					prometheus.Labels{"zone": z.Name, "hostname": r.Hostname, "type": r.Type},
				).Inc()
				return true, nil
			}
		}
	}

	ddns.Counters["errors"].With(
		prometheus.Labels{"error": errs.ErrNoZoneForHostname.Error()},
	).Inc()

	return false, errs.ErrNoZoneForHostname
}

// RaiseSerial checks if a serial needs to be risen or a complete new one.
func RaiseSerial(old string, now Nower) (string, error) {
	n := now.Now()

	// Check if the first two chars are the same like the one it would use now.
	if n == old[:8] {
		log.WithFields(log.Fields{"old": old[:8], "now": n}).Debug("needs a new serial")

		i, err := strconv.Atoi(old)
		if err != nil {
			return "", err
		}

		return strconv.Itoa(i + 1), nil
	}

	return n + "01", nil
}
