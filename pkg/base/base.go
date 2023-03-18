package base

import (
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/pathx"
)

type Opts struct {
	TmpDir string
}

func NewOpts(config *cfg.Config) *Opts {
	return &Opts{
		TmpDir: config.Values().GetString("base.tmp_dir"),
	}
}

func (o *Opts) Prepare() error {
	return pathx.Ensure(o.TmpDir)
}
