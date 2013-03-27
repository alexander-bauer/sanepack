package main

import (
	"os"
	"os/exec"
	"path"
)

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

	return
}
