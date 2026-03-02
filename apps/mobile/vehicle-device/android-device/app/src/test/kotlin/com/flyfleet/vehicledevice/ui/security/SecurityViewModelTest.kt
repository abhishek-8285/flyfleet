package com.flyfleet.vehicledevice.ui.security

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class SecurityViewModelTest {

    private lateinit var viewModel: SecurityViewModel
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        viewModel = SecurityViewModel()
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `TogglePrivacy updates state`() = runTest {
        val initial = viewModel.uiState.value.isPrivacyEnabled
        viewModel.onEvent(SecurityEvent.TogglePrivacy)
        assertNotEquals(initial, viewModel.uiState.value.isPrivacyEnabled)
    }

    @Test
    fun `Panic flow starts countdown and transitions to active`() = runTest {
        viewModel.onEvent(SecurityEvent.TriggerPanic)
        assertEquals(SecurityScreenState.PANIC_CONFIRM, viewModel.uiState.value.screenState)
        assertEquals(5, viewModel.uiState.value.panicCountdown)
        
        // Fast-forward 5 seconds
        advanceTimeBy(6000)
        
        assertEquals(SecurityScreenState.PANIC_ACTIVE, viewModel.uiState.value.screenState)
        assertTrue(viewModel.uiState.value.dispatchReference?.startsWith("EMER-") == true)
    }

    @Test
    fun `CancelPanic stops countdown`() = runTest {
        viewModel.onEvent(SecurityEvent.TriggerPanic)
        advanceTimeBy(2000)
        viewModel.onEvent(SecurityEvent.CancelPanic)
        
        assertEquals(SecurityScreenState.PRIVACY_TOGGLE, viewModel.uiState.value.screenState)
        
        // Advance time and check that it didn't transition to active
        advanceTimeBy(5000)
        assertEquals(SecurityScreenState.PRIVACY_TOGGLE, viewModel.uiState.value.screenState)
    }

    @Test
    fun `AcknowledgeTamper resets state`() = runTest {
        viewModel.onEvent(SecurityEvent.AcknowledgeTamper)
        assertFalse(viewModel.uiState.value.isTampered)
        assertEquals(SecurityScreenState.PRIVACY_TOGGLE, viewModel.uiState.value.screenState)
    }
}
