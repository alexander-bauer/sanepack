package main

import (
	"encoding/json"
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
	fFile   = flag.String("f", "sanepack.json", "sanepack file to read")
	fCreate = flag.Bool("c", false, "create a template sanepack file")

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

	if *fCreate {
		// If the -c flag is set, try to create a template file as
		// specified by -f.
		l.Debugf("Trying to write template file %q\n", *fFile)
		err := CreateSanepack(*fFile)
		if err != nil {
			l.Fatalf("Failed to write template file: %s", err)
		}
		// If the creation and writing was successful, report it and
		// return.
		l.Infof("Wrote template file %q successfully\n", *fFile)
		return
	}

	// If we aren't creating a template,, begin normal
	// operation. Start by trying to open the file for reading.
	l.Debugf("Trying to open file: %q\n", *fFile)
	f, err := os.Open(*fFile)
	if err != nil {
		// If the read fails, report it and exit. l.Fatalf() is
		// equivalent to l.Emergf() followed by a call to os.Exit(1).
		l.Fatalf("Failed to open %q: %s", *fFile, err)
	}
	defer f.Close()
	l.Debug("File opened successfully\n")

	// Now that the file has been opened and can be read from, try to
	// decode it into a Package.
	var p *Package
	err = json.NewDecoder(f).Decode(p)
	f.Close() // Close the file the moment we're done with it.
	if err != nil {
		// If the decode fails, report it. As before, use l.Fatalf().
		l.Fatalf("Could not decode project: %s", err)
	}
	l.Debug("Decode successful and file closed\n")
}

// Create a template sanepack file of the given filename.
func CreateSanepack(filename string) (err error) {
	// Create the file if it doesn't exist, or truncate and open it
	// for O_RDWR and file mode as inherited from parent.
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()
	l.Debugf("File %q opened for writing\n", filename)

	// MarshalIndent() a new Package object, then write it to the file.
	b, err := json.MarshalIndent(new(Package), "", "\t")
	if err != nil {
		l.Debug("JSON Marshalling failed\n")
		return
	}
	n, err := f.Write(b)
	if err != nil {
		// If the write fails, of course, return.
		l.Debug("File writing failed\n")
		return
	}
	err = f.Truncate(int64(n)) // Truncate the file.
	if err != nil {
		// If the truncation fails, (which would be a very strange
		// error), return.
		l.Debug("File truncation failed\n")
	}
	f.Close()
	l.Debug("Wrote template successfully\n")
	return
}
