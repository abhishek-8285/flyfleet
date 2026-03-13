import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface VehicleTelemetry {
  id: string;
  latitude: number;
  longitude: number;
  speed: number;
  heading: number;
  lastUpdate: string;
  status: "moving" | "idling" | "alert" | "offline";
  driverId?: string;
  tripId?: string;
  index?: number; // Pointer to SharedArrayBuffer slot
}

interface TelemetryState {
  vehicles: Record<string, VehicleTelemetry>;
  breadcrumbs: Record<string, number[][]>; // vehicleId -> [lng, lat][]
  indices: Record<string, number>; // Fast ID -> Index lookup
  lastUpdateBatch: string;
  count: number;
  selectedVehicleId: string | null;
}

const initialState: TelemetryState = {
  vehicles: {},
  breadcrumbs: {},
  indices: {},
  lastUpdateBatch: new Date().toISOString(),
  count: 0,
  selectedVehicleId: null,
};

const telemetrySlice = createSlice({
  name: "telemetry",
  initialState,
  reducers: {
    updateTelemetryBatch: (
      state,
      action: PayloadAction<VehicleTelemetry[]>,
    ) => {
      action.payload.forEach((update) => {
        state.vehicles[update.id] = {
          ...state.vehicles[update.id],
          ...update,
        };
        if (update.index !== undefined) {
          state.indices[update.id] = update.index;
        }
      });
      state.lastUpdateBatch = new Date().toISOString();
      state.count = Object.keys(state.vehicles).length;
    },
    updateBreadcrumbs: (
      state,
      action: PayloadAction<{ vehicleId: string; breadcrumbs: number[][] }>,
    ) => {
      // Cap breadcrumbs to 100 points to prevent memory bloat
      state.breadcrumbs[action.payload.vehicleId] =
        action.payload.breadcrumbs.slice(-100);
    },
    clearTelemetry: (state) => {
      state.vehicles = {};
      state.breadcrumbs = {};
      state.indices = {};
      state.count = 0;
      state.selectedVehicleId = null;
    },
    updateVehicle: (
      state,
      action: PayloadAction<
        Partial<VehicleTelemetry> & {
          vehicleId?: string;
          lat?: number;
          lng?: number;
        }
      >,
    ) => {
      const update = action.payload;
      const id = update.vehicleId || update.id;
      if (!id) return;

      state.vehicles[id] = {
        ...state.vehicles[id],
        ...(update as any),
        id,
        latitude: update.lat ?? update.latitude ?? state.vehicles[id]?.latitude,
        longitude:
          update.lng ?? update.longitude ?? state.vehicles[id]?.longitude,
      };
      state.count = Object.keys(state.vehicles).length;
    },
    setSelectedVehicle: (state, action: PayloadAction<string | null>) => {
      state.selectedVehicleId = action.payload;
    },
  },
});

export const {
  updateTelemetryBatch,
  updateBreadcrumbs,
  clearTelemetry,
  updateVehicle,
  setSelectedVehicle,
} = telemetrySlice.actions;
export default telemetrySlice.reducer;
