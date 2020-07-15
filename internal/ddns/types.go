package ddns

import (
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/prometheus/client_golang/prometheus"
)

type DDNS struct {
	Zones       []Zone `toml:"zones"`
	ZonesDir    string `toml:"zone_dir"`
	TokenSecret string `toml:"secret"`
	Nameserver  string `toml:"nameserver"`
	Admin       string `toml:"admin"`
	Scheme      string `toml:"scheme"`
	Port        string `toml:"port"`
	Tokenizer   Tokenizer
	Refresher   ZoneRefresher
	Counters    map[string]*prometheus.CounterVec
	sync.Mutex
}

type Zone struct {
	Name    string `toml:"name"`
	Records map[string]Record
	Serial  string
}

type Record struct {
	Hostname string `json:"hostname"`
	Type     string `json:"type"`
	Target   string
	jwt.StandardClaims
}

type UUID string
