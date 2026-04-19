import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { TripFilters, TripFilterState } from "../../../components/trips/TripFilters";
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
  distance: number;
  status: string;
  customer_name?: string;
}

const statusColors: Record<string, string> = {
  scheduled: "border-primary/40 bg-primary/10",
  in_progress: "border-amber-500/40 bg-amber-500/10",
  completed: "border-emerald-500/40 bg-emerald-500/10",
  cancelled: "border-rose-500/40 bg-rose-500/10",
};

const statusDot: Record<string, string> = {
  scheduled: "bg-primary",
  in_progress: "bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.5)]",
  completed: "bg-emerald-500",
  cancelled: "bg-rose-500",
};

const COLUMNS = [
  { key: "scheduled", label: "Scheduled", icon: "calendar_today" },
  { key: "in_progress", label: "In Progress", icon: "moving" },
  { key: "completed", label: "Completed", icon: "check_circle" },
  { key: "cancelled", label: "Cancelled", icon: "cancel" },
];

const defaultFilters: TripFilterState = {
  status: "",
  tripType: "",
  customerName: "",
  vehicleId: "",
  driverId: "",
  startDate: "",
  endDate: "",
  orderBy: "",
  orderDir: "DESC",
};

export default function TripsList() {
  const { uiStrings: rawUiStrings } = useUIStrings("tracking");
  const uiStrings = rawUiStrings || {};
  const [trips, setTrips] = useState<Trip[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState<TripFilterState>(defaultFilters);
  const [keyword, setKeyword] = useState("");
  const [vehicleMap, setVehicleMap] = useState<Record<string, string>>({});

  useEffect(() => {
    loadTrips();
    loadVehicles();
  }, [page, filters]);

  const loadVehicles = async () => {
    try {
      const res = await api.getVehicles({ limit: 1000 });
      const vehicles = res.vehicles || res.data || [];
      const map: Record<string, string> = {};
      vehicles.forEach((v: any) => {
        map[v.id] = v.registration_number || v.plate_number || v.name || v.id.slice(0, 8);
      });
      setVehicleMap(map);
    } catch (err) {
      console.warn("Failed to load vehicle registry for mapping:", err);
    }
  };

  const loadTrips = async () => {
    try {
      setLoading(true);
      const params: Record<string, string | number | undefined> = {
        page,
        limit: 200,
      };

      if (filters.status) params.status = filters.status;
      if (filters.tripType) params.trip_type = filters.tripType;
      if (filters.customerName) params.customer_name = filters.customerName;
      if (filters.vehicleId) params.vehicle_id = filters.vehicleId;
      if (filters.driverId) params.driver_id = filters.driverId;
      if (filters.startDate) params.from = filters.startDate;
      if (filters.endDate) params.to = filters.endDate;
      if (filters.orderBy) {
        params.order_by = filters.orderBy;
        params.order_dir = filters.orderDir;
      }

      const data = await api.getTrips(params);
      setTrips(data.trips || []);
      setTotal(data.total || 0);
    } catch (error) {
      console.error("Failed to load trips:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleFilterChange = (newFilters: TripFilterState) => {
    setFilters(newFilters);
    setPage(1);
  };

  const handleResetFilters = () => {
    setFilters(defaultFilters);
    setKeyword("");
    setPage(1);
  };

  // Client-side keyword search
  const filtered = useMemo(() => {
    if (!keyword.trim()) return trips;
    const q = keyword.toLowerCase();
    return trips.filter((t) => {
      const vehicleName = vehicleMap[t.vehicle_id]?.toLowerCase() || "";
      return (
        t.trip_id.toLowerCase().includes(q) ||
        t.vehicle_id.toLowerCase().includes(q) ||
        vehicleName.includes(q) ||
        (t.driver_id || "").toLowerCase().includes(q) ||
        (t.customer_name || "").toLowerCase().includes(q) ||
        t.start_location.toLowerCase().includes(q) ||
        t.end_location.toLowerCase().includes(q)
      );
    });
  }, [trips, keyword, vehicleMap]);

  const boardTripsByStatus = COLUMNS.map((col) => ({
    ...col,
    trips: filtered.filter((t) => t.status === col.key),
  }));

  return (
    <div className="p-6 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest text-on-surface-variant mb-1">
            <span>Fleet Operations</span>
            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
            <span className="text-white">Active Trips</span>
          </div>
          <h1 className="text-3xl font-bold text-white tracking-tight">
            Fleet Trips
          </h1>
        </div>
        <Link
          to="/admin/trips/add"
          className="px-5 py-2.5 bg-primary hover:bg-primary-dark text-white text-[11px] font-bold uppercase tracking-wider rounded-xl flex items-center gap-2 transition-all shadow-lg shadow-primary/20 active:scale-95"
        >
          <span className="material-symbols-outlined text-[18px]">add</span>
          Add New Trip
        </Link>
      </div>

      {/* Filters */}
      <TripFilters
        filters={filters}
        onFilterChange={handleFilterChange}
        onReset={handleResetFilters}
      />

      {/* Board View (Kanban) */}
      <BoardView
        columns={boardTripsByStatus}
        loading={loading}
        keyword={keyword}
        onKeywordChange={setKeyword}
        formatDate={formatDate}
        totalCount={filtered.length}
        vehicleMap={vehicleMap}
      />

      {/* List View (Table) */}
      <ListView
        trips={filtered}
        loading={loading}
        total={total}
        page={page}
        setPage={setPage}
        uiStrings={uiStrings}
        formatDate={formatDate}
        onRefresh={loadTrips}
        keyword={keyword}
        onKeywordChange={setKeyword}
        vehicleMap={vehicleMap}
      />
    </div>
  );
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

/* ── Board View (Kanban) ── */
function BoardView({
  columns,
  loading,
  keyword,
  onKeywordChange,
  formatDate,
  totalCount,
  vehicleMap,
}: {
  columns: { key: string; label: string; icon: string; trips: Trip[] }[];
  loading: boolean;
  keyword: string;
  onKeywordChange: (v: string) => void;
  formatDate: (d: string) => string;
  totalCount: number;
  vehicleMap: Record<string, string>;
}) {
  return (
    <div>
      <div className="flex items-center justify-between mb-4 mt-4 px-2">
        <h2 className="text-white text-xs font-bold uppercase tracking-widest">Trips Board</h2>
        <span className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider bg-surface-container px-2 py-0.5 rounded border border-outline">
          {totalCount} Active Trips
        </span>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-20 text-on-surface-variant">
          <span className="material-symbols-outlined animate-spin mr-3">refresh</span>
          <span className="text-[11px] font-bold uppercase tracking-widest">Initialising Board...</span>
        </div>
      ) : (
        <div className="flex gap-4 overflow-x-auto pb-2">
          {columns.map((col) => (
            <div
              key={col.key}
              className="min-w-[280px] w-[280px] flex-shrink-0 flex flex-col"
            >
              <div className="flex items-center gap-2 mb-4 px-2">
                <span className="material-symbols-outlined text-on-surface-variant text-[18px]">
                  {col.icon}
                </span>
                <span className="text-[11px] font-black text-white uppercase tracking-wider">{col.label}</span>
                <span className="bg-surface-container border border-outline text-on-surface-variant text-[10px] px-2 py-0.5 rounded font-black ml-auto">
                  {col.trips.length}
                </span>
              </div>

              <div className="space-y-2">
                {col.trips.length === 0 ? (
                  <div className="text-center py-12 text-on-surface-variant text-[10px] font-bold uppercase tracking-widest border border-dashed border-outline rounded-2xl bg-surface-container/10">
                    No Trips
                  </div>
                ) : (
                  col.trips.map((trip) => (
                    <Link
                      key={trip.id}
                      to={`/admin/trips/${trip.id}`}
                      className={`block bg-surface-container border border-outline rounded-2xl p-4 hover:border-primary transition-all shadow-xl shadow-black/5 group ${
                        statusColors[trip.status] || ""
                      }`}
                    >
                      <div className="flex items-center justify-between mb-3">
                        <span className="font-mono text-white text-[11px] font-bold bg-black/20 px-2 py-0.5 rounded">
                          #{trip.trip_id}
                        </span>
                        <div className="flex items-center gap-2">
                          <span className={`w-2 h-2 rounded-full ${statusDot[trip.status] || ""}`} />
                        </div>
                      </div>
                      <div className="text-[11px] font-bold text-white mb-3 line-clamp-2 leading-relaxed">
                        {trip.start_location} 
                        <span className="material-symbols-outlined text-[10px] mx-1 text-on-surface-variant">arrow_forward</span> 
                        {trip.end_location}
                      </div>
                      <div className="flex items-center justify-between text-[10px] font-black uppercase tracking-widest text-on-surface-variant border-t border-outline/30 pt-3 mt-3 group-hover:border-primary/20 transition-colors">
                        <span className="flex items-center gap-1.5" title={trip.vehicle_id}>
                          <span className="material-symbols-outlined text-[14px]">local_shipping</span>
                          {vehicleMap[trip.vehicle_id] || trip.vehicle_id?.slice(0, 8) + "…"}
                        </span>
                        <span className="font-mono">{trip.distance?.toFixed(1) || "—"} KM</span>
                      </div>
                      <div className="flex items-center justify-between text-[9px] font-bold text-on-surface-variant/70 mt-3 uppercase tracking-wider">
                        <span className="truncate max-w-[120px]">{trip.customer_name || "Internal Fleet"}</span>
                        <span>{formatDate(trip.start_time)}</span>
                      </div>
                    </Link>
                  ))
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/* ── List View (Table) ── */
function ListView({
  trips,
  loading,
  total,
  page,
  setPage,
  uiStrings,
  formatDate,
  onRefresh,
  keyword,
  onKeywordChange,
  vehicleMap,
}: {
  trips: Trip[];
  loading: boolean;
  total: number;
  page: number;
  setPage: (p: number | ((p: number) => number)) => void;
  uiStrings: Record<string, string>;
  formatDate: (d: string) => string;
  onRefresh: () => void;
  keyword: string;
  onKeywordChange: (v: string) => void;
  vehicleMap: Record<string, string>;
}) {
  return (
    <div className="bg-surface-container border border-outline rounded-3xl overflow-hidden shadow-2xl mt-8">
      {/* Table Header with Search */}
      <div className="p-5 border-b border-outline flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-surface-container/50">
        <div className="flex items-center gap-4">
          <h3 className="text-white font-bold text-sm uppercase tracking-wider">
            Detailed Trip Log
          </h3>
          <span className="bg-emerald-500/10 text-emerald-400 text-[10px] px-3 py-1 rounded-full font-black uppercase tracking-widest border border-emerald-500/20">
            {trips.filter((t) => t.status === "in_progress").length} Live Tracking
          </span>
        </div>
        <div className="flex items-center gap-3 w-full sm:w-auto">
          {/* Enhanced search integration */}
          <div className="relative flex-1 sm:flex-none sm:min-w-[280px]">
            <span className="material-symbols-outlined absolute left-3 top-2.5 text-on-surface-variant text-sm">
              search
            </span>
            <input
              type="text"
              value={keyword}
              onChange={(e) => onKeywordChange(e.target.value)}
              placeholder="Filter by Trip ID, Vehicle, or Route..."
              className="w-full bg-surface-dim border border-outline text-white text-xs rounded-xl pl-10 pr-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-on-surface-variant/50 shadow-inner font-bold tracking-tight"
            />
          </div>
          <button onClick={onRefresh} className="p-2.5 hover:bg-surface-container rounded-xl text-on-surface-variant hover:text-white border border-outline transition-all active:scale-95" title="Refresh Log">
            <span className="material-symbols-outlined text-[20px]">refresh</span>
          </button>
          <button className="p-2.5 hover:bg-surface-container rounded-xl text-on-surface-variant hover:text-white border border-outline transition-all active:scale-95" title="Export Trips">
            <span className="material-symbols-outlined text-[20px]">download</span>
          </button>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left">
          <thead className="border-b border-outline text-[10px] text-on-surface-variant font-bold uppercase tracking-wider bg-surface-container/80 backdrop-blur-sm">
            <tr>
              <th className="px-6 py-4">Trip ID</th>
              <th className="px-6 py-4">Vehicle</th>
              <th className="px-6 py-4">Route Path</th>
              <th className="px-6 py-4">Planned / Actual</th>
              <th className="px-6 py-4">Distance</th>
              <th className="px-6 py-4">Customer</th>
              <th className="px-6 py-4 text-right">Status</th>
            </tr>
          </thead>
          <tbody className="text-sm">
            {loading ? (
              <tr>
                <td colSpan={7} className="px-6 py-12 text-center text-on-surface-variant">
                   <span className="material-symbols-outlined animate-spin text-2xl">refresh</span>
                </td>
              </tr>
            ) : trips.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-text-muted">No trips found</td>
              </tr>
            ) : (
              trips.map((trip) => (
                <tr
                  key={trip.id}
                  className="border-b border-outline/50 hover:bg-surface-container/40 transition-all group"
                >
                  <td className="px-6 py-4">
                    <Link to={`/admin/trips/${trip.id}`} className="font-mono text-white font-bold bg-white/5 px-2 py-1 rounded hover:bg-primary/20 hover:text-primary transition-all">
                      #{trip.trip_id}
                    </Link>
                  </td>
                  <td className="px-6 py-4 text-white font-bold tracking-tight">
                    <div className="flex items-center gap-3" title={trip.vehicle_id}>
                      <span className="material-symbols-outlined text-on-surface-variant text-lg">local_shipping</span>
                      {vehicleMap[trip.vehicle_id] || trip.vehicle_id?.slice(0, 8) + "…"}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-on-surface-variant font-bold text-xs max-w-[240px] truncate">
                    {trip.start_location} <span className="text-white/20 mx-1">→</span> {trip.end_location}
                  </td>
                  <td className="px-6 py-4 text-on-surface-variant font-bold text-xs whitespace-nowrap">
                    {formatDate(trip.start_time)}
                  </td>
                  <td className="px-6 py-4 text-white font-mono font-bold text-xs">
                    {trip.distance?.toFixed(1) || "---"} <span className="text-[9px] text-on-surface-variant uppercase ml-1">km</span>
                  </td>
                  <td className="px-6 py-4 text-on-surface-variant font-black uppercase text-[10px] tracking-widest max-w-[160px] truncate">
                    {trip.customer_name || "Internal Fleet"}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <span
                      className={`inline-flex items-center gap-2 px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border border-outline shadow-sm ${
                        statusColors[trip.status] || statusColors.scheduled
                      }`}
                    >
                      <span className={`w-1.5 h-1.5 rounded-full ${statusDot[trip.status] || statusDot.scheduled}`} />
                      {trip.status?.toUpperCase().replace("_", " ")}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Footer Area */}
      <div className="p-4 border-t border-outline bg-surface-container/50 flex items-center justify-between">
        <span className="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest bg-surface-dim px-3 py-1 rounded-full border border-outline">
          Trips {(page - 1) * 20 + 1} – {Math.min(page * 20, total)} of {total}
        </span>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setPage((p: number) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="p-2 rounded-xl text-on-surface-variant hover:text-white hover:bg-surface-container border border-outline disabled:opacity-20 transition-all active:scale-95"
          >
            <span className="material-symbols-outlined text-[20px]">chevron_left</span>
          </button>
          <span className="px-3 py-1.5 bg-primary text-white text-[11px] font-black rounded-xl shadow-lg shadow-primary/20 tracking-widest">{page}</span>
          <button
            onClick={() => setPage((p: number) => p + 1)}
            disabled={page * 20 >= total}
            className="p-2 rounded-xl text-on-surface-variant hover:text-white hover:bg-surface-container border border-outline disabled:opacity-20 transition-all active:scale-95"
          >
            <span className="material-symbols-outlined text-[20px]">chevron_right</span>
          </button>
        </div>
      </div>
    </div>
  );
}
