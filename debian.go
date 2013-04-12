package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"
)

type DebianFrameworker struct {
	t *template.Template
}

const (
	DebianInfo = `To complete building the package, invoke:
    fakeroot dpkg-buildpackage`

	debianStandardsVersion = "3.9.3"
	debianCompatVersion    = "8"
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

	l.Debug("Attempting to create debian/ directory\n")
	err = os.Mkdir("debian", 0777)
	if err != nil {
		return
	}

	l.Debug("Attempting to create debain/changelog\n")
	err = d.changelog(p.ProjectName, p.Maintainer)
	if err != nil {
		return
	}

	l.Debug("Attempting to create debian/control\n")
	err = d.control(p.ProjectName, p.Description, "",
		p.Section, p.Priority,
		p.Homepage, p.Architecture, p.Maintainer, p.BuildDepends, p.Depends,
		p.Recommends, p.Suggests, p.Conflicts, p.Provides, p.Replaces)
	if err != nil {
		return
	}

	l.Debug("Attempting to create debian/compat\n")
	err = d.compat()
	if err != nil {
		return
	}

	l.Debug("Attempting to create debian/copyright\n")
	err = d.copyright(*p.Copyright, p.Homepage)
	if err != nil {
		return
	}

	l.Debug("Attempting to create debian/rules\n")
	err = d.rules()
	if err != nil {
		return
	}

	l.Debug("Attempting to create debian/docs\n")
	err = d.docs(p.Docs)
	if err != nil {
		return
	}

	if len(p.InitScript) > 0 { // Only do this if p.InitScript is set
		l.Debugf("Attempting to create debian/%s.init\n", p.ProjectName)
		err = d.initscript(p.ProjectName, p.InitScript)
		if err != nil {
			return
		}
	} else {
		l.Debugf("Skipped creating debian/%s.init file\n", p.ProjectName)
	}

	l.Debugf("Attempting to create debian/%s.manpages\n", p.ProjectName)
	err = d.manpages(p.ProjectName, p.ManPages)
	if err != nil {
		return
	}

	return
}

// changelog creates a debian/changelog and reads the version control
// changelog in order to populate it.
func (d DebianFrameworker) changelog(name string, maintainer Person) (err error) {
	// First, read the log to get a list of changes.
	logoutput, err := exec.Command(
		"git", "--no-pager", "log", "--simplify-merges",
		"--pretty=format:%s").Output()
	if err != nil {
		return
	}

	// Second, use git describe to get the tag, and only the tag.
	tag, err := exec.Command(
		"git", "describe", "--abbrev=0", "--tags", "--match=v*").Output()
	if err != nil {
		return
	}
	// The Version is slightly more finnicky than the tag; it must
	// start with a decimal number, and we must make sure not to
	// include the final newline from the command. Thus, we trim "\n"
	// from the right and "v" or "V" from the left.
	version := strings.TrimLeft(
		strings.TrimRight(string(tag), "\n"), "vV")

	changelog := &debianChangelogFile{
		Name:       name,
		Version:    version,
		Date:       time.Now().Format(time.RFC1123Z),
		Maintainer: maintainer,
		Changes:    strings.Split(string(logoutput), "\n"),
	}

	// Now, create and open debian/changelog for writing.
	f, err := os.Create("debian/changelog")
	if err != nil {
		return
	}
	defer f.Close()

	// Finally, run t.ExecuteTemplate() and return any errors.
	return d.t.ExecuteTemplate(f, "changelog.template", changelog)
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

// compat creats a "debian/compat" file using the global compat
// version.
func (d DebianFrameworker) compat() (err error) {
	// Begin by trying to open the debian/compat file.
	f, err := os.Create("debian/compat")
	if err != nil {
		return
	}
	defer f.Close()

	// Now, write in the contents.
	fmt.Fprintln(f, debianCompatVersion)
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

// docs creates a "debian/docs" file containing every given path to a
// non-manpage document, one per line.
func (d DebianFrameworker) docs(documents []string) (err error) {
	// Begin by opening the file.
	f, err := os.Create("debian/docs")
	if err != nil {
		return
	}
	defer f.Close()

	// Now use fmt.Fprintln() to populate it.
	for _, document := range documents {
		fmt.Fprintln(f, document)
	}

	f.Close()
	return
}

// initscript creates a "debian/<name>.init" file with the contents of
// the specified file.
func (d DebianFrameworker) initscript(name, initscript string) (err error) {
	// Begin by trying to open the initscript file.
	fi, err := os.Open(initscript)
	if err != nil {
		return
	}
	defer fi.Close()
	// If that succeeds, open the debian/<name>.init file.
	fo, err := os.Create("debian/" + name + ".init")
	if err != nil {
		return
	}
	defer fo.Close()

	// Now, copy the contents directly using io.Copy().
	_, err = io.Copy(fo, fi) // (destination, source)
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

// rules copies the template file to "debian/rules" and does *not*
// interpret the template in any way.
func (d DebianFrameworker) rules() (err error) {
	// Begin by trying to open the initscript file.
	fi, err := os.Open(path.Join(*fTemp, "debian/rules"))
	if err != nil {
		return
	}
	defer fi.Close()
	// If that succeeds, open the debian/rules file in the same manner
	// that os.Create() would, but with the executable permission set.
	fo, err := os.OpenFile("debian/rules",
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer fo.Close()

	// Now, copy the contents directly using io.Copy().
	_, err = io.Copy(fo, fi) // (destination, source)
	return
}

type debianChangelogFile struct {
	Name, Version, Date string
	Maintainer          Person
	Changes             []string
}

type debianControlFile struct {
	Name, Section, Priority, Architecture, StandardsVersion string
	Homepage, Description, LongDescription                  string
	Maintainer                                              Person
	BuildDepends, Depends, Recommends                       string
	Suggests, Conflicts, Provides, Replaces                 string
	Include                                                 map[string]bool
}
