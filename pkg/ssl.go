package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/cryptox"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"golang.org/x/net/html"
	"io"
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
		// If the response is an error, we try to read the error message from the response body.
		// The resty library does not provide a way to read the response body if the response is an error.
		body, err := io.ReadAll(response.RawBody())
		if err != nil {
			return false, fmt.Errorf("%s > failed to setup SSL: %s, %w", s.instance.ID(), response.Status(), err)
		}
		errorMessage, err := findErrorMessage(string(body[:]))
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
func findErrorMessage(body string) (errorMessage string, err error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return "", err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "dt" && hasClass(n, "foundation-form-response-status-message") {
			// If the <dt> node with class 'foundation-form-response-status-message' is found,
			// we search for the corresponding <dd> node containing the error message.
			ddNode := n.NextSibling
			for ddNode != nil {
				if ddNode.Type == html.ElementNode && ddNode.Data == "dd" {
					errorMessage = ddNode.FirstChild.Data
					return
				}
				ddNode = ddNode.NextSibling
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return errorMessage, nil
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			return strings.Contains(attr.Val, class)
		}
	}
	return false
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
