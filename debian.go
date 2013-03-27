package main

type DebianFrameworker struct{}

const (
	DebianInfo = `To complete building the package, invoke:
    fakeroot dpkg-buildpackage`
)

func (d DebianFrameworker) Info() string {
	return DebianInfo
}

func (d DebianFrameworker) Framework(p *Package) (err error) {
	l.Debugf("Pretended to build Debian framework")
	return nil
}
