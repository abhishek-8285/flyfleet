import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import api from '../services/api';
import { Alert } from '../types';

interface AlertsState {
  alerts: Alert[];
  unreadCount: number;
  loading: boolean;
  error: string | null;
}

const initialState: AlertsState = {
  alerts: [],
  unreadCount: 0,
  loading: false,
  error: null,
};

export const fetchAlerts = createAsyncThunk('alerts/fetchAll', async (params: any = {}) => {
  const response = await api.getAlerts(params);
  return response.alerts || response;
});

export const acknowledgeAlert = createAsyncThunk('alerts/acknowledge', async (id: string, { rejectWithValue }) => {
  try {
    await api.acknowledgeAlert(id);
    return id;
  } catch (error: any) {
    return rejectWithValue(error.response?.data?.message || 'Failed to acknowledge alert');
  }
});

const alertsSlice = createSlice({
  name: 'alerts',
  initialState,
  reducers: {
    addAlert: (state, action) => {
      state.alerts.unshift(action.payload);
      if (!action.payload.acknowledged) {
        state.unreadCount += 1;
      }
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchAlerts.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchAlerts.fulfilled, (state, action) => {
        state.loading = false;
        state.alerts = action.payload;
        state.unreadCount = action.payload.filter((a: Alert) => !a.acknowledged).length;
      })
      .addCase(fetchAlerts.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch alerts';
      })
      .addCase(acknowledgeAlert.fulfilled, (state, action) => {
        const alert = state.alerts.find((a) => a.id === action.payload);
        if (alert && !alert.acknowledged) {
          alert.acknowledged = true;
          state.unreadCount -= 1;
        }
      });
  },
});

export const { addAlert } = alertsSlice.actions;
export default alertsSlice.reducer;
