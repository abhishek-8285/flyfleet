package com.flyfleet.vehicledevice.events

import kotlin.test.Test
import kotlin.test.assertEquals

class EventPriorityResolverTest {
    private val resolver = EventPriorityResolver()

    @Test
    fun `critical driving and panic events map to tier1`() {
        assertEquals(UploadPriorityTier.TIER1, resolver.resolveTier(DeviceEventType.PANIC_TRIGGERED))
        assertEquals(UploadPriorityTier.TIER1, resolver.resolveTier(DeviceEventType.HARSH_BRAKE_DETECTED))
        assertEquals(UploadPriorityTier.TIER1, resolver.resolveTier(DeviceEventType.SPEED_SPIKE_DETECTED))
    }

    @Test
    fun `camera stall and thermal events map to tier2`() {
        assertEquals(UploadPriorityTier.TIER2, resolver.resolveTier(DeviceEventType.CAMERA_FRAME_STALL))
        assertEquals(UploadPriorityTier.TIER2, resolver.resolveTier(DeviceEventType.THERMAL_THROTTLE))
    }

    @Test
    fun `operational telemetry events map to tier3`() {
        assertEquals(UploadPriorityTier.TIER3, resolver.resolveTier(DeviceEventType.GPS_SIGNAL_LOSS))
        assertEquals(UploadPriorityTier.TIER3, resolver.resolveTier(DeviceEventType.BLE_DISCONNECT))
        assertEquals(UploadPriorityTier.TIER3, resolver.resolveTier(DeviceEventType.NETWORK_QUALITY_DROP))
        assertEquals(UploadPriorityTier.TIER3, resolver.resolveTier(DeviceEventType.AI_MODEL_UNAVAILABLE))
        assertEquals(UploadPriorityTier.TIER3, resolver.resolveTier(DeviceEventType.STORAGE_LOW))
        assertEquals(UploadPriorityTier.TIER3, resolver.resolveTier(DeviceEventType.CAN_BUS_DEGRADED))
        assertEquals(UploadPriorityTier.TIER3, resolver.resolveTier(DeviceEventType.TWIN_DRIFT_DETECTED))
    }
}
