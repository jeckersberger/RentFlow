package com.rentflow.scanner.domain.model

data class ScanSession(
    val id: String,
    val type: ScanType,
    val scannedItems: List<ScanResult>,
    val startedAt: Long,
)
