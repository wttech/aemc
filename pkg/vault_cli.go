package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/execx"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"net/url"
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

func (v VaultCli) execDir() string {
	vaultDir, _, _ := strings.Cut(filepath.Base(v.DownloadURL), "-bin")
	return filepath.Join(v.dir(), vaultDir, "bin")
}

func (v VaultCli) lock() osx.Lock[VaultCliLock] {
	return osx.NewLock(v.dir()+"/lock/create.yml", func() (VaultCliLock, error) { return VaultCliLock{DownloadURL: v.DownloadURL}, nil })
}

func (v VaultCli) prepare() error {
	lock := v.lock()
	check, err := lock.State()
	if err != nil {
		return err
	}
	if check.UpToDate {
		log.Debugf("existing Vault-Cli '%s' is up-to-date", v.DownloadURL)
		return nil
	}
	log.Infof("preparing new Vault-Cli '%s'", v.DownloadURL)
	err = v.install()
	if err != nil {
		return err
	}
	err = lock.Lock()
	if err != nil {
		return err
	}
	log.Infof("prepared new Vault-Cli '%s'", v.DownloadURL)

	return nil
}

func (v VaultCli) archiveFile() string {
	return pathx.Canonical(fmt.Sprintf("%s/%s", v.dir(), filepath.Base(v.DownloadURL)))
}

func (v VaultCli) install() error {
	archiveFile := v.archiveFile()
	log.Infof("downloading Vault-Cli from URL '%s' to file '%s'", v.DownloadURL, archiveFile)
	if err := httpx.DownloadOnce(v.DownloadURL, archiveFile); err != nil {
		return err
	}
	log.Infof("downloaded Vault-Cli from URL '%s' to file '%s'", v.DownloadURL, archiveFile)

	log.Infof("unarchiving Vault-Cli from file '%s'", archiveFile)
	if err := filex.Unarchive(archiveFile, v.dir()); err != nil {
		return err
	}
	log.Infof("unarchived Vault-Cli from file '%s'", archiveFile)
	return nil
}

func (v VaultCli) CommandShell(args []string) error {
	if err := v.prepare(); err != nil {
		return fmt.Errorf("cannot run Vault-Cli command: %w", err)
	}
	cmd := execx.CommandShell(args)
	cmd.Dir = v.execDir()
	v.aem.CommandOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("cannot run Vault-Cli command: %w", err)
	}
	return nil
}

func (v VaultCli) PushContent(instance *Instance, mainDir string, jcrPath string) error {
	vaultCliArgs := []string{
		"vlt",
		"--credentials", fmt.Sprintf("%s:%s", instance.user, instance.password),
		"import",
		fmt.Sprintf("%s/crx/-/jcr:root%s", instance.http.baseURL, jcrPath),
		mainDir,
	}
	return v.CommandShell(vaultCliArgs)
}

func (v VaultCli) PullContent(instance *Instance, mainDir string, filterFile string) error {
	vaultCliArgs := []string{
		"vlt",
		"--credentials", fmt.Sprintf("%s:%s", instance.user, instance.password),
		"checkout",
		"--force",
		"--filter", filterFile,
		fmt.Sprintf("%s/crx/server/crx.default", instance.http.baseURL),
		mainDir,
	}
	return v.CommandShell(vaultCliArgs)
}

func (v VaultCli) CopyContent(srcInstance *Instance, destInstance *Instance, rcpArgs []string, jcrPath string) error {
	parsedURLSrc, err := url.Parse(srcInstance.http.baseURL)
	if err != nil {
		return err
	}
	parsedURLDest, err := url.Parse(destInstance.http.baseURL)
	if err != nil {
		return err
	}
	vaultCliArgs := []string{"vlt", "rcp"}
	vaultCliArgs = append(vaultCliArgs, rcpArgs...)
	vaultCliArgs = append(vaultCliArgs, []string{
		fmt.Sprintf("%s://%s:%s@%s/crx/-/jcr:root%s",
			parsedURLSrc.Scheme, srcInstance.user, srcInstance.password,
			parsedURLSrc.Host, jcrPath),
		fmt.Sprintf("%s://%s:%s@%s/crx/-/jcr:root%s",
			parsedURLDest.Scheme, destInstance.user, destInstance.password,
			parsedURLDest.Host, jcrPath),
	}...)
	return v.CommandShell(vaultCliArgs)
}
