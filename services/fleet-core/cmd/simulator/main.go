package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofrs/uuid/v5"
	"google.golang.org/protobuf/proto"

	telemetryv1 "github.com/flyfleet/proto/telemetry/v1"
)

// ANSI Colors for premium dashboard printout
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorPurple = "\033[35m"
	ColorGray   = "\033[90m"
)

type JSONPayload struct {
	EventID    string  `json:"event_id"`
	DeviceID   string  `json:"device_id"`
	VehicleID  string  `json:"vehicle_id"`
	TenantID   string  `json:"tenant_id"`
	Latitude   float64 `json:"lat"`
	Longitude  float64 `json:"lng"`
	Speed      float64 `json:"speed"`
	Heading    float64 `json:"heading"`
	Ignition   bool    `json:"ignition"`
	RecordedAt string  `json:"recorded_at"`
	RawData    string  `json:"raw_data,omitempty"`
}

func main() {
	// 1. Command Flags
	modeFlag := flag.String("mode", "protobuf", "Simulation mode: 'json' or 'protobuf'")
	urlFlag := flag.String("url", "http://localhost:8110/api/v1/telemetry", "Ingress API endpoint URL")
	vehicleIDFlag := flag.String("vehicle-id", "550e8400-e29b-41d4-a716-446655440001", "Vehicle UUID")
	tenantIDFlag := flag.String("tenant-id", "550e8400-e29b-41d4-a716-446655440000", "Tenant UUID")
	deviceIDFlag := flag.String("device-id", "DEV-DEVICE-001", "Hardware GPS Device ID")
	regFlag := flag.String("reg", "DL01AB1234", "Vehicle Registration Plate")
	intervalFlag := flag.Duration("interval", 5*time.Second, "Update frequency interval (e.g. 5s)")
	flag.Parse()

	// Title Banner
	fmt.Printf("%s%s================================================================%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s%s   🚀 FLYFLEET HIGH-FIDELITY VEHICLE TELEMETRY SIMULATOR 🚀%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s%s================================================================%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s• Mode:       %s%s%s\n", ColorGray, ColorBold, *modeFlag, ColorReset)
	fmt.Printf("%s• Endpoint:   %s%s%s\n", ColorGray, ColorBold, *urlFlag, ColorReset)
	fmt.Printf("%s• Vehicle:    %s%s (%s)%s\n", ColorGray, ColorBold, *regFlag, *vehicleIDFlag, ColorReset)
	fmt.Printf("%s• Device ID:  %s%s%s\n", ColorGray, ColorBold, *deviceIDFlag, ColorReset)
	fmt.Printf("%s• Interval:   %s%v%s\n", ColorGray, ColorBold, *intervalFlag, ColorReset)
	fmt.Printf("%s----------------------------------------------------------------%s\n", ColorGray, ColorReset)

	// Simulation Variables
	lat := 28.6139
	lng := 77.2090
	prevLat := lat
	prevLng := lng
	heading := 135.0
	fuelLevel := 85.0
	odometerMeters := 1500000.0

	// Handle Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(*intervalFlag)
	defer ticker.Stop()

	// Seed math/rand
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Press Ctrl+C to terminate the simulator...\n")

	for {
		select {
		case <-sigChan:
			fmt.Printf("\n%sStopping vehicle simulation gracefully. Safe travels! 👋%s\n", ColorYellow, ColorReset)
			return
		case <-ticker.C:
			// Update spatial coordinates
			lat += 0.0003
			lng += 0.0004

			// Calculate distance moved in meters (simplified Euclidean)
			dLat := lat - prevLat
			dLng := lng - prevLng
			// 1 degree lat is ~111,139 meters, 1 degree lng is ~97,500 meters at Delhi latitude
			distM := math.Sqrt(math.Pow(dLat*111139.0, 2) + math.Pow(dLng*97500.0, 2))
			odometerMeters += distM

			// Calculate Speed in MPH (meters / seconds * 2.237)
			speedMph := (distM / intervalFlag.Seconds()) * 2.237
			if speedMph < 0.1 {
				speedMph = 0.0
			}

			// Update Heading (rotate slightly)
			heading = math.Mod(heading+2.0, 360.0)

			// Slowly reduce fuel
			fuelLevel -= 0.01 * intervalFlag.Seconds()
			if fuelLevel < 5.0 {
				fuelLevel = 98.0 // Refuel!
			}

			// Fluctuate J1939 CAN values
			engineRpm := int32(1400 + rand.Intn(600))
			engineLoad := 40.0 + rand.Float64()*30.0
			coolantTemp := int32(80 + rand.Intn(10))
			batteryVolts := 13.5 + rand.Float64()*0.8
			cellSignal := int32(-75 + rand.Intn(15))
			hdop := 0.9 + rand.Float64()*0.4

			// Check for random mock DTC occurrence (10% chance)
			var activeDtcs []*telemetryv1.ActiveDTC
			if rand.Float64() < 0.10 {
				activeDtcs = append(activeDtcs, &telemetryv1.ActiveDTC{
					Spn:   fmt.Sprintf("%d", 91+rand.Intn(5)),
					Fmi:   "19",
					Count: 1,
				})
			}

			eventIDUUID, _ := uuid.NewV4()
			eventID := eventIDUUID.String()
			now := time.Now().UTC()
			timestampStr := now.Format(time.RFC3339)

			// 2. Generate HMAC Signature
			// Signature computes EventID + DeviceID with secret "telemetry-secret"
			mac := hmac.New(sha256.New, []byte("telemetry-secret"))
			mac.Write([]byte(eventID + *deviceIDFlag))
			signature := hex.EncodeToString(mac.Sum(nil))

			var req *http.Request
			var err error
			var payloadBytes []byte
			var contentType string

			if *modeFlag == "protobuf" {
				contentType = "application/x-protobuf"
				// Build protobuf payload
				pbPayload := &telemetryv1.TelemetryPayload{
					DeviceId:            *deviceIDFlag,
					TimestampMs:         now.UnixNano() / int64(time.Millisecond),
					Latitude:            lat,
					Longitude:           lng,
					SpeedMph:            float32(speedMph),
					HeadingDegrees:      int32(heading),
					Hdop:                float32(hdop),
					EngineRpm:           uint32(engineRpm),
					EngineLoadPercent:   float32(engineLoad),
					FuelLevelPercent:    float32(fuelLevel),
					TotalOdometerMeters: odometerMeters,
					CoolantTempCelsius:  coolantTemp,
					IgnitionStatus:      true,
					BatteryVoltage:      float32(batteryVolts),
					CellularSignalDbm:   cellSignal,
					ActiveDtcs:          activeDtcs,
				}

				payloadBytes, err = proto.Marshal(pbPayload)
				if err != nil {
					fmt.Printf("%s[ERROR]%s Failed to marshal protobuf: %v\n", ColorRed, ColorReset, err)
					continue
				}
			} else {
				contentType = "application/json"
				// Build J1939 CAN custom telemetry structure inside raw_data JSON
				var rawDataMap = map[string]any{
					"engine_rpm":           engineRpm,
					"engine_load_percent":  engineLoad,
					"fuel_level_percent":   fuelLevel,
					"coolant_temp_celsius": coolantTemp,
					"battery_voltage":      batteryVolts,
					"cellular_signal_dbm":  cellSignal,
					"hdop":                 hdop,
				}
				if len(activeDtcs) > 0 {
					var list []map[string]any
					for _, d := range activeDtcs {
						list = append(list, map[string]any{
							"spn":   d.Spn,
							"fmi":   d.Fmi,
							"count": d.Count,
						})
					}
					rawDataMap["active_dtcs"] = list
				}
				rawDataBytes, _ := json.Marshal(rawDataMap)

				jsonPayload := JSONPayload{
					EventID:    eventID,
					DeviceID:   *deviceIDFlag,
					VehicleID:  *vehicleIDFlag,
					TenantID:   *tenantIDFlag,
					Latitude:   lat,
					Longitude:  lng,
					Speed:      speedMph * 1.60934, // JSON expects standard speed metric km/h
					Heading:    heading,
					Ignition:   true,
					RecordedAt: timestampStr,
					RawData:    string(rawDataBytes),
				}

				payloadBytes, err = json.Marshal(jsonPayload)
				if err != nil {
					fmt.Printf("%s[ERROR]%s Failed to marshal JSON: %v\n", ColorRed, ColorReset, err)
					continue
				}
			}

			// 3. Post HTTP request
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			req, err = http.NewRequestWithContext(ctx, http.MethodPost, *urlFlag, bytes.NewReader(payloadBytes))
			if err != nil {
				cancel()
				fmt.Printf("%s[ERROR]%s Failed to create HTTP request: %v\n", ColorRed, ColorReset, err)
				continue
			}

			req.Header.Set("Content-Type", contentType)
			req.Header.Set("X-HMAC-Signature", signature)
			req.Header.Set("X-Event-ID", eventID)

			resp, err := http.DefaultClient.Do(req)
			cancel()
			if err != nil {
				fmt.Printf("%s[FAIL]%s Network/Ingress unreachable: %v\n", ColorRed, ColorReset, err)
				continue
			}

			// Clean output logging format
			statusStr := fmt.Sprintf("%s202 Accepted%s", ColorGreen, ColorReset)
			if resp.StatusCode != http.StatusAccepted {
				bodyBytes, _ := io.ReadAll(resp.Body)
				statusStr = fmt.Sprintf("%s%d Mismatch (%s)%s", ColorRed, resp.StatusCode, string(bodyBytes), ColorReset)
			}
			resp.Body.Close()

			// Print telemetry metrics beautifully
			fmt.Printf("%s[%s]%s Pos: %s%.6f, %.6f%s | Odo: %s%.2f km%s | Speed: %s%.1f MPH%s | RPM: %s%d%s | Status: %s\n",
				ColorGray, timestampStr, ColorReset,
				ColorBold, lat, lng, ColorReset,
				ColorCyan, odometerMeters/1000.0, ColorReset,
				ColorYellow, speedMph, ColorReset,
				ColorPurple, engineRpm, ColorReset,
				statusStr,
			)

			if len(activeDtcs) > 0 {
				fmt.Printf("  ⚠️  %sSPN %s Fault Code Active on Device CAN Interface%s\n", ColorRed, activeDtcs[0].Spn, ColorReset)
			}

			prevLat = lat
			prevLng = lng
		}
	}
}
