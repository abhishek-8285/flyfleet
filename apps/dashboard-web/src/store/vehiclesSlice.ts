import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import api from '../services/api';
import { Vehicle } from '../types';

interface VehiclesState {
  vehicles: Vehicle[];
  selectedVehicle: Vehicle | null;
  loading: boolean;
  error: string | null;
}

const initialState: VehiclesState = {
  vehicles: [],
  selectedVehicle: null,
  loading: false,
  error: null,
};

export const fetchVehicles = createAsyncThunk('vehicles/fetchAll', async (params: any = {}) => {
  const response = await api.getVehicles(params);
  return response.vehicles || response;
});

export const fetchVehicle = createAsyncThunk('vehicles/fetchOne', async (id: string, { rejectWithValue }) => {
  try {
    return await api.getVehicle(id);
  } catch (error: any) {
    return rejectWithValue(error.response?.data?.message || 'Failed to fetch vehicle');
  }
});

export const createVehicle = createAsyncThunk('vehicles/create', async (data: any, { rejectWithValue }) => {
  try {
    return await api.createVehicle(data);
  } catch (error: any) {
    return rejectWithValue(error.response?.data?.message || 'Failed to create vehicle');
  }
});

export const updateVehicle = createAsyncThunk(
  'vehicles/update',
  async ({ id, data }: { id: string; data: any }, { rejectWithValue }) => {
    try {
      return await api.updateVehicle(id, data);
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || 'Failed to update vehicle');
    }
  }
);

export const deleteVehicle = createAsyncThunk('vehicles/delete', async (id: string, { rejectWithValue }) => {
  try {
    await api.deleteVehicle(id);
    return id;
  } catch (error: any) {
    return rejectWithValue(error.response?.data?.message || 'Failed to delete vehicle');
  }
});

const vehiclesSlice = createSlice({
  name: 'vehicles',
  initialState,
  reducers: {
    setSelectedVehicle: (state, action) => {
      state.selectedVehicle = action.payload;
    },
    updateVehicleLocation: (state, action) => {
      const { vehicle_id, location } = action.payload;
      const vehicle = state.vehicles.find((v) => v.id === vehicle_id);
      if (vehicle) {
        vehicle.current_location = location;
        vehicle.last_seen = location.timestamp;
      }
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch all
      .addCase(fetchVehicles.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchVehicles.fulfilled, (state, action) => {
        state.loading = false;
        state.vehicles = action.payload;
      })
      .addCase(fetchVehicles.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch vehicles';
      })
      // Fetch one
      .addCase(fetchVehicle.fulfilled, (state, action) => {
        state.selectedVehicle = action.payload;
      })
      // Create
      .addCase(createVehicle.fulfilled, (state, action) => {
        state.vehicles.push(action.payload);
      })
      // Update
      .addCase(updateVehicle.fulfilled, (state, action) => {
        const index = state.vehicles.findIndex((v) => v.id === action.payload.id);
        if (index !== -1) {
          state.vehicles[index] = action.payload;
        }
      })
      // Delete
      .addCase(deleteVehicle.fulfilled, (state, action) => {
        state.vehicles = state.vehicles.filter((v) => v.id !== action.payload);
      });
  },
});

export const { setSelectedVehicle, updateVehicleLocation } = vehiclesSlice.actions;
export default vehiclesSlice.reducer;
