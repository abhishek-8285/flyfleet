import { createAsyncThunk, createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface Alert {
  id: string;
  type: string;
  vehicleId: string;
  severity: "critical" | "warning" | "info";
  message: string;
  timestamp: string;
  acknowledged: boolean;
  metadata?: any;
}

export const fetchAlerts = createAsyncThunk(
  "alerts/fetchAlerts",
  async (_: any) => {
    const response = await fetch("/api/v1/alerts");
    return (await response.json()) as Alert[];
  },
);

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

const alertsSlice = createSlice({
  name: "alerts",
  initialState,
  reducers: {
    setAlerts: (state, action: PayloadAction<Alert[]>) => {
      state.alerts = action.payload;
      state.unreadCount = action.payload.filter((a) => !a.acknowledged).length;
      state.loading = false;
    },
    addAlertEvent: (state, action: PayloadAction<Alert>) => {
      state.alerts.unshift(action.payload);
      if (!action.payload.acknowledged) {
        state.unreadCount += 1;
      }
      if (state.alerts.length > 500) state.alerts.pop();
    },
    acknowledgeAlert: (state, action: PayloadAction<string>) => {
      const alert = state.alerts.find((a) => a.id === action.payload);
      if (alert && !alert.acknowledged) {
        alert.acknowledged = true;
        state.unreadCount = Math.max(0, state.unreadCount - 1);
      }
    },
    setAlertsLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setAlertsError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
      state.loading = false;
    },
  },
});

export const {
  setAlerts,
  addAlertEvent,
  acknowledgeAlert,
  setAlertsLoading,
  setAlertsError,
} = alertsSlice.actions;

export default alertsSlice.reducer;
