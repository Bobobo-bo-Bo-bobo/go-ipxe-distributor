package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	// "strconv"
)

func readConfigurationFile(f string) ([]byte, error) {
	content, err := ioutil.ReadFile(f)
	return content, err
}

func parseYAML(y []byte) (*Configuration, error) {
	var cfg Configuration

	// raw_map := make(map[interface{}]interface{})
	raw_map := make(map[string]interface{})
	err := yaml.Unmarshal(y, &raw_map)
	if err != nil {
		return nil, err
	}

	log.Println("%+v\n", raw_map)

	// Initialise configuration
	cfg.Global.Host = DEFAULT_HOST
	cfg.Global.Port = DEFAULT_PORT

	cfg.Default.IPXEPrepend = make([]string, 0)
	cfg.Default.IPXEAppend = make([]string, 0)
	cfg.Default.DefaultImage = make([]string, 0)

	cfg.Images = make(map[string]ConfigImages)
	cfg.Nodes = make(map[string]ConfigNodes)

	// parse raw map into configuration ... which is kind of ugly
	for key := range raw_map {
		if key == "global" {
			_global := raw_map[key.(string)].(map[string]string)
			for g_key := range _global {
				if g_key == "host" {
				} else if g_key == "port" {
					/*                    switch _global[g_key].(type) {
					                      case string:
					                          cfg.Global.Port = _global[g_key].(string)
					                          if err != nil {
					                              return nil, err
					                          }
					                      case int:
					                          cfg.Global.Port = strconv.Itoa(_global[g_key].(int))
					                      }*/
				} else {
					log.Printf("Warning: Skipping unsupported key for global dictionary: %s\n", g_key)
				}
			}
		} else if key == "default" {
		} else if key == "images" {
		} else if key == "nodes" {
		} else {
			log.Printf("Warning: Skipping unsupported key: %s\n", key)
		}
	}
	return &cfg, nil
}
