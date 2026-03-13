import { createSlice, PayloadAction } from "@reduxjs/toolkit";

interface AnalyticsState {
  fleetEfficiency: any[];
  tripTrends: any[];
  driverPerformance: any[];
  predictiveTrips: any[];
  loading: boolean;
  error: string | null;
}

const initialState: AnalyticsState = {
  fleetEfficiency: [],
  tripTrends: [],
  driverPerformance: [],
  predictiveTrips: [],
  loading: false,
  error: null,
};

const analyticsSlice = createSlice({
  name: "analytics",
  initialState,
  reducers: {
    setAnalyticsLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setFleetEfficiency: (state, action: PayloadAction<any[]>) => {
      state.fleetEfficiency = action.payload;
    },
    setTripTrends: (state, action: PayloadAction<any[]>) => {
      state.tripTrends = action.payload;
    },
    setDriverPerformance: (state, action: PayloadAction<any[]>) => {
      state.driverPerformance = action.payload;
    },
    setPredictiveTrips: (state, action: PayloadAction<any[]>) => {
      state.predictiveTrips = action.payload;
    },
    setAnalyticsError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

export const {
  setAnalyticsLoading,
  setFleetEfficiency,
  setTripTrends,
  setDriverPerformance,
  setPredictiveTrips,
  setAnalyticsError,
} = analyticsSlice.actions;

export default analyticsSlice.reducer;
