package cacher

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func newTestCache(t *testing.T) *Redis {
	t.Helper()
	server := miniredis.RunT(t)
	SetConfig(Config{
		RedisAddr:     server.Addr(),
		RedisPassword: "",
	})
	return NewRedis(0, 0)
}

func TestRedisSetGet(t *testing.T) {
	cache := newTestCache(t)

	if err := cache.Set(context.Background(), "a", "1"); err != nil {
		t.Fatalf("set error: %v", err)
	}
	got, ok, err := cache.Get(context.Background(), "a")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if got != "1" {
		t.Fatalf("expected value %q, got %q", "1", got)
	}

	_, ok, err = cache.Get(context.Background(), "missing")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if ok {
		t.Fatalf("expected missing key to return ok=false")
	}
}

func TestRedisConcurrentAccess(t *testing.T) {
	cache := newTestCache(t)

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
				if err := cache.Set(context.Background(), key, value); err != nil {
					errCh <- fmt.Errorf("set error for key %q: %w", key, err)
					continue
				}
				got, ok, err := cache.Get(context.Background(), key)
				if err != nil {
					errCh <- fmt.Errorf("get error for key %q: %w", key, err)
					continue
				}
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

func TestRedisDeletePrefix(t *testing.T) {
	cache := newTestCache(t)

	if err := cache.Set(context.Background(), "req#0", "a"); err != nil {
		t.Fatalf("set error: %v", err)
	}
	if err := cache.Set(context.Background(), "req#1", "b"); err != nil {
		t.Fatalf("set error: %v", err)
	}
	if err := cache.Set(context.Background(), "other#0", "c"); err != nil {
		t.Fatalf("set error: %v", err)
	}

	if err := cache.DeletePrefix(context.Background(), "req"); err != nil {
		t.Fatalf("delete prefix error: %v", err)
	}

	if _, ok, err := cache.Get(context.Background(), "req#0"); err != nil {
		t.Fatalf("get error: %v", err)
	} else if ok {
		t.Fatalf("expected req#0 to be deleted")
	}
	if _, ok, err := cache.Get(context.Background(), "req#1"); err != nil {
		t.Fatalf("get error: %v", err)
	} else if ok {
		t.Fatalf("expected req#1 to be deleted")
	}
	if _, ok, err := cache.Get(context.Background(), "other#0"); err != nil {
		t.Fatalf("get error: %v", err)
	} else if !ok {
		t.Fatalf("expected other#0 to remain")
	}
}
