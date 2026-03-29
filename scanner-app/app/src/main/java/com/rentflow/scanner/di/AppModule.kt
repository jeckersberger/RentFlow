package com.rentflow.scanner.di

import android.content.Context
import androidx.room.Room
import com.rentflow.scanner.data.api.*
import com.rentflow.scanner.data.db.ScannerDatabase
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.data.preferences.TokenManager
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object AppModule {

    private const val PLACEHOLDER_BASE_URL = "http://localhost/"

    @Provides
    @Singleton
    fun provideDynamicBaseUrlInterceptor(settingsDataStore: SettingsDataStore): DynamicBaseUrlInterceptor {
        return DynamicBaseUrlInterceptor(settingsDataStore)
    }

    @Provides
    @Singleton
    fun provideOkHttpClient(
        dynamicBaseUrlInterceptor: DynamicBaseUrlInterceptor,
        authInterceptor: AuthInterceptor,
        tokenManager: TokenManager,
        settingsDataStore: SettingsDataStore,
    ): OkHttpClient {
        val logging = HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        }
        return OkHttpClient.Builder()
            .addInterceptor(dynamicBaseUrlInterceptor)
            .addInterceptor(authInterceptor)
            .addInterceptor(logging)
            .authenticator(TokenAuthenticator(tokenManager, settingsDataStore))
            .connectTimeout(15, TimeUnit.SECONDS)
            .readTimeout(15, TimeUnit.SECONDS)
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(client: OkHttpClient): Retrofit {
        return Retrofit.Builder()
            .baseUrl(PLACEHOLDER_BASE_URL)
            .client(client)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    @Provides
    @Singleton
    fun provideAuthApi(retrofit: Retrofit): AuthApi = retrofit.create(AuthApi::class.java)

    @Provides
    @Singleton
    fun provideScannerApi(retrofit: Retrofit): ScannerApi = retrofit.create(ScannerApi::class.java)

    @Provides
    @Singleton
    fun provideProjectApi(retrofit: Retrofit): ProjectApi = retrofit.create(ProjectApi::class.java)

    @Provides
    @Singleton
    fun provideWarehouseApi(retrofit: Retrofit): WarehouseApi = retrofit.create(WarehouseApi::class.java)

    @Provides
    @Singleton
    fun provideConfigApi(retrofit: Retrofit): ConfigApi = retrofit.create(ConfigApi::class.java)

    @Provides
    @Singleton
    fun provideDatabase(@ApplicationContext context: Context): ScannerDatabase {
        return Room.databaseBuilder(context, ScannerDatabase::class.java, "scanner_db").build()
    }

    @Provides
    fun providePendingScanDao(db: ScannerDatabase) = db.pendingScanDao()
}
