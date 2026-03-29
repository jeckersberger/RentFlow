package com.rentflow.scanner.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.SystemUpdate
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.rentflow.scanner.R
import com.rentflow.scanner.data.service.UpdateInfo
import com.rentflow.scanner.ui.theme.Cyan

@Composable
fun UpdateBanner(
    updateInfo: UpdateInfo,
    onUpdate: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Card(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp)
            .clickable { onUpdate() },
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = Cyan.copy(alpha = 0.15f)),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                Icons.Default.SystemUpdate,
                contentDescription = null,
                tint = Cyan,
                modifier = Modifier.size(32.dp),
            )
            Spacer(Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    stringResource(R.string.update_available, updateInfo.versionName),
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.Bold,
                    color = Cyan,
                )
                if (updateInfo.size > 0) {
                    Text(
                        "%.1f MB".format(updateInfo.size / 1_048_576.0),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                    )
                }
            }
            Button(
                onClick = onUpdate,
                colors = ButtonDefaults.buttonColors(containerColor = Cyan),
            ) {
                Text(stringResource(R.string.update_button), color = MaterialTheme.colorScheme.surface)
            }
        }
    }
}
