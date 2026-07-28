package pkg

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceCacheGetCachesValue(t *testing.T) {
	t.Parallel()

	cache := &InstanceCache{Enabled: true}
	loads := 0
	loader := func() (string, error) {
		loads++
		return "value", nil
	}

	for i := 0; i < 3; i++ {
		value, err := InstanceCacheGet(cache, "key", loader)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
	}
	assert.Equal(t, 1, loads)
}

func TestInstanceCacheGetDoesNotCacheError(t *testing.T) {
	t.Parallel()

	cache := &InstanceCache{Enabled: true}
	loads := 0
	loader := func() (string, error) {
		loads++
		if loads == 1 {
			return "", fmt.Errorf("instance not running yet")
		}
		return "value", nil
	}

	_, err := InstanceCacheGet(cache, "key", loader)
	assert.Error(t, err)

	value, err := InstanceCacheGet(cache, "key", loader)
	assert.NoError(t, err)
	assert.Equal(t, "value", value)
	assert.Equal(t, 2, loads)
}

func TestInstanceCacheGetDisabled(t *testing.T) {
	t.Parallel()

	cache := &InstanceCache{Enabled: false}
	loads := 0
	loader := func() (string, error) {
		loads++
		return "value", nil
	}

	for i := 0; i < 3; i++ {
		value, err := InstanceCacheGet(cache, "key", loader)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
	}
	assert.Equal(t, 3, loads)
}

func TestInstanceCacheClear(t *testing.T) {
	t.Parallel()

	cache := &InstanceCache{Enabled: true}
	loads := 0
	loader := func() (string, error) {
		loads++
		return "value", nil
	}

	_, _ = InstanceCacheGet(cache, "key", loader)
	cache.Clear()
	_, _ = InstanceCacheGet(cache, "key", loader)

	assert.Equal(t, 2, loads)
}

func TestInstanceCacheGetConcurrently(t *testing.T) {
	t.Parallel()

	cache := &InstanceCache{Enabled: true}
	var mutex sync.Mutex
	loads := 0
	loader := func() (string, error) {
		mutex.Lock()
		defer mutex.Unlock()
		loads++
		return "value", nil
	}

	var group sync.WaitGroup
	for i := 0; i < 50; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			value, err := InstanceCacheGet(cache, "key", loader)
			assert.NoError(t, err)
			assert.Equal(t, "value", value)
		}()
	}
	group.Wait()

	assert.Equal(t, 1, loads)
}
