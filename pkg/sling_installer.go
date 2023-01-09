package pkg

type SlingInstaller struct {
	instance *Instance
}

const (
	SlingInstallerJMXBeanName = "org/apache/sling/installer/Installer/Sling%20OSGi%20Installer"
	SlingInstallerPauseRoot   = "/system/sling/installer/jcr/pauseInstallation"
)

func NewSlingInstaller(instance *Instance) *SlingInstaller {
	return &SlingInstaller{instance}
}

func (i SlingInstaller) State() (*SlingInstallerJMXBean, error) {
	bean := &SlingInstallerJMXBean{}
	if err := i.instance.sling.jmx.ReadBean(SlingInstallerJMXBeanName, bean); err != nil {
		return nil, err
	}
	return bean, nil
}

func (i SlingInstaller) CountPauses() (int, error) {
	pauseNodes := i.instance.Repo().Node(SlingInstallerPauseRoot).ChildrenList() // TODO handle error
	return len(pauseNodes), nil
}

type SlingInstallerJMXBean struct {
	Active                 bool `json:"Active"`
	SuspendedSince         int  `json:"SuspendedSince"`
	ActiveResourceCount    int  `json:"ActiveResourceCount"`
	InstalledResourceCount int  `json:"InstalledResourceCount"`
}

func (b SlingInstallerJMXBean) IsBusy() bool {
	return b.Active || b.ActiveResourceCount > 0
}
