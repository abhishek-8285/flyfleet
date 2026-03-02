package com.flyfleet.vehicledevice.ui.login

import com.flyfleet.vehicledevice.domain.model.DriverInfo
import com.flyfleet.vehicledevice.domain.model.RosterDriver
import com.flyfleet.vehicledevice.domain.repository.DriverAuthRepository

class MockDriverAuthRepository(
    private val mockRoster: List<RosterDriver> = emptyList(),
    private val mockDriverInfo: DriverInfo? = null,
    private val shouldSucceed: Boolean = true
) : DriverAuthRepository {

    override suspend fun getDriverRoster(vehicleId: String): Result<List<RosterDriver>> {
        return if (shouldSucceed) Result.success(mockRoster)
        else Result.failure(Exception("Failed to load roster"))
    }

    override suspend fun pinLogin(
        deviceId: String,
        vehicleId: String,
        driverId: String,
        pin: String,
        gpsLat: Double?,
        gpsLon: Double?
    ): Result<DriverInfo> {
        return if (shouldSucceed && pin == "123456" && mockDriverInfo != null) {
            Result.success(mockDriverInfo)
        } else {
            Result.failure(Exception("Invalid PIN"))
        }
    }

    override suspend fun rfidLogin(
        deviceId: String,
        vehicleId: String,
        rfidTag: String,
        gpsLat: Double?,
        gpsLon: Double?
    ): Result<DriverInfo> {
        return if (shouldSucceed && mockDriverInfo != null) {
            Result.success(mockDriverInfo)
        } else {
            Result.failure(Exception("RFID login failed"))
        }
    }

    override suspend fun checkActiveSession(driverId: String): Result<Boolean> {
        return Result.success(false)
    }

    override suspend fun createSession(
        driverId: String,
        vehicleId: String,
        deviceId: String
    ): Result<String> {
        return Result.success("MOCK-SESSION-ID")
    }

    override suspend fun logout(sessionId: String): Result<Unit> {
        return Result.success(Unit)
    }
}
