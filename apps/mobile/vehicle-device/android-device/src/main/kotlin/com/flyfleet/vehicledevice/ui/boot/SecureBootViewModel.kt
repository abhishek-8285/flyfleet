package com.flyfleet.vehicledevice.ui.boot

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.delay
import com.flyfleet.vehicledevice.security.IntegrityManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDateTime
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter

class SecureBootViewModel : ViewModel() {

    private val _uiState = MutableStateFlow(BootState())
    val uiState: StateFlow<BootState> = _uiState.asStateFlow()

    init {
        initializeChecks()
        startBootSequence()
        StartTimeUpdates()
    }

    private fun initializeChecks() {
        _uiState.update { currentState ->
            currentState.copy(
                checks = listOf(
                    SystemCheck("hardware_root", "Hardware Root", "TPM 2.0 Validated"),
                    SystemCheck("os_integrity", "OS Integrity", "Kernel signatures match"),
                    SystemCheck("app_signature", "App Signature", "Cert: Apple Distribution"),
                    SystemCheck("secure_storage", "Secure Storage", "Encrypted partition mounted"),
                    SystemCheck("crypto_keys", "Crypto Keys", "Keys loaded"),
                    SystemCheck("time_sync", "Time Sync", "Time verified"),
                    SystemCheck("sensor_init", "Sensor Initialization", "Sensors initialized"),
                    SystemCheck("storage_val", "Storage Validation", "Evidence storage OK")
                )
            )
        }
    }

    private fun startBootSequence() {
        viewModelScope.launch {
            val checkFunctions = listOf(
                "hardware_root" to IntegrityManager::checkHardwareRoot,
                "os_integrity" to IntegrityManager::checkOsIntegrity,
                "app_signature" to IntegrityManager::checkAppSignature,
                "secure_storage" to IntegrityManager::checkSecureStorage,
                "crypto_keys" to IntegrityManager::checkCryptoKeys,
                "time_sync" to IntegrityManager::checkTimeSync,
                "sensor_init" to IntegrityManager::checkSensors,
                "storage_val" to IntegrityManager::checkStorage
            )
            
            val totalChecks = checkFunctions.size
            
            checkFunctions.forEachIndexed { index, (id, checkFunc) ->
                // Mark check as running
                updateCheckStatus(id, CheckStatus.RUNNING)
                
                // Perform check
                val isSuccess = checkFunc()
                
                // Update status based on result
                val status = if (isSuccess) CheckStatus.VERIFIED else CheckStatus.FAILED
                updateCheckStatus(id, status)
                
                // Update progress
                val progress = (index + 1) / totalChecks.toFloat()
                _uiState.update { it.copy(overallProgress = progress) }
                
                if (!isSuccess) {
                    // Stop sequence on failure (or handle retry logic)
                    _uiState.update { it.copy(bootStatus = "Boot Failed") }
                    return@launch
                }
            }

            _uiState.update { 
                it.copy(
                    isSystemVerified = true,
                    bootStatus = "Verified"
                ) 
            }
        }
    }

    private fun updateCheckStatus(checkId: String, status: CheckStatus) {
        _uiState.update { currentState ->
            val updatedChecks = currentState.checks.map {
                if (it.id == checkId) it.copy(status = status) else it
            }
            currentState.copy(checks = updatedChecks)
        }
    }

    private fun StartTimeUpdates() {
        viewModelScope.launch {
            while (true) {
                val now = LocalDateTime.now(ZoneOffset.UTC)
                val formatter = DateTimeFormatter.ofPattern("HH:mm:ss 'UTC'")
                _uiState.update { it.copy(currentTime = now.format(formatter)) }
                delay(1000)
            }
        }
    }

    fun onLaunchDashboard() {
        // Handle navigation or other actions
    }
}
