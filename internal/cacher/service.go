package cacher

import (
	"strings"
	"sync"
)

// InMem is a simple in-memory cache with string keys and values.
type InMem struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMem() *InMem {
	return &InMem{
		data: make(map[string]string),
	}
}

func (c *InMem) Set(key, value string) {
	c.mu.Lock()
	c.data[key] = value
	c.mu.Unlock()
}

func (c *InMem) Get(key string) (string, bool) {
	c.mu.RLock()
	value, ok := c.data[key]
	c.mu.RUnlock()
	return value, ok
}

func (c *InMem) DeletePrefix(prefix string) {
	c.mu.Lock()
	for key := range c.data {
		if strings.HasPrefix(key, prefix) {
			delete(c.data, key)
		}
	}
	c.mu.Unlock()
}
