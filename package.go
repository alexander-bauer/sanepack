package main

import ()

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
