package main

import (
	"os"
	"os/exec"
	"path"
	"time"
)

// A Frameworker is a type which is capable of building the framework
// necessary for the user to create a distributable package. For
// example, a Debian Frameworker would create a debian/ directory and
// inform the user how to proceed.
type Frameworker interface {
	// Info is information for the user on how to create the
	// distributable package. It will be invoked if Framework()
	// succeeds.
	Info() string

	// Framework invokes any and all commands necessary to create the
	// framework for a distributable package, such as creating a
	// debian/ directory and placing control files, copying
	// documentation, and such.
	Framework(*Package) error
}

type Package struct {
	// ProjectName is the name of the project as it should appear on
	// the final package.
	ProjectName string

	// ProjectOwners is a list of owners of the project.
	ProjectOwners []Person

	// Maintainer is the person who maintains the package, and may be
	// separate from the ProjectOwners.
	Maintainer Person

	// Description is a brief, one line description of the project.
	Description string

	// Homepage is a link (HTTP or HTTPS) to the project homepage.
	Homepage string

	// ManPages is a slice containing paths (relative to the top of
	// the package repository) of any and all manpages. Note that the
	// manpages should be named like "packagename.1," etc.
	ManPages []string

	// Copyright contains a series of file globs, owners, and license
	// types, such as GPL 3.0+.
	Copyright *Copyright

	// BuildDepends is a slice containing the package names (and
	// versions) of any packages required to build this one.
	BuildDepends []string

	// Depends is a slice containing the package names (and versions)
	// of any packages required to run this one.
	Depends []string

	// Recommends, Suggests, Conflicts, Provides, and Replaces are
	// non-required slices containing package names (and versions) of
	// any packages as indicated by the name.
	Recommends, Suggests, Conflicts, Provides, Replaces []string

	// Section is the section of the repository, if applicable, to
	// mark the package as part of, such as "devel" for Debian.
	Section string

	// Priority is the priority at which the package should be
	// installed. This is usually "optional."
	Priority string

	// Architecture is the processor architecture for which the
	// package is or can be compiled. If the package is in an
	// interpreted language such as Python, it should be "all," and if
	// it does not require a specific platform, it should be "any."
	Architecture string
}

type Person struct {
	Name, Email string
}

type Copyright struct {
	Name, License string
	Homepage      string `json:",omitempty"`
	Files         []*fileCopyright
}

type fileCopyright struct {
	Glob, License string
	Year          int
	Owner         Person
}

// templatePackage attempts to use the current working directory to
// fill out a template Package object.
func templatePackage() (p *Package) {
	p = new(Package)
	// First, try to find the ProjectName. We will assume that the
	// package name is the name of the directory that sanepack was
	// invoked from.
	if wd, err := os.Getwd(); err == nil { // Note that err == nil
		p.ProjectName = path.Base(wd)
		l.Debugf("Found ProjectName: %q\n", p.ProjectName)
	} else { // If this fails, it will be left blank.
		l.Debugf("Could not get ProjectName: %s", err)
	}
	// Now, try to find the current user's name using git
	// defaults. This will be used to initialize ProjectOwners and
	// Maintainer.
	// TODO: allow use of other systems than git
	var user Person
	if vcsName, err := exec.Command("git", "config", "--global",
		"user.name").Output(); err == nil { // Again, err == nil
		user.Name = string(vcsName[:len(vcsName)-1]) // -1 to remove \n
		l.Debugf("Found current user name: %q\n", user.Name)
	} else {
		l.Debugf("Could not get current user name: %q", err)
	}

	// Same as above, but for email.
	if vcsEmail, err := exec.Command("git", "config", "--global",
		"user.email").Output(); err == nil {
		user.Email = string(vcsEmail[:len(vcsEmail)-1]) // as above
		l.Debugf("Found current user email: %q\n", user.Email)
	} else {
		l.Debugf("Could not get current user email: %s", err)
	}

	// Whether user could be initialized or not, we'll use it to fill
	// out the fields.
	p.ProjectOwners = []Person{user}
	p.Maintainer = user

	// Set up ManPages with an initialized slice.
	p.ManPages = make([]string, 1)
	p.ManPages[0] = "path/to/manpage.1"

	// Try to initialize Copyright with sane defaults.
	p.Copyright = &Copyright{
		Name:     p.ProjectName,
		License:  "abbreviated license name (such as GPL 3.0+)",
		Homepage: p.Homepage,
		Files:    make([]*fileCopyright, 1),
	}
	year, _, _ := time.Now().Date()
	p.Copyright.Files[0] = &fileCopyright{
		Glob:    "*",
		License: "GPL 3.0+",
		Year:    year,
		Owner:   user,
	}

	// Set up BuildDepends and Depends with initialized slices.
	p.BuildDepends = make([]string, 1)
	p.BuildDepends[0] = "package for your compiler here"

	p.Depends = make([]string, 1)
	p.Depends[0] = "package(s) required to run this package"

	// Here, we assume that the section is "main." This may be
	// incorrect, but it will serve as a template.
	p.Section = "main"

	// Similarly, "optional" is a common package priority.
	p.Priority = "optional"

	// We will make the assumption that the package is compiled and
	// could be compiled on any processor architecture.
	p.Architecture = "any"

	return
}
