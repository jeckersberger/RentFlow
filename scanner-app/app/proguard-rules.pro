# Retrofit
-keep class de.cratedesk.scanner.data.api.dto.** { *; }
-keepattributes Signature
-keepattributes *Annotation*
-dontwarn retrofit2.**
-keep class retrofit2.** { *; }

# Gson
-keepattributes SerializedName
-keep class com.google.gson.** { *; }

# Hilt
-dontwarn dagger.hilt.**
