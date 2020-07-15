package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"github.com/xsteadfastx/carson/internal/ddns"
	"github.com/xsteadfastx/carson/internal/token"
)

func NewDDNS(fp string) (*ddns.DDNS, error) {
	logger := log.WithFields(log.Fields{"config": fp})
	c := &ddns.DDNS{}

	logger.Debug("read config")

	raw, err := ioutil.ReadFile(fp)

	if err != nil {
		return &ddns.DDNS{}, err
	}

	_, err = toml.Decode(string(raw), c)
	if err != nil {
		return &ddns.DDNS{}, err
	}

	c.Tokenizer = token.Token{}
	c.Refresher = ddns.Refresher{}

	return c, nil
}
