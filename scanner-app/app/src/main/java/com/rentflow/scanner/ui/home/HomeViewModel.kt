package com.rentflow.scanner.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.data.repository.AuthRepository
import com.rentflow.scanner.data.repository.ConfigRepository
import com.rentflow.scanner.data.repository.ScannerRepository
import com.rentflow.scanner.data.service.AppUpdateService
import com.rentflow.scanner.data.service.UpdateInfo
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

data class HomeUiState(
    val userName: String = "",
    val pendingScanCount: Int = 0,
    val updateInfo: UpdateInfo? = null,
    val homeCards: List<String> = SettingsDataStore.DEFAULT_HOME_CARDS,
    val industryLabels: Map<String, String> = emptyMap(),
)

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val scannerRepository: ScannerRepository,
    private val updateService: AppUpdateService,
    private val settingsDataStore: SettingsDataStore,
    private val configRepository: ConfigRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(HomeUiState())
    val uiState: StateFlow<HomeUiState> = _uiState

    init {
        viewModelScope.launch {
            authRepository.getCurrentUser().onSuccess { user ->
                _uiState.update { it.copy(userName = user.displayName) }
            }
        }
        viewModelScope.launch {
            scannerRepository.observePendingCount().collect { count ->
                _uiState.update { it.copy(pendingScanCount = count) }
            }
        }
        // Check for updates on start
        viewModelScope.launch {
            updateService.checkForUpdate()
        }
        viewModelScope.launch {
            updateService.updateAvailable.collect { info ->
                _uiState.update { it.copy(updateInfo = info) }
            }
        }
        // Load home cards from preferences
        viewModelScope.launch {
            settingsDataStore.scannerHomeCards.collect { cards ->
                _uiState.update { it.copy(homeCards = cards) }
            }
        }
        // Load industry labels from preferences
        viewModelScope.launch {
            settingsDataStore.industryLabels.collect { labels ->
                _uiState.update { it.copy(industryLabels = labels) }
            }
        }
        // Fetch latest industry config from server (non-blocking, updates preferences)
        viewModelScope.launch {
            configRepository.fetchAndStoreIndustryConfig()
        }
    }

    fun logout() {
        authRepository.logout()
    }

    fun downloadUpdate() {
        val info = _uiState.value.updateInfo ?: return
        updateService.downloadAndInstall(info)
    }
}
