package com.rentflow.scanner.data.repository

import com.rentflow.scanner.data.api.WarehouseApi
import com.rentflow.scanner.data.preferences.TokenManager
import com.rentflow.scanner.domain.model.WarehouseZone
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WarehouseRepository @Inject constructor(
    private val warehouseApi: WarehouseApi,
    private val tokenManager: TokenManager,
) {
    suspend fun listZones(): Result<List<WarehouseZone>> {
        // Always try server first
        return try {
            val response = warehouseApi.listZones()
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                // Server returned error — fallback to demo only in demo mode
                if (tokenManager.getAccessToken() == "demo-access-token") {
                    Result.success(demoZones())
                } else {
                    Result.failure(Exception("Failed to load zones"))
                }
            }
        } catch (e: Exception) {
            // Network error — fallback to demo only in demo mode
            if (tokenManager.getAccessToken() == "demo-access-token") {
                Result.success(demoZones())
            } else {
                Result.failure(e)
            }
        }
    }

    private fun demoZones() = listOf(
        WarehouseZone("1", "Lager", 0),
        WarehouseZone("2", "Halle A", 0),
        WarehouseZone("3", "Halle B", 0),
        WarehouseZone("4", "Außenlager", 0),
        WarehouseZone("5", "Werkstatt", 0),
        WarehouseZone("6", "Baustelle", 0),
    )
}
