package pkg

// OSGi Facade for communicating with OSGi framework.
type OSGi struct {
	bundleManager *OSGiBundleManager
	eventManager  *OSGiEventManager
	configManager *OSGiConfigManager
}

func NewOSGi(instance *Instance) *OSGi {
	return &OSGi{
		bundleManager: NewBundleManager(instance),
		eventManager:  &OSGiEventManager{instance: instance},
		configManager: &OSGiConfigManager{instance: instance},
	}
}

func (o *OSGi) BundleManager() *OSGiBundleManager {
	return o.bundleManager
}

func (o *OSGi) EventManager() *OSGiEventManager {
	return o.eventManager
}

func (o *OSGi) ConfigManager() *OSGiConfigManager {
	return o.configManager
}
