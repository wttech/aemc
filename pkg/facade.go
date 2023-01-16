// Package pkg provides configuration and AEM facade
package pkg

import (
	"github.com/wttech/aemc/pkg/base"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/java"
	"io"
	"os"
	"os/exec"
)

// Aem is a facade to access AEM-related API
type Aem struct {
	output          io.Writer
	baseOpts        *base.Opts
	javaOpts        *java.Opts
	instanceManager *InstanceManager
	appManager      *AppManager
}

// NewAem creates Aem facade
func NewAem() *Aem {
	result := new(Aem)
	result.output = os.Stdout
	result.baseOpts = base.NewOpts()
	result.javaOpts = java.NewOpts()
	result.instanceManager = NewInstanceManager(result)
	result.appManager = NewAppManager(result)
	return result
}

func (a *Aem) SetOutput(output io.Writer) {
	a.output = output
}

func (a *Aem) BaseOpts() *base.Opts {
	return a.baseOpts
}

func (a *Aem) JavaOpts() *java.Opts {
	return a.javaOpts
}

func (a *Aem) AppManager() *AppManager {
	return a.appManager
}

func (a *Aem) InstanceManager() *InstanceManager {
	return a.instanceManager
}

func (a *Aem) Configure(config *cfg.Config) {
	a.baseOpts.Configure(config)
	a.javaOpts.Configure(config)
	a.appManager.Configure(config)
	a.instanceManager.Configure(config)
}

func (a *Aem) CommandOutput(cmd *exec.Cmd) {
	cmd.Stdout = a.output
	cmd.Stderr = a.output
}
