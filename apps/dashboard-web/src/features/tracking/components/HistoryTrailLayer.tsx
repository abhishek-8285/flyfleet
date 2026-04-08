// components/HistoryTrailLayer.tsx

import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { RootState } from '../../../store';
import maplibregl from 'maplibre-gl';

interface HistoryTrailLayerProps {
    map: maplibregl.Map | null;
}

export const HistoryTrailLayer: React.FC<HistoryTrailLayerProps> = ({ map }) => {
    const historyTrail = useSelector((state: RootState) => state.tracking.historyTrail);
    const playbackPosition = useSelector((state: RootState) => state.tracking.playbackPosition);

    useEffect(() => {
        if (!map || !historyTrail) return;

        // Remove existing trail layers
        ['trail-line', 'trail-points', 'playback-marker'].forEach(id => {
            if (map.getLayer(id)) map.removeLayer(id);
            if (map.getSource(id)) map.removeSource(id);
        });

        // Filter points up to playback position
        const points = playbackPosition !== null
            ? historyTrail.points.slice(0, playbackPosition + 1)
            : historyTrail.points;

        if (points.length < 2) return;

        // Build GeoJSON for trail line
        const lineCoords = points.map(p => [p.longitude, p.latitude]);
        const lineGeoJSON = {
            type: 'Feature',
            geometry: {
                type: 'LineString',
                coordinates: lineCoords
            },
            properties: {}
        };

        // Add trail line source and layer
        map.addSource('trail-line', {
            type: 'geojson',
            data: lineGeoJSON as any
        });

        map.addLayer({
            id: 'trail-line',
            type: 'line',
            source: 'trail-line',
            layout: {
                'line-join': 'round',
                'line-cap': 'round'
            },
            paint: {
                'line-color': '#3b82f6',
                'line-width': 3,
                'line-opacity': 0.8
            }
        });

        // Add trail points as circles
        const pointFeatures = points.map((p, i) => ({
            type: 'Feature',
            geometry: {
                type: 'Point',
                coordinates: [p.longitude, p.latitude]
            },
            properties: {
                speed: p.speed_kmh || 0,
                index: i
            }
        }));

        map.addSource('trail-points', {
            type: 'geojson',
            data: { type: 'FeatureCollection', features: pointFeatures } as any
        });

        // Color points by speed
        map.addLayer({
            id: 'trail-points',
            type: 'circle',
            source: 'trail-points',
            paint: {
                'circle-radius': 4,
                'circle-color': [
                    'case',
                    ['>', ['get', 'speed'], 60], '#ef4444',   // Red: fast
                    ['>', ['get', 'speed'], 30], '#f59e0b',   // Amber: medium
                    '#22c55e'                                    // Green: slow
                ],
                'circle-opacity': 0.7
            }
        });

        // Add playback marker
        if (playbackPosition !== null && points[playbackPosition]) {
            const currentPoint = points[playbackPosition];
            
            if (!map.getSource('playback-marker')) {
                map.addSource('playback-marker', {
                    type: 'geojson',
                    data: {
                        type: 'Feature',
                        geometry: {
                            type: 'Point',
                            coordinates: [currentPoint.longitude, currentPoint.latitude]
                        },
                        properties: {}
                    } as any
                });

                map.addLayer({
                    id: 'playback-marker',
                    type: 'circle',
                    source: 'playback-marker',
                    paint: {
                        'circle-radius': 8,
                        'circle-color': '#1d4ed8',
                        'circle-stroke-width': 2,
                        'circle-stroke-color': '#ffffff'
                    }
                });
            } else {
                // Update existing marker position
                const source = map.getSource('playback-marker') as maplibregl.GeoJSONSource;
                source.setData({
                    type: 'Feature',
                    geometry: {
                        type: 'Point',
                        coordinates: [currentPoint.longitude, currentPoint.latitude]
                    },
                    properties: {}
                } as any);
            }
        }
    }, [map, historyTrail, playbackPosition]);

    return null;
};
