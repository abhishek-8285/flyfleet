package com.flyfleet.vehicledevice.ui.navigation

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.navigation.compose.rememberNavController
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import androidx.test.ext.junit.runners.AndroidJUnit4

@RunWith(AndroidJUnit4::class)
class DiagnosticsWorkflowTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun verifyDiagnosticsNavigationFlow() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.SystemHealth.route)
            }
        }

        // 1. Verify System Health Dashboard
        composeTestRule.onNodeWithText("System Health").assertExists()
        composeTestRule.onNodeWithText("SUBSYSTEM STATUS").assertExists()

        // 2. Navigate to Network Status
        composeTestRule.onNodeWithText("Cellular Connectivity", substring = true).performClick()
        composeTestRule.onNodeWithText("Connectivity").assertExists()
        composeTestRule.onNodeWithText("5G ACTIVE", substring = true).assertExists()
        
        // Go Back
        composeTestRule.onNodeWithContentDescription("ChevronLeft", substring = true).performClick()
        
        // 3. Navigate to ECM Details (Verify path exists)
        composeTestRule.onNodeWithText("ECM / J1939", substring = true).performClick()
        composeTestRule.onNodeWithText("ECM Diagnostics").assertExists()
        
        // Go Back
        composeTestRule.onNodeWithContentDescription("ChevronLeft", substring = true).performClick()

        // 4. Navigate to Storage Monitor
        composeTestRule.onNodeWithText("Storage & Memory", substring = true).performClick()
        composeTestRule.onNodeWithText("Storage Integrity").assertExists()
    }

    @Test
    fun verifyCriticalMalfunctionVisibility() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.MalfunctionHud.route)
            }
        }

        // Verify that the Malfunction HUD can be reached and has critical warning text
        composeTestRule.onNodeWithText("CRITICAL MALFUNCTION", substring = true).assertExists()
        composeTestRule.onNodeWithText("DO NOT DRIVE", substring = true).assertExists()
    }
}
