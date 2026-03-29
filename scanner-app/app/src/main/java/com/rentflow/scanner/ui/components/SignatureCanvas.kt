package com.rentflow.scanner.ui.components

import android.graphics.Bitmap
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.asAndroidPath
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.gestures.detectDragGestures
import com.rentflow.scanner.R

/**
 * A reusable digital signature canvas composable.
 *
 * Displays a white drawing area where the user can draw a signature with their finger.
 * Includes "Clear" and "Confirm" buttons below the canvas.
 *
 * @param onSignatureComplete Called with the drawn signature as a [Bitmap] when the user
 *   taps the confirm button. Only called if at least one stroke has been drawn.
 * @param modifier Optional modifier for the outer column container.
 */
@Composable
fun SignatureCanvas(
    onSignatureComplete: (Bitmap) -> Unit,
    modifier: Modifier = Modifier,
) {
    val paths = remember { mutableStateListOf<Path>() }
    var currentPath by remember { mutableStateOf<Path?>(null) }
    var hasDrawn by remember { mutableStateOf(false) }

    val density = LocalDensity.current
    val strokeWidthPx = with(density) { 3.dp.toPx() }

    // Track canvas size for bitmap capture
    var canvasWidth by remember { mutableStateOf(0) }
    var canvasHeight by remember { mutableStateOf(0) }

    Column(
        modifier = modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        // Signature drawing area
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .heightIn(min = 200.dp)
                .border(
                    width = 2.dp,
                    color = MaterialTheme.colorScheme.outline,
                    shape = RoundedCornerShape(8.dp),
                )
                .background(
                    color = Color.White,
                    shape = RoundedCornerShape(8.dp),
                )
                .clipToBounds(),
            contentAlignment = Alignment.Center,
        ) {
            // Placeholder text when empty
            if (!hasDrawn) {
                Text(
                    text = stringResource(R.string.signature_hint),
                    style = MaterialTheme.typography.bodyLarge,
                    color = Color.Gray,
                    textAlign = TextAlign.Center,
                )
            }

            Canvas(
                modifier = Modifier
                    .fillMaxSize()
                    .pointerInput(Unit) {
                        detectDragGestures(
                            onDragStart = { offset ->
                                val newPath = Path().apply {
                                    moveTo(offset.x, offset.y)
                                }
                                currentPath = newPath
                            },
                            onDrag = { change, _ ->
                                change.consume()
                                currentPath?.let { path ->
                                    path.lineTo(
                                        change.position.x,
                                        change.position.y,
                                    )
                                    // Force recomposition by replacing the reference
                                    currentPath = Path().apply { addPath(path) }
                                }
                            },
                            onDragEnd = {
                                currentPath?.let { path ->
                                    paths.add(path)
                                    hasDrawn = true
                                }
                                currentPath = null
                            },
                            onDragCancel = {
                                currentPath?.let { path ->
                                    paths.add(path)
                                    hasDrawn = true
                                }
                                currentPath = null
                            },
                        )
                    },
            ) {
                canvasWidth = size.width.toInt()
                canvasHeight = size.height.toInt()

                val stroke = Stroke(
                    width = strokeWidthPx,
                    cap = StrokeCap.Round,
                    join = StrokeJoin.Round,
                )

                // Draw completed paths
                for (path in paths) {
                    drawPath(
                        path = path,
                        color = Color.Black,
                        style = stroke,
                    )
                }

                // Draw current in-progress path
                currentPath?.let { path ->
                    drawPath(
                        path = path,
                        color = Color.Black,
                        style = stroke,
                    )
                }
            }
        }

        // Action buttons
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            OutlinedButton(
                onClick = {
                    paths.clear()
                    currentPath = null
                    hasDrawn = false
                },
                modifier = Modifier.weight(1f),
                enabled = hasDrawn,
            ) {
                Text(stringResource(R.string.signature_clear))
            }

            Button(
                onClick = {
                    if (hasDrawn && canvasWidth > 0 && canvasHeight > 0) {
                        val bitmap = captureSignatureBitmap(
                            paths = paths.toList(),
                            width = canvasWidth,
                            height = canvasHeight,
                            strokeWidthPx = strokeWidthPx,
                        )
                        onSignatureComplete(bitmap)
                    }
                },
                modifier = Modifier.weight(1f),
                enabled = hasDrawn,
            ) {
                Text(stringResource(R.string.signature_confirm))
            }
        }
    }
}

/**
 * Captures the drawn paths as a [Bitmap] with a white background.
 */
private fun captureSignatureBitmap(
    paths: List<Path>,
    width: Int,
    height: Int,
    strokeWidthPx: Float,
): Bitmap {
    val bitmap = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888)
    val canvas = android.graphics.Canvas(bitmap)

    // White background
    canvas.drawColor(android.graphics.Color.WHITE)

    // Draw all paths
    val paint = android.graphics.Paint().apply {
        color = android.graphics.Color.BLACK
        style = android.graphics.Paint.Style.STROKE
        strokeWidth = strokeWidthPx
        strokeCap = android.graphics.Paint.Cap.ROUND
        strokeJoin = android.graphics.Paint.Join.ROUND
        isAntiAlias = true
    }

    for (path in paths) {
        canvas.drawPath(path.asAndroidPath(), paint)
    }

    return bitmap
}
