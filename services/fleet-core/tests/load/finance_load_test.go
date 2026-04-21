package load

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/service"
)

// RunFinanceLoadTest simulates high volume of fuel transaction ingestions
func RunFinanceLoadTest(svc *service.FinanceService, concurrency int, totalRequests int) {
	fmt.Printf("Starting load test: %d concurrency, %d total requests\n", concurrency, totalRequests)

	start := time.Now()
	var wg sync.WaitGroup
	reqChan := make(chan int, totalRequests)

	for i := 0; i < totalRequests; i++ {
		reqChan <- i
	}
	close(reqChan)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range reqChan {
				tx := &models.FuelTransaction{
					TenantID:           "7ca5b00c-b261-4a57-9d7e-721fb0d3bc64",
					VehicleID:       "5e9f8f4a-9b4e-4b4e-9b4e-9b4e9b4e9b4e",
					Liters:          50.0,
					Amount:          5000.0,
					Currency:        "INR",
					TransactionDate: time.Now(),
				}
				_ = svc.IngestFuelData(context.Background(), tx)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("Load test completed in %v. Throughput: %.2f req/s\n", duration, float64(totalRequests)/duration.Seconds())
}
