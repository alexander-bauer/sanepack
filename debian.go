package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"text/template"
)

type DebianFrameworker struct{}

const (
	DebianInfo = `To complete building the package, invoke:
    fakeroot dpkg-buildpackage`

	debianStandardsVersion = "3.9.3"
)

func (d DebianFrameworker) Info() string {
	return DebianInfo
}

func (d DebianFrameworker) Framework(p *Package) (err error) {
	// Begin by trying to create the most difficult file,
	// debian/control.
	l.Debug("Attempting to create debian/control\n")
	err = d.control(p.ProjectName, p.Description, "",
		p.Section, p.Priority,
		p.Homepage, p.Architecture, p.Maintainer, p.BuildDepends, p.Depends,
		p.Recommends, p.Suggests, p.Conflicts, p.Provides, p.Replaces)
	if err != nil {
		return
	}

	// Try to create the debian/<name>.manpages file.
	l.Debugf("Attempting to create debian/%s.manpages\n", p.ProjectName)
	err = d.manpages(p.ProjectName, p.ManPages)
	if err != nil {
		return
	}

	return
}

// control creates a debian/control file and populates it with the
// given fields. name, description, section, priority, architecture,
// maintainer, buildDepends, and depends are required.
func (d DebianFrameworker) control(name, description, longDescription, section, priority, homepage, architecture string, maintainer Person, buildDepends, depends, recommends, suggests, conflicts, provides, replaces []string) (err error) {

	// First, check that all required fields are given.
	if len(name) == 0 || len(description) == 0 || len(section) == 0 ||
		len(priority) == 0 || len(architecture) == 0 ||
		len(maintainer.Name) == 0 || len(maintainer.Email) == 0 ||
		buildDepends == nil || depends == nil {
		return errors.New("debian: not all required fields are given")
	}

	// Next, create a debianControlFile object.
	control := &debianControlFile{
		Name:             name,
		Section:          section,
		Priority:         priority,
		Architecture:     architecture,
		StandardsVersion: debianStandardsVersion,
		Homepage:         homepage,
		Description:      description,
		LongDescription:  longDescription,
		Maintainer:       maintainer,
		BuildDepends:     concat(", ", buildDepends...),
		Depends:          concat(", ", depends...),
		Recommends:       concat(", ", recommends...),
		Suggests:         concat(", ", suggests...),
		Conflicts:        concat(", ", conflicts...),
		Provides:         concat(", ", provides...),
		Replaces:         concat(", ", replaces...),
		Include:          make(map[string]bool, 6),
	}

	if len(homepage) != 0 {
		control.Include["Homepage"] = true
	}
	if len(recommends) != 0 {
		control.Include["Recommends"] = true
	}
	if len(suggests) != 0 {
		control.Include["Suggests"] = true
	}
	if len(conflicts) != 0 {
		control.Include["Conflicts"] = true
	}
	if len(provides) != 0 {
		control.Include["Provides"] = true
	}
	if len(replaces) != 0 {
		control.Include["Replaces"] = true
	}

	// Attempt to open debian/control.
	f, err := os.Create("debian/control")
	if err != nil {
		return
	}
	defer f.Close()

	t, err := template.ParseFiles(path.Join(*fTemp, "debian/control.template"))
	if err != nil {
		return
	}

	err = t.Execute(f, control)
	f.Close()
	return
}

// manpages creates a "debian/<name>.manpages" file containing every
// path in the manpages slice, one per line.
func (d DebianFrameworker) manpages(name string, manpages []string) (err error) {
	// Begin by trying to open the debian/<name>.manpages file.
	f, err := os.Create("debian/" + name + ".manpages")
	if err != nil {
		return
	}
	defer f.Close()

	// If the file is opened properly, use fmt.Fprintln() to write the
	// manpages in, one per line.
	for _, page := range manpages {
		fmt.Fprintln(f, page)
	}

	f.Close()
	return
}

type debianControlFile struct {
	Name, Section, Priority, Architecture, StandardsVersion string
	Homepage, Description, LongDescription                  string
	Maintainer                                              Person
	BuildDepends, Depends, Recommends                       string
	Suggests, Conflicts, Provides, Replaces                 string
	Include                                                 map[string]bool
}
