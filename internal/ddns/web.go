package ddns

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// root is the root handler.
func (ddns *DDNS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ddns.Counters["requests"].With(prometheus.Labels{}).Inc()

	reqID := GetUUID(r.Context())
	logger := log.WithFields(
		log.Fields{
			"remote-addr": r.RemoteAddr,
			"url":         r.URL,
			"uuid":        reqID,
		},
	)

	tokens, ok := r.URL.Query()["token"]

	if !ok || len(tokens) != 1 {
		logger.Error("only one token parameter allowed")
		ddns.Counters["errors"].With(prometheus.Labels{"error": "only one token parameter allowed"}).Inc()
		http.Error(w, fmt.Sprintf("%s: only one token parameter allowed", reqID), http.StatusUnprocessableEntity)

		return
	}

	p, err := ddns.Tokenizer.Parse(ddns.TokenSecret, tokens[0])
	if err != nil {
		logger.Error("could not parse token")
		ddns.Counters["errors"].With(prometheus.Labels{"error": "could not parse token"}).Inc()
		http.Error(w, fmt.Sprintf("%s: could not parse token", reqID), http.StatusUnauthorized)

		return
	}

	logger.WithFields(log.Fields{"parsed-token": p}).Info("parsed token")

	go ddns.Refresher.Refresh(r.Context(), ddns, p, r.RemoteAddr)

	fmt.Fprintf(w, "%s: triggered zone refresh", reqID)
}

// middlewareLogger logs the requests.
func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ReqID, uuid.New().String())
		reqID := GetUUID(ctx)

		log.WithFields(log.Fields{
			"method":      r.Method,
			"url":         r.URL,
			"proto":       r.Proto,
			"user-agent":  r.Header["User-Agent"],
			"remote-addr": r.RemoteAddr,
			"uuid":        reqID,
		}).Info("got request")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Run runs the webserver.
func (ddns *DDNS) Run() {
	ddns.Counters = NewCounters()
	ddns.RegisterCounters()

	http.Handle("/", middlewareLogger(http.Handler(ddns)))
	http.Handle("/metrics", promhttp.Handler())

	log.Infof("listen on %s", ddns.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", ddns.Port), nil))
}
