package com.rentflow.scanner.data.util

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.net.Uri
import java.io.File
import java.io.FileOutputStream

object ImageCompressor {
    private const val MAX_DIMENSION = 1920
    private const val QUALITY = 80

    fun compressPhoto(context: Context, sourceUri: Uri): Uri? {
        return try {
            val inputStream = context.contentResolver.openInputStream(sourceUri) ?: return null
            val original = BitmapFactory.decodeStream(inputStream)
            inputStream.close()

            val scaled = scaleDown(original)
            val outputFile = File(context.cacheDir, "compressed_${System.currentTimeMillis()}.jpg")
            FileOutputStream(outputFile).use { out ->
                scaled.compress(Bitmap.CompressFormat.JPEG, QUALITY, out)
            }
            if (scaled != original) scaled.recycle()
            original.recycle()

            Uri.fromFile(outputFile)
        } catch (e: Exception) {
            null
        }
    }

    private fun scaleDown(bitmap: Bitmap): Bitmap {
        val width = bitmap.width
        val height = bitmap.height
        if (width <= MAX_DIMENSION && height <= MAX_DIMENSION) return bitmap

        val ratio = minOf(MAX_DIMENSION.toFloat() / width, MAX_DIMENSION.toFloat() / height)
        val newWidth = (width * ratio).toInt()
        val newHeight = (height * ratio).toInt()
        return Bitmap.createScaledBitmap(bitmap, newWidth, newHeight, true)
    }
}
