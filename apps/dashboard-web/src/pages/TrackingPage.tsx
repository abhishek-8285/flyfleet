// pages/TrackingPage.tsx

import React, { useEffect, useRef, useState, useCallback } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import maplibregl from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';

import { VehicleMap } from '../features/tracking/components/VehicleMap';
import { VehicleDetailPanel } from '../features/tracking/components/VehicleDetailPanel';
import { SearchBar } from '../features/tracking/components/SearchBar';
import { FilterBar } from '../features/tracking/components/FilterBar';
import { ConnectionIndicator } from '../features/tracking/components/ConnectionIndicator';
import { PlaybackControls } from '../features/tracking/components/PlaybackControls';
import { GeofenceLayer } from '../features/tracking/components/GeofenceLayer';
import { HistoryTrailLayer } from '../features/tracking/components/HistoryTrailLayer';
import { telemetryWS } from '../features/tracking/services/websocketService';
import { setAllVehiclePositions, setLoading, setError } from '../features/tracking/store/trackingSlice';
import { useVehiclePositions } from '../features/tracking/hooks/useVehiclePositions';
import { AppDispatch, RootState } from '../store';

const TrackingPage: React.FC = () => {
    const dispatch = useDispatch<AppDispatch>();
    const mapRef = useRef<maplibregl.Map | null>(null);
    const markersRef = useRef<maplibregl.Marker[]>([]);
    const wsInitialized = useRef(false);

    const vehicles = useSelector((state: RootState) => state.tracking.vehicles);
    const selectedVehicleId = useSelector((state: RootState) => state.tracking.selectedVehicleId);
    const wsConnected = useSelector((state: RootState) => state.tracking.wsConnected);
    const fallbackToPolling = useSelector((state: RootState) => state.tracking.fallbackToPolling);

    const filteredVehicles = useVehiclePositions();

    // Load initial positions
    useEffect(() => {
        const loadVehicles = async () => {
            try {
                dispatch(setLoading({ key: 'vehicles', value: true }));
                const token = localStorage.getItem('auth_token');
                const response = await fetch('/api/v1/vehicles/locations', {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (!response.ok) throw new Error('Failed to load vehicles');
                const data = await response.json();
                dispatch(setAllVehiclePositions(data.data));
            } catch (err: any) {
                dispatch(setError(err.message));
            } finally {
                dispatch(setLoading({ key: 'vehicles', value: false }));
            }
        };

        loadVehicles();
    }, [dispatch]);

    // Initialize WebSocket connection
    useEffect(() => {
        if (wsInitialized.current) return;
        wsInitialized.current = true;

        const token = localStorage.getItem('auth_token');
        const orgId = localStorage.getItem('org_id');

        if (token && orgId) {
            telemetryWS.connect(token, orgId);
        }

        return () => {
            telemetryWS.disconnect();
        };
    }, []);

    // Polling fallback
    useEffect(() => {
        if (!fallbackToPolling) return;

        const poll = async () => {
            try {
                const token = localStorage.getItem('auth_token');
                const response = await fetch('/api/v1/vehicles/locations', {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (response.ok) {
                    const data = await response.json();
                    dispatch(setAllVehiclePositions(data.data));
                }
            } catch (err) {
                console.error('Polling failed:', err);
            }
        };

        poll();
        const interval = setInterval(poll, 10000);
        return () => clearInterval(interval);
    }, [fallbackToPolling, dispatch]);

    return (
        <div className="relative h-screen w-screen flex flex-col">
            {/* Top Bar */}
            <div className="absolute top-0 left-0 right-0 z-10 p-4 space-y-3">
                {/* Search Bar */}
                <div className="max-w-md mx-auto">
                    <SearchBar />
                </div>

                {/* Filter Bar */}
                <div className="max-w-2xl mx-auto">
                    <FilterBar />
                </div>
            </div>

            {/* Connection Status */}
            <div className="absolute bottom-4 left-4 z-10">
                <ConnectionIndicator />
            </div>

            {/* Vehicle Count Badge */}
            <div className="absolute bottom-4 right-4 z-10 bg-white rounded-full shadow-lg px-4 py-2">
                <span className="text-sm text-gray-600">
                    {filteredVehicles.length} / {Object.keys(vehicles).length} vehicles
                </span>
            </div>

            {/* Map */}
            <VehicleMap />

            {/* Detail Panel */}
            <VehicleDetailPanel />

            {/* Playback Controls (when history is loaded) */}
            <div className="absolute bottom-16 left-4 z-10 max-w-sm">
                <PlaybackControls />
            </div>
        </div>
    );
};

export default TrackingPage;
