// store/trackingSlice.test.ts

import trackingReducer, {
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
} from '../store/trackingSlice';
import { VehiclePosition, VehicleStatus } from '../types/tracking';

describe('trackingSlice', () => {
    const initialState = {
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

    describe('setAllVehiclePositions', () => {
        it('should set all vehicle positions', () => {
            const vehicles: VehiclePosition[] = [
                {
                    vehicle_id: 'v1',
                    org_id: 'org1',
                    registration: 'DL-01-AB-1234',
                    make: 'Tata',
                    model: 'Ace',
                    year: 2024,
                    vehicle_status: 'active',
                    current_driver_id: null,
                    driver_name: null,
                    driver_phone: null,
                    last_update: '2026-04-07T10:30:00Z',
                    latitude: 28.6139,
                    longitude: 77.2090,
                    speed_kmh: 45.5,
                    heading: 180,
                    ignition: true,
                    fuel_level: 72.5,
                    derived_status: 'moving',
                },
                {
                    vehicle_id: 'v2',
                    org_id: 'org1',
                    registration: 'MH-02-CD-5678',
                    make: 'Ashok Leyland',
                    model: 'Dost',
                    year: 2023,
                    vehicle_status: 'active',
                    current_driver_id: null,
                    driver_name: null,
                    driver_phone: null,
                    last_update: '2026-04-07T10:29:00Z',
                    latitude: 19.0760,
                    longitude: 72.8777,
                    speed_kmh: 0,
                    heading: null,
                    ignition: false,
                    fuel_level: 50,
                    derived_status: 'stopped',
                },
            ];

            const state = trackingReducer(initialState, setAllVehiclePositions(vehicles));

            expect(Object.keys(state.vehicles)).toHaveLength(2);
            expect(state.vehicles['v1'].registration).toBe('DL-01-AB-1234');
            expect(state.vehicles['v2'].derived_status).toBe('stopped');
        });

        it('should replace existing vehicles', () => {
            const vehicles1: VehiclePosition[] = [
                {
                    vehicle_id: 'v1',
                    org_id: 'org1',
                    registration: 'OLD',
                    make: '',
                    model: '',
                    year: 2024,
                    vehicle_status: 'active',
                    current_driver_id: null,
                    driver_name: null,
                    driver_phone: null,
                    last_update: '2026-04-07T10:30:00Z',
                    latitude: 0,
                    longitude: 0,
                    speed_kmh: 0,
                    heading: null,
                    ignition: null,
                    fuel_level: null,
                    derived_status: 'idle',
                },
            ];

            const state1 = trackingReducer(initialState, setAllVehiclePositions(vehicles1));

            const vehicles2: VehiclePosition[] = [
                {
                    vehicle_id: 'v1',
                    org_id: 'org1',
                    registration: 'UPDATED',
                    make: '',
                    model: '',
                    year: 2024,
                    vehicle_status: 'active',
                    current_driver_id: null,
                    driver_name: null,
                    driver_phone: null,
                    last_update: '2026-04-07T10:31:00Z',
                    latitude: 1,
                    longitude: 1,
                    speed_kmh: 10,
                    heading: null,
                    ignition: null,
                    fuel_level: null,
                    derived_status: 'moving',
                },
            ];

            const state2 = trackingReducer(state1, setAllVehiclePositions(vehicles2));

            expect(state2.vehicles['v1'].registration).toBe('UPDATED');
            expect(state2.vehicles['v1'].speed_kmh).toBe(10);
        });
    });

    describe('updateVehiclePosition', () => {
        it('should update a single vehicle position', () => {
            const state = trackingReducer(initialState, updateVehiclePosition({
                vehicle_id: 'v1',
                org_id: 'org1',
                registration: 'DL-01-AB-1234',
                make: 'Tata',
                model: 'Ace',
                year: 2024,
                vehicle_status: 'active',
                current_driver_id: null,
                driver_name: null,
                driver_phone: null,
                last_update: '2026-04-07T10:30:00Z',
                latitude: 28.6139,
                longitude: 77.2090,
                speed_kmh: 50,
                heading: 180,
                ignition: true,
                fuel_level: 70,
                derived_status: 'moving',
            }));

            expect(state.vehicles['v1']).toBeDefined();
            expect(state.vehicles['v1'].speed_kmh).toBe(50);
        });
    });

    describe('setSelectedVehicle', () => {
        it('should set selected vehicle and open detail panel', () => {
            const state = trackingReducer(initialState, setSelectedVehicle('v1'));
            expect(state.selectedVehicleId).toBe('v1');
            expect(state.detailPanelOpen).toBe(true);
        });

        it('should clear selection when null', () => {
            const state1 = trackingReducer(initialState, setSelectedVehicle('v1'));
            const state2 = trackingReducer(state1, setSelectedVehicle(null));
            expect(state2.selectedVehicleId).toBeNull();
            expect(state2.detailPanelOpen).toBe(false);
        });
    });

    describe('setSearchQuery', () => {
        it('should set search query', () => {
            const state = trackingReducer(initialState, setSearchQuery('DL-01'));
            expect(state.searchQuery).toBe('DL-01');
        });
    });

    describe('setStatusFilters', () => {
        it('should set status filters', () => {
            const filters: VehicleStatus[] = ['moving', 'idle'];
            const state = trackingReducer(initialState, setStatusFilters(filters));
            expect(state.statusFilters).toEqual(['moving', 'idle']);
        });

        it('should allow clearing filters', () => {
            const state1 = trackingReducer(initialState, setStatusFilters(['moving']));
            const state2 = trackingReducer(state1, setStatusFilters([]));
            expect(state2.statusFilters).toEqual([]);
        });
    });

    describe('WebSocket state', () => {
        it('should set wsConnected', () => {
            const state = trackingReducer(initialState, setWsConnected(true));
            expect(state.wsConnected).toBe(true);
        });

        it('should set reconnect attempts', () => {
            const state = trackingReducer(initialState, setWsReconnectAttempts(3));
            expect(state.wsReconnectAttempts).toBe(3);
        });

        it('should set fallback to polling', () => {
            const state = trackingReducer(initialState, setFallbackToPolling(true));
            expect(state.fallbackToPolling).toBe(true);
        });
    });

    describe('setHistoryTrail', () => {
        it('should set history trail data', () => {
            const trailData = {
                vehicle_id: 'v1',
                points: [
                    { time: '2026-04-07T10:00:00Z', vehicle_id: 'v1', latitude: 28.6139, longitude: 77.2090, speed_kmh: 40, heading: null, altitude_m: null, ignition: null, fuel_level: null },
                    { time: '2026-04-07T10:01:00Z', vehicle_id: 'v1', latitude: 28.6140, longitude: 77.2091, speed_kmh: 45, heading: null, altitude_m: null, ignition: null, fuel_level: null },
                ],
                summary: {
                    total_distance_km: 125.3,
                    avg_speed_kmh: 42.1,
                    max_speed_kmh: 85,
                    duration_minutes: 180,
                    stops_count: 3,
                },
            };

            const state = trackingReducer(initialState, setHistoryTrail(trailData));

            expect(state.historyTrail).not.toBeNull();
            expect(state.historyTrail?.vehicle_id).toBe('v1');
            expect(state.historyTrail?.points).toHaveLength(2);
            expect(state.historyTrail?.summary.total_distance_km).toBe(125.3);
        });
    });

    describe('setGeofences', () => {
        it('should set geofences', () => {
            const geofences = [
                {
                    id: 'g1',
                    org_id: 'org1',
                    name: 'Delhi Warehouse',
                    description: '',
                    type: 'circle' as const,
                    center_lat: 28.6139,
                    center_lng: 77.2090,
                    radius_m: 500,
                    polygon: null,
                    status: 'active' as const,
                    alert_channels: ['in_app'],
                    created_at: '2026-04-01T08:00:00Z',
                    updated_at: '2026-04-01T08:00:00Z',
                },
            ];

            const state = trackingReducer(initialState, setGeofences(geofences));

            expect(state.geofences).toHaveLength(1);
            expect(state.geofences[0].name).toBe('Delhi Warehouse');
        });
    });

    describe('addGeofenceEvent', () => {
        it('should add geofence event for vehicle', () => {
            const event = {
                id: 'e1',
                geofence_id: 'g1',
                geofence_name: 'Delhi Warehouse',
                vehicle_id: 'v1',
                registration: 'DL-01-AB-1234',
                event_type: 'entry' as const,
                occurred_at: '2026-04-07T10:30:00Z',
                latitude: 28.6139,
                longitude: 77.2090,
                alert_sent: false,
            };

            const state = trackingReducer(initialState, addGeofenceEvent(event));

            expect(state.geofenceEvents['v1']).toHaveLength(1);
            expect(state.geofenceEvents['v1'][0].event_type).toBe('entry');
        });

        it('should append multiple events for same vehicle', () => {
            const event1 = {
                id: 'e1',
                geofence_id: 'g1',
                geofence_name: 'Zone A',
                vehicle_id: 'v1',
                registration: 'DL-01',
                event_type: 'entry' as const,
                occurred_at: '2026-04-07T10:00:00Z',
                latitude: 28.6139,
                longitude: 77.2090,
                alert_sent: false,
            };

            const event2 = {
                id: 'e2',
                geofence_id: 'g1',
                geofence_name: 'Zone A',
                vehicle_id: 'v1',
                registration: 'DL-01',
                event_type: 'exit' as const,
                occurred_at: '2026-04-07T11:00:00Z',
                latitude: 28.6150,
                longitude: 77.2100,
                alert_sent: false,
            };

            const state1 = trackingReducer(initialState, addGeofenceEvent(event1));
            const state2 = trackingReducer(state1, addGeofenceEvent(event2));

            expect(state2.geofenceEvents['v1']).toHaveLength(2);
        });
    });

    describe('setLoading', () => {
        it('should set loading state for vehicles', () => {
            const state = trackingReducer(initialState, setLoading({ key: 'vehicles', value: true }));
            expect(state.loading.vehicles).toBe(true);
            expect(state.loading.history).toBe(false);
            expect(state.loading.geofences).toBe(false);
        });

        it('should set loading state for history', () => {
            const state = trackingReducer(initialState, setLoading({ key: 'history', value: true }));
            expect(state.loading.history).toBe(true);
        });
    });

    describe('setError', () => {
        it('should set error message', () => {
            const state = trackingReducer(initialState, setError('Connection failed'));
            expect(state.error).toBe('Connection failed');
        });

        it('should clear error', () => {
            const state1 = trackingReducer(initialState, setError('Some error'));
            const state2 = trackingReducer(state1, setError(null));
            expect(state2.error).toBeNull();
        });
    });

    describe('setPlaybackState', () => {
        it('should set playing state', () => {
            const state = trackingReducer(initialState, setPlaybackState({ isPlaying: true }));
            expect(state.isPlaying).toBe(true);
        });

        it('should set playback speed', () => {
            const state = trackingReducer(initialState, setPlaybackState({ isPlaying: true, speed: 4 }));
            expect(state.playbackSpeed).toBe(4);
        });

        it('should set playback position', () => {
            const state = trackingReducer(initialState, setPlaybackState({ isPlaying: false, position: 42 }));
            expect(state.playbackPosition).toBe(42);
        });

        it('should pause and reset position', () => {
            const state1 = trackingReducer(initialState, setPlaybackState({ isPlaying: true, position: 50 }));
            const state2 = trackingReducer(state1, setPlaybackState({ isPlaying: false, position: 0 }));
            expect(state2.isPlaying).toBe(false);
            expect(state2.playbackPosition).toBe(0);
        });
    });
});
