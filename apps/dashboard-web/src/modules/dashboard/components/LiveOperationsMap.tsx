import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import React from 'react';

import { MapContainer, Marker, Popup, TileLayer, useMap, ZoomControl, Tooltip, Circle, Polygon } from 'react-leaflet';
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
  status: 'active' | 'idle' | 'alert' | 'offline';
  label: string;
  heading?: number;
}

export interface Geofence {
  id: string;
  name: string;
  type: 'circle' | 'polygon' | 'corridor';
  center_lat: number | null;
  center_lng: number | null;
  radius_m: number | null;
  polygon?: [number, number][] | null;
  polygon_wkt?: string | null;
  status: 'active' | 'inactive';
}

// Custom Marker Generators
const createMovingIcon = (heading: number = 0) => L.divIcon({
  className: 'custom-vehicle-marker',
  html: `
    <div class="relative flex items-center justify-center">
      <div class="absolute w-10 h-10 bg-emerald-500/20 rounded-full animate-ping duration-1000"></div>
      <div class="relative w-8 h-8 bg-white dark:bg-[#1a2b3c] rounded-lg shadow-xl flex items-center justify-center border border-emerald-500/30">
        <svg viewBox="0 0 24 24" class="w-5 h-5 text-emerald-500 fill-emerald-500 transition-transform duration-500" style="transform: rotate(${heading}deg)">
          <path d="M12 2L4.5 20.29l.71.71L12 18l6.79 3 .71-.71z"/>
        </svg>
      </div>
    </div>
  `,
  iconSize: [32, 32],
  iconAnchor: [16, 16],
});

const createIdleIcon = () => L.divIcon({
  className: 'custom-vehicle-marker',
  html: `
    <div class="relative flex items-center justify-center">
      <div class="w-4 h-4 bg-emerald-500 rounded-full border-2 border-white shadow-lg shadow-emerald-500/40"></div>
    </div>
  `,
  iconSize: [16, 16],
  iconAnchor: [8, 8],
});

const createOfflineIcon = () => L.divIcon({
  className: 'custom-vehicle-marker',
  html: `
    <div class="relative flex items-center justify-center">
      <div class="w-4 h-4 bg-gray-400 rounded-full border-2 border-white shadow-lg"></div>
    </div>
  `,
  iconSize: [16, 16],
  iconAnchor: [8, 8],
});

const createAlertIcon = () => L.divIcon({
  className: 'custom-vehicle-marker',
  html: `
    <div class="relative flex items-center justify-center">
      <div class="absolute w-8 h-8 bg-red-500/30 rounded-full animate-ping"></div>
      <div class="w-4 h-4 bg-red-500 rounded-full border-2 border-white shadow-lg shadow-red-500/40"></div>
    </div>
  `,
  iconSize: [16, 16],
  iconAnchor: [8, 8],
});

const ResizeHandler: React.FC = () => {
  const map = useMap();
  React.useEffect(() => {
    const container = map.getContainer();
    if (!container) return;

    // Recalculate map size immediately on mount
    map.invalidateSize();

    // Use ResizeObserver to detect layout, sidebar, devtools, or flexbox container size changes
    const resizeObserver = new ResizeObserver(() => {
      map.invalidateSize();
    });

    resizeObserver.observe(container);

    // Fallback timer for smooth CSS transitions (e.g. sidebar collapse animations)
    const timer = setTimeout(() => {
      map.invalidateSize();
    }, 300);

    return () => {
      resizeObserver.disconnect();
      clearTimeout(timer);
    };
  }, [map]);
  return null;
};

const MapController: React.FC<{ vehicles: VehicleMarker[]; selectedId: string | null }> = ({
  vehicles,
  selectedId,
}) => {
  const map = useMap();

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

const VehicleMarkers: React.FC<{ vehicles: VehicleMarker[] }> = ({ vehicles }) => {
  const map = useMap();

  return (
    <>
      {vehicles.filter(v => typeof v.lat === 'number' && typeof v.lng === 'number' && v.lat !== 0 && v.lng !== 0).map((v) => {
        const icon = v.status === 'active' ? createMovingIcon(v.heading) : v.status === 'alert' ? createAlertIcon() : v.status === 'offline' ? createOfflineIcon() : createIdleIcon();
        
        return (
          <Marker 
            key={v.id} 
            position={[v.lat, v.lng]} 
            icon={icon}
            eventHandlers={{
              dblclick: (e) => {
                map.flyTo(e.latlng, 18);
              }
            }}
          >
            <Tooltip 
              direction="top" 
              offset={[0, -10]} 
              opacity={1} 
              className="custom-vehicle-tooltip"
            >
              <div className="bg-white dark:bg-[#1a2b3c] px-2.5 py-1 rounded-lg shadow-2xl border border-gray-200 dark:border-white/10">
                <p className="text-[10px] font-black uppercase tracking-tighter text-gray-700 dark:text-white leading-none">
                  {v.label}
                </p>
              </div>
            </Tooltip>
            <Popup minWidth={200} className="custom-dispatch-popup">
              <div className="text-gray-900 p-2 min-w-[180px]">
                <p className="font-black text-[15px] uppercase tracking-tighter mb-1.5 border-b border-gray-100 pb-1.5 leading-none">{v.label}</p>
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <div className={`w-2.5 h-2.5 rounded-full ${v.status === 'active' ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]' : v.status === 'offline' ? 'bg-gray-400' : 'bg-emerald-500'}`} />
                    <p className="text-[11px] font-black text-gray-500 uppercase tracking-widest">{v.status}</p>
                  </div>
                  <div className="bg-gray-100 dark:bg-white/5 px-2 py-0.5 rounded text-[10px] font-mono text-gray-400">{v.id.substring(0, 8)}</div>
                </div>
                <div className="grid grid-cols-3 gap-2 bg-gray-50 dark:bg-white/[0.03] p-2.5 rounded-lg border border-gray-100 dark:border-white/5">
                  <div className="flex flex-col gap-0.5">
                    <p className="text-[9px] text-gray-400 uppercase font-black tracking-tighter">Lat</p>
                    <p className="text-[11px] font-mono font-bold text-gray-700">{v.lat.toFixed(4)}</p>
                  </div>
                  <div className="flex flex-col gap-0.5 border-l border-gray-200/50 dark:border-white/10 pl-2">
                    <p className="text-[9px] text-gray-400 uppercase font-black tracking-tighter">Lng</p>
                    <p className="text-[11px] font-mono font-bold text-gray-700">{v.lng.toFixed(4)}</p>
                  </div>
                  <div className="flex flex-col gap-0.5 border-l border-gray-200/50 dark:border-white/10 pl-2">
                    <p className="text-[9px] text-gray-400 uppercase font-black tracking-tighter">HDG</p>
                    <p className="text-[11px] font-mono font-bold text-emerald-600">{Math.round(v.heading || 0)}°</p>
                  </div>
                </div>
              </div>
            </Popup>
          </Marker>
        );
      })}
    </>
  );
};

const GeofenceLayers: React.FC<{ geofences: Geofence[] }> = ({ geofences }) => {
  return (
    <>
      {geofences.map((gf) => {
        const isActive = gf.status === 'active';
        const color = isActive ? '#3b82f6' : '#9ca3af';
        
        if (gf.type === 'circle' && gf.center_lat && gf.center_lng && gf.radius_m) {
          return (
            <Circle
              key={gf.id}
              center={[gf.center_lat, gf.center_lng]}
              radius={gf.radius_m}
              pathOptions={{
                fillColor: color,
                fillOpacity: 0.2,
                color: color,
                weight: 2,
                dashArray: isActive ? undefined : '5, 10'
              }}
            >
              <Tooltip sticky>{gf.name}</Tooltip>
            </Circle>
          );
        }
        
        let polygonCoords = gf.polygon;
        if (!polygonCoords && gf.polygon_wkt) {
          try {
            const cleanWkt = gf.polygon_wkt.trim();
            if (cleanWkt.startsWith('[')) {
              polygonCoords = JSON.parse(cleanWkt);
            } else if (cleanWkt.toUpperCase().startsWith('POLYGON')) {
              const match = cleanWkt.match(/\(\((.*)\)\)/);
              if (match && match[1]) {
                const points = match[1].split(',');
                polygonCoords = points.map(pt => {
                  const parts = pt.trim().split(/\s+/);
                  return [parseFloat(parts[1]), parseFloat(parts[0])] as [number, number]; // convert [lng, lat] WKT to [lat, lng] Leaflet
                });
              }
            }
          } catch (e) {
            console.error('Failed to parse polygon_wkt:', e);
          }
        }

        if (gf.type === 'polygon' && polygonCoords && polygonCoords.length > 0) {
          return (
            <Polygon
              key={gf.id}
              positions={polygonCoords.map(p => [p[0], p[1]] as [number, number])}
              pathOptions={{
                fillColor: color,
                fillOpacity: 0.15,
                color: color,
                weight: 2,
                dashArray: isActive ? undefined : '5, 10'
              }}
            >
              <Tooltip sticky>{gf.name}</Tooltip>
            </Polygon>
          );
        }
        
        return null;
      })}
    </>
  );
};

const LiveOperationsMap: React.FC<{
  vehicles: VehicleMarker[];
  geofences?: Geofence[];
  height?: string | number;
  showLegend?: boolean;
  selectedId?: string | null;
  className?: string;
  tileLayerUrl?: string;
}> = ({ 
  vehicles, 
  geofences = [],
  height = 500, 
  showLegend = true, 
  selectedId = null, 
  className = '',
  tileLayerUrl = "https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}"
}) => {
  const center: [number, number] = [20.5937, 78.9629]; // Default center (India)
  const [showGeofences, setShowGeofences] = React.useState(true);

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

            <div className="flex items-center gap-3 pt-1 border-t border-gray-100 dark:border-white/5">
              <div className="w-5 h-5 flex items-center justify-center">
                <div className="w-3 h-3 rounded-full bg-gray-600 dark:bg-gray-500" />
              </div>
              <span className="text-[11px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-tight leading-none">Offline or not moving</span>
            </div>

            {geofences && geofences.length > 0 && (
              <div className="flex items-center justify-between pt-3 mt-1 border-t border-gray-100 dark:border-white/5">
                <span className="text-[11px] font-bold text-gray-600 dark:text-gray-300 uppercase tracking-tight">Show Geofences</span>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input 
                    type="checkbox" 
                    checked={showGeofences} 
                    onChange={(e) => setShowGeofences(e.target.checked)}
                    className="sr-only peer"
                  />
                  <div className="w-8 h-4 bg-gray-200 dark:bg-white/10 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-3 after:w-3 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>
            )}
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
        
        {showGeofences && <GeofenceLayers geofences={geofences} />}

        {vehicles.length > 10 ? (
          <MarkerClusterGroup chunkedLoading maxClusterRadius={60}>
            <VehicleMarkers vehicles={vehicles} />
          </MarkerClusterGroup>
        ) : (
          <VehicleMarkers vehicles={vehicles} />
        )}
      </MapContainer>
    </div>
  );
};

export default LiveOperationsMap;
