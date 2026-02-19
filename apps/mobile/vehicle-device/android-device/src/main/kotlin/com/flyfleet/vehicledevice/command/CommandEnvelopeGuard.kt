package com.flyfleet.vehicledevice.command

import com.flyfleet.vehicledevice.sync.Clock
import java.nio.charset.StandardCharsets
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

data class SignedCommandEnvelope(
    val commandId: String,
    val nonce: String,
    val issuedAtMs: Long,
    val commandType: DeviceCommandType,
    val payload: String,
    val signature: String,
)

data class CommandEnvelopeDecision(
    val allowed: Boolean,
    val reason: String,
)

interface CommandEnvelopeSigner {
    fun sign(canonical: String): String
    fun verify(canonical: String, signature: String): Boolean
}

class HmacCommandEnvelopeSigner(secret: ByteArray) : CommandEnvelopeSigner {
    private val key = SecretKeySpec(secret, "HmacSHA256")

    override fun sign(canonical: String): String {
        val mac = Mac.getInstance("HmacSHA256")
        mac.init(key)
        val digest = mac.doFinal(canonical.toByteArray(StandardCharsets.UTF_8))
        return digest.joinToString(separator = "") { "%02x".format(it) }
    }

    override fun verify(canonical: String, signature: String): Boolean {
        return sign(canonical).equals(signature, ignoreCase = true)
    }
}

interface NonceReplayCache {
    /**
     * @return true if nonce was new and is now recorded, false if nonce already exists.
     */
    fun putIfAbsent(nonce: String): Boolean
}

class InMemoryNonceReplayCache : NonceReplayCache {
    private val nonces = mutableSetOf<String>()

    override fun putIfAbsent(nonce: String): Boolean {
        if (nonces.contains(nonce)) {
            return false
        }
        nonces += nonce
        return true
    }
}

interface CommandIdRegistry {
    /**
     * @return true if command id was new and is now recorded, false if already seen.
     */
    fun putIfAbsent(commandId: String): Boolean
}

class InMemoryCommandIdRegistry : CommandIdRegistry {
    private val ids = mutableSetOf<String>()

    override fun putIfAbsent(commandId: String): Boolean {
        if (ids.contains(commandId)) {
            return false
        }
        ids += commandId
        return true
    }
}

class CommandEnvelopeGuard(
    private val clock: Clock,
    private val signer: CommandEnvelopeSigner,
    private val replayCache: NonceReplayCache,
    private val commandIdRegistry: CommandIdRegistry,
    private val acceptedClockSkewMs: Long = 120_000,
) {
    fun verify(envelope: SignedCommandEnvelope): CommandEnvelopeDecision {
        val now = clock.nowMs()
        val drift = kotlin.math.abs(now - envelope.issuedAtMs)
        if (drift > acceptedClockSkewMs) {
            return CommandEnvelopeDecision(false, "stale_or_future_command")
        }

        val canonical = canonical(envelope)
        if (!signer.verify(canonical, envelope.signature)) {
            return CommandEnvelopeDecision(false, "invalid_signature")
        }

        if (!replayCache.putIfAbsent(envelope.nonce)) {
            return CommandEnvelopeDecision(false, "replayed_nonce")
        }

        if (!commandIdRegistry.putIfAbsent(envelope.commandId)) {
            return CommandEnvelopeDecision(false, "duplicate_command_id")
        }

        return CommandEnvelopeDecision(true, "accepted")
    }

    fun canonical(envelope: SignedCommandEnvelope): String {
        return listOf(
            envelope.commandId,
            envelope.nonce,
            envelope.issuedAtMs.toString(),
            envelope.commandType.name,
            envelope.payload,
        ).joinToString("|")
    }
}
