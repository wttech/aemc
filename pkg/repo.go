package pkg

import (
	"bytes"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/mapsx"
	"github.com/wttech/aemc/pkg/repo"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

// Repo Facade for communicating with JCR repository.
type Repo struct {
	instance Instance

	PropertyChangeIgnored []string
}

func NewRepo(i *Instance) *Repo {
	cv := i.manager.aem.config.Values()

	return &Repo{
		instance: *i,

		PropertyChangeIgnored: cv.GetStringSlice("instance.repo.property_change_ignored"),
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
	log.Infof("%s > saving node '%s'", r.instance.ID(), path)
	resp, err := r.requestFormData("", props).Post(path)
	if err := r.handleResponse(fmt.Sprintf("%s > cannot save node '%s'", r.instance.ID(), path), resp, err); err != nil {
		return err
	}
	log.Infof("%s > saved node '%s'", r.instance.ID(), path)
	return nil
}

func (r Repo) Delete(path string) error {
	log.Infof("%s > deleting node '%s'", r.instance.ID(), path)
	resp, err := r.requestFormData("delete", map[string]any{}).Post(path)
	if err = r.handleResponse(fmt.Sprintf("%s > cannot delete node '%s'", r.instance.ID(), path), resp, err); err != nil {
		return err
	}
	log.Infof("%s > deleted node '%s'", r.instance.ID(), path)
	return nil
}

func (r Repo) Copy(sourcePath string, targetPath string) error {
	log.Infof("%s > copying node from '%s' to '%s'", r.instance.ID(), sourcePath, targetPath)
	resp, err := r.requestFormData("copy", map[string]any{":dest": targetPath}).Post(sourcePath)
	if err = r.handleResponse(fmt.Sprintf("%s > cannot copy node from '%s' to '%s'", r.instance.ID(), sourcePath, targetPath), resp, err); err != nil {
		return err
	}
	log.Infof("%s > copied node from '%s' to '%s'", r.instance.ID(), sourcePath, targetPath)
	return nil
}

func (r Repo) Move(sourcePath string, targetPath string, replace bool) error {
	log.Infof("%s > moving node from '%s' to '%s'", r.instance.ID(), sourcePath, targetPath)
	resp, err := r.requestFormData("move", map[string]any{":dest": targetPath, ":replace": replace}).Post(sourcePath)
	if err = r.handleResponse(fmt.Sprintf("%s > cannot move node from '%s' to '%s'", r.instance.ID(), sourcePath, targetPath), resp, err); err != nil {
		return err
	}
	log.Infof("%s > moved node from '%s' to '%s'", r.instance.ID(), sourcePath, targetPath)
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
	return mapsx.EqualIgnoring(current, updated, r.PropertyChangeIgnored)
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

func NewRepoNodeList(nodes []RepoNode) NodeList {
	var sortedNodes []RepoNode
	sortedNodes = append(sortedNodes, nodes...)
	sort.SliceStable(sortedNodes, func(i, j int) bool { return strings.Compare(nodes[i].Name(), nodes[j].Name()) < 0 })
	return NodeList{Nodes: sortedNodes, Total: len(nodes)}
}

type NodeList struct {
	Total int        `json:"total" yaml:"total"`
	Nodes []RepoNode `json:"nodes" yaml:"nodes"`
}

func (r Repo) Download(remotePath string, localFile string) error {
	log.Infof("%s > downloading node '%s'", r.instance.ID(), remotePath)
	if err := httpx.DownloadWithOpts(httpx.DownloadOpts{
		Client:   r.instance.http.Client(),
		URL:      remotePath,
		File:     localFile,
		Override: true,
	}); err != nil {
		return fmt.Errorf("%s > cannot download node '%s': %w", r.instance.ID(), remotePath, err)
	}
	log.Infof("%s > downloaded node '%s'", r.instance.ID(), remotePath)
	return nil
}

func (nl NodeList) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("stats", "stat", "value", map[string]any{"total": len(nl.Nodes)}))
	bs.WriteString("\n")
	bs.WriteString(fmtx.TblRows("list", true, []string{"path"}, lo.Map(nl.Nodes, func(node RepoNode, _ int) map[string]any {
		return map[string]any{"path": node.path}
	})))
	return bs.String()
}
