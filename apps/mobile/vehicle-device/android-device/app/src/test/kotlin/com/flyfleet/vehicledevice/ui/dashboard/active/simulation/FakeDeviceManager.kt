package com.flyfleet.vehicledevice.ui.dashboard.active.simulation

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** Fake Device Manager for injecting failure states (Storage, Network, Clock tampering). */
class FakeDeviceManager {

    private val _storagePercent = MutableStateFlow(100)
    val storagePercent: StateFlow<Int> = _storagePercent

    private val _networkConnected = MutableStateFlow(true)
    val networkConnected: StateFlow<Boolean> = _networkConnected

    private val _timeDriftMs = MutableStateFlow(0L)
    val timeDriftMs: StateFlow<Long> = _timeDriftMs

    private val _isTampered = MutableStateFlow(false)
    val isTampered: StateFlow<Boolean> = _isTampered

    fun simulateStorageFull() {
        _storagePercent.value = 3 // Drops below 5%
    }

    fun simulateNetworkLoss() {
        _networkConnected.value = false
    }

    fun simulateClockTamper(driftMs: Long) {
        _timeDriftMs.value = driftMs
    }

    fun simulateRootDetected() {
        _isTampered.value = true
    }
}
