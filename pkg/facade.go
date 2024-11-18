// Package pkg provides configuration and AEM facade
package pkg

import (
	"github.com/wttech/aemc/pkg/cfg"
	"io"
	"os"
	"os/exec"
)

// AEM is a facade to access AEM-related API
type AEM struct {
	output   io.Writer
	config   *cfg.Config
	project  *Project
	baseOpts *BaseOpts

	vendorManager   *VendorManager
	instanceManager *InstanceManager
	contentManager  *ContentManager
}

func DefaultAEM() *AEM {
	return NewAEM(cfg.NewConfig())
}

func NewAEM(config *cfg.Config) *AEM {
	result := new(AEM)
	result.output = os.Stdout
	result.config = config
	result.project = NewProject(result)
	result.baseOpts = NewBaseOpts(result)
	result.vendorManager = NewVendorManager(result)
	result.instanceManager = NewInstanceManager(result)
	result.contentManager = NewContentManager(result)
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

func (a *AEM) BaseOpts() *BaseOpts {
	return a.baseOpts
}

func (a *AEM) VendorManager() *VendorManager {
	return a.vendorManager
}

func (a *AEM) InstanceManager() *InstanceManager {
	return a.instanceManager
}

func (a *AEM) ContentManager() *ContentManager {
	return a.contentManager
}

func (a *AEM) Project() *Project {
	return a.project
}

func (a *AEM) Detached() bool {
	return !a.config.TemplateFileExists()
}
