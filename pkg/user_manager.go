package pkg

import (
	"fmt"

	"github.com/wttech/aemc/pkg/keystore"
	"github.com/wttech/aemc/pkg/user"
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
	userKeystorePath := UsersPath + "/" + scope + "/" + id + ".ks.json"

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

	if statusResponse.Created == true {
		return false, nil
	}

	pathParams := map[string]string{
		"newPassword": keystorePassword,
		"rePassword":  keystorePassword,
		":operation":  "createStore",
	}

	userKeystoreCreatePath := UsersPath + "/" + scope + "/" + id + ".ks.html"
	postResponse, postError := um.instance.http.Request().SetQueryParams(pathParams).Post(userKeystoreCreatePath)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot create user keystore: %w", um.instance.IDColor(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot create user keystore: %s", um.instance.IDColor(), postResponse.Status())
	}

	return true, nil
}

func (um *UserManager) UserStatus(scope string, id string) (*user.Status, error) {
	userPath := UsersPath + "/" + scope + "/" + id

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

func (um *UserManager) UserPasswordSet(scope string, id string, password string) (bool, error) {
	userStatus, err := um.UserStatus(scope, id)

	if err != nil {
		return false, err
	}

	userPath := UsersPath + "/" + scope + "/" + id

	passwordCheckResponse, passwordCheckError := um.instance.http.Request().
		SetBasicAuth(userStatus.AuthorizableID, password).
		Get(userPath + ".json")

	if passwordCheckError != nil {
		return false, fmt.Errorf("%s > cannot check user password: %w", um.instance.IDColor(), passwordCheckError)
	}
	if !passwordCheckResponse.IsError() {
		return false, nil
	}

	props := map[string]any{
		"rep:password": password,
	}

	postResponse, postError := um.instance.http.RequestFormData(props).Post(userPath)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot set user password: %w", um.instance.IDColor(), postError)
	}
	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot set user password: %s", um.instance.IDColor(), postResponse.Status())
	}

	return true, nil
}
