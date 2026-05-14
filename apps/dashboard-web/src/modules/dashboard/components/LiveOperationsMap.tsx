import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import React from 'react';

import { MapContainer, Marker, Popup, TileLayer, useMap, ZoomControl } from 'react-leaflet';
import MarkerClusterGroup from 'react-leaflet-cluster';

// Fix for default marker icon in Leaflet + React
import markerIcon from 'leaflet/dist/images/marker-icon.png';
import markerShadow from 'leaflet/dist/images/marker-shadow.png';

let DefaultIcon = L.icon({
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
});

L.Marker.prototype.options.icon = DefaultIcon;

interface VehicleMarker {
  id: string;
  lat: number;
  lng: number;
  status: 'active' | 'idle' | 'alert';
  label: string;
}

const ResizeHandler: React.FC = () => {
  const map = useMap();
  React.useEffect(() => {
    const timer = setTimeout(() => {
      map.invalidateSize();
    }, 300); // Wait for sidebar transition
    return () => clearTimeout(timer);
  }, [map]);
  return null;
};

const MapController: React.FC<{ vehicles: VehicleMarker[]; selectedId: string | null }> = ({
  vehicles,
  selectedId,
}) => {
  const map = useMap();
  // ... existing code ...

  React.useEffect(() => {
    if (selectedId) {
      const v = vehicles.find((m) => m.id === selectedId);
      if (v && typeof v.lat === 'number' && typeof v.lng === 'number' && v.lat !== 0 && v.lng !== 0) {
        map.flyTo([v.lat, v.lng], 18);
      }
    }
  }, [selectedId, vehicles, map]);

  return null;
};

const LiveOperationsMap: React.FC<{
  vehicles: VehicleMarker[];
  height?: string | number;
  showLegend?: boolean;
  selectedId?: string | null;
  className?: string;
  tileLayerUrl?: string;
}> = ({ 
  vehicles, 
  height = 500, 
  showLegend = true, 
  selectedId = null, 
  className = '',
  tileLayerUrl = "https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}"
}) => {
  const center: [number, number] = [20.5937, 78.9629]; // Default center (India)

  return (
    <div className={`bg-gray-900 overflow-hidden relative ${className}`} style={{ height }}>
      {showLegend && (
        <div className="absolute top-24 right-6 z-[1000] bg-white dark:bg-[#1a2b3c] border border-gray-200 dark:border-white/10 p-4 rounded-xl shadow-2xl min-w-[200px] animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <div className="w-5 h-5 flex items-center justify-center">
                <svg viewBox="0 0 24 24" className="w-4 h-4 text-emerald-500 fill-emerald-500 rotate-45"><path d="M12 2L4.5 20.29l.71.71L12 18l6.79 3 .71-.71z"/></svg>
              </div>
              <span className="text-[11px] font-bold text-gray-600 dark:text-gray-300 uppercase tracking-tight">Moving Vehicles</span>
            </div>
            
            <div className="flex items-center gap-3">
              <div className="w-5 h-5 flex items-center justify-center">
                <div className="w-3 h-3 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]" />
              </div>
              <span className="text-[11px] font-bold text-gray-600 dark:text-gray-300 uppercase tracking-tight">Online Vehicles</span>
            </div>

            <div className="flex items-center gap-3">
              <div className="w-5 h-5 flex items-center justify-center">
                <div className="w-3 h-3 rounded-[2px] bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]" />
              </div>
              <span className="text-[11px] font-bold text-gray-600 dark:text-gray-300 uppercase tracking-tight">Online Assets</span>
            </div>

            <div className="flex items-center gap-3 pt-1 border-t border-gray-100 dark:border-white/5">
              <div className="w-5 h-5 flex items-center justify-center">
                <div className="w-3 h-3 rounded-full bg-gray-600 dark:bg-gray-500" />
              </div>
              <span className="text-[11px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-tight leading-none">Offline or not moving vehicles</span>
            </div>

            <div className="flex items-center gap-3">
              <div className="w-5 h-5 flex items-center justify-center">
                <div className="w-3 h-3 rounded-[2px] bg-gray-600 dark:bg-gray-500" />
              </div>
              <span className="text-[11px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-tight">Offline Assets</span>
            </div>
          </div>
        </div>
      )}

      <MapContainer
        center={center}
        zoom={5}
        maxZoom={22}
        style={{ height: '100%', width: '100%' }}
        zoomControl={false}
        attributionControl={false}
        className="z-0"
      >
        <TileLayer
          attribution="&copy; Google Maps"
          url={tileLayerUrl}
          maxZoom={22}
          maxNativeZoom={20}
        />
        <ZoomControl position="bottomright" />
        <ResizeHandler />
        <MapController vehicles={vehicles} selectedId={selectedId} />

        <MarkerClusterGroup chunkedLoading maxClusterRadius={60}>
          {vehicles.filter(v => typeof v.lat === 'number' && typeof v.lng === 'number' && v.lat !== 0 && v.lng !== 0).map((v) => (
            <Marker key={v.id} position={[v.lat, v.lng]}>
              <Popup>
                <div className="text-gray-900">
                  <p className="font-bold">{v.label}</p>
                  <p className="text-xs">ID: {v.id}</p>
                  <p className="text-xs">Status: {v.status}</p>
                </div>
              </Popup>
            </Marker>
          ))}
        </MarkerClusterGroup>
      </MapContainer>
    </div>
  );
};

export default LiveOperationsMap;
