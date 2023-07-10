package pkg

import (
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

func (c Mover) Move(scrPackageManager *PackageManager, descPackageManager *PackageManager, filter string) error {
	err := NewDownloader(c.config).Download(scrPackageManager, "", filter, false, false)
	if err != nil {
		return err
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
