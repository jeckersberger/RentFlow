package com.rentflow.scanner.data.api

data class ApiResponse<T>(
    val data: T?,
    val message: String?,
    val error: String? = null,
    val code: String? = null,
)

data class LoginResponse(
    val access_token: String,
    val refresh_token: String,
    val expires_in: Int = 3600,
    val token_type: String = "Bearer",
    val tenant_id: String? = null,
    val user: LoginUser? = null,
)

data class LoginUser(
    val id: String,
    val email: String,
    val first_name: String = "",
    val last_name: String = "",
    val roles: List<String> = emptyList(),
    val tenant_id: String? = null,
)

data class LoginRequest(
    val email: String,
    val password: String,
)

data class RefreshRequest(
    val refresh_token: String,
)

data class QrLoginRequest(
    val qr_token: String,
)

data class ScanRequest(
    val barcode: String,
    val scan_type: String,
    val user_id: String,
    val device_id: String,
    val device_type: String = "scanner",
    val project_id: String? = null,
    val location_id: String? = null,
    val notes: String? = null,
)

data class SessionCreateRequest(
    val type: String,
    val project_id: String? = null,
    val location_id: String? = null,
)

data class AdHocBookingRequest(
    val equipment_id: String,
    val barcode: String,
    val notes: String? = null,
)

data class DeviceRegistration(
    val device_id: String,
    val device_name: String,
    val device_type: String = "cf-h906",
    val fcm_token: String? = null,
)

data class RingCommandResponse(
    val ring: Boolean = false,
)
