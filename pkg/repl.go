package pkg

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
