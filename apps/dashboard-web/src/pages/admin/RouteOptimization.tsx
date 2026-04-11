import {
    CheckCircleIcon,
    ClockIcon,
    CurrencyDollarIcon,
    MapIcon,
    PlusIcon,
    TruckIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import L from "leaflet";
import React, { useEffect, useState } from "react";
import { MapContainer, Marker, Popup, TileLayer } from "react-leaflet";
import { useUIStrings } from "../../hooks/useUIStrings";
import { api } from "../../services/api";
import iconRetinaUrl from "leaflet/dist/images/marker-icon-2x.png";
import iconUrl from "leaflet/dist/images/marker-icon.png";
import shadowUrl from "leaflet/dist/images/marker-shadow.png";

// Fix leaflet icon issues
delete (L.Icon.Default.prototype as any)._getIconUrl;

L.Icon.Default.mergeOptions({
  iconRetinaUrl,
  iconUrl,
  shadowUrl,
});

// Types
interface RouteRequest {
  id: string;
  name: string;
  status: "pending" | "processing" | "completed" | "failed";
  created_at: string;
  stops_count: number;
  vehicle_count: number;
  result?: any;
}

interface Stats {
  total: number;
  completed: number;
  saved_km: number;
  saved_cost: number;
}

const RouteOptimization: React.FC = () => {
  const [requests, setRequests] = useState<RouteRequest[]>([]);
  const [selectedRequest, setSelectedRequest] = useState<RouteRequest | null>(
    null,
  );
  const [stats, setStats] = useState<Stats>({
    total: 0,
    completed: 0,
    saved_km: 0,
    saved_cost: 0,
  });
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { uiStrings, loading: uiLoading } = useUIStrings("route_optimization");

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      // Fetch requests
      const res = await api.listOptimizationRequests({ limit: 20 });
      const data = (res as any)?.data ?? res;
      const items = data?.requests || data?.items || data?.data || data || [];
      const list = Array.isArray(items) ? items : [];
      setRequests(list);

      // Calculate stats from list
      const completed = list.filter(
        (r: any) => r.status === "completed",
      ).length;
      setStats({
        total: data?.total_count || data?.total || list.length,
        completed: completed,
        saved_km: completed * 15.2,
        saved_cost: completed * 45,
      });

      if (list.length > 0) {
        handleSelectRequest(list[0]);
      }
    } catch (error) {
      console.error("Failed to fetch optimization data", error);
      setError(
        uiStrings?.msg_load_failed || "Failed to load optimization requests.",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleSelectRequest = async (req: RouteRequest) => {
    if (req.status === "completed" && !req.result) {
      // Fetch result if missing
      try {
        // Fetch full request details which includes the result
        const fullRequest = await api.getOptimizationRequest(req.id);
        if (fullRequest && fullRequest.result) {
          req.result = fullRequest.result;
        }
      } catch (e) {
        console.warn("Could not fetch result details");
      }
    }
    setSelectedRequest(req);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "completed":
        return "text-emerald-400 bg-emerald-400/10 border-emerald-400/20";
      case "processing":
        return "text-blue-400 bg-blue-400/10 border-blue-400/20";
      case "failed":
        return "text-red-400 bg-red-400/10 border-red-400/20";
      default:
        return "text-slate-400 bg-slate-400/10 border-slate-400/20"; // Pending
    }
  };

  return (
    <div className="h-full flex flex-col bg-background-dark text-white space-y-6">
      {/* Header & Stats */}
      <div className="flex flex-col space-y-4">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">
              {uiStrings?.title}
            </h1>
            <p className="text-text-muted text-sm">{uiStrings?.subtitle}</p>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded-lg font-medium transition-all shadow-lg shadow-primary/20"
          >
            <PlusIcon className="w-5 h-5" />
            <span>{uiStrings?.btn_new}</span>
          </button>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <StatsCard
            label={uiStrings?.stat_total || "Total Requests"}
            value={stats.total}
            icon={<MapIcon className="w-6 h-6 text-primary" />}
          />
          <StatsCard
            label={uiStrings?.stat_completed || "Completed Operations"}
            value={stats.completed}
            icon={<CheckCircleIcon className="w-6 h-6 text-emerald-400" />}
          />
          <StatsCard
            label={uiStrings?.stat_distance || "Est. Distance Saved"}
            value={`${stats.saved_km.toFixed(1)} km`}
            icon={<TruckIcon className="w-6 h-6 text-blue-400" />}
          />
          <StatsCard
            label={uiStrings?.stat_savings || "Est. Cost Savings"}
            value={`$${stats.saved_cost.toFixed(2)}`}
            icon={<CurrencyDollarIcon className="w-6 h-6 text-yellow-400" />}
          />
        </div>
      </div>

      {/* Main Content: Split View */}
      <div className="flex-1 flex gap-6 overflow-hidden min-h-0">
        {/* Left: Request List */}
        <div className="w-1/3 flex flex-col bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="p-4 border-b border-border-color bg-surface-darker/50">
            <h3 className="font-semibold text-sm uppercase tracking-wide text-text-muted">
              {uiStrings?.title_recent}
            </h3>
          </div>
          <div className="flex-1 overflow-y-auto custom-scrollbar p-2 space-y-2">
            {loading || uiLoading ? (
              <div className="p-6 text-text-muted">
                {uiStrings?.msg_loading || "Loading requests..."}
              </div>
            ) : error ? (
              <div className="p-6 text-red-400">{error}</div>
            ) : (
              requests.map((req) => (
                <div
                  key={req.id}
                  onClick={() => handleSelectRequest(req)}
                  className={`p-4 rounded-lg cursor-pointer transition-all border ${
                    selectedRequest?.id === req.id
                      ? "bg-primary/5 border-primary/50 shadow-md"
                      : "bg-surface-dark border-transparent hover:bg-surface-darker hover:border-border-color"
                  }`}
                >
                  <div className="flex justify-between items-start mb-2">
                    <h4 className="font-medium truncate pr-2">{req.name}</h4>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded-full border uppercase tracking-wider font-bold ${getStatusColor(req.status)}`}
                    >
                      {req.status}
                    </span>
                  </div>
                  <div className="flex justify-between text-xs text-text-muted">
                    <span className="flex items-center gap-1">
                      <MapIcon className="w-3 h-3" /> {req.stops_count}{" "}
                      {uiStrings?.th_stops || "Stops"}
                    </span>
                    <span className="flex items-center gap-1">
                      <ClockIcon className="w-3 h-3" />{" "}
                      {format(new Date(req.created_at), "MMM d, HH:mm")}
                    </span>
                  </div>
                </div>
              ))
            )}
            {requests.length === 0 && !loading && !uiLoading && !error && (
              <div className="p-8 text-center text-text-muted">
                {uiStrings?.msg_no_requests}
              </div>
            )}
          </div>
        </div>

        {/* Right: Map / Details */}
        <div className="flex-1 bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col relative">
          {selectedRequest ? (
            <>
              {/* Map Header Overlay */}
              <div className="absolute top-4 left-4 z-[400] glass-card p-4 rounded-lg shadow-xl max-w-sm">
                <h3 className="font-bold text-lg mb-1">
                  {selectedRequest.name}
                </h3>
                <div className="space-y-1 text-sm text-text-muted">
                  <div className="flex justify-between">
                    <span>{uiStrings?.label_status}</span>
                    <span className="text-white capitalize">
                      {selectedRequest.status}
                    </span>
                  </div>
                  {selectedRequest.result && (
                    <>
                      <div className="flex justify-between">
                        <span>{uiStrings?.label_distance}</span>
                        <span className="text-white font-mono">
                          {selectedRequest.result.total_distance?.toFixed(1)} km
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>{uiStrings?.label_time}</span>
                        <span className="text-white font-mono">
                          {Math.round(selectedRequest.result.total_time)} min
                        </span>
                      </div>
                    </>
                  )}
                </div>
              </div>

              {/* Map Container */}
              <div className="flex-1 bg-[#1a1a1a] relative">
                <MapContainer
                  center={[20.5937, 78.9629]}
                  zoom={5}
                  style={{ height: "100%", width: "100%" }}
                  className="z-0"
                >
                  <TileLayer
                    attribution='&copy; Google Maps'
                    url="https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}&gl=IN"
                  />
                  {/* Render Markers and Routes if result exists */}
                  {selectedRequest.result?.routes?.map(
                    (route: any, idx: number) => (
                      <React.Fragment key={idx}>
                        {route.stops?.map((stop: any, stopIdx: number) => (
                          <Marker
                            key={stop.stop_id}
                            position={[
                              stop.location?.latitude || 0,
                              stop.location?.longitude || 0,
                            ]} // Fallback or need standard location field in stop result
                          >
                            <Popup>
                              Stop #{stop.sequence} <br />
                              Arrival:{" "}
                              {stop.arrival_time
                                ? format(new Date(stop.arrival_time), "HH:mm")
                                : "-"}
                            </Popup>
                          </Marker>
                        ))}
                      </React.Fragment>
                    ),
                  )}
                </MapContainer>
              </div>
            </>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-text-muted">
              <MapIcon className="w-16 h-16 opacity-20 mb-4" />
              <p>{uiStrings?.msg_select_request}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

const StatsCard = ({
  label,
  value,
  icon,
}: {
  label: string;
  value: string | number;
  icon: React.ReactNode;
}) => (
  <div className="bg-surface-dark border border-border-color rounded-xl p-4 flex items-center gap-4 hover:border-primary/30 transition-colors cursor-default">
    <div className="p-3 rounded-lg bg-surface-darker border border-border-color">
      {icon}
    </div>
    <div>
      <p className="text-xs text-text-muted uppercase tracking-wider font-semibold">
        {label}
      </p>
      <p className="text-xl font-bold text-white mt-0.5">{value}</p>
    </div>
  </div>
);

export default RouteOptimization;
