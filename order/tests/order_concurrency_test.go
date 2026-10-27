package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"order/config"
	"order/database"
	"order/model"
	"order/repository"
	"order/service"
)

type mockClient struct{}

func (m *mockClient) GetProductWithResilience(ctx context.Context, id uint) (*service.ProductDetails, error) {
	return &service.ProductDetails{ProductID: id, Name: "Item", Price: 50.0, InStock: true}, nil
}

func TestConcurrentOrderCreation(t *testing.T) {
	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		t.Skip("Postgres not available, skipping test")
	}

	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo, &mockClient{})

	totalRequests := 100
	var wg sync.WaitGroup
	var successCount int64

	for i := 1; i <= totalRequests; i++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()
			_, err := svc.CreateOrder(context.Background(), model.CreateOrderInput{
				CustomerID: uint(cid),
				ProductID:  1,
				Quantity:   1,
			})
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != int64(totalRequests) {
		t.Errorf("expected %d orders, successfully placed %d", totalRequests, successCount)
	}
}
