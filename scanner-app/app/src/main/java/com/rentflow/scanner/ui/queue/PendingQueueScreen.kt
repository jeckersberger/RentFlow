package com.rentflow.scanner.ui.queue

import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.data.db.PendingScanEntity
import com.rentflow.scanner.ui.theme.*
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PendingQueueScreen(
    onBack: () -> Unit,
    viewModel: PendingQueueViewModel = hiltViewModel(),
) {
    val scans by viewModel.pendingScans.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Offline-Warteschlange") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
                actions = {
                    if (scans.isNotEmpty()) {
                        Badge(
                            containerColor = Cyan,
                            contentColor = DarkBackground,
                            modifier = Modifier.padding(end = 16.dp),
                        ) {
                            Text("${scans.size}")
                        }
                    }
                },
            )
        },
        bottomBar = {
            if (scans.isNotEmpty()) {
                Surface(tonalElevation = 3.dp) {
                    Button(
                        onClick = { viewModel.retryAll() },
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(16.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = Cyan),
                    ) {
                        Icon(
                            Icons.Default.Replay,
                            contentDescription = null,
                            modifier = Modifier.padding(end = 8.dp),
                        )
                        Text("Alle erneut versuchen", color = DarkBackground)
                    }
                }
            }
        },
    ) { padding ->
        if (scans.isEmpty()) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Icon(
                        Icons.Default.CloudDone,
                        contentDescription = null,
                        modifier = Modifier.size(72.dp),
                        tint = Success,
                    )
                    Spacer(Modifier.height(16.dp))
                    Text(
                        "Keine ausstehenden Scans",
                        style = MaterialTheme.typography.titleMedium,
                        color = TextSecondary,
                    )
                }
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                items(scans, key = { it.id }) { scan ->
                    PendingScanItem(
                        scan = scan,
                        onRetry = { viewModel.retryItem(scan.id) },
                        onDelete = { viewModel.deleteItem(scan.id) },
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun PendingScanItem(
    scan: PendingScanEntity,
    onRetry: () -> Unit,
    onDelete: () -> Unit,
) {
    val isFailed = scan.retryCount >= 5 || scan.failed
    val dismissState = rememberSwipeToDismissBoxState(
        confirmValueChange = { value ->
            if (value == SwipeToDismissBoxValue.EndToStart) {
                onDelete()
                true
            } else {
                false
            }
        },
    )

    SwipeToDismissBox(
        state = dismissState,
        backgroundContent = {
            val color by animateColorAsState(
                when (dismissState.targetValue) {
                    SwipeToDismissBoxValue.EndToStart -> Error
                    else -> Color.Transparent
                },
                label = "swipeBg",
            )
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(color, MaterialTheme.shapes.medium)
                    .padding(horizontal = 20.dp),
                contentAlignment = Alignment.CenterEnd,
            ) {
                Icon(
                    Icons.Default.Delete,
                    contentDescription = "Loeschen",
                    tint = Color.White,
                )
            }
        },
        enableDismissFromStartToEnd = false,
    ) {
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = if (isFailed) Error.copy(alpha = 0.12f) else CardBackground,
            ),
            border = if (isFailed) {
                CardDefaults.outlinedCardBorder().copy(
                    width = 1.dp,
                    brush = androidx.compose.ui.graphics.SolidColor(Error),
                )
            } else {
                null
            },
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    // Barcode
                    Text(
                        scan.barcode,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        fontFamily = FontFamily.Monospace,
                        color = if (isFailed) Error else TextPrimary,
                    )
                    // Status chip
                    val statusLabel = if (isFailed) "Fehlgeschlagen" else "Ausstehend"
                    val statusColor = if (isFailed) Error else Warning
                    Surface(
                        color = statusColor.copy(alpha = 0.15f),
                        shape = MaterialTheme.shapes.small,
                    ) {
                        Text(
                            statusLabel,
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                            style = MaterialTheme.typography.labelSmall,
                            color = statusColor,
                            fontWeight = FontWeight.SemiBold,
                        )
                    }
                }

                Spacer(Modifier.height(8.dp))

                // Scan type & timestamp row
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            Icons.Default.QrCodeScanner,
                            contentDescription = null,
                            modifier = Modifier.size(16.dp),
                            tint = TextSecondary,
                        )
                        Spacer(Modifier.width(4.dp))
                        Text(
                            scan.scanType,
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary,
                        )
                    }
                    Text(
                        formatTimestamp(scan.timestamp),
                        style = MaterialTheme.typography.bodySmall,
                        color = TextSecondary,
                    )
                }

                // Retry count
                if (scan.retryCount > 0) {
                    Spacer(Modifier.height(4.dp))
                    Text(
                        "Versuche: ${scan.retryCount}",
                        style = MaterialTheme.typography.bodySmall,
                        color = if (isFailed) Error else Warning,
                    )
                }

                Spacer(Modifier.height(12.dp))

                // Action buttons
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    OutlinedButton(
                        onClick = onDelete,
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = Error),
                    ) {
                        Icon(
                            Icons.Default.Delete,
                            contentDescription = null,
                            modifier = Modifier.size(18.dp),
                        )
                        Spacer(Modifier.width(4.dp))
                        Text("Loeschen")
                    }
                    Spacer(Modifier.width(8.dp))
                    Button(
                        onClick = onRetry,
                        colors = ButtonDefaults.buttonColors(containerColor = Cyan),
                    ) {
                        Icon(
                            Icons.Default.Replay,
                            contentDescription = null,
                            modifier = Modifier.size(18.dp),
                            tint = DarkBackground,
                        )
                        Spacer(Modifier.width(4.dp))
                        Text("Wiederholen", color = DarkBackground)
                    }
                }
            }
        }
    }
}

private fun formatTimestamp(millis: Long): String {
    val sdf = SimpleDateFormat("dd.MM.yyyy HH:mm", Locale.GERMANY)
    return sdf.format(Date(millis))
}
