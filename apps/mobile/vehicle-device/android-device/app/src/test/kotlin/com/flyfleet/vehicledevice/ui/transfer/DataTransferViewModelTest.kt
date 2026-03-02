package com.flyfleet.vehicledevice.ui.transfer

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class DataTransferViewModelTest {

    private lateinit var viewModel: DataTransferViewModel
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        viewModel = DataTransferViewModel()
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `MethodSelected transitions to CONFIGURE_EXPORT`() = runTest {
        viewModel.onEvent(TransferEvent.MethodSelected(TransferMethod.WEB_SERVICE))
        assertEquals(TransferMethod.WEB_SERVICE, viewModel.uiState.value.selectedMethod)
        assertEquals(TransferScreenState.CONFIGURE_EXPORT, viewModel.uiState.value.screenState)
    }

    @Test
    fun `VerifyCompliance starts verification flow`() = runTest {
        viewModel.onEvent(TransferEvent.VerifyCompliance)
        assertEquals(TransferScreenState.COMPLIANCE_GATE, viewModel.uiState.value.screenState)
        
        advanceTimeBy(2500)
        assertTrue(viewModel.uiState.value.isVerified)
    }

    @Test
    fun `StartTransfer simulates progress and success`() = runTest {
        // Select Method -> Configure Export
        viewModel.onEvent(TransferEvent.MethodSelected(TransferMethod.WEB_SERVICE))
        viewModel.onEvent(TransferEvent.StartTransfer)
        assertEquals(TransferScreenState.IN_PROGRESS, viewModel.uiState.value.screenState)
        
        // Advance time to complete progress
        advanceTimeBy(6000)
        
        assertEquals(1.0f, viewModel.uiState.value.progress)
        assertEquals(TransferScreenState.SUCCESS_RECEIPT, viewModel.uiState.value.screenState)
        assertNotNull(viewModel.uiState.value.transferId)
        assertNotNull(viewModel.uiState.value.confirmationCode)
    }

    @Test
    fun `FinishFlow resets state`() = runTest {
        viewModel.onEvent(TransferEvent.FinishFlow)
        assertEquals(TransferScreenState.SELECT_METHOD, viewModel.uiState.value.screenState)
        assertNull(viewModel.uiState.value.selectedMethod)
    }
}
