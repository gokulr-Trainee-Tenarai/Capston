package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

type ProductDetails struct {
	ProductID uint    `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	InStock   bool    `json:"inStock"`
}

type ProductClient interface {
	GetProductWithResilience(ctx context.Context, productID uint) (*ProductDetails, error)
}

type productClient struct {
	baseURL string
	cb      *gobreaker.CircuitBreaker
	client  *http.Client
}

func NewProductClient(baseURL string) ProductClient {
	st := gobreaker.Settings{
		Name:        "ProductClientCircuitBreaker",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &productClient{
		baseURL: baseURL,
		cb:      gobreaker.NewCircuitBreaker(st),
		client:  &http.Client{Timeout: 2 * time.Second},
	}
}

func (pc *productClient) GetProductWithResilience(ctx context.Context, productID uint) (*ProductDetails, error) {
	res, err := pc.cb.Execute(func() (interface{}, error) {
		var lastErr error
		maxRetries := 3
		backoff := 50 * time.Millisecond

		for i := 0; i < maxRetries; i++ {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			reqCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
			reqURL := fmt.Sprintf("%s/products/%d", pc.baseURL, productID)
			req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, reqURL, nil)
			if err != nil {
				cancel()
				return nil, err
			}

			resp, err := pc.client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				var prod ProductDetails
				if decodeErr := json.NewDecoder(resp.Body).Decode(&prod); decodeErr == nil {
					cancel()
					return &prod, nil
				}
				lastErr = errors.New("failed to decode response")
			} else {
				if err != nil {
					lastErr = err
				} else {
					lastErr = fmt.Errorf("status code: %d", resp.StatusCode)
					resp.Body.Close()
				}
			}
			cancel()

			time.Sleep(backoff)
			backoff *= 2
		}
		return nil, fmt.Errorf("retries exhausted: %w", lastErr)
	})

	if err != nil {
		return nil, err
	}
	return res.(*ProductDetails), nil
}