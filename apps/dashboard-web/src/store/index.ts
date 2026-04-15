import { configureStore } from "@reduxjs/toolkit";
import alertsReducer from "./alertsSlice";
import analyticsReducer from "./analyticsSlice";
import authReducer from "./authSlice";
import customerAuthReducer from "./customerAuthSlice";
import dashboardReducer from "./dashboardSlice";
import geofencesReducer from "./geofencesSlice";
import telemetryReducer from "./telemetrySlice";
import timelineReducer from "./timelineSlice";
import trackingReducer from "../features/tracking/store/trackingSlice";
import tripsReducer from "./tripsSlice";
import vehiclesReducer from "./vehiclesSlice";

export const store = configureStore({
  reducer: {
    auth: authReducer,
    customerAuth: customerAuthReducer,
    vehicles: vehiclesReducer,
    telemetry: telemetryReducer,
    trips: tripsReducer,
    alerts: alertsReducer,
    geofences: geofencesReducer,
    dashboard: dashboardReducer,
    analytics: analyticsReducer,
    timeline: timelineReducer,
    tracking: trackingReducer,
  },
});

// Expose store globally for API service (breaks circular dependency with api.ts)
if (typeof window !== 'undefined') {
  (window as any).__REDUX_STORE__ = store;
}

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
