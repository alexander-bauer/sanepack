package main

import (
	"flag"
	"github.com/inhies/go-utils/log"
	"os"
)

var (
	Version = "0.0"

	l        *log.Logger
	loglevel log.LogLevel = log.WARNING // Normally only show warnings
)

var (
	fFile = flag.String("f", "sanepack.json", "package data file to read")

	fQuiet = flag.Bool("q", false, "disable logging") // not implemented
	fVerb  = flag.Bool("v", false, "enable verbose log output")
	fDebug = flag.Bool("debug", false, "enable debugging log output")
)

func main() {
	flag.Parse()

	// Create the logger as appropriate.
	if *fDebug { // If the debug flag is set:
		loglevel = log.DEBUG
	} else if *fVerb { // If the verbose flag is set:
		loglevel = log.INFO
	}
	l, _ = log.NewLevel(loglevel, true, os.Stdout, "", 0)

	l.Infof("Starting sanepack version %s\n", Version)
}
