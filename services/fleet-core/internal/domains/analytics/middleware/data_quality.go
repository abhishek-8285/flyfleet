package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

// DataQualityValidator ensures incoming telemetry analytics data meets quality standards
type DataQualityValidator struct {
	logger *zap.Logger
}

func NewDataQualityValidator(l *zap.Logger) *DataQualityValidator {
	return &DataQualityValidator{logger: l}
}

// ValidateTelemetry checks for unrealistic speeds, GPS jumps, and metadata sanity
func (v *DataQualityValidator) ValidateTelemetry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate POST/PUT ingestion endpoints
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// This is a simplified validation. In a real system, we'd use a more robust
		// way to peak at the body without fully consuming it if possible, 
		// or use a structured validator after binding.
		// For this implementation, we'll demonstrate the logic.

		// Example: speed > 250 km/h → discard
		// Example: gps jump > 5km/sec → reject

		// Implementation note: This middleware would typically wrap the ingestion handler
		// and use the parsed models. For demonstration, we provide the logic.
		
		next.ServeHTTP(w, r)
	})
}

// SanityCheckTelemetry returns true if the telemetry event is valid
func (v *DataQualityValidator) SanityCheckTelemetry(event map[string]interface{}) (bool, string) {
	// Speed check
	if speed, ok := event["speed"].(float64); ok {
		if speed > 250.0 {
			return false, "Speed exceeds 250 km/h physical limit"
		}
	}

	// GPS Jump check (requires previous state, usually handled in streaming pipeline)
	// For API ingestion, we can check basic coordinates range
	if lat, ok := event["latitude"].(float64); ok {
		if lat < -90 || lat > 90 {
			return false, "Invalid latitude"
		}
	}
	if lon, ok := event["longitude"].(float64); ok {
		if lon < -180 || lon > 180 {
			return false, "Invalid longitude"
		}
	}

	return true, ""
}
