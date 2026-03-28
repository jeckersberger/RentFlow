package de.cratedesk.scanner.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import de.cratedesk.scanner.ui.scan.ScanScreen
import de.cratedesk.scanner.ui.scan.ScanViewModel
import de.cratedesk.scanner.ui.settings.LoginScreen
import de.cratedesk.scanner.ui.settings.LoginViewModel
import de.cratedesk.scanner.ui.settings.SettingsScreen

@Composable
fun ScannerNavHost() {
    val navController = rememberNavController()
    val loginViewModel: LoginViewModel = hiltViewModel()
    val isLoggedIn by loginViewModel.isLoggedIn.collectAsState(initial = false)

    val startDestination = if (isLoggedIn) "scan" else "login"

    NavHost(navController = navController, startDestination = startDestination) {
        composable("login") {
            LoginScreen(
                viewModel = loginViewModel,
                onLoginSuccess = {
                    navController.navigate("scan") {
                        popUpTo("login") { inclusive = true }
                    }
                },
            )
        }
        composable("scan") {
            val scanViewModel: ScanViewModel = hiltViewModel()
            ScanScreen(
                viewModel = scanViewModel,
                onSettingsClick = { navController.navigate("settings") },
            )
        }
        composable("settings") {
            SettingsScreen(
                onBack = { navController.popBackStack() },
                onLogout = {
                    loginViewModel.logout()
                    navController.navigate("login") {
                        popUpTo(0) { inclusive = true }
                    }
                },
            )
        }
    }
}
