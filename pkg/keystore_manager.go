package pkg

import (
	"fmt"
	"os"
	"slices"

	jks "github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/keystore"
)

type KeystoreManager struct {
	instance *Instance
}

func (km *KeystoreManager) Status(scope, id string) (*keystore.Status, error) {
	userKeystorePath := composeKeystoreStatusPath(scope, id)

	response, err := km.instance.http.Request().Get(userKeystorePath)

	if err != nil {
		return nil, fmt.Errorf("%s > cannot read user keystore: %w", km.instance.IDColor(), err)
	}

	if response.IsError() {
		return nil, fmt.Errorf("%s > cannot read user keystore: %s", km.instance.IDColor(), response.Status())
	}

	result, err := keystore.UnmarshalStatus(response.RawBody())
	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse user keystore status response: %w", km.instance.IDColor(), err)
	}

	return result, nil
}

func (km *KeystoreManager) Create(scope, id, keystorePassword string) (bool, error) {
	statusResponse, statusError := km.Status(scope, id)
	if statusError != nil {
		return false, statusError
	}

	if statusResponse.Created {
		return false, nil
	}

	pathParams := map[string]string{
		"newPassword": keystorePassword,
		"rePassword":  keystorePassword,
		":operation":  "createStore",
	}

	userKeystoreCreatePath := composeKeystoreOperationsPath(scope, id)
	postResponse, postError := km.instance.http.Request().SetQueryParams(pathParams).Post(userKeystoreCreatePath)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot create user keystore: %w", km.instance.IDColor(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot create user keystore: %s", km.instance.IDColor(), postResponse.Status())
	}

	return true, nil
}

func (km *KeystoreManager) AddKey(scope, id, keystoreFilePath, keystoreFilePassword, privateKeyAlias, privateKeyPassword, privateKeyNewAlias string) (bool, error) {
	if !pathx.Exists(keystoreFilePath) {
		return false, fmt.Errorf("%s > keystore file does not exist: %s", km.instance.IDColor(), keystoreFilePath)
	}
	if privateKeyNewAlias == "" {
		privateKeyNewAlias = privateKeyAlias
	}
	if privateKeyPassword == "" {
		privateKeyPassword = keystoreFilePassword
	}

	readKeystore, err := readKeyStore(keystoreFilePath, []byte(keystoreFilePassword))
	if err != nil {
		return false, fmt.Errorf("%s > cannot read keystore file %s: %w", km.instance.IDColor(), keystoreFilePath, err)
	}

	aliases := readKeystore.Aliases()
	if aliases == nil {
		return false, fmt.Errorf("%s > keystore file does not contain any aliases", km.instance.IDColor())
	}
	if !slices.Contains(aliases, privateKeyAlias) {
		return false, fmt.Errorf("%s > keystore file does not contain alias: %s", km.instance.IDColor(), privateKeyAlias)
	}

	status, err := km.Status(scope, id)
	if err != nil {
		return false, err
	}

	if status == nil || !status.Created {
		return false, fmt.Errorf("%s > cannot add key as keystore does not exist", km.instance.IDColor())
	}
	if status.HasAlias(privateKeyAlias) {
		return false, nil
	}

	requestFiles := map[string]string{
		"keyStore": keystoreFilePath,
	}

	keystorePath := composeKeystoreOperationsPath(scope, id)
	formData := map[string]string{
		"keyStorePass": keystoreFilePassword,
		"alias":        privateKeyAlias,
		"keyPassword":  privateKeyPassword,
		"newAlias":     privateKeyNewAlias,
		"keyStoreType": "jks",
	}

	response, err := km.instance.http.Request().
		SetFiles(requestFiles).
		SetFormData(formData).
		Post(keystorePath)

	if err != nil {
		return false, fmt.Errorf("%s > cannot add key: %w", km.instance.IDColor(), err)
	}
	if response.IsError() {
		return false, fmt.Errorf("%s > cannot add key: %s", km.instance.IDColor(), response.Status())
	}
	return true, nil
}

func (km *KeystoreManager) DeleteKey(scope, id, privateKeyAlias string) (bool, error) {
	status, err := km.Status(scope, id)
	if err != nil {
		return false, err
	}

	if status == nil || !status.Created {
		return false, fmt.Errorf("%s > cannot delete key: keystore does not exist", km.instance.IDColor())
	}
	if !status.HasAlias(privateKeyAlias) {
		return false, nil
	}

	formData := map[string]string{
		"removeAlias": privateKeyAlias,
	}

	userKeystorePath := composeKeystoreOperationsPath(scope, id)
	response, err := km.instance.http.Request().
		SetFormData(formData).
		Post(userKeystorePath)

	if err != nil {
		return false, fmt.Errorf("%s > cannot delete key: %w", km.instance.IDColor(), err)
	}
	if response.IsError() {
		return false, fmt.Errorf("%s > cannot delete key: %s", km.instance.IDColor(), response.Status())
	}

	return true, nil
}

func composeKeystoreStatusPath(scope, id string) string {
	return composeUserPath(scope, id) + ".ks.json"
}

func composeKeystoreOperationsPath(scope, id string) string {
	return composeUserPath(scope, id) + ".ks.html"
}

func readKeyStore(filename string, password []byte) (*jks.KeyStore, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	ks := jks.New()
	if err := ks.Load(f, password); err != nil {
		return nil, err
	}

	return &ks, nil
}
