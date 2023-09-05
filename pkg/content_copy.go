package pkg

import (
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"path/filepath"
)

type Copier struct {
	config *content.Opts
}

func NewCopier(config *content.Opts) *Copier {
	return &Copier{
		config: config,
	}
}

func (c Copier) Copy(scrPackageManager *PackageManager, destPackageManager *PackageManager, pid string, roots []string, filter string, clean bool) error {
	var tmpResultFile string
	if clean {
		tmpResultFile = pathx.RandomTemporaryFileName(c.config.BaseOpts.TmpDir, "vault_result", ".zip")
		tmpResultDir := pathx.RandomTemporaryPathName(c.config.BaseOpts.TmpDir, "vault_result")
		defer func() {
			_ = pathx.DeleteIfExists(tmpResultDir)
			_ = pathx.DeleteIfExists(tmpResultFile)
		}()
		if err := NewDownloader(c.config).DownloadContent(scrPackageManager, filepath.Join(tmpResultDir, content.JCRRoot), "", roots, filter, true, true); err != nil {
			return err
		}
		if err := filex.Archive(tmpResultDir, tmpResultFile); err != nil {
			return err
		}
	} else {
		var err error
		tmpResultFile, err = NewDownloader(c.config).DownloadPackage(scrPackageManager, pid, roots, filter)
		if err != nil {
			return err
		}
	}
	defer func() { _ = pathx.DeleteIfExists(tmpResultFile) }()
	remotePath, err := destPackageManager.Upload(tmpResultFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = destPackageManager.Delete(remotePath)
	}()
	if err = destPackageManager.Install(remotePath); err != nil {
		return err
	}
	return nil
}
