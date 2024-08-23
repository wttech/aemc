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
