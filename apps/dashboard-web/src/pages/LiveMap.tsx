import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import React, { useEffect, useState } from 'react';
import { MapContainer, Marker, Polyline, Popup, TileLayer } from 'react-leaflet';
import { Trip } from '../types';

// Fix Leaflet's default icon issue with Webpack
import icon from 'leaflet/dist/images/marker-icon.png';
import iconShadow from 'leaflet/dist/images/marker-shadow.png';

let DefaultIcon = L.icon({
  iconUrl: icon,
  shadowUrl: iconShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41]
});

L.Marker.prototype.options.icon = DefaultIcon;

interface LiveMapProps {
  trips?: Trip[];
  height?: string;
}

const LiveMap: React.FC<LiveMapProps> = ({ trips = [], height = '500px' }) => {
  const [center, setCenter] = useState<[number, number]>([20.5937, 78.9629]); // India center

  useEffect(() => {
    if (trips.length > 0 && trips[0].start_location) {
        // Center on the first active trip
        setCenter([trips[0].start_location.latitude, trips[0].start_location.longitude]);
    }
  }, [trips]);

  return (
    <MapContainer 
        center={center} 
        zoom={5} 
        style={{ height: height, width: '100%', borderRadius: '0.75rem', zIndex: 0 }}
        className="shadow-lg border border-gray-200 dark:border-gray-700"
    >
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      
      {trips.map(trip => (
        <React.Fragment key={trip.id}>
            {/* Start Marker */}
            <Marker position={[trip.start_location.latitude, trip.start_location.longitude]}>
                <Popup>
                    <div className="p-2">
                        <strong className="block text-sm font-bold text-gray-900">Trip: {trip.id.substring(0, 8)}</strong>
                        <span className="text-xs text-blue-600 font-medium capitalize">{trip.status}</span>
                        <p className="text-xs text-gray-600 mt-1">Driver: {trip.driver_id || 'Unassigned'}</p>
                    </div>
                </Popup>
            </Marker>

            {/* End Marker (if exists) */}
            {trip.end_location && (
                <Marker position={[trip.end_location.latitude, trip.end_location.longitude]} opacity={0.6}>
                    <Popup>Destination</Popup>
                </Marker>
            )}

            {/* Route Polyline (if waypoints exist) */}
            {trip.route && trip.route.length > 0 && (
                <Polyline 
                    positions={trip.route.map(pt => [pt.latitude, pt.longitude])}
                    color={trip.status === 'ongoing' ? '#3b82f6' : '#9ca3af'}
                    weight={3}
                />
            )}
        </React.Fragment>
      ))}
    </MapContainer>
  );
};

export default LiveMap;
