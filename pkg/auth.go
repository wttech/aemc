package pkg

type Auth struct {
	instance    *Instance
	userManager *UserManager
}

func NewAuth(instance *Instance) *Auth {
	auth := &Auth{instance: instance}
	auth.userManager = NewUserManager(instance)

	return auth
}

func (auth *Auth) UserManager() *UserManager {
	return auth.userManager
}
