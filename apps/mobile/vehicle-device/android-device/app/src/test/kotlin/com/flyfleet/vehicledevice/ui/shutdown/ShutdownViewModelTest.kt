package com.flyfleet.vehicledevice.ui.shutdown

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class ShutdownViewModelTest {

    private lateinit var viewModel: ShutdownViewModel
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        viewModel = ShutdownViewModel()
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial state contains mock summary`() = runTest {
        assertNotNull(viewModel.uiState.value.summary)
        assertEquals("DRV-9921", viewModel.uiState.value.summary?.driverId)
    }

    @Test
    fun `Logout flow transitions through progress to complete`() = runTest {
        viewModel.onEvent(ShutdownEvent.CertifyAndLogout)
        runCurrent()
        
        assertEquals(ShutdownScreenState.SAFE_LOGOUT_PROGRESS, viewModel.uiState.value.screenState)
        assertTrue(viewModel.uiState.value.isProcessing)
        
        // Fast-forward through delays
        advanceTimeBy(4000)
        
        assertEquals(ShutdownScreenState.SHUTDOWN_COMPLETE, viewModel.uiState.value.screenState)
        assertFalse(viewModel.uiState.value.isProcessing)
        assertTrue(viewModel.uiState.value.navigateToLogin)
    }

    @Test
    fun `LoadSummary reloads data`() = runTest {
        viewModel.onEvent(ShutdownEvent.LoadSummary)
        assertNotNull(viewModel.uiState.value.summary)
    }
}
