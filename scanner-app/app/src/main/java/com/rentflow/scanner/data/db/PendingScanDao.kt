package com.rentflow.scanner.data.db

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Dao
interface PendingScanDao {
    @Insert
    suspend fun insert(scan: PendingScanEntity): Long

    @Query("SELECT * FROM pending_scans WHERE failed = 0 ORDER BY timestamp ASC")
    suspend fun getPending(): List<PendingScanEntity>

    @Query("SELECT * FROM pending_scans ORDER BY timestamp DESC")
    fun observeAll(): Flow<List<PendingScanEntity>>

    @Query("SELECT COUNT(*) FROM pending_scans WHERE failed = 0")
    fun observePendingCount(): Flow<Int>

    @Query("UPDATE pending_scans SET retryCount = retryCount + 1, failed = CASE WHEN retryCount >= 4 THEN 1 ELSE 0 END WHERE id = :id")
    suspend fun incrementRetry(id: Long)

    @Delete
    suspend fun delete(scan: PendingScanEntity)

    @Query("DELETE FROM pending_scans WHERE id = :id")
    suspend fun deleteById(id: Long)

    @Query("UPDATE pending_scans SET failed = 0, retryCount = 0 WHERE id = :id")
    suspend fun resetRetry(id: Long)
}
