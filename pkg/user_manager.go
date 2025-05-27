package pkg

import (
	"fmt"
	"os"

	jks "github.com/pavlo-v-chernykh/keystore-go/v4"
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

func (um *UserManager) Keystore() *KeystoreManager {
	return &KeystoreManager{instance: um.instance}
}

func (um *UserManager) ReadState(scope string, id string) (*user.Status, error) {
	userPath := composeUserPath(scope, id)

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

	userPath := composeUserPath(scope, id)
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

func composeUserPath(scope string, id string) string {
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
