package main

const name = "go-ipxe-distributor"
const version = "1.0.0-20200608"

const defaultConfigFile = "/etc/ipxe-distributor/config.yaml"
const defaultURL = "http://localhost:8080"
const (
	// TypeNil - interface{} is nil
	TypeNil int = iota
	// TypeBool - interface{} is bool
	TypeBool
	// TypeString - interface{} is string
	TypeString
	// TypeInt - interface{} is int
	TypeInt
	// TypeByte - interface{} is byte
	TypeByte
	// TypeFloat - interface{} is float
	TypeFloat
	// TypeOther - anything else
	TypeOther
)

const macPath = "/mac"
const serialPath = "/serial"
const groupPath = "/group"
const defaultPath = "/default"
