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

func NewVaultCli(aem *AEM) *VaultCli {
	cv := aem.baseOpts.Config().Values()

	return &VaultCli{
		aem: aem,

		DownloadURL: cv.GetString("vault.download_url"),
	}
}

type VaultCli struct {
	aem *AEM

	DownloadURL string
}

type VaultCliLock struct {
	DownloadURL string `yaml:"download_url"`
}

func (v VaultCli) dir() string {
	if v.aem.Detached() {
		return filepath.Join(os.TempDir(), "vault-cli")
	}
	return filepath.Join(v.aem.baseOpts.ToolDir, "vault-cli")
}

func (v VaultCli) execFile() string {
	vaultDir, _, _ := strings.Cut(filepath.Base(v.DownloadURL), "-bin")
	execDir := filepath.Join(v.dir(), vaultDir, "bin")
	if osx.IsWindows() {
		return pathx.Canonical(execDir + "/vlt.bat")
	}
	return pathx.Canonical(execDir + "/vlt")
}

func (v VaultCli) lock() osx.Lock[VaultCliLock] {
	return osx.NewLock(v.dir()+"/lock/create.yml", func() (VaultCliLock, error) { return VaultCliLock{DownloadURL: v.DownloadURL}, nil })
}

func (v VaultCli) Prepare() error {
	lock := v.lock()
	check, err := lock.State()
	if err != nil {
		return err
	}
	if check.UpToDate {
		log.Debugf("existing Vault '%s' is up-to-date", v.DownloadURL)
		return nil
	}
	log.Infof("preparing new Vault '%s'", v.DownloadURL)
	if err = v.prepare(); err != nil {
		return err
	}
	if err = lock.Lock(); err != nil {
		return err
	}
	log.Infof("prepared new Vault '%s'", v.DownloadURL)

	return nil
}

func (v VaultCli) archiveFile() string {
	return pathx.Canonical(fmt.Sprintf("%s/%s", v.dir(), filepath.Base(v.DownloadURL)))
}

func (v VaultCli) prepare() error {
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

func (v VaultCli) CommandShell(args []string) error {
	if err := v.Prepare(); err != nil {
		return fmt.Errorf("cannot prepare Vault before running command: %w", err)
	}
	vaultCliArgs := append([]string{v.execFile()}, args...)
	cmd := execx.CommandShell(vaultCliArgs)
	v.aem.CommandOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot run Vault command: %w", err)
	}
	return nil
}
