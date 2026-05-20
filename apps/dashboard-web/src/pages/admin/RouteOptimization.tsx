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
import { MapContainer, Marker, Popup, TileLayer, Polyline, Polygon, useMap } from "react-leaflet";
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

// Custom sequenced stop marker with number indicator (Google Maps style)
const createSequencedPin = (sequenceNum: number, colorIndex: number) => {
  const colors = [
    "#3b82f6", // Blue
    "#10b981", // Emerald
    "#6366f1", // Indigo
    "#f59e0b", // Amber
    "#ec4899", // Pink
    "#06b6d4", // Cyan
  ];
  const color = colors[colorIndex % colors.length];
  return L.divIcon({
    className: "custom-stop-marker",
    html: `
      <div style="transform: translate(-17px, -42px);">
        <svg width="34" height="42" viewBox="0 0 34 42" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0px 3px 5px rgba(0, 0, 0, 0.35));">
          <path d="M17 0C7.6 0 0 7.6 0 17C0 27.7 14.9 40.8 16.2 41.9C16.4 42.1 16.7 42.2 17 42.2C17.3 42.2 17.6 42.1 17.8 41.9C19.1 40.8 34 27.7 34 17C34 7.6 26.4 0 17 0Z" fill="${color}"/>
          <circle cx="17" cy="17" r="9" fill="white"/>
          <text x="17" y="20.5" fill="${color}" font-size="10" font-weight="900" text-anchor="middle" font-family="Inter, system-ui, sans-serif">${sequenceNum}</text>
        </svg>
      </div>
    `,
    iconSize: [34, 42],
    iconAnchor: [0, 0],
  });
};

// Custom depot marker with home/warehouse icon (Google Maps style)
const createDepotPin = () => {
  return L.divIcon({
    className: "custom-depot-marker",
    html: `
      <div style="transform: translate(-18px, -44px);">
        <svg width="36" height="44" viewBox="0 0 36 44" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0px 3px 6px rgba(0, 0, 0, 0.4));">
          <path d="M18 0C8.1 0 0 8.1 0 18C0 29.3 15.8 42.8 17.2 44.0C17.4 44.2 17.7 44.3 18 44.3C18.3 44.3 18.6 44.2 18.8 44.0C20.2 42.8 36 29.3 36 18C36 8.1 27.9 0 18 0Z" fill="#ea580c"/>
          <circle cx="18" cy="18" r="10" fill="white"/>
          <path d="M18 12L13 16.5H15.5V21H20.5V16.5H23L18 12Z" fill="#ea580c"/>
        </svg>
      </div>
    `,
    iconSize: [36, 44],
    iconAnchor: [0, 0],
  });
};

// Custom Google Maps style floating route badge (displays travel time and distance on top of path)
const createRouteBadgeIcon = (timeMinutes: number, distanceKM: number) => {
  return L.divIcon({
    className: "custom-route-badge",
    html: `
      <div style="transform: translate(-50%, -100%); margin-top: -8px;">
        <div style="
          background: white;
          color: #1f2937;
          font-family: 'Inter', system-ui, sans-serif;
          font-size: 11px;
          padding: 6px 10px;
          border-radius: 8px;
          box-shadow: 0px 4px 10px rgba(0, 0, 0, 0.22);
          border: 1px solid #e5e7eb;
          display: flex;
          flex-direction: column;
          align-items: center;
          white-space: nowrap;
          pointer-events: none;
        ">
          <div style="display: flex; align-items: center; gap: 4px; font-weight: 800;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round" style="color: #2563eb;">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
            <span style="color: #2563eb;">${Math.round(timeMinutes)} min</span>
          </div>
          <div style="color: #4b5563; font-size: 9.5px; font-weight: 700; margin-top: 2px;">
            ${distanceKM.toFixed(1)} km
          </div>
        </div>
        <div style="
          width: 0;
          height: 0;
          border-left: 5px solid transparent;
          border-right: 5px solid transparent;
          border-top: 5px solid white;
          margin: 0 auto;
          margin-top: -1px;
        "></div>
      </div>
    `,
    iconSize: [1, 1],
    iconAnchor: [0, 0],
  });
};

const polylineColors = [
  "#3b82f6", // Blue
  "#10b981", // Emerald
  "#8b5cf6", // Violet
  "#f59e0b", // Amber
  "#ec4899", // Pink
  "#06b6d4", // Cyan
];

// Helper to fit map bounds to depots + stops + route paths
const MapBoundsController: React.FC<{ request: any }> = ({ request }) => {
  const map = useMap();
  
  useEffect(() => {
    if (!request) return;
    
    const bounds: [number, number][] = [];
    
    // Add depots to bounds
    if (request.depots && request.depots.length > 0) {
      request.depots.forEach((d: any) => {
        if (d.location?.latitude && d.location?.longitude) {
          bounds.push([d.location.latitude, d.location.longitude]);
        }
      });
    }
    
    // Add stops to bounds
    if (request.stops && request.stops.length > 0) {
      request.stops.forEach((s: any) => {
        if (s.latitude && s.longitude) {
          bounds.push([s.latitude, s.longitude]);
        }
      });
    }
    
    // Add polyline coordinates if available
    if (request.result?.routes) {
      request.result.routes.forEach((route: any) => {
        if (route.polyline && route.polyline.length > 0) {
          route.polyline.forEach((coord: any) => {
            bounds.push([coord[0], coord[1]]);
          });
        }
      });
    }
    
    if (bounds.length > 0) {
      map.fitBounds(bounds, { padding: [50, 50] });
    }
  }, [request, map]);
  
  return null;
};

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
  const [recalculatingVehicles, setRecalculatingVehicles] = useState<Set<string>>(new Set());

  // Form states for VRP request creation
  const [formName, setFormName] = useState("Mumbai Fast-Track Delivery Tour");
  const [formDesc, setFormDesc] = useState("Optimal retail shipment distribution in Mumbai districts.");
  const [isDemoLoaded, setIsDemoLoaded] = useState(false);
  const [formSubmitting, setFormSubmitting] = useState(false);

  // Premium non-blocking toast notifications
  const [toastMsg, setToastMsg] = useState<{ text: string; type: "success" | "error" } | null>(null);

  const showToast = (text: string, type: "success" | "error" = "success") => {
    setToastMsg({ text, type });
    setTimeout(() => {
      setToastMsg(null);
    }, 4000);
  };

  useEffect(() => {
    fetchData();
  }, []);

  // Live status polling if any route requests are still pending or processing
  useEffect(() => {
    const hasActiveJobs = requests.some(r => r.status === "processing" || r.status === "pending");
    if (!hasActiveJobs) return;

    const interval = setInterval(() => {
      fetchDataSilent();
    }, 2000);

    return () => clearInterval(interval);
  }, [requests]);

  const fetchDataSilent = async () => {
    try {
      const res = await api.listOptimizationRequests({ limit: 20 });
      const data = (res as any)?.data ?? res;
      const items = data?.requests || data?.items || data?.data || data || [];
      const list = Array.isArray(items) ? items : [];
      
      setRequests(list);

      // Silently sync the currently selected request status if it finished
      if (selectedRequest) {
        const matching = list.find(r => r.id === selectedRequest.id);
        if (matching && matching.status !== selectedRequest.status) {
          handleSelectRequest(matching);
        }
      }

      const completed = list.filter((r: any) => r.status === "completed").length;
      setStats({
        total: data?.total_count || data?.total || list.length,
        completed: completed,
        saved_km: completed * 15.2,
        saved_cost: completed * 45,
      });
    } catch (e) {
      console.warn("Silent fetch failed", e);
    }
  };

  const handleLoadMumbaiDemo = () => {
    setFormName("Mumbai Fast-Track Delivery Tour");
    setFormDesc("Optimal distribution of retail shipments from Mumbai Port Central Depot across Bandra, Andheri, Dadar, and Chembur.");
    setIsDemoLoaded(true);
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormSubmitting(true);
    try {
      const stops = [
        {
          id: "stop-bandra",
          address: "Bandra West, Mumbai",
          latitude: 19.0596,
          longitude: 72.8295,
          demand: 40,
          service_duration: 15,
          state: "MH",
        },
        {
          id: "stop-andheri",
          address: "Andheri East, Mumbai",
          latitude: 19.1136,
          longitude: 72.8697,
          demand: 60,
          service_duration: 20,
          state: "MH",
        },
        {
          id: "stop-dadar",
          address: "Dadar West, Mumbai",
          latitude: 19.0178,
          longitude: 72.8478,
          demand: 30,
          service_duration: 15,
          state: "MH",
        },
        {
          id: "stop-chembur",
          address: "Chembur, Mumbai",
          latitude: 19.0622,
          longitude: 72.8974,
          demand: 50,
          service_duration: 20,
          state: "MH",
        },
      ];

      const vehicles = [
        {
          id: "Tata-Ace-Gold",
          vehicle_class: "PICKUP",
          capacity: 150,
          speed: 40.0,
          cost_per_km: 12.0,
          start_location: { latitude: 18.9438, longitude: 72.8396 },
          end_location: { latitude: 18.9438, longitude: 72.8396 },
          available: true,
        },
        {
          id: "Mahindra-Bolero",
          vehicle_class: "TRUCK",
          capacity: 250,
          speed: 45.0,
          cost_per_km: 18.0,
          start_location: { latitude: 18.9438, longitude: 72.8396 },
          end_location: { latitude: 18.9438, longitude: 72.8396 },
          available: true,
        },
      ];

      const depots = [
        {
          id: "depot-mumbai-port",
          name: "Mumbai Port Central Depot",
          location: {
            latitude: 18.9438,
            longitude: 72.8396,
            address: "Mumbai Port, Fort, Mumbai",
          },
        },
      ];

      const payload = {
        name: formName,
        description: formDesc,
        start_time: new Date().toISOString(),
        end_time: new Date(Date.now() + 12 * 60 * 60 * 1000).toISOString(),
        constraints: {
          max_travel_time: 480,
          max_distance: 300,
        },
        stops,
        vehicles,
        depots,
        priorities: {
          time_priority: 0.5,
          cost_priority: 0.5,
          service_priority: 0.5,
        },
        optimization_goals: {
          minimize_distance: true,
          minimize_vehicles: true,
        },
      };

      // 1. Create request
      const res = await api.createOptimizationRequest(payload);
      const data = res?.data ?? res;
      const requestId = data?.id;

      if (!requestId) {
        throw new Error("Did not receive a valid Request ID from the creation endpoint");
      }

      // 2. Submit VRP solver job
      await api.submitOptimizationJob({ request_id: requestId });

      // 3. Complete and clear
      setShowCreateModal(false);
      setIsDemoLoaded(false);
      await fetchData();
      
      showToast("VRP route optimization successfully submitted and running!", "success");
    } catch (err) {
      console.error("VRP Request failed", err);
      showToast(`VRP creation failed: ${err instanceof Error ? err.message : String(err)}`, "error");
    } finally {
      setFormSubmitting(false);
    }
  };

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
    try {
      // Fetch full request details to always get stops, vehicles and depots
      const fullRequest = await api.getOptimizationRequest(req.id);
      if (fullRequest) {
        req.result = fullRequest.result;
        (req as any).stops = fullRequest.stops;
        (req as any).vehicles = fullRequest.vehicles;
        (req as any).depots = fullRequest.depots;
      }
    } catch (e) {
      console.warn("Could not fetch request details", e);
    }
    setSelectedRequest({ ...req });
  };

  const handleRecalculate = async (vehicleId: string) => {
    if (!selectedRequest?.result?.id) return;
    
    let currentLat = 19.0760; // Mumbai Depot
    let currentLon = 72.8777;
    
    if (selectedRequest.depots && selectedRequest.depots.length > 0) {
      currentLat = selectedRequest.depots[0].location?.latitude || currentLat;
      currentLon = selectedRequest.depots[0].location?.longitude || currentLon;
    }
    
    setRecalculatingVehicles(prev => {
      const next = new Set(prev);
      next.add(vehicleId);
      return next;
    });

    try {
      const res = await api.recalculateRoute({
        result_id: selectedRequest.result.id,
        vehicle_id: vehicleId,
        current_lat: currentLat,
        current_lon: currentLon,
      });
      
      setSelectedRequest(prev => {
        if (!prev) return null;
        const updated = { ...prev, result: res };
        setRequests(list => list.map(r => r.id === prev.id ? updated : r));
        return updated;
      });
      
      showToast(`Successfully recalculated route for vehicle ${vehicleId}!`, "success");
    } catch (e) {
      console.error("Failed to recalculate route:", e);
      showToast(`Failed to recalculate route: ${e instanceof Error ? e.message : String(e)}`, "error");
    } finally {
      setRecalculatingVehicles(prev => {
        const next = new Set(prev);
        next.delete(vehicleId);
        return next;
      });
    }
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

        {/* Right: Map & Schedules Pane */}
        <div className="flex-1 bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col relative">
          {selectedRequest ? (
            <div className="flex-1 flex overflow-hidden">
              {/* Map View Workspace (2/3 width) */}
              <div className="flex-1 relative bg-[#1a1a1a]">
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

                <MapContainer
                  center={[20.5937, 78.9629]}
                  zoom={5}
                  attributionControl={false}
                  style={{ height: "100%", width: "100%" }}
                  className="z-0"
                >
                  <TileLayer
                    attribution='&copy; Google Maps'
                    url="https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}&gl=IN"
                  />
                  <MapBoundsController request={selectedRequest} />

                  {/* Render Depots */}
                  {selectedRequest.depots?.map((depot: any) => (
                    <Marker
                      key={depot.id}
                      position={[depot.location?.latitude || 0, depot.location?.longitude || 0]}
                      icon={createDepotPin()}
                    >
                      <Popup>
                        <div className="text-slate-900 p-1">
                          <h4 className="font-bold text-xs">{depot.name}</h4>
                          <p className="text-[10px] text-slate-500">{depot.location?.address}</p>
                        </div>
                      </Popup>
                    </Marker>
                  ))}

                  {/* Render Markers and Route Paths if result exists */}
                  {(() => {
                    const stopMap = new Map(selectedRequest.stops?.map((s: any) => [s.id, s]) || []);
                    return selectedRequest.result?.routes?.map((route: any, idx: number) => {
                      const color = polylineColors[idx % polylineColors.length];
                      
                      // 1. Gather all stop markers for this route
                      const markers = route.stops?.map((stop: any) => {
                        const stopData = stopMap.get(stop.stop_id);
                        if (!stopData) return null;
                        return (
                          <Marker
                            key={stop.stop_id}
                            position={[stopData.latitude, stopData.longitude]}
                            icon={createSequencedPin(stop.sequence, idx)}
                          >
                            <Popup>
                              <div className="text-slate-900 p-1">
                                <h4 className="font-bold text-xs">Stop #{stop.sequence}</h4>
                                <p className="text-[10px] text-slate-500 mb-1">{stopData.address}</p>
                                <p className="text-[10px] font-mono">
                                  Arrival: {stop.arrival_time ? format(new Date(stop.arrival_time), "HH:mm") : "-"}
                                </p>
                              </div>
                            </Popup>
                          </Marker>
                        );
                      });

                      // 2. Build route path coordinates
                      let pathCoords: [number, number][] = [];
                      if (route.polyline && route.polyline.length > 0) {
                        pathCoords = route.polyline.map((c: any) => [c[0], c[1]]);
                      } else {
                        // Fallback straight lines connecting depot -> stops -> depot
                        if (selectedRequest.depots && selectedRequest.depots.length > 0) {
                          pathCoords.push([selectedRequest.depots[0].location?.latitude || 0, selectedRequest.depots[0].location?.longitude || 0]);
                        }
                        route.stops?.forEach((stop: any) => {
                          const stopData = stopMap.get(stop.stop_id);
                          if (stopData) {
                            pathCoords.push([stopData.latitude, stopData.longitude]);
                          }
                        });
                        if (selectedRequest.depots && selectedRequest.depots.length > 0) {
                          pathCoords.push([selectedRequest.depots[0].location?.latitude || 0, selectedRequest.depots[0].location?.longitude || 0]);
                        }
                      }

                      return (
                        <React.Fragment key={idx}>
                          {markers}
                          {pathCoords.length > 1 && (
                            <>
                              {/* Google Maps Outer Casing */}
                              <Polyline
                                positions={pathCoords}
                                color="#18181b"
                                weight={7}
                                opacity={0.25}
                              />
                              {/* Google Maps Inner Core Path */}
                              <Polyline
                                positions={pathCoords}
                                color={color}
                                weight={4}
                                opacity={0.9}
                              />
                            </>
                          )}
                          {pathCoords.length >= 3 && (
                            <Polygon
                              positions={pathCoords}
                              pathOptions={{
                                fillColor: color,
                                fillOpacity: 0.12,
                                color: color,
                                weight: 1,
                                dashArray: "5, 5",
                                opacity: 0.4
                              }}
                            />
                          )}
                          {pathCoords.length >= 2 && (
                            <Marker
                              position={pathCoords[Math.floor(pathCoords.length / 2)]}
                              icon={createRouteBadgeIcon(route.time || 0, route.distance || 0)}
                            />
                          )}
                        </React.Fragment>
                      );
                    });
                  })()}
                </MapContainer>
              </div>

              {/* Drill-down Schedule Sidebar (1/3 width) */}
              <div className="w-80 border-l border-border-color bg-surface-dark flex flex-col overflow-hidden">
                <div className="p-4 border-b border-border-color bg-surface-darker/30">
                  <h3 className="font-semibold text-xs uppercase tracking-widest text-text-muted">
                    Route Schedules
                  </h3>
                </div>
                <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-4">
                  {(() => {
                    const stopMap = new Map(selectedRequest.stops?.map((s: any) => [s.id, s]) || []);
                    return selectedRequest.result?.routes?.map((route: any, routeIdx: number) => {
                      const color = polylineColors[routeIdx % polylineColors.length];
                      const capacityPct = Math.min(100, Math.round((route.capacity_used / 150) * 100));
                      const isRecalculating = recalculatingVehicles.has(route.vehicle_id);

                      return (
                        <div key={routeIdx} className="bg-surface-darker/40 rounded-xl border border-border-color p-4 hover:border-primary/20 transition-all">
                          {/* Header */}
                          <div className="flex justify-between items-start mb-3">
                            <div>
                              <div className="flex items-center gap-2">
                                <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: color }} />
                                <h4 className="font-bold text-sm text-white">
                                  {route.vehicle_id}
                                </h4>
                              </div>
                              <p className="text-[10px] text-text-muted mt-1 uppercase tracking-wider">
                                {route.stops?.length || 0} stops
                              </p>
                            </div>
                            <button
                              disabled={isRecalculating}
                              onClick={() => handleRecalculate(route.vehicle_id)}
                              className={`text-[9px] px-2 py-1 rounded border transition-all uppercase tracking-wide font-black ${
                                isRecalculating
                                  ? "bg-slate-800 text-slate-500 border-slate-700 cursor-not-allowed"
                                  : "bg-primary/10 hover:bg-primary/25 text-primary border-primary/20"
                              }`}
                            >
                              {isRecalculating ? "Running..." : "Recalculate"}
                            </button>
                          </div>

                          {/* Metrics */}
                          <div className="grid grid-cols-3 gap-2 bg-black/25 p-2 rounded-lg border border-border-color mb-3 text-center">
                            <div>
                              <p className="text-[8px] text-text-muted uppercase font-black tracking-tight">Distance</p>
                              <p className="text-[10px] font-mono font-bold text-slate-200">{route.distance?.toFixed(1)} km</p>
                            </div>
                            <div className="border-l border-border-color">
                              <p className="text-[8px] text-text-muted uppercase font-black tracking-tight">Duration</p>
                              <p className="text-[10px] font-mono font-bold text-slate-200">{Math.round(route.time)} min</p>
                            </div>
                            <div className="border-l border-border-color">
                              <p className="text-[8px] text-text-muted uppercase font-black tracking-tight">Capacity</p>
                              <p className="text-[10px] font-mono font-bold text-slate-200">{capacityPct}%</p>
                            </div>
                          </div>

                          {/* Stops Sequence */}
                          <div className="relative pl-3 border-l border-slate-700/60 space-y-3 my-2 ml-1.5">
                            {/* Start */}
                            <div className="relative flex items-center">
                              <span className="absolute -left-[16.5px] w-1.5 h-1.5 rounded-full bg-slate-500 border border-slate-900" />
                              <p className="text-[9px] font-bold text-slate-400 uppercase tracking-wide">
                                Depot Start
                              </p>
                            </div>

                            {route.stops?.map((stop: any) => {
                              const stopData = stopMap.get(stop.stop_id);
                              return (
                                <div key={stop.stop_id} className="relative">
                                  <span
                                    className="absolute -left-[20.5px] top-0.5 w-3.5 h-3.5 rounded-full flex items-center justify-center text-[8px] font-black text-white"
                                    style={{ backgroundColor: color }}
                                  >
                                    {stop.sequence}
                                  </span>
                                  <div className="text-xs ml-1">
                                    <p className="font-bold text-slate-200 truncate max-w-[170px]" title={stopData?.address}>
                                      {stopData?.address || `Stop ${stop.stop_id}`}
                                    </p>
                                    <div className="text-[9px] text-text-muted flex justify-between mt-0.5">
                                      <span>Arr: {stop.arrival_time ? format(new Date(stop.arrival_time), "HH:mm") : "-"}</span>
                                      <span>Demand: {stopData?.demand || 0}</span>
                                    </div>
                                  </div>
                                </div>
                              );
                            })}

                            {/* End */}
                            <div className="relative flex items-center">
                              <span className="absolute -left-[16.5px] w-1.5 h-1.5 rounded-full bg-slate-500 border border-slate-900" />
                              <p className="text-[9px] font-bold text-slate-400 uppercase tracking-wide">
                                Depot End
                              </p>
                            </div>
                          </div>
                        </div>
                      );
                    });
                  })()}
                </div>
              </div>
            </div>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-text-muted">
              <MapIcon className="w-16 h-16 opacity-20 mb-4" />
              <p>{uiStrings?.msg_select_request}</p>
            </div>
          )}
        </div>
      </div>

      {/* Beautiful high-fidelity Modal for VRP creation */}
      {showCreateModal && (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/65 backdrop-blur-sm p-4 animate-fade-in">
          <div className="bg-[#121824] border border-[#232d42] rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl flex flex-col animate-slide-up text-white">
            <div className="p-6 border-b border-[#232d42] flex justify-between items-center bg-[#0d121c]/30">
              <h3 className="text-lg font-bold text-white">Create VRP Optimization</h3>
              <button 
                onClick={() => setShowCreateModal(false)}
                className="text-[#64748b] hover:text-white transition-colors"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <form onSubmit={handleCreateSubmit} className="p-6 space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-[#64748b] uppercase tracking-wider">Request Name</label>
                <input 
                  type="text" 
                  value={formName} 
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full bg-[#0d121c] border border-[#232d42] rounded-xl px-4 py-2.5 text-white focus:outline-none focus:border-[#3b82f6]/50 text-sm"
                  required
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-[#64748b] uppercase tracking-wider">Description</label>
                <textarea 
                  value={formDesc} 
                  onChange={(e) => setFormDesc(e.target.value)}
                  className="w-full bg-[#0d121c] border border-[#232d42] rounded-xl px-4 py-2 text-white focus:outline-none focus:border-[#3b82f6]/50 text-sm h-20 resize-none"
                />
              </div>

              <div className="bg-[#3b82f6]/5 border border-[#3b82f6]/20 rounded-xl p-4 flex flex-col space-y-3">
                <div className="flex justify-between items-center">
                  <span className="text-xs font-bold text-[#3b82f6] uppercase tracking-wide">Mumbai Regional Dataset</span>
                  <span className="text-[10px] bg-[#3b82f6]/20 text-[#3b82f6] px-2 py-0.5 rounded font-black uppercase">India Region</span>
                </div>
                <p className="text-xs text-[#94a3b8] leading-relaxed">
                  Automatically populates the request with 1 Depot (Mumbai Port), 2 Vehicles (Tata Ace, Mahindra Bolero), and 4 sequenced stops in Bandra, Andheri, Dadar, and Chembur.
                </p>
                <button
                  type="button"
                  onClick={handleLoadMumbaiDemo}
                  className="bg-[#3b82f6] hover:bg-[#2563eb] text-white font-semibold text-xs py-2 px-4 rounded-lg self-start transition-all shadow-md shadow-[#3b82f6]/10 font-bold"
                >
                  Apply Mumbai Demo Setup
                </button>
              </div>

              {isDemoLoaded && (
                <div className="grid grid-cols-3 gap-2 bg-[#0d121c] p-3 rounded-xl border border-[#232d42] text-center animate-fade-in">
                  <div>
                    <span className="text-[9px] text-[#64748b] uppercase font-bold">Stops</span>
                    <p className="text-sm font-black text-white">4 Stops</p>
                  </div>
                  <div className="border-l border-[#232d42]">
                    <span className="text-[9px] text-[#64748b] uppercase font-bold">Vehicles</span>
                    <p className="text-sm font-black text-white">2 Vehicles</p>
                  </div>
                  <div className="border-l border-[#232d42]">
                    <span className="text-[9px] text-[#64748b] uppercase font-bold">Depots</span>
                    <p className="text-sm font-black text-white">1 Depot</p>
                  </div>
                </div>
              )}

              <div className="pt-4 border-t border-[#232d42] flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 border border-[#232d42] hover:bg-[#0d121c] rounded-lg text-xs font-bold transition-all"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={formSubmitting}
                  className="px-5 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-lg text-xs font-black shadow-lg shadow-[#3b82f6]/15 transition-all flex items-center gap-2"
                >
                  {formSubmitting ? (
                    <>
                      <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                      <span>Optimizing...</span>
                    </>
                  ) : (
                    <span>Create & Optimize</span>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
      {/* Premium Auto-Fading Toast Notification */}
      {toastMsg && (
        <div className={`fixed top-4 right-4 z-[10000] flex items-center gap-3 px-5 py-3 rounded-xl border shadow-xl animate-fade-in text-xs font-bold text-white ${
          toastMsg.type === "success" 
            ? "bg-emerald-600/90 border-emerald-500/30 backdrop-blur-md" 
            : "bg-rose-600/90 border-rose-500/30 backdrop-blur-md"
        }`}>
          {toastMsg.type === "success" ? (
            <svg className="w-5 h-5 text-emerald-100" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          ) : (
            <svg className="w-5 h-5 text-rose-100" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          )}
          <span>{toastMsg.text}</span>
        </div>
      )}
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
