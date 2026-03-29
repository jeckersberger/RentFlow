package com.rentflow.scanner.ui.home

import android.widget.Toast
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.R
import com.rentflow.scanner.ui.components.UpdateBanner

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    onNavigateToScan: () -> Unit,
    onNavigateToCheckOut: () -> Unit,
    onNavigateToCheckIn: () -> Unit,
    onNavigateToInventory: () -> Unit,
    onNavigateToRfidAssign: () -> Unit,
    onNavigateToCreateProject: () -> Unit,
    onNavigateToSettings: () -> Unit,
    onNavigateToQueue: () -> Unit,
    onLogout: () -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()
    val context = LocalContext.current

    val comingSoonAction: () -> Unit = {
        Toast.makeText(context, "Coming soon", Toast.LENGTH_SHORT).show()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.home_title)) },
                actions = {
                    IconButton(onClick = onNavigateToQueue) {
                        BadgedBox(
                            badge = {
                                if (state.pendingScanCount > 0) {
                                    Badge(containerColor = MaterialTheme.colorScheme.error) {
                                        Text("${state.pendingScanCount}")
                                    }
                                }
                            }
                        ) {
                            Icon(Icons.Default.CloudQueue, contentDescription = "Offline Queue")
                        }
                    }
                    IconButton(onClick = onNavigateToSettings) {
                        Icon(Icons.Default.Settings, contentDescription = stringResource(R.string.nav_settings))
                    }
                    IconButton(onClick = {
                        viewModel.logout()
                        onLogout()
                    }) {
                        Icon(Icons.AutoMirrored.Filled.Logout, contentDescription = "Logout")
                    }
                },
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            state.updateInfo?.let { info ->
                UpdateBanner(
                    updateInfo = info,
                    onUpdate = { viewModel.downloadUpdate() },
                    modifier = Modifier.padding(bottom = 8.dp),
                )
            }

            if (state.userName.isNotBlank()) {
                Text(stringResource(R.string.hello_user, state.userName), style = MaterialTheme.typography.headlineMedium)
            }

            for (cardId in state.homeCards) {
                val cardDef = getCardDefinition(cardId, state.industryLabels)
                if (cardDef != null) {
                    val onClick = when (cardId) {
                        "scan_info" -> onNavigateToScan
                        "checkout" -> onNavigateToCheckOut
                        "checkin" -> onNavigateToCheckIn
                        "inventory" -> onNavigateToInventory
                        "rfid_assign" -> onNavigateToRfidAssign
                        "new_project" -> onNavigateToCreateProject
                        else -> comingSoonAction
                    }
                    WorkflowCard(cardDef.icon, cardDef.label, onClick = onClick)
                }
            }
        }
    }
}

private data class CardDefinition(
    val icon: ImageVector,
    val label: String,
)

@Composable
private fun getCardDefinition(cardId: String, labels: Map<String, String>): CardDefinition? {
    return when (cardId) {
        "scan_info" -> CardDefinition(
            Icons.Default.QrCodeScanner,
            labels["scan_info"] ?: stringResource(R.string.nav_scan),
        )
        "checkout" -> CardDefinition(
            Icons.Default.Output,
            labels["checkout"] ?: stringResource(R.string.nav_checkout),
        )
        "checkin" -> CardDefinition(
            Icons.Default.Input,
            labels["checkin"] ?: stringResource(R.string.nav_checkin),
        )
        "inventory" -> CardDefinition(
            Icons.Default.Inventory,
            labels["inventory"] ?: stringResource(R.string.nav_inventory),
        )
        "rfid_assign" -> CardDefinition(
            Icons.Default.Nfc,
            labels["rfid_assign"] ?: stringResource(R.string.rfid_assign_title),
        )
        "new_project" -> CardDefinition(
            Icons.Default.CreateNewFolder,
            labels["new_project"] ?: stringResource(R.string.create_project_title),
        )
        "operating_hours" -> CardDefinition(
            Icons.Default.Timer,
            labels["operating_hours"] ?: stringResource(R.string.card_operating_hours),
        )
        "fuel_entry" -> CardDefinition(
            Icons.Default.LocalGasStation,
            labels["fuel_entry"] ?: stringResource(R.string.card_fuel_entry),
        )
        "damage_report" -> CardDefinition(
            Icons.Default.ReportProblem,
            labels["damage_report"] ?: stringResource(R.string.card_damage_report),
        )
        "cleaning_confirm" -> CardDefinition(
            Icons.Default.CleaningServices,
            labels["cleaning_confirm"] ?: stringResource(R.string.card_cleaning_confirm),
        )
        "mileage_entry" -> CardDefinition(
            Icons.Default.Speed,
            labels["mileage_entry"] ?: stringResource(R.string.card_mileage_entry),
        )
        else -> null
    }
}

@Composable
private fun WorkflowCard(icon: ImageVector, label: String, onClick: () -> Unit, iconTint: Color? = null) {
    Card(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .height(80.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 24.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            Icon(icon, contentDescription = null, modifier = Modifier.size(32.dp), tint = iconTint ?: LocalContentColor.current)
            Text(label, style = MaterialTheme.typography.titleLarge)
        }
    }
}
