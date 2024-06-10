// Package pkg provides configuration and AEM facade
package pkg

import (
	"github.com/wttech/aemc/pkg/base"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/content"
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
	contentManager  *content.Manager
	instanceManager *InstanceManager
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
	result.contentManager = content.NewManager(result.baseOpts)
	result.instanceManager = NewInstanceManager(result)
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

func (a *AEM) ContentManager() *content.Manager {
	return a.contentManager
}

func (a *AEM) Detached() bool {
	return !a.config.TemplateFileExists()
}
