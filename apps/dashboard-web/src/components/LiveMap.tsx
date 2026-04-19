import L from "leaflet";
import "leaflet.markercluster/dist/MarkerCluster.css";
import "leaflet.markercluster/dist/MarkerCluster.Default.css";
import "leaflet/dist/leaflet.css";
import React, { useEffect, useMemo, useState } from "react";
import {
    MapContainer,
    Marker,
    Polyline,
    Popup,
    TileLayer,
} from "react-leaflet";
import { useUIStrings } from "../hooks/useUIStrings";
import { Trip, Vehicle } from "../types";

// Fix Leaflet's default icon issue with Webpack
import icon from "leaflet/dist/images/marker-icon.png";
import iconShadow from "leaflet/dist/images/marker-shadow.png";

let DefaultIcon = L.icon({
  iconUrl: icon,
  shadowUrl: iconShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
});

L.Marker.prototype.options.icon = DefaultIcon;

interface LiveMapProps {
  trips?: Trip[];
  vehicles?: Vehicle[];
  height?: string;
  onVehicleSelect?: (vehicle: Vehicle) => void;
}

const VehicleMarker: React.FC<{
  vehicle: Vehicle;
  onClick: () => void;
  uiStrings?: any;
}> = ({ vehicle, onClick, uiStrings }) => {
  const customIcon = useMemo(() => {
    const speed = vehicle.current_location?.speed || 0;
    const color = speed > 5 ? "#22c55e" : "#eab308"; // Green if moving, yellow if idle

    return L.divIcon({
      html: `
                <div style="position: relative;">
                    <div style="background-color: ${color}; width: 20px; height: 20px; border-radius: 50%; border: 3px solid white; box-shadow: 0 0 10px rgba(0,0,0,0.3); transform: rotate(${vehicle.current_location?.heading || 0}deg);">
                        <div style="position: absolute; top: -5px; left: 5px; width: 0; height: 0; border-left: 5px solid transparent; border-right: 5px solid transparent; border-bottom: 7px solid ${color};"></div>
                    </div>
                </div>
            `,
      className: "custom-vehicle-icon",
      iconSize: [20, 20],
      iconAnchor: [10, 10],
    });
  }, [vehicle.current_location?.speed, vehicle.current_location?.heading]);

  if (!vehicle.current_location) return null;

  return (
    <Marker
      position={[
        vehicle.current_location.latitude,
        vehicle.current_location.longitude,
      ]}
      icon={customIcon}
      eventHandlers={{ click: onClick }}
    >
      <Popup>
        <div className="p-1">
          <p className="text-xs font-bold">{vehicle.registration_number}</p>
          <p className="text-[10px] text-gray-500">
            {vehicle.current_location.speed} {uiStrings?.unit_speed || "km/h"}
          </p>
        </div>
      </Popup>
    </Marker>
  );
};

const LiveMap: React.FC<LiveMapProps> = ({
  trips = [],
  vehicles = [],
  height = "500px",
  onVehicleSelect,
}) => {
  const { uiStrings } = useUIStrings("comp_map");
  const [center, setCenter] = useState<[number, number]>([20.5937, 78.9629]); // India center

  // Extract first vehicle location for dependency tracking
  const firstVehicleLocation = vehicles[0]?.current_location;

  useEffect(() => {
    if (trips.length > 0 && trips[0].start_location) {
      setCenter([
        trips[0].start_location.latitude,
        trips[0].start_location.longitude,
      ]);
    } else if (firstVehicleLocation) {
      setCenter([
        firstVehicleLocation.latitude,
        firstVehicleLocation.longitude,
      ]);
    }
  }, [trips, firstVehicleLocation]); // Re-center only if initializing or first vehicle location changes

  return (
    <MapContainer
      center={center}
      zoom={5}
      attributionControl={false}
      style={{
        height: height,
        width: "100%",
        borderRadius: "0.75rem",
        zIndex: 0,
      }}
      className="shadow-lg border border-gray-200 dark:border-gray-700"
    >
      <TileLayer
        attribution='&copy; Google Maps'
        url="https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}&gl=IN"
      />

      {/* Clusters for Vehicles */}
      {vehicles.map((vehicle) => (
        <VehicleMarker
          key={vehicle.id}
          vehicle={vehicle}
          onClick={() => onVehicleSelect?.(vehicle)}
          uiStrings={uiStrings}
        />
      ))}

      {/* Trips Rendering */}
      {trips.map((trip) => (
        <React.Fragment key={trip.id}>
          <Marker
            position={[
              trip.start_location.latitude,
              trip.start_location.longitude,
            ]}
          >
            <Popup>
              <div className="p-2">
                <strong className="block text-sm font-bold text-gray-900">
                  {uiStrings?.label_trip || "Trip: "}
                  {trip.id.substring(0, 8)}
                </strong>
                <span className="text-xs text-blue-600 font-medium capitalize">
                  {trip.status}
                </span>
              </div>
            </Popup>
          </Marker>
          {trip.route && trip.route.length > 0 && (
            <Polyline
              positions={trip.route.map((pt) => [pt.latitude, pt.longitude])}
              color="#3b82f6"
              weight={3}
            />
          )}
        </React.Fragment>
      ))}
    </MapContainer>
  );
};

export default LiveMap;
