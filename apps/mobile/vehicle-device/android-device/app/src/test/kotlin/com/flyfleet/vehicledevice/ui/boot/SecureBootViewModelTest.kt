package com.flyfleet.vehicledevice.ui.boot

import android.app.Application
import android.content.Context
import android.content.pm.PackageInfo
import android.content.pm.PackageManager
import app.cash.turbine.test
import com.flyfleet.vehicledevice.security.IntegrityManager
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import kotlin.time.Duration.Companion.seconds

@OptIn(ExperimentalCoroutinesApi::class)
class SecureBootViewModelTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockApp: Application
    private lateinit var mockPackageManager: PackageManager
    private lateinit var mockPackageInfo: PackageInfo

    @Before
    fun setup() {
        mockApp = mockk(relaxed = true)
        mockPackageManager = mockk(relaxed = true)
        mockPackageInfo = mockk(relaxed = true)

        every { mockApp.packageManager } returns mockPackageManager
        every { mockApp.packageName } returns "com.flyfleet.vehicledevice"
        every { mockPackageManager.getPackageInfo(any<String>(), any<Int>()) } returns mockPackageInfo
        mockPackageInfo.versionName = "1.0.0"

        // Mock string resources
        every { mockApp.getString(any()) } returns "MockString"
        every { mockApp.getString(com.flyfleet.vehicledevice.R.string.boot_status_verified) } returns "Verified"
        every { mockApp.getString(com.flyfleet.vehicledevice.R.string.boot_status_failed_time) } returns "Boot Failed: Time Sync"
        every { mockApp.getString(com.flyfleet.vehicledevice.R.string.boot_status_failed_integrity) } returns "Boot Failed: Integrity"

        mockkObject(IntegrityManager)
    }

    @After
    fun tearDown() {
        unmockkAll()
    }

    @Test
    fun `boot sequence success emits BOOT_SUCCESS and transitions to VERIFIED`() = runTest {
        coEvery { IntegrityManager.checkHardwareRoot(any()) } returns Pair(true, "Mock Hardware")
        coEvery { IntegrityManager.checkOsIntegrity(any()) } returns Pair(true, "Mock OS")
        coEvery { IntegrityManager.checkAppSignature(any()) } returns Pair(true, "Mock Sig")
        coEvery { IntegrityManager.checkSecureStorage(any()) } returns Pair(true, "Mock Storage")
        coEvery { IntegrityManager.checkCryptoKeys(any()) } returns Pair(true, "Mock Keys")
        coEvery { IntegrityManager.checkTimeSync(any()) } returns Pair(true, "Mock Time")
        coEvery { IntegrityManager.checkSensors(any()) } returns Pair(true, "Mock Sensors")
        coEvery { IntegrityManager.checkStorage(any()) } returns Pair(true, "Mock Storage Reserve")

        val viewModel = SecureBootViewModel(mockApp, isTesting = true)

        viewModel.uiEvents.test {
            viewModel.startBootSequence()
            assertEquals(BootEvent.BOOT_SUCCESS, awaitItem())
            cancelAndIgnoreRemainingEvents()
        }

        assertEquals(BootStatus.VERIFIED, viewModel.uiState.value.status)
        assertEquals("Verified", viewModel.uiState.value.bootStatusMsg)
    }

    @Test
    fun `boot sequence failure on time sync emits TIME_SYNC_FAILED and stops`() = runTest {
        coEvery { IntegrityManager.checkHardwareRoot(any()) } returns Pair(true, "")
        coEvery { IntegrityManager.checkOsIntegrity(any()) } returns Pair(true, "")
        coEvery { IntegrityManager.checkAppSignature(any()) } returns Pair(true, "")
        coEvery { IntegrityManager.checkSecureStorage(any()) } returns Pair(true, "")
        coEvery { IntegrityManager.checkCryptoKeys(any()) } returns Pair(true, "")
        coEvery { IntegrityManager.checkTimeSync(any()) } returns Pair(false, "Time Drift > 10m")

        val viewModel = SecureBootViewModel(mockApp, isTesting = true)

        viewModel.uiEvents.test {
            viewModel.startBootSequence()
            assertEquals(BootEvent.TIME_SYNC_FAILED, awaitItem())
            cancelAndIgnoreRemainingEvents()
        }

        assertEquals(BootStatus.TIME_SYNC_FAILURE, viewModel.uiState.value.status)
    }

    @Test
    fun `boot sequence failure on OS Integrity emits TAMPER_DETECTED`() = runTest {
        coEvery { IntegrityManager.checkHardwareRoot(any()) } returns Pair(true, "")
        coEvery { IntegrityManager.checkOsIntegrity(any()) } returns Pair(false, "Root access detected: /sbin/su")

        val viewModel = SecureBootViewModel(mockApp, isTesting = true)

        viewModel.uiEvents.test {
            viewModel.startBootSequence()
            assertEquals(BootEvent.TAMPER_DETECTED, awaitItem())
            cancelAndIgnoreRemainingEvents()
        }

        assertEquals(BootStatus.INTEGRITY_FAILURE, viewModel.uiState.value.status)
        assertEquals("Root access detected: /sbin/su", viewModel.uiState.value.integrityFailure?.tamperType)
    }
}
