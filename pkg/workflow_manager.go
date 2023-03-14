package pkg

type WorkflowManager struct {
	instance *Instance
}

func NewWorkflowManager(i *Instance) *WorkflowManager {
	return &WorkflowManager{instance: i}
}

func (w *WorkflowManager) Launcher(path string) *WorkflowLauncher {
	return &WorkflowLauncher{manager: w, path: path}
}
