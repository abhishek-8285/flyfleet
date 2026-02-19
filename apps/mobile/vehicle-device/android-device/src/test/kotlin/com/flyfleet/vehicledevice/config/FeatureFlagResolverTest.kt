package com.flyfleet.vehicledevice.config

import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class FeatureFlagResolverTest {
    private val alwaysValidVerifier = object : FeatureFlagSignatureVerifier {
        override fun isValid(payload: String, signature: String): Boolean = true
    }

    private val alwaysInvalidVerifier = object : FeatureFlagSignatureVerifier {
        override fun isValid(payload: String, signature: String): Boolean = false
    }

    private fun baselineConfig(): FeatureFlagConfig {
        return FeatureFlagConfig(
            version = "v1",
            global = FeatureFlags(
                cameraEnabled = true,
                audioEventChecksEnabled = false,
                aiEnabled = true,
                canBusEnabled = true,
                driverLockoutEnabled = true,
            ),
            fleetOverrides = mapOf(
                "fleet-a" to FeatureFlagsPatch(aiEnabled = false),
            ),
            vehicleOverrides = mapOf(
                "veh-1" to FeatureFlagsPatch(cameraEnabled = false),
            ),
            driverOverrides = mapOf(
                "drv-9" to FeatureFlagsPatch(audioEventChecksEnabled = true),
            ),
            deviceOverrides = mapOf(
                "dev-7" to FeatureFlagsPatch(driverLockoutEnabled = false),
            ),
        )
    }

    @Test
    fun `applies precedence global to fleet to vehicle to driver to device`() {
        val resolver = FeatureFlagResolver(alwaysValidVerifier)
        val cfg = baselineConfig()

        val effective = resolver.resolveEffectiveFlags(
            signedRemoteConfig = SignedFeatureFlagConfig(cfg, payload = "signed", signature = "ok"),
            lastKnownGoodConfig = cfg,
            context = FeatureEvaluationContext(
                fleetId = "fleet-a",
                vehicleId = "veh-1",
                driverId = "drv-9",
                deviceId = "dev-7",
            ),
        )

        assertFalse(effective.cameraEnabled)
        assertTrue(effective.audioEventChecksEnabled)
        assertFalse(effective.aiEnabled)
        assertTrue(effective.canBusEnabled)
        assertFalse(effective.driverLockoutEnabled)
    }

    @Test
    fun `local emergency override has highest precedence`() {
        val resolver = FeatureFlagResolver(alwaysValidVerifier)
        val cfg = baselineConfig()

        val effective = resolver.resolveEffectiveFlags(
            signedRemoteConfig = SignedFeatureFlagConfig(cfg, payload = "signed", signature = "ok"),
            lastKnownGoodConfig = cfg,
            context = FeatureEvaluationContext(deviceId = "dev-7"),
            localEmergencyOverride = FeatureFlagsPatch(
                cameraEnabled = false,
                aiEnabled = false,
                canBusEnabled = false,
                driverLockoutEnabled = true,
            ),
        )

        assertFalse(effective.cameraEnabled)
        assertFalse(effective.aiEnabled)
        assertFalse(effective.canBusEnabled)
        assertTrue(effective.driverLockoutEnabled)
    }

    @Test
    fun `falls back to last known good when signature verification fails`() {
        val fallback = baselineConfig()
        val remote = fallback.copy(
            version = "v2",
            global = fallback.global.copy(
                cameraEnabled = false,
                aiEnabled = false,
                canBusEnabled = false,
            ),
        )

        val resolver = FeatureFlagResolver(alwaysInvalidVerifier)
        val effective = resolver.resolveEffectiveFlags(
            signedRemoteConfig = SignedFeatureFlagConfig(remote, payload = "tampered", signature = "bad"),
            lastKnownGoodConfig = fallback,
            context = FeatureEvaluationContext(),
        )

        assertTrue(effective.cameraEnabled)
        assertTrue(effective.aiEnabled)
        assertTrue(effective.canBusEnabled)
    }

    @Test
    fun `rejects signed config rollback when remote version is older`() {
        val lastKnownGood = baselineConfig().copy(
            version = "v3",
            global = baselineConfig().global.copy(aiEnabled = true),
        )
        val signedButOlder = baselineConfig().copy(
            version = "v2",
            global = baselineConfig().global.copy(aiEnabled = false),
        )

        val resolver = FeatureFlagResolver(alwaysValidVerifier)
        val effective = resolver.resolveEffectiveFlags(
            signedRemoteConfig = SignedFeatureFlagConfig(signedButOlder, payload = "signed", signature = "ok"),
            lastKnownGoodConfig = lastKnownGood,
            context = FeatureEvaluationContext(),
        )

        assertTrue(effective.aiEnabled)
    }

    @Test
    fun `accepts signed config when remote version is newer`() {
        val lastKnownGood = baselineConfig().copy(version = "v2")
        val signedNewer = baselineConfig().copy(
            version = "v3",
            global = baselineConfig().global.copy(canBusEnabled = false),
        )

        val resolver = FeatureFlagResolver(alwaysValidVerifier)
        val effective = resolver.resolveEffectiveFlags(
            signedRemoteConfig = SignedFeatureFlagConfig(signedNewer, payload = "signed", signature = "ok"),
            lastKnownGoodConfig = lastKnownGood,
            context = FeatureEvaluationContext(),
        )

        assertFalse(effective.canBusEnabled)
    }
}
