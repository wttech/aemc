package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/keystore"
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
