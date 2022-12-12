package java

import (
	"fmt"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/osx"
	"os"
)

type Opts struct {
	HomePath string
}

func NewOpts() *Opts {
	return &Opts{
		HomePath: os.Getenv("JAVA_HOME"),
	}
}

func (o *Opts) Validate() error {
	if len(o.HomePath) == 0 {
		return fmt.Errorf("java home path is not set; fix it by setting config property 'java.home_path' or environment variable 'JAVA_HOME'")
	}
	if !osx.PathExists(o.HomePath) {
		return fmt.Errorf("java home path '%s' does not exist", o.HomePath)
	}
	return nil
}

func (o *Opts) Executable() string {
	return o.HomePath + "/bin/java"
}

func (o *Opts) Configure(config *cfg.Config) {
	opts := config.Values().Java

	if len(opts.HomePath) > 0 {
		o.HomePath = opts.HomePath
	}
}
