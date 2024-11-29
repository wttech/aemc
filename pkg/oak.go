package pkg

// OAK Facade for managing OAK repository.
type OAK struct {
	instance *Instance

	indexManager *OAKIndexManager
}

func NewOAK(instance *Instance) *OAK {
	return &OAK{
		instance: instance,

		indexManager: NewOAKIndexManager(instance),
	}
}

func (o *OAK) IndexManager() *OAKIndexManager {
	return o.indexManager
}

func (o *OAK) oakRun() *OakRun {
	return o.instance.manager.aem.vendorManager.oakRun
}

func (o *OAK) Compact() error {
	return o.instance.manager.aem.vendorManager.oakRun.Compact(o.instance.local.Dir())
}
