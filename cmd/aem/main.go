package main

import (
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/osx"
)

func main() {
	osx.EnvVarsLoad()

	config := cfg.NewConfig()
	api := pkg.NewAem()

	app := NewCLI(api, config)
	app.Exec()
}
