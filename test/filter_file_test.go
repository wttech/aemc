//go:build int_test

package test

import (
	"github.com/wttech/aemc/pkg/common/tplx"
	"github.com/wttech/aemc/pkg/pkg"
	"testing"
)

func testFilterFile(t *testing.T, filterFile string, expectedFile string, data map[string]any) {
	bytes, err := pkg.VaultFS.ReadFile(filterFile)
	if err != nil {
		t.Fatalf("%v %v", bytes, err)
	}
	expected, err := VaultFS.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("%v %v", bytes, err)
	}
	actual, err := tplx.RenderString(string(bytes), data)
	if actual != string(expected) {
		t.Errorf("RenderString(%s, %v) = %s; want %s", string(bytes), data, actual, expected)
	}
}

func TestFilterRoots(t *testing.T) {
	testFilterFile(t, "vault/META-INF/vault/filter.xml", "resources/filter_roots.xml",
		map[string]any{
			"FilterRoots": []string{"/apps/my_site", "/content/my_site"},
		},
	)
}

func TestFilterRootExcludes(t *testing.T) {
	testFilterFile(t, "vault/META-INF/vault/filter.xml", "resources/exclude_patterns.xml",
		map[string]any{
			"FilterRoots":        []string{"/apps/my_site", "/content/my_site"},
			"FilterRootExcludes": []string{"/apps/my_site/cq:dialog(/.*)?", "/apps/my_site/rep:policy(/.*)?"},
		},
	)
}

func TestFilterRootsUpdate(t *testing.T) {
	testFilterFile(t, "vault/META-INF/vault/filter.xml", "resources/filter_roots_update.xml",
		map[string]any{
			"FilterRoots": []string{"/apps/my_site", "/content/my_site"},
			"FilterMode":  "update",
		},
	)
}

func TestFilterRootExcludesUpdate(t *testing.T) {
	testFilterFile(t, "vault/META-INF/vault/filter.xml", "resources/exclude_patterns_update.xml",
		map[string]any{
			"FilterRoots":        []string{"/apps/my_site", "/content/my_site"},
			"FilterRootExcludes": []string{"/apps/my_site/cq:dialog(/.*)?", "/apps/my_site/rep:policy(/.*)?"},
			"FilterMode":         "update",
		},
	)
}
