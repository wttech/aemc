package base

import (
	"github.com/wttech/aemc/pkg/cfg"
)

type Opts struct {
	TmpDir                 string
	ChecksumIgnorePatterns []string
}

func NewOpts() *Opts {
	return &Opts{
		TmpDir: "aem/home/tmp",
		ChecksumIgnorePatterns: []string{
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
			"**/aem/home/**",
		},
	}
}

func (o Opts) Configure(config *cfg.Config) {
	opts := config.Values().Base

	if len(opts.TmpDir) > 0 {
		o.TmpDir = opts.TmpDir
	}
	if len(opts.ChecksumExcludes) > 0 {
		o.ChecksumIgnorePatterns = opts.ChecksumExcludes
	}
}
