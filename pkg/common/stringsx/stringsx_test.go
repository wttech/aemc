package stringsx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wttech/aemc/pkg/common/stringsx"
)

func TestSlug(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	a.Equal("clean-with-prereqs", stringsx.Slug("Clean with Prereqs"))
	a.Equal("clean-with-prereqs", stringsx.Slug("  clean_with--prereqs! "))
	a.Equal("6-6-lts", stringsx.Slug("6.6 LTS"))
	a.Equal("", stringsx.Slug(" !? "))
}
