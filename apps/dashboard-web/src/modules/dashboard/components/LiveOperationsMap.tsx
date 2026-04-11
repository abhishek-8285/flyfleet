import L from "leaflet";
import "leaflet/dist/leaflet.css";
import {
    MapContainer,
    Marker,
    Popup,
    TileLayer,
    ZoomControl,
} from "react-leaflet";
import MarkerClusterGroup from "react-leaflet-cluster";

// Fix for default marker icon in Leaflet + React
import markerIcon from "leaflet/dist/images/marker-icon.png";
import markerShadow from "leaflet/dist/images/marker-shadow.png";

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
  status: "active" | "idle" | "alert";
  label: string;
}

const LiveOperationsMap: React.FC<{ vehicles: VehicleMarker[] }> = ({
  vehicles,
}) => {
  const center: [number, number] = [20.5937, 78.9629]; // Default center (India)

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden h-[500px] relative">
      <div className="absolute top-4 left-4 z-[1000] bg-gray-900/90 border border-gray-700 p-3 rounded-lg backdrop-blur-md">
        <h3 className="text-white font-semibold text-sm mb-2">
          Fleet Overview
        </h3>
        <div className="flex gap-4">
          <div className="flex items-center gap-2">
            <span className="w-3 h-3 rounded-full bg-green-500"></span>
            <span className="text-xs text-gray-300">Active</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-3 h-3 rounded-full bg-yellow-500"></span>
            <span className="text-xs text-gray-300">Idle</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-3 h-3 rounded-full bg-red-500"></span>
            <span className="text-xs text-gray-300">Critical</span>
          </div>
        </div>
      </div>

      <MapContainer
        center={center}
        zoom={5}
        style={{ height: "100%", width: "100%" }}
        zoomControl={false}
        className="z-0"
      >
        <TileLayer
          attribution='&copy; Google Maps'
          url="https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}&gl=IN"
        />
        <ZoomControl position="bottomright" />

        <MarkerClusterGroup chunkedLoading maxClusterRadius={60}>
          {vehicles.map((v) => (
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
