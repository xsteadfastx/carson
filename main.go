package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/xsteadfastx/carson/cmd"
)

var version string = "development"

func main() {
	log.SetLevel(log.DebugLevel)
	cmd.Run(version)
}
