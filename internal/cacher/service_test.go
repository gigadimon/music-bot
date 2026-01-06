package cacher

import (
	"fmt"
	"sync"
	"testing"
)

func TestInMemSetGet(t *testing.T) {
	cache := NewInMem()

	cache.Set("a", "1")
	got, ok := cache.Get("a")
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if got != "1" {
		t.Fatalf("expected value %q, got %q", "1", got)
	}

	_, ok = cache.Get("missing")
	if ok {
		t.Fatalf("expected missing key to return ok=false")
	}
}

func TestInMemConcurrentAccess(t *testing.T) {
	cache := NewInMem()

	const workers = 32
	const iterations = 200

	var wg sync.WaitGroup
	errCh := make(chan error, workers*iterations)

	for i := 0; i < workers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			for j := 0; j < iterations; j++ {
				value := fmt.Sprintf("val-%d-%d", i, j)
				cache.Set(key, value)
				got, ok := cache.Get(key)
				if !ok {
					errCh <- fmt.Errorf("missing key %q", key)
					continue
				}
				if got == "" {
					errCh <- fmt.Errorf("empty value for key %q", key)
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}
}

func TestInMemDeletePrefix(t *testing.T) {
	cache := NewInMem()

	cache.Set("req#0", "a")
	cache.Set("req#1", "b")
	cache.Set("other#0", "c")

	cache.DeletePrefix("req")

	if _, ok := cache.Get("req#0"); ok {
		t.Fatalf("expected req#0 to be deleted")
	}
	if _, ok := cache.Get("req#1"); ok {
		t.Fatalf("expected req#1 to be deleted")
	}
	if _, ok := cache.Get("other#0"); !ok {
		t.Fatalf("expected other#0 to remain")
	}
}
