# Retrofit
-keepattributes Signature
-keepattributes *Annotation*
-keep class com.rentflow.scanner.data.api.** { *; }
-keep class com.rentflow.scanner.domain.model.** { *; }

# Gson
-keep class com.google.gson.** { *; }

# Room
-keep class * extends androidx.room.RoomDatabase
