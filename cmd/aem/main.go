package main

import (
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/cfg"
)

func main() {
	config := cfg.NewConfig()
	api := pkg.NewAem()

	app := NewCLI(api, config)
	app.Exec()
}
