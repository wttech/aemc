package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBootstrapFlags(t *testing.T) {
	cases := map[string]struct {
		args     []string
		expected bootstrapFlags
	}{
		"none":            {[]string{"instance", "status"}, bootstrapFlags{}},
		"short cd":        {[]string{"-C", "/opt/aem", "instance", "status"}, bootstrapFlags{cd: "/opt/aem"}},
		"long both":       {[]string{"instance", "status", "--cd=/opt/aem", "--config", "my.yml"}, bootstrapFlags{cd: "/opt/aem", config: "my.yml"}},
		"among unknown":   {[]string{"pkg", "deploy", "-A", "--output-format", "yml", "--config", "my.yml", "--file", "a.zip"}, bootstrapFlags{config: "my.yml"}},
		"after help flag": {[]string{"-h", "--config", "my.yml"}, bootstrapFlags{config: "my.yml"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.expected, parseBootstrapFlags(c.args))
		})
	}
}
