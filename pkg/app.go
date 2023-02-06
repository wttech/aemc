package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/execx"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
)

type AppManager struct {
	aem *Aem

	SourcesIgnored []string
}

const (
	AppChecksumFileSuffix = ".build.md5"
)

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
	cmd := execx.CommandString(command)
	env, err := am.aem.javaOpts.Env()
	if err != nil {
		return err
	}
	cmd.Env = env
	am.aem.CommandOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute build command '%s': %w", command, err)
	}
	return nil
}

func (am *AppManager) BuildWithChanged(command string, filePattern string, sourcePaths []string) (string, bool, error) {
	file, err := pathx.GlobOne(filePattern)
	if file != "" {
		checksumFile := file + AppChecksumFileSuffix
		checksumExists, err := pathx.ExistsStrict(checksumFile)
		if err != nil {
			return file, false, err
		}
		checksumPrevious := ""
		if checksumExists {
			checksumPrevious, err = filex.ReadString(checksumFile)
			if err != nil {
				return file, false, err
			}
		}
		checksumCurrent, err := filex.ChecksumPaths(sourcePaths, am.SourcesIgnored)
		if err != nil {
			return file, false, err
		}
		if checksumPrevious == checksumCurrent {
			return file, false, nil
		}
	}
	err = am.Build(command)
	if err != nil {
		return "", false, err
	}
	file, err = pathx.GlobSome(filePattern)
	if err != nil {
		return "", false, err
	}
	checksumCurrent, err := filex.ChecksumPaths(sourcePaths, am.SourcesIgnored)
	if err != nil {
		return file, false, err
	}
	if err = filex.WriteString(file+AppChecksumFileSuffix, checksumCurrent); err != nil {
		return file, false, err
	}
	return file, true, nil
}

func (am *AppManager) Configure(config *cfg.Config) {
	opts := config.Values().App

	if len(opts.SourceExcludes) > 0 {
		am.SourcesIgnored = opts.SourceExcludes
	}
}
