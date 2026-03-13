import { createSlice, PayloadAction } from "@reduxjs/toolkit";

interface DashboardStats {
  total_vehicles: number;
  active_trips: number;
  active_drivers: number;
  alerts_count: number;
  fleet_health_score: number;
  vehicles_offline: number;
  sla_violations: number;
}

interface DashboardState {
  stats: DashboardStats | null;
  loading: boolean;
  error: string | null;
  lastUpdated: number | null;
}

const initialState: DashboardState = {
  stats: null,
  loading: false,
  error: null,
  lastUpdated: null,
};

const dashboardSlice = createSlice({
  name: "dashboard",
  initialState,
  reducers: {
    fetchStatsStart: (state) => {
      state.loading = true;
      state.error = null;
    },
    fetchStatsSuccess: (state, action: PayloadAction<DashboardStats>) => {
      state.loading = false;
      state.stats = action.payload;
      state.lastUpdated = Date.now();
    },
    fetchStatsFailure: (state, action: PayloadAction<string>) => {
      state.loading = false;
      state.error = action.payload;
    },
    updateKpi: (state, action: PayloadAction<Partial<DashboardStats>>) => {
      if (state.stats) {
        state.stats = { ...state.stats, ...action.payload };
      }
    },
    incrementAlertCount: (state) => {
      if (state.stats) {
        state.stats.alerts_count += 1;
      }
    },
  },
});

export const {
  fetchStatsStart,
  fetchStatsSuccess,
  fetchStatsFailure,
  updateKpi,
  incrementAlertCount,
} = dashboardSlice.actions;
export default dashboardSlice.reducer;
