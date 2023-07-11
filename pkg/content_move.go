package pkg

import (
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/content"
)

type Mover struct {
	config *content.Opts
}

func NewMover(config *content.Opts) *Mover {
	return &Mover{
		config: config,
	}
}

func (c Mover) Move(scrPackageManager *PackageManager, descPackageManager *PackageManager, filter string, clean bool) error {
	err := NewDownloader(c.config).Download(scrPackageManager, "/tmp/aemc_content", filter, clean, clean)
	if err != nil {
		return err
	}
	if clean {
		_, err = filex.ArchiveWithChanged("/tmp/aemc_content", "/tmp/aemc_content.zip")
		if err != nil {
			return err
		}
	}
	_, err = descPackageManager.Upload("/tmp/aemc_content.zip")
	if err != nil {
		return err
	}
	err = descPackageManager.Install("/etc/packages/my_packages/aemc_content.zip")
	if err != nil {
		return err
	}
	return nil
}
