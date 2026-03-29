package com.rentflow.scanner.ui.equipment

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rentflow.scanner.data.hardware.HardwareScanner
import com.rentflow.scanner.data.hardware.RfidReadEvent
import com.rentflow.scanner.data.hardware.RfidWriteResult
import com.rentflow.scanner.data.repository.ScannerRepository
import com.rentflow.scanner.data.repository.WarehouseRepository
import com.rentflow.scanner.domain.model.Equipment
import com.rentflow.scanner.domain.model.ScanResult
import com.rentflow.scanner.domain.model.WarehouseZone
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class EquipmentDetailUiState(
    val equipment: Equipment? = null,
    val isLoading: Boolean = true,
    val error: String? = null,
    val rfidWriteResult: RfidWriteResult? = null,
    val isWritingRfid: Boolean = false,
    val locationSaved: String? = null,
    val zones: List<WarehouseZone> = emptyList(),
    val showOverwriteDialog: Boolean = false,
    val existingTagEpc: String? = null,
    val rfidVerified: Boolean = false,
    val history: List<ScanResult> = emptyList(),
    val isLocating: Boolean = false,
    val rssiValue: Int = 0,
)

@HiltViewModel
class EquipmentDetailViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val scannerRepository: ScannerRepository,
    private val warehouseRepository: WarehouseRepository,
    private val hardwareScanner: HardwareScanner,
) : ViewModel() {
    private val barcode: String = savedStateHandle["barcode"] ?: ""
    private val _uiState = MutableStateFlow(EquipmentDetailUiState())
    val uiState: StateFlow<EquipmentDetailUiState> = _uiState

    private var locatingJob: Job? = null

    init {
        loadEquipment()
        loadZones()
        loadHistory()
    }

    private fun loadEquipment() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            scannerRepository.resolveBarcode(barcode).fold(
                onSuccess = { eq -> _uiState.update { it.copy(isLoading = false, equipment = eq) } },
                onFailure = { e -> _uiState.update { it.copy(isLoading = false, error = e.message) } },
            )
        }
    }

    private fun loadZones() {
        viewModelScope.launch {
            warehouseRepository.listZones().onSuccess { zones ->
                _uiState.update { it.copy(zones = zones) }
            }
        }
    }

    private fun loadHistory() {
        viewModelScope.launch {
            scannerRepository.getScanHistory(barcode).onSuccess { history ->
                _uiState.update { it.copy(history = history) }
            }
        }
    }

    fun startLocating() {
        _uiState.update { it.copy(isLocating = true, rssiValue = 0) }
        hardwareScanner.startRfidRead()
        locatingJob = viewModelScope.launch {
            hardwareScanner.rfidReadEvents.collect { event ->
                _uiState.update { it.copy(rssiValue = event.rssi) }
            }
        }
    }

    fun stopLocating() {
        locatingJob?.cancel()
        locatingJob = null
        hardwareScanner.stopRfid()
        _uiState.update { it.copy(isLocating = false, rssiValue = 0) }
    }

    fun updateLocation(newLocation: String) {
        val eq = _uiState.value.equipment ?: return
        viewModelScope.launch {
            scannerRepository.updateEquipmentLocation(eq.id, newLocation).fold(
                onSuccess = {
                    _uiState.update {
                        it.copy(
                            equipment = eq.copy(location = newLocation),
                            locationSaved = "Lagerort auf \"$newLocation\" geändert",
                        )
                    }
                },
                onFailure = { e ->
                    _uiState.update { it.copy(error = e.message) }
                },
            )
        }
    }

    private var pairJob: Job? = null

    fun pairRfidTag() {
        val eq = _uiState.value.equipment ?: return

        // If already pairing, cancel it
        if (_uiState.value.isWritingRfid) {
            cancelPairing()
            return
        }

        _uiState.update { it.copy(isWritingRfid = true, rfidWriteResult = null, rfidVerified = false) }

        // Start continuous RFID bulk read — waits for user to hold a tag
        hardwareScanner.startRfidBulkRead()

        pairJob = viewModelScope.launch {
            // Wait for the first tag to appear
            hardwareScanner.rfidReadEvents.collect { event ->
                // Got a tag! Stop reading immediately
                hardwareScanner.stopRfid()

                // Use TID as unique identifier (read-only, globally unique)
                val tid = event.tid.ifBlank {
                    hardwareScanner.readTid(event.epc) ?: ""
                }
                val identifier = tid.ifBlank { event.epc }

                // Send pairing to backend
                scannerRepository.pairRfidTag(eq.id, identifier).fold(
                    onSuccess = {
                        _uiState.update {
                            it.copy(
                                isWritingRfid = false,
                                rfidWriteResult = RfidWriteResult(true),
                                rfidVerified = true,
                                equipment = eq.copy(rfidTag = identifier),
                            )
                        }
                    },
                    onFailure = { e ->
                        _uiState.update {
                            it.copy(
                                isWritingRfid = false,
                                rfidWriteResult = RfidWriteResult(false, e.message),
                            )
                        }
                    },
                )
                // Stop collecting after first tag
                return@collect
            }
        }
    }

    private fun cancelPairing() {
        pairJob?.cancel()
        pairJob = null
        hardwareScanner.stopRfid()
        _uiState.update { it.copy(isWritingRfid = false) }
    }
}
