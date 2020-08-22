package ddns_test

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/xsteadfastx/carson/internal/ddns"
	"github.com/xsteadfastx/carson/internal/ddns/mocks"
)

func TestRaiseSerial(t *testing.T) {
	assert := assert.New(t)
	table := []struct {
		old      string
		now      string
		expected string
	}{
		{
			"2019052603",
			"20200526",
			"2020052601",
		},
		{
			"2020052606",
			"20200526",
			"2020052607",
		},
	}

	for _, t := range table {
		mockNower := &mocks.Nower{}
		mockNower.On("Now").Return(t.now)
		new, err := ddns.RaiseSerial(t.old, mockNower)

		assert.NoError(err)
		assert.Equal(t.expected, new)
	}
}

func TestParseZone(t *testing.T) {
	assert := assert.New(t)

	mockNower := &mocks.Nower{}
	mockNower.On("Now").Return("1111111111")

	records := make(map[string]ddns.Record)

	records["bla"] = ddns.Record{
		Target: "127.0.0.1",
		Type:   "A",
	}

	records["zonk"] = ddns.Record{
		Target: "127.0.1.1",
		Type:   "A",
	}

	expected := ddns.Zone{
		Name:    "foo.bar.tld",
		Records: records,
	}

	ddns := &ddns.DDNS{ZonesDir: "testdata/zone_files"}
	z, err := ddns.ParseZone(context.Background(), "foo.bar.tld", mockNower)

	assert.NoError(err)

	assert.Equal(expected.Name, z.Name)
	assert.Equal(expected.Records, z.Records)
}

func TestCreateZone(t *testing.T) {
	assert := assert.New(t)

	dir, err := ioutil.TempDir("/tmp", "zones")
	assert.NoError(err)

	defer func() {
		log.Printf("remove %s", dir)
		os.RemoveAll(dir)
	}()

	d := ddns.DDNS{
		Admin:      "marv@xsfx.dev",
		Nameserver: "ns.xsfx.dev",
		ZonesDir:   dir,
	}
	records := make(map[string]ddns.Record)
	records["bla"] = ddns.Record{Type: "A", Target: "127.0.0.1"}
	records["blubb"] = ddns.Record{Type: "A", Target: "127.0.1.1"}
	z := ddns.Zone{
		Name:    "foo.bar.tld",
		Serial:  "0000000000",
		Records: records,
	}
	err = d.CreateZone(context.Background(), z)

	assert.NoError(err)

	actual, err := ioutil.ReadFile(path.Join(dir, "foo.bar.tld.zone"))
	assert.NoError(err)

	expected, err := ioutil.ReadFile(path.Join("testdata/expected", "foo.bar.tld.zone"))
	assert.NoError(err)

	assert.Equal(expected, actual)
}

//nolint: funlen
func TestAddRecord(t *testing.T) {
	assert := assert.New(t)

	tables := []struct {
		claim      ddns.Record
		remoteAddr string
		err        string
		changed    bool
		records    map[string]ddns.Record
		expected   map[string]ddns.Record
	}{
		{
			ddns.Record{},
			"",
			"missing hostname in token",
			false,
			map[string]ddns.Record{},
			map[string]ddns.Record{},
		},
		{
			ddns.Record{Hostname: "zonk.foo.bar.tld"},
			"",
			"missing record type in token",
			false,
			map[string]ddns.Record{},
			map[string]ddns.Record{},
		},
		{
			ddns.Record{Hostname: "zonk.foo.bar.tld", Type: "A"},
			"127.0.0.1:8888",
			"",
			true,
			map[string]ddns.Record{},
			map[string]ddns.Record{"zonk": {Hostname: "zonk.foo.bar.tld", Type: "A", Target: "127.0.0.1"}},
		},
		{
			ddns.Record{Hostname: "zonk.foo.bar.tld", Type: "A"},
			"127.0.0.1:8888",
			"",
			false,
			map[string]ddns.Record{"zonk": {Hostname: "zonk.foo.bar.tld", Type: "A", Target: "127.0.0.1"}},
			map[string]ddns.Record{"zonk": {Hostname: "zonk.foo.bar.tld", Type: "A", Target: "127.0.0.1"}},
		},
		{
			ddns.Record{Hostname: "zonk.foo.bar.tld", Type: "A"},
			"127.0.1.1:8888",
			"",
			true,
			map[string]ddns.Record{"zonk": {Hostname: "zonk.foo.bar.tld", Type: "A", Target: "127.0.0.1"}},
			map[string]ddns.Record{"zonk": {Hostname: "zonk.foo.bar.tld", Type: "A", Target: "127.0.1.1"}},
		},
		{
			ddns.Record{},
			"",
			"",
			false,
			map[string]ddns.Record{},
			map[string]ddns.Record{},
		},
	}

	for _, table := range tables {
		ddns := ddns.DDNS{
			Zones: []ddns.Zone{
				{
					Name:    "foo.bar.tld",
					Records: table.records,
				},
			},
			Counters: ddns.NewCounters(),
		}

		changed, err := ddns.AddRecord(context.Background(), table.claim, table.remoteAddr)

		assert.Equal(table.changed, changed)

		if err != nil {
			assert.EqualError(err, table.err)
		}

		assert.Equal(table.expected, ddns.Zones[0].Records)
	}
}

func TestRefresh(t *testing.T) { //nolint:funlen
	assert := assert.New(t)

	table := []struct {
		record     ddns.Record
		remoteAddr string
		oldLine    string
		newLine    string
		newSerial  string
	}{
		{
			record: ddns.Record{
				Hostname: "bla.foo.bar.tld",
				Type:     "A",
			},
			remoteAddr: "127.0.1.1:8000",
			oldLine:    "bla\t\tIN\t\tA\t\t127.0.0.1",
			newLine:    "bla\t\tIN\t\tA\t\t127.0.1.1",
			newSerial:  "",
		},
		{
			record: ddns.Record{
				Hostname: "bla.foo.bar.tld",
				Type:     "A",
			},
			remoteAddr: "127.0.0.1:8000",
			oldLine:    "bla\t\tIN\t\tA\t\t127.0.0.1",
			newLine:    "bla\t\tIN\t\tA\t\t127.0.0.1",
			newSerial:  "0000000000",
		},
	}

	for _, t := range table {
		dir, err := ioutil.TempDir("/tmp", "zones")
		assert.NoError(err)

		input, err := ioutil.ReadFile("testdata/expected/foo.bar.tld.zone")
		assert.NoError(err)

		err = ioutil.WriteFile(path.Join(dir, "foo.bar.tld.zone"), input, 0600)
		assert.NoError(err)

		defer func() {
			log.Printf("remove %s", dir)
			os.RemoveAll(dir)
		}()

		d := &ddns.DDNS{
			Admin:      "marv@xsfx.dev",
			Nameserver: "ns.xsfx.dev",
			ZonesDir:   dir,
			Zones: []ddns.Zone{
				{
					Name: "foo.bar.tld",
				},
			},
			Refresher: ddns.Refresher{},
			Counters:  ddns.NewCounters(),
		}

		zf, err := ioutil.ReadFile(path.Join(dir, "foo.bar.tld.zone"))
		assert.NoError(err)

		// Check if line is in file.
		assert.Equal(true, strings.Contains(string(zf), t.oldLine))

		d.Refresher.Refresh(context.Background(), d, t.record, t.remoteAddr)

		nzf, err := ioutil.ReadFile(path.Join(dir, "foo.bar.tld.zone"))
		assert.NoError(err)

		// Check if changed line is in file.
		assert.Equal(true, strings.Contains(string(nzf), t.newLine))

		// Check that old line is not in file.
		if t.newLine != t.oldLine {
			assert.Equal(false, strings.Contains(string(nzf), t.oldLine))
		}

		// Check serial
		if t.newSerial != "" {
			assert.Equal(true, strings.Contains(string(nzf), t.newSerial))
		} else {
			assert.Equal(false, strings.Contains(string(nzf), "0000000000"))
		}
	}
}
