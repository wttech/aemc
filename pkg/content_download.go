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
	if err := pathx.DeleteIfExists(targetTmpDir); err != nil {
		return fmt.Errorf("cannot delete temporary dir '%s': %w", targetTmpDir, err)
	}
	return fs.WalkDir(efs, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if dirEntry.IsDir() {
			return nil
		}
		bytes, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		if err := filex.Write(targetTmpDir+strings.TrimPrefix(path, dirPrefix), bytes); err != nil {
			return err
		}
		return nil
	})
}

func (c Downloader) Download(packageManager *PackageManager, root string, filter string, unpack bool, clean bool) error {
	err := copyEmbedFiles(&aemcContent, "/tmp/aemc_content", "content/aemc_content")
	if err != nil {
		return err
	}
	err = filex.Copy(filter, "/tmp/aemc_content/META-INF/vault/filter.xml", true)
	if err != nil {
		return err
	}
	_, err = filex.ArchiveWithChanged("/tmp/aemc_content", "/tmp/aemc_content.zip")
	if err != nil {
		return err
	}
	_, err = packageManager.Upload("/tmp/aemc_content.zip")
	if err != nil {
		return err
	}
	err = packageManager.Build("/etc/packages/my_packages/aemc_content.zip")
	if err != nil {
		return err
	}
	err = packageManager.Download("/etc/packages/my_packages/aemc_content.zip", "/tmp/aemc_content.zip")
	if err != nil {
		return err
	}
	if unpack {
		_, err = filex.UnarchiveWithChanged("/tmp/aemc_content.zip", "/tmp/aemc_content")
		if err != nil {
			return err
		}
		err = filex.Copy("/tmp/aemc_content/jcr_root", root, true)
		if err != nil {
			return err
		}
	}
	if unpack && clean {
		err = content.NewCleaner(c.config).Clean(root)
		if err != nil {
			return err
		}
	}
	return nil
}
