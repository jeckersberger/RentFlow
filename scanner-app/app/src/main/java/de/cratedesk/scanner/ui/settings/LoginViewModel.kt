package de.cratedesk.scanner.ui.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import de.cratedesk.scanner.data.api.AuthApi
import de.cratedesk.scanner.data.api.dto.LoginRequest
import de.cratedesk.scanner.data.local.TokenManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import javax.inject.Inject

data class LoginUiState(
    val isLoading: Boolean = false,
    val loginSuccess: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authApi: AuthApi,
    private val tokenManager: TokenManager,
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState

    val isLoggedIn = tokenManager.accessToken.map { it != null }

    fun login(email: String, password: String, tenantSlug: String, serverUrl: String) {
        viewModelScope.launch {
            _uiState.value = LoginUiState(isLoading = true)
            try {
                tokenManager.saveServerUrl(serverUrl)
                val response = authApi.login(LoginRequest(email, password, tenantSlug))
                val tokens = response.data?.tokens
                if (tokens != null) {
                    tokenManager.saveTokens(tokens.accessToken, tokens.refreshToken)
                    _uiState.value = LoginUiState(loginSuccess = true)
                } else {
                    _uiState.value = LoginUiState(error = "Login fehlgeschlagen")
                }
            } catch (e: Exception) {
                _uiState.value = LoginUiState(error = e.message ?: "Login fehlgeschlagen")
            }
        }
    }

    fun logout() {
        viewModelScope.launch {
            tokenManager.clearAll()
            _uiState.value = LoginUiState()
        }
    }
}
