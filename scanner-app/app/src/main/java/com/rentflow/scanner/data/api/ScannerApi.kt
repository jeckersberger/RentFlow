package com.rentflow.scanner.data.api

import com.rentflow.scanner.domain.model.Equipment
import com.rentflow.scanner.domain.model.ScanResult
import com.rentflow.scanner.domain.model.ScanSession
import retrofit2.Response
import retrofit2.http.*

interface ScannerApi {
    @GET("api/v1/scan/resolve/{barcode}")
    suspend fun resolveBarcode(@Path("barcode") barcode: String): Response<ApiResponse<Equipment>>

    @POST("api/v1/scan")
    suspend fun scan(@Body request: ScanRequest): Response<ApiResponse<ScanResult>>

    @POST("api/v1/scan/batch")
    suspend fun scanBatch(@Body requests: List<ScanRequest>): Response<ApiResponse<List<ScanResult>>>

    @POST("api/v1/scan/sync")
    suspend fun syncOffline(@Body requests: List<ScanRequest>): Response<ApiResponse<Unit>>

    @POST("api/v1/scanner/sessions")
    suspend fun createSession(@Body request: SessionCreateRequest): Response<ApiResponse<ScanSession>>

    @PUT("api/v1/scanner/sessions/{id}/end")
    suspend fun endSession(@Path("id") id: String): Response<ApiResponse<ScanSession>>

    @POST("api/v1/scanner/sessions/{id}/scan")
    suspend fun sessionScan(@Path("id") id: String, @Body request: ScanRequest): Response<ApiResponse<ScanResult>>

    @GET("api/v1/scanner/sessions/{id}/protocol")
    suspend fun sessionProtocol(@Path("id") id: String): Response<ApiResponse<List<ScanResult>>>

    @POST("api/v1/scanner/adhoc-booking")
    suspend fun adHocBooking(@Body request: AdHocBookingRequest): Response<ApiResponse<Equipment>>

    @POST("api/v1/scanner/devices/register")
    suspend fun registerDevice(@Body request: DeviceRegistration): Response<ApiResponse<Unit>>

    @GET("api/v1/scanner/devices/{device_id}/ring")
    suspend fun checkRingCommand(@Path("device_id") deviceId: String): Response<ApiResponse<RingCommandResponse>>

    @POST("api/v1/scanner/devices/{device_id}/ring-ack")
    suspend fun acknowledgeRing(@Path("device_id") deviceId: String): Response<ApiResponse<Unit>>

    @POST("api/v1/scanner/sessions/{id}/signature")
    suspend fun uploadSignature(@Path("id") sessionId: String, @Body body: Map<String, String>): Response<ApiResponse<Unit>>

    @PUT("api/v1/scan/equipment/{id}/location")
    suspend fun updateLocation(@Path("id") id: String, @Body body: Map<String, String>): Response<ApiResponse<Unit>>

    @GET("api/v1/scan/history")
    suspend fun scanHistory(@Query("barcode") barcode: String): Response<ApiResponse<List<ScanResult>>>

    @PUT("api/v1/scan/equipment/{id}/rfid")
    suspend fun pairRfidTag(@Path("id") id: String, @Body body: Map<String, String>): Response<ApiResponse<Unit>>

    @GET("api/v1/scan/equipment/untagged")
    suspend fun getUntaggedEquipment(): Response<ApiResponse<List<Equipment>>>

    @POST("api/v1/projects")
    suspend fun createProject(@Body request: CreateProjectRequest): Response<ApiResponse<ProjectCreated>>

    @GET("api/v1/customers")
    suspend fun searchCustomers(@Query("search") query: String): Response<ApiResponse<List<CustomerDto>>>

    @POST("api/v1/customers")
    suspend fun createCustomer(@Body request: CreateCustomerRequest): Response<ApiResponse<CustomerDto>>

    @GET("api/v1/inventory/jobs")
    suspend fun listInventoryJobs(): Response<ApiResponse<List<InventoryJob>>>

    @GET("api/v1/inventory/jobs/{id}/items")
    suspend fun getInventoryJobItems(@Path("id") id: String): Response<ApiResponse<List<Equipment>>>

    @POST("api/v1/inventory/jobs/{id}/complete")
    suspend fun completeInventoryJob(@Path("id") id: String, @Body body: InventoryCompleteRequest): Response<ApiResponse<Unit>>
}

data class InventoryJob(
    val id: String,
    val name: String,
    val zone: String? = null,
    val status: String = "open",
    val expectedCount: Int = 0,
    val scannedCount: Int = 0,
    val createdAt: String? = null,
)

data class InventoryCompleteRequest(
    val scannedBarcodes: List<String>,
    val missingBarcodes: List<String>,
    val unexpectedBarcodes: List<String>,
)

data class CreateProjectRequest(
    val name: String,
    val client_name: String? = null,
    val client_email: String? = null,
    val client_phone: String? = null,
    val client_address: AddressDto? = null,
    val venue_address: AddressDto? = null,
    val start_date: String? = null,
    val end_date: String? = null,
    val notes: String? = null,
    val is_dry_hire: Boolean = false,
)

data class AddressDto(
    val street: String? = null,
    val city: String? = null,
    val postal_code: String? = null,
    val country: String? = null,
)

data class ProjectCreated(
    val id: String,
    val name: String,
    val status: String? = null,
    val client_name: String? = null,
)

data class CustomerDto(
    val id: String,
    val name: String,
    val email: String? = null,
    val phone: String? = null,
    val address_street: String? = null,
    val address_city: String? = null,
    val address_postcode: String? = null,
    val address_country: String? = null,
)

data class CreateCustomerRequest(
    val name: String,
    val email: String? = null,
    val phone: String? = null,
    val address_street: String? = null,
    val address_city: String? = null,
    val address_postcode: String? = null,
    val address_country: String? = null,
)
