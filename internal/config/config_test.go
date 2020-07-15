package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xsteadfastx/carson/internal/config"
	"github.com/xsteadfastx/carson/internal/ddns"
	"github.com/xsteadfastx/carson/internal/token"
)

func TestParseConfig(t *testing.T) {
	assert := assert.New(t)

	expected := &ddns.DDNS{
		Admin:       "marv@xsfx.dev",
		Nameserver:  "ns.foo.bar.tld",
		Port:        "8000",
		Scheme:      "http",
		TokenSecret: "foobar",
		Tokenizer:   token.Token{},
		ZonesDir:    "./tmp",
		Zones: []ddns.Zone{
			{
				Name:   "foo.bar.tld",
				Serial: "",
			},
		},
		Refresher: ddns.Refresher{},
	}

	c, err := config.NewDDNS("testdata/config.toml")

	assert.NoError(err)
	assert.Equal(expected, c)
}
