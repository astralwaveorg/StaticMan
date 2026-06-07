package cache

import (
	"sync"
	"time"
)

// Entry 缓存条目
type Entry struct {
	Value      interface{}
	ExpiresAt  time.Time
}

// Cache 内存缓存
type Cache struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New 创建新缓存
func New() *Cache {
	return &Cache{
		entries: make(map[string]Entry),
	}
}

// Get 获取缓存值
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.ExpiresAt) {
		c.Delete(key)
		return nil, false
	}
	return entry.Value, true
}

// Set 设置缓存值
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	c.entries[key] = Entry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Delete 删除缓存
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// DeletePrefix 删除匹配前缀的所有缓存
func (c *Cache) DeletePrefix(prefix string) {
	c.mu.Lock()
	for k := range c.entries {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}

// Clear 清空所有缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]Entry)
	c.mu.Unlock()
}

// Flush 清空所有缓存 (Clear 的别名)
func (c *Cache) Flush() {
	c.Clear()
}

// Stats 返回缓存统计
func (c *Cache) Stats() (total int, expired int) {
	now := time.Now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, entry := range c.entries {
		total++
		if now.After(entry.ExpiresAt) {
			expired++
		}
	}
	return
}
