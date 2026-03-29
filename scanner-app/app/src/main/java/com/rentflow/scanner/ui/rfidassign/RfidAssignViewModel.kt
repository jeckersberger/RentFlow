package com.rentflow.scanner.ui.rfidassign

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rentflow.scanner.data.hardware.HardwareScanner
import com.rentflow.scanner.data.hardware.ScanFeedback
import com.rentflow.scanner.data.repository.ScannerRepository
import com.rentflow.scanner.domain.model.Equipment
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class RfidAssignUiState(
    val equipmentList: List<Equipment> = emptyList(),
    val selectedEquipment: Equipment? = null,
    val isLoading: Boolean = false,
    val isScanning: Boolean = false,
    val error: String? = null,
    val successMessage: String? = null,
    val manualBarcode: String = "",
)

@HiltViewModel
class RfidAssignViewModel @Inject constructor(
    private val scannerRepository: ScannerRepository,
    private val hardwareScanner: HardwareScanner,
    private val scanFeedback: ScanFeedback,
) : ViewModel() {
    private val _uiState = MutableStateFlow(RfidAssignUiState())
    val uiState: StateFlow<RfidAssignUiState> = _uiState

    private var rfidJob: Job? = null

    init {
        loadUntagged()
    }

    fun loadUntagged() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            scannerRepository.getUntaggedEquipment().fold(
                onSuccess = { list ->
                    _uiState.update { it.copy(isLoading = false, equipmentList = list) }
                },
                onFailure = { e ->
                    _uiState.update { it.copy(isLoading = false, error = e.message) }
                },
            )
        }
    }

    fun updateManualBarcode(value: String) {
        _uiState.update { it.copy(manualBarcode = value) }
    }

    fun submitManualBarcode() {
        val barcode = _uiState.value.manualBarcode.trim()
        if (barcode.isBlank()) return
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            scannerRepository.resolveBarcode(barcode).fold(
                onSuccess = { equipment ->
                    _uiState.update { it.copy(isLoading = false, manualBarcode = "") }
                    selectEquipment(equipment)
                },
                onFailure = { e ->
                    scanFeedback.onScanError()
                    _uiState.update { it.copy(isLoading = false, error = e.message) }
                },
            )
        }
    }

    fun selectEquipment(equipment: Equipment) {
        _uiState.update { it.copy(selectedEquipment = equipment, isScanning = true, error = null, successMessage = null) }
        startRfidListening()
    }

    fun cancelAssign() {
        rfidJob?.cancel()
        hardwareScanner.stopRfid()
        _uiState.update { it.copy(selectedEquipment = null, isScanning = false) }
    }

    private fun startRfidListening() {
        hardwareScanner.startRfidBulkRead()
        rfidJob = viewModelScope.launch {
            hardwareScanner.rfidReadEvents.collect { event ->
                hardwareScanner.stopRfid()

                val tid = event.tid.ifBlank {
                    hardwareScanner.readTid(event.epc) ?: ""
                }
                val identifier = tid.ifBlank { event.epc }

                val eq = _uiState.value.selectedEquipment ?: return@collect

                scannerRepository.pairRfidTag(eq.id, identifier).fold(
                    onSuccess = {
                        scanFeedback.onScanSuccess()
                        _uiState.update {
                            it.copy(
                                isScanning = false,
                                selectedEquipment = null,
                                successMessage = "${eq.name}: $identifier",
                                // Remove from list
                                equipmentList = it.equipmentList.filter { e -> e.id != eq.id },
                            )
                        }
                    },
                    onFailure = { e ->
                        scanFeedback.onScanError()
                        _uiState.update { it.copy(error = e.message) }
                        // Restart for retry
                        startRfidListening()
                    },
                )
                return@collect
            }
        }
    }

    fun clearSuccess() {
        _uiState.update { it.copy(successMessage = null) }
    }

    override fun onCleared() {
        super.onCleared()
        rfidJob?.cancel()
        hardwareScanner.stopRfid()
        hardwareScanner.closeRfid()
    }
}
