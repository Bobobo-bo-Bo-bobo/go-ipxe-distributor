package main

const name = "go-ipxe-distributor"
const version = "1.0.1-20200703"

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

const macPath = "/mac/{mac}"
const serialPath = "/serial/{serial}"
const groupPath = "/group/{group}"
const defaultPath = "/default"

const versionText = `%s version %s
Copyright (C) 2020 by Andreas Maus <maus@ypbind.de>
This program comes with ABSOLUTELY NO WARRANTY.

%s is distributed under the Terms of the GNU General
Public License Version 3. (http://www.gnu.org/copyleft/gpl.html)

Build with go version: %s

`

const helpText = `Usage: %s [--config=<file>] [--help] [--test] [--version]
    --config=<file>     Read configuration from <file>
                        Default: %s

    --help              Show help text

    --test              Test configuration file for syntax errors

    --version           Show version information

`
