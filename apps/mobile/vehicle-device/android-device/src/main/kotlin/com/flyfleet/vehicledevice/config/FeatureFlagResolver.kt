package com.flyfleet.vehicledevice.config

import java.nio.charset.StandardCharsets
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

/**
 * Feature flag resolution precedence (lowest -> highest):
 * global -> fleet -> vehicle -> driver -> device -> localEmergency
 */
data class FeatureFlags(
    val cameraEnabled: Boolean = true,
    val audioEventChecksEnabled: Boolean = false,
    val aiEnabled: Boolean = true,
    val canBusEnabled: Boolean = true,
    val driverLockoutEnabled: Boolean = true,
)

data class FeatureFlagsPatch(
    val cameraEnabled: Boolean? = null,
    val audioEventChecksEnabled: Boolean? = null,
    val aiEnabled: Boolean? = null,
    val canBusEnabled: Boolean? = null,
    val driverLockoutEnabled: Boolean? = null,
) {
    companion object {
        fun none(): FeatureFlagsPatch = FeatureFlagsPatch()
    }

    fun isEmpty(): Boolean {
        return cameraEnabled == null &&
            audioEventChecksEnabled == null &&
            aiEnabled == null &&
            canBusEnabled == null &&
            driverLockoutEnabled == null
    }
}

data class FeatureFlagConfig(
    val version: String,
    val global: FeatureFlags,
    val fleetOverrides: Map<String, FeatureFlagsPatch> = emptyMap(),
    val vehicleOverrides: Map<String, FeatureFlagsPatch> = emptyMap(),
    val driverOverrides: Map<String, FeatureFlagsPatch> = emptyMap(),
    val deviceOverrides: Map<String, FeatureFlagsPatch> = emptyMap(),
)

data class FeatureEvaluationContext(
    val fleetId: String? = null,
    val vehicleId: String? = null,
    val driverId: String? = null,
    val deviceId: String? = null,
)

data class SignedFeatureFlagConfig(
    val config: FeatureFlagConfig,
    val payload: String,
    val signature: String,
)

interface FeatureFlagSignatureVerifier {
    fun isValid(payload: String, signature: String): Boolean
}

class HmacSha256SignatureVerifier(secret: ByteArray) : FeatureFlagSignatureVerifier {
    private val key = SecretKeySpec(secret, "HmacSHA256")

    override fun isValid(payload: String, signature: String): Boolean {
        val mac = Mac.getInstance("HmacSHA256")
        mac.init(key)
        val digest = mac.doFinal(payload.toByteArray(StandardCharsets.UTF_8))
        val expected = digest.joinToString(separator = "") { b -> "%02x".format(b) }
        return expected.equals(signature, ignoreCase = true)
    }
}

class FeatureFlagResolver(
    private val signatureVerifier: FeatureFlagSignatureVerifier,
) {
    fun resolveEffectiveFlags(
        signedRemoteConfig: SignedFeatureFlagConfig?,
        lastKnownGoodConfig: FeatureFlagConfig,
        context: FeatureEvaluationContext,
        localEmergencyOverride: FeatureFlagsPatch = FeatureFlagsPatch.none(),
    ): FeatureFlags {
        val activeConfig = chooseConfig(signedRemoteConfig, lastKnownGoodConfig)

        var flags = activeConfig.global
        context.fleetId?.let { flags = apply(flags, activeConfig.fleetOverrides[it]) }
        context.vehicleId?.let { flags = apply(flags, activeConfig.vehicleOverrides[it]) }
        context.driverId?.let { flags = apply(flags, activeConfig.driverOverrides[it]) }
        context.deviceId?.let { flags = apply(flags, activeConfig.deviceOverrides[it]) }

        if (!localEmergencyOverride.isEmpty()) {
            flags = apply(flags, localEmergencyOverride)
        }
        return flags
    }

    private fun chooseConfig(
        signedRemoteConfig: SignedFeatureFlagConfig?,
        lastKnownGoodConfig: FeatureFlagConfig,
    ): FeatureFlagConfig {
        if (signedRemoteConfig == null) {
            return lastKnownGoodConfig
        }
        val verified = signatureVerifier.isValid(signedRemoteConfig.payload, signedRemoteConfig.signature)
        if (!verified) {
            return lastKnownGoodConfig
        }
        return if (isVersionNewerOrEqual(signedRemoteConfig.config.version, lastKnownGoodConfig.version)) {
            signedRemoteConfig.config
        } else {
            lastKnownGoodConfig
        }
    }

    private fun isVersionNewerOrEqual(candidate: String, baseline: String): Boolean {
        val candidateNum = parseVersion(candidate)
        val baselineNum = parseVersion(baseline)
        if (candidateNum == null || baselineNum == null) {
            return false
        }
        return candidateNum >= baselineNum
    }

    private fun parseVersion(version: String): Int? {
        if (!version.startsWith("v")) {
            return null
        }
        return version.removePrefix("v").toIntOrNull()
    }

    private fun apply(base: FeatureFlags, patch: FeatureFlagsPatch?): FeatureFlags {
        if (patch == null) {
            return base
        }
        return FeatureFlags(
            cameraEnabled = patch.cameraEnabled ?: base.cameraEnabled,
            audioEventChecksEnabled = patch.audioEventChecksEnabled ?: base.audioEventChecksEnabled,
            aiEnabled = patch.aiEnabled ?: base.aiEnabled,
            canBusEnabled = patch.canBusEnabled ?: base.canBusEnabled,
            driverLockoutEnabled = patch.driverLockoutEnabled ?: base.driverLockoutEnabled,
        )
    }
}
