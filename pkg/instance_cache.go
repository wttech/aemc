package pkg

import (
	"sync"
)

// Cache keys are centralized here to avoid accidental collisions between managers sharing the instance cache.
const (
	cacheKeyStatusSystemProps   = "status.system_props"
	cacheKeyStatusSlingProps    = "status.sling_props"
	cacheKeyStatusSlingSettings = "status.sling_settings"
	cacheKeyStatusAemVersion    = "status.aem_version"
)

// InstanceCache holds in-memory values that are constant for a running instance
// (like system properties or AEM version) to avoid repeating HTTP requests.
// Values are never persisted; the cache lives as long as the Instance object.
type InstanceCache struct {
	instance *Instance

	Enabled bool

	mutex   sync.Mutex
	entries map[string]*instanceCacheEntry
}

type instanceCacheEntry struct {
	mutex  sync.Mutex
	loaded bool
	value  any
}

func NewInstanceCache(i *Instance) *InstanceCache {
	cv := i.manager.aem.config.Values()

	return &InstanceCache{
		instance: i,

		Enabled: cv.GetBool("instance.cache.enabled"),

		entries: map[string]*instanceCacheEntry{},
	}
}

// Clear forgets all cached values. Needed when instance state changes (start/stop).
func (c *InstanceCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries = map[string]*instanceCacheEntry{}
}

func (c *InstanceCache) entry(key string) *instanceCacheEntry {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.entries == nil {
		c.entries = map[string]*instanceCacheEntry{}
	}
	result, ok := c.entries[key]
	if !ok {
		result = &instanceCacheEntry{}
		c.entries[key] = result
	}
	return result
}

// InstanceCacheGet returns a cached value or computes it using the loader.
// Errors are never cached, so failures occurring while an instance is still
// starting up do not stick for the rest of the process lifetime.
// It is a package-level function as Go methods cannot have type parameters.
func InstanceCacheGet[T any](c *InstanceCache, key string, loader func() (T, error)) (T, error) {
	if c == nil || !c.Enabled {
		return loader()
	}
	e := c.entry(key)
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.loaded {
		return e.value.(T), nil
	}
	value, err := loader()
	if err != nil {
		var zero T
		return zero, err
	}
	e.value = value
	e.loaded = true
	return value, nil
}
