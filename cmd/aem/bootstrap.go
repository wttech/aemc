package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/wttech/aemc/pkg/cfg"
	"io"
	"os"
)

// Flags that decide where the project and its config are, so they are applied before the env files and the config are read.
// They are declared on the root command too, for help and validation.
const (
	flagCd     = "cd"
	flagConfig = "config"
)

type bootstrapFlags struct {
	cd     string
	config string
}

func parseBootstrapFlags(args []string) bootstrapFlags {
	var result bootstrapFlags
	fs := pflag.NewFlagSet("bootstrap", pflag.ContinueOnError)
	fs.ParseErrorsAllowlist.UnknownFlags = true
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	fs.BoolP("help", "h", false, "")
	fs.StringVarP(&result.cd, flagCd, "C", "", "")
	fs.StringVar(&result.config, flagConfig, "", "")
	_ = fs.Parse(args)
	return result
}

// applyBootstrapFlags changes the working directory (like 'git -C', 'make -C', 'mise -C') and points to the config file.
func applyBootstrapFlags(flags bootstrapFlags) error {
	if flags.cd != "" {
		if err := os.Chdir(flags.cd); err != nil {
			return fmt.Errorf("cannot change directory to '%s': %w", flags.cd, err)
		}
	}
	if flags.config != "" {
		if err := os.Setenv(cfg.FileEnvVar, flags.config); err != nil {
			return fmt.Errorf("cannot use config file '%s': %w", flags.config, err)
		}
	}
	return nil
}
