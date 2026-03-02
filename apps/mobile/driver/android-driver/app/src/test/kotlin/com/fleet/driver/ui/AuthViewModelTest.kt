package com.fleet.driver.ui

import com.fleet.driver.data.remote.model.AuthState
import com.fleet.driver.data.remote.model.UserResponse
import com.fleet.driver.data.repository.AuthRepository
import com.fleet.driver.presentation.auth.viewmodel.AuthViewModel
import com.fleet.driver.util.Resource
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class AuthViewModelTest {

    private lateinit var viewModel: AuthViewModel
    private val mockAuthRepository = mockk<AuthRepository>(relaxed = true)
    private val testDispatcher = UnconfinedTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        viewModel = AuthViewModel(mockAuthRepository)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial state is Unauthenticated if not logged in`() {
        every { mockAuthRepository.isAuthenticated() } returns false
        
        // Re-init to trigger checkInitialAuthState with mock result
        viewModel = AuthViewModel(mockAuthRepository)
        
        assertEquals(AuthState.Unauthenticated, viewModel.authState.value)
    }

    @Test
    fun `initial state is Authenticated if already logged in`() {
        val user = UserResponse("user-1", "test@example.com", "driver", emptyList(), "org-1")
        every { mockAuthRepository.isAuthenticated() } returns true
        every { mockAuthRepository.getCurrentUser() } returns user
        
        // Re-init
        viewModel = AuthViewModel(mockAuthRepository)
        
        val state = viewModel.authState.value
        assertTrue(state is AuthState.Authenticated)
        assertEquals(user, (state as AuthState.Authenticated).user)
    }

    @Test
    fun `login success updates state to Authenticated`() = runTest {
        val user = UserResponse("user-1", "test@example.com", "driver", emptyList(), "org-1")
        val loginResponse = mockk<com.fleet.driver.data.remote.model.LoginResponse>()
        every { loginResponse.user } returns user
        
        coEvery { mockAuthRepository.login(any(), any()) } returns Resource.Success(loginResponse)
        
        viewModel.login("test@example.com", "password")
        
        val state = viewModel.authState.value
        assertTrue(state is AuthState.Authenticated)
        assertEquals(user, (state as AuthState.Authenticated).user)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `login failure updates state to Error`() = runTest {
        val errorMessage = "Invalid credentials"
        coEvery { mockAuthRepository.login(any(), any()) } returns Resource.Error(errorMessage)
        
        viewModel.login("test@example.com", "wrong")
        
        val state = viewModel.authState.value
        assertTrue(state is AuthState.Error)
        assertEquals(errorMessage, (state as AuthState.Error).message)
        assertEquals(errorMessage, viewModel.errorMessage.value)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `logout updates state to Unauthenticated`() = runTest {
        coEvery { mockAuthRepository.logout() } returns Resource.Success(Unit)
        
        viewModel.logout()
        
        assertEquals(AuthState.Unauthenticated, viewModel.authState.value)
        assertFalse(viewModel.isLoading.value)
    }
}
