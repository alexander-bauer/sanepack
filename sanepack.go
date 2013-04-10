package main

import (
	"encoding/json"
	"flag"
	"github.com/inhies/go-utils/log"
	"os"
)

var (
	Version = "0.1"

	l        *log.Logger
	loglevel log.LogLevel = log.WARNING // Normally only show warnings
)

const (
	defaultTemplates = "/etc/sanepack"
)

var (
	fFile   = flag.String("f", "sanepack.json", "sanepack file to read")
	fCreate = flag.Bool("c", false, "create a template sanepack file")

	fType = flag.String("t", "deb", "package type (such as \"deb\")")
	fTemp = flag.String("temp", defaultTemplates, "template location")

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

	// If we aren't creating a template, begin normal operation. Start
	// by determining the package type.
	var fw Frameworker
	switch *fType {
	case "deb":
		fw = DebianFrameworker{}
	default:
		// If the type of package requested is invalid, exit.
		l.Fatalf("Invalid package type: %q\n", *fType)
	}
	l.Debugf("Framework type: %q", *fType)

	// Now continue on to try to open the file.
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
	p := new(Package)
	err = json.NewDecoder(f).Decode(p)
	f.Close() // Close the file the moment we're done with it.
	if err != nil {
		// If the decode fails, report it. As before, use l.Fatalf().
		l.Fatalf("Could not decode project: %s", err)
	}
	l.Debug("Decode successful and file closed\n")

	// Now, move on to creating the framework with the previously
	// selected type.
	l.Debugf("Trying to create framework with type %q\n", *fType)
	err = fw.Framework(p)
	if err != nil {
		l.Fatalf("Could not create framework: %s", err)
	}
	l.Debug("Successfully created framework\n")

	l.Println(fw.Info())
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

	// MarshalIndent() a template Package, then write it to the
	// file. Note the call to templatePackage().
	b, err := json.MarshalIndent(templatePackage(), "", "\t")
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
	l.Debug("Wrote template successfully\n")
	return
}

// concat is a small utility used to create long strings from
// []strings. It places the given separator between every item in the
// given array and returns it.
func concat(sep string, items ...string) (concatenated string) {
	for i, item := range items {
		if i == 0 {
			// The first item in the list needs no separator before
			// it.
			concatenated = item
			continue
		}
		concatenated += sep + item
	}
	return
}
