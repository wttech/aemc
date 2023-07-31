package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/cryptox"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
)

const (
	SSLSetupPath = "/libs/granite/security/post/sslSetup.html"
)

type SSL struct {
	instance *Instance
}

type sslLock struct {
	KeystorePassword   string `yaml:"keystore_password"`
	TrustStorePassword string `yaml:"trust_store_password"`
	Certificate        string `yaml:"certificate"`
	PrivateKey         string `yaml:"private_key"`
	HttpsHostname      string `yaml:"https_hostname"`
	HttpsPort          string `yaml:"https_port"`
}

func NewSSL(instance *Instance) *SSL {
	return &SSL{
		instance: instance,
	}
}

func (s SSL) Setup(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort string) (bool, error) {
	if !pathx.Exists(certificateFile) {
		return false, fmt.Errorf("%s > certificate file does not exist: %s", s.instance.ID(), certificateFile)
	}
	if !pathx.Exists(privateKeyFile) {
		return false, fmt.Errorf("%s > private key file does not exist: %s", s.instance.ID(), privateKeyFile)
	}

	lock := s.lock(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort)
	check, err := lock.State()
	if err != nil {
		return false, fmt.Errorf("%s > failed to check SSL setup: %w", s.instance.ID(), err)
	}
	if check.UpToDate {
		log.Debugf("%s > SSL already set up (up-to-date)", s.instance.ID())
		return false, nil
	}

	params := map[string]any{
		"keystorePassword":          keyStorePassword,
		"keystorePasswordConfirm":   keyStorePassword,
		"truststorePassword":        trustStorePassword,
		"truststorePasswordConfirm": trustStorePassword,
		"httpsHostname":             httpsHostname,
		"httpsPort":                 httpsPort,
	}

	response, err := s.instance.http.
		RequestFormData(params).
		SetFiles(map[string]string{
			"certificateFile": certificateFile,
			"privatekeyFile":  privateKeyFile,
		}).
		Post(SSLSetupPath)

	if err != nil {
		return false, fmt.Errorf("%s > failed to setup SSL: %w", s.instance.ID(), err)
	} else if response.IsError() {
		return false, fmt.Errorf("%s > failed to setup SSL: %s", s.instance.ID(), response.Status())
	}

	if err := lock.Lock(); err != nil {
		return true, fmt.Errorf("%s > failed to lock SSL setup: %w", s.instance.ID(), err)
	}

	return true, nil
}

func (s SSL) lock(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort string) osx.Lock[sslLock] {
	return osx.NewLock(fmt.Sprintf("%s/ssl.yml", s.instance.local.LockDir()), func() (sslLock, error) {
		certificateChecksum, err := filex.ChecksumFile(certificateFile)
		if err != nil {
			return sslLock{}, fmt.Errorf("%s > failed to calculate checksum for certificate file: %w", s.instance.ID(), err)
		}
		privateKeyChecksum, err := filex.ChecksumFile(privateKeyFile)
		if err != nil {
			return sslLock{}, fmt.Errorf("%s > failed to calculate checksum for private key file: %w", s.instance.ID(), err)
		}

		return sslLock{
			KeystorePassword:   cryptox.HashString(keyStorePassword),
			TrustStorePassword: cryptox.HashString(trustStorePassword),
			Certificate:        certificateChecksum,
			PrivateKey:         privateKeyChecksum,
			HttpsHostname:      httpsHostname,
			HttpsPort:          httpsPort,
		}, nil
	})
}
