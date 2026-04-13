import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { useEffect, useMemo, useState } from "react";
import { MapContainer, Marker, Polyline, Popup, TileLayer, useMap, useMapEvents } from "react-leaflet";
import markerIconUrl from "leaflet/dist/images/marker-icon.png";
import markerShadowUrl from "leaflet/dist/images/marker-shadow.png";

const pickupIcon = L.divIcon({
  className: "",
  html: `<div style="position:relative;width:28px;height:41px;">
    <div style="width:28px;height:28px;border-radius:50%;background:#10b981;border:3px solid white;box-shadow:0 2px 8px rgba(0,0,0,0.4);display:flex;align-items:center;justify-content:center;">
      <span style="font-size:12px;">📍</span>
    </div>
    <div style="width:0;height:0;border-left:6px solid transparent;border-right:6px solid transparent;border-top:8px solid #10b981;margin-left:8px;margin-top:-2px;"></div>
  </div>`,
  iconSize: [28, 41],
  iconAnchor: [14, 41],
  popupAnchor: [0, -36],
});

const dropIcon = L.divIcon({
  className: "",
  html: `<div style="position:relative;width:28px;height:41px;">
    <div style="width:28px;height:28px;border-radius:50%;background:#f59e0b;border:3px solid white;box-shadow:0 2px 8px rgba(0,0,0,0.4);display:flex;align-items:center;justify-content:center;">
      <span style="font-size:12px;">🏁</span>
    </div>
    <div style="width:0;height:0;border-left:6px solid transparent;border-right:6px solid transparent;border-top:8px solid #f59e0b;margin-left:8px;margin-top:-2px;"></div>
  </div>`,
  iconSize: [28, 41],
  iconAnchor: [14, 41],
  popupAnchor: [0, -36],
});

function FitBounds({ points }: { points: [number, number][] }) {
  const map = useMap();
  useEffect(() => {
    if (points.length >= 2) {
      map.fitBounds(points, { padding: [50, 50], maxZoom: 14 });
    }
  }, [points, map]);
  return null;
}

export interface RouteMapPoint {
  lat: number;
  lng: number;
  address?: string;
}

interface RouteMapProps {
  pickup: RouteMapPoint | null;
  drop: RouteMapPoint | null;
  onPickupChange: (pt: RouteMapPoint) => void;
  onDropChange: (pt: RouteMapPoint) => void;
  height?: string;
}

export default function RouteMap({
  pickup,
  drop,
  onPickupChange,
  onDropChange,
  height = "h-80",
}: RouteMapProps) {
  const defaultCenter: [number, number] = [28.6139, 77.209];
  const points = useMemo(() => {
    const pts: [number, number][] = [];
    if (pickup) pts.push([pickup.lat, pickup.lng]);
    if (drop) pts.push([drop.lat, drop.lng]);
    return pts;
  }, [pickup, drop]);

  const center = points.length > 0 ? points[0] : defaultCenter;
  const zoom = points.length >= 2 ? undefined : 5;

  const handleMapClick = async (lat: number, lng: number) => {
    const address = await reverseGeocode(lat, lng);
    const pt = { lat, lng, address };
    if (!pickup) {
      onPickupChange(pt);
    } else if (!drop) {
      onDropChange(pt);
    } else {
      // Reset: new click sets pickup, clears drop
      onPickupChange(pt);
      onDropChange({ lat: 0, lng: 0, address: "" });
    }
  };

  return (
    <div className={`rounded-xl border border-border-color overflow-hidden ${height} relative`}>
      <MapContainer center={center} zoom={zoom} className="w-full h-full z-0" scrollWheelZoom>
        <TileLayer
          attribution='&copy; OpenStreetMap'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <MapClickHandler onClick={handleMapClick} />
        {points.length >= 2 && <FitBounds points={points} />}

        {/* Route line */}
        {pickup && drop && (
          <Polyline
            positions={[[pickup.lat, pickup.lng], [drop.lat, drop.lng]]}
            color="#137fec"
            weight={4}
            dashArray="8, 8"
            opacity={0.8}
          />
        )}

        {/* Pickup marker */}
        {pickup && pickup.lat !== 0 && (
          <Marker position={[pickup.lat, pickup.lng]} icon={pickupIcon}>
            <Popup>
              <div className="text-xs">
                <p className="font-semibold text-emerald-600">📍 Pickup</p>
                <p className="text-text-muted mt-1 truncate max-w-[200px]">{pickup.address}</p>
              </div>
            </Popup>
          </Marker>
        )}

        {/* Drop marker */}
        {drop && drop.lat !== 0 && (
          <Marker position={[drop.lat, drop.lng]} icon={dropIcon}>
            <Popup>
              <div className="text-xs">
                <p className="font-semibold text-amber-600">🏁 Drop</p>
                <p className="text-text-muted mt-1 truncate max-w-[200px]">{drop.address}</p>
              </div>
            </Popup>
          </Marker>
        )}
      </MapContainer>

      {/* Map overlay hint */}
      {!pickup && (
        <div className="absolute inset-0 z-[400] flex items-center justify-center pointer-events-none">
          <div className="bg-surface-darker/80 backdrop-blur px-4 py-2 rounded-lg border border-border-color text-center">
            <span className="material-symbols-outlined text-text-muted text-2xl">location_on</span>
            <p className="text-xs text-text-muted mt-1">Click to set pickup location</p>
          </div>
        </div>
      )}
      {pickup && !drop && (
        <div className="absolute bottom-3 left-3 right-3 z-[400]">
          <div className="bg-surface-darker/90 backdrop-blur border border-border-color rounded-lg px-3 py-2 text-center">
            <p className="text-xs text-amber-400 font-medium">Now click to set drop location</p>
          </div>
        </div>
      )}

      {/* GPS button */}
      <GpsButton
        onLocation={(loc) => {
          if (!pickup) onPickupChange(loc);
          else if (!drop) onDropChange(loc);
          else onPickupChange(loc);
        }}
      />
    </div>
  );
}

function MapClickHandler({ onClick }: { onClick: (lat: number, lng: number) => void }) {
  useMapEvents({
    click(e) {
      onClick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

async function reverseGeocode(lat: number, lng: number): Promise<string> {
  try {
    const res = await fetch(
      `https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lng}&format=json`,
    );
    const data = await res.json();
    return data.display_name || `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
  } catch {
    return `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
  }
}

function GpsButton({ onLocation }: { onLocation: (loc: { lat: number; lng: number; address?: string }) => void }) {
  const [loading, setLoading] = useState(false);

  const useGps = () => {
    if (!navigator.geolocation) return;
    setLoading(true);
    navigator.geolocation.getCurrentPosition(
      async (pos) => {
        const lat = pos.coords.latitude;
        const lng = pos.coords.longitude;
        const address = await reverseGeocode(lat, lng);
        onLocation({ lat, lng, address });
        setLoading(false);
      },
      () => setLoading(false),
      { enableHighAccuracy: false, timeout: 8000 },
    );
  };

  return (
    <button
      type="button"
      onClick={useGps}
      disabled={loading}
      className="absolute top-3 right-3 z-[400] bg-surface-darker/90 backdrop-blur border border-border-color rounded-lg p-2 text-text-muted hover:text-primary hover:border-primary/40 transition-colors disabled:opacity-50"
      title="Use my location"
    >
      {loading ? (
        <span className="material-symbols-outlined text-sm animate-spin">progress_activity</span>
      ) : (
        <span className="material-symbols-outlined text-sm">my_location</span>
      )}
    </button>
  );
}
