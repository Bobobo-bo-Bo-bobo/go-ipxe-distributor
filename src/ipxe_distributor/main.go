package main

import (
	"flag"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var config Configuration

func main() {
	var help = flag.Bool("help", false, "Show help text")
	var _version = flag.Bool("version", false, "Show version information")
	var _test = flag.Bool("test", false, "Only test configuration file for syntax errors")
	var configfile = flag.String("config", defaultConfigFile, "Configuration file")
	var logFmt = new(log.TextFormatter)

	flag.Usage = showUsage
	flag.Parse()

	logFmt.FullTimestamp = true
	logFmt.TimestampFormat = time.RFC3339
	log.SetFormatter(logFmt)

	if *help {
		showUsage()
		os.Exit(0)
	}

	if *_version {
		showVersion()
		os.Exit(0)
	}

	raw, err := readConfigurationFile(*configfile)
	if err != nil {
		log.WithFields(log.Fields{
			"config_file": *configfile,
			"error":       err.Error(),
		}).Fatal("Can't read configuration file")
	}

	config, err = parseYAML(raw)
	if err != nil {
		log.WithFields(log.Fields{
			"config_file": *configfile,
			"error":       err.Error(),
		}).Fatal("Can't parse provided configuration file")
	}

	if *_test {
		log.WithFields(log.Fields{
			"config_file": *configfile,
		}).Info("Configuration file contains no syntax errors")
		os.Exit(0)
	}

	// log.Printf("%+v\n", rawParsed)
	handleHTTP(config)
}
