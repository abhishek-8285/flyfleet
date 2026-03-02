package com.flyfleet.vehicledevice.integration

import com.flyfleet.vehicledevice.domain.model.DriverStatus
import com.flyfleet.vehicledevice.domain.model.HosStatusEvent
import com.flyfleet.vehicledevice.domain.repository.HosRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.map

class MockHosRepository : HosRepository {
    private val logs = MutableStateFlow<Map<String, List<HosStatusEvent>>>(emptyMap())

    override fun getLogsForDriver(driverId: String): Flow<List<HosStatusEvent>> {
        return logs.map { it[driverId] ?: emptyList() }
    }

    override suspend fun addLog(driverId: String, status: DriverStatus, timestamp: Long) {
        val currentLogs = logs.value.toMutableMap()
        val driverLogs = (currentLogs[driverId] ?: emptyList()).toMutableList()
        driverLogs.add(HosStatusEvent(
            id = (driverLogs.size + 1).toLong(),
            status = status,
            startTime = timestamp,
            endTime = null,
            location = "Mock Location"
        ))
        currentLogs[driverId] = driverLogs
        logs.value = currentLogs
    }

    override suspend fun getLastStatus(driverId: String): DriverStatus? {
        return logs.value[driverId]?.lastOrNull()?.status
    }
}
