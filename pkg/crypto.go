package pkg

type Crypto struct {
	instance *Instance
}

func NewCrypto(instance *Instance) *Crypto {
	return &Crypto{instance: instance}
}

func (c Crypto) Configure(hmacFile bool, masterFile bool) (bool, error) {
	// TODO ...
	if err := c.instance.OSGI().Restart(); err != nil {
		return false, err
	}
	return true, nil
}

func (c Crypto) Protect(value string) (string, error) {
	return "", nil
}
