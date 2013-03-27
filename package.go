package main

import ()

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
}

type Person struct {
	Name, Email string
}
