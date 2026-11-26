package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneratedWrappersUpToDate fails when the committed wrappers differ from the template; run 'go generate ./pkg/project'.
func TestGeneratedWrappersUpToDate(t *testing.T) {
	for _, w := range Wrappers() {
		expected, err := Render(w)
		require.NoError(t, err)
		actual, err := os.ReadFile(filepath.Join("..", OutputDir, w.File))
		require.NoError(t, err)
		assert.Equal(t, expected, string(actual), "wrapper '%s' is out of date, run 'go generate ./pkg/project'", w.File)
	}
}
