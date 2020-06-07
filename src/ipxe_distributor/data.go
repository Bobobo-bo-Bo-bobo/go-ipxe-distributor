package main

type ConfigGlobal struct {
    URL string
	Host string
	Port string
}

type ConfigDefault struct {
	IPXEPrepend  []string
	IPXEAppend   []string
	DefaultImage []string
}

type ConfigImages struct {
	Action []string
}

type ConfigNodes struct {
	MAC    string
	Group  string
	Serial string
	Image  string
}

type Configuration struct {
	Global  ConfigGlobal
	Default ConfigDefault
	Images  map[string]ConfigImages
	Nodes   map[string]ConfigNodes
}
