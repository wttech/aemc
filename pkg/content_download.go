package pkg

import (
	"embed"
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"io/fs"
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

//go:embed content/aemc_content
var aemcContent embed.FS

func copyEmbedFiles(efs *embed.FS, targetTmpDir string, dirPrefix string) error {
	if err := pathx.DeleteIfExists(targetTmpDir); err != nil {
		return fmt.Errorf("cannot delete temporary dir '%s': %w", targetTmpDir, err)
	}
	return fs.WalkDir(efs, ".", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}
		bytes, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		return filex.Write(targetTmpDir+strings.ReplaceAll(strings.TrimPrefix(path, dirPrefix), "$", ""), bytes)
	})
}

func (c Downloader) DownloadPackage(packageManager *PackageManager, filter string) (string, error) {
	tmpDir := pathx.RandomTemporaryPathName(c.config.BaseOpts.TmpDir, "vault")
	tmpFile := pathx.RandomTemporaryFileName(c.config.BaseOpts.TmpDir, "vault", ".zip")
	tmpResultFile := pathx.RandomTemporaryFileName(c.config.BaseOpts.TmpDir, "vault_result", ".zip")
	defer func() {
		_ = pathx.DeleteIfExists(tmpDir)
		_ = pathx.DeleteIfExists(tmpFile)
	}()
	if err := copyEmbedFiles(&aemcContent, tmpDir, "content/aemc_content"); err != nil {
		return "", err
	}
	if err := filex.Copy(filter, filepath.Join(tmpDir, "META-INF", "vault", "filter.xml"), true); err != nil {
		return "", err
	}
	if err := filex.Archive(tmpDir, tmpFile); err != nil {
		return "", err
	}
	_, err := packageManager.Upload(tmpFile)
	if err != nil {
		return "", err
	}
	if err = packageManager.Build("/etc/packages/my_packages/aemc_content.zip"); err != nil {
		return "", err
	}
	if err = packageManager.Download("/etc/packages/my_packages/aemc_content.zip", tmpResultFile); err != nil {
		return "", err
	}
	return tmpResultFile, nil
}

func (c Downloader) DownloadContent(packageManager *PackageManager, root string, filter string, clean bool) error {
	tmpResultFile, err := c.DownloadPackage(packageManager, filter)
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
