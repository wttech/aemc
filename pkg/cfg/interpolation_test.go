package cfg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func testMapping(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func TestInterpolateYAML(t *testing.T) {
	content := `
# comment with ${NOT_INTERPOLATED} and [[ legacy ]]
instance:
  author:
    active: ${ACTIVE:-true}
    http_url: ${URL:-http://127.0.0.1:4502}
    user: "${USER_NAME}"
    port: ${PORT:-4502}
    quoted_bool: "${ACTIVE:-true}"
    password: ${PASSWORD}
    escaped: price$$
    "${KEY}": kept
    jvm_opts:
      - -Djava.io.tmpdir=${DIR}/aem/home/tmp
      - -Xmx${XMX:-2g}
    include: ['${PACKAGE:-com.mysite}.*', 'com.day.*']
`
	values, err := interpolateYAML([]byte(content), testMapping(map[string]string{
		"USER_NAME": "admin",
		"PASSWORD":  "p@ss: #word",
		"DIR":       "/project",
		"KEY":       "changed",
	}))
	require.NoError(t, err)

	author := values["instance"].(map[string]any)["author"].(map[string]any)
	assert.Equal(t, true, author["active"])
	assert.Equal(t, "http://127.0.0.1:4502", author["http_url"])
	assert.Equal(t, "admin", author["user"])
	assert.Equal(t, 4502, author["port"])
	assert.Equal(t, "true", author["quoted_bool"])
	assert.Equal(t, "p@ss: #word", author["password"])
	assert.Equal(t, "price$", author["escaped"])
	assert.Equal(t, "kept", author["${KEY}"])
	assert.Equal(t, []any{"-Djava.io.tmpdir=/project/aem/home/tmp", "-Xmx2g"}, author["jvm_opts"])
	assert.Equal(t, []any{"com.mysite.*", "com.day.*"}, author["include"])
}

func TestInterpolateYAMLRequired(t *testing.T) {
	content := "instance:\n  author:\n    http_url: ${URL:?AEM author URL is required}\n"
	_, err := interpolateYAML([]byte(content), testMapping(map[string]string{}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 3")
	assert.Contains(t, err.Error(), "AEM author URL is required")
}

func TestInterpolateYAMLEmpty(t *testing.T) {
	values, err := interpolateYAML([]byte(""), testMapping(map[string]string{}))
	require.NoError(t, err)
	assert.Empty(t, values)
}

func TestInterpolateJSON(t *testing.T) {
	values, err := interpolateYAML([]byte(`{"log": {"level": "${LEVEL:-info}"}}`), testMapping(map[string]string{}))
	require.NoError(t, err)
	assert.Equal(t, "info", values["log"].(map[string]any)["level"])
}
