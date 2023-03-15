package pkg

import (
	"fmt"
	"strings"
)

type WorkflowLauncher struct {
	manager *WorkflowManager

	path string
}

func (l WorkflowLauncher) LibNode() RepoNode {
	return l.manager.instance.repo.Node(l.path)
}

func (l WorkflowLauncher) ConfigNode() RepoNode {
	return l.manager.instance.repo.Node(strings.ReplaceAll(l.path, WorkflowLauncherLibRoot+"/", WorkflowLauncherConfigRoot+"/"))
}

func (l WorkflowLauncher) String() string {
	return fmt.Sprintf("workflow launcher '%s'", l.path)
}

const (
	WorkflowLauncherEnabledProp = "enabled"
	WorkflowLauncherToggledProp = "toggled"
)
