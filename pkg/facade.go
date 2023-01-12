// Package pkg provides configuration and AEM facade
package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/base"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/java"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Aem is a facade to access AEM-related API
type Aem struct {
	output          io.Writer
	baseOpts        *base.Opts
	javaOpts        *java.Opts
	instanceManager *InstanceManager
}

// NewAem creates Aem facade
func NewAem() *Aem {
	result := new(Aem)
	result.output = os.Stdout
	result.baseOpts = base.NewOpts()
	result.javaOpts = java.NewOpts()
	result.instanceManager = NewInstanceManager(result)
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

func (a *Aem) InstanceManager() *InstanceManager {
	return a.instanceManager
}

func (a *Aem) Configure(config *cfg.Config) {
	a.baseOpts.Configure(config)
	a.javaOpts.Configure(config)
	a.instanceManager.Configure(config)
}

func (a *Aem) Build(command string, outputFile string, inputPaths, inputExcludes []string) (bool, error) {
	outputFileExists, err := pathx.ExistsStrict(outputFile)
	if err != nil {
		return false, err
	}
	checksumFile := outputFile + ".build"
	checksumExists, err := pathx.ExistsStrict(checksumFile)
	if err != nil {
		return false, err
	}
	checksumPrevious := ""
	if checksumExists {
		checksumPrevious, err = filex.ReadString(checksumFile)
		if err != nil {
			return false, err
		}
	}
	checksumCurrent, err := filex.ChecksumPaths(inputPaths, inputExcludes)
	if err != nil {
		return false, err
	}
	upToDate := outputFileExists && checksumPrevious == checksumCurrent
	if upToDate {
		return false, nil
	}
	err = a.executeBuild(command)
	if err != nil {
		return false, err
	}
	if err = filex.Write(checksumFile, checksumCurrent); err != nil {
		return false, err
	}
	return true, nil
}

func (a *Aem) executeBuild(command string) error {
	commandParts := strings.Split(command, " ")
	commandName := commandParts[0]
	commandArgs := commandParts[1:]
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Stdout = a.output
	cmd.Stderr = a.output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute build command '%s': %w", command, err)
	}
	return nil
}
