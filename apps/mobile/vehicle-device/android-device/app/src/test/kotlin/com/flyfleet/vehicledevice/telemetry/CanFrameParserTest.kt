package com.flyfleet.vehicledevice.telemetry

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import kotlin.experimental.and

class CanFrameParserTest {

    @Test
    fun testParseEngineSpeed() {
        // PGN 61444 (0xF004) EEC1
        // ID: 0CF00400 (priority 3, PGN F004, source 00)
        // Data: FF FF FF 10 20 FF FF FF 
        // RPM bytes 3,4 (0-indexed) -> index 4 is high, 3 is low (Little Endian)
        // 0x20 << 8 | 0x10 = 0x2010 = 8208 decimal
        // 8208 * 0.125 = 1026.0 RPM
        
        val canId = 0x0CF00400
        val data = byteArrayOf(-1, -1, -1, 0x10, 0x20, -1, -1, -1)
        
        val result = CanFrameParser.parse(canId, data)
        
        assertEquals(1026.0, result?.engineRpm!!, 0.1)
    }

    @Test
    fun testParseVehicleSpeed() {
        // PGN 65265 (0xFEF1) CCVS
        // ID: 18FEF100
        // Data: FF 50 19 10 FF FF FF FF
        // Speed bytes 1,2 (0-indexed) -> index 2 is high, 1 is low
        // 0x19 << 8 | 0x50 = 0x1950 = 6480 decimal
        // 6480 / 256.0 = 25.3125 km/h
        
        // Brake byte 3 (0-indexed) -> 0x10 (decimal 16)
        // 0x10 & 0x10 = 0x10 (true)
        
        val canId = 0x18FEF100
        val data = byteArrayOf(-1, 0x50, 0x19, 0x10, -1, -1, -1, -1)
        
        val result = CanFrameParser.parse(canId, data)
        
        assertEquals(25.3125, result?.speedKmh!!, 0.001)
        assertTrue(result.brakePedalPressed!!)
    }
}
