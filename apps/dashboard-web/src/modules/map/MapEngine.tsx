import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import React, { useEffect, useRef } from "react";

interface MapEngineProps {
  center?: [number, number];
  zoom?: number;
  onMapLoad?: (map: maplibregl.Map) => void;
}

const MapEngine: React.FC<MapEngineProps> = ({
  center = [78.9629, 20.5937],
  zoom = 5,
  onMapLoad,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<maplibregl.Map | null>(null);

  useEffect(() => {
    if (map.current || !mapContainer.current) return;

    map.current = new maplibregl.Map({
      container: mapContainer.current,
      style: "/carto-proxy/gl/dark-matter-gl-style/style.json",
      transformRequest: (url) => {
        if (url.startsWith("https://basemaps.cartocdn.com")) {
          return {
            url: url.replace("https://basemaps.cartocdn.com", "/carto-proxy"),
          };
        }
      },
      center: center,
      zoom: zoom,
    });

    map.current.on("load", () => {
      if (onMapLoad && map.current) {
        onMapLoad(map.current);
      }
    });

    return () => {
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, [center, zoom, onMapLoad]);

  return (
    <div className="absolute inset-0 w-full h-full overflow-hidden">
      <div ref={mapContainer} className="absolute inset-0 w-full h-full" />
      {/* Map Overlays can be added here */}
    </div>
  );
};

export default MapEngine;
