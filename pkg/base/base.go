package base

import (
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/pathx"
)

type Opts struct {
	config *cfg.Config

	TmpDir   string
	ToolDir  string
	CacheDir string
}

func NewOpts(config *cfg.Config) *Opts {
	cv := config.Values()

	return &Opts{
		config: config,

		TmpDir:   cv.GetString("base.tmp_dir"),
		ToolDir:  cv.GetString("base.tool_dir"),
		CacheDir: cv.GetString("base.cache_dir"),
	}
}

func (o *Opts) Config() *cfg.Config {
	return o.config
}

func (o *Opts) Prepare() (bool, error) {
	changed := false
	dirs := []string{o.TmpDir, o.ToolDir, o.CacheDir}
	for _, dir := range dirs {
		dirEnsured, err := pathx.EnsureWithChanged(dir)
		changed = changed || dirEnsured
		if err != nil {
			return changed, err
		}
	}
	return changed, nil
}
