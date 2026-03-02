import {
    ArrowPathIcon,
    ChevronRightIcon,
    SignalIcon,
} from "@heroicons/react/24/outline";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import api from "../../../services/api";

interface TelemetryEvent {
  id: string;
  latitude: number;
  longitude: number;
  speed: number;
  heading: number;
  event_type: string;
  event_time: string;
}

export default function TripTelemetry() {
  const { id } = useParams<{ id: string }>();
  const { uiStrings: rawUiStrings } = useUIStrings("tracking");
  const uiStrings = rawUiStrings || {};
  const [events, setEvents] = useState<TelemetryEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [tripId, setTripId] = useState("");

  useEffect(() => {
    loadData();
  }, [id]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [tripData, telemetryData] = await Promise.all([
        api.getTrip(id!),
        api.getTripTelemetry(id!, 50),
      ]);
      setTripId(tripData.trip_id);
      setEvents(telemetryData.events || []);
    } catch (error) {
      console.error("Failed to load telemetry:", error);
    } finally {
      setLoading(false);
    }
  };

  const formatTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });
  };

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="p-6 pb-0">
        <div className="flex items-center justify-between mb-4">
          <div>
            <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
              <Link to="/admin/trips" className="hover:text-white">
                {uiStrings.trips_title || "Trips"}
              </Link>
              <ChevronRightIcon className="w-3 h-3" />
              <Link
                to={`/admin/trips/${id}`}
                className="hover:text-white font-mono"
              >
                {tripId}
              </Link>
              <ChevronRightIcon className="w-3 h-3" />
              <span className="text-primary">
                {uiStrings.breadcrumb_telemetry || "Telemetry"}
              </span>
            </div>
            <h1 className="text-2xl font-bold text-white flex items-center gap-2">
              {uiStrings.title_telemetry || "Trip Telemetry View"}
              <span className="w-2 h-2 rounded-full bg-success animate-pulse"></span>
            </h1>
          </div>
          <button
            onClick={loadData}
            className="p-2 bg-surface-dark border border-border-color text-text-muted hover:text-white rounded transition-colors"
          >
            <ArrowPathIcon className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 flex gap-6 p-6 pt-4 overflow-hidden">
        {/* Map */}
        <div className="flex-1 bg-surface-dark rounded-xl border border-border-color overflow-hidden relative">
          <div className="absolute top-4 left-4 z-10 bg-surface-darker/90 backdrop-blur border border-border-color rounded p-2">
            <div className="text-xs font-bold text-white mb-1">
              {uiStrings.label_live_tracking || "Continuous Map Polling"}
            </div>
            <div className="text-[10px] text-text-muted">
              {(
                uiStrings.vehicle_prefix || "Asset Sensor Unit: #{{id}}"
              ).replace("{{id}}", tripId)}
            </div>
          </div>
          <div className="w-full h-full bg-[#1b2631] flex items-center justify-center">
            <div className="text-center text-text-muted">
              <SignalIcon className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p className="text-sm">
                {uiStrings.label_live_map || "Live Rendering Map"}
              </p>
              <p className="text-xs">
                {uiStrings.map_integration_pending ||
                  "(GIS Map Server Starting...)"}
              </p>
            </div>
          </div>
        </div>

        {/* Telemetry Feed */}
        <div className="w-96 bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
          <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
            <h3 className="text-white font-semibold flex items-center gap-2">
              <SignalIcon className="w-4 h-4 text-text-muted" />
              {uiStrings.title_telem_feed || "Live Telemetry Feed"}
            </h3>
            <div className="w-2 h-2 rounded-full bg-success animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>
          </div>
          <div className="flex-1 overflow-y-auto">
            {loading ? (
              <div className="p-4 text-center text-text-muted">
                {uiStrings.msg_loading || "Loading..."}
              </div>
            ) : events.length === 0 ? (
              <div className="p-4 text-center text-text-muted">
                {uiStrings.msg_no_telem || "No telemetry data"}
              </div>
            ) : (
              <div className="divide-y divide-border-color">
                {events.map((event, index) => (
                  <div
                    key={event.id}
                    className={`p-4 hover:bg-surface-darker transition-colors cursor-pointer ${index === 0 ? "border-l-2 border-primary bg-surface-darker/30" : ""}`}
                  >
                    <div className="flex justify-between items-start mb-1">
                      <span className="text-xs font-mono text-primary font-bold">
                        {formatTime(event.event_time)}
                      </span>
                      {event.event_type && (
                        <span className="text-[10px] bg-success/10 text-success px-1.5 py-0.5 rounded border border-success/20">
                          {event.event_type.toUpperCase()}
                        </span>
                      )}
                    </div>
                    <div className="grid grid-cols-2 gap-y-1 gap-x-4 text-xs mt-2">
                      <div className="flex justify-between">
                        <span className="text-text-muted">
                          {uiStrings.label_speed || "Current Velocity"}:
                        </span>
                        <span className="text-white font-medium">
                          {event.speed?.toFixed(1)} km/h
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-text-muted">
                          {uiStrings.label_heading || "Heading"}:
                        </span>
                        <span className="text-white font-medium">
                          {event.heading?.toFixed(0)}°
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-text-muted">
                          {uiStrings.label_lat || "Lat"}:
                        </span>
                        <span className="text-white font-mono text-[10px]">
                          {event.latitude?.toFixed(6)}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-text-muted">
                          {uiStrings.label_lon || "Lon"}:
                        </span>
                        <span className="text-white font-mono text-[10px]">
                          {event.longitude?.toFixed(6)}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
