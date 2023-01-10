package pkg

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/mapsx"
	"github.com/wttech/aemc/pkg/repo"
	"net/http"
	"reflect"
)

// Repo Facade for communicating with JCR repository.
type Repo struct {
	instance Instance

	propertyChangeIgnored []string
}

func NewRepo(res *Instance) *Repo {
	return &Repo{
		instance: *res,

		propertyChangeIgnored: []string{"transportPassword"},
	}
}

// Node Creates a new node object.
func (r Repo) Node(path string) RepoNode {
	return NewNode(r, path)
}

// ReplAgent Creates a new node replication agent object.
func (r Repo) ReplAgent(location, name string) ReplAgent {
	return NewReplAgent(r, location, name)
}

func (r Repo) Exists(path string) (bool, error) {
	response, err := r.instance.http.Request().Head(fmt.Sprintf("%s.json", path))
	if err != nil {
		return false, fmt.Errorf("cannot check node existence '%s': %w", path, err)
	}
	switch response.StatusCode() {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound, http.StatusForbidden, http.StatusUnauthorized:
		return false, nil
	default:
		return false, fmt.Errorf("cannot check node existence '%s' - unexpected status code '%d'", path, response.StatusCode())
	}
}

func (r Repo) Read(path string) (map[string]any, error) {
	response, err := r.instance.http.Request().Get(fmt.Sprintf("%s.json", path))
	if err != nil {
		return nil, fmt.Errorf("cannot read properties of node '%s': %w", path, err)
	} else if response.IsError() {
		return nil, fmt.Errorf("cannot read properties of node '%s': %s", path, response.Status())
	}
	var props map[string]any
	err = fmtx.UnmarshalJSON(response.RawBody(), &props)
	return props, nil
}

func (r Repo) Save(path string, props map[string]any) error {
	log.Infof("saving node '%s' on instance '%s'", path, r.instance.ID())
	resp, err := r.requestFormData("", props).Post(path)
	err = r.handleResponse(fmt.Sprintf("cannot save node '%s' on instance '%s'", path, r.instance.ID()), resp, err)
	if err != nil {
		return err
	}
	log.Infof("saved node '%s' on instance '%s'", path, r.instance.ID())
	return nil
}

func (r Repo) Delete(path string) error {
	log.Infof("deleting node '%s' from instance '%s'", path, r.instance.ID())
	resp, err := r.requestFormData("delete", map[string]any{}).Post(path)
	err = r.handleResponse(fmt.Sprintf("cannot delete node '%s' from instance '%s'", path, r.instance.ID()), resp, err)
	if err != nil {
		return err
	}
	log.Infof("deleted node '%s' from instance '%s'", path, r.instance.ID())
	return nil
}

func (r Repo) requestFormData(operation string, props map[string]any) *resty.Request {
	request := r.instance.http.Request()
	request.SetHeader("Accept", "application/json")
	request.FormData.Set(":operation", operation)
	for k, v := range props {
		if v == nil {
			request.FormData.Set(fmt.Sprintf("%s@Delete", k), "")
		} else {
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
				for i := 0; i < rv.Len(); i++ {
					request.FormData.Add(k, fmt.Sprintf("%v", rv.Index(i).Interface()))
				}
				request.FormData.Set(fmt.Sprintf("%s@TypeHint", k), propTypeHint(rv.Type().Elem().Kind())+"[]")
			} else {
				request.FormData.Set(k, fmt.Sprintf("%v", v))
				request.FormData.Set(fmt.Sprintf("%s@TypeHint", k), propTypeHint(rv.Kind()))
			}
		}
	}
	return request
}

func (r Repo) PropsEqual(current map[string]any, updated map[string]any) bool {
	return mapsx.EqualIgnoring(current, updated, r.propertyChangeIgnored)
}

func propTypeHint(kind reflect.Kind) string {
	switch kind {
	case reflect.Bool:
		return "Boolean"
	case reflect.Int, reflect.Int64:
		return "Long"
	case reflect.Float32, reflect.Float64:
		return "Decimal"
	default:
		return "String"
	}
}

func (r Repo) handleResponse(action string, resp *resty.Response, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", action, err)
	} else if resp.IsError() {
		return fmt.Errorf("%s: %s", action, resp.Status())
	}
	var result repo.RepoResult
	err = fmtx.UnmarshalJSON(resp.RawBody(), &result)
	if err != nil {
		return fmt.Errorf("%s; cannot parse response: %w", action, err)
	} else if result.IsError() {
		return fmt.Errorf("%s: %s", action, result.ErrorMessage())
	}
	return nil
}
