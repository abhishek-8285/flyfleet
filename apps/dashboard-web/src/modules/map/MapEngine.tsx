import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import React, { useEffect, useRef } from "react";

interface MapEngineProps {
  center?: [number, number];
  zoom?: number;
  onMapLoad?: (map: maplibregl.Map) => void;
}

const DEFAULT_CENTER: [number, number] = [78.9629, 20.5937];
const DEFAULT_ZOOM = 5;

const MapEngine: React.FC<MapEngineProps> = ({
  center = DEFAULT_CENTER,
  zoom = DEFAULT_ZOOM,
  onMapLoad,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<maplibregl.Map | null>(null);
  const onMapLoadRef = useRef(onMapLoad);
  
  // Track previous prop values to only trigger flyTo/setZoom when the PROP actually changes
  const prevCenterRef = useRef<[number, number]>(center);
  const prevZoomRef = useRef<number>(zoom);

  // Sync ref for callback
  useEffect(() => {
    onMapLoadRef.current = onMapLoad;
  }, [onMapLoad]);

  // 1. Initialization Effect: Runs exactly once
  useEffect(() => {
    if (map.current || !mapContainer.current) return;

    console.log('[MapEngine] Creating maplibre map instance...');
    map.current = new maplibregl.Map({
      container: mapContainer.current,
      style: "/carto-proxy/gl/dark-matter-gl-style/style.json",
      transformRequest: (url) => {
        if (url.includes("basemaps.cartocdn.com")) {
          const proxiedUrl = url.replace(/^https:\/\/(tiles\.)?basemaps\.cartocdn\.com/, "/carto-proxy");
          return {
            url: proxiedUrl,
          };
        }
      },
      center: center,
      zoom: zoom,
      maxBounds: [
        [68.1766, 6.7521],  // Southwest India
        [97.3953, 35.5042], // Northeast India
      ] as [[number, number], [number, number]],
      minZoom: 4,
      maxZoom: 18,
    });

    map.current.on("load", () => {
      console.log('[MapEngine] Map loaded successfully!');
      if (onMapLoadRef.current && map.current) {
        onMapLoadRef.current(map.current);
      }
    });
    
    map.current.on("error", (e) => {
      console.error('[MapEngine] Map error:', e);
    });

    return () => {
      if (map.current) {
        console.log('[MapEngine] Cleaning up map instance...');
        map.current.remove();
        map.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 2. Camera Update Effect: Synchronize only on EXPLICIT prop changes
  useEffect(() => {
    if (!map.current) return;
    
    const [prevLng, prevLat] = prevCenterRef.current;
    const [nextLng, nextLat] = center;

    // Only update if the prop's VALUE changed significantly compared to the LAST prop value
    if (prevLng !== nextLng || prevLat !== nextLat) {
      console.log('[MapEngine] Explicit center update detected, moving camera...');
      map.current.flyTo({ center: center, duration: 1000 });
      prevCenterRef.current = center;
    }
  }, [center]);

  useEffect(() => {
    if (!map.current) return;
    
    if (prevZoomRef.current !== zoom) {
      console.log('[MapEngine] Explicit zoom update detected:', zoom);
      map.current.setZoom(zoom);
      prevZoomRef.current = zoom;
    }
  }, [zoom]);

  return (
    <div className="absolute inset-0 w-full h-full overflow-hidden">
      <div ref={mapContainer} className="absolute inset-0 w-full h-full" />
    </div>
  );
};

export default MapEngine;
