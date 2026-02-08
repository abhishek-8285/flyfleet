package processor

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/flyfleet/vehicle-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BulkProcessor struct {
	bulkRepo    *repository.BulkRepository
	vehicleRepo *repository.VehicleRepository
	logger      *zap.Logger
}

func NewBulkProcessor(bulkRepo *repository.BulkRepository, vehicleRepo *repository.VehicleRepository, logger *zap.Logger) *BulkProcessor {
	return &BulkProcessor{
		bulkRepo:    bulkRepo,
		vehicleRepo: vehicleRepo,
		logger:      logger,
	}
}

func (p *BulkProcessor) ProcessCSV(jobID uuid.UUID, reader io.Reader, orgID uuid.UUID) {
	// Update status to processing
	ctx := context.Background()
	p.bulkRepo.UpdateJobStatus(ctx, jobID, "processing", 0, 0)

	csvReader := csv.NewReader(reader)
	
	// Read header
	header, err := csvReader.Read()
	if err != nil {
		p.logger.Error("Failed to read CSV header", zap.Error(err))
		p.bulkRepo.UpdateJobStatus(ctx, jobID, "failed", 0, 0)
		return
	}

	// Map header to index
	headerMap := make(map[string]int)
	for i, col := range header {
		headerMap[col] = i
	}

	successCount := 0
	failedCount := 0
	rowNumber := 1 // Header is 0, but 1-based index for user

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			p.logger.Error("Failed to read CSV record", zap.Error(err))
			continue
		}
		rowNumber++

		// Parse vehicle data
		vehicleReq := &models.CreateVehicleRequest{}
		recordData := make(map[string]interface{})
		
		var rowErr error

		// Helper to get value
		getVal := func(col string) string {
			if idx, ok := headerMap[col]; ok && idx < len(record) {
				val := record[idx]
				recordData[col] = val
				return val
			}
			return ""
		}

		// Simple mapping (Case sensitive for now, ideally strictly defined)
		vehicleReq.VIN = getVal("vin")
		vehicleReq.LicensePlate = getVal("license_plate")
		vehicleReq.Make = getVal("make")
		vehicleReq.Model = getVal("model")
		
		yearStr := getVal("year")
		if yearStr != "" {
			year, err := strconv.Atoi(yearStr)
			if err == nil {
				vehicleReq.Year = year
			} else {
				rowErr = fmt.Errorf("invalid year format")
			}
		}

		vehicleReq.Color = getVal("color")
		vehicleReq.FuelType = getVal("fuel_type")
		vehicleReq.Registration = getVal("registration")
		vehicleReq.Insurance = getVal("insurance")
		
		capacityStr := getVal("capacity")
		if capacityStr != "" {
			cap, err := strconv.Atoi(capacityStr)
			if err == nil {
				vehicleReq.Capacity = cap
			}
		}

		// Validation
		if vehicleReq.VIN == "" || vehicleReq.LicensePlate == "" {
			rowErr = fmt.Errorf("missing mandatory fields (vin, license_plate)")
		}

		// Process Row
		status := "success"
		errMsg := ""
		
		if rowErr == nil {
			// Try to create vehicle
			vehicle := &models.Vehicle{
				ID:           uuid.New(),
				OrgID:        orgID,
				LicensePlate: vehicleReq.LicensePlate,
				Model:        vehicleReq.Model,
				Make:         vehicleReq.Make,
				Year:         vehicleReq.Year,
				Color:        vehicleReq.Color,
				Capacity:     vehicleReq.Capacity,
				FuelType:     vehicleReq.FuelType,
				VIN:          vehicleReq.VIN,
				Registration: vehicleReq.Registration,
				Insurance:    vehicleReq.Insurance,
				Status:       "active",
				Odometer:     0,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			
			if err := p.vehicleRepo.Create(ctx, vehicle); err != nil {
				rowErr = err
			}
		}

		if rowErr != nil {
			status = "failed"
			errMsg = rowErr.Error()
			failedCount++
		} else {
			successCount++
		}

		// Log Record
		// Only log errors or every Nth success to save space? 
		// For now log all as per prototype requirement (needs Detail View)
		
		// Optimization: Batch inserts for records? 
		// For MVP, single insert.
		
		recordDataJSON, _ := json.Marshal(recordData)
		
		jobRecord := &models.BulkJobRecord{
			ID:           uuid.New(),
			JobID:        jobID,
			RowNumber:    rowNumber,
			RecordData:   recordDataJSON,
			Status:       status,
			ErrorMessage: errMsg,
			CreatedAt:    time.Now(),
		}
		
		// We might want to fire and forget or batch this. 
		// Doing it synchronously here for simplicity.
		p.bulkRepo.AddJobRecord(ctx, jobRecord)
	}

	// Final Update
	finalStatus := "completed"
	if failedCount > 0 {
		if successCount > 0 {
			finalStatus = "partial_success"
		} else {
			finalStatus = "failed"
		}
	} else if successCount == 0 && failedCount == 0 {
        // Empty file?
        finalStatus = "completed"
    }

	p.bulkRepo.UpdateJobStatus(ctx, jobID, finalStatus, successCount, failedCount)
}
