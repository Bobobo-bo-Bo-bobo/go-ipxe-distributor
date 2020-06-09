package main

// ConfigGlobal - global configuration
type ConfigGlobal struct {
	URL string
}

// ConfigDefault - iPXE defaults
type ConfigDefault struct {
	IPXEPrepend  []string
	IPXEAppend   []string
	DefaultImage []string
}

// ConfigImages - images
type ConfigImages struct {
	Action []string
}

// ConfigNodes - nodes
type ConfigNodes struct {
	MAC    string
	Group  string
	Serial string
	Image  string
}

// Configuration - configuration
type Configuration struct {
	Global        ConfigGlobal
	Default       ConfigDefault
	Images        map[string]ConfigImages
	Nodes         map[string]ConfigNodes
	MACNodeMap    map[string]string // Map normalized MAC addresses to node names
	SerialNodeMap map[string]string // Map serial numbers to node names
	GroupNodeMap  map[string]string // Map group names to node names
}
