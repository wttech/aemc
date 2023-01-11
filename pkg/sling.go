package pkg

// Sling Facade for communicating with Sling framework.
type Sling struct {
	jmx       *JMX
	installer *SlingInstaller
}

func NewSling(instance *Instance) *Sling {
	return &Sling{NewJMX(instance), NewSlingInstaller(instance)}
}

func (s *Sling) JMX() *JMX {
	return s.jmx
}

func (s *Sling) Installer() *SlingInstaller {
	return s.installer
}
