package com.flyfleet.vehicledevice.ui.boot

import android.app.Application
import com.flyfleet.vehicledevice.domain.repository.DeviceRepository
import com.flyfleet.vehicledevice.domain.repository.DeviceStatus as DeviceStatusModel
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@ExperimentalCoroutinesApi
@RunWith(RobolectricTestRunner::class)
@Config(sdk = [31])
class SecureBootViewModelTest {

    private lateinit var viewModel: SecureBootViewModel
    private val mockApplication = mockk<Application>(relaxed = true)
    private val mockDeviceRepository = mockk<DeviceRepository>(relaxed = true)
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        mockkObject(com.flyfleet.vehicledevice.security.IntegrityManager)
        
        // Mock strings to avoid R.string dependency issues in pure JVM test
        every { mockApplication.getString(any()) } returns "Mock String"
        every { mockDeviceRepository.getDeviceId() } returns "FLT-TEST-123"
        every { mockDeviceRepository.getDeviceStatus() } returns flowOf(DeviceStatusModel(isRegistered = true, isAttested = true, isOnline = true))

        coEvery { com.flyfleet.vehicledevice.security.IntegrityManager.checkHardwareRoot(any()) } returns Pair(true, "Mock HW")
        coEvery { com.flyfleet.vehicledevice.security.IntegrityManager.checkOsIntegrity(any()) } returns Pair(true, "Mock OS")
        coEvery { com.flyfleet.vehicledevice.security.IntegrityManager.checkAppSignature(any()) } returns Pair(true, "Mock Sig")
        coEvery { com.flyfleet.vehicledevice.security.IntegrityManager.checkSecureStorage(any()) } returns Pair(true, "Mock Storage")
        coEvery { com.flyfleet.vehicledevice.security.IntegrityManager.checkCryptoKeys(any()) } returns Pair(true, "Mock Crypto")
        coEvery { com.flyfleet.vehicledevice.security.IntegrityManager.checkTimeSync(any()) } returns Pair(true, "Mock Time")
        coEvery { com.flyfleet.vehicledevice.security.IntegrityManager.checkStorage(any()) } returns Pair(true, "Mock Storage")
        
        viewModel = SecureBootViewModel(mockApplication, mockDeviceRepository)
    }

    @After
    fun tearDown() {
        unmockkAll()
        Dispatchers.resetMain()
    }

    @Test
    fun `initial state contains app version and device ID`() = runTest {
        assertEquals("FLT-TEST-123", viewModel.uiState.value.deviceId)
        assertTrue(viewModel.uiState.value.checks.isNotEmpty())
    }

    @Test
    fun `boot sequence completes with success when all integrity checks pass`() = runTest {
        // IntegrityManager is an object, so its static methods might be called.
        // In a pure JVM test without mockkObject, we rely on the implementation.
        // If it uses Build.BOARD etc, it might fail in JVM. 
        // We'll see if we need mockkObject(IntegrityManager).
        
        viewModel.startBootSequence()
        advanceUntilIdle()
        
        assertEquals(BootStatus.VERIFIED, viewModel.uiState.value.status)
        assertEquals(1.0f, viewModel.uiState.value.overallProgress)
    }

    @Test
    fun `onLaunchDashboard updates state`() = runTest {
        viewModel.onLaunchDashboard()
        assertTrue(viewModel.uiState.value.navigateToDashboard)
    }

    @Test
    fun `retryBootSequence resets state and starts sequence`() = runTest {
        viewModel.retryBootSequence()
        advanceUntilIdle()
        assertEquals(BootStatus.VERIFIED, viewModel.uiState.value.status)
    }
}
