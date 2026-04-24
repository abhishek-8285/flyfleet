package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestVehicleRepository_AssignDriver_Strict(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := &VehicleRepository{db: mock}
	ctx := context.Background()
	vehicleID := uuid.New()
	tenantID := uuid.New()
	driverID := uuid.New()

	t.Run("Successful Assignment", func(t *testing.T) {
		mock.ExpectBegin()
		
		// 1. Check vehicle assignment (returns nil)
		mock.ExpectQuery(`SELECT current_driver_id FROM vehicles`).
			WithArgs(vehicleID, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"current_driver_id"}).AddRow(nil))
		
		// 2. Check driver assignment (returns empty)
		mock.ExpectQuery(`SELECT current_vehicle_id FROM drivers`).
			WithArgs(driverID.String(), tenantID.String()).
			WillReturnRows(pgxmock.NewRows([]string{"current_vehicle_id"}).AddRow(nil))
		
		// 3. Update vehicle
		mock.ExpectExec(`UPDATE vehicles SET current_driver_id`).
			WithArgs(driverID, vehicleID, tenantID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		
		// 4. Update driver
		mock.ExpectExec(`UPDATE drivers SET current_vehicle_id`).
			WithArgs(vehicleID.String(), driverID.String(), tenantID.String()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		
		mock.ExpectCommit()

		err := repo.AssignDriver(ctx, vehicleID, tenantID, driverID)
		assert.NoError(t, err)
	})

	t.Run("Fail: Vehicle Already Assigned", func(t *testing.T) {
		otherDriverID := uuid.New()
		mock.ExpectBegin()
		
		mock.ExpectQuery(`SELECT current_driver_id FROM vehicles`).
			WithArgs(vehicleID, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"current_driver_id"}).AddRow(otherDriverID))
		
		mock.ExpectRollback()

		err := repo.AssignDriver(ctx, vehicleID, tenantID, driverID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vehicle is already assigned")
	})

	t.Run("Fail: Driver Already Assigned", func(t *testing.T) {
		otherVehicleID := uuid.New().String()
		mock.ExpectBegin()
		
		// Vehicle is free
		mock.ExpectQuery(`SELECT current_driver_id FROM vehicles`).
			WithArgs(vehicleID, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"current_driver_id"}).AddRow(nil))
		
		// Driver is busy
		mock.ExpectQuery(`SELECT current_vehicle_id FROM drivers`).
			WithArgs(driverID.String(), tenantID.String()).
			WillReturnRows(pgxmock.NewRows([]string{"current_vehicle_id"}).AddRow(otherVehicleID))
		
		mock.ExpectRollback()

		err := repo.AssignDriver(ctx, vehicleID, tenantID, driverID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "driver is already assigned")
	})
}

func TestVehicleRepository_UnassignDriver_Bidirectional(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := &VehicleRepository{db: mock}
	ctx := context.Background()
	vehicleID := uuid.New()
	tenantID := uuid.New()
	driverID := uuid.New()

	t.Run("Successful Unassignment", func(t *testing.T) {
		mock.ExpectBegin()
		
		// 1. Find current driver
		mock.ExpectQuery(`SELECT current_driver_id FROM vehicles`).
			WithArgs(vehicleID, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"current_driver_id"}).AddRow(driverID))
		
		// 2. Clear vehicle
		mock.ExpectExec(`UPDATE vehicles SET current_driver_id = NULL`).
			WithArgs(vehicleID, tenantID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		
		// 3. Clear driver
		mock.ExpectExec(`UPDATE drivers SET current_vehicle_id = NULL`).
			WithArgs(driverID.String(), tenantID.String()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		
		mock.ExpectCommit()

		err := repo.UnassignDriver(ctx, vehicleID, tenantID)
		assert.NoError(t, err)
	})
}
