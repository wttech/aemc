package pkg

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

const (
	WorkflowLauncherLibRoot    = "/libs/settings/workflow/launcher"
	WorkflowLauncherConfigRoot = "/conf/global/settings/workflow/launcher"
)
