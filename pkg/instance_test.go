package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceNewLocalAuthor(t *testing.T) {
	t.Parallel()

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()

	assert.Equal(t, "http://127.0.0.1:4502", instance.HTTP().BaseURL())
	assert.Equal(t, "admin", instance.User())
	assert.Equal(t, "admin", instance.Password())
}

func TestInstanceNewLocalPublish(t *testing.T) {
	t.Parallel()

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalPublish()

	assert.Equal(t, "http://127.0.0.1:4503", instance.HTTP().BaseURL())
	assert.Equal(t, "admin", instance.User())
	assert.Equal(t, "admin", instance.Password())
}
