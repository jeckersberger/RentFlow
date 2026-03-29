package com.rentflow.scanner.data.db

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "pending_scans")
data class PendingScanEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val barcode: String,
    val scanType: String,
    val projectId: String?,
    val notes: String?,
    val timestamp: Long,
    val retryCount: Int = 0,
    val failed: Boolean = false,
    val userId: String,
    val deviceId: String,
)
