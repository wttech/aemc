package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
)

const (
	CryptoProtectPath = "/system/console/crypto/.json"
)

type Crypto struct {
	instance *Instance

	keyBundleSymbolicName string
}

type CryptoProtectResult struct {
	Protected string
}

func NewCrypto(instance *Instance) *Crypto {
	cv := instance.manager.aem.config.Values()

	return &Crypto{
		instance: instance,

		keyBundleSymbolicName: cv.GetString("instance.crypto.key_bundle_symbolic_name"),
	}
}

func (c Crypto) Setup(hmacFile string, masterFile string) (bool, error) {
	if !c.instance.IsLocal() {
		return false, fmt.Errorf("%s > Crypto keys could only be set on local instance", c.instance.IDColor())
	}
	if !pathx.Exists(hmacFile) {
		return false, fmt.Errorf("%s > Crypto hmac file '%s' does not exist", c.instance.IDColor(), hmacFile)
	}
	if !pathx.Exists(masterFile) {
		return false, fmt.Errorf("%s > Crypto master file '%s' does not exist", c.instance.IDColor(), masterFile)
	}
	osgi := c.instance.OSGI()
	keyBundle, err := osgi.BundleManager().Find(c.keyBundleSymbolicName)
	if err != nil {
		return false, err
	}
	if keyBundle == nil {
		return false, fmt.Errorf("%s > cannot find Crypto key bundle using symbolic name '%s'", c.instance.IDColor(), c.keyBundleSymbolicName)
	}
	keyDir := fmt.Sprintf("%s/data", c.instance.Local().BundleDir(keyBundle.ID))
	hmacTargetFile := fmt.Sprintf("%s/hmac", keyDir)
	masterTargetFile := fmt.Sprintf("%s/master", keyDir)

	hmacOk, err := filex.Equals(hmacFile, hmacTargetFile)
	if err != nil {
		return false, err
	}
	masterOk, err := filex.Equals(masterFile, masterTargetFile)
	if err != nil {
		return false, err
	}
	if hmacOk && masterOk {
		log.Debugf("%s > skipping setting Crypto keys (hmac '%s', master '%s') as they are up-to-date", c.instance.IDColor(), hmacFile, masterFile)
		return false, nil
	}
	log.Infof("%s > copying Crypto hmac file from '%s' to '%s'", c.instance.IDColor(), hmacFile, hmacTargetFile)
	if err := filex.Copy(hmacFile, hmacTargetFile, true); err != nil {
		return false, fmt.Errorf("%s > cannot copy Crypto hmac file from '%s' to '%s': %w", c.instance.IDColor(), hmacFile, hmacTargetFile, err)
	}
	log.Infof("%s > copying Crypto master file from '%s' to '%s'", c.instance.IDColor(), masterFile, masterTargetFile)
	if err := filex.Copy(masterFile, masterTargetFile, true); err != nil {
		return false, fmt.Errorf("%s > cannot copy Crypto master file from '%s' to '%s'> %w", c.instance.IDColor(), masterFile, masterTargetFile, err)
	}
	if err := osgi.Restart(); err != nil {
		return false, err
	}
	return true, nil
}

func (c Crypto) Protect(value string) (string, error) {
	log.Infof("%s > encrypting text using Crypto", c.instance.IDColor())
	response, err := c.instance.http.RequestFormData(map[string]any{"datum": value}).Post(CryptoProtectPath)

	if err != nil {
		return "", fmt.Errorf("%s > cannot encrypt text using Crypto: %w", c.instance.IDColor(), err)
	} else if response.IsError() {
		return "", fmt.Errorf("%s > cannot encrypt text using Crypto: %s", c.instance.IDColor(), response.Status())
	}

	var result CryptoProtectResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &result); err != nil {
		return "", fmt.Errorf("%s > cannot parse Crypto response: %w", c.instance.IDColor(), err)
	}

	return result.Protected, nil
}
