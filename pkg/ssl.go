package pkg

import (
	"context"
	"encoding/pem"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/certx"
	"github.com/wttech/aemc/pkg/common/cryptox"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"os"
	"strings"
	"time"
)

const (
	SSLSetupPath = "/libs/granite/security/post/sslSetup.html"
)

type SSL struct {
	instance     *Instance
	setupTimeout time.Duration
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
	configValues := instance.manager.aem.config.Values()

	return &SSL{
		instance:     instance,
		setupTimeout: configValues.GetDuration("instance.ssl.setup_timeout"),
	}
}

func (s SSL) Setup(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort string) (bool, error) {
	if !pathx.Exists(certificateFile) {
		return false, fmt.Errorf("%s > SSL certificate file does not exist: %s", s.instance.IDColor(), certificateFile)
	}
	if !pathx.Exists(privateKeyFile) {
		return false, fmt.Errorf("%s > SSL private key file does not exist: %s", s.instance.IDColor(), privateKeyFile)
	}

	privateKeyData, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return false, fmt.Errorf("%s > failed to read SSL private key file: %w", s.instance.IDColor(), err)
	}
	pemBlock, _ := pem.Decode(privateKeyData)
	if pemBlock != nil {
		tmpDerFileNameBasedOnPemPath, cleanCallback, err := certx.CreateTmpDerFileBasedOnPem(pemBlock)

		defer cleanCallback()

		if err != nil {
			return false, fmt.Errorf("%s > failed to create temp file for storing DER SSL certificate: %w", s.instance.IDColor(), err)
		}

		privateKeyFile = *tmpDerFileNameBasedOnPemPath
	}

	lock := s.lock(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort)
	check, err := lock.State()
	if err != nil {
		return false, err
	}
	if check.UpToDate {
		log.Debugf("%s > SSL already set up (up-to-date)", s.instance.IDColor())
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

	files := map[string]string{
		"certificateFile": certificateFile,
		"privatekeyFile":  privateKeyFile,
	}

	response, err := s.sendSetupRequest(params, files)

	if err != nil {
		return false, fmt.Errorf("%s > failed to setup SSL: %w", s.instance.IDColor(), err)
	} else if response.IsError() {
		rawBody := response.RawBody()
		if rawBody == nil {
			return false, fmt.Errorf("%s > failed to setup SSL: %s", s.instance.IDColor(), response.Status())
		}
		defer rawBody.Close()
		body, err := io.ReadAll(rawBody)
		if err != nil {
			return false, fmt.Errorf("%s > failed to setup SSL: %s, %w", s.instance.IDColor(), response.Status(), err)
		}
		errorMessage, err := s.extractErrorMessage(string(body[:]))
		if err != nil {
			return false, fmt.Errorf("%s > failed to setup SSL: %s, %w", s.instance.IDColor(), response.Status(), err)
		}
		return false, fmt.Errorf("%s > failed to setup SSL: %s, %s", s.instance.IDColor(), response.Status(), errorMessage)
	}

	if err := lock.Lock(); err != nil {
		return true, fmt.Errorf("%s > failed to lock SSL setup: %w", s.instance.IDColor(), err)
	}

	return true, nil
}

func (s SSL) sendSetupRequest(params map[string]any, files map[string]string) (*resty.Response, error) {
	pause := time.Duration(2) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), s.setupTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			response, err := s.instance.http.
				RequestFormData(params).
				SetFiles(files).
				SetDoNotParseResponse(true).
				Post(SSLSetupPath)
			if err == nil {
				return response, err
			}
			log.Warnf("%s > failed to setup SSL: %s, retrying", s.instance.IDColor(), err)
			time.Sleep(pause)
		}
	}
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
func (s SSL) extractErrorMessage(body string) (errorMessage string, err error) {
	lines := strings.Split(string(body), "\n")
	errorMessageAhead := false
	for _, line := range lines {
		if strings.Contains(line, "foundation-form-response-status-message") {
			errorMessageAhead = true
			continue
		}
		if errorMessageAhead {
			// Extract error message from within <dd> tag:
			// <dd>Error message example</dd>
			line = strings.Split(line, ">")[1]
			// Error message example</dd
			line = strings.Split(line, "<")[0]
			// Error message example
			errorMessage = strings.TrimSpace(line)
			break
		}
	}
	if errorMessage == "" {
		return "", fmt.Errorf("%s > SSL error message not found", s.instance.IDColor())
	}
	return errorMessage, nil
}

func (s SSL) lock(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort string) osx.Lock[sslLock] {
	return osx.NewLock(fmt.Sprintf("%s/ssl.yml", s.instance.LockDir()), func() (sslLock, error) {
		certificateChecksum, err := filex.ChecksumFile(certificateFile)
		if err != nil {
			return sslLock{}, fmt.Errorf("%s > failed to calculate checksum for SSL certificate file: %w", s.instance.IDColor(), err)
		}
		privateKeyChecksum, err := filex.ChecksumFile(privateKeyFile)
		if err != nil {
			return sslLock{}, fmt.Errorf("%s > failed to calculate checksum for SSL private key file: %w", s.instance.IDColor(), err)
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
