import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { useCallback, useEffect, useRef, useState } from 'react';
import { MapContainer, Marker, Popup, TileLayer, useMap } from 'react-leaflet';

interface Suggestion {
  display_name: string;
  lat: string;
  lon: string;
  address: {
    road?: string;
    city?: string;
    state?: string;
    country?: string;
  };
}

interface AddressInputProps {
  label: string;
  type: 'pickup' | 'drop';
  value: string;
  lat?: number;
  lng?: number;
  onLocationSelected: (lat: number, lng: number, address: string) => void;
}

const pickupIcon = L.divIcon({
  className: '',
  html: `<div style="width:32px;height:44px;position:relative;">
    <div style="width:32px;height:32px;border-radius:50%;background:#10b981;border:3px solid white;box-shadow:0 2px 12px rgba(16,185,129,0.5);display:flex;align-items:center;justify-content:center;">
      <span style="font-size:14px;">📍</span>
    </div>
    <div style="width:0;height:0;border-left:7px solid transparent;border-right:7px solid transparent;border-top:10px solid #10b981;margin-left:9px;margin-top:-2px;"></div>
  </div>`,
  iconSize: [32, 44],
  iconAnchor: [16, 44],
  popupAnchor: [0, -38],
});

const dropIcon = L.divIcon({
  className: '',
  html: `<div style="width:32px;height:44px;position:relative;">
    <div style="width:32px;height:32px;border-radius:50%;background:#f59e0b;border:3px solid white;box-shadow:0 2px 12px rgba(245,158,11,0.5);display:flex;align-items:center;justify-content:center;">
      <span style="font-size:14px;">🏁</span>
    </div>
    <div style="width:0;height:0;border-left:7px solid transparent;border-right:7px solid transparent;border-top:10px solid #f59e0b;margin-left:9px;margin-top:-2px;"></div>
  </div>`,
  iconSize: [32, 44],
  iconAnchor: [16, 44],
  popupAnchor: [0, -38],
});

function DraggableMarker({
  position,
  icon,
  onDragEnd,
  onMapClick,
}: {
  position: [number, number];
  icon: L.DivIcon;
  onDragEnd: (lat: number, lng: number) => void;
  onMapClick: (lat: number, lng: number) => void;
}) {
  const map = useMap();

  const handleMapClick = useCallback(
    (e: L.LeafletMouseEvent) => {
      onMapClick(e.latlng.lat, e.latlng.lng);
    },
    [onMapClick]
  );

  useEffect(() => {
    map.on('click', handleMapClick);
    return () => {
      map.off('click', handleMapClick);
    };
  }, [map, handleMapClick]);

  useEffect(() => {
    map.setView(position, Math.max(map.getZoom(), 14));
  }, [position, map]);

  return (
    <Marker
      position={position}
      icon={icon}
      draggable
      eventHandlers={{
        dragend: (e) => {
          const { lat, lng } = e.target.getLatLng();
          onDragEnd(lat, lng);
        },
      }}
    >
      <Popup>Drag to adjust location</Popup>
    </Marker>
  );
}

async function reverseGeocode(lat: number, lng: number): Promise<string> {
  try {
    const res = await fetch(
      `https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lng}&format=json`
    );
    const data = await res.json();
    return data.display_name || `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
  } catch {
    return `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
  }
}

export default function AddressInput({
  label,
  type,
  value,
  lat,
  lng,
  onLocationSelected,
}: AddressInputProps) {
  const [showMap, setShowMap] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [mapCenter, setMapCenter] = useState<[number, number]>([28.6139, 77.209]);
  const [loading, setLoading] = useState(false);
  const [gpsLoading, setGpsLoading] = useState(false);
  const [addressResolved, setAddressResolved] = useState(value);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const isPickup = type === 'pickup';
  const accentColor = isPickup ? 'text-emerald-400' : 'text-amber-400';
  const accentBorder = isPickup ? 'border-emerald-500/30' : 'border-amber-500/30';
  const accentBg = isPickup ? 'bg-emerald-500/10' : 'bg-amber-500/10';

  useEffect(() => {
    setAddressResolved(value);
    setSearchQuery(value);
  }, [value]);

  useEffect(() => {
    if (lat && lng) {
      setMapCenter([lat, lng]);
    }
  }, [lat, lng]);

  const searchAddress = async (q: string) => {
    if (q.length < 3) {
      setSuggestions([]);
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(
        `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(q)}&limit=5&addressdetails=1&countrycodes=in`
      );
      const data = await res.json();
      setSuggestions(data);
    } catch {
      setSuggestions([]);
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value;
    setSearchQuery(v);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    if (v.length >= 3) {
      debounceRef.current = setTimeout(() => searchAddress(v), 400);
    } else {
      setSuggestions([]);
    }
  };

  const handleSelectSuggestion = (s: Suggestion) => {
    const latNum = parseFloat(s.lat);
    const lngNum = parseFloat(s.lon);
    setMapCenter([latNum, lngNum]);
    setAddressResolved(s.display_name);
    setSearchQuery(s.display_name.split(',').slice(0, 3).join(','));
    setSuggestions([]);
    onLocationSelected(latNum, lngNum, s.display_name);
  };

  const handleMapPin = async (pinLat: number, pinLng: number) => {
    setMapCenter([pinLat, pinLng]);
    const address = await reverseGeocode(pinLat, pinLng);
    setAddressResolved(address);
    setSearchQuery(address.split(',').slice(0, 3).join(','));
    onLocationSelected(pinLat, pinLng, address);
  };

  const useGps = () => {
    if (!navigator.geolocation) return;
    setGpsLoading(true);
    navigator.geolocation.getCurrentPosition(
      async (pos) => {
        const pinLat = pos.coords.latitude;
        const pinLng = pos.coords.longitude;
        setMapCenter([pinLat, pinLng]);
        const address = await reverseGeocode(pinLat, pinLng);
        setAddressResolved(address);
        setSearchQuery(address.split(',').slice(0, 3).join(','));
        onLocationSelected(pinLat, pinLng, address);
        setGpsLoading(false);
      },
      () => setGpsLoading(false),
      { enableHighAccuracy: false, timeout: 8000 }
    );
  };

  const openMap = () => {
    if (lat && lng) {
      setMapCenter([lat, lng]);
    }
    setShowMap(true);
  };

  const confirmAndClose = () => {
    setShowMap(false);
    if (mapCenter[0] !== 0 && mapCenter[1] !== 0) {
      reverseGeocode(mapCenter[0], mapCenter[1]).then((addr) => {
        setAddressResolved(addr);
        setSearchQuery(addr.split(',').slice(0, 3).join(','));
        onLocationSelected(mapCenter[0], mapCenter[1], addr);
      });
    }
  };

  return (
    <>
      {/* Address Input Field */}
      <div className="relative">
        <div className="flex items-center gap-2 mb-2">
          <span className={`text-lg ${accentColor}`}>{isPickup ? '📍' : '🏁'}</span>
          <span className="text-xs font-semibold text-white uppercase tracking-wider">{label}</span>
        </div>
        <div
          onClick={openMap}
          className={`flex items-center gap-3 bg-surface-darker border ${accentBorder} ${accentBg} rounded-xl px-4 py-3 cursor-pointer hover:border-primary/40 transition-colors`}
        >
          <span className="material-symbols-outlined text-text-muted text-lg">location_on</span>
          {addressResolved ? (
            <span className="text-sm text-white truncate flex-1">
              {addressResolved.split(',').slice(0, 2).join(',')}
            </span>
          ) : (
            <span className="text-sm text-text-muted/50 flex-1">
              {isPickup ? 'Tap to set pickup location' : 'Tap to set drop location'}
            </span>
          )}
          <span className="material-symbols-outlined text-text-muted text-lg">chevron_right</span>
        </div>
      </div>

      {/* Map Modal */}
      {showMap && (
        <div className="fixed inset-0 z-[9999] flex flex-col bg-black/80">
          {/* Modal Header */}
          <div className="flex items-center justify-between px-4 py-3 bg-surface-darker border-b border-border-color">
            <div>
              <h3 className="text-white text-sm font-semibold">
                {isPickup ? '📍 Set Pickup Location' : '🏁 Set Drop Location'}
              </h3>
              <p className="text-[10px] text-text-muted mt-0.5">
                Search, tap the map, or drag the pin
              </p>
            </div>
            <button onClick={confirmAndClose} className="p-2 text-text-muted hover:text-white">
              <span className="material-symbols-outlined text-xl">close</span>
            </button>
          </div>

          {/* Search Bar */}
          <div className="px-4 py-2 bg-surface-darker border-b border-border-color">
            <div className="relative">
              <span className="absolute inset-y-0 left-3 flex items-center text-text-muted">
                <span className="material-symbols-outlined text-lg">search</span>
              </span>
              <input
                type="text"
                value={searchQuery}
                onChange={handleInputChange}
                placeholder="Search address..."
                className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-10 pr-4 py-2.5 outline-none focus:ring-1 focus:ring-primary placeholder-text-muted/50"
              />
              {loading && (
                <span className="absolute right-3 top-2.5 text-text-muted">
                  <span className="material-symbols-outlined text-sm animate-spin">
                    progress_activity
                  </span>
                </span>
              )}
            </div>

            {/* Suggestions */}
            {suggestions.length > 0 && (
              <div className="mt-2 bg-surface-dark border border-border-color rounded-lg overflow-hidden max-h-40 overflow-y-auto">
                {suggestions.map((s, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => handleSelectSuggestion(s)}
                    className="w-full px-4 py-2.5 text-left hover:bg-border-color transition-colors border-b border-border-color last:border-b-0"
                  >
                    <div className="flex items-center gap-2">
                      <span className="material-symbols-outlined text-text-muted text-sm">
                        location_on
                      </span>
                      <div className="min-w-0">
                        <p className="text-xs text-white truncate">
                          {s.address.road || s.display_name.split(',')[0]}
                        </p>
                        <p className="text-[10px] text-text-muted truncate">
                          {[s.address.city, s.address.state].filter(Boolean).join(', ')}
                        </p>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Map */}
          <div className="flex-1 relative">
            <MapContainer
              center={mapCenter}
              zoom={14}
              className="w-full h-full z-0"
              scrollWheelZoom
              attributionControl={false}
            >
              <TileLayer
                attribution="&copy;"
                url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
              />
              <DraggableMarker
                position={mapCenter}
                icon={isPickup ? pickupIcon : dropIcon}
                onDragEnd={(pinLat, pinLng) => {
                  setMapCenter([pinLat, pinLng]);
                  reverseGeocode(pinLat, pinLng).then((addr) => {
                    setAddressResolved(addr);
                    setSearchQuery(addr.split(',').slice(0, 3).join(','));
                    onLocationSelected(pinLat, pinLng, addr);
                  });
                }}
                onMapClick={(pinLat, pinLng) => {
                  setMapCenter([pinLat, pinLng]);
                  reverseGeocode(pinLat, pinLng).then((addr) => {
                    setAddressResolved(addr);
                    setSearchQuery(addr.split(',').slice(0, 3).join(','));
                    onLocationSelected(pinLat, pinLng, addr);
                  });
                }}
              />
            </MapContainer>

            {/* GPS Button - top right, above map */}
            <button
              type="button"
              onClick={useGps}
              disabled={gpsLoading}
              className="absolute top-3 right-3 z-[9999] bg-surface-darker/90 backdrop-blur border border-border-color rounded-lg p-2.5 text-text-muted hover:text-primary hover:border-primary/40 transition-colors disabled:opacity-50 shadow-lg"
              title="Use my location"
            >
              {gpsLoading ? (
                <span className="material-symbols-outlined text-lg animate-spin">
                  progress_activity
                </span>
              ) : (
                <span className="material-symbols-outlined text-lg">my_location</span>
              )}
            </button>

            {/* Confirm Button - above Leaflet's z-index (Leaflet goes up to ~z-1000) */}
            <div className="absolute bottom-0 left-0 right-0 z-[9999] pointer-events-none">
              <div className="px-4 pb-6 pt-8 bg-gradient-to-t from-black/80 to-transparent">
                <button
                  onClick={confirmAndClose}
                  className="w-full bg-primary hover:bg-primary-dark text-white font-semibold text-sm rounded-xl py-3.5 shadow-lg shadow-primary/30 transition-colors pointer-events-auto active:scale-[0.98]"
                >
                  {isPickup ? '✓  Confirm Pickup Location' : '✓  Confirm Drop Location'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
