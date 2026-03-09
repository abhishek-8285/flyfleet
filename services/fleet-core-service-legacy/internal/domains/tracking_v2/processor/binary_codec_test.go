package processor

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBinaryTelemetryCodec_DecodeGPS(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	// Prepare a mocked 25-byte packet
	// Byte 0: Packet Type (0x01)
	// Bytes 1-8: Device ID (12345)
	// Bytes 9-12: Relative Timestamp (Current Time)
	// Bytes 13-16: Latitude (12.9716 * 1e6 = 12971600)
	// Bytes 17-20: Longitude (77.5946 * 1e6 = 77594600)
	// Byte 21: Battery (85)
	// Byte 22: Signal (90)
	// Bytes 23-24: Checksum

	data := make([]byte, 25)
	data[0] = 0x01
	binary.BigEndian.PutUint64(data[1:9], 12345)
	now := uint32(time.Now().Unix())
	binary.BigEndian.PutUint32(data[9:13], now)
	binary.BigEndian.PutUint32(data[13:17], uint32(12971600))
	binary.BigEndian.PutUint32(data[17:21], uint32(77594600))
	data[21] = 85
	data[22] = 90

	// Calculate Checksum
	checksum := uint16(0)
	for i := 0; i < 23; i++ {
		checksum += uint16(data[i])
	}
	binary.BigEndian.PutUint16(data[23:25], checksum)

	telemetry, deviceID, err := codec.DecodeGPS(data)

	assert.NoError(t, err)
	assert.Equal(t, "DEV-12345", deviceID)
	assert.InDelta(t, 12.9716, telemetry.Latitude, 0.0001)
	assert.InDelta(t, 77.5946, telemetry.Longitude, 0.0001)
	assert.Equal(t, 85, *telemetry.Battery)
}

func TestBinaryTelemetryCodec_DecodeGPS_InvalidChecksum(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	data := make([]byte, 25)
	data[0] = 0x01
	binary.BigEndian.PutUint64(data[1:9], 12345)
	binary.BigEndian.PutUint32(data[9:13], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(data[13:17], uint32(12971600))
	binary.BigEndian.PutUint32(data[17:21], uint32(77594600))
	data[21] = 85
	data[22] = 90
	// Wrong checksum intentionally
	binary.BigEndian.PutUint16(data[23:25], 0xFFFF)

	telemetry, deviceID, err := codec.DecodeGPS(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum")
	assert.Nil(t, telemetry)
	assert.Empty(t, deviceID)
}

func TestBinaryTelemetryCodec_DecodeGPS_InvalidPacketType(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	data := make([]byte, 25)
	data[0] = 0x99 // Invalid packet type
	// Set valid checksum for first 23 bytes
	binary.BigEndian.PutUint64(data[1:9], 12345)
	for i := 1; i < 9; i++ {
		data[i] = 0
	}
	checksum := uint16(0)
	for i := 0; i < 23; i++ {
		checksum += uint16(data[i])
	}
	binary.BigEndian.PutUint16(data[23:25], checksum)

	telemetry, deviceID, err := codec.DecodeGPS(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "packet type")
	assert.Nil(t, telemetry)
	assert.Empty(t, deviceID)
}

func TestBinaryTelemetryCodec_DecodeGPS_ShortPacket(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	data := make([]byte, 10) // Too short

	telemetry, deviceID, err := codec.DecodeGPS(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "packet length")
	assert.Nil(t, telemetry)
	assert.Empty(t, deviceID)
}

func TestBinaryTelemetryCodec_DecodeGPS_EdgeCoordinates(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	tests := []struct {
		name      string
		lat       float64
		lng       float64
		wantLat   float64
		wantLng   float64
		wantError bool
	}{
		{"Equator", 0.0, 0.0, 0.0, 0.0, false},
		{"North Pole", 90.0, 0.0, 90.0, 0.0, false},
		{"South Pole", -90.0, 0.0, -90.0, 0.0, false},
		{"Prime Meridian", 0.0, 180.0, 0.0, 180.0, false},
		{"International Date Line", 0.0, -180.0, 0.0, -180.0, false},
		{"San Francisco", 37.7749, -122.4194, 37.7749, -122.4194, false},
		{"Mumbai", 19.0760, 72.8777, 19.0760, 72.8777, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 25)
			data[0] = 0x01
			binary.BigEndian.PutUint64(data[1:9], 99999)
			binary.BigEndian.PutUint32(data[9:13], uint32(time.Now().Unix()))

			// Convert coordinates to fixed-point
			latFixed := int32(tt.lat * 1e6)
			lngFixed := int32(tt.lng * 1e6)

			binary.BigEndian.PutUint32(data[13:17], uint32(latFixed))
			binary.BigEndian.PutUint32(data[17:21], uint32(lngFixed))
			data[21] = 100
			data[22] = 100

			// Calculate checksum
			checksum := uint16(0)
			for i := 0; i < 23; i++ {
				checksum += uint16(data[i])
			}
			binary.BigEndian.PutUint16(data[23:25], checksum)

			telemetry, deviceID, err := codec.DecodeGPS(data)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "DEV-99999", deviceID)
				assert.InDelta(t, tt.wantLat, telemetry.Latitude, 0.0001)
				assert.InDelta(t, tt.wantLng, telemetry.Longitude, 0.0001)
			}
		})
	}
}

func TestBinaryTelemetryCodec_DecodeGPS_BatteryAndSignalLevels(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	tests := []struct {
		name        string
		battery     uint8
		signal      uint8
		wantBattery int
	}{
		{"Full Battery", 100, 100, 100},
		{"Half Battery", 50, 50, 50},
		{"Low Battery", 10, 20, 10},
		{"Critical Battery", 0, 0, 0},
		{"Max Values", 255, 255, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 25)
			data[0] = 0x01
			binary.BigEndian.PutUint64(data[1:9], 11111)
			binary.BigEndian.PutUint32(data[9:13], uint32(time.Now().Unix()))
			binary.BigEndian.PutUint32(data[13:17], uint32(0))
			binary.BigEndian.PutUint32(data[17:21], uint32(0))
			data[21] = tt.battery
			data[22] = tt.signal

			// Calculate checksum
			checksum := uint16(0)
			for i := 0; i < 23; i++ {
				checksum += uint16(data[i])
			}
			binary.BigEndian.PutUint16(data[23:25], checksum)

			telemetry, _, err := codec.DecodeGPS(data)

			assert.NoError(t, err)
			assert.NotNil(t, telemetry.Battery)
			assert.Equal(t, tt.wantBattery, *telemetry.Battery)
		})
	}
}

func TestBinaryTelemetryCodec_TimestampHandling(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	// Test with specific timestamp
	expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	timestamp := uint32(expectedTime.Unix())

	data := make([]byte, 25)
	data[0] = 0x01
	binary.BigEndian.PutUint64(data[1:9], 54321)
	binary.BigEndian.PutUint32(data[9:13], timestamp)
	binary.BigEndian.PutUint32(data[13:17], uint32(0))
	binary.BigEndian.PutUint32(data[17:21], uint32(0))
	data[21] = 100
	data[22] = 100

	// Calculate checksum
	checksum := uint16(0)
	for i := 0; i < 23; i++ {
		checksum += uint16(data[i])
	}
	binary.BigEndian.PutUint16(data[23:25], checksum)

	telemetry, _, err := codec.DecodeGPS(data)

	assert.NoError(t, err)
	assert.NotNil(t, telemetry.Time)
	assert.InDelta(t, expectedTime.Unix(), telemetry.Time.Unix(), 1)
}

func TestBinaryTelemetryCodec_DeviceIDRange(t *testing.T) {
	codec := NewBinaryTelemetryCodec()

	tests := []struct {
		name       string
		deviceNum  uint64
		wantDevice string
	}{
		{"Min ID", 0, "DEV-0"},
		{"Small ID", 1, "DEV-1"},
		{"Medium ID", 999999, "DEV-999999"},
		{"Large ID", math.MaxUint64, "DEV-18446744073709551615"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 25)
			data[0] = 0x01
			binary.BigEndian.PutUint64(data[1:9], tt.deviceNum)
			binary.BigEndian.PutUint32(data[9:13], uint32(time.Now().Unix()))
			binary.BigEndian.PutUint32(data[13:17], uint32(0))
			binary.BigEndian.PutUint32(data[17:21], uint32(0))
			data[21] = 100
			data[22] = 100

			// Calculate checksum
			checksum := uint16(0)
			for i := 0; i < 23; i++ {
				checksum += uint16(data[i])
			}
			binary.BigEndian.PutUint16(data[23:25], checksum)

			_, deviceID, err := codec.DecodeGPS(data)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantDevice, deviceID)
		})
	}
}

// Helper functions
func bytesRepeating(length int, b byte) []byte {
	data := make([]byte, length)
	for i := range data {
		data[i] = b
	}
	return data
}

func sequentialBytes(length int) []byte {
	data := make([]byte, length)
	for i := range data {
		data[i] = byte(i)
	}
	return data
}
