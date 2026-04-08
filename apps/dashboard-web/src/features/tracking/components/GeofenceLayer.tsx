// components/GeofenceLayer.tsx

import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { RootState } from '../../../store';
import maplibregl from 'maplibre-gl';

interface GeofenceLayerProps {
    map: maplibregl.Map | null;
}

export const GeofenceLayer: React.FC<GeofenceLayerProps> = ({ map }) => {
    const geofences = useSelector((state: RootState) => state.tracking.geofences);

    useEffect(() => {
        if (!map) return;

        // Clear existing geofence layers
        if (map.getLayer('geofence-circles')) {
            map.removeLayer('geofence-circles');
            map.removeSource('geofence-circles');
        }
        if (map.getLayer('geofence-polygons')) {
            map.removeLayer('geofence-polygons');
            map.removeSource('geofence-polygons');
        }

        // Build GeoJSON from geofences
        const circleFeatures: any[] = [];
        const polygonFeatures: any[] = [];

        geofences.forEach(gf => {
            if (gf.type === 'circle' && gf.center_lat && gf.center_lng && gf.radius_m) {
                circleFeatures.push({
                    type: 'Feature',
                    geometry: {
                        type: 'Point',
                        coordinates: [gf.center_lng, gf.center_lat]
                    },
                    properties: {
                        id: gf.id,
                        name: gf.name,
                        radius: gf.radius_m,
                        status: gf.status
                    }
                });
            } else if (gf.type === 'polygon' && gf.polygon) {
                polygonFeatures.push({
                    type: 'Feature',
                    geometry: {
                        type: 'Polygon',
                        coordinates: [[...gf.polygon.map(p => [p[0], p[1]]), [...gf.polygon[0]]]]
                    },
                    properties: {
                        id: gf.id,
                        name: gf.name,
                        status: gf.status
                    }
                });
            }
        });

        // Add circle geofences
        if (circleFeatures.length > 0) {
            map.addSource('geofence-circles', {
                type: 'geojson',
                data: { type: 'FeatureCollection', features: circleFeatures }
            });

            map.addLayer({
                id: 'geofence-circles',
                type: 'circle',
                source: 'geofence-circles',
                paint: {
                    'circle-radius': ['get', 'radius'],
                    'circle-color': [
                        'case',
                        ['==', ['get', 'status'], 'active'],
                        '#3b82f6',
                        '#9ca3af'
                    ],
                    'circle-opacity': 0.3,
                    'circle-stroke-width': 2,
                    'circle-stroke-color': [
                        'case',
                        ['==', ['get', 'status'], 'active'],
                        '#2563eb',
                        '#6b7280'
                    ]
                }
            });
        }

        // Add polygon geofences
        if (polygonFeatures.length > 0) {
            map.addSource('geofence-polygons', {
                type: 'geojson',
                data: { type: 'FeatureCollection', features: polygonFeatures }
            });

            map.addLayer({
                id: 'geofence-polygons',
                type: 'fill',
                source: 'geofence-polygons',
                paint: {
                    'fill-color': [
                        'case',
                        ['==', ['get', 'status'], 'active'],
                        '#3b82f6',
                        '#9ca3af'
                    ],
                    'fill-opacity': 0.2,
                    'fill-outline-color': [
                        'case',
                        ['==', ['get', 'status'], 'active'],
                        '#2563eb',
                        '#6b7280'
                    ]
                }
            });
        }
    }, [map, geofences]);

    return null; // This is a layer component, no DOM rendering
};
