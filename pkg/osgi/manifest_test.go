package osgi

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestJAR(t *testing.T, manifest string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test-bundle.jar")
	f, err := os.Create(path)
	assert.NoError(t, err)
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	entry, err := w.Create("META-INF/MANIFEST.MF")
	assert.NoError(t, err)
	_, err = entry.Write([]byte(manifest))
	assert.NoError(t, err)
	return path
}

func TestReadBundleManifest(t *testing.T) {
	t.Parallel()

	jar := createTestJAR(t, "Manifest-Version: 1.0\r\nBundle-SymbolicName: com.example.test\r\nBundle-Version: 1.2.3\r\n")

	manifest, err := ReadBundleManifest(jar)

	assert.NoError(t, err)
	assert.Equal(t, "com.example.test", manifest.SymbolicName)
	assert.Equal(t, "1.2.3", manifest.Version)
}

func TestReadBundleManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, err := ReadBundleManifest("/nonexistent/bundle.jar")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read OSGi bundle manifest from file")
}

func TestReadBundleManifestInvalidJAR(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "not-a-jar.jar")
	err := os.WriteFile(path, []byte("not a jar file"), 0644)
	assert.NoError(t, err)

	_, err = ReadBundleManifest(path)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read OSGi bundle manifest from file")
}

func TestBundleManifestMarshalText(t *testing.T) {
	t.Parallel()

	manifest := BundleManifest{SymbolicName: "com.example.test", Version: "1.0.0"}

	text := manifest.MarshalText()

	assert.Contains(t, text, "com.example.test")
	assert.Contains(t, text, "1.0.0")
}
