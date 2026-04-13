import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { useEffect, useRef, useState } from "react";
import { MapContainer, Marker, Popup, TileLayer, useMap, useMapEvents } from "react-leaflet";
import markerIconUrl from "leaflet/dist/images/marker-icon.png";
import markerShadowUrl from "leaflet/dist/images/marker-shadow.png";

const markerIcon = L.icon({
  iconUrl: markerIconUrl,
  shadowUrl: markerShadowUrl,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
});

// Blue dot for current GPS position
const gpsIcon = L.divIcon({
  className: "",
  html: `<div style="position:relative;">
    <div style="width:20px;height:20px;border-radius:50%;background:rgba(19,127,236,0.2);position:absolute;top:-6px;left:-6px;animation:pulse-gps 2s infinite;"></div>
    <div style="width:14px;height:14px;border-radius:50%;background:#137fec;border:3px solid white;box-shadow:0 0 8px rgba(19,127,236,0.5);"></div>
  </div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
});

export interface MapLocation {
  lat: number;
  lng: number;
  address?: string;
}

interface LocationPickerMapProps {
  label: string;
  color: "primary" | "warning" | "success";
  value: MapLocation | null;
  onChange: (loc: MapLocation) => void;
  height?: string;
}

function ClickHandler({
  onChange,
  draggableMarker,
  markerPosition,
}: {
  onChange: (lat: number, lng: number) => void;
  draggableMarker?: boolean;
  markerPosition?: [number, number] | null;
}) {
  const map = useMap();

  useMapEvents({
    click(e) {
      onChange(e.latlng.lat, e.latlng.lng);
    },
  });

  // Fit bounds when marker position changes
  useEffect(() => {
    if (markerPosition) {
      map.setView(markerPosition, Math.max(map.getZoom(), 13));
    }
  }, [markerPosition, map]);

  return null;
}

function RecenterControl({ center, zoom }: { center: [number, number]; zoom?: number }) {
  const map = useMap();
  useEffect(() => {
    if (center) map.setView(center, zoom || map.getZoom());
  }, [center, zoom, map]);
  return null;
}

const reverseGeocode = async (lat: number, lng: number): Promise<string> => {
  try {
    const res = await fetch(
      `https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lng}&format=json`,
    );
    const data = await res.json();
    return data.display_name || `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
  } catch {
    return `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
  }
};

const getIpLocation = async (): Promise<{ lat: number; lng: number } | null> => {
  try {
    const res = await fetch("https://ip-api.com/json/?fields=lat,lon");
    const data = await res.json();
    if (data.lat && data.lon) return { lat: data.lat, lng: data.lon };
  } catch { /* ignore */ }
  // Fallback: default to a major city (Delhi)
  return null;
};

const colorMap: Record<string, string> = {
  primary: "#137fec",
  warning: "#f59e0b",
  success: "#10b981",
};

export default function LocationPickerMap({
  label,
  color,
  value,
  onChange,
  height = "h-64",
}: LocationPickerMapProps) {
  const dotColor = colorMap[color] || colorMap.primary;
  const [gpsLoading, setGpsLoading] = useState(false);
  const [gpsError, setGpsError] = useState("");
  const [gpsLocation, setGpsLocation] = useState<MapLocation | null>(null);
  const [gpsMode, setGpsMode] = useState<"ip" | "gps" | null>(null);
  const mapRef = useRef<HTMLDivElement>(null);

  const handleMapClick = async (lat: number, lng: number) => {
    const address = await reverseGeocode(lat, lng);
    onChange({ lat, lng, address });
  };

  const useMyLocation = () => {
    setGpsMode("ip");
    setGpsLoading(true);
    setGpsError("");
    fallbackToIp().finally(() => setGpsMode(null));
  };

  const tryBrowserGps = () => {
    setGpsMode("gps");
    if (!navigator.geolocation) {
      setGpsLoading(false);
      setGpsMode(null);
      setGpsError("GPS not available, using IP instead");
      setTimeout(() => setGpsError(""), 4000);
      fallbackToIp();
      return;
    }

    setGpsLoading(true);
    setGpsError("");

    navigator.geolocation.getCurrentPosition(
      async (pos) => {
        const lat = pos.coords.latitude;
        const lng = pos.coords.longitude;
        const address = await reverseGeocode(lat, lng);
        const loc = { lat, lng, address };
        setGpsLocation(loc);
        onChange(loc);
        setGpsLoading(false);
        setGpsMode(null);
      },
      () => {
        setGpsLoading(false);
        setGpsMode(null);
        // GPS failed → fall back to IP
        fallbackToIp();
      },
      { enableHighAccuracy: false, timeout: 5000, maximumAge: 0 },
    );
  };

  const fallbackToIp = async () => {
    const ipLoc = await getIpLocation();
    if (ipLoc) {
      const address = await reverseGeocode(ipLoc.lat, ipLoc.lng);
      const loc = { lat: ipLoc.lat, lng: ipLoc.lng, address };
      setGpsLocation(loc);
      onChange(loc);
      setGpsLoading(false);
    } else {
      setGpsLoading(false);
      setGpsError("Tap the map to pick a location");
      setTimeout(() => setGpsError(""), 4000);
    }
  };

  return (
    <div
      ref={mapRef}
      className={`rounded-xl border border-border-color overflow-hidden ${height} relative`}
    >
      <MapContainer
        center={value ? [value.lat, value.lng] : [28.6139, 77.209]}
        zoom={value ? 14 : 5}
        className="w-full h-full z-0"
        scrollWheelZoom
      >
        <TileLayer
          attribution='&copy; OpenStreetMap'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <ClickHandler onChange={handleMapClick} />
        {gpsLocation && (
          <RecenterControl center={[gpsLocation.lat, gpsLocation.lng]} zoom={15} />
        )}
        {value && (
          <Marker
            position={[value.lat, value.lng]}
            icon={gpsLocation && value.lat === gpsLocation.lat && value.lng === gpsLocation.lng ? gpsIcon : markerIcon}
          >
            <Popup>{value.address || "Selected location"}</Popup>
          </Marker>
        )}
      </MapContainer>

      {/* Overlay Label */}
      <div className="absolute top-3 left-3 z-[400] flex items-center gap-2 bg-surface-darker/90 backdrop-blur px-3 py-1.5 rounded-lg border border-border-color">
        <span className="w-2 h-2 rounded-full" style={{ backgroundColor: dotColor }} />
        <span className="text-xs text-white font-medium">{label}</span>
      </div>

      {/* GPS Button with Dropdown */}
      <div className="absolute top-3 right-3 z-[400]">
        {gpsMode === "ip" && (
          <div className="bg-surface-darker/90 backdrop-blur border border-border-color rounded-lg px-2.5 py-1.5 text-[10px] text-warning animate-pulse">
            🌐 Locating by IP...
          </div>
        )}
        {gpsMode === "gps" && (
          <div className="bg-surface-darker/90 backdrop-blur border border-border-color rounded-lg px-2.5 py-1.5 text-[10px] text-primary animate-pulse">
            📡 Requesting GPS...
          </div>
        )}
        {!gpsLoading && (
          <div className="relative group">
            <button
              type="button"
              className="bg-surface-darker/90 backdrop-blur border border-border-color rounded-lg p-2 text-text-muted hover:text-primary hover:border-primary/40 transition-colors"
              title="Use my location"
            >
              <span className="material-symbols-outlined text-sm">my_location</span>
            </button>
            {/* Dropdown */}
            <div className="absolute right-0 top-full mt-1 bg-surface-darker border border-border-color rounded-lg shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-[500] min-w-[180px]">
              <button
                type="button"
                onClick={useMyLocation}
                className="w-full flex items-center gap-2 px-3 py-2 text-left text-sm text-text-muted hover:bg-border-color hover:text-white first:rounded-t-lg last:rounded-b-lg"
              >
                <span className="material-symbols-outlined text-sm">public</span>
                <div>
                  <p className="text-white text-xs font-medium">IP Location</p>
                  <p className="text-[9px] text-text-muted">Instant, ~city-level</p>
                </div>
              </button>
              <button
                type="button"
                onClick={tryBrowserGps}
                className="w-full flex items-center gap-2 px-3 py-2 text-left text-sm text-text-muted hover:bg-border-color hover:text-white border-t border-border-color"
              >
                <span className="material-symbols-outlined text-sm">gps_fixed</span>
                <div>
                  <p className="text-white text-xs font-medium">GPS (Precise)</p>
                  <p className="text-[9px] text-text-muted">Needs permission</p>
                </div>
              </button>
            </div>
          </div>
        )}
        {gpsLoading && (
          <button
            type="button"
            disabled
            className="bg-surface-darker/90 backdrop-blur border border-border-color rounded-lg p-2 text-primary"
          >
            <span className="material-symbols-outlined text-sm animate-spin">progress_activity</span>
          </button>
        )}
      </div>

      {/* GPS Error */}
      {gpsError && (
        <div className="absolute top-14 right-3 z-[400] bg-danger/10 border border-danger/30 rounded-lg px-3 py-1.5 max-w-[220px]">
          <p className="text-[10px] text-danger">{gpsError}</p>
        </div>
      )}

      {/* Coordinates Display */}
      {value && (
        <div className="absolute bottom-3 left-3 z-[400] bg-surface-darker/90 backdrop-blur px-3 py-1.5 rounded-lg border border-border-color max-w-[80%]">
          <p className="text-[11px] text-text-muted truncate">
            {value.lat.toFixed(5)}, {value.lng.toFixed(5)}
          </p>
          {value.address && (
            <p className="text-[10px] text-text-muted/70 truncate mt-0.5">
              {value.address}
            </p>
          )}
        </div>
      )}

      {!value && (
        <div className="absolute inset-0 z-[400] flex items-center justify-center pointer-events-none">
          <div className="bg-surface-darker/80 backdrop-blur px-4 py-2 rounded-lg border border-border-color text-center">
            <span className="material-symbols-outlined text-text-muted text-2xl">
              location_on
            </span>
            <p className="text-xs text-text-muted mt-1">Click or use GPS to pick location</p>
          </div>
        </div>
      )}
    </div>
  );
}
