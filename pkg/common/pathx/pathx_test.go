package pathx

import (
	"os"
	"path/filepath"
	"testing"
)

// The pattern is built with the OS separators, as callers do on Windows from paths given by the user.
func TestGlobSomeWithOsSeparators(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "all", "target")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"app.all-1.0.0.zip", "app.all-1.0.1.zip", "other.zip"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	path, err := GlobSome(filepath.Join(dir, "app.all-*.zip"))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "app.all-1.0.1.zip"); path != want {
		t.Errorf("GlobSome() = %q, want %q", path, want)
	}
}

func TestGlobAllInCurrentDir(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("app.zip", nil, 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := GlobAll("*.zip")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0] != "app.zip" {
		t.Errorf("GlobAll() = %q, want [\"app.zip\"]", paths)
	}
}
