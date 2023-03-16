package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"time"
)

type WorkflowManager struct {
	instance *Instance

	LibRoot            string
	ConfigRoot         string
	ToggleRetryTimeout time.Duration
	ToggleRetryDelay   time.Duration
}

func NewWorkflowManager(i *Instance) *WorkflowManager {
	return &WorkflowManager{
		instance: i,

		LibRoot:            WorkflowLauncherLibRoot,
		ConfigRoot:         WorkflowLauncherConfigRoot,
		ToggleRetryTimeout: time.Minute * 5,
		ToggleRetryDelay:   time.Second * 10,
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

func (w *WorkflowManager) doLauncherAction(name string, callback func() error) error {
	started := time.Now()
	for {
		err := callback()
		if err == nil {
			return nil
		}
		if time.Now().After(started.Add(w.ToggleRetryTimeout)) {
			return fmt.Errorf("%s > awaiting workflow launcher action '%s' timed out after %s: %w", w.instance.ID(), name, w.ToggleRetryTimeout, err)
		}
		time.Sleep(w.ToggleRetryDelay)
	}
}

func (w *WorkflowManager) disableLaunchers(launchers []WorkflowLauncher) {
	for _, launcher := range launchers {
		if err := w.doLauncherAction("disable", func() error {
			return w.disableLauncher(launcher)
		}); err != nil {
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
		return fmt.Errorf("%s > cannot read workflow launcher config '%s': %w", w.instance.ID(), configNode.path, err)
	}
	if !configState.Exists {
		if err := libNode.Copy(configNode.Path()); err != nil {
			return fmt.Errorf("%s > workflow launcher config node '%s' cannot be copied from lib node '%s': %s", w.instance.ID(), configNode.path, libNode.path, err)
		}
		configState, err = configNode.State()
		if err != nil {
			return fmt.Errorf("%s > cannot read workflow launcher config just copied '%s': %w", w.instance.ID(), configNode.path, err)
		}
	}
	enabledAny, enabledFound := configState.Properties[WorkflowLauncherEnabledProp]
	enabled := enabledFound && cast.ToBool(enabledAny)
	if enabled {
		if err := configNode.Save(map[string]any{
			WorkflowLauncherEnabledProp: false,
			WorkflowLauncherToggledProp: true,
		}); err != nil {
			return fmt.Errorf("%s > cannot disable workflow launcher '%s': %w", w.instance.ID(), configNode.path, err)
		}
		log.Infof("%s > disabled workflow launcher '%s'", w.instance.ID(), configNode.path)
	}
	return nil
}

func (w *WorkflowManager) enableLaunchers(launchers []WorkflowLauncher) {
	for _, launcher := range launchers {
		if err := w.doLauncherAction("enable", func() error {
			return w.enableLauncher(launcher)
		}); err != nil {
			log.Warnf("%s", err)
			continue
		}
	}
}

func (w *WorkflowManager) enableLauncher(launcher WorkflowLauncher) error {
	configNode := launcher.ConfigNode()
	configState, err := configNode.State()
	if err != nil {
		return fmt.Errorf("%s > cannot read config of workflow launcher '%s': %w", w.instance.ID(), configNode.path, err)
	}
	if !configState.Exists {
		return fmt.Errorf("%s > config node of workflow launcher does not exist '%s': %w", w.instance.ID(), configNode.path, err)
	}
	toggledAny, toggledFound := configState.Properties[WorkflowLauncherToggledProp]
	toggled := toggledFound && cast.ToBool(toggledAny)
	if toggled {
		if err := configNode.Save(map[string]any{
			WorkflowLauncherEnabledProp: true,
			WorkflowLauncherToggledProp: nil,
		}); err != nil {
			return fmt.Errorf("%s > cannot enable workflow launcher '%s': %w", w.instance.ID(), configNode.path, err)
		}
		log.Infof("%s > enabled workflow launcher '%s'", w.instance.ID(), configNode.path)
	}
	return nil
}

const (
	WorkflowLauncherLibRoot    = "/libs/settings/workflow/launcher"
	WorkflowLauncherConfigRoot = "/conf/global/settings/workflow/launcher"
)
