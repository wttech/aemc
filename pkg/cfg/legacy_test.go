package cfg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLegacyTemplateUsed(t *testing.T) {
	assert.True(t, legacyTemplateUsed([]byte("active: [[.Env.ACTIVE | default true ]]")))
	assert.False(t, legacyTemplateUsed([]byte("# env var '[[.Env.JAVA_HOME]]'\nactive: ${ACTIVE:-true}")))
	assert.False(t, legacyTemplateUsed([]byte("include: ['com.mysite.*']")))
}

func TestMigrateLegacy(t *testing.T) {
	content := `instance:
  config:
    local_author:
      active: [[.Env.AEM_AUTHOR_ACTIVE | default true ]]
      http_url: [[.Env.AEM_AUTHOR_HTTP_URL | default "http://127.0.0.1:4502" ]]
      password: [[.Env.AEM_AUTHOR_PASSWORD]]
      price: 5$
      jvm_opts:
        - -Djava.io.tmpdir=[[canonicalPath .Path "aem/home/tmp"]]
      name: [[ .Env.NAME | upper ]]
  local:
    backup_dir: aem/home/var/backup
base:
  tool_dir: aem/home/opt
vendor:
  java:
    # env var '[[.Env.JAVA_HOME]]'
    url: "jdk_[[.Arch]]_[[.Os]].[[.ArchiveExt]]"
`
	expected := `instance:
  config:
    local_author:
      active: ${AEM_AUTHOR_ACTIVE:-true}
      http_url: ${AEM_AUTHOR_HTTP_URL:-http://127.0.0.1:4502}
      password: ${AEM_AUTHOR_PASSWORD}
      price: 5$$
      jvm_opts:
        - -Djava.io.tmpdir=${AEM_PROJECT_DIR}/aem/home/tmp
      name: [[ .Env.NAME | upper ]]
  local:
    backup_dir: aem/home/var/backup
base:
  tool_dir: aem/home/opt
vendor:
  java:
    # env var '${JAVA_HOME}'
    url: "jdk_${AEM_ARCH}_${AEM_OS}.${AEM_ARCHIVE_EXT}"
`
	result := MigrateLegacy(content)
	assert.Equal(t, expected, result.Content)
	assert.Equal(t, []string{
		"line 10: template expression without a counterpart: name: [[ .Env.NAME | upper ]]",
		"line 12: setting 'instance.local.backup_dir' is removed, use 'instance.local.backup.dir'",
		"line 14: set 'base.tool_dir' to 'aem/home/opt/vendor'",
	}, result.Unresolved)
}
