package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
)

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
	launchers, err := w.findLaunchers(libPaths)
	if err != nil {
		return err
	}
	defer func(launchers []WorkflowLauncher) { w.enableLaunchers(launchers) }(launchers)
	w.disableLaunchers(launchers)
	return action()
}

func (w *WorkflowManager) findLaunchers(paths []string) ([]WorkflowLauncher, error) {
	var result []WorkflowLauncher
	for _, libPath := range paths {
		dir, fileName := pathx.DirAndFileName(libPath)
		dirNode := w.instance.repo.Node(dir)
		children, err := dirNode.Children()
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			if stringsx.Match(child.Name(), fileName) {
				result = append(result, *(w.Launcher(child.Path())))
			}
		}
	}
	return result, nil
}

func (w *WorkflowManager) disableLaunchers(launchers []WorkflowLauncher) {
	for _, launcher := range launchers {
		if err := w.disableLauncher(launcher); err != nil {
			log.Warnf("%s", err)
			continue
		}
	}
}

func (w *WorkflowManager) disableLauncher(launcher WorkflowLauncher) error {
	configNode := launcher.ConfigNode()
	libNode := launcher.LibNode()
	configState, err := configNode.State()
	if err != nil {
		return fmt.Errorf("%s > workflow launcher config state cannot be read '%s'", w.instance.ID(), err)
	}
	if !configState.Exists {
		if err := libNode.Copy(configNode.Path()); err != nil {
			return fmt.Errorf("%s > workflow launcher config node '%s' cannot be copied from lib node '%s': %s", w.instance.ID(), configNode.path, libNode.path, err)
		}
	}
	enabledAny, enabledFound := configState.Properties[WorkflowLauncherEnabledProp]
	enabled := enabledFound && cast.ToBool(enabledAny)
	if enabled {
		if err := configNode.Save(map[string]any{
			WorkflowLauncherEnabledProp: false,
			WorkflowLauncherToggledProp: true,
		}); err != nil {
			return fmt.Errorf("%s > workflow launcher '%s' cannot be disabled: %s", w.instance.ID(), configNode.path, err)
		}
		log.Infof("%s > workflow launcher '%s' disabled", w.instance.ID(), configNode.path)
	}
	return nil
}

func (w *WorkflowManager) enableLaunchers(launchers []WorkflowLauncher) {
	for _, launcher := range launchers {
		if err := w.enableLauncher(launcher); err != nil {
			log.Warnf("%s", err)
			continue
		}
	}
}

func (w *WorkflowManager) enableLauncher(launcher WorkflowLauncher) error {
	configNode := launcher.ConfigNode()
	configState, err := configNode.State()
	if err != nil {
		return fmt.Errorf("%s > workflow launcher config cannot be read: %w", w.instance.ID(), err)
	}
	if !configState.Exists {
		return fmt.Errorf("%s > workflow launcher config does not exist: %w", w.instance.ID(), err)
	}
	toggledAny, toggledFound := configState.Properties[WorkflowLauncherToggledProp]
	toggled := toggledFound && cast.ToBool(toggledAny)
	if toggled {
		if err := configNode.Save(map[string]any{
			WorkflowLauncherEnabledProp: true,
			WorkflowLauncherToggledProp: nil,
		}); err != nil {
			return fmt.Errorf("%s > workflow launcher '%s' cannot be disabled: %s", w.instance.ID(), configNode.path, err)
		}
		log.Infof("%s > workflow launcher '%s' disabled", w.instance.ID(), configNode.path)
	}
	return nil
}

const (
	WorkflowLauncherLibRoot    = "/libs/settings/workflow/launcher"
	WorkflowLauncherConfigRoot = "/conf/global/settings/workflow/launcher"
)
