package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var config Configuration

func main() {
	var configfile = defaultConfigFile

	var logFmt = new(log.TextFormatter)
	logFmt.FullTimestamp = true
	logFmt.TimestampFormat = time.RFC3339
	log.SetFormatter(logFmt)

	if len(os.Args) >= 2 {
		configfile = os.Args[1]
		if len(os.Args) >= 3 {
			log.Warning("Ignoring additional command line parameters")
		}
	}

	raw, err := readConfigurationFile(configfile)
	if err != nil {
		log.WithFields(log.Fields{
			"config_file": configfile,
			"error":       err.Error(),
		}).Fatal("Can't read configuration file")
	}

	config, err = parseYAML(raw)
	if err != nil {
		log.WithFields(log.Fields{
			"config_file": configfile,
			"error":       err.Error(),
		}).Fatal("Can't parse provided configuration file")
	}

	// log.Printf("%+v\n", rawParsed)
	handleHTTP(config)
}
