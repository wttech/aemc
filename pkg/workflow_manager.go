package pkg

import log "github.com/sirupsen/logrus"

type WorkflowManager struct {
	instance *Instance

	libRoot    string
	configRoot string
}

func NewWorkflowManager(i *Instance) *WorkflowManager {
	return &WorkflowManager{
		instance: i,

		libRoot:    WorkflowLauncherLibRoot,
		configRoot: WorkflowLauncherConfigRoot,
	}
}

func (w *WorkflowManager) Launcher(path string) *WorkflowLauncher {
	return &WorkflowLauncher{manager: w, path: path}
}

func (w *WorkflowManager) ToggleLaunchers(libPaths []string, action func() error) error {
	all := w.FindLaunchers(libPaths)
	var disabled []WorkflowLauncher
	for _, launcher := range all {
		_, err := launcher.Disable()
		if err == nil {
			disabled = append(disabled, launcher)
		} else {
			log.Warnf("%s > workflow launcher '%s' cannot be temporarily disabled", w.instance.ID(), launcher.path)
		}
	}
	defer func() {
		// TODO add retry mechanism until done
		for _, launcher := range disabled {
			if _, err := launcher.Enable(); err != nil {
				log.Warnf("%s > workflow launcher '%s' temporarily disabled cannot be reenabled", w.instance.ID(), launcher.path)
			}
		}
	}()
	return action()
}

func (w *WorkflowManager) FindLaunchers(paths []string) []WorkflowLauncher {
	return nil // TODO impl this
}

const (
	WorkflowLauncherLibRoot    = "/libs/settings/workflow/launcher"
	WorkflowLauncherConfigRoot = "/conf/global/settings/workflow/launcher"
)
