// store/trackingSlice.ts

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import {
    VehiclePosition,
    VehicleStatus,
    Geofence,
    GeofenceEvent,
    TelemetryPoint,
    HistorySummary,
} from '../types/tracking';

interface TrackingState {
    vehicles: Record<string, VehiclePosition>;
    geofences: Geofence[];
    geofenceEvents: Record<string, GeofenceEvent[]>;
    selectedVehicleId: string | null;
    historyTrail: {
        vehicle_id: string;
        points: TelemetryPoint[];
        summary: HistorySummary;
    } | null;
    searchQuery: string;
    statusFilters: VehicleStatus[];
    typeFilters: string[];
    wsConnected: boolean;
    wsLastMessage: number | null;
    wsReconnectAttempts: number;
    isPlaying: boolean;
    playbackSpeed: number;
    playbackPosition: number | null;
    detailPanelOpen: boolean;
    historyPanelOpen: boolean;
    loading: {
        vehicles: boolean;
        history: boolean;
        geofences: boolean;
    };
    error: string | null;
    fallbackToPolling: boolean;
}

const initialState: TrackingState = {
    vehicles: {},
    geofences: [],
    geofenceEvents: {},
    selectedVehicleId: null,
    historyTrail: null,
    searchQuery: '',
    statusFilters: [],
    typeFilters: [],
    wsConnected: false,
    wsLastMessage: null,
    wsReconnectAttempts: 0,
    isPlaying: false,
    playbackSpeed: 1,
    playbackPosition: null,
    detailPanelOpen: false,
    historyPanelOpen: false,
    loading: {
        vehicles: false,
        history: false,
        geofences: false,
    },
    error: null,
    fallbackToPolling: false,
};

const trackingSlice = createSlice({
    name: 'tracking',
    initialState,
    reducers: {
        setAllVehiclePositions: (state, action: PayloadAction<VehiclePosition[]>) => {
            const vehicles: Record<string, VehiclePosition> = {};
            action.payload.forEach(v => {
                vehicles[v.vehicle_id] = v;
            });
            state.vehicles = vehicles;
        },
        updateVehiclePosition: (state, action: PayloadAction<VehiclePosition>) => {
            const vehicle = action.payload;
            state.vehicles[vehicle.vehicle_id] = vehicle;
        },
        setSelectedVehicle: (state, action: PayloadAction<string | null>) => {
            state.selectedVehicleId = action.payload;
            state.detailPanelOpen = action.payload !== null;
        },
        setSearchQuery: (state, action: PayloadAction<string>) => {
            state.searchQuery = action.payload;
        },
        setStatusFilters: (state, action: PayloadAction<VehicleStatus[]>) => {
            state.statusFilters = action.payload;
        },
        setWsConnected: (state, action: PayloadAction<boolean>) => {
            state.wsConnected = action.payload;
        },
        setWsReconnectAttempts: (state, action: PayloadAction<number>) => {
            state.wsReconnectAttempts = action.payload;
        },
        setFallbackToPolling: (state, action: PayloadAction<boolean>) => {
            state.fallbackToPolling = action.payload;
        },
        setHistoryTrail: (state, action: PayloadAction<{
            vehicle_id: string;
            points: TelemetryPoint[];
            summary: HistorySummary;
        }>) => {
            state.historyTrail = action.payload;
        },
        setGeofences: (state, action: PayloadAction<Geofence[]>) => {
            state.geofences = action.payload;
        },
        addGeofenceEvent: (state, action: PayloadAction<GeofenceEvent>) => {
            const event = action.payload;
            if (!state.geofenceEvents[event.vehicle_id]) {
                state.geofenceEvents[event.vehicle_id] = [];
            }
            state.geofenceEvents[event.vehicle_id].push(event);
        },
        setLoading: (state, action: PayloadAction<{ key: keyof TrackingState['loading'], value: boolean }>) => {
            state.loading[action.payload.key] = action.payload.value;
        },
        setError: (state, action: PayloadAction<string | null>) => {
            state.error = action.payload;
        },
        setPlaybackState: (state, action: PayloadAction<{
            isPlaying: boolean;
            speed?: number;
            position?: number;
        }>) => {
            state.isPlaying = action.payload.isPlaying;
            if (action.payload.speed) state.playbackSpeed = action.payload.speed;
            if (action.payload.position !== undefined) state.playbackPosition = action.payload.position;
        },
    },
});

export const {
    setAllVehiclePositions,
    updateVehiclePosition,
    setSelectedVehicle,
    setSearchQuery,
    setStatusFilters,
    setWsConnected,
    setWsReconnectAttempts,
    setFallbackToPolling,
    setHistoryTrail,
    setGeofences,
    addGeofenceEvent,
    setLoading,
    setError,
    setPlaybackState,
} = trackingSlice.actions;

export default trackingSlice.reducer;
