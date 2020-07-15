package ddns

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func NewCounters() map[string]*prometheus.CounterVec {
	cs := make(map[string]*prometheus.CounterVec)

	cs["requests"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "carson_requests_total",
			Help: "The total number of carson requests",
		},
		[]string{},
	)

	cs["errors"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "carson_errors",
			Help: "The number of carson errors",
		},
		[]string{"error"},
	)

	cs["update"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "carson_updates",
			Help: "The number of record updates",
		},
		[]string{"zone", "hostname", "type"},
	)

	return cs
}

func (ddns *DDNS) RegisterCounters() {
	for n, c := range ddns.Counters {
		log.WithFields(log.Fields{"counter": n}).Debug("register counter")
		prometheus.MustRegister(c)
	}
}
