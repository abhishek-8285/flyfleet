package com.flyfleet.vehicledevice.safety

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.flyfleet.vehicledevice.telemetry.simulation.FakeTelemetryService
import com.flyfleet.vehicledevice.ui.MainActivity
import dagger.hilt.android.testing.HiltAndroidRule
import dagger.hilt.android.testing.HiltAndroidTest
import java.lang.Thread.sleep
import javax.inject.Inject
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

@HiltAndroidTest
@RunWith(AndroidJUnit4::class)
class SafetyComplianceTest {

    @get:Rule(order = 0) var hiltRule = HiltAndroidRule(this)

    @get:Rule(order = 1) val composeTestRule = createAndroidComposeRule<MainActivity>()

    @Inject lateinit var telemetryService: FakeTelemetryService

    @Before
    fun setup() {
        hiltRule.inject()
    }

    @Test
    fun testMandatorySpeedLockout_Layer5() {
        // 1. Driving starts (Simulate via Layer 4 Field Simulation)
        telemetryService.simulateSpeed(25.0)
        sleep(2000)

        // 2. Verify Layer 5 Compliance: UI MUST be locked
        composeTestRule.onNodeWithText("SCREEN LOCKED", substring = true).assertIsDisplayed()

        // 3. Verify interaction is blocked
        composeTestRule.onAllNodesWithText("CHANGE STATUS").assertCountEquals(0)
    }

    @Test
    fun testHardwareFailureResponse_Layer4() {
        // 1. Simulate OBD Disconnect
        telemetryService.simulateObd(false)
        sleep(2000)

        // 2. Verify degraded mode banner/screen
        composeTestRule.onNodeWithText("ObdDisconnected", substring = true).assertIsDisplayed()
    }

    @Test
    fun testTamperDetectionResponse_Layer5() {
        // 1. Simulate Tamper Event (High Severity)
        telemetryService.simulateTamper(true)
        sleep(2000)

        // 2. Verify mandatory device lock
        composeTestRule.onNodeWithText("TamperDetected", substring = true).assertIsDisplayed()
    }
}
