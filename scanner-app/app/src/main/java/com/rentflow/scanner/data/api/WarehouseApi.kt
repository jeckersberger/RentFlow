package com.rentflow.scanner.data.api

import com.rentflow.scanner.domain.model.WarehouseZone
import retrofit2.Response
import retrofit2.http.GET

interface WarehouseApi {
    @GET("api/v1/warehouse/zones")
    suspend fun listZones(): Response<ApiResponse<List<WarehouseZone>>>
}
