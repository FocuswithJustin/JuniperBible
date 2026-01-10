package cache

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ttl := 5 * time.Minute
	cache := New[string, int](ttl)

	if cache == nil {
		t.Fatal("New returned nil")
	}
	if cache.ttl != ttl {
		t.Errorf("TTL mismatch: got %v, want %v", cache.ttl, ttl)
	}
	if cache.data == nil {
		t.Error("data map not initialized")
	}
	if !cache.timestamp.IsZero() {
		t.Error("timestamp should be zero on initialization")
	}
}

func TestSetAndGet(t *testing.T) {
	cache := New[string, int](1 * time.Minute)

	// Set a value
	cache.Set("key1", 42)

	// Get the value
	value, ok := cache.Get("key1")
	if !ok {
		t.Fatal("Get returned ok=false for existing key")
	}
	if value != 42 {
		t.Errorf("Get returned wrong value: got %d, want 42", value)
	}

	// Get non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Get returned ok=true for non-existent key")
	}
}

func TestGetExpired(t *testing.T) {
	cache := New[string, int](50 * time.Millisecond)

	// Set a value
	cache.Set("key1", 42)

	// Verify it's cached
	value, ok := cache.Get("key1")
	if !ok || value != 42 {
		t.Fatal("Initial Get failed")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should return false after expiration
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Get returned ok=true for expired cache")
	}
}

func TestSetAll(t *testing.T) {
	cache := New[string, int](1 * time.Minute)

	data := map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	}

	cache.SetAll(data)

	// Verify all values
	for key, want := range data {
		got, ok := cache.Get(key)
		if !ok {
			t.Errorf("Get(%q) returned ok=false", key)
			continue
		}
		if got != want {
			t.Errorf("Get(%q) = %d, want %d", key, got, want)
		}
	}
}

func TestGetAll(t *testing.T) {
	cache := New[string, int](1 * time.Minute)

	// Empty cache
	all := cache.GetAll()
	if all != nil {
		t.Error("GetAll on empty cache should return nil (expired)")
	}

	// Populate cache
	data := map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	}
	cache.SetAll(data)

	// GetAll should return a copy
	all = cache.GetAll()
	if all == nil {
		t.Fatal("GetAll returned nil for valid cache")
	}
	if len(all) != len(data) {
		t.Errorf("GetAll returned wrong count: got %d, want %d", len(all), len(data))
	}
	for key, want := range data {
		got, ok := all[key]
		if !ok {
			t.Errorf("GetAll missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("GetAll[%q] = %d, want %d", key, got, want)
		}
	}

	// Verify it's a copy by modifying it
	all["key1"] = 999
	value, _ := cache.Get("key1")
	if value == 999 {
		t.Error("GetAll did not return a copy - cache was modified")
	}
}

func TestGetAllExpired(t *testing.T) {
	cache := New[string, int](50 * time.Millisecond)

	cache.SetAll(map[string]int{"key1": 1})

	// Should work initially
	all := cache.GetAll()
	if all == nil {
		t.Fatal("GetAll returned nil for valid cache")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should return nil after expiration
	all = cache.GetAll()
	if all != nil {
		t.Error("GetAll returned non-nil for expired cache")
	}
}

func TestIsExpired(t *testing.T) {
	cache := New[string, int](50 * time.Millisecond)

	// New cache is expired (zero timestamp)
	if !cache.IsExpired() {
		t.Error("New cache should be expired")
	}

	// Set a value
	cache.Set("key1", 42)

	// Should not be expired immediately
	if cache.IsExpired() {
		t.Error("Cache should not be expired immediately after Set")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be expired now
	if !cache.IsExpired() {
		t.Error("Cache should be expired after TTL")
	}
}

func TestInvalidate(t *testing.T) {
	cache := New[string, int](1 * time.Minute)

	// Populate cache
	cache.SetAll(map[string]int{
		"key1": 1,
		"key2": 2,
	})

	// Verify it's populated
	if cache.Len() != 2 {
		t.Fatalf("Cache should have 2 items, got %d", cache.Len())
	}
	if cache.IsExpired() {
		t.Fatal("Cache should not be expired")
	}

	// Invalidate
	cache.Invalidate()

	// Should be empty and expired
	if cache.Len() != 0 {
		t.Errorf("Invalidated cache should be empty, got %d items", cache.Len())
	}
	if !cache.IsExpired() {
		t.Error("Invalidated cache should be expired")
	}

	// Get should fail
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Get should fail after Invalidate")
	}
}

func TestLen(t *testing.T) {
	cache := New[string, int](1 * time.Minute)

	if cache.Len() != 0 {
		t.Errorf("New cache should have length 0, got %d", cache.Len())
	}

	cache.Set("key1", 1)
	if cache.Len() != 1 {
		t.Errorf("After Set, length should be 1, got %d", cache.Len())
	}

	cache.Set("key2", 2)
	if cache.Len() != 2 {
		t.Errorf("After second Set, length should be 2, got %d", cache.Len())
	}

	cache.Invalidate()
	if cache.Len() != 0 {
		t.Errorf("After Invalidate, length should be 0, got %d", cache.Len())
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := New[int, string](1 * time.Minute)
	var wg sync.WaitGroup

	// Multiple goroutines writing
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Set(n, "value")
		}(i)
	}

	// Multiple goroutines reading
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Get(n)
		}(i)
	}

	// Some invalidations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Invalidate()
		}()
	}

	wg.Wait()

	// Test should complete without race conditions
}

func TestMultipleTypes(t *testing.T) {
	// Test string -> int
	intCache := New[string, int](1 * time.Minute)
	intCache.Set("answer", 42)
	val, ok := intCache.Get("answer")
	if !ok || val != 42 {
		t.Error("String->int cache failed")
	}

	// Test int -> string
	strCache := New[int, string](1 * time.Minute)
	strCache.Set(1, "one")
	str, ok := strCache.Get(1)
	if !ok || str != "one" {
		t.Error("Int->string cache failed")
	}

	// Test with struct values
	type Person struct {
		Name string
		Age  int
	}
	structCache := New[string, Person](1 * time.Minute)
	structCache.Set("alice", Person{Name: "Alice", Age: 30})
	person, ok := structCache.Get("alice")
	if !ok || person.Name != "Alice" || person.Age != 30 {
		t.Error("Struct cache failed")
	}
}

func TestTimestampUpdate(t *testing.T) {
	cache := New[string, int](1 * time.Minute)

	// Set initial value
	cache.Set("key1", 1)
	firstTimestamp := cache.timestamp

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Set another value
	cache.Set("key2", 2)
	secondTimestamp := cache.timestamp

	if !secondTimestamp.After(firstTimestamp) {
		t.Error("Timestamp should be updated on each Set")
	}

	// SetAll should also update timestamp
	time.Sleep(10 * time.Millisecond)
	cache.SetAll(map[string]int{"key3": 3})
	thirdTimestamp := cache.timestamp

	if !thirdTimestamp.After(secondTimestamp) {
		t.Error("Timestamp should be updated on SetAll")
	}
}

func TestZeroValue(t *testing.T) {
	cache := New[string, int](1 * time.Minute)

	// Store a zero value
	cache.Set("zero", 0)

	// Should be retrievable
	value, ok := cache.Get("zero")
	if !ok {
		t.Error("Get returned ok=false for zero value")
	}
	if value != 0 {
		t.Errorf("Get returned wrong zero value: got %d, want 0", value)
	}
}
