package ddns_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xsteadfastx/carson/internal/ddns"
	"github.com/xsteadfastx/carson/internal/token"
)

type mockRefresh struct{}

func (m *mockRefresh) Refresh(ctx context.Context, _a0 *ddns.DDNS, r ddns.Record, remoteAddr string) {
}

func TestOnlyOneToken(t *testing.T) {
	assert := assert.New(t)

	tables := []struct {
		url string
	}{
		{
			"/",
		},
		{
			"/?token=foo&token=bar",
		},
	}

	for _, table := range tables {
		ddns := &ddns.DDNS{
			TokenSecret: "foobar",
			Counters:    ddns.NewCounters(),
		}

		req, err := http.NewRequest("GET", table.url, nil)
		assert.NoError(err)

		rr := httptest.NewRecorder()

		ddns.ServeHTTP(rr, req)

		assert.Equal(422, rr.Code)
		assert.Equal("none: only one token parameter allowed\n", rr.Body.String())
	}
}

func TestWrongToken(t *testing.T) {
	assert := assert.New(t)

	tables := []struct {
		url string
	}{
		{
			"/?token=foobar",
		},
	}

	for _, table := range tables {
		ddns := &ddns.DDNS{
			TokenSecret: "foobar",
			Tokenizer:   token.Token{},
			Counters:    ddns.NewCounters(),
		}

		req, err := http.NewRequest("GET", table.url, nil)
		assert.NoError(err)

		rr := httptest.NewRecorder()

		ddns.ServeHTTP(rr, req)

		assert.Equal(401, rr.Code)
		assert.Equal("none: could not parse token\n", rr.Body.String())
	}
}

func TestRightToken(t *testing.T) {
	assert := assert.New(t)

	tables := []struct {
		url string
	}{
		{
			"/?token=" +
				"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
				"eyJob3N0bmFtZSI6InpvbmsuZm9vLmJhci50bGQiLCJ0eXBlIjoiQSJ9." +
				"joHyH5FmEB3izP5hjU5Ye-MkRwJsF-Ve1vYNh_rMwRI",
		},
	}

	for _, table := range tables {
		ddns := &ddns.DDNS{
			TokenSecret: "foobar",
			Zones: []ddns.Zone{
				{
					Name:    "foo.bar.tld",
					Records: make(map[string]ddns.Record),
				},
			},
			Tokenizer: &token.Token{},
			Refresher: &mockRefresh{},
			Counters:  ddns.NewCounters(),
		}

		req, err := http.NewRequest("GET", table.url, nil)
		req.RemoteAddr = "127.0.0.1:9999"

		assert.NoError(err)

		rr := httptest.NewRecorder()

		ddns.ServeHTTP(rr, req)

		assert.Equal(200, rr.Code)
	}
}
