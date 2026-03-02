package processor

import (
	"encoding/binary"
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
