package de.cratedesk.scanner.data.api

import de.cratedesk.scanner.data.api.dto.*
import retrofit2.http.*

interface ScannerApi {

    @POST("api/v1/scanner/scan")
    suspend fun scan(@Body request: ScanRequest): ApiResponse<ScanEventDto>

    @POST("api/v1/scanner/checkout")
    suspend fun checkout(@Body request: CheckoutRequest): ApiResponse<List<ScanEventDto>>

    @POST("api/v1/scanner/checkin")
    suspend fun checkin(@Body request: CheckinRequest): ApiResponse<List<ScanEventDto>>

    @POST("api/v1/scanner/bulk")
    suspend fun bulkSync(@Body request: BulkSyncRequest): ApiResponse<BulkSyncResponseDto>

    @POST("api/v1/scanner/adhoc-booking")
    suspend fun adhocBooking(@Body request: AdhocBookingRequest): ApiResponse<ScanEventDto>

    @GET("api/v1/scanner/events")
    suspend fun listEvents(
        @Query("page") page: Int = 1,
        @Query("per_page") perPage: Int = 50,
    ): ApiResponse<List<ScanEventDto>>

    @POST("api/v1/scanner/devices/register")
    suspend fun registerDevice(@Body request: RegisterDeviceRequest): ApiResponse<ScannerDeviceDto>

    @GET("api/v1/scanner/devices")
    suspend fun listDevices(): ApiResponse<List<ScannerDeviceDto>>

    @GET("api/v1/scanner/devices/ring")
    suspend fun checkRing(@Query("device_id") deviceId: String): ApiResponse<RingStatusDto>

    @POST("api/v1/scanner/devices/ring-ack")
    suspend fun ackRing(@Body request: RingAckRequest): Unit

    @POST("api/v1/scanner/devices/{id}/ring")
    suspend fun triggerRing(@Path("id") id: String): Unit
}

interface AuthApi {
    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): ApiResponse<LoginResponseDto>
}
