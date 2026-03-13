import { configureStore } from "@reduxjs/toolkit";
import alertsReducer from "./alertsSlice";
import analyticsReducer from "./analyticsSlice";
import authReducer from "./authSlice";
import dashboardReducer from "./dashboardSlice";
import geofencesReducer from "./geofencesSlice";
import telemetryReducer from "./telemetrySlice";
import timelineReducer from "./timelineSlice";
import tripsReducer from "./tripsSlice";
import vehiclesReducer from "./vehiclesSlice";

export const store = configureStore({
  reducer: {
    auth: authReducer,
    vehicles: vehiclesReducer,
    telemetry: telemetryReducer,
    trips: tripsReducer,
    alerts: alertsReducer,
    geofences: geofencesReducer,
    dashboard: dashboardReducer,
    analytics: analyticsReducer,
    timeline: timelineReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
