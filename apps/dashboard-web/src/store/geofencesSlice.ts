import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import api from '../services/api';
import { Geofence } from '../types';

interface GeofencesState {
  geofences: Geofence[];
  loading: boolean;
  error: string | null;
}

const initialState: GeofencesState = {
  geofences: [],
  loading: false,
  error: null,
};

export const fetchGeofences = createAsyncThunk('geofences/fetchAll', async () => {
  const response = await api.getGeofences();
  return response.geofences || response;
});

export const createGeofence = createAsyncThunk('geofences/create', async (data: any) => {
  return await api.createGeofence(data);
});

export const deleteGeofence = createAsyncThunk('geofences/delete', async (id: string) => {
  await api.deleteGeofence(id);
  return id;
});

const geofencesSlice = createSlice({
  name: 'geofences',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(fetchGeofences.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchGeofences.fulfilled, (state, action) => {
        state.loading = false;
        state.geofences = action.payload;
      })
      .addCase(fetchGeofences.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch geofences';
      })
      .addCase(createGeofence.fulfilled, (state, action) => {
        state.geofences.push(action.payload);
      })
      .addCase(deleteGeofence.fulfilled, (state, action) => {
        state.geofences = state.geofences.filter((g) => g.id !== action.payload);
      });
  },
});

export default geofencesSlice.reducer;
