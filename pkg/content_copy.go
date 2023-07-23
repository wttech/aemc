package pkg

import (
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/content"
)

type Copier struct {
	config *content.Opts
}

func NewCopier(config *content.Opts) *Copier {
	return &Copier{
		config: config,
	}
}

func (c Copier) Copy(scrPackageManager *PackageManager, destPackageManager *PackageManager, filter string, clean bool) error {
	if err := NewDownloader(c.config).Download(scrPackageManager, "/tmp/aemc_content", filter, clean); err != nil {
		return err
	}
	if clean {
		if err := filex.Archive("/tmp/aemc_content", "/tmp/aemc_content.zip"); err != nil {
			return err
		}
	}
	_, err := destPackageManager.Upload("/tmp/aemc_content.zip")
	if err != nil {
		return err
	}
	if err = destPackageManager.Install("/etc/packages/my_packages/aemc_content.zip"); err != nil {
		return err
	}
	return nil
}
