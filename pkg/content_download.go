package pkg

import (
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"path/filepath"
	"strings"
	"time"
)

const (
	FilterXML = "filter.xml"
)

type Downloader struct {
	config *content.Opts
}

func NewDownloader(config *content.Opts) *Downloader {
	return &Downloader{config}
}

func (c Downloader) DownloadPackage(packageManager *PackageManager, pid string, roots []string, filter string) (string, error) {
	if pid == "" {
		pid = "my_packages:aemc_content:" + time.Now().Format("2006.102.304") + "-SNAPSHOT"
	}
	remotePath, err := packageManager.Create(pid, roots, filter)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = packageManager.Delete(remotePath)
	}()
	if err := packageManager.Build(remotePath); err != nil {
		return "", err
	}
	tmpResultFile := filepath.Join(c.config.BaseOpts.TmpDir, filepath.Base(remotePath))
	if err := packageManager.Download(remotePath, tmpResultFile); err != nil {
		return "", err
	}
	return tmpResultFile, nil
}

func (c Downloader) DownloadContent(packageManager *PackageManager, pid string, root string, roots []string, filter string, clean bool, unpack bool) error {
	tmpResultFile, err := c.DownloadPackage(packageManager, pid, roots, filter)
	if err != nil {
		return err
	}
	tmpResultDir := pathx.RandomTemporaryPathName(c.config.BaseOpts.TmpDir, "download_content")
	defer func() {
		_ = pathx.DeleteIfExists(tmpResultDir)
		_ = pathx.DeleteIfExists(tmpResultFile)
	}()
	if unpack {
		if err = filex.Unarchive(tmpResultFile, tmpResultDir); err != nil {
			return err
		}
		if err := pathx.Ensure(root); err != nil {
			return err
		}
		before, _, _ := strings.Cut(root, content.JCRRoot)
		if clean {
			if err = content.NewCleaner(c.config).BeforeClean(root); err != nil {
				return err
			}
		}
		if err = filex.CopyDir(filepath.Join(tmpResultDir, content.JCRRoot), before+content.JCRRoot); err != nil {
			return err
		}
		if clean {
			if err = content.NewCleaner(c.config).Clean(root); err != nil {
				return err
			}
		}
	} else {
		if err = filex.Copy(tmpResultFile, filepath.Join(root, filepath.Base(tmpResultFile)), true); err != nil {
			return err
		}
	}
	return nil
}
