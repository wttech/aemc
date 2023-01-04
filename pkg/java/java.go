package java

import (
	"fmt"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/osx"
	"os"
)

type Opts struct {
	HomeDir string
}

func NewOpts() *Opts {
	return &Opts{
		HomeDir: os.Getenv("JAVA_HOME"),
	}
}

func (o *Opts) Validate() error {
	if len(o.HomeDir) == 0 {
		return fmt.Errorf("java home dir is not set; fix it by setting config property 'java.home_dir' or environment variable 'JAVA_HOME'")
	}
	if !osx.PathExists(o.HomeDir) {
		return fmt.Errorf("java home dir '%s' does not exist", o.HomeDir)
	}
	return nil
}

func (o *Opts) Executable() string {
	return o.HomeDir + "/bin/java"
}

func (o *Opts) Configure(config *cfg.Config) {
	opts := config.Values().Java

	if len(opts.HomeDir) > 0 {
		o.HomeDir = opts.HomeDir
	}
}
