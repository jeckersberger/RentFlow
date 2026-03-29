package com.rentflow.scanner.domain.model

data class Equipment(
    val id: String,
    val barcode: String,
    val name: String,
    val category: String,
    val status: EquipmentStatus,
    val location: String?,
    val projectName: String?,
    val projectId: String? = null,
    val rfidTag: String?,
    val imageUrl: String?,
)

enum class EquipmentStatus {
    AVAILABLE, CHECKED_OUT, IN_MAINTENANCE, DAMAGED, RETIRED
}
