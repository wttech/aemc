//go:build int_test

package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoSaveThenRead(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.NewAem()
	instance := aem.InstanceManager().NewLocalAuthor()
	err := instance.Repo().Save("/var/repo_int_test_saveThenRead", map[string]any{
		"integer": 123,
		"float":   12.5,
		"bool":    true,
		"strings": []string{"a", "b", "c"},
		"string":  "hello world",
	})
	a.Nil(err)
	props, err := instance.Repo().Read("/var/repo_int_test_saveThenRead")
	a.Nil(err)
	a.Equal(true, props["bool"])
	a.Equal("12.5", props["float"])
	a.Equal(123.0, props["integer"])
	a.Equal("hello world", props["string"])
	a.Equal([]interface{}{"a", "b", "c"}, props["strings"])
}

func TestRepoSaveThenRemoveProp(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.NewAem()
	instance := aem.InstanceManager().NewLocalAuthor()
	err := instance.Repo().Save("/var/repo_int_test_removeProp", map[string]any{
		"first":  "1",
		"second": "2",
	})
	a.Nil(err)
	props, err := instance.Repo().Read("/var/repo_int_test_removeProp")
	_, ok := props["second"]
	a.True(ok)
	err = instance.Repo().Save("/var/repo_int_test_removeProp", map[string]any{
		"first":  "1",
		"second": nil,
	})
	props, err = instance.Repo().Read("/var/repo_int_test_removeProp")
	a.Nil(err)
	a.Equal("1", props["first"])
	_, ok = props["second"]
	a.False(ok)
}

func TestRepoReadChildren(t *testing.T) {
	t.Parallel()

	a := assert.New(t)
	aem := pkg.NewAem()

	instance := aem.InstanceManager().NewLocalAuthor()
	children, err := instance.Repo().Node("/content").Children()
	a.Nil(err)
	a.NotEmpty(children)
}

func TestRepoReadParents(t *testing.T) {
	t.Parallel()

	a := assert.New(t)
	aem := pkg.NewAem()

	instance := aem.InstanceManager().NewLocalAuthor()
	parents := instance.Repo().Node("/content/dam/projects").Parents()
	a.Len(parents, 2)
}

func TestRepoTraverse(t *testing.T) {
	t.Parallel()

	a := assert.New(t)
	aem := pkg.NewAem()

	instance := aem.InstanceManager().NewLocalAuthor()
	it := instance.Repo().Node("/etc/dam").Traversor()

	traversed := 0
	for {
		_, ok, err := it.Next()
		if !ok {
			break
		}
		a.Nil(err)
		traversed++
	}
	a.GreaterOrEqual(traversed, 20)
}
