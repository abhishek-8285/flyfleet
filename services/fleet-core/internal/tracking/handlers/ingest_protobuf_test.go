package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
	telemetryv1 "github.com/flyfleet/proto/telemetry/v1"
)

type mockIngestService struct {
	ingestFunc func(ctx context.Context, event *models.TelemetryEvent) error
}

func (m *mockIngestService) Ingest(ctx context.Context, event *models.TelemetryEvent) error {
	if m.ingestFunc != nil {
		return m.ingestFunc(ctx, event)
	}
	return nil
}

func TestServeHTTP_JSON_Mode(t *testing.T) {
	logger := zap.NewNop()
	mockSvc := &mockIngestService{}

	var capturedEvent *models.TelemetryEvent
	mockSvc.ingestFunc = func(ctx context.Context, event *models.TelemetryEvent) error {
		capturedEvent = event
		return nil
	}

	handler := NewIngestHandler(mockSvc, logger)

	reqBody := ingestRequest{
		VehicleID:  "4e650119-bc78-43d9-9524-7e82200199e4",
		DeviceID:   "device-1234",
		Latitude:   12.3456,
		Longitude:  78.9012,
		Speed:      65.5,
		Heading:    180.0,
		Ignition:   true,
		FuelLevel:  85.0,
		Odometer:   12345.6,
		SequenceNo: 42,
		RecordedAt: "2026-05-17T10:00:00Z",
		RawData:    "JSON_TEST",
		EventID:    "e6a6f671-bc01-4475-b6d0-42fe1299a9e3",
		TenantID:   "a3a4b5c6-d7e8-4f90-8888-999900001111",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	if capturedEvent == nil {
		t.Fatal("expected event to be captured")
	}

	if capturedEvent.DeviceID != "device-1234" {
		t.Errorf("expected DeviceID device-1234, got %s", capturedEvent.DeviceID)
	}

	if capturedEvent.Lat != 12.3456 || capturedEvent.Lng != 78.9012 {
		t.Errorf("coordinates mismatch")
	}

	if capturedEvent.Speed != 65.5 {
		t.Errorf("speed mismatch")
	}
}

func TestServeHTTP_Protobuf_Mode(t *testing.T) {
	logger := zap.NewNop()
	mockSvc := &mockIngestService{}

	var capturedEvent *models.TelemetryEvent
	mockSvc.ingestFunc = func(ctx context.Context, event *models.TelemetryEvent) error {
		capturedEvent = event
		return nil
	}

	handler := NewIngestHandler(mockSvc, logger)

	pbPayload := &telemetryv1.TelemetryPayload{
		DeviceId:            "pb-device-999",
		TimestampMs:         1779012000000, // May 17, 2026
		Latitude:            28.6139,
		Longitude:           77.2090,
		SpeedMph:            60.0, // Should convert to 60 * 1.60934 = 96.5604 km/h
		HeadingDegrees:      90,
		Hdop:                1.2,
		EngineRpm:           1800,
		EngineLoadPercent:   45.5,
		FuelLevelPercent:    92.0,
		TotalOdometerMeters: 1500000.0, // 1500 km
		CoolantTempCelsius:  85,
		IgnitionStatus:      true,
		BatteryVoltage:      13.8,
		CellularSignalDbm:   -72,
		ActiveDtcs: []*telemetryv1.ActiveDTC{
			{
				Spn:   "111",
				Fmi:   "12",
				Count: 3,
			},
		},
	}

	pbBytes, err := proto.Marshal(pbPayload)
	if err != nil {
		t.Fatalf("failed to marshal protobuf payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry", bytes.NewReader(pbBytes))
	req.Header.Set("Content-Type", "application/x-protobuf")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d. Body: %s", w.Code, w.Body.String())
	}

	if capturedEvent == nil {
		t.Fatal("expected event to be captured")
	}

	if capturedEvent.DeviceID != "pb-device-999" {
		t.Errorf("expected DeviceID pb-device-999, got %s", capturedEvent.DeviceID)
	}

	// Speed converted to km/h
	expectedSpeed := 60.0 * 1.60934
	if capturedEvent.Speed != expectedSpeed {
		t.Errorf("expected speed %f km/h, got %f", expectedSpeed, capturedEvent.Speed)
	}

	// Odometer converted to km
	if capturedEvent.Odometer != 1500.0 {
		t.Errorf("expected odometer 1500 km, got %f", capturedEvent.Odometer)
	}

	// Active DTCs and other J1939 telemetry packed in RawData JSON
	var rawData map[string]any
	if err := json.Unmarshal([]byte(capturedEvent.RawData), &rawData); err != nil {
		t.Fatalf("failed to parse raw_data JSON: %v", err)
	}

	if rpm, ok := rawData["engine_rpm"].(float64); !ok || rpm != 1800 {
		t.Errorf("expected engine_rpm 1800, got %v", rawData["engine_rpm"])
	}

	dtcs, ok := rawData["active_dtcs"].([]any)
	if !ok || len(dtcs) != 1 {
		t.Fatalf("expected 1 active DTC, got %v", rawData["active_dtcs"])
	}

	dtcMap := dtcs[0].(map[string]any)
	if dtcMap["spn"] != "111" || dtcMap["fmi"] != "12" || dtcMap["count"].(float64) != 3 {
		t.Errorf("active DTC details mismatch")
	}
}
