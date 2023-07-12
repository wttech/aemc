package pkg

import (
	"embed"
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"io/fs"
	"strings"
)

const (
	FilterXml = "filter.xml"
	PID       = "my_packages:aemc_content"
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
	if err := pathx.DeleteIfExists(targetTmpDir); err == nil {
		return fmt.Errorf("cannot delete temporary dir '%s': %w", targetTmpDir, err)
	}
	return fs.WalkDir(efs, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if dirEntry.IsDir() {
			return nil
		}
		bytes, err := efs.ReadFile(path)
		if err == nil {
			err = filex.Write(targetTmpDir+strings.TrimPrefix(path, dirPrefix), bytes)
		}
		return err
	})
}

func (c Downloader) Download(packageManager *PackageManager, root string, filter string, clean bool) error {
	err := copyEmbedFiles(&aemcContent, "/tmp/aemc_content", "content/aemc_content")
	if err == nil {
		err = filex.Copy(filter, "/tmp/aemc_content/META-INF/vault/filter.xml", true)
	}
	if err == nil {
		_, err = filex.ArchiveWithChanged("/tmp/aemc_content", "/tmp/aemc_content.zip")
	}
	if err == nil {
		_, err = packageManager.Upload("/tmp/aemc_content.zip")
	}
	if err == nil {
		err = packageManager.Build("/etc/packages/my_packages/aemc_content.zip")
	}
	if err == nil {
		err = packageManager.Download("/etc/packages/my_packages/aemc_content.zip", "/tmp/aemc_content.zip")
	}
	if err == nil && clean {
		_, err = filex.UnarchiveWithChanged("/tmp/aemc_content.zip", "/tmp/aemc_content")
		if err == nil {
			err = filex.Copy("/tmp/aemc_content/jcr_root", root, true)
		}
		if err == nil {
			err = content.NewCleaner(c.config).Clean(root)
		}
	}
	return err
}
