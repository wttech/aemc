package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

// OSGi Facade for communicating with OSGi framework.
type OSGi struct {
	instance *Instance

	bundleManager    *OSGiBundleManager
	componentManager *OSGiComponentManager
	eventManager     *OSGiEventManager
	configManager    *OSGiConfigManager

	shutdownDelay time.Duration
}

func NewOSGi(instance *Instance) *OSGi {
	cv := instance.manager.aem.config.Values()

	return &OSGi{
		instance: instance,

		bundleManager:    NewBundleManager(instance),
		componentManager: NewComponentManager(instance),
		eventManager:     &OSGiEventManager{instance: instance},
		configManager:    &OSGiConfigManager{instance: instance},

		shutdownDelay: cv.GetDuration("instance.osgi.shutdown_delay"),
	}
}

func (o *OSGi) BundleManager() *OSGiBundleManager {
	return o.bundleManager
}

func (o *OSGi) ComponentManager() *OSGiComponentManager {
	return o.componentManager
}

func (o *OSGi) EventManager() *OSGiEventManager {
	return o.eventManager
}

func (o *OSGi) ConfigManager() *OSGiConfigManager {
	return o.configManager
}

func (o *OSGi) Shutdown() error {
	return o.shutdown("Stop")
}

func (o *OSGi) Restart() error {
	return o.shutdown("Restart")
}

const (
	VMStatPath = "/system/console/vmstat"
)

func (o *OSGi) shutdown(shutdownType string) error {
	log.Infof("%s > triggering OSGi shutdown of type '%s'", o.instance.ID(), shutdownType)
	response, err := o.instance.http.Request().SetFormData(map[string]string{
		"shutdown_type": shutdownType,
	}).Post(VMStatPath)
	if err != nil {
		return fmt.Errorf("%s > cannot trigger OSGi shutdown of type '%s': %w", o.instance.ID(), shutdownType, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot trigger OSGi shutdown of type '%s': %s", o.instance.ID(), shutdownType, response.Status())
	}
	time.Sleep(o.shutdownDelay)
	log.Infof("%s > triggered OSGi shutdown of type '%s'", o.instance.ID(), shutdownType)
	return nil
}
