package com.rentflow.scanner.ui.components

import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.rentflow.scanner.R

@Composable
fun PendingQueueBadge(count: Int) {
    if (count > 0) {
        Badge(containerColor = MaterialTheme.colorScheme.error) {
            Text(stringResource(R.string.pending_scans, count))
        }
    }
}
