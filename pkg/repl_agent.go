package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common"
	"golang.org/x/exp/maps"
	"strings"
)

type ReplAgent struct {
	page RepoNode
}

func NewReplAgent(repo Repo, location string, name string) ReplAgent {
	page := NewNode(repo, fmt.Sprintf("/etc/replication/agents.%s/%s", location, name))
	return ReplAgent{page: page}
}

func (ra ReplAgent) Name() string {
	return ra.page.Name()
}

func (ra ReplAgent) Setup(props map[string]any) (bool, error) {
	props["provisioner"] = common.AppId // enforce at least one-time provisioning due to 'transportPassword' property

	changed := false
	pageState, err := ra.page.State()
	if err != nil {
		return false, fmt.Errorf("%s > cannot read replication agent '%s': %w", ra.instanceID(), ra.page.Path(), err)
	}
	if !pageState.Exists {
		err = ra.page.Save(map[string]any{
			"jcr:primaryType": "cq:Page",
		})
		if err != nil {
			return false, fmt.Errorf("%s > cannot setup replication agent '%s': %w", ra.instanceID(), ra.page.Path(), err)
		}
		changed = true
	}
	pageContent := ra.page.Content()
	pageContentState, err := pageContent.State()
	if err != nil {
		return changed, fmt.Errorf("%s > cannot read replication agent '%s' exist: %w", ra.instanceID(), pageContent.Path(), err)
	}
	if !pageContentState.Exists {
		maps.Copy(props, map[string]any{
			"jcr:primaryType":    "nt:unstructured",
			"jcr:title":          strings.ToTitle(ra.Name()),
			"sling:resourceType": "cq/replication/components/agent",
			"cq:template":        "/libs/cq/replication/templates/agent",
		})
		err = pageContent.Save(props)
		if err != nil {
			return changed, fmt.Errorf("%s > cannot create replication agent '%s': %w", ra.instanceID(), pageContent.Path(), err)
		}
		changed = true
	} else {
		changed, err = pageContent.SaveWithChanged(props) // TODO react when transportPassword changes externally, now is ignored
		if err != nil {
			return changed, fmt.Errorf("%s > cannot update replication agent '%s': %w", ra.instanceID(), pageContent.Path(), err)
		}
	}

	return changed, nil
}

func (ra ReplAgent) Delete() (bool, error) {
	pageState, err := ra.page.State()
	if err != nil {
		return false, fmt.Errorf("%s > cannot read replication agent '%s': %w", ra.instanceID(), ra.page.Path(), err)
	}
	if !pageState.Exists {
		return false, nil
	}
	err = ra.page.Delete()
	if err != nil {
		return false, fmt.Errorf("%s > cannot delete replication agent '%s': %w", ra.instanceID(), ra.page.Path(), err)
	}
	return true, nil
}

func (ra ReplAgent) instanceID() string {
	return ra.page.repo.instance.id
}

func (ra ReplAgent) MarshalJSON() ([]byte, error) {
	return ra.page.Content().MarshalJSON()
}

func (ra ReplAgent) MarshalYAML() (interface{}, error) {
	return ra.page.Content().MarshalYAML()
}

func (ra ReplAgent) MarshalText() string {
	return ra.page.Content().MarshalText()
}
