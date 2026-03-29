package com.rentflow.scanner.domain.model

data class ScanResult(
    val id: String,
    val barcode: String,
    val equipment: Equipment?,
    val timestamp: Long,
    val scanType: ScanType,
)

enum class ScanType { IN, OUT, INVENTORY, LOOKUP }
