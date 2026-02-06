package com.fleet.driver.data.repository

import com.fleet.driver.data.remote.api.DriverService
import com.fleet.driver.data.remote.model.*
import com.fleet.driver.util.Resource
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Driver Repository
 * Handles all driver profile and management operations
 */
@Singleton
class DriverRepository @Inject constructor(
    private val driverService: DriverService
) {
    
    suspend fun getDriverProfile(): Resource<DriverProfile> {
        return withContext(Dispatchers.IO) {
            try {
                val response = driverService.getDriverProfile()
                if (response.isSuccessful) {
                    response.body()?.let { Resource.Success(it) } 
                        ?: Resource.Error("Empty response")
                } else {
                    Resource.Error(response.errorBody()?.string() ?: "Failed to fetch profile")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    suspend fun getDriverAnalytics(): Resource<DriverAnalytics> {
        return withContext(Dispatchers.IO) {
            try {
                val response = driverService.getDriverAnalytics()
                if (response.isSuccessful) {
                    response.body()?.let { Resource.Success(it) } 
                        ?: Resource.Error("Empty response")
                } else {
                    Resource.Error(response.errorBody()?.string() ?: "Failed to fetch analytics")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    suspend fun getBehaviorEvents(fromDate: String?, toDate: String?): Resource<List<BehaviorEvent>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = driverService.getBehaviorEvents(fromDate, toDate)
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    Resource.Error(response.errorBody()?.string() ?: "Failed to fetch behavior events")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    suspend fun getAttendanceRecords(fromDate: String?, toDate: String?): Resource<List<AttendanceRecord>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = driverService.getAttendanceRecords(fromDate, toDate)
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    Resource.Error(response.errorBody()?.string() ?: "Failed to fetch attendance")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    suspend fun getWorkSchedules(): Resource<List<WorkSchedule>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = driverService.getWorkSchedules()
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    Resource.Error(response.errorBody()?.string() ?: "Failed to fetch schedules")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    suspend fun getDocuments(): Resource<List<DriverDocument>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = driverService.getDocuments()
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    Resource.Error(response.errorBody()?.string() ?: "Failed to fetch documents")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
}