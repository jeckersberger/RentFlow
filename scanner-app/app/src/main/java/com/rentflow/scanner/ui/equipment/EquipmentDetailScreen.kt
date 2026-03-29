package com.rentflow.scanner.ui.equipment

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.rentflow.scanner.domain.model.ScanType
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.R
import com.rentflow.scanner.ui.components.EquipmentCard
import com.rentflow.scanner.ui.components.LoadingScreen
import com.rentflow.scanner.ui.theme.Cyan
import com.rentflow.scanner.ui.theme.Success
import com.rentflow.scanner.ui.theme.Warning

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EquipmentDetailScreen(
    onBack: () -> Unit,
    viewModel: EquipmentDetailViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()
    var showLocationDialog by remember { mutableStateOf(false) }

    // Location picker dialog
    if (showLocationDialog && state.zones.isNotEmpty()) {
        val currentLocation = state.equipment?.location
        AlertDialog(
            onDismissRequest = { showLocationDialog = false },
            title = { Text(stringResource(R.string.equipment_select_location)) },
            text = {
                LazyColumn {
                    items(state.zones) { zone ->
                        val isSelected = zone.name == currentLocation
                        ListItem(
                            headlineContent = {
                                Text(
                                    zone.name,
                                    fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Normal,
                                )
                            },
                            trailingContent = {
                                if (isSelected) {
                                    Icon(Icons.Default.Check, contentDescription = null, tint = Cyan)
                                }
                            },
                            modifier = Modifier.clickable {
                                viewModel.updateLocation(zone.name)
                                showLocationDialog = false
                            },
                        )
                    }
                }
            },
            confirmButton = {},
            dismissButton = {
                OutlinedButton(onClick = { showLocationDialog = false }) {
                    Text(stringResource(R.string.cancel))
                }
            },
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.equipment_detail_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
            )
        },
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp).verticalScroll(rememberScrollState())) {
            if (state.isLoading) {
                LoadingScreen()
            } else if (state.equipment != null) {
                EquipmentCard(state.equipment!!)

                Spacer(Modifier.height(24.dp))

                // Location change
                Button(
                    onClick = { showLocationDialog = true },
                    modifier = Modifier.fillMaxWidth().height(56.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = Cyan),
                ) {
                    Icon(Icons.Default.EditLocationAlt, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
                    Text(stringResource(R.string.equipment_change_location))
                }

                state.locationSaved?.let {
                    Spacer(Modifier.height(8.dp))
                    Text(it, color = Success)
                }

                Spacer(Modifier.height(16.dp))

                // RFID section
                state.equipment!!.rfidTag?.let {
                    Text(stringResource(R.string.equipment_rfid_label, it), style = MaterialTheme.typography.bodyLarge)
                    Spacer(Modifier.height(16.dp))
                }

                Button(
                    onClick = viewModel::pairRfidTag,
                    modifier = Modifier.fillMaxWidth().height(56.dp),
                    colors = if (state.isWritingRfid) {
                        ButtonDefaults.buttonColors(containerColor = Warning)
                    } else {
                        ButtonDefaults.buttonColors()
                    },
                ) {
                    if (state.isWritingRfid) {
                        CircularProgressIndicator(modifier = Modifier.size(24.dp), color = MaterialTheme.colorScheme.onPrimary)
                        Spacer(Modifier.width(8.dp))
                        Text(stringResource(R.string.rfid_writing))
                    } else {
                        Icon(Icons.Default.Nfc, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
                        Text(stringResource(R.string.equipment_write_rfid))
                    }
                }

                state.rfidWriteResult?.let { result ->
                    Spacer(Modifier.height(16.dp))
                    if (result.success) {
                        Card(
                            colors = CardDefaults.cardColors(containerColor = Success.copy(alpha = 0.15f)),
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                                Icon(
                                    if (state.rfidVerified) Icons.Default.VerifiedUser else Icons.Default.Check,
                                    contentDescription = null,
                                    tint = Success,
                                )
                                Spacer(Modifier.width(8.dp))
                                Column {
                                    Text(stringResource(R.string.rfid_write_success), color = Success, fontWeight = FontWeight.Bold)
                                    if (state.rfidVerified) {
                                        Text(stringResource(R.string.rfid_write_verified), style = MaterialTheme.typography.bodySmall, color = Success)
                                    } else {
                                        Text(stringResource(R.string.rfid_write_not_verified), style = MaterialTheme.typography.bodySmall, color = Warning)
                                    }
                                }
                            }
                        }
                    } else {
                        Card(
                            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer),
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                                Icon(Icons.Default.Error, contentDescription = null, tint = MaterialTheme.colorScheme.error)
                                Spacer(Modifier.width(8.dp))
                                Column {
                                    Text(
                                        result.error ?: stringResource(R.string.rfid_write_failed),
                                        color = MaterialTheme.colorScheme.error,
                                        fontWeight = FontWeight.Bold,
                                    )
                                    Spacer(Modifier.height(4.dp))
                                    OutlinedButton(onClick = viewModel::pairRfidTag) {
                                        Icon(Icons.Default.Refresh, contentDescription = null, modifier = Modifier.size(16.dp))
                                        Spacer(Modifier.width(4.dp))
                                        Text(stringResource(R.string.retry))
                                    }
                                }
                            }
                        }
                    }
                }

                // RSSI Equipment Locator
                Spacer(Modifier.height(24.dp))
                if (state.isLocating) {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant),
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(
                                stringResource(R.string.equipment_locating),
                                style = MaterialTheme.typography.titleSmall,
                                fontWeight = FontWeight.Bold,
                            )
                            Spacer(Modifier.height(8.dp))
                            Text(
                                stringResource(R.string.equipment_signal_strength),
                                style = MaterialTheme.typography.bodySmall,
                            )
                            Spacer(Modifier.height(4.dp))
                            // RSSI bar: -70 dBm = weak (0%), -30 dBm = strong (100%)
                            val rssiNormalized = ((state.rssiValue + 70).toFloat() / 40f).coerceIn(0f, 1f)
                            val animatedProgress by animateFloatAsState(targetValue = rssiNormalized, label = "rssi")
                            LinearProgressIndicator(
                                progress = { animatedProgress },
                                modifier = Modifier.fillMaxWidth().height(12.dp),
                                color = when {
                                    rssiNormalized > 0.66f -> Success
                                    rssiNormalized > 0.33f -> Warning
                                    else -> MaterialTheme.colorScheme.error
                                },
                                trackColor = MaterialTheme.colorScheme.surfaceVariant,
                            )
                            Spacer(Modifier.height(4.dp))
                            Text(
                                "${state.rssiValue} dBm",
                                style = MaterialTheme.typography.bodySmall,
                                fontFamily = FontFamily.Monospace,
                            )
                            Spacer(Modifier.height(12.dp))
                            Button(
                                onClick = viewModel::stopLocating,
                                modifier = Modifier.fillMaxWidth(),
                                colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error),
                            ) {
                                Icon(Icons.Default.Stop, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
                                Text(stringResource(R.string.cancel))
                            }
                        }
                    }
                } else {
                    Button(
                        onClick = viewModel::startLocating,
                        modifier = Modifier.fillMaxWidth().height(56.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = Cyan),
                    ) {
                        Icon(Icons.Default.TrackChanges, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
                        Text(stringResource(R.string.equipment_locate))
                    }
                }

                // Scan History
                Spacer(Modifier.height(24.dp))
                Text(
                    stringResource(R.string.equipment_history),
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                )
                Spacer(Modifier.height(8.dp))
                if (state.history.isEmpty()) {
                    Text(
                        stringResource(R.string.equipment_no_history),
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                } else {
                    val dateFormat = remember { SimpleDateFormat("dd.MM.yyyy HH:mm", Locale.getDefault()) }
                    state.history.forEach { entry ->
                        Card(
                            modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant),
                        ) {
                            Row(
                                modifier = Modifier.padding(12.dp),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Icon(
                                    when (entry.scanType) {
                                        ScanType.OUT -> Icons.Default.Output
                                        ScanType.IN -> Icons.Default.Login
                                        ScanType.INVENTORY -> Icons.Default.Inventory
                                        ScanType.LOOKUP -> Icons.Default.Search
                                    },
                                    contentDescription = null,
                                    tint = Cyan,
                                    modifier = Modifier.size(20.dp),
                                )
                                Spacer(Modifier.width(12.dp))
                                Column(modifier = Modifier.weight(1f)) {
                                    Text(
                                        entry.scanType.name,
                                        style = MaterialTheme.typography.bodyMedium,
                                        fontWeight = FontWeight.Bold,
                                    )
                                    Text(
                                        dateFormat.format(Date(entry.timestamp)),
                                        style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    )
                                }
                            }
                        }
                    }
                }
            } else {
                state.error?.let { Text(it, color = MaterialTheme.colorScheme.error) }
            }
        }
    }
}
