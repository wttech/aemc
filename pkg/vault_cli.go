package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/execx"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"os"
	"path/filepath"
	"strings"
)

func NewVaultCLI(vendorManager *VendorManager) *VaultCLI {
	cv := vendorManager.aem.Config().Values()

	return &VaultCLI{
		vendorManager: vendorManager,

		DownloadURL: cv.GetString("vendor.vault.download_url"),
	}
}

type VaultCLI struct {
	vendorManager *VendorManager

	DownloadURL string
}

type VaultCLILock struct {
	DownloadURL string `yaml:"download_url"`
}

func (v VaultCLI) dir() string {
	if v.vendorManager.aem.Detached() {
		return filepath.Join(os.TempDir(), "vault-cli")
	}
	return filepath.Join(v.vendorManager.aem.baseOpts.ToolDir, "vault-cli")
}

func (v VaultCLI) execDir() string {
	vaultDir, _, _ := strings.Cut(filepath.Base(v.DownloadURL), "-bin")
	return filepath.Join(v.dir(), vaultDir, "bin")
}

func (v VaultCLI) JarFile() string {
	return "<path>/bin/vlt" // TODO c.aem.VaultCLI().JarFile()
}

func (v VaultCLI) lock() osx.Lock[VaultCLILock] {
	return osx.NewLock(v.dir()+"/lock/create.yml", func() (VaultCLILock, error) { return VaultCLILock{DownloadURL: v.DownloadURL}, nil })
}

func (v VaultCLI) PrepareWithChanged() (bool, error) {
	lock := v.lock()
	check, err := lock.State()
	if err != nil {
		return false, err
	}
	if check.UpToDate {
		log.Debugf("existing Vault '%s' is up-to-date", v.DownloadURL)
		return false, nil
	}
	log.Infof("preparing new Vault '%s'", v.DownloadURL)
	err = v.prepare()
	if err != nil {
		return false, err
	}
	err = lock.Lock()
	if err != nil {
		return false, err
	}
	log.Infof("prepared new Vault '%s'", v.DownloadURL)

	return true, nil
}

func (v VaultCLI) archiveFile() string {
	return pathx.Canonical(fmt.Sprintf("%s/%s", v.dir(), filepath.Base(v.DownloadURL)))
}

func (v VaultCLI) prepare() error {
	if err := pathx.DeleteIfExists(v.dir()); err != nil {
		return err
	}
	archiveFile := v.archiveFile()
	log.Infof("downloading Vault from URL '%s' to file '%s'", v.DownloadURL, archiveFile)
	if err := httpx.DownloadOnce(v.DownloadURL, archiveFile); err != nil {
		return err
	}
	log.Infof("downloaded Vault from URL '%s' to file '%s'", v.DownloadURL, archiveFile)

	log.Infof("unarchiving Vault from file '%s'", archiveFile)
	if err := filex.Unarchive(archiveFile, v.dir()); err != nil {
		return err
	}
	log.Infof("unarchived Vault from file '%s'", archiveFile)
	return nil
}

func (v VaultCLI) CommandShell(args []string) error {
	// TODO check if vault is prepared

	cmd := execx.CommandShell(args)
	cmd.Dir = v.execDir() // TODO do not change dir, but prepend with absolute executable; so args should skip 'vlt'
	v.vendorManager.aem.CommandOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot run Vault command: %w", err)
	}
	return nil
}
