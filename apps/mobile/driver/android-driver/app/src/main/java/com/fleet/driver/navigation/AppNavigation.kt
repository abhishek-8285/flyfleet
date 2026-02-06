package com.fleet.driver.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.fleet.driver.presentation.auth.screen.EnrollmentScreen
import com.fleet.driver.presentation.auth.screen.LoginScreen
import com.fleet.driver.presentation.main.screen.MainScreen
import com.fleet.driver.presentation.vehicle.screen.VehicleListScreen

/**
 * App Navigation Graph
 */
@Composable
fun AppNavigation(
    navController: NavHostController = rememberNavController()
) {
    NavHost(
        navController = navController,
        startDestination = NavigationScreens.Login.route
    ) {
        // Authentication Screens
        composable(NavigationScreens.Login.route) {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate(NavigationScreens.Main.route) {
                        popUpTo(NavigationScreens.Login.route) { inclusive = true }
                    }
                },
                onNavigateToEnrollment = {
                    navController.navigate(NavigationScreens.Enrollment.route)
                }
            )
        }
        
        composable(NavigationScreens.Enrollment.route) {
            EnrollmentScreen(
                onEnrollmentSuccess = {
                    navController.navigate(NavigationScreens.Main.route) {
                        popUpTo(NavigationScreens.Enrollment.route) { inclusive = true }
                    }
                },
                onNavigateToLogin = {
                    navController.popBackStack()
                }
            )
        }
        
        // Main App Screens
        composable(NavigationScreens.Main.route) {
            MainScreen(
                onLogout = {
                    navController.navigate(NavigationScreens.Login.route) {
                        popUpTo(NavigationScreens.Main.route) { inclusive = true }
                    }
                }
            )
        }
    }
}

/**
 * Navigation Screens Definition
 */
sealed class NavigationScreens(val route: String) {
    object Login : NavigationScreens("login")
    object Enrollment : NavigationScreens("enrollment")
    object Main : NavigationScreens("main")
}