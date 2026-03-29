package com.rentflow.scanner.data.db

import androidx.room.Database
import androidx.room.RoomDatabase

@Database(entities = [PendingScanEntity::class], version = 1, exportSchema = false)
abstract class ScannerDatabase : RoomDatabase() {
    abstract fun pendingScanDao(): PendingScanDao
}
