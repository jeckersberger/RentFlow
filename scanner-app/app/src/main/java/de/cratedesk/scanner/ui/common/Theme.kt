package de.cratedesk.scanner.ui.common

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val DarkColorScheme = darkColorScheme(
    primary = Color(0xFF06B6D4),
    onPrimary = Color.White,
    secondary = Color(0xFFA855F7),
    onSecondary = Color.White,
    background = Color(0xFF0F1117),
    onBackground = Color(0xFFF1F5F9),
    surface = Color(0xFF1A1D2E),
    onSurface = Color(0xFFF1F5F9),
    error = Color(0xFFEF4444),
)

private val LightColorScheme = lightColorScheme(
    primary = Color(0xFF0891B2),
    onPrimary = Color.White,
    secondary = Color(0xFF9333EA),
    onSecondary = Color.White,
)

@Composable
fun CrateDeskTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme
    MaterialTheme(
        colorScheme = colorScheme,
        content = content,
    )
}
