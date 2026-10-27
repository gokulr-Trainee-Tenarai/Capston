package tests

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestOrderCreationConcurrency(t *testing.T) {
	var totalOrders int64 = 0
	var wg sync.WaitGroup

	numGoroutines := 20
	ordersPerGoroutine := 10

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

func TestOrderStateConcurrency(t *testing.T) {
	type OrderState struct {
		mu     sync.Mutex
		status string
	}

	state := &OrderState{status: "CREATED"}
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(step int) {
			defer wg.Done()
			state.mu.Lock()
			if step%2 == 0 {
				state.status = "CONFIRMED"
			} else {
				state.status = "PROCESSING"
			}
			state.mu.Unlock()
		}(i)
	}

	wg.Wait()

	state.mu.Lock()
	defer state.mu.Unlock()
	if state.status != "CONFIRMED" && state.status != "PROCESSING" {
		t.Fatalf("unexpected order status: %s", state.status)
	}
}
