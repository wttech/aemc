package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"os/exec"
	"strings"
)

type AppManager struct {
	aem *Aem

	SourcesIgnored []string
}

func NewAppManager(aem *Aem) *AppManager {
	result := new(AppManager)
	result.aem = aem

	result.SourcesIgnored = []string{
		// AEM files
		"**/aem/home/**",

		// meta files
		"**/.*",
		"**/.*/**",
		"!.content.xml",

		// build files
		"**/target",
		"**/target/**",
		"**/build",
		"**/build/**",
		"**/dist",
		"**/dist/**",
		"**/generated",
		"**/generated/**",
		"package-lock.json",
		"**/package-lock.json",

		// temporary files
		"*.log",
		"*.tmp",
		"**/node_modules",
		"**/node_modules/**",
		"**/node",
		"**/node/**",
	}

	return result
}

func (am *AppManager) Build(command string) error {
	parts := strings.Split(command, " ")
	name := parts[0]
	args := parts[1:]

	cmd := exec.Command(name, args...)
	cmd.Stdout = am.aem.output
	cmd.Stderr = am.aem.output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute build command '%s': %w", command, err)
	}
	return nil
}

func (am *AppManager) BuildWithChanged(command string, file string, sourcePaths []string) (bool, error) {
	fileExists, err := pathx.ExistsStrict(file)
	if err != nil {
		return false, err
	}
	checksumFile := file + ".build"
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
	checksumCurrent, err := filex.ChecksumPaths(sourcePaths, am.SourcesIgnored)
	if err != nil {
		return false, err
	}
	upToDate := fileExists && checksumPrevious == checksumCurrent
	if upToDate {
		return false, nil
	}
	err = am.Build(command)
	if err != nil {
		return false, err
	}
	if err = filex.Write(checksumFile, checksumCurrent); err != nil {
		return false, err
	}
	return true, nil
}

func (am *AppManager) Configure(config *cfg.Config) {
	opts := config.Values().App

	if len(opts.SourceExcludes) > 0 {
		am.SourcesIgnored = opts.SourceExcludes
	}
}
