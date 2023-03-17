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

// Aem is a facade to access AEM-related API
type Aem struct {
	output          io.Writer
	config          *cfg.Config
	project         *project.Project
	baseOpts        *base.Opts
	javaOpts        *java.Opts
	instanceManager *InstanceManager
}

// NewAem creates Aem facade
func NewAem() *Aem {
	result := new(Aem)
	result.output = os.Stdout
	result.config = cfg.NewConfig()
	result.project = project.New(result)
	result.baseOpts = base.NewOpts(result.config)
	result.javaOpts = java.NewOpts(result.config)
	result.instanceManager = NewInstanceManager(result)
	return result
}

func (a *Aem) Output() io.Writer {
	return a.output
}

func (a *Aem) SetOutput(output io.Writer) {
	a.output = output
}

func (a *Aem) CommandOutput(cmd *exec.Cmd) {
	cmd.Stdout = a.output
	cmd.Stderr = a.output
}

func (a *Aem) Config() *cfg.Config {
	return a.config
}

func (a *Aem) BaseOpts() *base.Opts {
	return a.baseOpts
}

func (a *Aem) JavaOpts() *java.Opts {
	return a.javaOpts
}

func (a *Aem) InstanceManager() *InstanceManager {
	return a.instanceManager
}
