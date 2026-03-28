package de.cratedesk.scanner.data.api.dto

import com.google.gson.annotations.SerializedName

data class ApiResponse<T>(
    val data: T? = null,
    val error: String? = null,
    val meta: ApiMeta? = null,
)

data class ApiMeta(
    val page: Int = 1,
    @SerializedName("per_page") val perPage: Int = 20,
    val total: Long = 0,
)

data class ScanEventDto(
    val id: String = "",
    @SerializedName("tenant_id") val tenantId: String = "",
    @SerializedName("user_id") val userId: String = "",
    @SerializedName("device_id") val deviceId: String = "",
    val barcode: String? = null,
    @SerializedName("rfid_tag") val rfidTag: String? = null,
    @SerializedName("equipment_id") val equipmentId: String? = null,
    val action: String = "scan",
    @SerializedName("project_id") val projectId: String? = null,
    @SerializedName("location_id") val locationId: String? = null,
    @SerializedName("condition_rating") val conditionRating: Int? = null,
    @SerializedName("condition_notes") val conditionNotes: String? = null,
    @SerializedName("gps_lat") val gpsLat: Double? = null,
    @SerializedName("gps_lng") val gpsLng: Double? = null,
    val timestamp: String = "",
    @SerializedName("synced_at") val syncedAt: String? = null,
    @SerializedName("created_at") val createdAt: String = "",
)

data class ScannerDeviceDto(
    val id: String = "",
    @SerializedName("tenant_id") val tenantId: String = "",
    @SerializedName("device_id") val deviceId: String = "",
    @SerializedName("device_name") val deviceName: String = "",
    @SerializedName("device_type") val deviceType: String = "cf-h906",
    @SerializedName("fcm_token") val fcmToken: String? = null,
    @SerializedName("ring_requested") val ringRequested: Boolean = false,
    @SerializedName("last_seen") val lastSeen: String? = null,
    @SerializedName("created_at") val createdAt: String = "",
)

data class BulkSyncResponseDto(
    val inserted: Int = 0,
    val skipped: Int = 0,
    val total: Int = 0,
)

data class RingStatusDto(
    @SerializedName("ring_requested") val ringRequested: Boolean = false,
)

// Request DTOs
data class ScanRequest(
    @SerializedName("device_id") val deviceId: String,
    val barcode: String? = null,
    @SerializedName("rfid_tag") val rfidTag: String? = null,
)

data class CheckoutRequest(
    @SerializedName("device_id") val deviceId: String,
    @SerializedName("project_id") val projectId: String,
    val items: List<CheckoutItem>,
)

data class CheckoutItem(
    @SerializedName("equipment_id") val equipmentId: String,
    val barcode: String? = null,
    @SerializedName("rfid_tag") val rfidTag: String? = null,
)

data class CheckinRequest(
    @SerializedName("device_id") val deviceId: String,
    @SerializedName("project_id") val projectId: String? = null,
    val items: List<CheckinItem>,
)

data class CheckinItem(
    @SerializedName("equipment_id") val equipmentId: String,
    val barcode: String? = null,
    @SerializedName("rfid_tag") val rfidTag: String? = null,
    @SerializedName("condition_rating") val conditionRating: Int? = null,
    @SerializedName("condition_notes") val conditionNotes: String? = null,
)

data class BulkSyncRequest(
    val events: List<BulkScanEventDto>,
)

data class BulkScanEventDto(
    @SerializedName("device_id") val deviceId: String,
    val barcode: String? = null,
    @SerializedName("rfid_tag") val rfidTag: String? = null,
    @SerializedName("equipment_id") val equipmentId: String? = null,
    val action: String,
    @SerializedName("project_id") val projectId: String? = null,
    @SerializedName("location_id") val locationId: String? = null,
    @SerializedName("condition_rating") val conditionRating: Int? = null,
    @SerializedName("condition_notes") val conditionNotes: String? = null,
    @SerializedName("gps_lat") val gpsLat: Double? = null,
    @SerializedName("gps_lng") val gpsLng: Double? = null,
    val timestamp: String,
)

data class AdhocBookingRequest(
    @SerializedName("device_id") val deviceId: String,
    val barcode: String? = null,
    @SerializedName("rfid_tag") val rfidTag: String? = null,
    @SerializedName("equipment_id") val equipmentId: String? = null,
    @SerializedName("project_id") val projectId: String? = null,
)

data class RegisterDeviceRequest(
    @SerializedName("device_id") val deviceId: String,
    @SerializedName("device_name") val deviceName: String,
    @SerializedName("device_type") val deviceType: String = "cf-h906",
)

data class RingAckRequest(
    @SerializedName("device_id") val deviceId: String,
)

data class LoginRequest(
    val email: String,
    val password: String,
    @SerializedName("tenant_slug") val tenantSlug: String,
)

data class LoginResponseDto(
    val tokens: TokensDto,
    val user: UserDto,
)

data class TokensDto(
    @SerializedName("access_token") val accessToken: String,
    @SerializedName("refresh_token") val refreshToken: String,
)

data class UserDto(
    val id: String,
    val email: String,
    @SerializedName("first_name") val firstName: String,
    @SerializedName("last_name") val lastName: String,
    val role: String,
)
