import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { TripFilters, TripFilterState } from "../../../components/trips/TripFilters";
import { useUIStrings } from "../../../hooks/useUIStrings";
import api from "../../../services/api";

interface Location {
  latitude: number;
  longitude: number;
  address: string;
}

interface Trip {
  id: string;
  trip_id: string;
  vehicle_id: string;
  driver_id?: string;
  start_location: string | Location;
  end_location: string | Location;
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
      setTrips(data.data || data.trips || []);
      setTotal(data.total_count || data.total || 0);
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
      const startLoc = formatLocation(t.start_location).toLowerCase();
      const endLoc = formatLocation(t.end_location).toLowerCase();
      return (
        t.trip_id.toLowerCase().includes(q) ||
        t.vehicle_id.toLowerCase().includes(q) ||
        vehicleName.includes(q) ||
        (t.driver_id || "").toLowerCase().includes(q) ||
        (t.customer_name || "").toLowerCase().includes(q) ||
        startLoc.includes(q) ||
        endLoc.includes(q)
      );
    });
  }, [trips, keyword, vehicleMap]);



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

      {/* Status KPI Board (Case Numbers) */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4 mt-2">
        {COLUMNS.map((col) => {
          const count = trips.filter((t) => t.status === col.key).length;
          const isSelected = filters.status === col.key;
          
          let activeStyle = "";
          if (col.key === "in_progress") {
            activeStyle = isSelected 
              ? "border-amber-500 bg-amber-500/20 text-amber-400 ring-2 ring-amber-500/20 shadow-amber-500/5" 
              : "border-amber-500/20 bg-amber-500/5 hover:bg-amber-500/10 text-amber-400/80 hover:text-amber-400";
          } else if (col.key === "completed") {
            activeStyle = isSelected 
              ? "border-emerald-500 bg-emerald-500/20 text-emerald-400 ring-2 ring-emerald-500/20 shadow-emerald-500/5" 
              : "border-emerald-500/20 bg-emerald-500/5 hover:bg-emerald-500/10 text-emerald-400/80 hover:text-emerald-400";
          } else if (col.key === "cancelled") {
            activeStyle = isSelected 
              ? "border-rose-500 bg-rose-500/20 text-rose-400 ring-2 ring-rose-500/20 shadow-rose-500/5" 
              : "border-rose-500/20 bg-rose-500/5 hover:bg-rose-500/10 text-rose-400/80 hover:text-rose-400";
          } else {
            activeStyle = isSelected 
              ? "border-primary bg-primary/20 text-primary ring-2 ring-primary/20 shadow-primary/5" 
              : "border-primary/20 bg-primary/5 hover:bg-primary/10 text-primary/80 hover:text-primary";
          }
          
          return (
            <button 
              key={col.key}
              onClick={() => handleFilterChange({ 
                ...filters, 
                status: filters.status === col.key ? "" : col.key 
              })}
              className={`p-5 rounded-2xl border transition-all duration-300 flex items-center justify-between shadow-xl text-left ${activeStyle} active:scale-95`}
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-white/5 flex items-center justify-center">
                  <span className="material-symbols-outlined text-lg">{col.icon}</span>
                </div>
                <div>
                  <h4 className="text-white text-[10px] font-black uppercase tracking-widest opacity-80">{col.label}</h4>
                  <p className="text-[9px] font-bold text-on-surface-variant/80 mt-0.5">
                    {isSelected ? "Click to clear filter" : "Filter detailed log"}
                  </p>
                </div>
              </div>
              <div className="text-3xl font-black tracking-tight text-white font-mono">
                {count}
              </div>
            </button>
          );
        })}
      </div>

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

function getPredefinedAddress(lat: number, lng: number): string {
  const roundedLat = parseFloat(lat.toFixed(3));
  const roundedLng = parseFloat(lng.toFixed(3));

  // Delhi NCR bounds
  if (roundedLat >= 28.5 && roundedLat <= 28.9 && roundedLng >= 77.0 && roundedLng <= 77.3) {
    if (roundedLat === 28.636 && roundedLng === 77.220) return "Connaught Place, New Delhi";
    if (roundedLat === 28.614 && roundedLng === 77.209) return "New Delhi Hub";
    return "Delhi Terminal Depot";
  }
  // Gorakhpur
  if (roundedLat >= 26.6 && roundedLat <= 26.9 && roundedLng >= 83.2 && roundedLng <= 83.5) {
    return "Gorakhpur Logistic Park";
  }
  // Siwan
  if (roundedLat >= 26.0 && roundedLat <= 26.3 && roundedLng >= 84.2 && roundedLng <= 84.5) {
    return "Siwan Distribution Center";
  }
  // Jaipur
  if (roundedLat >= 26.8 && roundedLat <= 27.1 && roundedLng >= 75.6 && roundedLng <= 75.9) {
    return "Jaipur Transport Nagar";
  }

  return `${lat.toFixed(4)}, ${lng.toFixed(4)}`;
}

function formatLocation(loc: string | Location) {
  if (!loc) return "Unknown Location";
  if (typeof loc === "string") return loc;
  if (loc.address && loc.address.trim() !== "") return loc.address;
  return getPredefinedAddress(loc.latitude, loc.longitude);
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
  const navigate = useNavigate();
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
                  onClick={() => navigate(`/admin/trips/${trip.id}`)}
                  className="border-b border-outline/50 hover:bg-white/[0.06] transition-all group cursor-pointer"
                >
                  <td className="px-6 py-4">
                    <Link to={`/admin/trips/${trip.id}`} className="font-mono text-white font-bold bg-white/5 px-2 py-1 rounded group-hover:bg-primary/20 group-hover:text-primary transition-all">
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
                    {formatLocation(trip.start_location)} <span className="text-white/20 mx-1">→</span> {formatLocation(trip.end_location)}
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
