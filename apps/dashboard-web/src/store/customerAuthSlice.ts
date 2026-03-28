import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { AppDispatch } from './index';

interface CustomerAuthState {
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

const initialState: CustomerAuthState = {
  token: null, // SECURITY: Stored in memory only, NOT localStorage
  isAuthenticated: false,
  loading: false,
  error: null,
};

const customerAuthSlice = createSlice({
  name: 'customerAuth',
  initialState,
  reducers: {
    setCustomerToken: (state, action: PayloadAction<string | null>) => {
      state.token = action.payload;
      state.isAuthenticated = !!action.payload;
    },
    clearCustomerAuth: (state) => {
      state.token = null;
      state.isAuthenticated = false;
      state.error = null;
    },
    setCustomerError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
  extraReducers: (builder) => {
    // Add async thunks if needed
  },
});

export const { setCustomerToken, clearCustomerAuth, setCustomerError } = customerAuthSlice.actions;
export default customerAuthSlice.reducer;
