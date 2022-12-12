package pkg

const (
	SlingPropsPath = "/system/console/status-slingprops.txt"
)

type StatusManager struct {
	instance *Instance
}

// sling.run.mode.options = s7connect,dynamicmedia_scene7,dynamicmedia
func (sm StatusManager) SlingProps() map[string]string {
	return map[string]string{}
}

// Run Modes = [s7connect, crx3, author, sdk, local, live, crx3tar]
func (sm StatusManager) SlingSettings() map[string]string {
	return map[string]string{}
}
