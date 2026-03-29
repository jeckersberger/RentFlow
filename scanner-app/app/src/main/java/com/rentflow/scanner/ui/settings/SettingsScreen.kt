package com.rentflow.scanner.ui.settings

import androidx.compose.foundation.layout.*
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
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.BuildConfig
import com.rentflow.scanner.R
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.data.service.DownloadState
import com.rentflow.scanner.ui.theme.Success
import com.rentflow.scanner.ui.components.UpdateBanner
import com.rentflow.scanner.ui.theme.Cyan
import com.rentflow.scanner.ui.theme.Success

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    onBack: () -> Unit,
    viewModel: SettingsViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.settings_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
            )
        },
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp).verticalScroll(rememberScrollState())) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                OutlinedTextField(
                    value = state.serverUrl,
                    onValueChange = viewModel::onServerUrlChange,
                    label = { Text(stringResource(R.string.settings_server_url)) },
                    modifier = Modifier.weight(1f),
                    singleLine = true,
                )
                Spacer(Modifier.width(8.dp))
                IconButton(onClick = viewModel::checkConnection) {
                    when (state.connectionStatus) {
                        ConnectionStatus.CHECKING -> CircularProgressIndicator(modifier = Modifier.size(24.dp), strokeWidth = 2.dp)
                        ConnectionStatus.CONNECTED -> Icon(Icons.Default.CheckCircle, contentDescription = null, tint = Success)
                        ConnectionStatus.FAILED -> Icon(Icons.Default.Cancel, contentDescription = null, tint = com.rentflow.scanner.ui.theme.Error)
                        ConnectionStatus.UNKNOWN -> Icon(Icons.Default.Wifi, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
            Spacer(Modifier.height(16.dp))
            Text(stringResource(R.string.settings_language), style = MaterialTheme.typography.titleLarge)
            Spacer(Modifier.height(8.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                FilterChip(
                    selected = state.language == "de",
                    onClick = { viewModel.onLanguageChange("de") },
                    label = { Text("Deutsch") },
                )
                FilterChip(
                    selected = state.language == "en",
                    onClick = { viewModel.onLanguageChange("en") },
                    label = { Text("English") },
                )
            }
            Spacer(Modifier.height(24.dp))
            Text(stringResource(R.string.settings_scan_mode), style = MaterialTheme.typography.titleLarge)
            Spacer(Modifier.height(8.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                FilterChip(
                    selected = state.scanMode == SettingsDataStore.SCAN_MODE_BARCODE,
                    onClick = { viewModel.onScanModeChange(SettingsDataStore.SCAN_MODE_BARCODE) },
                    label = { Text("Barcode") },
                    leadingIcon = {
                        Icon(Icons.Default.QrCodeScanner, contentDescription = null, modifier = Modifier.size(18.dp))
                    },
                )
                FilterChip(
                    selected = state.scanMode == SettingsDataStore.SCAN_MODE_RFID,
                    onClick = { viewModel.onScanModeChange(SettingsDataStore.SCAN_MODE_RFID) },
                    label = { Text("RFID") },
                    leadingIcon = {
                        Icon(Icons.Default.Sensors, contentDescription = null, modifier = Modifier.size(18.dp))
                    },
                )
            }
            Spacer(Modifier.height(24.dp))
            Text(stringResource(R.string.settings_logout_timeout), style = MaterialTheme.typography.titleLarge)
            Spacer(Modifier.height(8.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                listOf(15, 30, 60, 120, 480).forEach { minutes ->
                    FilterChip(
                        selected = state.logoutTimeoutMinutes == minutes,
                        onClick = { viewModel.onLogoutTimeoutChange(minutes) },
                        label = {
                            Text(if (minutes < 60) stringResource(R.string.settings_minutes, minutes)
                                 else stringResource(R.string.settings_hours, minutes / 60))
                        },
                    )
                }
            }
            Spacer(Modifier.height(24.dp))
            Button(
                onClick = viewModel::save,
                modifier = Modifier.fillMaxWidth().height(56.dp),
            ) {
                Text(stringResource(R.string.confirm))
            }
            if (state.saved) {
                Spacer(Modifier.height(16.dp))
                Text(stringResource(R.string.saved), color = Success)
            }

            Spacer(Modifier.height(32.dp))
            HorizontalDivider()
            Spacer(Modifier.height(16.dp))

            // Version & Update
            Text(stringResource(R.string.settings_version), style = MaterialTheme.typography.titleLarge)
            Spacer(Modifier.height(8.dp))
            Text(
                "v${BuildConfig.VERSION_NAME} (Build ${BuildConfig.VERSION_CODE})",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
            )
            Spacer(Modifier.height(12.dp))

            state.updateInfo?.let { info ->
                UpdateBanner(
                    updateInfo = info,
                    onUpdate = { viewModel.downloadUpdate() },
                )
            }

            // Download state feedback
            when (state.downloadState) {
                DownloadState.DOWNLOADING -> {
                    Spacer(Modifier.height(8.dp))
                    Row {
                        CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp)
                        Spacer(Modifier.width(8.dp))
                        Text(stringResource(R.string.update_downloading), color = Cyan)
                    }
                }
                DownloadState.INSTALLING -> {
                    Spacer(Modifier.height(8.dp))
                    Text(stringResource(R.string.update_installing), color = Success)
                }
                DownloadState.FAILED -> {
                    Spacer(Modifier.height(8.dp))
                    Text(stringResource(R.string.update_failed), color = MaterialTheme.colorScheme.error)
                }
                DownloadState.IDLE -> {}
            }

            Spacer(Modifier.height(12.dp))

            OutlinedButton(
                onClick = { viewModel.checkForUpdate() },
                modifier = Modifier.fillMaxWidth(),
                enabled = !state.isCheckingUpdate && state.downloadState != DownloadState.DOWNLOADING,
            ) {
                if (state.isCheckingUpdate) {
                    CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp)
                    Spacer(Modifier.width(8.dp))
                }
                Text(
                    if (state.isCheckingUpdate) stringResource(R.string.settings_checking)
                    else stringResource(R.string.settings_check_update)
                )
            }
        }
    }
}
