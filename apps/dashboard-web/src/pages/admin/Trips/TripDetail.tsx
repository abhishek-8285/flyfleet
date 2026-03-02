import {
    ChevronRightIcon,
    MapPinIcon,
    PencilIcon,
    SignalIcon,
    TruckIcon,
    UserIcon,
} from "@heroicons/react/24/outline";
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import api from "../../../services/api";

interface Trip {
  id: string;
  trip_id: string;
  vehicle_id: string;
  driver_id?: string;
  start_location: string;
  end_location: string;
  start_time: string;
  end_time?: string;
  distance: number;
  status: string;
}

const statusColors: Record<string, string> = {
  scheduled: "bg-primary/10 text-primary border-primary/20",
  in_progress: "bg-warning/10 text-warning border-warning/20",
  completed: "bg-success/10 text-success border-success/20",
  cancelled: "bg-danger/10 text-danger border-danger/20",
};

export default function TripDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("tracking");
  const { uiStrings: commonStrings } = useUIStrings("common");
  const uiStrings = rawUiStrings || {};
  const [trip, setTrip] = useState<Trip | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadTrip();
  }, [id]);

  const loadTrip = async () => {
    try {
      const data = await api.getTrip(id!);
      setTrip(data);
    } catch (error) {
      console.error("Failed to load trip:", error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  if (loading) {
    return (
      <div className="p-6 text-center text-text-muted">
        {uiStrings.msg_loading || "Loading..."}
      </div>
    );
  }

  if (!trip) {
    return (
      <div className="p-6 text-center text-danger">
        {uiStrings.msg_trip_not_found || "Trip not found"}
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <Link to="/admin/trips" className="hover:text-white">
              {uiStrings.trips_title || "Trips"}
            </Link>
            <ChevronRightIcon className="w-3 h-3" />
            <span className="text-white font-medium">{trip.trip_id}</span>
          </div>
          <h1 className="text-2xl font-bold text-white">
            {uiStrings.trip_detail_title || "Dispatch Command View"}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to={`/admin/trips/${id}/telemetry`}
            className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white text-sm rounded flex items-center gap-2 transition-colors"
          >
            <SignalIcon className="w-4 h-4" />
            {uiStrings.btn_telemetry || "Telemetry"}
          </Link>
          <Link
            to={`/admin/trips/${id}/edit`}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded flex items-center gap-2 transition-colors"
          >
            <PencilIcon className="w-4 h-4" />
            {commonStrings?.btn_edit || uiStrings.btn_edit_trip || "Edit Trip"}
          </Link>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_total_distance || "Total Distance"}
          </div>
          <div className="text-2xl font-bold text-white">
            {trip.distance?.toFixed(1) || "---"}{" "}
            <span className="text-sm font-normal text-text-muted">
              {uiStrings.miles_label || "km"}
            </span>
          </div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_status || "Status"}
          </div>
          <span
            className={`inline-flex items-center px-2 py-1 rounded text-sm font-bold border ${statusColors[trip.status]}`}
          >
            {trip.status?.toUpperCase().replace("_", " ")}
          </span>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_start_time || "Start Time"}
          </div>
          <div className="text-lg font-medium text-white">
            {formatDate(trip.start_time)}
          </div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_end_time || "End Time"}
          </div>
          <div className="text-lg font-medium text-white">
            {trip.end_time ? formatDate(trip.end_time) : "---"}
          </div>
        </div>
      </div>

      {/* Details */}
      <div className="grid grid-cols-2 gap-6">
        {/* Map Placeholder */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden h-96">
          <div className="absolute top-4 left-4 z-10 bg-surface-darker/90 backdrop-blur border border-border-color rounded p-2">
            <div className="text-xs font-bold text-white mb-1">
              {(
                uiStrings.route_label || "Planned Execution Path: {{id}}"
              ).replace("{{id}}", trip.trip_id)}
            </div>
            <div className="text-[10px] text-text-muted">
              {trip.start_location} → {trip.end_location}
            </div>
          </div>
          <div className="w-full h-full bg-[#1b2631] flex items-center justify-center">
            <div className="text-center text-text-muted">
              <MapPinIcon className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p className="text-sm">
                {uiStrings.map_view_label || "Global GPS View"}
              </p>
              <p className="text-xs">
                {uiStrings.map_integration_pending ||
                  "(GIS Map Server Starting...)"}
              </p>
            </div>
          </div>
        </div>

        {/* Info Panel */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="flex border-b border-border-color bg-surface-darker">
            <button className="px-6 py-3 text-sm font-medium text-white border-b-2 border-primary bg-surface-dark">
              {uiStrings.tab_general_info || "Core Dispatch Details"}
            </button>
            <button className="px-6 py-3 text-sm font-medium text-text-muted hover:text-white">
              {uiStrings.tab_vehicle_specs || "Asset Specifications"}
            </button>
          </div>
          <div className="p-5 space-y-6">
            <div>
              <h3 className="text-xs font-bold text-text-muted uppercase mb-3">
                {uiStrings.section_shipment_details ||
                  "Freight & Route Context"}
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 bg-surface-darker rounded border border-border-color">
                  <div className="text-[10px] text-text-muted mb-1">
                    {uiStrings.label_origin || "Origin"}
                  </div>
                  <div className="text-sm font-medium text-white">
                    {trip.start_location}
                  </div>
                </div>
                <div className="p-3 bg-surface-darker rounded border border-border-color">
                  <div className="text-[10px] text-text-muted mb-1">
                    {uiStrings.label_destination || "Destination"}
                  </div>
                  <div className="text-sm font-medium text-white">
                    {trip.end_location}
                  </div>
                </div>
              </div>
            </div>
            <div>
              <h3 className="text-xs font-bold text-text-muted uppercase mb-3">
                {uiStrings.section_assignment || "Labor & Asset Assignment"}
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 bg-surface-darker rounded border border-border-color flex items-center gap-3">
                  <TruckIcon className="w-5 h-5 text-primary" />
                  <div>
                    <div className="text-[10px] text-text-muted mb-0.5">
                      {uiStrings.label_vehicle || "Vehicle"}
                    </div>
                    <div className="text-sm font-medium text-white font-mono">
                      {trip.vehicle_id?.slice(0, 8)}...
                    </div>
                  </div>
                </div>
                <div className="p-3 bg-surface-darker rounded border border-border-color flex items-center gap-3">
                  <UserIcon className="w-5 h-5 text-success" />
                  <div>
                    <div className="text-[10px] text-text-muted mb-0.5">
                      {uiStrings.label_driver || "Driver"}
                    </div>
                    <div className="text-sm font-medium text-white font-mono">
                      {trip.driver_id?.slice(0, 8) ||
                        uiStrings.not_assigned ||
                        "Not assigned"}
                      ...
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
