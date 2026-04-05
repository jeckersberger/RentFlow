package com.rentflow.scanner.ui.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.data.repository.AuthRepository
import com.rentflow.scanner.data.service.SessionTimeoutManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class LoginUiState(
    val serverUrl: String = "",
    val email: String = "",
    val password: String = "",
    val isLoading: Boolean = false,
    val error: String? = null,
    val loginSuccess: Boolean = false,
    val showQrScanner: Boolean = false,
    val devMode: Boolean = false,
    val devTapCount: Int = 0,
)

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val sessionTimeoutManager: SessionTimeoutManager,
    private val settingsDataStore: SettingsDataStore,
) : ViewModel() {
    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState

    init {
        viewModelScope.launch {
            settingsDataStore.serverUrl.collect { url ->
                _uiState.update { it.copy(serverUrl = url) }
            }
        }
    }

    fun onServerUrlChange(url: String) {
        _uiState.update { it.copy(serverUrl = url, error = null) }
    }

    fun onLogoTap() {
        val current = _uiState.value
        val newCount = current.devTapCount + 1
        if (newCount >= 5) {
            _uiState.update { it.copy(devMode = true, devTapCount = 0) }
        } else {
            _uiState.update { it.copy(devTapCount = newCount) }
            // Reset counter after 2 seconds of inactivity
            viewModelScope.launch {
                delay(2000)
                _uiState.update { it.copy(devTapCount = 0) }
            }
        }
    }

    fun onEmailChange(email: String) {
        _uiState.update { it.copy(email = email, error = null) }
    }

    fun onPasswordChange(password: String) {
        _uiState.update { it.copy(password = password, error = null) }
    }

    fun onLoginClick() {
        val state = _uiState.value
        // In normal mode, server URL comes from settings (default or QR); never blank
        if (state.serverUrl.isBlank()) {
            _uiState.update { it.copy(error = "Server-URL erforderlich") }
            return
        }
        if (state.email.isBlank() || state.password.isBlank()) {
            _uiState.update { it.copy(error = "Benutzername und Passwort erforderlich") }
            return
        }
        viewModelScope.launch {
            // Save server URL before login
            settingsDataStore.setServerUrl(state.serverUrl)
            _uiState.update { it.copy(isLoading = true, error = null) }
            val result = authRepository.login(state.email, state.password)
            _uiState.update {
                if (result.isSuccess) {
                    sessionTimeoutManager.reset()
                    it.copy(isLoading = false, loginSuccess = true)
                } else {
                    it.copy(isLoading = false, error = result.exceptionOrNull()?.message ?: "Login fehlgeschlagen")
                }
            }
        }
    }

    fun openQrScanner() {
        _uiState.update { it.copy(showQrScanner = true, error = null) }
    }

    fun closeQrScanner() {
        _uiState.update { it.copy(showQrScanner = false) }
    }

    fun onQrCodeScanned(qrData: String) {
        _uiState.update { it.copy(showQrScanner = false) }

        // Try to parse as JSON: {"token":"...","server":"https://..."} or {"url":"...","token":"..."}
        try {
            val json = org.json.JSONObject(qrData)
            val url = json.optString("server", "").ifBlank { json.optString("url", "") }
            val token = json.optString("token", "")

            if (url.isNotBlank()) {
                _uiState.update { it.copy(serverUrl = url) }
            }

            if (token.isNotBlank()) {
                performQrLogin(token, url)
            } else if (url.isNotBlank()) {
                viewModelScope.launch { settingsDataStore.setServerUrl(url) }
            }
            return
        } catch (_: Exception) {
            // Not JSON — try other formats
        }

        // Deep link: rentflow://qr-login/TOKEN
        if (qrData.startsWith("rentflow://qr-login/")) {
            val token = qrData.removePrefix("rentflow://qr-login/")
            if (token.isNotBlank()) {
                performQrLogin(token)
                return
            }
        }

        // If it looks like a URL, use as server URL
        if (qrData.startsWith("http")) {
            viewModelScope.launch { settingsDataStore.setServerUrl(qrData) }
            _uiState.update { it.copy(serverUrl = qrData) }
            return
        }

        // Fallback: treat as plain login token
        performQrLogin(qrData)
    }

    private fun performQrLogin(token: String, serverUrl: String? = null) {
        val currentServerUrl = serverUrl ?: _uiState.value.serverUrl
        if (currentServerUrl.isBlank()) {
            _uiState.update { it.copy(error = "Server-URL erforderlich. Bitte zuerst Server-URL eingeben.") }
            return
        }
        _uiState.update { it.copy(isLoading = true, error = null) }
        viewModelScope.launch {
            settingsDataStore.setServerUrl(currentServerUrl)
            val result = authRepository.qrLogin(token)
            _uiState.update {
                if (result.isSuccess) {
                    sessionTimeoutManager.reset()
                    it.copy(isLoading = false, loginSuccess = true)
                } else {
                    it.copy(isLoading = false, error = result.exceptionOrNull()?.message ?: "QR-Login fehlgeschlagen")
                }
            }
        }
    }
}
