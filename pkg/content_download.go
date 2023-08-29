package pkg

import (
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"path/filepath"
	"strings"
)

const (
	FilterXml = "filter.xml"
)

type Downloader struct {
	config *content.Opts
}

func NewDownloader(config *content.Opts) *Downloader {
	return &Downloader{
		config: config,
	}
}

func (c Downloader) DownloadPackage(packageManager *PackageManager, roots []string, filter string) (string, error) {
	tmpResultFile := pathx.RandomTemporaryFileName(c.config.BaseOpts.TmpDir, "vault_result", ".zip")
	remotePath, err := packageManager.Create("my_packages:aemc_content", roots, filter)
	if err != nil {
		return "", err
	}
	if err := packageManager.Build(remotePath); err != nil {
		return "", err
	}
	if err := packageManager.Download(remotePath, tmpResultFile); err != nil {
		return "", err
	}
	return tmpResultFile, nil
}

func (c Downloader) DownloadContent(packageManager *PackageManager, root string, roots []string, filter string, clean bool) error {
	tmpResultFile, err := c.DownloadPackage(packageManager, roots, filter)
	if err != nil {
		return err
	}
	tmpResultDir := pathx.RandomTemporaryPathName(c.config.BaseOpts.TmpDir, "vault_result")
	defer func() {
		_ = pathx.DeleteIfExists(tmpResultDir)
		_ = pathx.DeleteIfExists(tmpResultFile)
	}()
	if err = filex.Unarchive(tmpResultFile, tmpResultDir); err != nil {
		return err
	}
	before, _, _ := strings.Cut(root, content.JcrRoot)
	if clean {
		if err = content.NewCleaner(c.config).BeforeClean(root); err != nil {
			return err
		}
	}
	if err = filex.CopyDir(filepath.Join(tmpResultDir, content.JcrRoot), before+content.JcrRoot); err != nil {
		return err
	}
	if clean {
		if err = content.NewCleaner(c.config).Clean(root); err != nil {
			return err
		}
	}
	return nil
}
