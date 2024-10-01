package pkg

import (
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	FlattenFilePattern = "[\\\\/]_[a-zA-Z0-9]+_[^\\\\/]+\\.xml$"
)

type ContentManager struct {
	aem            *AEM
	contentManager *content.Manager
}

func NewContentManager(aem *AEM) *ContentManager {
	return &ContentManager{
		aem:            aem,
		contentManager: content.NewManager(aem.baseOpts),
	}
}

func (cm *ContentManager) tmpDir() string {
	if cm.aem.Detached() {
		return os.TempDir()
	}
	return cm.aem.baseOpts.TmpDir
}

func (cm *ContentManager) Clean(path string) error {
	return cm.contentManager.Clean(path)
}

func (cm *ContentManager) Download(instance *Instance, localFile string, clean bool, opts PackageCreateOpts) error {
	if clean {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
		defer func() { _ = pathx.DeleteIfExists(workDir) }()
		if err := cm.pullContent(instance, workDir, opts); err != nil {
			return err
		}
		if err := cm.Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
		if err := content.Zip(workDir, localFile); err != nil {
			return err
		}
	} else {
		if err := cm.downloadContent(instance, localFile, opts); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullDir(instance *Instance, dir string, clean bool, replace bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.pullContent(instance, workDir, opts); err != nil {
		return err
	}
	if replace {
		if err := cm.contentManager.DeleteDir(dir); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(dir, content.JCRRoot)
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot, jcrPath), dir); err != nil {
		return err
	}
	if clean {
		if err := cm.Clean(dir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullFile(instance *Instance, file string, clean bool, replace bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.pullContent(instance, workDir, opts); err != nil {
		return err
	}
	syncFile := DetermineSyncFile(file)
	if file != syncFile || replace {
		if err := cm.contentManager.DeleteFile(file, nil); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(syncFile, content.JCRRoot)
	if err := filex.Copy(filepath.Join(workDir, content.JCRRoot, jcrPath), syncFile, true); err != nil {
		return err
	}
	if clean {
		if err := cm.Clean(syncFile); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) Push(instances []Instance, clean bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_push")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := copyPackageAllFiles(workDir, opts); err != nil {
		return err
	}
	if clean {
		if err := cm.Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
	}
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_push", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if err := content.Zip(workDir, pkgFile); err != nil {
		return err
	}
	if err := cm.pushContent(instances, pkgFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) Copy(srcInstance *Instance, destInstances []Instance, clean bool, opts PackageCreateOpts) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_copy", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if err := cm.Download(srcInstance, pkgFile, clean, opts); err != nil {
		return err
	}
	if err := cm.pushContent(destInstances, pkgFile); err != nil {
		return err
	}
	return nil
}

func DetermineSyncFile(file string) string {
	if regexp.MustCompile(FlattenFilePattern).MatchString(file) {
		return filepath.Join(strings.ReplaceAll(file, content.XmlFileSuffix, ""), content.JCRContentFile)
	}
	return file
}

func (cm *ContentManager) downloadContent(instance *Instance, pkgFile string, opts PackageCreateOpts) error {
	remotePath, err := instance.PackageManager().Create(opts)
	defer func() { _ = instance.PackageManager().Delete(remotePath) }()
	if err != nil {
		return err
	}
	if err = instance.PackageManager().Build(remotePath); err != nil {
		return err
	}
	if err = instance.PackageManager().Download(remotePath, pkgFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) pullContent(instance *Instance, workDir string, opts PackageCreateOpts) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if err := cm.downloadContent(instance, pkgFile, opts); err != nil {
		return err
	}
	if err := content.Unzip(pkgFile, workDir); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) pushContent(instances []Instance, pkgFile string) error {
	_, err := InstanceProcess(cm.aem, instances, func(instance Instance) (any, error) {
		remotePath, err := instance.PackageManager().Upload(pkgFile)
		defer func() { _ = instance.PackageManager().Delete(remotePath) }()
		if err != nil {
			return nil, err
		}
		if err = instance.PackageManager().Install(remotePath); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
