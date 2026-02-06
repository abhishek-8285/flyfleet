package com.fleet.driver.presentation.driver.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.fleet.driver.data.remote.model.*
import com.fleet.driver.data.repository.DriverRepository
import com.fleet.driver.util.Resource
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * Driver Profile ViewModel
 */
@HiltViewModel
class DriverViewModel @Inject constructor(
    private val driverRepository: DriverRepository
) : ViewModel() {
    
    private val _driverProfile = MutableStateFlow<DriverProfile?>(null)
    val driverProfile: StateFlow<DriverProfile?> = _driverProfile.asStateFlow()
    
    private val _analytics = MutableStateFlow<DriverAnalytics?>(null)
    val analytics: StateFlow<DriverAnalytics?> = _analytics.asStateFlow()
    
    private val _behaviorEvents = MutableStateFlow<List<BehaviorEvent>>(emptyList())
    val behaviorEvents: StateFlow<List<BehaviorEvent>> = _behaviorEvents.asStateFlow()
    
    private val _attendanceRecords = MutableStateFlow<List<AttendanceRecord>>(emptyList())
    val attendanceRecords: StateFlow<List<AttendanceRecord>> = _attendanceRecords.asStateFlow()
    
    private val _workSchedules = MutableStateFlow<List<WorkSchedule>>(emptyList())
    val workSchedules: StateFlow<List<WorkSchedule>> = _workSchedules.asStateFlow()
    
    private val _documents = MutableStateFlow<List<DriverDocument>>(emptyList())
    val documents: StateFlow<List<DriverDocument>> = _documents.asStateFlow()
    
    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()
    
    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage.asStateFlow()
    
    init {
        loadDriverData()
    }
    
    private fun loadDriverData() {
        loadProfile()
        loadAnalytics()
        loadBehaviorEvents()
        loadAttendance()
        loadSchedules()
        loadDocuments()
    }
    
    fun loadProfile() {
        viewModelScope.launch {
            _isLoading.value = true
            when (val result = driverRepository.getDriverProfile()) {
                is Resource.Success -> {
                    _driverProfile.value = result.data
                    _isLoading.value = false
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
            }
        }
    }
    
    fun loadAnalytics() {
        viewModelScope.launch {
            when (val result = driverRepository.getDriverAnalytics()) {
                is Resource.Success -> {
                    _analytics.value = result.data
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                }
            }
        }
    }
    
    fun loadBehaviorEvents(fromDate: String? = null, toDate: String? = null) {
        viewModelScope.launch {
            when (val result = driverRepository.getBehaviorEvents(fromDate, toDate)) {
                is Resource.Success -> {
                    _behaviorEvents.value = result.data
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                }
            }
        }
    }
    
    fun loadAttendance(fromDate: String? = null, toDate: String? = null) {
        viewModelScope.launch {
            when (val result = driverRepository.getAttendanceRecords(fromDate, toDate)) {
                is Resource.Success -> {
                    _attendanceRecords.value = result.data
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                }
            }
        }
    }
    
    fun loadSchedules() {
        viewModelScope.launch {
            when (val result = driverRepository.getWorkSchedules()) {
                is Resource.Success -> {
                    _workSchedules.value = result.data
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                }
            }
        }
    }
    
    fun loadDocuments() {
        viewModelScope.launch {
            when (val result = driverRepository.getDocuments()) {
                is Resource.Success -> {
                    _documents.value = result.data
                }
                is Resource.Error -> {
                    _errorMessage.value = result.message
                }
            }
        }
    }
    
    fun refreshAll() {
        loadDriverData()
    }
    
    fun clearError() {
        _errorMessage.value = null
    }
    
    fun getBehaviorScoreColor(score: Int): androidx.compose.ui.graphics.Color {
        return when {
            score >= 80 -> androidx.compose.ui.graphics.Color.Green
            score >= 60 -> androidx.compose.ui.graphics.Color.Yellow
            score >= 40 -> androidx.compose.ui.graphics.Color.Orange
            else -> androidx.compose.ui.graphics.Color.Red
        }
    }
    
    fun getRatingStars(rating: Double): String {
        val fullStars = rating.toInt()
        val hasHalfStar = rating % 1 >= 0.5
        return "★".repeat(fullStars) + if (hasHalfStar) "☆" else "" + "☆".repeat(5 - fullStars - if (hasHalfStar) 1 else 0)
    }
}