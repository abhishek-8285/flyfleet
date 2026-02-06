package com.fleet.driver.presentation.auth.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.fleet.driver.data.remote.model.AuthState
import com.fleet.driver.data.remote.model.EnrollmentState
import com.fleet.driver.data.remote.model.UserResponse
import com.fleet.driver.data.repository.AuthRepository
import com.fleet.driver.util.Resource
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * Authentication ViewModel
 * Handles authentication logic and state management
 */
@HiltViewModel
class AuthViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _authState = MutableStateFlow<AuthState>(AuthState.Unauthenticated)
    val authState: StateFlow<AuthState> = _authState.asStateFlow()

    private val _enrollmentState = MutableStateFlow<EnrollmentState>(EnrollmentState.Idle)
    val enrollmentState: StateFlow<EnrollmentState> = _enrollmentState.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage.asStateFlow()

    init {
        checkInitialAuthState()
    }

    /**
     * Check initial authentication state
     */
    private fun checkInitialAuthState() {
        if (authRepository.isAuthenticated()) {
            val user = authRepository.getCurrentUser()
            user?.let {
                _authState.value = AuthState.Authenticated(it)
            }
        }
    }

    /**
     * User login
     */
    fun login(email: String, password: String) {
        viewModelScope.launch {
            _isLoading.value = true
            _errorMessage.value = null
            
            when (val result = authRepository.login(email, password)) {
                is Resource.Success -> {
                    _authState.value = AuthState.Authenticated(result.data.user)
                    _isLoading.value = false
                    
                    // Register device for push notifications
                    registerDevice()
                }
                is Resource.Error -> {
                    _authState.value = AuthState.Error(result.message)
                    _errorMessage.value = result.message
                    _isLoading.value = false
                }
            }
        }
    }

    /**
     * Driver enrollment using QR code
     */
    fun enrollDriver(enrollmentCode: String) {
        viewModelScope.launch {
            _enrollmentState.value = EnrollmentState.Enrolling
            _errorMessage.value = null
            
            when (val result = authRepository.enrollDriver(enrollmentCode)) {
                is Resource.Success -> {
                    if (result.data.success) {
                        _enrollmentState.value = EnrollmentState.Success(result.data.driverId)
                        
                        // Get user info and update auth state
                        val user = authRepository.getCurrentUser()
                        user?.let {
                            _authState.value = AuthState.Authenticated(it)
                        }
                        
                        // Register device for push notifications
                        registerDevice()
                    } else {
                        _enrollmentState.value = EnrollmentState.Error("Enrollment failed")
                        _errorMessage.value = "Enrollment failed"
                    }
                }
                is Resource.Error -> {
                    _enrollmentState.value = EnrollmentState.Error(result.message)
                    _errorMessage.value = result.message
                }
            }
        }
    }

    /**
     * Register device for push notifications
     */
    private fun registerDevice() {
        viewModelScope.launch {
            when (val result = authRepository.registerDevice()) {
                is Resource.Success -> {
                    // Device registered successfully
                }
                is Resource.Error -> {
                    // Log error but don't block user flow
                    println("Device registration failed: ${result.message}")
                }
            }
        }
    }

    /**
     * Logout user
     */
    fun logout() {
        viewModelScope.launch {
            _isLoading.value = true
            
            when (val result = authRepository.logout()) {
                is Resource.Success -> {
                    _authState.value = AuthState.Unauthenticated
                    _isLoading.value = false
                }
                is Resource.Error -> {
                    // Even if API fails, clear local state
                    _authState.value = AuthState.Unauthenticated
                    _isLoading.value = false
                }
            }
        }
    }

    /**
     * Refresh authentication token
     */
    fun refreshToken() {
        viewModelScope.launch {
            when (val result = authRepository.refreshToken()) {
                is Resource.Success -> {
                    _authState.value = AuthState.Authenticated(result.data.user)
                }
                is Resource.Error -> {
                    _authState.value = AuthState.Unauthenticated
                }
            }
        }
    }

    /**
     * Clear error message
     */
    fun clearError() {
        _errorMessage.value = null
    }

    /**
     * Reset enrollment state
     */
    fun resetEnrollmentState() {
        _enrollmentState.value = EnrollmentState.Idle
    }

    /**
     * Check if user is authenticated
     */
    fun isAuthenticated(): Boolean {
        return authRepository.isAuthenticated()
    }

    /**
     * Get current user
     */
    fun getCurrentUser(): UserResponse? {
        return authRepository.getCurrentUser()
    }
}