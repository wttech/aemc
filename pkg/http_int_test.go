//go:build int_test

package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceHTTPGetRequest(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.NewAem()
	instance := aem.InstanceManager().NewLocalAuthor()
	response, err := instance.HTTP().Request().Get("/system/console/bundles.json")

	a.Nil(err)
	a.Equal(http.StatusOK, response.StatusCode())
	a.Equal("application/json;charset=utf-8", response.Header().Get("Content-Type"))
}
