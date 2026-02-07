import { configureStore } from '@reduxjs/toolkit';
import alertsReducer from './alertsSlice';
import authReducer from './authSlice';
import geofencesReducer from './geofencesSlice';
import tripsReducer from './tripsSlice';
import vehiclesReducer from './vehiclesSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    vehicles: vehiclesReducer,
    trips: tripsReducer,
    alerts: alertsReducer,
    geofences: geofencesReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
