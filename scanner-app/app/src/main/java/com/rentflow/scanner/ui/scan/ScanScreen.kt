package com.rentflow.scanner.ui.scan

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.R
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.domain.model.EquipmentStatus
import com.rentflow.scanner.ui.components.EquipmentCard
import com.rentflow.scanner.ui.theme.Cyan

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ScanScreen(
    onBack: () -> Unit,
    onCheckOut: () -> Unit,
    onCheckIn: () -> Unit,
    onEquipmentDetail: (String) -> Unit,
    onNavigateToJob: ((String) -> Unit)? = null,
    viewModel: ScanViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()
    val isRfidMode = state.scanMode == SettingsDataStore.SCAN_MODE_RFID

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.nav_scan)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
                actions = {
                    // Show current scan mode indicator
                    Icon(
                        if (isRfidMode) Icons.Default.Sensors else Icons.Default.QrCodeScanner,
                        contentDescription = null,
                        tint = Cyan,
                        modifier = Modifier.padding(end = 16.dp),
                    )
                },
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            if (state.isLoading) {
                CircularProgressIndicator(modifier = Modifier.padding(32.dp))
            } else if (state.equipment != null) {
                // Barcode scan result
                val eq = state.equipment!!
                EquipmentCard(eq, onClick = { onEquipmentDetail(eq.barcode) })

                if (eq.status == EquipmentStatus.CHECKED_OUT && eq.projectName != null) {
                    Spacer(Modifier.height(12.dp))
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.primaryContainer,
                        ),
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth().padding(16.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Column(modifier = Modifier.weight(1f)) {
                                Text(
                                    stringResource(R.string.scan_on_job, eq.projectName),
                                    style = MaterialTheme.typography.bodyLarge,
                                    fontWeight = FontWeight.Bold,
                                    color = MaterialTheme.colorScheme.onPrimaryContainer,
                                )
                            }
                            if (eq.projectId != null && onNavigateToJob != null) {
                                FilledTonalButton(onClick = { onNavigateToJob(eq.projectId) }) {
                                    Text(stringResource(R.string.scan_go_to_job))
                                    Spacer(Modifier.width(4.dp))
                                    Icon(Icons.AutoMirrored.Filled.ArrowForward, contentDescription = null, modifier = Modifier.size(18.dp))
                                }
                            }
                        }
                    }
                }

                Spacer(Modifier.height(24.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                    Button(onClick = onCheckOut) {
                        Icon(Icons.Default.Output, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
                        Text(stringResource(R.string.nav_checkout))
                    }
                    Button(onClick = onCheckIn) {
                        Icon(Icons.Default.Input, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
                        Text(stringResource(R.string.nav_checkin))
                    }
                }
                Spacer(Modifier.height(16.dp))
                OutlinedButton(onClick = { viewModel.clearResult() }) {
                    Text(stringResource(R.string.scan_ready))
                }
            } else {
                // Waiting for scan
                Spacer(Modifier.height(48.dp))
                Icon(
                    if (isRfidMode) Icons.Default.Sensors else Icons.Default.QrCodeScanner,
                    contentDescription = null,
                    modifier = Modifier.size(96.dp),
                    tint = if (state.isScanning) Cyan else Cyan.copy(alpha = 0.5f),
                )
                Spacer(Modifier.height(16.dp))

                Text(
                    if (isRfidMode) {
                        if (state.isScanning) stringResource(R.string.scan_rfid_reading) else stringResource(R.string.scan_rfid_trigger)
                    } else {
                        stringResource(R.string.scan_barcode_trigger)
                    },
                    style = MaterialTheme.typography.headlineSmall,
                    color = if (state.isScanning) Cyan else MaterialTheme.colorScheme.onSurface,
                )

                Text(
                    if (isRfidMode) stringResource(R.string.scan_mode_rfid) else stringResource(R.string.scan_mode_barcode),
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )

                state.error?.let {
                    Spacer(Modifier.height(16.dp))
                    Text(it, color = MaterialTheme.colorScheme.error)
                    if (state.pendingRfidTag != null) {
                        Spacer(Modifier.height(12.dp))
                        Button(
                            onClick = { viewModel.startAssignFlow() },
                            modifier = Modifier.fillMaxWidth().padding(horizontal = 32.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = Cyan),
                        ) {
                            Icon(Icons.Default.Link, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
                            Text(stringResource(R.string.scan_assign_rfid))
                        }
                        Text(
                            stringResource(R.string.scan_assign_rfid_hint),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.padding(top = 4.dp),
                        )
                    }
                }

                // Show RFID tags if in RFID mode
                if (isRfidMode && state.rfidTags.isNotEmpty()) {
                    Spacer(Modifier.height(24.dp))
                    Text(
                        stringResource(R.string.scan_tags_found, state.rfidTags.size),
                        style = MaterialTheme.typography.titleMedium,
                        color = Cyan,
                    )
                    Spacer(Modifier.height(8.dp))
                    LazyColumn(
                        modifier = Modifier.weight(1f),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        items(state.rfidTags) { tag ->
                            val identifier = tag.tid.ifBlank { tag.epc }
                            Card(
                                onClick = { viewModel.onRfidTagTapped(identifier) },
                                modifier = Modifier.fillMaxWidth(),
                            ) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(12.dp),
                                    horizontalArrangement = Arrangement.SpaceBetween,
                                    verticalAlignment = Alignment.CenterVertically,
                                ) {
                                    Column(modifier = Modifier.weight(1f)) {
                                        Text(
                                            tag.epc,
                                            style = MaterialTheme.typography.bodyLarge,
                                            fontFamily = FontFamily.Monospace,
                                            fontWeight = FontWeight.Bold,
                                        )
                                        if (tag.tid.isNotBlank()) {
                                            Text(
                                                "TID: ${tag.tid}",
                                                style = MaterialTheme.typography.bodySmall,
                                                fontFamily = FontFamily.Monospace,
                                                color = Cyan.copy(alpha = 0.7f),
                                            )
                                        }
                                    }
                                    Column(horizontalAlignment = Alignment.End) {
                                        Text(
                                            "${tag.rssi} dBm",
                                            style = MaterialTheme.typography.bodySmall,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                                        )
                                        Icon(
                                            Icons.Default.TouchApp,
                                            contentDescription = null,
                                            tint = Cyan.copy(alpha = 0.5f),
                                            modifier = Modifier.size(16.dp),
                                        )
                                    }
                                }
                            }
                        }
                    }
                    Spacer(Modifier.height(8.dp))
                    OutlinedButton(onClick = { viewModel.clearResult() }) {
                        Text(stringResource(R.string.scan_tags_clear))
                    }
                }
            }
        }
    }
}
