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
        <div className="relative h-screen w-screen flex flex-col bg-surface-dim overflow-hidden font-sans antialiased">
            {/* Intel Command Overlay (Top Bar) */}
            <div className="absolute top-8 left-8 right-8 z-10 flex flex-col md:flex-row gap-6 items-start justify-between pointer-events-none">
                {/* Search & Insight Discovery */}
                <div className="w-full max-w-xl pointer-events-auto">
                    <div className="bg-surface-container/90 backdrop-blur-xl border border-outline rounded-[2.5rem] p-6 shadow-2xl">
                        <div className="flex items-center gap-4 mb-4 opacity-50">
                            <span className="material-symbols-outlined text-lg">radar</span>
                            <span className="text-[10px] font-black uppercase tracking-[0.3em] text-white">Live Asset Surveillance</span>
                        </div>
                        <SearchBar />
                    </div>
                </div>

                {/* Filter Matrix */}
                <div className="w-full max-w-2xl pointer-events-auto self-end">
                    <div className="bg-surface-container/90 backdrop-blur-xl border border-outline rounded-[2rem] p-4 shadow-2xl">
                        <FilterBar />
                    </div>
                </div>
            </div>

            {/* System Readiness Indicator */}
            <div className="absolute bottom-8 left-8 z-10 pointer-events-auto">
                <div className="bg-surface-container/90 backdrop-blur-xl border border-outline rounded-3xl p-4 shadow-2xl">
                    <ConnectionIndicator />
                </div>
            </div>

            {/* Asset Density Counter */}
            <div className="absolute bottom-8 right-8 z-10 pointer-events-auto">
                <div className="bg-surface-container/90 backdrop-blur-xl border border-outline rounded-2xl px-6 py-4 shadow-2xl flex items-center gap-4">
                    <div className="h-2 w-2 rounded-full bg-primary animate-pulse" />
                    <span className="text-[11px] font-black text-white uppercase tracking-widest">
                        <span className="text-primary">{filteredVehicles.length}</span> / <span className="opacity-50">{Object.keys(vehicles).length}</span> Active Fleet Units
                    </span>
                </div>
            </div>

            {/* Operational Theater (Map) */}
            <div className="absolute inset-0 z-0">
                <VehicleMap />
            </div>

            {/* Intel Detail Panel */}
            <div className="absolute right-8 top-1/2 -translate-y-1/2 z-10 z-20 pointer-events-auto">
                <VehicleDetailPanel />
            </div>

            {/* Chronological Playback (when history is loaded) */}
            <div className="absolute bottom-28 left-8 z-10 max-w-sm pointer-events-auto">
                <PlaybackControls />
            </div>
        </div>
    );
};

export default TrackingPage;
