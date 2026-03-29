package com.rentflow.scanner.ui.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.rentflow.scanner.ui.checkin.CheckInScreen
import com.rentflow.scanner.ui.checkout.CheckOutScreen
import com.rentflow.scanner.ui.equipment.EquipmentDetailScreen
import com.rentflow.scanner.ui.home.HomeScreen
import com.rentflow.scanner.ui.inventory.InventoryScreen
import com.rentflow.scanner.ui.login.LoginScreen
import com.rentflow.scanner.ui.queue.PendingQueueScreen
import com.rentflow.scanner.ui.createproject.CreateProjectScreen
import com.rentflow.scanner.ui.rfidassign.RfidAssignScreen
import com.rentflow.scanner.ui.scan.ScanScreen
import com.rentflow.scanner.ui.settings.SettingsScreen

object Routes {
    const val LOGIN = "login"
    const val HOME = "home"
    const val SCAN = "scan"
    const val CHECKOUT = "checkout"
    const val CHECKOUT_PROJECT = "checkout/{projectId}"
    const val CHECKIN = "checkin"
    const val INVENTORY = "inventory"
    const val EQUIPMENT_DETAIL = "equipment/{barcode}"
    const val RFID_ASSIGN = "rfid_assign"
    const val CREATE_PROJECT = "create_project"
    const val SETTINGS = "settings"
    const val QUEUE = "queue"

    fun equipmentDetail(barcode: String) = "equipment/$barcode"
    fun checkoutProject(projectId: String) = "checkout/$projectId"
}

@Composable
fun AppNavigation(startDestination: String = Routes.LOGIN) {
    val navController = rememberNavController()

    NavHost(navController = navController, startDestination = startDestination) {
        composable(Routes.LOGIN) {
            LoginScreen(onLoginSuccess = {
                navController.navigate(Routes.HOME) {
                    popUpTo(Routes.LOGIN) { inclusive = true }
                }
            })
        }
        composable(Routes.HOME) {
            HomeScreen(
                onNavigateToScan = { navController.navigate(Routes.SCAN) },
                onNavigateToCheckOut = { navController.navigate(Routes.CHECKOUT) },
                onNavigateToCheckIn = { navController.navigate(Routes.CHECKIN) },
                onNavigateToInventory = { navController.navigate(Routes.INVENTORY) },
                onNavigateToRfidAssign = { navController.navigate(Routes.RFID_ASSIGN) },
                onNavigateToCreateProject = { navController.navigate(Routes.CREATE_PROJECT) },
                onNavigateToSettings = { navController.navigate(Routes.SETTINGS) },
                onNavigateToQueue = { navController.navigate(Routes.QUEUE) },
                onLogout = {
                    navController.navigate(Routes.LOGIN) {
                        popUpTo(0) { inclusive = true }
                    }
                },
            )
        }
        composable(Routes.SCAN) {
            ScanScreen(
                onBack = { navController.popBackStack() },
                onCheckOut = { navController.navigate(Routes.CHECKOUT) },
                onCheckIn = { navController.navigate(Routes.CHECKIN) },
                onEquipmentDetail = { barcode -> navController.navigate(Routes.equipmentDetail(barcode)) },
                onNavigateToJob = { projectId -> navController.navigate(Routes.checkoutProject(projectId)) },
            )
        }
        composable(Routes.CHECKOUT) {
            CheckOutScreen(
                onBack = { navController.popBackStack() },
                onCompleted = {
                    navController.navigate(Routes.HOME) {
                        popUpTo(Routes.HOME) { inclusive = true }
                    }
                },
            )
        }
        composable(Routes.CHECKOUT_PROJECT) {
            CheckOutScreen(
                onBack = { navController.popBackStack() },
                onCompleted = {
                    navController.navigate(Routes.HOME) {
                        popUpTo(Routes.HOME) { inclusive = true }
                    }
                },
            )
        }
        composable(Routes.CHECKIN) {
            CheckInScreen(
                onBack = { navController.popBackStack() },
                onCompleted = {
                    navController.navigate(Routes.HOME) {
                        popUpTo(Routes.HOME) { inclusive = true }
                    }
                },
            )
        }
        composable(Routes.INVENTORY) {
            InventoryScreen(
                onBack = { navController.popBackStack() },
                onCompleted = {
                    navController.navigate(Routes.HOME) {
                        popUpTo(Routes.HOME) { inclusive = true }
                    }
                },
            )
        }
        composable(Routes.RFID_ASSIGN) {
            RfidAssignScreen(onBack = { navController.popBackStack() })
        }
        composable(Routes.CREATE_PROJECT) {
            CreateProjectScreen(onBack = { navController.popBackStack() })
        }
        composable(Routes.EQUIPMENT_DETAIL) {
            EquipmentDetailScreen(onBack = { navController.popBackStack() })
        }
        composable(Routes.SETTINGS) {
            SettingsScreen(onBack = { navController.popBackStack() })
        }
        composable(Routes.QUEUE) {
            PendingQueueScreen(onBack = { navController.popBackStack() })
        }
    }
}
