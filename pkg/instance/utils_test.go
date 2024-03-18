package instance

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInstanceMatchers(t *testing.T) {
	t.Parallel()

	assert.False(t, Match("local_publish", "com.acme.aem.core.ExampleService", "com.acme.aem.core.OtherService"))
	assert.False(t, Match("local_publish", "com.acme.aem.core.OtherService", "com.acme.aem.core.ExampleService"))
	assert.True(t, Match("local_publish", "com.acme.aem.core.ExampleService", "com.acme.aem.core.ExampleService"))
	assert.True(t, Match("local_publish", "com.acme.aem.core.ExampleService", "local_publish:com.acme.aem.core.ExampleService"))
	assert.True(t, Match("local_publish", "com.acme.aem.core.ExampleService", "local_publish:com.acme.aem.core.*Service"))
	assert.True(t, Match("local_publish", "com.acme.aem.core.ExampleService", "*_publish*:com.acme.aem.core.*Service"))
	assert.True(t, Match("local_publish", "com.acme.aem.core.ExampleService", "local_publish,int_publish_1,int_publish_2:com.acme.aem.core.*Service"))
}
