//go:build int_test

package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSGiBundleList(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()

	response, err := instance.OSGI().BundleManager().List()
	a.Nil(err, "cannot read bundle list properly")
	a.NotEmpty(response.List, "bundle list should not be empty")
	a.NotEmpty(response.Status, "bundle status message should not be empty")

	a.False(response.StatusUnknown(), "bundle status should not be unknown")
}

func TestOSGiEventList(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()

	response, err := instance.OSGI().EventManager().List()
	a.Nil(err, "cannot read event list properly")
	a.NotEmpty(response.List, "event list should not be empty")

	a.False(response.StatusUnknown(), "event status should not be unknown")
}
