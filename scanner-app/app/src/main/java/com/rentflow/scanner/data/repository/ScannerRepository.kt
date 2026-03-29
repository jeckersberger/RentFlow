package com.rentflow.scanner.data.repository

import android.content.Context
import com.rentflow.scanner.data.api.AdHocBookingRequest
import com.rentflow.scanner.data.api.CreateCustomerRequest
import com.rentflow.scanner.data.api.CreateProjectRequest
import com.rentflow.scanner.data.api.CustomerDto
import com.rentflow.scanner.data.api.ProjectCreated
import com.rentflow.scanner.data.api.ScanRequest
import com.rentflow.scanner.data.api.ScannerApi
import com.rentflow.scanner.data.api.SessionCreateRequest
import com.rentflow.scanner.data.db.PendingScanDao
import com.rentflow.scanner.data.db.PendingScanEntity
import com.rentflow.scanner.data.preferences.TokenManager
import com.rentflow.scanner.domain.model.Equipment
import com.rentflow.scanner.domain.model.EquipmentStatus
import com.rentflow.scanner.domain.model.ScanResult
import com.rentflow.scanner.domain.model.ScanSession
import com.rentflow.scanner.worker.SyncWorker
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ScannerRepository @Inject constructor(
    private val scannerApi: ScannerApi,
    private val pendingScanDao: PendingScanDao,
    private val tokenManager: TokenManager,
    @ApplicationContext private val context: Context,
) {
    suspend fun resolveBarcode(barcode: String): Result<Equipment> {
        // Always try server first
        return try {
            val response = scannerApi.resolveBarcode(barcode)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                // Server error — fallback to demo data only in demo mode
                if (tokenManager.getAccessToken() == "demo-access-token") {
                    Result.success(demoEquipment(barcode))
                } else {
                    Result.failure(Exception(response.body()?.message ?: "Equipment not found"))
                }
            }
        } catch (e: Exception) {
            if (tokenManager.getAccessToken() == "demo-access-token") {
                Result.success(demoEquipment(barcode))
            } else {
                Result.failure(e)
            }
        }
    }

    private fun demoEquipment(barcode: String) = Equipment(
        id = "demo-${barcode}",
        barcode = barcode,
        name = "Equipment $barcode",
        category = "Demo",
        status = EquipmentStatus.AVAILABLE,
        location = "Lager",
        projectName = null,
        rfidTag = null,
        imageUrl = null,
    )

    suspend fun scan(barcode: String, scanType: String, projectId: String? = null, notes: String? = null): Result<ScanResult> {
        val request = ScanRequest(
            barcode = barcode,
            scan_type = scanType,
            user_id = "",
            device_id = android.os.Build.SERIAL,
            project_id = projectId,
            notes = notes,
        )
        return try {
            val response = scannerApi.scan(request)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                queueOffline(barcode, scanType, projectId, notes)
                Result.failure(Exception(response.body()?.message ?: "Scan failed"))
            }
        } catch (e: Exception) {
            queueOffline(barcode, scanType, projectId, notes)
            Result.failure(e)
        }
    }

    suspend fun createSession(type: String, projectId: String? = null): Result<ScanSession> {
        return try {
            val response = scannerApi.createSession(SessionCreateRequest(type, projectId))
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Failed to create session"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun endSession(sessionId: String): Result<ScanSession> {
        return try {
            val response = scannerApi.endSession(sessionId)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Failed to end session"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun sessionScan(sessionId: String, barcode: String, scanType: String): Result<ScanResult> {
        val request = ScanRequest(
            barcode = barcode,
            scan_type = scanType,
            user_id = "",
            device_id = android.os.Build.SERIAL,
        )
        return try {
            val response = scannerApi.sessionScan(sessionId, request)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Session scan failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun observePendingCount(): Flow<Int> = pendingScanDao.observePendingCount()
    fun observePendingScans() = pendingScanDao.observeAll()

    suspend fun retryPendingScan(id: Long) {
        pendingScanDao.resetRetry(id)
        SyncWorker.enqueue(context)
    }

    suspend fun deletePendingScan(id: Long) {
        pendingScanDao.deleteById(id)
    }

    suspend fun adHocBooking(equipmentId: String, barcode: String, notes: String? = null): Result<Equipment> {
        return try {
            val response = scannerApi.adHocBooking(AdHocBookingRequest(equipmentId, barcode, notes))
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception(response.body()?.message ?: "Ad-hoc booking failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateEquipmentLocation(equipmentId: String, location: String): Result<Unit> {
        return try {
            val response = scannerApi.updateLocation(equipmentId, mapOf("location" to location))
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                if (tokenManager.getAccessToken() == "demo-access-token") Result.success(Unit)
                else Result.failure(Exception(response.body()?.message ?: "Standort-Update fehlgeschlagen"))
            }
        } catch (e: Exception) {
            if (tokenManager.getAccessToken() == "demo-access-token") Result.success(Unit)
            else Result.failure(e)
        }
    }

    suspend fun pairRfidTag(equipmentId: String, tid: String): Result<Unit> {
        return try {
            val response = scannerApi.pairRfidTag(equipmentId, mapOf("rfid_tid" to tid))
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                if (tokenManager.getAccessToken() == "demo-access-token") Result.success(Unit)
                else Result.failure(Exception("RFID-Zuordnung fehlgeschlagen"))
            }
        } catch (e: Exception) {
            if (tokenManager.getAccessToken() == "demo-access-token") Result.success(Unit)
            else Result.failure(e)
        }
    }

    suspend fun createProject(request: CreateProjectRequest): Result<ProjectCreated> {
        return try {
            val response = scannerApi.createProject(request)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception(response.body()?.message ?: "Projekt konnte nicht erstellt werden"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun searchCustomers(query: String): Result<List<CustomerDto>> {
        return try {
            val response = scannerApi.searchCustomers(query)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.success(emptyList())
            }
        } catch (_: Exception) {
            Result.success(emptyList())
        }
    }

    suspend fun createCustomer(request: CreateCustomerRequest): Result<CustomerDto> {
        return try {
            val response = scannerApi.createCustomer(request)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception(response.body()?.message ?: "Kunde konnte nicht erstellt werden"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getUntaggedEquipment(): Result<List<Equipment>> {
        return try {
            val response = scannerApi.getUntaggedEquipment()
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.success(emptyList())
            }
        } catch (_: Exception) {
            Result.success(emptyList())
        }
    }

    suspend fun listInventoryJobs(): Result<List<com.rentflow.scanner.data.api.InventoryJob>> {
        return try {
            val response = scannerApi.listInventoryJobs()
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.success(emptyList())
            }
        } catch (_: Exception) {
            Result.success(emptyList())
        }
    }

    suspend fun getInventoryJobItems(jobId: String): Result<List<Equipment>> {
        return try {
            val response = scannerApi.getInventoryJobItems(jobId)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.success(emptyList())
            }
        } catch (_: Exception) {
            Result.success(emptyList())
        }
    }

    suspend fun completeInventoryJob(
        jobId: String,
        scannedBarcodes: List<String>,
        missingBarcodes: List<String>,
        unexpectedBarcodes: List<String>,
    ): Result<Unit> {
        return try {
            val response = scannerApi.completeInventoryJob(
                jobId,
                com.rentflow.scanner.data.api.InventoryCompleteRequest(scannedBarcodes, missingBarcodes, unexpectedBarcodes),
            )
            if (response.isSuccessful) Result.success(Unit)
            else Result.failure(Exception("Inventur konnte nicht abgeschlossen werden"))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun uploadSignature(sessionId: String, signatureBase64: String): Result<Unit> {
        return try {
            val response = scannerApi.uploadSignature(sessionId, mapOf("signature" to signatureBase64))
            if (response.isSuccessful) Result.success(Unit)
            else Result.failure(Exception(response.body()?.message ?: "Signature upload failed"))
        } catch (e: Exception) {
            // Non-critical: session can complete without signature upload
            Result.failure(e)
        }
    }

    suspend fun getScanHistory(barcode: String): Result<List<ScanResult>> {
        return try {
            val response = scannerApi.scanHistory(barcode)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.success(emptyList())
            }
        } catch (_: Exception) {
            Result.success(emptyList())
        }
    }

    private suspend fun queueOffline(barcode: String, scanType: String, projectId: String?, notes: String?) {
        pendingScanDao.insert(
            PendingScanEntity(
                barcode = barcode,
                scanType = scanType,
                projectId = projectId,
                notes = notes,
                timestamp = System.currentTimeMillis(),
                userId = "",
                deviceId = android.os.Build.SERIAL,
            )
        )
        SyncWorker.enqueue(context)
    }
}
