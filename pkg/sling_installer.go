package pkg

import "strconv"

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
	pauseNodes, err := i.instance.Repo().Node(SlingInstallerPauseRoot).Children()
	if err != nil {
		return -1, err
	}
	return len(pauseNodes), nil
}

type SlingInstallerJMXBean struct {
	Active                 bool   `json:"Active"`
	SuspendedSince         int    `json:"SuspendedSince"`
	ActiveResourceCount    string `json:"ActiveResourceCount"`
	InstalledResourceCount string `json:"InstalledResourceCount"`
}

func (b SlingInstallerJMXBean) IsBusy() bool {
	return b.Active || b.ActiveResources() > 0
}

func (b SlingInstallerJMXBean) ActiveResources() int {
	count, _ := strconv.Atoi(b.ActiveResourceCount)
	return count
}

func (b SlingInstallerJMXBean) InstalledResources() int {
	count, _ := strconv.Atoi(b.InstalledResourceCount)
	return count
}
