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
class PlatformEdgeCasesUITest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun verifyDriverRoleSwapFlow() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.RoleSwap.route)
            }
        }

        // Verify Swap Screen UI
        composeTestRule.onNodeWithText("DRIVER ROLE SWAP").assertExists()
        composeTestRule.onNodeWithText("Current Driver: John Doe").assertExists()
        composeTestRule.onNodeWithText("CONFIRM SWAP").performClick()
        
        // Navigation results in popBackStack usually, but in mock E2E it works
    }

    @Test
    fun verifyCoSignatureFlow() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.CoSignature.route)
            }
        }

        composeTestRule.onNodeWithText("CO-DRIVER SIGNATURE").assertExists()
        composeTestRule.onNodeWithText("John Doe").assertExists()
        composeTestRule.onNodeWithText("AUTHORIZE").performClick()
    }

    @Test
    fun verifyPrivacyToggleUI() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.SecurityProtocols.route)
            }
        }

        composeTestRule.onNodeWithText("Security Protocols").assertExists()
        composeTestRule.onNodeWithText("Privacy Mode", substring = true).assertExists()
        // Toggle interaction check
        composeTestRule.onNode(hasClickAction()).performClick()
    }

    @Test
    fun verifyHardwareDiagnosticDetails() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.HardwareStatus.route)
            }
        }

        composeTestRule.onNodeWithText("Hardware Peripherals").assertExists()
        composeTestRule.onNodeWithText("CAMERA").assertExists()
        composeTestRule.onNodeWithText("CONNECTED", substring = true).assertExists()
    }

    @Test
    fun verifyEcmDiagnosticDetails() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.EcmDetails.route)
            }
        }

        composeTestRule.onNodeWithText("ECM Diagnostics").assertExists()
        composeTestRule.onNodeWithText("CAN BUS", substring = true).assertExists()
        composeTestRule.onNodeWithText("J1939 Primary", substring = true).assertExists()
    }
}
