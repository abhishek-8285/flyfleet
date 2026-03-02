package com.flyfleet.vehicledevice.di

import android.content.Context
import androidx.room.Room
import com.flyfleet.vehicledevice.data.db.AppDatabase
import com.flyfleet.vehicledevice.data.db.SyncDao
import com.flyfleet.vehicledevice.data.db.DashboardDao
import com.flyfleet.vehicledevice.data.db.UserPreferenceDao
import com.flyfleet.vehicledevice.data.db.HosDao
import com.flyfleet.vehicledevice.data.di.DatabaseModule
import com.flyfleet.vehicledevice.data.security.SecurityModule
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import dagger.hilt.testing.TestInstallIn
import io.mockk.every
import io.mockk.mockk
import javax.inject.Singleton

@Module
@TestInstallIn(
    components = [SingletonComponent::class],
    replaces = [DatabaseModule::class]
)
object TestDatabaseModule {

    @Provides
    @Singleton
    fun provideSecurityModule(@ApplicationContext context: Context): SecurityModule {
        val mock = mockk<SecurityModule>(relaxed = true)
        every { mock.getDatabasePassphrase() } returns "mock-pass".toByteArray()
        every { mock.getSecureString(any(), any()) } answers { it.invocation.args[1] as String }
        return mock
    }

    @Provides
    @Singleton
    fun provideAppDatabase(@ApplicationContext context: Context, securityModule: SecurityModule): AppDatabase {
        // Use in-memory database without SQLCipher factory for Robolectric
        return Room.inMemoryDatabaseBuilder(context, AppDatabase::class.java)
            .allowMainThreadQueries()
            .build()
    }

    @Provides
    @Singleton
    fun provideSyncDao(database: AppDatabase): SyncDao {
        return database.syncDao()
    }

    @Provides
    @Singleton
    fun provideDashboardDao(database: AppDatabase): DashboardDao {
        return database.dashboardDao()
    }

    @Provides
    @Singleton
    fun provideUserPreferenceDao(database: AppDatabase): UserPreferenceDao {
        return database.userPreferenceDao()
    }

    @Provides
    @Singleton
    fun provideHosDao(database: AppDatabase): HosDao {
        return database.hosDao()
    }
}
