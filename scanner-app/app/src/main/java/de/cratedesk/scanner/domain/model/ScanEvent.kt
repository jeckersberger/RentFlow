package de.cratedesk.scanner.domain.model

data class ScanEvent(
    val id: String = "",
    val tenantId: String = "",
    val userId: String = "",
    val deviceId: String = "",
    val barcode: String? = null,
    val rfidTag: String? = null,
    val equipmentId: String? = null,
    val action: String = "scan",
    val projectId: String? = null,
    val locationId: String? = null,
    val conditionRating: Int? = null,
    val conditionNotes: String? = null,
    val gpsLat: Double? = null,
    val gpsLng: Double? = null,
    val timestamp: String = "",
    val syncedAt: String? = null,
    val createdAt: String = "",
)

data class ScannerDevice(
    val id: String = "",
    val tenantId: String = "",
    val deviceId: String = "",
    val deviceName: String = "",
    val deviceType: String = "cf-h906",
    val fcmToken: String? = null,
    val ringRequested: Boolean = false,
    val lastSeen: String? = null,
    val createdAt: String = "",
)

data class BulkSyncResponse(
    val inserted: Int = 0,
    val skipped: Int = 0,
    val total: Int = 0,
)

enum class ScanAction(val value: String) {
    SCAN("scan"),
    CHECKOUT("checkout"),
    CHECKIN("checkin"),
    INVENTORY_SCAN("inventory_scan"),
    ADHOC_BOOKING("adhoc_booking"),
}
