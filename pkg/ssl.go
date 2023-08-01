package pkg

import (
	"encoding/pem"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/cryptox"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"os"
	"strings"
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
	return &SSL{instance: instance}
}

func (s SSL) Setup(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort string) (bool, error) {
	if !pathx.Exists(certificateFile) {
		return false, fmt.Errorf("%s > SSL certificate file does not exist: %s", s.instance.ID(), certificateFile)
	}
	if !pathx.Exists(privateKeyFile) {
		return false, fmt.Errorf("%s > SSL private key file does not exist: %s", s.instance.ID(), privateKeyFile)
	}

	privateKeyData, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return false, fmt.Errorf("%s > failed to read SSL private key file: %w", s.instance.ID(), err)
	}
	pemBlock, _ := pem.Decode(privateKeyData)
	if pemBlock != nil {
		tempDerFile, err := os.CreateTemp("", "aemc-private-key-*.der")
		if err != nil {
			return false, fmt.Errorf("%s > failed to create temp file for storing DER SSL certificate: %w", s.instance.ID(), err)
		}
		defer os.Remove(tempDerFile.Name())
		err = s.writeDER(tempDerFile, pemBlock)
		if err != nil {
			return false, fmt.Errorf("%s > failed to write DER SSL certificate: %w", s.instance.ID(), err)
		}
		privateKeyFile = tempDerFile.Name()
	}

	lock := s.lock(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort)
	check, err := lock.State()
	if err != nil {
		return false, err
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
		SetDoNotParseResponse(true).
		Post(SSLSetupPath)

	if err != nil {
		return false, fmt.Errorf("%s > failed to setup SSL: %w", s.instance.ID(), err)
	} else if response.IsError() {
		rawBody := response.RawBody()
		if rawBody == nil {
			return false, fmt.Errorf("%s > failed to setup SSL: %s", s.instance.ID(), response.Status())
		}
		defer rawBody.Close()
		body, err := io.ReadAll(rawBody)
		if err != nil {
			return false, fmt.Errorf("%s > failed to setup SSL: %s, %w", s.instance.ID(), response.Status(), err)
		}
		errorMessage, err := s.getErrorMessage(string(body[:]))
		if err != nil {
			return false, fmt.Errorf("%s > failed to setup SSL: %s, %w", s.instance.ID(), response.Status(), err)
		}
		return false, fmt.Errorf("%s > failed to setup SSL: %s, %s", s.instance.ID(), response.Status(), errorMessage)
	}

	if err := lock.Lock(); err != nil {
		return true, fmt.Errorf("%s > failed to lock SSL setup: %w", s.instance.ID(), err)
	}

	return true, nil
}

func (s SSL) writeDER(tempDerFile *os.File, pemBlock *pem.Block) error {
	if _, err := tempDerFile.Write(pemBlock.Bytes); err != nil {
		return err
	}
	err := tempDerFile.Close()
	if err != nil {
		return err
	}
	return nil
}

// From HTML response body, e.g.:
// <!DOCTYPE html>
// <html lang='en'>
// <head>
// <title>Error</title>
// </head>
// <body>
// <h1>Error</h1>
// <dl>
// <dt class='foundation-form-response-status-code'>Status</dt>
// <dd>500</dd>
// <dt class='foundation-form-response-status-message'>Message</dt>
// <dd>Invalid password for existing key store</dd>
// <dt class='foundation-form-response-title'>Title</dt>
// <dd>Error</dd>
// </dl>
// </body>
// </html>
// returns "Invalid password for existing key store"
func (s SSL) getErrorMessage(body string) (errorMessage string, err error) {
	lines := strings.Split(string(body), "\n")
	errorMessageAhead := false
	for _, line := range lines {
		if strings.Contains(line, "foundation-form-response-status-message") {
			errorMessageAhead = true
			continue
		}
		if errorMessageAhead {
			line = strings.Split(line, ">")[1]
			line = strings.Split(line, "<")[0]
			errorMessage = strings.TrimSpace(line)
			break
		}
	}
	if errorMessage == "" {
		return "", fmt.Errorf("error message not found")
	}
	return errorMessage, nil
}

func (s SSL) lock(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort string) osx.Lock[sslLock] {
	return osx.NewLock(fmt.Sprintf("%s/ssl.yml", s.instance.local.LockDir()), func() (sslLock, error) {
		certificateChecksum, err := filex.ChecksumFile(certificateFile)
		if err != nil {
			return sslLock{}, fmt.Errorf("%s > failed to calculate checksum for SSL certificate file: %w", s.instance.ID(), err)
		}
		privateKeyChecksum, err := filex.ChecksumFile(privateKeyFile)
		if err != nil {
			return sslLock{}, fmt.Errorf("%s > failed to calculate checksum for SSL private key file: %w", s.instance.ID(), err)
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
