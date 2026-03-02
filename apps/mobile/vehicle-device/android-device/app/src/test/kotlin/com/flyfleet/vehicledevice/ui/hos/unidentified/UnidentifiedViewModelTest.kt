package com.flyfleet.vehicledevice.ui.hos.unidentified

import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class UnidentifiedViewModelTest {

    private lateinit var viewModel: UnidentifiedViewModel

    @Before
    fun setup() {
        viewModel = UnidentifiedViewModel()
    }

    @Test
    fun `mock events are loaded on init`() = runTest {
        assertTrue(viewModel.uiState.value.events.isNotEmpty())
    }

    @Test
    fun `ToggleSelection updates selected events`() = runTest {
        val eventId = 101L
        viewModel.onEvent(UnidentifiedEventTrigger.ToggleSelection(eventId))
        assertTrue(viewModel.uiState.value.selectedEvents.contains(eventId))
        
        viewModel.onEvent(UnidentifiedEventTrigger.ToggleSelection(eventId))
        assertFalse(viewModel.uiState.value.selectedEvents.contains(eventId))
    }

    @Test
    fun `LoadEvents reloads mock data`() = runTest {
        viewModel.onEvent(UnidentifiedEventTrigger.LoadEvents)
        assertTrue(viewModel.uiState.value.events.isNotEmpty())
    }
}
