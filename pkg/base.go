package pkg

import (
	"github.com/wttech/aemc/pkg/common/pathx"
)

type BaseOpts struct {
	aem *AEM

	LibDir   string
	TmpDir   string
	ToolDir  string
	CacheDir string
}

func NewBaseOpts(aem *AEM) *BaseOpts {
	cv := aem.config.Values()

	return &BaseOpts{
		aem: aem,

		LibDir:   cv.GetString("base.lib_dir"),
		TmpDir:   cv.GetString("base.tmp_dir"),
		ToolDir:  cv.GetString("base.tool_dir"),
		CacheDir: cv.GetString("base.cache_dir"),
	}
}

func (o *BaseOpts) PrepareWithChanged() (bool, error) {
	changed := false
	dirs := []string{o.LibDir, o.TmpDir, o.ToolDir, o.CacheDir}
	for _, dir := range dirs {
		dirEnsured, err := pathx.EnsureWithChanged(dir)
		changed = changed || dirEnsured
		if err != nil {
			return changed, err
		}
	}
	return changed, nil
}
