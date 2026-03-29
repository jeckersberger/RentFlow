package com.rentflow.scanner.ui.queue

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rentflow.scanner.data.db.PendingScanEntity
import com.rentflow.scanner.data.repository.ScannerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PendingQueueViewModel @Inject constructor(
    private val scannerRepository: ScannerRepository,
) : ViewModel() {

    val pendingScans: StateFlow<List<PendingScanEntity>> =
        scannerRepository.observePendingScans()
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), emptyList())

    fun retryItem(id: Long) {
        viewModelScope.launch {
            scannerRepository.retryPendingScan(id)
        }
    }

    fun deleteItem(id: Long) {
        viewModelScope.launch {
            scannerRepository.deletePendingScan(id)
        }
    }

    fun retryAll() {
        viewModelScope.launch {
            pendingScans.value.forEach { scan ->
                scannerRepository.retryPendingScan(scan.id)
            }
        }
    }
}
