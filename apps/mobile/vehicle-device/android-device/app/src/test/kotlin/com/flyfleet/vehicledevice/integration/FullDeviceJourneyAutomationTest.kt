package com.flyfleet.vehicledevice.integration

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.flyfleet.vehicledevice.ui.MainActivity
import com.flyfleet.vehicledevice.ui.navigation.AppNavigation
import com.flyfleet.vehicledevice.ui.navigation.Screen
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme
import dagger.hilt.android.testing.HiltAndroidRule
import dagger.hilt.android.testing.HiltAndroidTest
import dagger.hilt.android.testing.HiltTestApplication
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.shadows.ShadowLog
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.setMain
import kotlinx.coroutines.test.resetMain
import org.junit.After
import org.junit.Before
import timber.log.Timber

@HiltAndroidTest
@Config(application = HiltTestApplication::class, sdk = [33], qualifiers = "w1280dp-h800dp-xhdpi")
@RunWith(RobolectricTestRunner::class)
class FullDeviceJourneyAutomationTest {

    @get:Rule(order = 0)
    val hiltRule = HiltAndroidRule(this)

    @get:Rule(order = 1)
    val composeTestRule = createAndroidComposeRule<MainActivity>()

    @OptIn(ExperimentalCoroutinesApi::class)
    private val testDispatcher = UnconfinedTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        ShadowLog.stream = System.out
        if (Timber.treeCount == 0) {
            Timber.plant(Timber.DebugTree())
        }
        resetMockServer()
        hiltRule.inject()
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun resetMockServer() {
        try {
            val url = java.net.URL("http://localhost:3000/mock/scenario/reset")
            (url.openConnection() as java.net.HttpURLConnection).apply {
                requestMethod = "POST"
                setRequestProperty("Content-Type", "application/json")
                connect()
                println(">>> MOCK SERVER RESET: HTTP $responseCode")
                disconnect()
            }
        } catch (e: Exception) {
            println(">>> MOCK SERVER RESET FAILED: ${e.message}")
        }
    }

    @Test
    fun executeFullRegulatoryJourney_onDevice() {
        println(">>> STARTING FullDeviceJourneyAutomationTest")
        // Activity is automatically launched by createAndroidComposeRule<MainActivity>()
        println(">>> ACTIVITY LAUNCHED")
            // Wait for system integrity checks to complete
            println(">>> STEP 1: SECURE BOOT - Waiting for button")
        // Wait for system integrity checks to complete
        composeTestRule.waitUntil(20000) {
            composeTestRule.onAllNodesWithTag("launch_dashboard_button").fetchSemanticsNodes().isNotEmpty()
        }
        println(">>> STEP 1: BUTTON FOUND - Clicking")
        composeTestRule.onNodeWithTag("launch_dashboard_button").assertIsEnabled().performClick()

        // --- STEP 2: DRIVER LOGIN (Screens 7-18) ---
        println(">>> STEP 2: DRIVER LOGIN - Waiting for 'Tap your card to login'")
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("Tap your card to login").fetchSemanticsNodes().isNotEmpty()
        }
        println(">>> STEP 2: 'Tap your card to login' found - Switching to PIN mode")
        
        // Switch to PIN mode
        composeTestRule.onNodeWithTag("login_using_pin_button").performClick()
        composeTestRule.waitForIdle()
        
        // Select Driver
        println(">>> STEP 2: Waiting for driver DRV-001")
        try {
            composeTestRule.waitUntil(30000) {
                composeTestRule.onAllNodesWithTag("driver_DRV-001").fetchSemanticsNodes().isNotEmpty()
            }
        } catch (e: Exception) {
            println(">>> DRIVER SELECTION TIMEOUT: Printing UI Tree")
            composeTestRule.onRoot().printToLog("UI_TREE")
            throw e
        }
        composeTestRule.onNodeWithTag("driver_DRV-001").performClick()
        
        // Enter PIN: 123456
        try {
            composeTestRule.waitUntil(20000) {
                composeTestRule.onAllNodesWithTag("numeric_keypad").fetchSemanticsNodes().isNotEmpty()
            }
        } catch (e: Exception) {
            println(">>> PIN ENTRY TIMEOUT: Printing UI Tree")
            composeTestRule.onRoot().printToLog("UI_TREE")
            throw e
        }
        composeTestRule.onNodeWithTag("numeric_keypad").assertIsDisplayed()
        val keys = listOf(1, 2, 3, 4, 5, 6)
        keys.forEach { digit ->
            composeTestRule.onNodeWithTag("key_$digit").performClick()
        }

        // --- STEP 2: LOGIN SUCCESS & NAVIGATION ---
        android.util.Log.e("DEBUG_NAV", "TEST: PIN entered, waiting for dashboard...")

        // --- STEP 3: DASHBOARD & HOS (Screens 19-30) ---
        // Wait for navigation to Dashboard
        try {
            android.util.Log.e("DEBUG_NAV", "TEST: Waiting for Dashboard...")
            composeTestRule.waitUntil(30000) {
                val nodes = composeTestRule.onAllNodesWithText("CHANGE STATUS").fetchSemanticsNodes()
                if (nodes.isNotEmpty()) {
                    android.util.Log.e("DEBUG_NAV", "TEST: DASHBOARD DETECTED!")
                    true
                } else {
                    false
                }
            }
        } catch (e: Exception) {
            android.util.Log.e("DEBUG_NAV", "TEST: DASHBOARD TIMEOUT: Printing UI Tree")
            composeTestRule.onRoot().printToLog("UI_TREE")
            throw e
        }
        android.util.Log.e("DEBUG_NAV", "TEST: DASHBOARD LOADED")
        
        // Navigate to HOS Clocks
        composeTestRule.onNodeWithText("HOS LOGS").performClick()
        
        composeTestRule.waitUntil(15000) {
            composeTestRule.onAllNodesWithText("HOS Compliance").fetchSemanticsNodes().isNotEmpty()
        }

        // Navigate to HOS Logs Summary via bottom nav
        composeTestRule.onNodeWithText("LOGS", ignoreCase = true).performClick()

        composeTestRule.waitUntil(20000) {
            composeTestRule.onAllNodesWithText("HOS Daily Log Summary", ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        // Select a log entry - click the date to avoid the inner certification clickable
        try {
            composeTestRule.onAllNodesWithText("Monday, June 17", substring = true, ignoreCase = true)
                .onFirst()
                .performClick()
        } catch (e: Exception) {
            println(">>> LOG SELECTION FAILED: Printing UI Tree")
            composeTestRule.onRoot().printToLog("UI_TREE")
            throw e
        }
        
        // This lands on HOSGraphDetailScreen, but HOSDailyLogSummaryScreen used it too?
        // Wait, AppNavigation says HOSGraphDetailScreen(date=...)
        // Let's just wait for "LOG SEGMENTS" which is in those detail screens
        try {
            composeTestRule.waitUntil(30000) {
                composeTestRule.onAllNodesWithText("LOG SEGMENTS", substring = true, ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
            }
        } catch (e: Exception) {
            println(">>> GRAPH DETAIL TIMEOUT: Printing UI Tree")
            composeTestRule.onRoot().printToLog("UI_TREE")
            throw e
        }
        
        // --- STEP 4: ROADSIDE INSPECTION (Screens 31-40) ---
        // Go back to Dashboard
        composeTestRule.onNodeWithContentDescription("Back").performClick() // to Summary/History
        composeTestRule.onNodeWithContentDescription("Back").performClick() // to Clocks
        composeTestRule.onNodeWithContentDescription("Back").performClick() // to Dashboard
        
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("INSPECT", ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        // Launch Inspection Mode
        composeTestRule.onNodeWithText("INSPECT", ignoreCase = true).performClick()
        
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("Roadside Inspection", substring = true, ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        // --- STEP 5: DIAGNOSTICS & MALFUNCTIONS (Screens 41-50) ---
        composeTestRule.onNodeWithText("Exit Mode").performClick()
        
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("DIAGS", ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        composeTestRule.onNodeWithText("DIAGS", ignoreCase = true).performClick()
        
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("System Diagnostics", ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        // View Subsystem Status (Telemetry section)
        composeTestRule.onNodeWithText("DEVICE TELEMETRY", substring = true, ignoreCase = true).assertIsDisplayed()
        
        // Go back to Dashboard
        composeTestRule.onNodeWithContentDescription("Back").performClick()

        // --- STEP 6: EMERGENCY & PANIC (Screens 51-57) ---
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("PANIC", ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        composeTestRule.onNodeWithTag("panic_button").performScrollTo().performTouchInput { longClick() }
        
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("PANIC DISPATCHED", substring = true, ignoreCase = true).fetchSemanticsNodes().isNotEmpty()
        }

        // --- STEP 7: LOGOUT (Screen 57) ---
        // Wait for panic acknowledgment or just logout if possible
        // For automated test, we might need to wait for the overlay to disappear or navigate via back
        // composeTestRule.onNodeWithText("DONE").performClick() 
    }
}
