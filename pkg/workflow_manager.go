package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"time"
)

type WorkflowManager struct {
	instance *Instance

	LauncherLibRoot            string
	LauncherConfigRoot         string
	LauncherToggleRetryTimeout time.Duration
	LauncherToggleRetryDelay   time.Duration
}

func NewWorkflowManager(i *Instance) *WorkflowManager {
	cv := i.manager.aem.config.Values()

	return &WorkflowManager{
		instance: i,

		LauncherLibRoot:            cv.GetString("instance.workflow.launcher.lib_root"),
		LauncherConfigRoot:         cv.GetString("instance.workflow.launcher.config_root"),
		LauncherToggleRetryTimeout: cv.GetDuration("instance.workflow.launcher.toggle_retry.timeout"),
		LauncherToggleRetryDelay:   cv.GetDuration("instance.workflow.launcher.toggle_retry.delay"),
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
		if time.Now().After(started.Add(w.LauncherToggleRetryTimeout)) {
			return fmt.Errorf("%s > awaiting workflow launcher action '%s' timed out after %s: %w", w.instance.ID(), name, w.LauncherToggleRetryTimeout, err)
		}
		time.Sleep(w.LauncherToggleRetryDelay)
	}
}

func (w *WorkflowManager) disableLaunchers(launchers []WorkflowLauncher) {
	for _, launcher := range launchers {
		if err := w.doLauncherAction("disable", func() error {
			if err := launcher.Prepare(); err != nil {
				return err
			}
			enabled, err := launcher.IsEnabled()
			if err != nil {
				return err
			}
			if enabled {
				if err := launcher.Disable(); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			log.Warnf("%s", err)
			continue
		}
	}
}

func (w *WorkflowManager) enableLaunchers(launchers []WorkflowLauncher) {
	for _, launcher := range launchers {
		if err := w.doLauncherAction("enable", func() error {
			toggled, err := launcher.IsToggled()
			if err != nil {
				return err
			}
			if toggled {
				if err := launcher.Enable(); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			log.Warnf("%s", err)
			continue
		}
	}
}

const (
	WorkflowLauncherLibRoot    = "/libs/settings/workflow/launcher"
	WorkflowLauncherConfigRoot = "/conf/global/settings/workflow/launcher"
)
