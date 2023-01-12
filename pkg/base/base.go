package base

import (
	"github.com/wttech/aemc/pkg/cfg"
)

type Opts struct {
	TmpDir           string
	ChecksumExcludes []string
}

func NewOpts() *Opts {
	return &Opts{
		TmpDir: "aem/home/tmp",
		ChecksumExcludes: []string{
			// meta files
			".*",
			".*/**",
			"**/.*",
			"**/.*/**",

			// build files
			"target/**",
			"**/target",
			"**/target/**",

			"build/**",
			"**/build",
			"**/build/**",

			"dist/**",
			"**/dist",
			"**/dist/**",

			"generated/**",
			"**/generated",
			"**/generated/**",

			"package-lock.json",
			"**/package-lock.json",

			// temporary files
			"*.log",
			"*.tmp",

			"node_modules",
			"**/node_modules",
			"**/node_modules/**",

			"node/**",
			"**/node",
			"**/node/**",
		},
	}
}

func (o Opts) Configure(config *cfg.Config) {
	opts := config.Values().Base

	if len(opts.TmpDir) > 0 {
		o.TmpDir = opts.TmpDir
	}
	if len(opts.ChecksumExcludes) > 0 {
		o.ChecksumExcludes = opts.ChecksumExcludes
	}
}
