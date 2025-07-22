package main

import (
	"fmt"
	"sync"
	"time"
)

// Problem: Thread-Safe LRU Cache with TTL
// Implement a thread-safe LRU (Least Recently Used) cache with Time-To-Live (TTL) expiration that supports concurrent access and automatic cleanup.

// Requirements:
// 1. LRU eviction when capacity is exceeded
// 2. TTL-based expiration for entries
// 3. Thread-safe concurrent access
// 4. Automatic cleanup of expired entries
// 5. Support for generic key-value types

type Cache[K comparable, V any] interface {
	// Get retrieves a value by key, returns value and true if found
	Get(key K) (V, bool)

	// Set stores a key-value pair with TTL
	Set(key K, value V, ttl time.Duration)

	// Delete removes a key from cache
	Delete(key K) bool

	// Size returns current number of entries
	Size() int

	// Clear removes all entries
	Clear()

	// Stats returns cache statistics
	Stats() CacheStats
}

type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Expirations int64
	Size        int
	Capacity    int
}

type CacheEntry[V any] struct {
	Value     V
	ExpiresAt time.Time
	// TODO: Add fields for LRU tracking
}

type lruCache[K comparable, V any] struct {
	capacity        int
	mu              sync.RWMutex
	head            *node[K, V]
	tail            *node[K, V]
	cleanupInterval time.Duration
	items           map[K]*node[K, V]
	stopCleanup     chan struct{}
	stats           CacheStats

	// TODO: Implement doubly linked list for LRU
	// head, tail *node[K, V]

	// TODO: Add mutex for thread safety
	// mu sync.RWMutex

	// TODO: Add statistics tracking
	// stats CacheStats

	// TODO: Add cleanup mechanism
	// cleanupInterval time.Duration
	//
}

type node[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
	prev      *node[K, V]
	next      *node[K, V]
	// TODO: Add prev/next pointers for doubly linked list
	// prev, next *node[K, V]
}

func NewCache[K comparable, V any](capacity int) Cache[K, V] {
	// TODO: Initialize cache with:
	// - map for O(1) key lookup
	// - doubly linked list for O(1) LRU operations
	// - background cleanup goroutine
	lhead := &node[K, V]{}
	ltail := &node[K, V]{}
	lhead.next = ltail
	ltail.prev = lhead
	cache := lruCache[K, V]{
		items:           make(map[K]*node[K, V]),
		capacity:        capacity,
		cleanupInterval: 10 * time.Second,
		stopCleanup:     make(chan struct{}),
		head:            lhead,
		tail:            ltail,
	}
	return &cache
}

func (c *lruCache[K, V]) Get(key K) (V, bool) {
	// TODO: Implement Get with:
	// - Thread safety (read lock)
	// - TTL expiration check
	// - LRU update (move to front)
	// - Statistics update
	var zero V
	c.mu.RLock()
	val, found := c.items[key]
	if !found {
		c.mu.RUnlock()
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return zero, false
	}

	if c.isExpired(val) {
		c.mu.RUnlock()
		c.mu.Lock()
		if c.isExpired(val) {
			c.removeNode(val)
			c.stats.Expirations++
			c.stats.Misses++
		}
		c.mu.Unlock()
		return zero, false
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.moveToFront(val)
	c.stats.Hits++
	return val.value, true
}
func (c *lruCache[K, V]) addToFront(n *node[K, V]) {
	n.prev = c.head
	n.next = c.head.next
	c.head.next.prev = n
	c.head.next = n
}
func (c *lruCache[K, V]) Set(key K, value V, ttl time.Duration) {
	// TODO: Implement Set with:
	// - Thread safety (write lock)
	// - Check if key exists (update vs insert)
	// - LRU eviction if at capacity
	// - Move to front of LRU list
	// - Statistics update
	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.items[key]; ok {
		val.value = value
		val.expiresAt = time.Now().Add(ttl)
		c.moveToFront(val)
		return
	}

	if len(c.items) >= c.capacity {
		c.removeLRU()
		c.stats.Evictions++
	}
	NewNode := &node[K, V]{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	c.items[key] = NewNode
	c.addToFront(NewNode)
	c.stats.Hits++
}

func (c *lruCache[K, V]) Delete(key K) bool {
	// TODO: Implement Delete with:
	// - Thread safety (write lock)
	// - Remove from map and linked list
	// - Statistics update
	if _, ok := c.items[key]; ok {
		delete(c.items, key)
		return true
	}
	c.stats.Evictions++
	return false
}

func (c *lruCache[K, V]) Size() int {
	// TODO: Thread-safe size check
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *lruCache[K, V]) Clear() {
	// TODO: Remove all entries thread-safely
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*node[K, V])
	c.head.next = c.tail
	c.tail.prev = c.head
}

func (c *lruCache[K, V]) Stats() CacheStats {
	// TODO: Return current statistics
	c.stats.Size = len(c.items)
	c.stats.Capacity = c.capacity
	return c.stats
}

// Helper methods for doubly linked list operations
func (c *lruCache[K, V]) moveToFront(n *node[K, V]) {
	// TODO: Move node to front of LRU list
	n.prev.next = n.next
	n.next.prev = n.prev
	c.addToFront(n)
}

func (c *lruCache[K, V]) removeLRU() {
	// TODO: Remove least recently used item
	c.removeNode(c.tail.prev)
}

func (c *lruCache[K, V]) removeNode(n *node[K, V]) {
	// TODO: Remove specific node from linked list
	n.prev.next = n.next
	n.next.prev = n.prev
	delete(c.items, n.key)
}

func (c *lruCache[K, V]) startCleanup() {
	// TODO: Start background cleanup goroutine
	// - Periodically scan for expired entries
	// - Remove expired items
	// - Handle graceful shutdown
	ticker := time.NewTicker(c.cleanupInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				for key, n := range c.items {
					if c.isExpired(n) {
						fmt.Printf("Cleanup: removing expired key '%v'\n", key)
						c.removeNode(n)
						c.stats.Expirations++
					}
				}
				c.mu.Unlock()
			case <-c.stopCleanup:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *lruCache[K, V]) isExpired(n *node[K, V]) bool {
	// TODO: Check if node has expired
	if n.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(n.expiresAt)
}

func main() {
	// Test 1: Basic LRU functionality
	cache := NewCache[string, int](3)

	cache.Set("a", 1, 5*time.Second)
	cache.Set("b", 2, 5*time.Second)
	cache.Set("c", 3, 5*time.Second)

	// Should evict "a" (least recently used)
	cache.Set("d", 4, 5*time.Second)

	if _, found := cache.Get("a"); found {
		fmt.Println("ERROR: 'a' should have been evicted")
	}

	// Test 2: TTL expiration
	cache.Set("expire", 100, 1*time.Second)

	if val, found := cache.Get("expire"); !found || val != 100 {
		fmt.Println("ERROR: Should find 'expire' immediately")
	}

	time.Sleep(2 * time.Second)

	if _, found := cache.Get("expire"); found {
		fmt.Println("ERROR: 'expire' should have expired")
	}

	// Test 3: LRU order update on access
	cache.Clear()
	cache.Set("x", 1, 10*time.Second)
	cache.Set("y", 2, 10*time.Second)
	cache.Set("z", 3, 10*time.Second)

	// Access "x" to make it most recently used
	cache.Get("x")

	// Add new item, should evict "y" (now LRU)
	cache.Set("w", 4, 10*time.Second)

	if _, found := cache.Get("y"); found {
		fmt.Println("ERROR: 'y' should have been evicted")
	}

	// Test 4: Concurrent access
	testConcurrentAccess(cache)

	// Test 5: Statistics
	stats := cache.Stats()
	fmt.Printf("Cache Stats: %+v\n", stats)
}

func testConcurrentAccess(cache Cache[string, int]) {
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				cache.Set(key, j, 2*time.Second)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d-%d", id%10, j)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
}
