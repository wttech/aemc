package project

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMiseAemcVersion guards that the release bumped AEM Compose in every mise project file.
func TestMiseAemcVersion(t *testing.T) {
	files, err := fs.Glob(RunnerFiles, "runner/mise/*/mise.toml")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, file := range files {
		content, err := fs.ReadFile(RunnerFiles, file)
		require.NoError(t, err)
		assert.Contains(t, string(content), `"github:wttech/aemc" = "`+AemcVersion+`"`, file)
	}
}
