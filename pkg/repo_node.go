package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/langx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"golang.org/x/exp/maps"
	"strings"
)

// RepoNode represents single node in JCR repository
type RepoNode struct {
	repo Repo

	path string
}

func NewNode(repo Repo, path string) RepoNode {
	return RepoNode{
		repo: repo,
		path: "/" + strings.Trim(path, "/"),
	}
}

func (n RepoNode) Path() string {
	return n.path
}

func (n RepoNode) Root() bool {
	return n.path == "/"
}

func (n RepoNode) Name() string {
	return stringsx.AfterLast(n.path, "/")
}

func (n RepoNode) Extension() string {
	return stringsx.AfterLast(n.Name(), ".")
}

func (n RepoNode) Content() RepoNode {
	return n.Child("jcr:content")
}

func (n RepoNode) Parent() RepoNode {
	parentPath := stringsx.BeforeLast(n.path, "/")
	if parentPath == "" {
		parentPath = "/"
	}
	return NewNode(n.repo, parentPath)
}

func (n RepoNode) Parents() []RepoNode {
	result := []RepoNode{}
	current := n
	for {
		current = current.Parent()
		if current.Root() {
			break
		}
		result = append(result, current)
	}
	return result
}

func (n RepoNode) Breadcrumb() []RepoNode {
	result := []RepoNode{n}
	current := n
	for {
		current = current.Parent()
		if current.Root() {
			break
		}
		result = append(result, current)
	}
	return lo.Reverse(result)
}

func (n RepoNode) Child(name string) RepoNode {
	return NewNode(n.repo, fmt.Sprintf("%s/%s", n.path, name))
}

func (n RepoNode) Children() ([]RepoNode, error) {
	response, err := n.repo.instance.http.Request().Get(fmt.Sprintf("%s.harray.1.json", n.path))
	if err != nil {
		return nil, fmt.Errorf("cannot read children of node '%s': %w", n.path, err)
	} else if response.IsError() {
		return nil, fmt.Errorf("cannot read children of node '%s': %s", n.path, response.Status())
	}
	var children nodeArrayChildren
	err = fmtx.UnmarshalJSON(response.RawBody(), &children)
	if err != nil {
		return nil, fmt.Errorf("cannot parse children of node '%s': %w", n.path, err)
	}
	childrenWithType := lo.Filter(children.Children, func(c nodeArrayChild, _ int) bool { return c.PrimaryType != "" })
	return lo.Map(childrenWithType, func(child nodeArrayChild, _ int) RepoNode { return n.Child(child.Name) }), nil
}

type nodeArrayChildren struct {
	Children []nodeArrayChild `json:"__children__"`
}

type nodeArrayChild struct {
	Name        string `json:"__name__"`
	PrimaryType string `json:"jcr:primaryType,omitempty"`
}

func (n RepoNode) Siblings() ([]RepoNode, error) {
	parentChildren, err := n.Parent().Children()
	if err != nil {
		return nil, err
	}
	return lo.Filter(parentChildren, func(child RepoNode, _ int) bool { return child.path != n.path }), nil
}

func (n RepoNode) Sibling(name string) RepoNode {
	return n.Parent().Child(name)
}

type RepoNodeState struct {
	Path       string         `yaml:"path" json:"path"`
	Exists     bool           `yaml:"exists" json:"exists"`
	Properties map[string]any `yaml:"properties" json:"properties"`
}

func (n RepoNode) State() (*RepoNodeState, error) {
	exists, err := n.Exists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return &RepoNodeState{
			Path:   n.path,
			Exists: false,
		}, nil
	}
	props, err := n.ReadProps()
	if err != nil {
		return nil, err
	}
	return &RepoNodeState{
		Path:       n.path,
		Exists:     true,
		Properties: props,
	}, nil
}

func (n RepoNode) Exists() (bool, error) {
	return n.repo.Exists(n.path)
}

func (n RepoNode) ReadProps() (map[string]any, error) {
	return n.repo.Read(n.path)
}

func (n RepoNode) Save(props map[string]any) error {
	return n.repo.Save(n.path, props)
}

func (n RepoNode) SaveWithChanged(props map[string]any) (bool, error) {
	state, err := n.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		err = n.repo.Save(n.path, props)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	propsBefore := maps.Clone(state.Properties)
	if n.repo.PropsEqual(propsBefore, props) {
		return false, nil
	}
	err = n.repo.Save(n.path, props)
	if err != nil {
		return false, err
	}
	state, err = n.State()
	if err != nil {
		return false, err
	}
	return !n.repo.PropsEqual(propsBefore, state.Properties), nil
}

func (n RepoNode) DeleteWithChanged() (bool, error) {
	exists, err := n.Exists()
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	err = n.repo.Delete(n.path)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (n RepoNode) Delete() error {
	exists, err := n.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s > node '%s' cannot be deleted as it does not exist", n.repo.instance.ID(), n.path)
	}
	return n.repo.Delete(n.path)
}

func (n RepoNode) DeleteProp(name string) error {
	return n.SaveProp(name, nil)
}

func (n RepoNode) SaveProp(name string, value any) error {
	return n.repo.Save(n.path, map[string]any{name: value})
}

func (n RepoNode) Copy(targetPath string) error {
	exists, err := n.repo.Exists(n.path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s > node '%s' cannot be copied as it does not exist", n.repo.instance.ID(), n.path)
	}
	return n.repo.Copy(n.path, targetPath)
}

func (n RepoNode) CopyTo(targetDir string) error {
	return n.Copy(fmt.Sprintf("%s/%s", targetDir, n.Name()))
}

func (n RepoNode) CopyWithChanged(targetPath string) (bool, error) {
	targetExists, err := n.repo.Exists(targetPath)
	if err != nil {
		return false, err
	}
	if targetExists {
		return false, nil
	}
	return true, n.Copy(targetPath)
}

func (n RepoNode) Move(targetPath string, replace bool) error {
	sourceExists, err := n.repo.Exists(n.path)
	if err != nil {
		return err
	}
	if !sourceExists {
		return fmt.Errorf("%s > node '%s' cannot be moved as it does not exist", n.repo.instance.ID(), n.path)
	}
	if !replace {
		targetExists, err := n.repo.Exists(targetPath)
		if err != nil {
			return err
		}
		if targetExists {
			return fmt.Errorf("%s > node '%s' cannot be moved to path '%s' as it already exists", n.repo.instance.ID(), n.path, targetPath)
		}
	}
	return n.repo.Move(n.path, targetPath, replace)
}

func (n RepoNode) MoveTo(targetDir string, replace bool) error {
	return n.Move(fmt.Sprintf("%s/%s", targetDir, n.Name()), replace)
}

func (n RepoNode) MoveWithChanged(targetPath string, replace bool) (bool, error) {
	sourceExists, err := n.repo.Exists(n.path)
	if err != nil {
		return false, err
	}
	if !sourceExists {
		return false, nil
	}
	if !replace {
		targetExists, err := n.repo.Exists(targetPath)
		if err != nil {
			return false, err
		}
		if targetExists {
			return false, fmt.Errorf("%s > node '%s' cannot be moved to path '%s' as it already exists", n.repo.instance.ID(), n.path, targetPath)
		}
	}
	return true, n.repo.Move(n.path, targetPath, replace)
}

func (n RepoNode) Download(localFile string) error {
	state, err := n.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("%s > node '%s' cannot be downloaded as it does not exist", n.repo.instance.ID(), n.path)
	}
	return n.repo.Download(state.Path, localFile)
}

func (n RepoNode) DownloadWithChanged(localFile string) (bool, error) {
	state, err := n.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, fmt.Errorf("%s > node '%s' cannot be downloaded as it does not exist", n.repo.instance.ID(), n.path)
	}
	if !pathx.Exists(localFile) {
		return true, n.repo.Download(state.Path, localFile)
	}
	return false, nil
}

func (n RepoNode) String() string {
	return fmt.Sprintf("node '%s'", n.path)
}

func (n RepoNode) MarshalJSON() ([]byte, error) {
	state, err := n.State()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state)
}

func (n RepoNode) MarshalYAML() (interface{}, error) {
	return n.State()
}

func (n RepoNode) MarshalText() string {
	state, err := n.State()
	if err != nil {
		return fmt.Sprintf("path '%s' state cannot be read\n", n.path)
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("path '%s'\n", n.path))
		sb.WriteString(fmtx.TblProps(state.Properties))
	} else {
		sb.WriteString(fmt.Sprintf("path '%s' does not exist\n", n.path))
	}
	return sb.String()
}

func (n RepoNode) Traversor() RepoNodeTraversor {
	return RepoNodeTraversor{nodes: langx.NewStackWithValue(n)}
}

type RepoNodeTraversor struct {
	nodes langx.Stack[RepoNode]
}

func (i *RepoNodeTraversor) Next() (RepoNode, bool, error) {
	var zero RepoNode
	if i.nodes.IsEmpty() {
		return zero, false, nil
	}
	current := i.nodes.Pop()
	children, err := current.Children()
	if err != nil {
		return zero, true, err
	}
	for _, child := range children {
		i.nodes.Push(child)
	}
	return current, true, nil
}
