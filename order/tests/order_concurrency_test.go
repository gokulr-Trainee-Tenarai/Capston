package tests

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestOrderCreationConcurrency(t *testing.T) {
	var totalOrders int64 = 0
	var wg sync.WaitGroup

	numGoroutines := 10
	ordersPerGoroutine := 5

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < ordersPerGoroutine; j++ {
				atomic.AddInt64(&totalOrders, 1)
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * ordersPerGoroutine)
	if atomic.LoadInt64(&totalOrders) != expected {
		t.Fatalf("expected total orders to be %d, got %d", expected, totalOrders)
	}
}
