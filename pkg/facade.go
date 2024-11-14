// Package pkg provides configuration and AEM facade
package pkg

import (
	"github.com/wttech/aemc/pkg/base"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/java"
	"github.com/wttech/aemc/pkg/project"
	"io"
	"os"
	"os/exec"
)

// AEM is a facade to access AEM-related API
type AEM struct {
	output          io.Writer
	config          *cfg.Config
	project         *project.Project
	baseOpts        *base.Opts
	javaOpts        *java.Opts
	contentManager  *ContentManager
	instanceManager *InstanceManager
	vaultCLI        *VaultCLI
}

func DefaultAEM() *AEM {
	return NewAEM(cfg.NewConfig())
}

func NewAEM(config *cfg.Config) *AEM {
	result := new(AEM)
	result.output = os.Stdout
	result.config = config
	result.project = project.New(result.config)
	result.baseOpts = base.NewOpts(result.config)
	result.javaOpts = java.NewOpts(result.baseOpts)
	result.contentManager = NewContentManager(result)
	result.instanceManager = NewInstanceManager(result)
	result.vaultCLI = NewVaultCLI(result)
	return result
}

func (a *AEM) Output() io.Writer {
	return a.output
}

func (a *AEM) SetOutput(output io.Writer) {
	a.output = output
}

func (a *AEM) CommandOutput(cmd *exec.Cmd) {
	cmd.Stdout = a.output
	cmd.Stderr = a.output
}

func (a *AEM) Config() *cfg.Config {
	return a.config
}

func (a *AEM) BaseOpts() *base.Opts {
	return a.baseOpts
}

func (a *AEM) JavaOpts() *java.Opts {
	return a.javaOpts
}

func (a *AEM) InstanceManager() *InstanceManager {
	return a.instanceManager
}

func (a *AEM) Project() *project.Project {
	return a.project
}

func (a *AEM) ContentManager() *ContentManager {
	return a.contentManager
}

func (a *AEM) VaultCLI() *VaultCLI {
	return a.vaultCLI
}

func (a *AEM) Detached() bool {
	return !a.config.TemplateFileExists()
}

// TODO move SDK, OakRun, Quickstart under 'vendor' key in the config and namespace in the code
func (a *AEM) Prepare() error {
	// validation phase (quick feedback)
	sdk, err := a.InstanceManager().LocalOpts.Quickstart.IsDistSDK()
	if err != nil {
		return err
	}
	// preparation phase (slow feedback)
	if err := a.BaseOpts().Prepare(); err != nil {
		return err
	}
	if err := a.JavaOpts().Prepare(); err != nil {
		return err
	}
	// TODO move SDK and Quickstart outside of instance manager
	if sdk {
		if err := a.InstanceManager().LocalOpts.SDK.Prepare(); err != nil {
			return err
		}
	}
	// TODO move OakRun outside of instance manager
	if err := a.InstanceManager().LocalOpts.OakRun.Prepare(); err != nil {
		return err
	}
	return nil
}
