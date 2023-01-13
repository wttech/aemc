package base

import (
	"github.com/wttech/aemc/pkg/cfg"
)

type Opts struct {
	TmpDir string
}

func NewOpts() *Opts {
	return &Opts{
		TmpDir: "aem/home/tmp",
	}
}

func (o *Opts) Configure(config *cfg.Config) {
	opts := config.Values().Base

	if len(opts.TmpDir) > 0 {
		o.TmpDir = opts.TmpDir
	}
}
