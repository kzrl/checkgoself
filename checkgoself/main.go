package main

import (
	"flag"
	"fmt"
	"github.com/kzrl/checkgoself"
)

const version = "0.0.1"

func main() {

	var configFile = flag.String("config", "config.json", "Path to config.json")
	var helpFlag = flag.Bool("help", false, "Show usage")
	var versionFlag = flag.Bool("version", false, "Show version")
	var sendEmails = flag.Bool("emails", true, "Send email alerts")

	flag.Parse()

	if *helpFlag || *versionFlag {
		help()
		return
	}

	checkgoself.ConfigFile = configFile
	checkgoself.SendEmails = sendEmails

	config := checkgoself.ParseConfig()
	checkgoself.Check(config.Metrics)
}

func help() {
	fmt.Printf("checkgoself v%s\n", version)
	flag.PrintDefaults()
}
