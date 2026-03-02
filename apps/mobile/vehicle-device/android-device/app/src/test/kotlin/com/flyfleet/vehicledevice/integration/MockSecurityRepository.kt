package com.flyfleet.vehicledevice.integration

import com.flyfleet.vehicledevice.domain.repository.SecurityRepository

class MockSecurityRepository : SecurityRepository {
    private val data = mutableMapOf<String, String>()

    override fun getAuthToken(): String = data["session_token"] ?: ""
    override fun setAuthToken(token: String) { data["session_token"] = token }
    override fun clearAuthToken() { data.remove("session_token") }

    override fun getCurrentDriverId(): String = data["current_driver_id"] ?: ""
    override fun setCurrentDriverId(driverId: String) { 
        android.util.Log.e("DEBUG_NAV", "MockSecurityRepo: Setting driverId=$driverId")
        data["current_driver_id"] = driverId 
    }

    override fun getSecureString(key: String, defaultValue: String): String = data[key] ?: defaultValue
    override fun storeSecureString(key: String, value: String) { data[key] = value }
    override fun clearSecureData(key: String) { data.remove(key) }
}
