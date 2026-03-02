import axios from "axios";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import React, { useEffect, useState } from "react";
import {
    MapContainer,
    Marker,
    Polyline,
    Popup,
    TileLayer,
} from "react-leaflet";
import { useParams } from "react-router-dom";

// Fix Leaflet icon
import icon from "leaflet/dist/images/marker-icon.png";
import iconShadow from "leaflet/dist/images/marker-shadow.png";

let DefaultIcon = L.icon({
  iconUrl: icon,
  shadowUrl: iconShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
});
L.Marker.prototype.options.icon = DefaultIcon;

const ShipmentTracking: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [shipment, setShipment] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [uiStrings, setUiStrings] = useState<any>(null);

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.portal || data.system_metadata?.ui_strings?.portal,
        ),
      )
      .catch((err) => console.error("Failed to fetch UI strings", err));
  }, []);

  useEffect(() => {
    const fetchShipment = async () => {
      try {
        // This endpoint is public as per backend implementation
        const res = await axios.get(`/api/v1/customer/shipments/${id}`);
        setShipment(res.data);
      } catch (err: any) {
        console.error("Failed to fetch shipment", err);
        setError("Shipment not found or access denied.");
      } finally {
        setLoading(false);
      }
    };

    if (id) fetchShipment();
  }, [id]);

  if (loading)
    return (
      <div className="min-h-screen bg-background-dark flex items-center justify-center text-white">
        <span className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin mr-2"></span>
        {uiStrings?.loading_data || "Loading shipment data..."}
      </div>
    );

  if (error || !shipment)
    return (
      <div className="min-h-screen bg-background-dark flex flex-col items-center justify-center p-4">
        <div className="w-16 h-16 bg-red-500/10 rounded-full flex items-center justify-center text-red-500 mb-4">
          <span className="material-symbols-outlined text-[32px]">
            error_outline
          </span>
        </div>
        <h1 className="text-xl font-bold text-white mb-2">
          {uiStrings?.tracking_error_title || "Tracking Error"}
        </h1>
        <p className="text-text-muted">{error}</p>
      </div>
    );

  const currentLocation = shipment.current_location
    ? [shipment.current_location.lat, shipment.current_location.lng]
    : null;
  const origin = shipment.origin
    ? [shipment.origin.lat, shipment.origin.lng]
    : null;

  return (
    <div className="min-h-screen bg-background-dark flex flex-col font-display">
      {/* Header */}
      <header className="h-16 bg-surface-darker border-b border-border-color flex items-center px-6">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white">
            <span className="material-symbols-outlined text-[20px]">
              local_shipping
            </span>
          </div>
          <div>
            <h1 className="text-sm font-bold text-white">
              {uiStrings?.brand_name || "FlyFleet Tracking"}
            </h1>
            <p className="text-xs text-text-muted">
              {uiStrings?.tracking_id_prefix || "Tracking ID:"}{" "}
              {shipment.tracking_number}
            </p>
          </div>
        </div>
        <div className="ml-auto">
          <span
            className={`px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wide 
                  ${
                    shipment.status === "DELIVERED"
                      ? "bg-green-500/10 text-green-500"
                      : shipment.status === "IN_TRANSIT"
                        ? "bg-blue-500/10 text-blue-500"
                        : "bg-gray-500/10 text-gray-500"
                  }`}
          >
            {shipment.status}
          </span>
        </div>
      </header>

      <div className="flex-1 flex flex-col md:flex-row">
        {/* Sidebar Info */}
        <div className="w-full md:w-80 bg-surface-dark border-r border-border-color p-6 space-y-6 z-10 flex-shrink-0">
          <div>
            <h3 className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-3">
              {uiStrings?.details_title || "Shipment Details"}
            </h3>
            <div className="space-y-4">
              <div className="flex gap-3">
                <div className="mt-0.5 text-primary">
                  <span className="material-symbols-outlined text-[18px]">
                    trip_origin
                  </span>
                </div>
                <div>
                  <p className="text-sm font-medium text-white">
                    {shipment.origin?.address || "N/A"}
                  </p>
                  <p className="text-xs text-text-muted">
                    {uiStrings?.label_origin || "Departure Node"}
                  </p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="mt-0.5 text-green-500">
                  <span className="material-symbols-outlined text-[18px]">
                    location_on
                  </span>
                </div>
                <div>
                  <p className="text-sm font-medium text-white">
                    {shipment.destination?.address || "N/A"}
                  </p>
                  <p className="text-xs text-text-muted">
                    {uiStrings?.label_destination || "Arrival Node"}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div className="pt-6 border-t border-border-color">
            <h3 className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-3">
              {uiStrings?.est_delivery_label || "Estimated Delivery"}
            </h3>
            <p className="text-lg font-bold text-white">
              {shipment.estimated_delivery
                ? new Date(shipment.estimated_delivery).toLocaleDateString(
                    undefined,
                    {
                      weekday: "long",
                      year: "numeric",
                      month: "long",
                      day: "numeric",
                    },
                  )
                : uiStrings?.pending_label || "Pending"}
            </p>
          </div>

          {shipment.cargo_details && (
            <div className="pt-6 border-t border-border-color">
              <h3 className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-3">
                {uiStrings?.cargo_label || "Cargo"}
              </h3>
              <div className="bg-surface-darker p-3 rounded-lg border border-border-color">
                <p className="text-sm text-white">
                  <span className="text-text-muted">
                    {uiStrings?.type_label || "Type:"}
                  </span>{" "}
                  {shipment.cargo_details.type}
                </p>
                <p className="text-sm text-white">
                  <span className="text-text-muted">
                    {uiStrings?.weight_label || "Weight:"}
                  </span>{" "}
                  {shipment.cargo_details.weight} kg
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Map */}
        <div className="flex-1 bg-surface-darker relative">
          {currentLocation ? (
            <MapContainer
              center={[
                currentLocation[0] as number,
                currentLocation[1] as number,
              ]}
              zoom={13}
              style={{ height: "100%", width: "100%" }}
              className="z-0"
            >
              <TileLayer
                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
              />
              <Marker
                position={[
                  currentLocation[0] as number,
                  currentLocation[1] as number,
                ]}
              >
                <Popup>
                  {uiStrings?.current_location_label || "Last Known Ping"}
                </Popup>
              </Marker>
              {origin && currentLocation && (
                <Polyline
                  positions={[
                    origin as [number, number],
                    currentLocation as [number, number],
                  ]}
                  color="#3b82f6"
                  dashArray="10, 10"
                />
              )}
            </MapContainer>
          ) : (
            <div className="flex items-center justify-center h-full text-text-muted">
              {uiStrings?.no_location_data || "No location data available"}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ShipmentTracking;
