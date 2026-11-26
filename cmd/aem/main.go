package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/osx"
	"os"
)

func main() {
	if err := applyBootstrapFlags(parseBootstrapFlags(os.Args[1:])); err != nil {
		log.Fatal(err)
	}
	osx.EnvVarsLoad()

	cli := NewCLI()
	cli.MustExec()
}
