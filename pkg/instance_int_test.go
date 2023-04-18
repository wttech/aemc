//go:build int_test

package pkg_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wttech/aemc/pkg"
	"testing"
)

func TestInstanceTimeLocation(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()
	instanceLocation := instance.TimeLocation()

	a.NotEmpty(instanceLocation.String())
}
