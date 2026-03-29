package com.rentflow.scanner.ui.checkin

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.core.content.FileProvider
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage
import com.rentflow.scanner.R
import com.rentflow.scanner.domain.model.Project
import com.rentflow.scanner.ui.components.SignatureCanvas
import com.rentflow.scanner.ui.theme.Cyan
import com.rentflow.scanner.ui.theme.Error
import com.rentflow.scanner.ui.theme.Success
import com.rentflow.scanner.ui.theme.Warning
import java.io.File

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CheckInScreen(
    onBack: () -> Unit,
    onCompleted: () -> Unit,
    viewModel: CheckInViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()
    val context = LocalContext.current

    LaunchedEffect(state.completed) {
        if (state.completed) onCompleted()
    }

    // Photo capture
    var photoUri by remember { mutableStateOf<Uri?>(null) }
    val cameraLauncher = rememberLauncherForActivityResult(ActivityResultContracts.TakePicture()) { success ->
        viewModel.onPhotoTaken(if (success) photoUri else null)
    }

    // When photoTargetIndex changes, launch camera
    LaunchedEffect(state.photoTargetIndex) {
        if (state.photoTargetIndex >= 0) {
            val file = File(context.cacheDir, "checkin_photo_${System.currentTimeMillis()}.jpg")
            val uri = FileProvider.getUriForFile(context, "${context.packageName}.fileprovider", file)
            photoUri = uri
            cameraLauncher.launch(uri)
        }
    }

    // Signature dialog
    if (state.showSignature) {
        AlertDialog(
            onDismissRequest = viewModel::dismissSignature,
            title = { Text(stringResource(R.string.signature_title)) },
            text = {
                SignatureCanvas(
                    onSignatureComplete = { bitmap -> viewModel.completeWithSignature(bitmap) },
                )
            },
            confirmButton = {},
            dismissButton = {
                OutlinedButton(onClick = viewModel::dismissSignature) {
                    Text(stringResource(R.string.cancel))
                }
            },
        )
    }

    // Auto-detection dialog
    state.autoDetectedProject?.let { project ->
        AlertDialog(
            onDismissRequest = viewModel::dismissAutoDetectedProject,
            icon = {
                Icon(
                    Icons.Default.AssignmentReturn,
                    contentDescription = null,
                    tint = Cyan,
                    modifier = Modifier.size(32.dp),
                )
            },
            title = { Text(stringResource(R.string.checkin_auto_detected, project.name)) },
            confirmButton = {
                Button(onClick = viewModel::confirmAutoDetectedProject) {
                    Text(stringResource(R.string.confirm))
                }
            },
            dismissButton = {
                OutlinedButton(onClick = viewModel::dismissAutoDetectedProject) {
                    Text(stringResource(R.string.cancel))
                }
            },
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.checkin_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
            )
        },
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp)) {
            state.error?.let {
                Text(it, color = MaterialTheme.colorScheme.error)
                Spacer(Modifier.height(8.dp))
            }

            if (state.showSummary) {
                // Summary screen after check-in completion
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Spacer(Modifier.height(32.dp))
                    Icon(
                        Icons.Default.CheckCircle,
                        contentDescription = null,
                        tint = Success,
                        modifier = Modifier.size(64.dp),
                    )
                    Spacer(Modifier.height(16.dp))
                    Text(
                        stringResource(R.string.summary_checkin_done),
                        style = MaterialTheme.typography.headlineSmall,
                        fontWeight = FontWeight.Bold,
                    )
                    Spacer(Modifier.height(8.dp))
                    state.selectedProject?.let {
                        Text(
                            it.name,
                            style = MaterialTheme.typography.titleMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                    Spacer(Modifier.height(16.dp))
                    val okCount = state.scannedItems.count { it.condition == ItemCondition.OK }
                    val damagedCount = state.scannedItems.count { it.condition == ItemCondition.DAMAGED }
                    val defectiveCount = state.scannedItems.count { it.condition == ItemCondition.DEFECTIVE }
                    Text(
                        stringResource(R.string.summary_items_ok, okCount),
                        style = MaterialTheme.typography.bodyLarge,
                        color = Success,
                    )
                    if (damagedCount > 0) {
                        Text(
                            stringResource(R.string.summary_items_damaged, damagedCount),
                            style = MaterialTheme.typography.bodyLarge,
                            color = Warning,
                        )
                    }
                    if (defectiveCount > 0) {
                        Text(
                            stringResource(R.string.summary_items_defective, defectiveCount),
                            style = MaterialTheme.typography.bodyLarge,
                            color = Error,
                        )
                    }
                    val problemItems = state.scannedItems.filter {
                        it.condition == ItemCondition.DAMAGED || it.condition == ItemCondition.DEFECTIVE
                    }
                    if (problemItems.isNotEmpty()) {
                        Spacer(Modifier.height(16.dp))
                        Text(
                            stringResource(R.string.summary_damaged_items),
                            style = MaterialTheme.typography.titleSmall,
                            fontWeight = FontWeight.Bold,
                            color = Warning,
                            modifier = Modifier.fillMaxWidth(),
                        )
                        Spacer(Modifier.height(8.dp))
                        problemItems.forEach { item ->
                            Row(
                                modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Icon(
                                    if (item.condition == ItemCondition.DEFECTIVE) Icons.Default.Error else Icons.Default.Warning,
                                    contentDescription = null,
                                    tint = if (item.condition == ItemCondition.DEFECTIVE) Error else Warning,
                                    modifier = Modifier.size(20.dp),
                                )
                                Spacer(Modifier.width(8.dp))
                                Column {
                                    Text(item.equipment.name, style = MaterialTheme.typography.bodyMedium)
                                    if (item.notes.isNotBlank()) {
                                        Text(
                                            item.notes,
                                            style = MaterialTheme.typography.bodySmall,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                                        )
                                    }
                                }
                            }
                        }
                    }
                    Spacer(Modifier.weight(1f))
                    Button(
                        onClick = viewModel::dismissSummary,
                        modifier = Modifier.fillMaxWidth().height(56.dp),
                    ) {
                        Text(stringResource(R.string.summary_done))
                    }
                }
            } else if (state.selectedProject == null) {
                // Job selection phase
                Text(stringResource(R.string.checkin_returning_today), style = MaterialTheme.typography.titleLarge)
                Spacer(Modifier.height(12.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End,
                ) {
                    FilledTonalButton(onClick = viewModel::toggleShowAll) {
                        Icon(
                            if (state.showAllJobs) Icons.Default.VisibilityOff else Icons.Default.Visibility,
                            contentDescription = null,
                            modifier = Modifier.size(18.dp),
                        )
                        Spacer(Modifier.width(6.dp))
                        Text(
                            if (state.showAllJobs) stringResource(R.string.checkin_today_only)
                            else stringResource(R.string.checkin_all_jobs)
                        )
                    }
                }
                Spacer(Modifier.height(12.dp))

                val jobs = if (state.showAllJobs) state.allCheckedOut else state.returningToday

                if (jobs.isEmpty() && !state.isLoading) {
                    Text(
                        stringResource(R.string.checkin_no_returns_today),
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }

                if (state.isLoading) {
                    CircularProgressIndicator(modifier = Modifier.padding(16.dp).align(Alignment.CenterHorizontally))
                }

                LazyColumn(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(jobs) { project ->
                        JobCard(project = project, onClick = { viewModel.selectProject(project) })
                    }
                }
            } else {
                // Scanning phase
                Text(
                    "${stringResource(R.string.checkin_title)}: ${state.selectedProject!!.name}",
                    style = MaterialTheme.typography.titleLarge,
                )
                Spacer(Modifier.height(8.dp))

                // Progress display
                if (state.expectedReturnCount > 0) {
                    val scanned = state.scannedItems.size
                    val expected = state.expectedReturnCount
                    val progress = scanned.toFloat() / expected
                    Text(
                        stringResource(R.string.checkin_progress, scanned, expected),
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.Bold,
                        color = if (scanned >= expected) Cyan else MaterialTheme.colorScheme.onSurface,
                    )
                    Spacer(Modifier.height(4.dp))
                    LinearProgressIndicator(
                        progress = { progress.coerceIn(0f, 1f) },
                        modifier = Modifier.fillMaxWidth().height(8.dp).clip(RoundedCornerShape(4.dp)),
                        color = if (scanned >= expected) Cyan else MaterialTheme.colorScheme.primary,
                        trackColor = MaterialTheme.colorScheme.surfaceVariant,
                    )
                    Spacer(Modifier.height(8.dp))
                } else {
                    Text(
                        stringResource(R.string.checkin_items_scanned, state.scannedItems.size),
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                Spacer(Modifier.height(16.dp))

                LazyColumn(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    itemsIndexed(state.scannedItems) { index, item ->
                        CheckInItemCard(
                            item = item,
                            onConditionChange = { viewModel.updateCondition(index, it) },
                            onNotesChange = { viewModel.updateNotes(index, it) },
                            onPhoto = { viewModel.requestPhoto(index) },
                            onRemovePhoto = { photoIndex -> viewModel.removePhoto(index, photoIndex) },
                            onRemove = { viewModel.removeItem(index) },
                        )
                    }
                }
                if (state.scannedItems.isNotEmpty()) {
                    Spacer(Modifier.height(16.dp))
                    Button(
                        onClick = viewModel::completeCheckIn,
                        enabled = !state.isLoading,
                        modifier = Modifier.fillMaxWidth().height(56.dp),
                    ) {
                        if (state.isLoading) {
                            CircularProgressIndicator(modifier = Modifier.size(24.dp), color = MaterialTheme.colorScheme.onPrimary)
                        } else {
                            Text("${stringResource(R.string.checkin_confirm)} (${state.scannedItems.size})")
                        }
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun CheckInItemCard(
    item: CheckInItem,
    onConditionChange: (ItemCondition) -> Unit,
    onNotesChange: (String) -> Unit,
    onPhoto: () -> Unit,
    onRemovePhoto: (Int) -> Unit,
    onRemove: () -> Unit,
) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(item.equipment.name, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Text(item.equipment.barcode, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                IconButton(onClick = onRemove) {
                    Icon(Icons.Default.Close, contentDescription = null, tint = MaterialTheme.colorScheme.error)
                }
            }

            Spacer(Modifier.height(8.dp))

            // Condition chips
            Text(stringResource(R.string.checkin_condition), style = MaterialTheme.typography.bodyMedium)
            Spacer(Modifier.height(4.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                FilterChip(
                    selected = item.condition == ItemCondition.OK,
                    onClick = { onConditionChange(ItemCondition.OK) },
                    label = { Text(stringResource(R.string.condition_ok)) },
                    leadingIcon = {
                        Icon(Icons.Default.Check, contentDescription = null, modifier = Modifier.size(18.dp))
                    },
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = Success.copy(alpha = 0.2f),
                        selectedLabelColor = Success,
                        selectedLeadingIconColor = Success,
                    ),
                )
                FilterChip(
                    selected = item.condition == ItemCondition.DAMAGED,
                    onClick = { onConditionChange(ItemCondition.DAMAGED) },
                    label = { Text(stringResource(R.string.condition_damaged)) },
                    leadingIcon = {
                        Icon(Icons.Default.Warning, contentDescription = null, modifier = Modifier.size(18.dp))
                    },
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = Warning.copy(alpha = 0.2f),
                        selectedLabelColor = Warning,
                        selectedLeadingIconColor = Warning,
                    ),
                )
                FilterChip(
                    selected = item.condition == ItemCondition.DEFECTIVE,
                    onClick = { onConditionChange(ItemCondition.DEFECTIVE) },
                    label = { Text(stringResource(R.string.condition_defective)) },
                    leadingIcon = {
                        Icon(Icons.Default.Error, contentDescription = null, modifier = Modifier.size(18.dp))
                    },
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = Error.copy(alpha = 0.2f),
                        selectedLabelColor = Error,
                        selectedLeadingIconColor = Error,
                    ),
                )
            }

            Spacer(Modifier.height(8.dp))

            // Notes
            OutlinedTextField(
                value = item.notes,
                onValueChange = onNotesChange,
                label = { Text(stringResource(R.string.checkin_notes)) },
                modifier = Modifier.fillMaxWidth(),
                maxLines = 2,
            )

            // Photos — only show when condition is DAMAGED or DEFECTIVE
            if (item.condition == ItemCondition.DAMAGED || item.condition == ItemCondition.DEFECTIVE) {
                Spacer(Modifier.height(8.dp))

                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    item.photos.forEachIndexed { photoIndex, uri ->
                        Box(modifier = Modifier.size(48.dp)) {
                            AsyncImage(
                                model = uri,
                                contentDescription = null,
                                modifier = Modifier
                                    .size(48.dp)
                                    .clip(RoundedCornerShape(8.dp))
                                    .border(1.dp, Cyan, RoundedCornerShape(8.dp)),
                                contentScale = ContentScale.Crop,
                            )
                            IconButton(
                                onClick = { onRemovePhoto(photoIndex) },
                                modifier = Modifier
                                    .size(18.dp)
                                    .align(Alignment.TopEnd),
                            ) {
                                Icon(
                                    Icons.Default.Close,
                                    contentDescription = null,
                                    tint = MaterialTheme.colorScheme.error,
                                    modifier = Modifier.size(14.dp),
                                )
                            }
                        }
                    }
                    if (item.photos.size < 5) {
                        OutlinedButton(onClick = onPhoto, modifier = Modifier.height(48.dp)) {
                            Icon(Icons.Default.CameraAlt, contentDescription = null, modifier = Modifier.size(18.dp))
                            Spacer(Modifier.width(6.dp))
                            Text(stringResource(R.string.checkin_add_photo))
                        }
                    } else {
                        Text(
                            stringResource(R.string.checkin_max_photos),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun JobCard(project: Project, onClick: () -> Unit) {
    Card(onClick = onClick, modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(project.name, style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold)
            project.client?.let {
                Text(it, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            project.end_date?.let {
                Text(
                    stringResource(R.string.checkin_return_date, it),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        }
    }
}
