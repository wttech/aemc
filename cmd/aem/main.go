package main

import (
	"github.com/wttech/aemc/pkg/common/osx"
)

func main() {
	osx.EnvVarsLoad()

	cli := NewCLI()
	cli.MustExec()
}
