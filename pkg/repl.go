package pkg

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/wttech/aemc/pkg/replication"
	"github.com/wttech/aemc/pkg/sling"
)

type Replication struct {
	instance *Instance

	bundleSymbolicName string
}

func NewReplication(instance *Instance) *Replication {
	cv := instance.manager.aem.config.Values()

	return &Replication{
		instance: instance,

		bundleSymbolicName: cv.GetString("instance.replication.bundle_symbolic_name"),
	}
}

func (r Replication) Agent(location, name string) ReplAgent {
	return r.instance.Repo().ReplAgent(location, name)
}

func (r Replication) Bundle() OSGiBundle {
	return r.instance.OSGI().BundleManager().New(r.bundleSymbolicName)
}

func (r Replication) Activate(path string) error {
	log.Infof("%s > activating path '%s'", r.instance.IDColor(), path)
	if err := r.replicate("activate", path); err != nil {
		return err
	}
	log.Infof("%s > activated path '%s'", r.instance.IDColor(), path)
	return nil
}

func (r Replication) Deactivate(path string) error {
	log.Infof("%s > deactivating path '%s'", r.instance.IDColor(), path)
	if err := r.replicate("deactivate", path); err != nil {
		return err
	}
	log.Infof("%s > deactivated path '%s'", r.instance.IDColor(), path)
	return nil
}

func (r Replication) replicate(cmd string, path string) error {
	response, err := r.instance.http.Request().
		SetFormData(map[string]string{
			"cmd":  cmd,
			"path": path,
		}).
		Post(replication.ReplicateJsonPath)
	if err != nil {
		return fmt.Errorf("%s > cannot do replication command '%s' for path '%s': %w", r.instance.IDColor(), cmd, path, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot do replication command '%s' for path '%s': %s", r.instance.IDColor(), cmd, path, response.Status())
	}
	jsonBytes, err := io.ReadAll(response.RawBody())
	if err != nil {
		return fmt.Errorf("%s > cannot read replication command '%s' response for path '%s': %w", r.instance.IDColor(), cmd, path, err)
	}
	jsonData, err := sling.JsonData(string(jsonBytes))
	if err != nil {
		return fmt.Errorf("%s > cannot parse replication command '%s' JSON response for path '%s': %w", r.instance.IDColor(), cmd, path, err)
	}
	if jsonData.IsError() {
		return fmt.Errorf("%s > replication command '%s' failed for path '%s': %s", r.instance.IDColor(), cmd, path, jsonData.Message)
	}
	return nil
}

func (r Replication) ActivateTree(opts replication.ActivateTreeOpts) error {
	log.Infof("%s > activating tree at path '%s'", r.instance.IDColor(), opts.StartPath)

	cmd := "activate"
	if opts.DryRun {
		cmd = "dryrun"
	}
	response, err := r.instance.http.Request().
		SetFormData(map[string]string{
			"cmd":               cmd,
			"path":              opts.StartPath,
			"onlymodified":      cast.ToString(opts.OnlyModified),
			"reactivate":        cast.ToString(opts.OnlyActivated),
			"ignoredeactivated": cast.ToString(opts.IgnoreDeactivated),
			"__charset__":       "UTF-8",
		}).
		Post(replication.ActivateTreePath)
	if err != nil {
		return fmt.Errorf("%s > cannot activate tree at path '%s': %w", r.instance.IDColor(), opts.StartPath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot activate tree at path: %s: %s", r.instance.IDColor(), opts.StartPath, response.Status())
	}
	log.Infof("%s > activated tree at path '%s'", r.instance.IDColor(), opts.StartPath)
	return nil
}
