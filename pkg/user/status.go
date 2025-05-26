package user

import (
	"fmt"
	"io"

	"github.com/wttech/aemc/pkg/common/fmtx"
)

const (
	// Repository user types (as stored in JCR)
	RepUserType       = "rep:User"
	RepSystemUserType = "rep:SystemUser"

	// Maped user types
	UserType       = "user"
	SystemUserType = "systemUser"
)

type Status struct {
	Type           string `json:"jcr:primaryType"`
	AuthorizableID string `json:"rep:authorizableId"`
}

func UnmarshalStatus(readCloser io.ReadCloser) (*Status, error) {
	var status = Status{
		Type:           "rep:User",
		AuthorizableID: "",
	}
	if err := fmtx.UnmarshalJSON(readCloser, &status); err != nil {
		return nil, err
	}

	switch status.Type {
	case RepUserType:
		status.Type = UserType
	case RepSystemUserType:
		status.Type = SystemUserType
	default:
		return nil, fmt.Errorf("unknown user type: %s", status.Type)
	}

	return &status, nil
}
