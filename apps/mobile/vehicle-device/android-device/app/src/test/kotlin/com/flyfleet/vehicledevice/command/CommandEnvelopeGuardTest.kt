package com.flyfleet.vehicledevice.command

import com.flyfleet.vehicledevice.sync.Clock
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class CommandEnvelopeGuardTest {
    private class FakeClock(var now: Long) : Clock {
        override fun nowMs(): Long = now
    }

    private fun signedEnvelope(
        signer: CommandEnvelopeSigner,
        commandId: String,
        nonce: String,
        issuedAtMs: Long,
        commandType: DeviceCommandType = DeviceCommandType.FORCE_SYNC,
        payload: String = "{}",
    ): SignedCommandEnvelope {
        val unsigned = SignedCommandEnvelope(
            commandId = commandId,
            nonce = nonce,
            issuedAtMs = issuedAtMs,
            commandType = commandType,
            payload = payload,
            signature = "",
        )
        val canonical = listOf(
            unsigned.commandId,
            unsigned.nonce,
            unsigned.issuedAtMs.toString(),
            unsigned.commandType.name,
            unsigned.payload,
        ).joinToString("|")
        return unsigned.copy(signature = signer.sign(canonical))
    }

    @Test
    fun `accepts valid envelope once`() {
        val clock = FakeClock(now = 1_000)
        val signer = HmacCommandEnvelopeSigner("secret".toByteArray())
        val guard = CommandEnvelopeGuard(
            clock = clock,
            signer = signer,
            replayCache = InMemoryNonceReplayCache(),
            commandIdRegistry = InMemoryCommandIdRegistry(),
        )

        val env = signedEnvelope(signer, commandId = "cmd-1", nonce = "n-1", issuedAtMs = 1_000)
        val result = guard.verify(env)

        assertTrue(result.allowed)
        assertEquals("accepted", result.reason)
    }

    @Test
    fun `rejects stale timestamp outside accepted skew`() {
        val clock = FakeClock(now = 1_000_000)
        val signer = HmacCommandEnvelopeSigner("secret".toByteArray())
        val guard = CommandEnvelopeGuard(
            clock = clock,
            signer = signer,
            replayCache = InMemoryNonceReplayCache(),
            commandIdRegistry = InMemoryCommandIdRegistry(),
            acceptedClockSkewMs = 10_000,
        )

        val env = signedEnvelope(signer, commandId = "cmd-2", nonce = "n-2", issuedAtMs = 900_000)
        val result = guard.verify(env)

        assertFalse(result.allowed)
        assertEquals("stale_or_future_command", result.reason)
    }

    @Test
    fun `rejects invalid signature`() {
        val clock = FakeClock(now = 2_000)
        val signer = HmacCommandEnvelopeSigner("secret".toByteArray())
        val guard = CommandEnvelopeGuard(
            clock = clock,
            signer = signer,
            replayCache = InMemoryNonceReplayCache(),
            commandIdRegistry = InMemoryCommandIdRegistry(),
        )

        val env = signedEnvelope(signer, commandId = "cmd-3", nonce = "n-3", issuedAtMs = 2_000)
            .copy(signature = "bad")
        val result = guard.verify(env)

        assertFalse(result.allowed)
        assertEquals("invalid_signature", result.reason)
    }

    @Test
    fun `rejects replayed nonce`() {
        val clock = FakeClock(now = 3_000)
        val signer = HmacCommandEnvelopeSigner("secret".toByteArray())
        val guard = CommandEnvelopeGuard(
            clock = clock,
            signer = signer,
            replayCache = InMemoryNonceReplayCache(),
            commandIdRegistry = InMemoryCommandIdRegistry(),
        )

        val first = signedEnvelope(signer, commandId = "cmd-4", nonce = "n-replay", issuedAtMs = 3_000)
        val second = signedEnvelope(signer, commandId = "cmd-5", nonce = "n-replay", issuedAtMs = 3_000)

        assertTrue(guard.verify(first).allowed)
        val replay = guard.verify(second)
        assertFalse(replay.allowed)
        assertEquals("replayed_nonce", replay.reason)
    }

    @Test
    fun `rejects duplicate command id even with new nonce`() {
        val clock = FakeClock(now = 4_000)
        val signer = HmacCommandEnvelopeSigner("secret".toByteArray())
        val guard = CommandEnvelopeGuard(
            clock = clock,
            signer = signer,
            replayCache = InMemoryNonceReplayCache(),
            commandIdRegistry = InMemoryCommandIdRegistry(),
        )

        val first = signedEnvelope(signer, commandId = "cmd-dup", nonce = "n-a", issuedAtMs = 4_000)
        val second = signedEnvelope(signer, commandId = "cmd-dup", nonce = "n-b", issuedAtMs = 4_000)

        assertTrue(guard.verify(first).allowed)
        val dup = guard.verify(second)
        assertFalse(dup.allowed)
        assertEquals("duplicate_command_id", dup.reason)
    }
}
