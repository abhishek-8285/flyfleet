// components/VehicleMap.tsx

import React, { useEffect, useRef, useCallback } from 'react';
import maplibregl from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { useSelector } from 'react-redux';
import { RootState } from '../../../store';
import { statusColor } from '../utils/statusColor';

export const VehicleMap: React.FC = () => {
    const mapRef = useRef<maplibregl.Map | null>(null);
    const markersRef = useRef<maplibregl.Marker[]>([]);
    const vehicles = useSelector((state: RootState) => state.tracking.vehicles);
    const selectedVehicleId = useSelector((state: RootState) => state.tracking.selectedVehicleId);

    useEffect(() => {
        // Initialize MapLibre
        mapRef.current = new maplibregl.Map({
            container: 'tracking-map',
            style: {
                version: 8,
                name: 'FlyFleet Map',
                sources: {
                    'osm-tiles': {
                        type: 'raster',
                        tiles: ['https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}&gl=IN'],
                        tileSize: 256,
                        attribution: '© Google Maps'
                    }
                },
                layers: [{
                    id: 'osm-tiles',
                    type: 'raster',
                    source: 'osm-tiles',
                    minzoom: 0,
                    maxzoom: 19
                }]
            },
            center: [78.9629, 20.5937], // India center
            zoom: 4,
        });

        // Add navigation controls
        mapRef.current.addControl(new maplibregl.NavigationControl(), 'top-right');
        return () => {
            mapRef.current?.remove();
        };
    }, []);

    // Update markers when vehicles change
    useEffect(() => {
        if (!mapRef.current) return;

        // Clear old markers
        markersRef.current.forEach(m => m.remove());
        markersRef.current = [];

        // Add markers for each vehicle
        Object.values(vehicles).forEach(vehicle => {
            const color = statusColor(vehicle.derived_status);

            // Create marker element
            const el = document.createElement('div');
            el.className = 'vehicle-marker';
            el.style.width = '24px';
            el.style.height = '24px';
            el.style.borderRadius = '50%';
            el.style.backgroundColor = color;
            el.style.border = '2px solid white';
            el.style.boxShadow = '0 2px 4px rgba(0,0,0,0.3)';
            el.style.cursor = 'pointer';

            // Click handler
            el.addEventListener('click', () => {
                // Dispatch select vehicle action
            });

            const marker = new maplibregl.Marker({ element: el })
                .setLngLat([vehicle.longitude, vehicle.latitude])
                .setPopup(new maplibregl.Popup({ offset: 25 })
                    .setHTML(`
                        <div class="marker-popup">
                            <strong>${vehicle.registration}</strong><br/>
                            <span style="color: ${color}">●</span> ${vehicle.derived_status}<br/>
                            Speed: ${vehicle.speed_kmh?.toFixed(1) || 0} km/h
                        </div>
                    `)
                )
                .addTo(mapRef.current!);

            markersRef.current.push(marker);
        });

        // Fit map to show all vehicles if none selected
        if (!selectedVehicleId && Object.keys(vehicles).length > 0) {
            const bounds = new maplibregl.LngLatBounds();
            Object.values(vehicles).forEach(v => {
                bounds.extend([v.longitude, v.latitude]);
            });
            mapRef.current.fitBounds(bounds, { padding: 50 });
        }

        // Center on selected vehicle
        if (selectedVehicleId && vehicles[selectedVehicleId]) {
            const vehicle = vehicles[selectedVehicleId];
            mapRef.current.flyTo({
                center: [vehicle.longitude, vehicle.latitude],
                zoom: 14,
                duration: 1000
            });
        }
    }, [vehicles, selectedVehicleId]);

    return (
        <div id="tracking-map" className="w-full h-full" />
    );
};
