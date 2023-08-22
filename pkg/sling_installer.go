package pkg

import (
	"github.com/spf13/cast"
)

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
	Active                 bool `json:"Active"`
	SuspendedSince         int  `json:"SuspendedSince"`
	ActiveResourceCount    any  `json:"ActiveResourceCount"`    // AEM type bug: sometimes 'int' or 'string'
	InstalledResourceCount any  `json:"InstalledResourceCount"` // AEM type bug: sometimes 'int' or 'string'
}

func (b SlingInstallerJMXBean) IsActive() bool {
	return b.Active /* || b.ActiveResources() > 0 */ // sometimes ActiveResourceCount > 0 but Active == false
}

func (b SlingInstallerJMXBean) ActiveResources() int {
	return cast.ToInt(b.ActiveResourceCount)
}

func (b SlingInstallerJMXBean) InstalledResources() int {
	return cast.ToInt(b.InstalledResourceCount)
}
