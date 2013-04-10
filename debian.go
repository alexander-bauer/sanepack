package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"text/template"
)

type DebianFrameworker struct {
	t *template.Template
}

const (
	DebianInfo = `To complete building the package, invoke:
    fakeroot dpkg-buildpackage`

	debianStandardsVersion = "3.9.3"
)

func (d DebianFrameworker) Info() string {
	return DebianInfo
}

func (d DebianFrameworker) Framework(p *Package) (err error) {
	// Begin by trying to load the templates.
	d.t, err = template.ParseGlob(path.Join(*fTemp, "debian", "*.template"))
	if err != nil {
		return
	}
	l.Debug("Loaded debian/*.template files")

	// First, create the debian/ directory.
	l.Debug("Attempting to create debian/ directory\n")
	err = os.Mkdir("debian", 0777)
	if err != nil {
		return
	}

	// Now go on to create the debian/control file.
	l.Debug("Attempting to create debian/control\n")
	err = d.control(p.ProjectName, p.Description, "",
		p.Section, p.Priority,
		p.Homepage, p.Architecture, p.Maintainer, p.BuildDepends, p.Depends,
		p.Recommends, p.Suggests, p.Conflicts, p.Provides, p.Replaces)
	if err != nil {
		return
	}

	// Try to create the debian/copyright file.
	l.Debug("Attempting to create debian/copyright\n")
	err = d.copyright(*p.Copyright, p.Homepage)
	if err != nil {
		return
	}

	// Try to copy over the debian/rules file.
	l.Debug("Attempting to create debian/rules\n")
	err = d.rules()
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
		Depends:          concat(", ", append(depends, "debhelper")...),
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

	err = d.t.ExecuteTemplate(f, "control.template", control)
	f.Close()
	return
}

// copyright creates a "debian/copyright" file using the given license
// type.
func (d DebianFrameworker) copyright(c Copyright, homepage string) (err error) {
	// Begin by trying to open the debian/copyright file.
	f, err := os.Create("debian/copyright")
	if err != nil {
		return
	}
	defer f.Close()

	c.Homepage = homepage

	// If the file is opened properly, run it through
	// d.t.ExecuteTemplate() and return any errors.
	return d.t.ExecuteTemplate(f, "copyright.template", c)
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

// rules copies the template file to "debian/rules" and does *not*
// interpret the template in any way.
func (d DebianFrameworker) rules() (err error) {
	f, err := os.Create("debian/rules")
	if err != nil {
		return
	}
	defer f.Close()

	// Though this is being run through t.ExecuteTemplate(), this
	// should perform a simple copy. In the future, this may be
	// revised to actually use the template.
	return d.t.ExecuteTemplate(f, "rules.template", nil)
}

type debianControlFile struct {
	Name, Section, Priority, Architecture, StandardsVersion string
	Homepage, Description, LongDescription                  string
	Maintainer                                              Person
	BuildDepends, Depends, Recommends                       string
	Suggests, Conflicts, Provides, Replaces                 string
	Include                                                 map[string]bool
}
