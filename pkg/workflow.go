package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
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

func (l WorkflowLauncher) Prepare() error {
	configNode := l.ConfigNode()
	libNode := l.LibNode()
	configExists, err := configNode.Exists()
	if err != nil {
		return fmt.Errorf("%s > cannot read workflow launcher config '%s': %w", l.manager.instance.IDColor(), configNode.path, err)
	}
	if !configExists {
		if err := libNode.Copy(configNode.Path()); err != nil {
			return fmt.Errorf("%s > workflow launcher config node '%s' cannot be copied from lib node '%s': %s", l.manager.instance.IDColor(), configNode.path, libNode.path, err)
		}
	}
	return nil
}

func (l WorkflowLauncher) Disable() error {
	node := l.ConfigNode()
	if err := node.Save(map[string]any{
		WorkflowLauncherEnabledProp: false,
		WorkflowLauncherToggledProp: true,
	}); err != nil {
		return fmt.Errorf("%s > cannot disable workflow launcher '%s': %w", l.manager.instance.IDColor(), node.path, err)
	}
	log.Infof("%s > disabled workflow launcher '%s'", l.manager.instance.IDColor(), node.path)
	return nil
}

func (l WorkflowLauncher) Enable() error {
	node := l.ConfigNode()
	if err := node.Save(map[string]any{
		WorkflowLauncherEnabledProp: true,
		WorkflowLauncherToggledProp: nil,
	}); err != nil {
		return fmt.Errorf("%s > cannot enable workflow launcher '%s': %w", l.manager.instance.IDColor(), node.path, err)
	}
	log.Infof("%s > enabled workflow launcher '%s'", l.manager.instance.IDColor(), node.path)
	return nil
}

func (l WorkflowLauncher) IsEnabled() (bool, error) {
	configNode := l.ConfigNode()
	configState, err := configNode.State()
	if err != nil {
		return false, fmt.Errorf("%s > cannot read workflow launcher config just copied '%s': %w", l.manager.instance.IDColor(), configNode.path, err)
	}
	enabledAny, enabledFound := configState.Properties[WorkflowLauncherEnabledProp]
	enabled := enabledFound && cast.ToBool(enabledAny)
	return enabled, nil
}

func (l WorkflowLauncher) IsToggled() (bool, error) {
	configNode := l.ConfigNode()
	configState, err := configNode.State()
	if err != nil {
		return false, fmt.Errorf("%s > cannot read config of workflow launcher '%s': %w", l.manager.instance.IDColor(), configNode.path, err)
	}
	if !configState.Exists {
		return false, fmt.Errorf("%s > config node of workflow launcher does not exist '%s': %w", l.manager.instance.IDColor(), configNode.path, err)
	}
	toggledAny, toggledFound := configState.Properties[WorkflowLauncherToggledProp]
	toggled := toggledFound && cast.ToBool(toggledAny)
	return toggled, nil
}

func (l WorkflowLauncher) String() string {
	return fmt.Sprintf("workflow launcher '%s'", l.path)
}

const (
	WorkflowLauncherEnabledProp = "enabled"
	WorkflowLauncherToggledProp = "toggled"
)
