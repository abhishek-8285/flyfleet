package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TelemetryData struct {
	GPSLat         float64   `json:"gps_lat"`
	GPSLon         float64   `json:"gps_lon"`
	Speed          float64   `json:"speed"`
	BatteryVoltage float64   `json:"battery_voltage"`
	Timestamp      time.Time `json:"timestamp"`
}

func main() {
	numDevices := flag.Int("devices", 100, "Number of simulated devices")
	ingestionURL := flag.String("url", "http://localhost:8080/diagnostics/devices/%s/telemetry", "Ingestion API URL pattern")
	ratePerSecond := flag.Int("rate", 1, "Successful reports per second per device")
	duration := flag.Duration("duration", 1*time.Minute, "Simulation duration")
	flag.Parse()

	log.Printf("Starting simulation: %d devices, %v duration at %d reports/sec\n", *numDevices, *duration, *ratePerSecond)

	var wg sync.WaitGroup
	ctx := time.After(*duration)

	for i := 0; i < *numDevices; i++ {
		deviceID := uuid.New().String()
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			ticker := time.NewTicker(time.Second / time.Duration(*ratePerSecond))
			defer ticker.Stop()

			for {
				select {
				case <-ctx:
					return
				case <-ticker.C:
					sendTelemetry(id, *ingestionURL)
				}
			}
		}(deviceID)
	}

	wg.Wait()
	log.Println("Simulation completed.")
}

func sendTelemetry(deviceID, urlPattern string) {
	url := fmt.Sprintf(urlPattern, deviceID)

	data := TelemetryData{
		GPSLat:         12.9716 + (rand.Float64() * 0.1),
		GPSLon:         77.5946 + (rand.Float64() * 0.1),
		Speed:          40 + (rand.Float64() * 40),
		BatteryVoltage: 11.5 + (rand.Float64() * 2.0),
		Timestamp:      time.Now(),
	}

	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeviceToken sim-token-"+deviceID)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Suppress spam in high rate tests, just log errors occasionally
		if rand.Float64() < 0.01 {
			log.Printf("Error sending telemetry for %s: %v\n", deviceID, err)
		}
		return
	}
	defer resp.Body.Close()
}
