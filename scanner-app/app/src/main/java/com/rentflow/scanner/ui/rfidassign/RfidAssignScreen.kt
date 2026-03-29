package com.rentflow.scanner.ui.rfidassign

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.R
import com.rentflow.scanner.ui.theme.Cyan
import com.rentflow.scanner.ui.theme.Success

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RfidAssignScreen(
    onBack: () -> Unit,
    viewModel: RfidAssignViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(state.successMessage) {
        state.successMessage?.let {
            snackbarHostState.showSnackbar(it)
            viewModel.clearSuccess()
        }
    }

    // RFID scan dialog
    if (state.isScanning && state.selectedEquipment != null) {
        AlertDialog(
            onDismissRequest = { viewModel.cancelAssign() },
            icon = { Icon(Icons.Default.Nfc, contentDescription = null, tint = Cyan, modifier = Modifier.size(48.dp)) },
            title = { Text(stringResource(R.string.rfid_assign_scan_rfid)) },
            text = {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        state.selectedEquipment!!.name,
                        style = MaterialTheme.typography.titleMedium,
                    )
                    Spacer(Modifier.height(8.dp))
                    Text(
                        stringResource(R.string.rfid_assign_scan_rfid_hint),
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        textAlign = TextAlign.Center,
                    )
                    Spacer(Modifier.height(16.dp))
                    CircularProgressIndicator(color = Cyan)
                    state.error?.let {
                        Spacer(Modifier.height(8.dp))
                        Text(it, color = MaterialTheme.colorScheme.error)
                    }
                }
            },
            confirmButton = {},
            dismissButton = {
                TextButton(onClick = { viewModel.cancelAssign() }) {
                    Text(stringResource(R.string.rfid_assign_cancel))
                }
            },
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(stringResource(R.string.rfid_assign_title))
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.loadUntagged() }) {
                        Icon(Icons.Default.Refresh, contentDescription = null)
                    }
                },
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) },
    ) { padding ->
        when {
            state.isLoading && state.equipmentList.isEmpty() -> {
                Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            }

            else -> {
                LazyColumn(
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    item {
                        OutlinedTextField(
                            value = state.manualBarcode,
                            onValueChange = { viewModel.updateManualBarcode(it) },
                            label = { Text(stringResource(R.string.rfid_assign_manual_label)) },
                            singleLine = true,
                            modifier = Modifier.fillMaxWidth(),
                            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                            keyboardActions = KeyboardActions(onSearch = { viewModel.submitManualBarcode() }),
                            trailingIcon = {
                                IconButton(
                                    onClick = { viewModel.submitManualBarcode() },
                                    enabled = state.manualBarcode.isNotBlank(),
                                ) {
                                    Icon(Icons.Default.Search, contentDescription = null)
                                }
                            },
                        )
                    }

                    if (state.equipmentList.isEmpty() && !state.isLoading) {
                        item {
                            Column(
                                modifier = Modifier.fillMaxWidth().padding(vertical = 32.dp),
                                horizontalAlignment = Alignment.CenterHorizontally,
                            ) {
                                Icon(
                                    Icons.Default.CheckCircle,
                                    contentDescription = null,
                                    modifier = Modifier.size(64.dp),
                                    tint = Success,
                                )
                                Spacer(Modifier.height(16.dp))
                                Text(
                                    stringResource(R.string.rfid_assign_all_done),
                                    style = MaterialTheme.typography.titleLarge,
                                    color = Success,
                                )
                                Spacer(Modifier.height(8.dp))
                                Text(
                                    stringResource(R.string.rfid_assign_all_done_hint),
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                )
                            }
                        }
                    }

                    if (state.equipmentList.isNotEmpty()) {
                        item {
                            Text(
                                stringResource(R.string.rfid_assign_count, state.equipmentList.size),
                                style = MaterialTheme.typography.labelLarge,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                                modifier = Modifier.padding(top = 8.dp, bottom = 8.dp),
                            )
                        }
                    }

                    items(state.equipmentList, key = { it.id }) { equipment ->
                        Card(
                            onClick = { viewModel.selectEquipment(equipment) },
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            Row(
                                modifier = Modifier.fillMaxWidth().padding(16.dp),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Column(modifier = Modifier.weight(1f)) {
                                    Text(
                                        equipment.name,
                                        style = MaterialTheme.typography.titleMedium,
                                        maxLines = 1,
                                        overflow = TextOverflow.Ellipsis,
                                    )
                                    Text(
                                        equipment.barcode,
                                        style = MaterialTheme.typography.bodySmall,
                                        fontFamily = FontFamily.Monospace,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    )
                                    equipment.category.takeIf { it.isNotBlank() }?.let {
                                        Text(
                                            it,
                                            style = MaterialTheme.typography.bodySmall,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                                        )
                                    }
                                }
                                Icon(
                                    Icons.Default.Nfc,
                                    contentDescription = null,
                                    tint = Cyan,
                                    modifier = Modifier.size(28.dp),
                                )
                            }
                        }
                    }
                }
            }
        }

        state.error?.takeIf { !state.isScanning }?.let {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                Text(it, color = MaterialTheme.colorScheme.error, textAlign = TextAlign.Center)
            }
        }
    }
}
