package com.flyfleet.vehicledevice.ui.diagnostics

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class DiagnosticsViewModelTest {

    private lateinit var viewModel: DiagnosticsViewModel
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        viewModel = DiagnosticsViewModel()
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial state contains mock sync queue items`() = runTest {
        val state = viewModel.uiState.value
        assertFalse(state.syncQueue.isEmpty())
        assertEquals(3, state.syncQueue.size)
    }

    @Test
    fun `switch screen updates screenState in UI state`() = runTest {
        viewModel.onEvent(DiagnosticsEvent.SwitchScreen(DiagnosticsScreenState.SYNC_QUEUE_MANAGER))
        assertEquals(DiagnosticsScreenState.SYNC_QUEUE_MANAGER, viewModel.uiState.value.screenState)
    }

    @Test
    fun `total queue size is correctly calculated`() = runTest {
        val expectedSize = viewModel.uiState.value.syncQueue.sumOf { it.sizeBytes }
        assertEquals(expectedSize, viewModel.uiState.value.totalQueueSize)
    }
}
