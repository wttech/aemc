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
	testFilterFile(t, "vault/META-INF/vault/filter.xml", "output/filter_roots.xml",
		map[string]any{
			"FilterRoots": []string{"/apps/my_site", "/content/my_site"},
		},
	)
}

func TestOnlyOneContent(t *testing.T) {
	testFilterFile(t, "vault/META-INF/vault/filter.xml", "output/only_one_content.xml",
		map[string]any{
			"FilterRoots":    []string{"/apps/my_site", "/content/my_site"},
			"OnlyOneContent": true,
		},
	)
}
