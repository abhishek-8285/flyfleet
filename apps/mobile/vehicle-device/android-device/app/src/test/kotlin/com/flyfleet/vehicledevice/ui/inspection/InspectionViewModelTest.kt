package com.flyfleet.vehicledevice.ui.inspection

import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class InspectionViewModelTest {

    private lateinit var viewModel: InspectionViewModel

    @Before
    fun setup() {
        viewModel = InspectionViewModel()
    }

    @Test
    fun `initial state is INTRO`() = runTest {
        assertEquals(InspectionScreenState.INTRO, viewModel.uiState.value.screenState)
    }

    @Test
    fun `UpdateOfficerId updates state`() = runTest {
        viewModel.onEvent(InspectionEvent.UpdateOfficerId("OFF-123"))
        assertEquals("OFF-123", viewModel.uiState.value.officerId)
    }

    @Test
    fun `StartInspection transitions to OFFICER_ID_ENTRY`() = runTest {
        viewModel.onEvent(InspectionEvent.StartInspection)
        assertEquals(InspectionScreenState.OFFICER_ID_ENTRY, viewModel.uiState.value.screenState)
    }

    @Test
    fun `EnterLogViewer transitions to LOG_VIEWER`() = runTest {
        viewModel.onEvent(InspectionEvent.EnterLogViewer)
        assertEquals(InspectionScreenState.LOG_VIEWER, viewModel.uiState.value.screenState)
    }

    @Test
    fun `ExitInspection resets state`() = runTest {
        viewModel.onEvent(InspectionEvent.UpdateOfficerId("OFF-123"))
        viewModel.onEvent(InspectionEvent.ExitInspection)
        assertEquals(InspectionScreenState.INTRO, viewModel.uiState.value.screenState)
        assertEquals("", viewModel.uiState.value.officerId)
    }
}
