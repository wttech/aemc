package pkg

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/wttech/aemc/pkg/common/cert"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/gts"
	"os"
)

type GTSManager struct {
	instance *Instance
}

func NewGTSMananger(instance *Instance) *GTSManager {
	return &GTSManager{
		instance: instance,
	}
}

const (
	GTSPathJson = "/libs/granite/security/truststore.json"
	GTSPath     = "/libs/granite/security/post/truststore"
)

func (gtsManager *GTSManager) Status() (*gts.Status, error) {
	response, err := gtsManager.instance.http.Request().Get(GTSPathJson)

	if err != nil {
		return nil, fmt.Errorf("%s > cannot read global trust store: %w", gtsManager.instance.ID(), err)
	}

	if response.IsError() {
		return nil, fmt.Errorf("%s > cannot read global trust store: %s", gtsManager.instance.ID(), response.Status())
	}

	result, err := gts.UnmarshalStatus(response.RawBody())

	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse global trust store status response: %w", gtsManager.instance.ID(), err)
	}

	return result, nil
}

func (gtsManager *GTSManager) Create(trustStorePassword string) (bool, error) {
	statusResponse, statusError := gtsManager.Status()

	if statusError != nil {
		return false, statusError
	}

	if statusResponse.Created == nil {
		return false, nil
	}

	postResponse, postError := gtsManager.instance.http.Request().Post(GTSPath + "?newPassword=" + trustStorePassword + "&rePassword=" + trustStorePassword + "&:operation=createStore")

	if postError != nil {
		return false, fmt.Errorf("%s > cannot create global trust store: %w", gtsManager.instance.ID(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot create global trust store: %s", gtsManager.instance.ID(), postResponse.Status())
	}

	return true, nil
}

func (gtsManager *GTSManager) AddCertificate(certificateFilePath string) (*gts.Certificate, bool, error) {

	if !pathx.Exists(certificateFilePath) {
		return nil, false, fmt.Errorf("%s > certificate file does not exist: %s", gtsManager.instance.ID(), certificateFilePath)
	}

	cretificateData, err := os.ReadFile(certificateFilePath)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to read certificate file '%s': %w", gtsManager.instance.ID(), certificateFilePath, err)
	}

	pemBlock, _ := pem.Decode(cretificateData)
	if pemBlock != nil {
		tmpDerFileNameBasedOnPemPath, functionToDefer, err := cert.CreateTmpDerFileBasedOnPem(pemBlock)

		defer functionToDefer()

		if err != nil {
			return nil, false, fmt.Errorf("%s > %w", gtsManager.instance.ID(), err)
		}
		cretificateData, err = os.ReadFile(*tmpDerFileNameBasedOnPemPath)
	}

	certificate, err := x509.ParseCertificate(cretificateData)

	if err != nil {
		return nil, false, err
	}

	statusResponse, statusError := gtsManager.Status()

	if statusError != nil {
		return nil, false, statusError
	}

	certificateInTrustStore, err := statusResponse.FindCertificate(*certificate)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to compare certificate with certificates in trust store: %w", gtsManager.instance.ID(), err)
	}

	if certificateInTrustStore != nil {
		return certificateInTrustStore, false, nil
	}

	response, err := gtsManager.instance.http.Request().SetFiles(map[string]string{
		"certificate": certificateFilePath,
	}).Post(GTSPath)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to add certificate to trust store: %w", gtsManager.instance.ID(), err)
	}

	if response.IsError() {
		return nil, false, fmt.Errorf("%s > failed to add certificate to trust store: %s", gtsManager.instance.ID(), response.Status())
	}

	statusResponse, statusError = gtsManager.Status()

	if statusError != nil {
		return nil, false, statusError
	}

	certificateInTrustStore, err = statusResponse.FindCertificate(*certificate)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to find added certificate in trust store: %w", gtsManager.instance.ID(), err)
	}

	return certificateInTrustStore, true, nil
}

func (gtsManager *GTSManager) RemoveCertificate(certifiacteAlias string) (bool, error) {
	statusResponse, statusError := gtsManager.Status()

	if statusError != nil {
		return false, statusError
	}

	certificate := statusResponse.FindCertificateByAlias(certifiacteAlias)

	if certificate == nil {
		return false, nil
	}

	postResponse, postError := gtsManager.instance.http.Request().Post(GTSPath + "?removeAlias=" + certifiacteAlias)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot remove certificate from trust store: %w", gtsManager.instance.ID(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot remove certificate from trust store: %s", gtsManager.instance.ID(), postResponse.Status())
	}

	return true, nil
}

func (gtsManager *GTSManager) ReadCertificate(certifiacteAlias string) (*gts.Certificate, error) {
	statusResponse, statusError := gtsManager.Status()

	if statusError != nil {
		return nil, statusError
	}

	certificate := statusResponse.FindCertificateByAlias(certifiacteAlias)

	return certificate, nil
}
