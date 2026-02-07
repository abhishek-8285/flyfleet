import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import api from '../services/api';
import { Trip } from '../types';

interface TripsState {
  trips: Trip[];
  selectedTrip: Trip | null;
  loading: boolean;
  error: string | null;
}

const initialState: TripsState = {
  trips: [],
  selectedTrip: null,
  loading: false,
  error: null,
};

export const fetchTrips = createAsyncThunk('trips/fetchAll', async (params: any = {}) => {
  const response = await api.getTrips(params);
  return response.trips || response;
});

export const fetchTrip = createAsyncThunk('trips/fetchOne', async (id: string, { rejectWithValue }) => {
  try {
    return await api.getTrip(id);
  } catch (error: any) {
    return rejectWithValue(error.response?.data?.message || 'Failed to fetch trip');
  }
});

export const createTrip = createAsyncThunk('trips/create', async (data: any, { rejectWithValue }) => {
  try {
    return await api.createTrip(data);
  } catch (error: any) {
    return rejectWithValue(error.response?.data?.message || 'Failed to create trip');
  }
});

const tripsSlice = createSlice({
  name: 'trips',
  initialState,
  reducers: {
    setSelectedTrip: (state, action) => {
      state.selectedTrip = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchTrips.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchTrips.fulfilled, (state, action) => {
        state.loading = false;
        state.trips = action.payload;
      })
      .addCase(fetchTrips.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch trips';
      })
      .addCase(fetchTrip.fulfilled, (state, action) => {
        state.selectedTrip = action.payload;
      })
      .addCase(createTrip.fulfilled, (state, action) => {
        state.trips.push(action.payload);
      });
  },
});

export const { setSelectedTrip } = tripsSlice.actions;
export default tripsSlice.reducer;
