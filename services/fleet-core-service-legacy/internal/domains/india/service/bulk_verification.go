package service

import (
	"context"
	"fmt"
	"sync"
)

type BulkVerificationService interface {
	BulkVerifyVahan(ctx context.Context, registrationNumbers []string) (map[string]interface{}, error)
}

type bulkVerificationService struct {
	vahanSvc VahanIntegrationService
}

func NewBulkVerificationService(vahanSvc VahanIntegrationService) BulkVerificationService {
	return &bulkVerificationService{vahanSvc: vahanSvc}
}

func (s *bulkVerificationService) BulkVerifyVahan(ctx context.Context, registrationNumbers []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit batch size to 100 as per requirement
	if len(registrationNumbers) > 100 {
		registrationNumbers = registrationNumbers[:100]
	}

	for _, regNo := range registrationNumbers {
		wg.Add(1)
		go func(rn string) {
			defer wg.Done()
			// In a real scenario, we'd call the VAHAN API here
			// For now, we simulate the bulk processing
			// We use a dummy UUID as we don't have org/vehicle context here in this generic bulk call
			details, err := s.vahanSvc.VerifyVehicle(ctx, [16]byte{}, [16]byte{}, rn)
			
			mu.Lock()
			if err != nil {
				results[rn] = fmt.Sprintf("Error: %v", err)
			} else {
				results[rn] = details
			}
			mu.Unlock()
		}(regNo)
	}

	wg.Wait()
	return results, nil
}
