package com.rentflow.scanner.ui.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rentflow.scanner.RentFlowScannerApp
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.data.service.AppUpdateService
import com.rentflow.scanner.data.service.DownloadState
import com.rentflow.scanner.data.service.UpdateInfo
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.net.HttpURLConnection
import java.net.URL
import javax.inject.Inject

enum class ConnectionStatus { UNKNOWN, CHECKING, CONNECTED, FAILED }

data class SettingsUiState(
    val serverUrl: String = "",
    val language: String = "de",
    val scanMode: String = "barcode",
    val logoutTimeoutMinutes: Int = 60,
    val saved: Boolean = false,
    val connectionStatus: ConnectionStatus = ConnectionStatus.UNKNOWN,
    val updateInfo: UpdateInfo? = null,
    val isCheckingUpdate: Boolean = false,
    val downloadState: DownloadState = DownloadState.IDLE,
)

@HiltViewModel
class SettingsViewModel @Inject constructor(
    private val settingsDataStore: SettingsDataStore,
    private val updateService: AppUpdateService,
) : ViewModel() {
    private val _uiState = MutableStateFlow(SettingsUiState())
    val uiState: StateFlow<SettingsUiState> = _uiState

    init {
        viewModelScope.launch {
            settingsDataStore.serverUrl.collect { url ->
                _uiState.update { it.copy(serverUrl = url) }
            }
        }
        viewModelScope.launch {
            settingsDataStore.language.collect { lang ->
                _uiState.update { it.copy(language = lang) }
            }
        }
        viewModelScope.launch {
            settingsDataStore.scanMode.collect { mode ->
                _uiState.update { it.copy(scanMode = mode) }
            }
        }
        viewModelScope.launch {
            settingsDataStore.logoutTimeoutMinutes.collect { minutes ->
                _uiState.update { it.copy(logoutTimeoutMinutes = minutes) }
            }
        }
        viewModelScope.launch {
            updateService.updateAvailable.collect { info ->
                _uiState.update { it.copy(updateInfo = info) }
            }
        }
        viewModelScope.launch {
            updateService.downloadState.collect { state ->
                _uiState.update { it.copy(downloadState = state) }
            }
        }
    }

    fun onServerUrlChange(url: String) {
        _uiState.update { it.copy(serverUrl = url, saved = false) }
    }

    fun onLanguageChange(lang: String) {
        _uiState.update { it.copy(language = lang, saved = false) }
    }

    fun onScanModeChange(mode: String) {
        _uiState.update { it.copy(scanMode = mode, saved = false) }
    }

    fun onLogoutTimeoutChange(minutes: Int) {
        _uiState.update { it.copy(logoutTimeoutMinutes = minutes, saved = false) }
    }

    fun save() {
        viewModelScope.launch {
            settingsDataStore.setServerUrl(_uiState.value.serverUrl)
            settingsDataStore.setLanguage(_uiState.value.language)
            settingsDataStore.setScanMode(_uiState.value.scanMode)
            settingsDataStore.setLogoutTimeout(_uiState.value.logoutTimeoutMinutes)
            RentFlowScannerApp.applyLanguage(_uiState.value.language)
            _uiState.update { it.copy(saved = true) }
        }
    }

    fun checkConnection() {
        val url = _uiState.value.serverUrl.trim()
        if (url.isBlank()) {
            _uiState.update { it.copy(connectionStatus = ConnectionStatus.FAILED) }
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(connectionStatus = ConnectionStatus.CHECKING) }
            val result = withContext(Dispatchers.IO) {
                try {
                    val base = if (url.endsWith("/")) url else "$url/"
                    val conn = URL("${base}api/v1/auth/me").openConnection() as HttpURLConnection
                    conn.connectTimeout = 5000
                    conn.readTimeout = 5000
                    conn.requestMethod = "GET"
                    val code = conn.responseCode
                    conn.disconnect()
                    code in 200..499 // Any response means server is reachable
                } catch (_: Exception) {
                    false
                }
            }
            _uiState.update {
                it.copy(connectionStatus = if (result) ConnectionStatus.CONNECTED else ConnectionStatus.FAILED)
            }
        }
    }

    fun checkForUpdate() {
        viewModelScope.launch {
            _uiState.update { it.copy(isCheckingUpdate = true) }
            updateService.checkForUpdate()
            _uiState.update { it.copy(isCheckingUpdate = false) }
        }
    }

    fun downloadUpdate() {
        val info = _uiState.value.updateInfo ?: return
        updateService.downloadAndInstall(info)
    }
}
