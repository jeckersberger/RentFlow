package de.cratedesk.scanner.ui.scan

import android.annotation.SuppressLint
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.ImageProxy
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.google.mlkit.vision.barcode.common.Barcode
import com.google.mlkit.vision.common.InputImage

/**
 * CameraX ImageAnalyzer that detects barcodes/QR codes using ML Kit.
 * Calls [onBarcodeDetected] with the raw barcode value on first detection.
 * Throttles to prevent duplicate callbacks.
 */
class BarcodeAnalyzer(
    private val onBarcodeDetected: (String, Int) -> Unit,
) : ImageAnalysis.Analyzer {

    private val scanner = BarcodeScanning.getClient()
    private var lastDetectedValue: String? = null
    private var lastDetectedTime: Long = 0

    @SuppressLint("UnsafeOptInUsageError")
    override fun analyze(imageProxy: ImageProxy) {
        val mediaImage = imageProxy.image ?: run {
            imageProxy.close()
            return
        }

        val inputImage = InputImage.fromMediaImage(
            mediaImage,
            imageProxy.imageInfo.rotationDegrees,
        )

        scanner.process(inputImage)
            .addOnSuccessListener { barcodes ->
                for (barcode in barcodes) {
                    val rawValue = barcode.rawValue ?: continue
                    val now = System.currentTimeMillis()

                    // Throttle: skip if same barcode detected within 2 seconds
                    if (rawValue == lastDetectedValue && now - lastDetectedTime < 2000) {
                        continue
                    }

                    lastDetectedValue = rawValue
                    lastDetectedTime = now
                    onBarcodeDetected(rawValue, barcode.format)
                }
            }
            .addOnCompleteListener {
                imageProxy.close()
            }
    }
}
