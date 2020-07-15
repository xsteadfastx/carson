package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/xsteadfastx/carson/internal/config"
)

func Run(version string) { //nolint: funlen
	carsonVersion := flag.Bool("version", false, "version")

	flag.Parse()

	if *carsonVersion {
		fmt.Print(version)
		os.Exit(0)
	}

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runConfig := runCmd.String("config", "", "config file")

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createConfig := createCmd.String("config", "", "config file")
	createRecordType := createCmd.String("type", "", "DNS record type")
	createFQDN := createCmd.String("fqdn", "", "FQDN hostname")

	if len(os.Args) < 2 { //nolint:gomnd
		log.Fatal("expected 'run' or 'create' subcommands")
	}

	switch os.Args[1] {
	case "run":
		if err := runCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}

		if *runConfig == "" {
			fmt.Print("needs this parameters:\n\n")
			runCmd.PrintDefaults()
			fmt.Print("\n")
			os.Exit(1)
		}

		ddns, err := config.NewDDNS(*runConfig)
		if err != nil {
			log.Fatal(err)
		}

		ddns.Run()

	case "create":
		if err := createCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}

		if *createConfig == "" || *createRecordType == "" || *createFQDN == "" {
			fmt.Print("needs this parameters:\n\n")
			createCmd.PrintDefaults()
			fmt.Print("\n")
			os.Exit(1)
		}

		ddns, err := config.NewDDNS(*createConfig)
		if err != nil {
			log.Fatal(err)
		}

		token, err := ddns.Tokenizer.Create(ddns.TokenSecret, *createFQDN, *createRecordType)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s://%s:%s?token=%s", ddns.Scheme, ddns.Nameserver, ddns.Port, token)
	}
}
