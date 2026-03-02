package com.flyfleet.vehicledevice.integration

import com.flyfleet.vehicledevice.data.di.RepositoryModule
import com.flyfleet.vehicledevice.data.repository.MockDriverAuthRepository
import com.flyfleet.vehicledevice.domain.model.RosterDriver
import com.flyfleet.vehicledevice.domain.repository.*
import dagger.Module
import dagger.Provides
import dagger.hilt.components.SingletonComponent
import dagger.hilt.testing.TestInstallIn
import io.mockk.mockk
import javax.inject.Singleton

@Module
@TestInstallIn(
    components = [SingletonComponent::class],
    replaces = [RepositoryModule::class]
)
object TestRepositoryModule {

    @Provides
    @Singleton
    fun provideDeviceRepository(): DeviceRepository = mockk(relaxed = true)

    @Provides
    @Singleton
    fun provideHosRepository(): HosRepository = MockHosRepository()

    @Provides
    @Singleton
    fun providePreferenceRepository(): PreferenceRepository {
        val mock = mockk<PreferenceRepository>(relaxed = true)
        io.mockk.coEvery { mock.fetchAndSyncPreferences(any()) } returns Result.success(
            com.flyfleet.vehicledevice.domain.model.UserPreference("DRV-001")
        )
        return mock
    }

    @Provides
    @Singleton
    fun provideDriverAuthRepository(): DriverAuthRepository {
        return MockDriverAuthRepository(
            mockRoster = listOf(
                RosterDriver("DRV-001", "Marcus Rodriguez", "MR")
            ),
            validPin = "123456",
            validRfid = "04:A2:B3:C4:D5:E6"
        )
    }

    @Provides
    @Singleton
    fun provideSyncRepository(): SyncRepository = mockk(relaxed = true)

    @Provides
    @Singleton
    fun provideSecurityRepository(): SecurityRepository = MockSecurityRepository()
}
