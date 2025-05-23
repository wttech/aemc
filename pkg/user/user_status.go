package user

import (
	"fmt"
	"io"

	"github.com/wttech/aemc/pkg/common/fmtx"
)

type Status struct {
	Type           string
	AuthorizableID string
}

func UnmarshalStatus(readCloser io.ReadCloser) (*Status, error) {
	var aux struct {
		JcrPrimaryType    string `json:"jcr:primaryType"`
		RepAuthorizableID string `json:"rep:authorizableId"`
	}
	if err := fmtx.UnmarshalJSON(readCloser, &aux); err != nil {
		return nil, err
	}

	status := &Status{}
	switch aux.JcrPrimaryType {
	case "rep:User":
		status.Type = "user"
	case "rep:SystemUser":
		status.Type = "systemUser"
	default:
		return nil, fmt.Errorf("unknown user type: %s", aux.JcrPrimaryType)
	}

	status.AuthorizableID = aux.RepAuthorizableID
	return status, nil
}
