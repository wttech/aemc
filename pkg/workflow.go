package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

func (l WorkflowLauncher) State() (*RepoNodeState, error) {
	return l.LibNode().State()
}

func (l WorkflowLauncher) Enable() (bool, error) {
	return l.Toggle(true)
}

func (l WorkflowLauncher) Disable() (bool, error) {
	return l.Toggle(false)
}

func (l WorkflowLauncher) Toggle(flag bool) (bool, error) {
	configNode := l.ConfigNode()
	configNodeExists, err := configNode.ReadExists()
	if err != nil {
		return false, err
	}
	if !configNodeExists {
		if err := l.LibNode().Copy(configNode.Path()); err != nil {
			return false, err
		}
	}
	state, err := configNode.State()
	if err != nil {
		return false, err
	}
	flagCurrentAny, ok := state.Properties[WorkflowLauncherEnabledProp]
	flagChangeNeeded := true
	if ok {
		flagChangeNeeded = flagCurrentAny.(bool) != flag
	}
	if !flagChangeNeeded {
		return false, nil
	}
	if flag {
		log.Infof("%s > enabling workflow launcher '%s'", l.manager.instance.ID(), l.path)
	} else {
		log.Infof("%s > disabling workflow launcher '%s'", l.manager.instance.ID(), l.path)
	}
	if err := configNode.SaveProp(WorkflowLauncherEnabledProp, flag); err != nil {
		return false, err
	}
	if flag {
		log.Infof("%s > enabled workflow launcher '%s'", l.manager.instance.ID(), l.path)
	} else {
		log.Infof("%s > disabled workflow launcher '%s'", l.manager.instance.ID(), l.path)
	}
	return true, nil
}

func (l WorkflowLauncher) String() string {
	return fmt.Sprintf("workflow launcher '%s'", l.path)
}

const (
	WorkflowLauncherEnabledProp = "enabled"
)
