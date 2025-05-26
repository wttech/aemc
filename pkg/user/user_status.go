package user

import (
	"fmt"
	"io"

	"github.com/wttech/aemc/pkg/common/fmtx"
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
	case "rep:User":
		status.Type = "user"
	case "rep:SystemUser":
		status.Type = "systemUser"
	default:
		return nil, fmt.Errorf("unknown user type: %s", status.Type)
	}

	return &status, nil
}
