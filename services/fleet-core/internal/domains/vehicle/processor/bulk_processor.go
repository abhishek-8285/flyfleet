package processor

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/repository"
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

		// Helper to get value with SECURITY FIX: sanitize spreadsheet formula injection
		getVal := func(col string) string {
			if idx, ok := headerMap[col]; ok && idx < len(record) {
				val := record[idx]
				recordData[col] = val
				// SECURITY FIX: Prevent spreadsheet formula injection (CWE-1236)
				// If the value starts with =, +, -, @, or tab, prefix with a single quote
				// to prevent Excel/Sheets from treating it as a formula
				if len(val) > 0 {
					first := strings.ToUpper(val)[0]
					if first == '=' || first == '+' || first == '-' || first == '@' || val[0] == '\t' {
						val = "'" + val
					}
				}
				return val
			}
			return ""
		}

		// Simple mapping (CSV columns to enterprise model fields)
		vinVal := getVal("vin")
		var vinPtr *string
		if vinVal != "" {
			vinPtr = &vinVal
		}
		regNo := getVal("license_plate")
		vehicleReq.RegistrationNumber = regNo
		if vehicleReq.RegistrationNumber == "" {
			vehicleReq.RegistrationNumber = getVal("registration_number")
		}
		makeVal := getVal("make")
		if makeVal != "" {
			vehicleReq.Make = &makeVal
		}
		modelVal := getVal("model")
		if modelVal != "" {
			vehicleReq.Model = &modelVal
		}

		yearStr := getVal("year")
		if yearStr != "" {
			year, err := strconv.Atoi(yearStr)
			if err == nil {
				vehicleReq.Year = &year
			} else {
				rowErr = fmt.Errorf("invalid year format")
			}
		}

		colorVal := getVal("color")
		if colorVal != "" {
			vehicleReq.Color = &colorVal
		}
		vehicleReq.FuelType = getVal("fuel_type")
		if vehicleReq.FuelType == "" {
			vehicleReq.FuelType = "DIESEL"
		}
		vehicleReq.VehicleType = getVal("vehicle_type")
		if vehicleReq.VehicleType == "" {
			vehicleReq.VehicleType = "TRUCK"
		}
		vehicleReq.OwnershipType = getVal("ownership_type")
		if vehicleReq.OwnershipType == "" {
			vehicleReq.OwnershipType = "OWNED"
		}

		// Validation
		if vehicleReq.RegistrationNumber == "" {
			rowErr = fmt.Errorf("missing mandatory field: registration_number / license_plate")
		}

		// Process Row
		status := "success"
		errMsg := ""

		if rowErr == nil {
			// Build enterprise vehicle aggregate
			vehicle := &models.Vehicle{
				ID:                 uuid.New(),
				TenantID:           orgID,
				RegistrationNumber: vehicleReq.RegistrationNumber,
				VINNumber:          vinPtr,
				VehicleType:        vehicleReq.VehicleType,
				FuelType:           vehicleReq.FuelType,
				OwnershipType:      vehicleReq.OwnershipType,
				Make:               vehicleReq.Make,
				Model:              vehicleReq.Model,
				Year:               vehicleReq.Year,
				Color:              vehicleReq.Color,
				Status:             models.StatusActive,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
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
