package com.rentflow.scanner.domain.model

data class PendingScan(
    val id: Long = 0,
    val barcode: String,
    val scanType: String,
    val projectId: String?,
    val notes: String?,
    val timestamp: Long,
    val retryCount: Int = 0,
    val failed: Boolean = false,
)
