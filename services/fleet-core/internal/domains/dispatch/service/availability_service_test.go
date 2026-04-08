package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock driver repository for availability tests
type mockDriverRepo struct {
	drivers map[string]*models.Driver
}

func (m *mockDriverRepo) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	if d, ok := m.drivers[id]; ok {
		return d, nil
	}
	return nil, assert.AnError
}

func TestAvailabilityService_IsDriverAvailable(t *testing.T) {
	tests := []struct {
		name     string
		driver   *models.Driver
		expected bool
	}{
		{
			name: "active driver no trip",
			driver: &models.Driver{
				Status:        "active",
				CurrentTripID: nil,
			},
			expected: true,
		},
		{
			name: "active driver with empty string trip ID",
			driver: &models.Driver{
				Status:        "active",
				CurrentTripID: strPtr(""),
			},
			expected: true,
		},
		{
			name: "active driver on active trip",
			driver: &models.Driver{
				Status:        "active",
				CurrentTripID: strPtr("trip-123"),
			},
			expected: false,
		},
		{
			name: "inactive driver",
			driver: &models.Driver{
				Status:        "suspended",
				CurrentTripID: nil,
			},
			expected: false,
		},
		{
			name: "resigned driver",
			driver: &models.Driver{
				Status:        "resigned",
				CurrentTripID: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDriverRepo{drivers: map[string]*models.Driver{"d1": tt.driver}}
			svc := &AvailabilityService{driverRepo: repo}

			available, err := svc.IsDriverAvailable(context.Background(), "d1")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, available)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
