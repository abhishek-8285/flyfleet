package com.fleet.driver.data.repository

import com.fleet.driver.data.remote.api.VehicleService
import com.fleet.driver.data.remote.model.*
import com.fleet.driver.util.Resource
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Vehicle Repository
 * Handles all vehicle-related data operations
 */
@Singleton
class VehicleRepository @Inject constructor(
    private val vehicleService: VehicleService
) {
    
    /**
     * Get list of vehicles with filtering
     */
    suspend fun getVehicles(filter: VehicleFilter = VehicleFilter()): Resource<VehicleListResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getVehicles(
                    status = filter.status,
                    type = filter.type,
                    search = filter.search,
                    page = filter.page,
                    pageSize = filter.pageSize
                )
                
                if (response.isSuccessful) {
                    response.body()?.let { 
                        Resource.Success(it)
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to fetch vehicles")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Get specific vehicle by ID
     */
    suspend fun getVehicle(vehicleId: String): Resource<Vehicle> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getVehicle(vehicleId)
                
                if (response.isSuccessful) {
                    response.body()?.let {
                        Resource.Success(it)
                    } ?: Resource.Error("Vehicle not found")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to fetch vehicle")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Get vehicles assigned to current driver
     */
    suspend fun getMyVehicles(): Resource<List<Vehicle>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getMyVehicles()
                
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to fetch my vehicles")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Assign vehicle to driver
     */
    suspend fun assignVehicle(vehicleId: String, driverId: String): Resource<VehicleAssignmentResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val request = AssignVehicleRequest(vehicleId, driverId)
                val response = vehicleService.assignVehicle(vehicleId, request)
                
                if (response.isSuccessful) {
                    response.body()?.let {
                        Resource.Success(it)
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to assign vehicle")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Unassign vehicle from driver
     */
    suspend fun unassignVehicle(vehicleId: String): Resource<Unit> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.unassignVehicle(vehicleId)
                
                if (response.isSuccessful) {
                    Resource.Success(Unit)
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to unassign vehicle")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Update vehicle status
     */
    suspend fun updateVehicleStatus(vehicleId: String, status: String): Resource<Vehicle> {
        return withContext(Dispatchers.IO) {
            try {
                val request = VehicleStatusUpdateRequest(status)
                val response = vehicleService.updateVehicleStatus(vehicleId, request)
                
                if (response.isSuccessful) {
                    response.body()?.let {
                        Resource.Success(it)
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to update vehicle status")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Get vehicle telemetry data
     */
    suspend fun getVehicleTelemetry(
        vehicleId: String, 
        from: String? = null, 
        to: String? = null
    ): Resource<List<VehicleTelemetry>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getVehicleTelemetry(vehicleId, from, to)
                
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to fetch telemetry data")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Get vehicle assignment history
     */
    suspend fun getVehicleAssignments(vehicleId: String): Resource<List<VehicleAssignment>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getVehicleAssignments(vehicleId)
                
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to fetch assignment history")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Report vehicle issue
     */
    suspend fun reportVehicleIssue(
        vehicleId: String,
        title: String,
        description: String,
        priority: String = "medium",
        category: String = "other",
        images: List<String>? = null
    ): Resource<VehicleIssueResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val request = VehicleIssueRequest(title, description, priority, category, images)
                val response = vehicleService.reportVehicleIssue(vehicleId, request)
                
                if (response.isSuccessful) {
                    response.body()?.let {
                        Resource.Success(it)
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to report issue")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Get vehicle documents
     */
    suspend fun getVehicleDocuments(vehicleId: String): Resource<List<VehicleDocument>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getVehicleDocuments(vehicleId)
                
                if (response.isSuccessful) {
                    Resource.Success(response.body() ?: emptyList())
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to fetch documents")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Upload vehicle document
     */
    suspend fun uploadVehicleDocument(
        vehicleId: String,
        documentType: String,
        documentNumber: String,
        issueDate: String,
        expiryDate: String,
        fileUrl: String
    ): Resource<VehicleDocument> {
        return withContext(Dispatchers.IO) {
            try {
                val request = UploadDocumentRequest(documentType, documentNumber, issueDate, expiryDate, fileUrl)
                val response = vehicleService.uploadVehicleDocument(vehicleId, request)
                
                if (response.isSuccessful) {
                    response.body()?.let {
                        Resource.Success(it)
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to upload document")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Get vehicles by status
     */
    suspend fun getVehiclesByStatus(status: String): Resource<List<Vehicle>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getVehicles(status = status, pageSize = 100)
                
                if (response.isSuccessful) {
                    Resource.Success(response.body()?.vehicles ?: emptyList())
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to fetch vehicles")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
    
    /**
     * Search vehicles
     */
    suspend fun searchVehicles(query: String): Resource<List<Vehicle>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = vehicleService.getVehicles(search = query, pageSize = 50)
                
                if (response.isSuccessful) {
                    Resource.Success(response.body()?.vehicles ?: emptyList())
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Failed to search vehicles")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }
}