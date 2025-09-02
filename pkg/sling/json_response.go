package sling

import (
	"encoding/json"

	"github.com/spf13/cast"
)

type JSONData struct {
	Status        int
	Message       string
	Path          []string
	StatusMessage string
	StatusCode    int
}

func JsonData(jsonStr string) (data JSONData, err error) {
	var rawData map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return data, err
	}

	data.Path = cast.ToStringSlice(rawData["path"])
	data.StatusMessage = cast.ToString(rawData["status.message"])
	data.Message = data.StatusMessage
	data.StatusCode = cast.ToInt(rawData["status.code"])
	data.Status = data.StatusCode

	return data, nil
}

func (d JSONData) IsError() bool {
	return d.Status <= 0 || d.Status > 399
}
