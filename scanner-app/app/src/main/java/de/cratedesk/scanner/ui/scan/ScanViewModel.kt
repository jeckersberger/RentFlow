package de.cratedesk.scanner.ui.scan

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import de.cratedesk.scanner.data.api.ScannerApi
import de.cratedesk.scanner.data.api.dto.ScanEventDto
import de.cratedesk.scanner.data.api.dto.ScanRequest
import de.cratedesk.scanner.data.local.TokenManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ScanUiState(
    val isLoading: Boolean = false,
    val events: List<ScanEventDto> = emptyList(),
    val lastScanResult: ScanEventDto? = null,
    val error: String? = null,
)

@HiltViewModel
class ScanViewModel @Inject constructor(
    private val scannerApi: ScannerApi,
    private val tokenManager: TokenManager,
) : ViewModel() {

    private val _uiState = MutableStateFlow(ScanUiState())
    val uiState: StateFlow<ScanUiState> = _uiState

    init {
        loadEvents()
    }

    fun scan(barcodeOrRfid: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            try {
                val deviceId = tokenManager.deviceId.first() ?: "unknown"
                val isRfid = barcodeOrRfid.startsWith("RFID-") || barcodeOrRfid.length > 20
                val request = ScanRequest(
                    deviceId = deviceId,
                    barcode = if (!isRfid) barcodeOrRfid else null,
                    rfidTag = if (isRfid) barcodeOrRfid else null,
                )
                val response = scannerApi.scan(request)
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    lastScanResult = response.data,
                )
                loadEvents()
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = e.message ?: "Scan fehlgeschlagen",
                )
            }
        }
    }

    fun loadEvents() {
        viewModelScope.launch {
            try {
                val response = scannerApi.listEvents()
                _uiState.value = _uiState.value.copy(
                    events = response.data ?: emptyList(),
                    error = null,
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Events laden fehlgeschlagen",
                )
            }
        }
    }
}
