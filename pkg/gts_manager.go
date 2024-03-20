package pkg

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/wttech/aemc/pkg/common/certx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/gts"
	"os"
)

type GTSManager struct {
	instance *Instance
}

func NewGTSManager(instance *Instance) *GTSManager {
	return &GTSManager{instance: instance}
}

const (
	GTSPathJson = "/libs/granite/security/truststore.json"
	GTSPath     = "/libs/granite/security/post/truststore"
)

func (gm *GTSManager) Status() (*gts.Status, error) {
	response, err := gm.instance.http.Request().Get(GTSPathJson)

	if err != nil {
		return nil, fmt.Errorf("%s > cannot read GTS: %w", gm.instance.IDColor(), err)
	}

	if response.IsError() {
		return nil, fmt.Errorf("%s > cannot read GTS: %s", gm.instance.IDColor(), response.Status())
	}

	result, err := gts.UnmarshalStatus(response.RawBody())

	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse GTS status response: %w", gm.instance.IDColor(), err)
	}

	return result, nil
}

func (gm *GTSManager) Create(trustStorePassword string) (bool, error) {
	statusResponse, statusError := gm.Status()

	if statusError != nil {
		return false, statusError
	}

	if statusResponse.Created == true {
		return false, nil
	}

	pathParams := map[string]string{
		"newPassword": trustStorePassword,
		"rePassword":  trustStorePassword,
		":operation":  "createStore",
	}

	postResponse, postError := gm.instance.http.Request().SetQueryParams(pathParams).Post(GTSPath)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot create GTS: %w", gm.instance.IDColor(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot create GTS: %s", gm.instance.IDColor(), postResponse.Status())
	}

	return true, nil
}

func (gm *GTSManager) AddCertificate(certificateFilePath string) (*gts.Certificate, bool, error) {

	if !pathx.Exists(certificateFilePath) {
		return nil, false, fmt.Errorf("%s > certificate file does not exist: %s", gm.instance.IDColor(), certificateFilePath)
	}

	cretificateData, err := os.ReadFile(certificateFilePath)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to read certificate file '%s': %w", gm.instance.IDColor(), certificateFilePath, err)
	}

	pemBlock, _ := pem.Decode(cretificateData)
	if pemBlock != nil {
		tmpDerFileNameBasedOnPemPath, cleanCallback, err := certx.CreateTmpDerFileBasedOnPem(pemBlock)

		defer cleanCallback()

		if err != nil {
			return nil, false, fmt.Errorf("%s > %w", gm.instance.IDColor(), err)
		}
		cretificateData, err = os.ReadFile(*tmpDerFileNameBasedOnPemPath)
	}

	certificate, err := x509.ParseCertificate(cretificateData)

	if err != nil {
		return nil, false, err
	}

	statusResponse, statusError := gm.Status()

	if statusError != nil {
		return nil, false, statusError
	}

	certificateInTrustStore, err := statusResponse.FindCertificate(*certificate)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to compare certificate with certificates in GTS: %w", gm.instance.IDColor(), err)
	}

	if certificateInTrustStore != nil {
		return certificateInTrustStore, false, nil
	}

	requestFiles := map[string]string{
		"certificate": certificateFilePath,
	}

	response, err := gm.instance.http.Request().SetFiles(requestFiles).Post(GTSPath)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to add certificate to GTS: %w", gm.instance.IDColor(), err)
	}

	if response.IsError() {
		return nil, false, fmt.Errorf("%s > failed to add certificate to GTS: %s", gm.instance.IDColor(), response.Status())
	}

	statusResponse, statusError = gm.Status()

	if statusError != nil {
		return nil, false, statusError
	}

	certificateInTrustStore, err = statusResponse.FindCertificate(*certificate)

	if err != nil {
		return nil, false, fmt.Errorf("%s > failed to find added certificate in GTS: %w", gm.instance.IDColor(), err)
	}

	return certificateInTrustStore, true, nil
}

func (gm *GTSManager) RemoveCertificate(certifiacteAlias string) (bool, error) {
	statusResponse, statusError := gm.Status()

	if statusError != nil {
		return false, statusError
	}

	certificate := statusResponse.FindCertificateByAlias(certifiacteAlias)

	if certificate == nil {
		return false, nil
	}

	pathParams := map[string]string{
		"removeAlias": certifiacteAlias,
	}

	postResponse, postError := gm.instance.http.Request().SetQueryParams(pathParams).Post(GTSPath)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot remove certificate from GTS: %w", gm.instance.IDColor(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot remove certificate from GTS: %s", gm.instance.IDColor(), postResponse.Status())
	}

	return true, nil
}

func (gm *GTSManager) ReadCertificate(certifiacteAlias string) (*gts.Certificate, error) {
	statusResponse, statusError := gm.Status()

	if statusError != nil {
		return nil, statusError
	}

	certificate := statusResponse.FindCertificateByAlias(certifiacteAlias)

	return certificate, nil
}
