package com.rentflow.scanner.ui.findme

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.VolumeOff
import androidx.compose.material.icons.filled.VolumeUp
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.rentflow.scanner.ui.theme.Cyan
import com.rentflow.scanner.ui.theme.Error

@Composable
fun FindMeOverlay(
    onStop: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Error.copy(alpha = 0.95f)),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center,
        ) {
            Icon(
                Icons.Default.VolumeUp,
                contentDescription = null,
                modifier = Modifier.size(120.dp),
                tint = MaterialTheme.colorScheme.onError,
            )
            Spacer(Modifier.height(24.dp))
            Text(
                "SCANNER GESUCHT",
                style = MaterialTheme.typography.headlineLarge,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onError,
                textAlign = TextAlign.Center,
            )
            Spacer(Modifier.height(8.dp))
            Text(
                "Jemand sucht diesen Scanner",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onError.copy(alpha = 0.8f),
            )
            Spacer(Modifier.height(48.dp))
            Button(
                onClick = onStop,
                modifier = Modifier
                    .fillMaxWidth(0.7f)
                    .height(72.dp),
                colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.onError),
            ) {
                Icon(Icons.Default.VolumeOff, contentDescription = null, modifier = Modifier.size(28.dp), tint = Error)
                Spacer(Modifier.width(12.dp))
                Text("STOPP", fontSize = 24.sp, fontWeight = FontWeight.Bold, color = Error)
            }
        }
    }
}
