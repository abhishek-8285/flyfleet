package com.fleet.driver.presentation.vehicle.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.fleet.driver.data.remote.model.*
import com.fleet.driver.data.repository.VehicleRepository
import com.fleet.driver.util.Resource
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * Vehicle ViewModel
 * Handles vehicle management logic and state
 */
@HiltViewModel
class VehicleViewModel @Inject constructor(
    private val vehicleRepository: VehicleRepository
) : ViewModel() {
    
    // Vehicle list state
    private val _vehicles = MutableStateFlow<List<Vehicle>>(emptyList())
    val vehicles: StateFlow<List<Vehicle>> = _vehicles.asStateFlow()
    
    // Selected vehicle state
    private val _selectedVehicle = MutableStateFlow<Vehicle?>(null)
    val selectedVehicle: StateFlow<Vehicle?> = _selectedVehicle.asStateFlow()
    
    // My vehicles state
    private val _myVehicles = MutableStateFlow<List<Vehicle>>(emptyList())
    val myVehicles: StateFlow<List<Vehicle>> = _myVehicles.asStateFlow()
    
    // Loading state
    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()
    
    // Error state
    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage.asStateFlow()
    
    // Search query
    private val _searchQuery = MutableStateFlow("")
    val searchQuery: StateFlow<String> = _searchQuery.asStateFlow()
    
    // Filter state
    private val _currentFilter = MutableStateFlow(VehicleFilter())
    val currentFilter: StateFlow<VehicleFilter> = _currentFilter.asStateFlow()
    
    // Telemetry data
    private val _vehicleTelemetry = MutableStateFlow<List<VehicleTelemetry>>(emptyList())
    val vehicleTelemetry: StateFlow<List<VehicleTelemetry>> = _vehicleTelemetry.asStateFlow()
    
    // Assignment history
    private val _assignmentHistory = MutableStateFlow<List<VehicleAssignment>>(emptyList())
    val assignmentHistory: StateFlow<List<VehicleAssignment>> = _assignmentHistory.asStateFlow()
    
    init {
        loadMyVehicles()
    }
    
    /**
     * Load vehicles with current filter
     */
    fun loadVehicles(filter: VehicleFilter = _currentFilter.value) {
        viewModelScope.launch {
            _isLoading.value = true
            _currentFilter.value = filter
            
            when (val result = vehicleRepository.getVehicles(filter)) {
                is Resource.Success -> {
                    result.data?.let { _vehicles.value = it.vehicles }
                    _isLoading.value = false
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
                else -> {}
            }
        }
    }
    
    /**
     * Load vehicles assigned to current driver
     */
    fun loadMyVehicles() {
        viewModelScope.launch {
            _isLoading.value = true
            
            when (val result = vehicleRepository.getMyVehicles()) {
                is Resource.Success -> {
                    _myVehicles.value = result.data ?: emptyList()
                    _isLoading.value = false
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
                else -> {}
            }
        }
    }
    
    /**
     * Select a vehicle
     */
    fun selectVehicle(vehicleId: String) {
        viewModelScope.launch {
            when (val result = vehicleRepository.getVehicle(vehicleId)) {
                is Resource.Success -> {
                    result.data?.let {
                        _selectedVehicle.value = it
                        loadVehicleDetails(it)
                    }
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                }
                else -> {}
            }
        }
    }
    
    /**
     * Load detailed information for selected vehicle
     */
    private fun loadVehicleDetails(vehicle: Vehicle) {
        viewModelScope.launch {
            // Load telemetry data
            loadVehicleTelemetry(vehicle.id)
            
            // Load assignment history
            loadAssignmentHistory(vehicle.id)
        }
    }
    
    /**
     * Load vehicle telemetry data
     */
    private suspend fun loadVehicleTelemetry(vehicleId: String) {
        when (val result = vehicleRepository.getVehicleTelemetry(vehicleId)) {
            is Resource.Success -> {
                _vehicleTelemetry.value = result.data ?: emptyList()
            }
            is Resource.Error -> {
                _errorMessage.value = result.message
            }
            else -> {}
        }
    }
    
    /**
     * Load assignment history
     */
    private suspend fun loadAssignmentHistory(vehicleId: String) {
        when (val result = vehicleRepository.getVehicleAssignments(vehicleId)) {
            is Resource.Success -> {
                _assignmentHistory.value = result.data ?: emptyList()
            }
            is Resource.Error -> {
                _errorMessage.value = result.message
            }
            else -> {}
        }
    }
    
    /**
     * Assign vehicle to driver
     */
    fun assignVehicle(vehicleId: String, driverId: String) {
        viewModelScope.launch {
            _isLoading.value = true
            
            when (val result = vehicleRepository.assignVehicle(vehicleId, driverId)) {
                is Resource.Success -> {
                    _isLoading.value = false
                    // Refresh vehicle lists
                    loadMyVehicles()
                    loadVehicles(_currentFilter.value)
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
                else -> {}
            }
        }
    }
    
    /**
     * Unassign vehicle from driver
     */
    fun unassignVehicle(vehicleId: String) {
        viewModelScope.launch {
            _isLoading.value = true
            
            when (val result = vehicleRepository.unassignVehicle(vehicleId)) {
                is Resource.Success -> {
                    _isLoading.value = false
                    // Refresh vehicle lists
                    loadMyVehicles()
                    loadVehicles(_currentFilter.value)
                    _selectedVehicle.value = null
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
                else -> {}
            }
        }
    }
    
    /**
     * Update vehicle status
     */
    fun updateVehicleStatus(vehicleId: String, status: String) {
        viewModelScope.launch {
            _isLoading.value = true
            
            when (val result = vehicleRepository.updateVehicleStatus(vehicleId, status)) {
                is Resource.Success -> {
                    _isLoading.value = false
                    // Update local state
                    result.data?.let { _selectedVehicle.value = it }
                    // Refresh lists
                    loadMyVehicles()
                    loadVehicles(_currentFilter.value)
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
                else -> {}
            }
        }
    }
    
    /**
     * Report vehicle issue
     */
    fun reportVehicleIssue(
        vehicleId: String,
        title: String,
        description: String,
        priority: String = "medium",
        category: String = "other"
    ) {
        viewModelScope.launch {
            _isLoading.value = true
            
            when (val result = vehicleRepository.reportVehicleIssue(
                vehicleId, title, description, priority, category
            )) {
                is Resource.Success -> {
                    _isLoading.value = false
                    // Show success message
                    _errorMessage.value = "Issue reported successfully"
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
                else -> {}
            }
        }
    }
    
    /**
     * Search vehicles
     */
    fun searchVehicles(query: String) {
        _searchQuery.value = query
        if (query.isNotEmpty()) {
            viewModelScope.launch {
                _isLoading.value = true
                
                when (val result = vehicleRepository.searchVehicles(query)) {
                    is Resource.Success -> {
                        _vehicles.value = result.data ?: emptyList()
                        _isLoading.value = false
                    }
                    is Resource.Error -> {
                        _errorMessage.value = result.message
                        _isLoading.value = false
                    }
                    else -> {}
                }
            }
        } else {
            // Load all vehicles with current filter
            loadVehicles(_currentFilter.value)
        }
    }
    
    /**
     * Filter vehicles by status
     */
    fun filterByStatus(status: String?) {
        val newFilter = _currentFilter.value.copy(status = status)
        loadVehicles(newFilter)
    }
    
    /**
     * Filter vehicles by type
     */
    fun filterByType(type: String?) {
        val newFilter = _currentFilter.value.copy(type = type)
        loadVehicles(newFilter)
    }
    
    /**
     * Refresh all data
     */
    fun refresh() {
        loadVehicles(_currentFilter.value)
        loadMyVehicles()
        _selectedVehicle.value?.let { loadVehicleDetails(it) }
    }
    
    /**
     * Clear error message
     */
    fun clearError() {
        _errorMessage.value = null
    }
    
    /**
     * Clear search
     */
    fun clearSearch() {
        _searchQuery.value = ""
        loadVehicles(_currentFilter.value)
    }
    
    /**
     * Get vehicle by ID from loaded vehicles
     */
    fun getVehicleById(vehicleId: String): Vehicle? {
        return _vehicles.value.find { it.id == vehicleId }
    }
    
    /**
     * Check if vehicle is assigned to current driver
     */
    fun isVehicleAssignedToMe(vehicleId: String): Boolean {
        return _myVehicles.value.any { it.id == vehicleId }
    }
    
    /**
     * Get vehicle status color
     */
    fun getVehicleStatusColor(status: String): androidx.compose.ui.graphics.Color {
        return when (status.lowercase()) {
            "available" -> androidx.compose.ui.graphics.Color.Green
            "in_use" -> androidx.compose.ui.graphics.Color.Blue
            "maintenance" -> androidx.compose.ui.graphics.Color.Yellow
            "retired" -> androidx.compose.ui.graphics.Color.Red
            else -> androidx.compose.ui.graphics.Color.Gray
        }
    }
    
    /**
     * Get vehicle type icon
     */
    fun getVehicleTypeIcon(type: String): String {
        return when (type.lowercase()) {
            "truck" -> "🚚"
            "van" -> "🚐"
            "car" -> "🚗"
            "motorcycle" -> "🏍️"
            else -> "🚙"
        }
    }
}