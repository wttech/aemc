package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceNewLocalAuthor(t *testing.T) {
	t.Parallel()

	aem := pkg.NewAem()
	instance := aem.InstanceManager().NewLocalAuthor()

	assert.Equal(t, "http://localhost:4502", instance.HTTP().BaseURL())
	assert.Equal(t, "admin", instance.User())
	assert.Equal(t, "admin", instance.Password())
}

func TestInstanceNewLocalPublish(t *testing.T) {
	t.Parallel()

	aem := pkg.NewAem()
	instance := aem.InstanceManager().NewLocalPublish()

	assert.Equal(t, "http://localhost:4503", instance.HTTP().BaseURL())
	assert.Equal(t, "admin", instance.User())
	assert.Equal(t, "admin", instance.Password())
}
