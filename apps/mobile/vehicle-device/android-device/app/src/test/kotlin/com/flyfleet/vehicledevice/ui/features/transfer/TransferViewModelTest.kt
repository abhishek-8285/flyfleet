package com.flyfleet.vehicledevice.ui.features.transfer

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class TransferViewModelTest {

    private lateinit var viewModel: TransferViewModel
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        viewModel = TransferViewModel()
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial state is IDLE and not verified`() = runTest {
        val state = viewModel.state.value
        assertEquals(TransferStatus.IDLE, state.status)
        assertFalse(state.isVerified)
        assertNull(state.selectedMethod)
    }

    @Test
    fun `verify pin with target code updates isVerified`() = runTest {
        viewModel.onEvent(TransferEvent.VerifyPin("1234"))
        assertTrue(viewModel.state.value.isVerified)
    }

    @Test
    fun `verify pin with wrong code does not update isVerified`() = runTest {
        viewModel.onEvent(TransferEvent.VerifyPin("0000"))
        assertFalse(viewModel.state.value.isVerified)
    }

    @Test
    fun `select method updates state`() = runTest {
        viewModel.onEvent(TransferEvent.SelectMethod(TransferMethod.WEB_SERVICE))
        assertEquals(TransferMethod.WEB_SERVICE, viewModel.state.value.selectedMethod)
    }

    @Test
    fun `start transfer simulates progress and ends with success`() = runTest {
        viewModel.onEvent(TransferEvent.StartTransfer)
        runCurrent()
        
        // Initial state
        assertEquals(TransferStatus.IN_PROGRESS, viewModel.state.value.status)
        
        // Advance time to complete simulation
        advanceUntilIdle()
        
        assertEquals(1.0f, viewModel.state.value.progress)
        assertEquals(TransferStatus.SUCCESS, viewModel.state.value.status)
    }
}
