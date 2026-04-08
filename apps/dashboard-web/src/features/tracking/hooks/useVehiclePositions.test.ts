// hooks/useVehiclePositions.test.ts

import { renderHook } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import trackingReducer from '../store/trackingSlice';
import { useVehiclePositions } from './useVehiclePositions';
import { VehiclePosition } from '../types/tracking';

const createTestStore = (vehicles: Record<string, VehiclePosition> = {}) => {
    return configureStore({
        reducer: {
            tracking: trackingReducer,
        },
        preloadedState: {
            tracking: {
                vehicles,
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
                loading: { vehicles: false, history: false, geofences: false },
                error: null,
                fallbackToPolling: false,
            },
        },
    });
};

const wrapper = (store: ReturnType<typeof createTestStore>) => {
    return ({ children }: { children: React.ReactNode }) => (
        <Provider store={store}>{children}</Provider>
    );
};

const createMockVehicle = (overrides: Partial<VehiclePosition> = {}): VehiclePosition => ({
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
    speed_kmh: 40,
    heading: 180,
    ignition: true,
    fuel_level: 70,
    derived_status: 'moving',
    ...overrides,
});

describe('useVehiclePositions', () => {
    it('should return all vehicles when no filters applied', () => {
        const vehicles = {
            v1: createMockVehicle({ vehicle_id: 'v1', derived_status: 'moving' as const }),
            v2: createMockVehicle({ vehicle_id: 'v2', derived_status: 'idle' as const }),
            v3: createMockVehicle({ vehicle_id: 'v3', derived_status: 'offline' as const }),
        };

        const store = createTestStore(vehicles);
        const { result } = renderHook(() => useVehiclePositions(), {
            wrapper: wrapper(store),
        });

        expect(result.current).toHaveLength(3);
    });

    it('should filter by status', () => {
        const vehicles = {
            v1: createMockVehicle({ vehicle_id: 'v1', derived_status: 'moving' as const }),
            v2: createMockVehicle({ vehicle_id: 'v2', derived_status: 'idle' as const }),
            v3: createMockVehicle({ vehicle_id: 'v3', derived_status: 'moving' as const }),
        };

        const store = createTestStore(vehicles);
        // Manually set status filter
        store.dispatch({ type: 'tracking/setStatusFilters', payload: ['moving'] });

        const { result } = renderHook(() => useVehiclePositions(), {
            wrapper: wrapper(store),
        });

        expect(result.current).toHaveLength(2);
        expect(result.current.every(v => v.derived_status === 'moving')).toBe(true);
    });

    it('should filter by search query', () => {
        const vehicles = {
            v1: createMockVehicle({ vehicle_id: 'v1', registration: 'DL-01-AB-1234' }),
            v2: createMockVehicle({ vehicle_id: 'v2', registration: 'MH-02-CD-5678' }),
        };

        const store = createTestStore(vehicles);
        store.dispatch({ type: 'tracking/setSearchQuery', payload: 'DL-01' });

        const { result } = renderHook(() => useVehiclePositions(), {
            wrapper: wrapper(store),
        });

        expect(result.current).toHaveLength(1);
        expect(result.current[0].registration).toBe('DL-01-AB-1234');
    });

    it('should apply both status and search filters', () => {
        const vehicles = {
            v1: createMockVehicle({ vehicle_id: 'v1', registration: 'DL-01-MOVING', derived_status: 'moving' as const }),
            v2: createMockVehicle({ vehicle_id: 'v2', registration: 'DL-01-IDLE', derived_status: 'idle' as const }),
            v3: createMockVehicle({ vehicle_id: 'v3', registration: 'MH-02-MOVING', derived_status: 'moving' as const }),
        };

        const store = createTestStore(vehicles);
        store.dispatch({ type: 'tracking/setStatusFilters', payload: ['moving'] });
        store.dispatch({ type: 'tracking/setSearchQuery', payload: 'DL-01' });

        const { result } = renderHook(() => useVehiclePositions(), {
            wrapper: wrapper(store),
        });

        expect(result.current).toHaveLength(1);
        expect(result.current[0].registration).toBe('DL-01-MOVING');
    });

    it('should return empty array when no vehicles', () => {
        const store = createTestStore({});
        const { result } = renderHook(() => useVehiclePositions(), {
            wrapper: wrapper(store),
        });
        expect(result.current).toHaveLength(0);
    });

    it('should search by make and model', () => {
        const vehicles = {
            v1: createMockVehicle({ vehicle_id: 'v1', make: 'Tata', model: 'Ace' }),
            v2: createMockVehicle({ vehicle_id: 'v2', make: 'Ashok Leyland', model: 'Dost' }),
        };

        const store = createTestStore(vehicles);
        store.dispatch({ type: 'tracking/setSearchQuery', payload: 'tata' });

        const { result } = renderHook(() => useVehiclePositions(), {
            wrapper: wrapper(store),
        });

        expect(result.current).toHaveLength(1);
        expect(result.current[0].make).toBe('Tata');
    });
});
