package main

import (
	"log"
	"os"
)

func main() {
	var configfile string = DEFAULT_CONFIG_FILE

	if len(os.Args) >= 2 {
		configfile = os.Args[1]
		if len(os.Args) >= 3 {
			log.Println("Ignoring additional command line parameters")
		}
	}

	raw, err := readConfigurationFile(configfile)
	if err != nil {
		log.Fatal(err)
	}

	raw_parsed, err := parseYAML(raw)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(raw_parsed)
}
