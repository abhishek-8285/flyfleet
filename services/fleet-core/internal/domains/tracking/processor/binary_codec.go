package processor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/models"
)

// BinaryTelemetryCodec handles bit-packed edge payloads
type BinaryTelemetryCodec struct{}

// NewBinaryTelemetryCodec creates a new codec
func NewBinaryTelemetryCodec() *BinaryTelemetryCodec {
	return &BinaryTelemetryCodec{}
}

// GPSPacket represents the decoded bit-packed data
type GPSPacket struct {
	PacketType     uint8  // 1 byte
	DeviceID       uint64 // 8 bytes
	RelativeTime   uint32 // 4 bytes (seconds since epoch)
	LatitudeFixed  int32  // 4 bytes (decimal * 1e6)
	LongitudeFixed int32  // 4 bytes (decimal * 1e6)
	Battery        uint8  // 1 byte
	Signal         uint8  // 1 byte
	Checksum       uint16 // 2 bytes
}

// DecodeGPS decodes a 25-byte bit-packed GPS packet
func (c *BinaryTelemetryCodec) DecodeGPS(data []byte) (*models.GPSTelemetry, string, error) {
	if len(data) < 25 {
		return nil, "", fmt.Errorf("invalid packet length: expected 25, got %d", len(data))
	}

	// Verify Checksum (Simple CRC-16 simulation for MVP)
	// In production, we would use a real CRC algorithm
	expectedChecksum := binary.BigEndian.Uint16(data[23:25])
	actualChecksum := uint16(0)
	for i := 0; i < 23; i++ {
		actualChecksum += uint16(data[i])
	}
	if actualChecksum != expectedChecksum {
		return nil, "", errors.New("checksum mismatch")
	}

	packet := GPSPacket{
		PacketType:     data[0],
		DeviceID:       binary.BigEndian.Uint64(data[1:9]),
		RelativeTime:   binary.BigEndian.Uint32(data[9:13]),
		LatitudeFixed:  int32(binary.BigEndian.Uint32(data[13:17])),
		LongitudeFixed: int32(binary.BigEndian.Uint32(data[17:21])),
		Battery:        data[21],
		Signal:         data[22],
	}

	if packet.PacketType != 0x01 {
		return nil, "", fmt.Errorf("unexpected packet type: 0x%02x", packet.PacketType)
	}

	timestamp := time.Unix(int64(packet.RelativeTime), 0)
	lat := float64(packet.LatitudeFixed) / 1e6
	lng := float64(packet.LongitudeFixed) / 1e6
	battery := int(packet.Battery)

	telemetry := &models.GPSTelemetry{
		Time:      timestamp,
		Latitude:  lat,
		Longitude: lng,
		Battery:   &battery,
		Ignition:  boolPtr(true), // Fixed for heartbeat
	}

	deviceIDStr := fmt.Sprintf("DEV-%d", packet.DeviceID)
	return telemetry, deviceIDStr, nil
}

func boolPtr(b bool) *bool {
	return &b
}
