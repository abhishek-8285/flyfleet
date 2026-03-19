import React, { useEffect, useRef } from "react";
import maplibregl from "maplibre-gl";
import { useMapStore } from "./store";
import MapEngine from "./MapEngine";

const API_BASE_WS = process.env.REACT_APP_API_URL?.replace("http", "ws") || "ws://localhost:8084/api/v1";

const LiveMap: React.FC = () => {
  const mapRef = useRef<maplibregl.Map | null>(null);
  const { pulse, initializeWorker, destroyWorker } = useMapStore();
  const rAFRef = useRef<number | null>(null);

  useEffect(() => {
    const orgId = localStorage.getItem("org_id");
    let token = localStorage.getItem("token");
    if (!token) {
        const authPayload = localStorage.getItem("auth_payload");
        if (authPayload) {
            try { token = JSON.parse(authPayload).access_token || ""; } catch(e){}
        }
    }

    if (!orgId) return;

    // Boot the worker
    initializeWorker(orgId, token || "", API_BASE_WS);

    return () => {
      destroyWorker();
      if (rAFRef.current) cancelAnimationFrame(rAFRef.current);
    };
  }, [initializeWorker, destroyWorker]);

  // The Hardware-Accelerated Bridge 
  useEffect(() => {
    if (!mapRef.current || pulse === 0) return;
    
    // Throttle via rAF to ensure 60FPS sync
    if (rAFRef.current) cancelAnimationFrame(rAFRef.current);

    rAFRef.current = requestAnimationFrame(() => {
      const map = mapRef.current;
      if (!map) return;

      const source = map.getSource("vehicles") as maplibregl.GeoJSONSource;
      if (source) {
        // Direct read from non-reactive state store to bypass all React overhead
        const geoJSON = useMapStore.getState().currentGeoJSON;
        
        // Camera-Based Throttling / Viewport Cull
        // MapLibre's `setData` natively handles bounding box clipping, 
        // but we ensure we are literally just tossing the reference into the GL engine.
        source.setData(geoJSON);
      }
    });

  }, [pulse]);

  const handleMapLoad = (map: maplibregl.Map) => {
    mapRef.current = map;

    // Load simple placeholder sprites for status-based icon mapping
    // Note: In a real app we would load actual PNGs here. 
    // We'll use MapLibre's basic handling or a fallback. 
    // For this rewrite, we assume the style has "truck-moving" / "truck-idle".

    map.addSource("vehicles", {
      type: "geojson",
      data: { type: "FeatureCollection", features: [] },
      cluster: true,
      clusterMaxZoom: 9, // Native Mapbox clustering < 10 zoom
      clusterRadius: 50,
    });

    map.addLayer({
      id: "clusters",
      type: "circle",
      source: "vehicles",
      filter: ["has", "point_count"],
      paint: {
        "circle-color": [
          "step", ["get", "point_count"],
          "#51bbd6", 100,
          "#f1f075", 750,
          "#f28cb1"
        ],
        "circle-radius": ["step", ["get", "point_count"], 20, 100, 30, 750, 40],
        "circle-stroke-width": 1,
        "circle-stroke-color": "#fff"
      },
    });

    map.addLayer({
      id: "cluster-count",
      type: "symbol",
      source: "vehicles",
      filter: ["has", "point_count"],
      layout: {
        "text-field": "{point_count_abbreviated}",
        "text-font": ["monospace", "sans-serif"],
        "text-size": 12,
      },
    });

    // Hardware-Accelerated Symbol Layer for Active Vehicles
    map.addLayer({
      id: "unclustered-point",
      type: "symbol",
      source: "vehicles",
      filter: ["!", ["has", "point_count"]],
      layout: {
        // Use icon-rotate tied to the heading property in the telemetry data
        "icon-rotate": ["get", "heading"],
        "icon-rotation-alignment": "map", // Align to map projection
        "icon-allow-overlap": false,
        
        // Note: For now we'll use a standard fallback circle below as well if icons aren't loaded,
        // but this prepares the symbol engine for 'truck-[status]' icons.
        // "icon-image": ["concat", "truck-", ["get", "status"]],
        // "icon-size": 0.8
      },
    });

    // Fallback paint layer in case there's no sprite loaded for symbols
    map.addLayer({
      id: "unclustered-point-fallback",
      type: "circle",
      source: "vehicles",
      filter: ["!", ["has", "point_count"]],
      paint: {
        "circle-color": [
            "match",
            ["get", "status"],
            "moving", "#22c55e",
            /* default for idle */ "#eab308"
        ],
        "circle-radius": 8,
        "circle-stroke-width": 2,
        "circle-stroke-color": "#fff",
      },
    });
  };

  return (
    <div style={{ height: "100%", minHeight: "600px", width: "100%", borderRadius: "0.5rem", overflow: "hidden" }}>
      <MapEngine onMapLoad={handleMapLoad} />
    </div>
  );
};

export default LiveMap;
