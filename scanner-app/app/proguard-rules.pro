# Retrofit
-keepattributes Signature
-keepattributes *Annotation*
-keep class com.rentflow.scanner.data.api.** { *; }
-keep class com.rentflow.scanner.domain.model.** { *; }

# Gson
-keep class com.google.gson.** { *; }
-keepclassmembers,allowobfuscation class * {
    @com.google.gson.annotations.SerializedName <fields>;
}

# Room
-keep class * extends androidx.room.RoomDatabase
-keep @androidx.room.Entity class * { *; }
-keep @androidx.room.Dao interface * { *; }
-keep class com.rentflow.scanner.data.db.** { *; }

# Hilt / Dagger
-dontwarn dagger.hilt.**
-keep class dagger.hilt.** { *; }
-keep class * extends dagger.hilt.android.internal.managers.ViewComponentManager$FragmentContextWrapper { *; }
-keep,allowobfuscation,allowshrinking @dagger.hilt.android.EarlyEntryPoint class *
-keep,allowobfuscation,allowshrinking @dagger.hilt.EntryPoint class *

# WorkManager + HiltWorker
-keep class * extends androidx.work.Worker
-keep class * extends androidx.work.ListenableWorker
-keep class com.rentflow.scanner.worker.** { *; }
-keep class androidx.hilt.work.** { *; }

# OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**

# Kotlin Coroutines
-keepnames class kotlinx.coroutines.internal.MainDispatcherFactory {}
-keepnames class kotlinx.coroutines.CoroutineExceptionHandler {}

# Enum values (used by Gson deserialization)
-keepclassmembers enum * { *; }
