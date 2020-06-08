package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

func readConfigurationFile(f string) ([]byte, error) {
	content, err := ioutil.ReadFile(f)
	return content, err
}

func parseYAML(y []byte) (*Configuration, error) {
	var cfg Configuration

	// set defaults and initialise data structures
	cfg.Global.URL = defaultURL
	cfg.Default.IPXEPrepend = make([]string, 0)
	cfg.Default.IPXEAppend = make([]string, 0)
	cfg.Default.DefaultImage = make([]string, 0)
	cfg.Images = make(map[string]ConfigImages)
	cfg.Nodes = make(map[string]ConfigNodes)
	cfg.MACNodeMap = make(map[string]string)
	cfg.SerialNodeMap = make(map[string]string)
	cfg.GroupNodeMap = make(map[string][]string)

	// parse YAML with dynamic keys
	rawMap := make(map[string]interface{})
	err := yaml.Unmarshal(y, &rawMap)
	if err != nil {
		return nil, err
	}

	for key, sub := range rawMap {
		switch key {
		case "default":
			for dkey, dvalue := range sub.(map[interface{}]interface{}) {
				if getInterfaceType(dvalue) != TypeOther {
					return nil, fmt.Errorf("Invalid type for value of default")
				}
				if getInterfaceType(dkey) != TypeString {
					return nil, fmt.Errorf("Invalid key type for default")
				}

				switch dkey {
				case "default_image":
					if getInterfaceType(dvalue) != TypeOther {
						return nil, fmt.Errorf("Invalid type for default_image")
					}

					for _, k := range dvalue.([]interface{}) {
						if getInterfaceType(k) != TypeString {
							return nil, fmt.Errorf("Invalid type for default_image")
						}
						cfg.Default.DefaultImage = append(cfg.Default.DefaultImage, k.(string))
					}
				case "ipxe_append":
					if getInterfaceType(dvalue) != TypeOther {
						return nil, fmt.Errorf("Invalid type for ipxe_append")
					}
					for _, k := range dvalue.([]interface{}) {
						if getInterfaceType(k) != TypeString {
							return nil, fmt.Errorf("Invalid type for ipxe_append")
						}
						cfg.Default.IPXEAppend = append(cfg.Default.IPXEAppend, k.(string))
					}
				case "ipxe_prepend":
					if getInterfaceType(dvalue) != TypeOther {
						return nil, fmt.Errorf("Invalid type for ipxe_prepend")
					}
					for _, k := range dvalue.([]interface{}) {
						if getInterfaceType(k) != TypeString {
							return nil, fmt.Errorf("Invalid type for ipxe_prepend")
						}
						cfg.Default.IPXEPrepend = append(cfg.Default.IPXEPrepend, k.(string))
					}
				default:
					log.WithFields(log.Fields{
						"key": dkey.(string),
					}).Warning("Ignoring unsuppored configuration key for default")
				}
			}

		case "global":
			for gkey, gvalue := range sub.(map[interface{}]interface{}) {
				if getInterfaceType(gvalue) != TypeString {
					return nil, fmt.Errorf("Invalid type for global")
				}

				switch gkey {
				case "url":
					cfg.Global.URL = gvalue.(string)
				default:
					log.WithFields(log.Fields{
						"key": gkey.(string),
					}).Warning("Ignoring unsuppored configuration key for global")
				}
			}
		case "images":
			for imgname, ivalue := range sub.(map[interface{}]interface{}) {
				if getInterfaceType(imgname) != TypeString {
					return nil, fmt.Errorf("Key for image name is not a string")
				}

				for a, aval := range ivalue.(map[interface{}]interface{}) {
					if getInterfaceType(a) != TypeString {
						return nil, fmt.Errorf("Invalid type for images value")
					}
					switch a {
					case "action":
						if getInterfaceType(aval) != TypeOther {
							return nil, fmt.Errorf("Invalid type for action of image %s", imgname)
						}
						var imgact []string
						for _, k := range aval.([]interface{}) {
							if getInterfaceType(k) != TypeString {
								return nil, fmt.Errorf("Invalid type for action value of image %s", imgname)
							}
							imgact = append(imgact, k.(string))
						}
						cfg.Images[imgname.(string)] = ConfigImages{
							Action: imgact,
						}

					default:
						log.WithFields(log.Fields{
							"key": a.(string),
						}).Warning("Ignoring unsuppored configuration key for images")
					}
				}
			}
		case "nodes":
			for nodename, nvalue := range sub.(map[interface{}]interface{}) {
				if getInterfaceType(nodename) != TypeString {
					return nil, fmt.Errorf("Key for nodes is not a string")
				}
				if getInterfaceType(nvalue) != TypeOther {
					return nil, fmt.Errorf("Invalid type of value for node %s", nodename)
				}

				var ncfg ConfigNodes

				for key, value := range nvalue.(map[interface{}]interface{}) {
					if getInterfaceType(key) != TypeString {
						return nil, fmt.Errorf("Invalid key type for node %s", nodename)
					}
					if getInterfaceType(value) != TypeString {
						return nil, fmt.Errorf("Invalid value type for node %s", nodename)
					}

					switch key.(string) {
					case "image":
						ncfg.Image = value.(string)
					case "serial":
						ncfg.Serial = value.(string)
					case "group":
						ncfg.Group = value.(string)
					case "mac":
						ncfg.MAC = value.(string)
					default:
						log.WithFields(log.Fields{
							"key":  key.(string),
							"node": nodename.(string),
						}).Warning("Ignoring unsuppored configuration key for node")
					}
				}
				cfg.Nodes[nodename.(string)] = ncfg
			}
		}
	}

	for name, ncfg := range cfg.Nodes {
		// Map MACs to node name
		if ncfg.MAC != "" {
			_normalized := strings.Replace(strings.Replace(strings.ToLower(ncfg.MAC), ":", "", -1), "-", "", -1)
			cfg.MACNodeMap[_normalized] = name
		}

		// Map serial numbers to node name
		if ncfg.Serial != "" {
			cfg.SerialNodeMap[ncfg.Serial] = name
		}

		// Map group name to node name
		if ncfg.Group != "" {
			cfg.GroupNodeMap[ncfg.Group] = append(cfg.GroupNodeMap[ncfg.Group], name)
		}
	}

	return &cfg, nil
}
