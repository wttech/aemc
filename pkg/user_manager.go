package pkg

import (
	"fmt"
	"os"

	jks "github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/keystore"
	"github.com/wttech/aemc/pkg/user"
	"golang.org/x/exp/slices"
)

type UserManager struct {
	instance *Instance
}

func NewUserManager(instance *Instance) *UserManager {
	return &UserManager{instance: instance}
}

const (
	UsersPath = "/home/users"
)

func (um *UserManager) KeystoreStatus(scope, id string) (*keystore.Status, error) {
	userKeystorePath := assembleUserPath(scope, id) + ".ks.json"

	response, err := um.instance.http.Request().Get(userKeystorePath)

	if err != nil {
		return nil, fmt.Errorf("%s > cannot read user Keystore: %w", um.instance.IDColor(), err)
	}

	if response.IsError() {
		return nil, fmt.Errorf("%s > cannot read user keystore: %s", um.instance.IDColor(), response.Status())
	}

	result, err := keystore.UnmarshalStatus(response.RawBody())

	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse user Keystore status response: %w", um.instance.IDColor(), err)
	}

	return result, nil
}

func (um *UserManager) KeystoreCreate(scope, id, keystorePassword string) (bool, error) {
	statusResponse, statusError := um.KeystoreStatus(scope, id)

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

	userKeystoreCreatePath := assembleUserPath(scope, id) + ".ks.html"
	postResponse, postError := um.instance.http.Request().SetQueryParams(pathParams).Post(userKeystoreCreatePath)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot create user keystore: %w", um.instance.IDColor(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot create user keystore: %s", um.instance.IDColor(), postResponse.Status())
	}

	return true, nil
}

func (um *UserManager) AddKeystoreKey(scope, id, keystoreFilePath, keystoreFilePassword, privateKeyAlias, privateKeyPassword, privateKeyNewAlias string) (bool, error) {
	if !pathx.Exists(keystoreFilePath) {
		return false, fmt.Errorf("%s > keystore file does not exist: %s", um.instance.IDColor(), keystoreFilePath)
	}
	if privateKeyNewAlias == "" {
		privateKeyNewAlias = privateKeyAlias
	}
	if privateKeyPassword == "" {
		privateKeyPassword = keystoreFilePassword
	}

	readKeystore, err := readKeyStore(keystoreFilePath, []byte(keystoreFilePassword))
	if err != nil {
		return false, fmt.Errorf("%s > cannot read keystore file %s: %w", um.instance.IDColor(), keystoreFilePath, err)
	}

	aliases := readKeystore.Aliases()

	if aliases == nil {
		return false, fmt.Errorf("%s > keystore does not contain any aliases", um.instance.IDColor())
	}
	if !slices.Contains(aliases, privateKeyAlias) {
		return false, fmt.Errorf("%s > keystore does not contain alias: %s", um.instance.IDColor(), privateKeyAlias)
	}

	keystorePath := assembleUserPath(scope, id) + ".ks.html"
	keystoreStatusPath := assembleUserPath(scope, id) + ".ks.json"

	statusResponse, err := um.instance.http.Request().Get(keystoreStatusPath)

	if err != nil {
		return false, fmt.Errorf("%s > cannot read user Keystore: %w", um.instance.IDColor(), err)
	}

	if statusResponse.IsError() {
		return false, fmt.Errorf("%s > cannot read user keystore: %s", um.instance.IDColor(), statusResponse.Status())
	}

	status, err := keystore.UnmarshalStatus(statusResponse.RawBody())

	if err != nil {
		return false, fmt.Errorf("%s > cannot parse user Keystore status response: %w", um.instance.IDColor(), err)
	}
	if status == nil || !status.Created {
		return false, fmt.Errorf("%s > cannot delete keystore key: keystore does not exist", um.instance.IDColor())
	}
	if status.HasAlias(privateKeyAlias) {
		return false, nil
	}

	requestFiles := map[string]string{
		"keyStore": keystoreFilePath,
	}

	formData := map[string]string{
		"keyStorePass": keystoreFilePassword,
		"alias":        privateKeyAlias,
		"keyPassword":  privateKeyPassword,
		"newAlias":     privateKeyNewAlias,
		"keyStoreType": "jks",
	}

	response, err := um.instance.http.Request().
		SetFiles(requestFiles).
		SetFormData(formData).
		Post(keystorePath)

	if err != nil {
		return false, fmt.Errorf("%s > cannot add keystore key: %w", um.instance.IDColor(), err)
	}
	if response.IsError() {
		return false, fmt.Errorf("%s > cannot add keystore key: %s", um.instance.IDColor(), response.Status())
	}
	return true, nil

}

func (um *UserManager) DeleteKeystoreKey(scope, id, privateKeyAlias string) (bool, error) {
	userKeystorePath := assembleUserPath(scope, id) + ".ks.html"
	userKeystoreStatusPath := assembleUserPath(scope, id) + ".ks.json"

	statusResponse, err := um.instance.http.Request().Get(userKeystoreStatusPath)

	if err != nil {
		return false, fmt.Errorf("%s > cannot read user Keystore: %w", um.instance.IDColor(), err)
	}

	if statusResponse.IsError() {
		return false, fmt.Errorf("%s > cannot read user keystore: %s", um.instance.IDColor(), statusResponse.Status())
	}

	status, err := keystore.UnmarshalStatus(statusResponse.RawBody())

	if err != nil {
		return false, fmt.Errorf("%s > cannot parse user Keystore status response: %w", um.instance.IDColor(), err)
	}
	if status == nil || !status.Created {
		return false, fmt.Errorf("%s > cannot delete keystore key: keystore does not exist", um.instance.IDColor())
	}
	if !status.HasAlias(privateKeyAlias) {
		return false, nil
	}

	formData := map[string]string{
		"removeAlias": privateKeyAlias,
	}

	response, err := um.instance.http.Request().
		SetFormData(formData).
		Post(userKeystorePath)

	if err != nil {
		return false, fmt.Errorf("%s > cannot delete keystore key: %w", um.instance.IDColor(), err)
	}
	if response.IsError() {
		return false, fmt.Errorf("%s > cannot delete keystore key: %s", um.instance.IDColor(), response.Status())
	}

	return true, nil
}

func (um *UserManager) ReadState(scope string, id string) (*user.Status, error) {
	userPath := assembleUserPath(scope, id)

	response, err := um.instance.http.Request().Get(userPath + ".json")

	if err != nil {
		return nil, fmt.Errorf("%s > cannot read user: %w", um.instance.IDColor(), err)
	}
	if response.IsError() {
		return nil, fmt.Errorf("%s > cannot read user: %s", um.instance.IDColor(), response.Status())
	}

	result, err := user.UnmarshalStatus(response.RawBody())

	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse user status response: %w", um.instance.IDColor(), err)
	}

	return result, nil
}

func (um *UserManager) SetPassword(scope string, id string, password string) (bool, error) {
	userStatus, err := um.ReadState(scope, id)

	if err != nil {
		return false, err
	}

	userPath := assembleUserPath(scope, id)

	passwordCheckResponse, err := um.instance.http.Request().
		SetBasicAuth(userStatus.AuthorizableID, password).
		Get(userPath + ".json")

	if err != nil {
		return false, fmt.Errorf("%s > cannot check user password: %w", um.instance.IDColor(), err)
	}
	if !passwordCheckResponse.IsError() {
		return false, nil
	}

	props := map[string]any{
		"rep:password": password,
	}

	postResponse, err := um.instance.http.RequestFormData(props).Post(userPath)

	if err != nil {
		return false, fmt.Errorf("%s > cannot set user password: %w", um.instance.IDColor(), err)
	}
	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot set user password: %s", um.instance.IDColor(), postResponse.Status())
	}

	return true, nil
}

func assembleUserPath(scope string, id string) string {
	if scope == "" {
		return UsersPath + "/" + id
	}
	return UsersPath + "/" + scope + "/" + id
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
