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

func (userManager *UserManager) KeystoreStatus(scope, id string) (*keystore.Status, error) {
	userKeystorePath := UsersPath + "/" + scope + "/" + id + ".ks.json"

	response, err := userManager.instance.http.Request().Get(userKeystorePath)

	if err != nil {
		return nil, fmt.Errorf("%s > cannot read user Keystore: %w", userManager.instance.ID(), err)
	}

	if response.IsError() {
		return nil, fmt.Errorf("%s > cannot read user keystore: %s", userManager.instance.ID(), response.Status())
	}

	result, err := keystore.UnmarshalStatus(response.RawBody())

	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse user Keystore status response: %w", userManager.instance.ID(), err)
	}

	return result, nil
}

func (userManager *UserManager) KeystoreCreate(scope, id, keystorePassword string) (bool, error) {
	statusResponse, statusError := userManager.KeystoreStatus(scope, id)

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
	postResponse, postError := userManager.instance.http.Request().SetQueryParams(pathParams).Post(userKeystoreCreatePath)

	if postError != nil {
		return false, fmt.Errorf("%s > cannot create user keystore: %w", userManager.instance.ID(), postError)
	}

	if postResponse.IsError() {
		return false, fmt.Errorf("%s > cannot create user keystore: %s", userManager.instance.ID(), postResponse.Status())
	}

	return true, nil
}
