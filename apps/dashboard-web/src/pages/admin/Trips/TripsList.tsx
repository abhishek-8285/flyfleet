import {
  ArrowDownTrayIcon,
  ArrowPathIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  PlusIcon,
  TruckIcon,
} from "@heroicons/react/24/outline";
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
  scheduled: "border-primary/40 bg-primary/5",
  in_progress: "border-warning/40 bg-warning/5",
  completed: "border-success/40 bg-success/5",
  cancelled: "border-danger/40 bg-danger/5",
};

const statusDot: Record<string, string> = {
  scheduled: "bg-primary",
  in_progress: "bg-warning animate-pulse",
  completed: "bg-success",
  cancelled: "bg-danger",
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

  useEffect(() => {
    loadTrips();
  }, [page, filters]);

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
    return trips.filter(
      (t) =>
        t.trip_id.toLowerCase().includes(q) ||
        t.vehicle_id.toLowerCase().includes(q) ||
        (t.driver_id || "").toLowerCase().includes(q) ||
        (t.customer_name || "").toLowerCase().includes(q) ||
        t.start_location.toLowerCase().includes(q) ||
        t.end_location.toLowerCase().includes(q),
    );
  }, [trips, keyword]);

  const boardTripsByStatus = COLUMNS.map((col) => ({
    ...col,
    trips: filtered.filter((t) => t.status === col.key),
  }));

  return (
    <div className="p-6 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <span>{uiStrings.fleet_management_label || "Fleet Management"}</span>
            <ChevronRightIcon className="w-3 h-3" />
            <span className="text-white">{uiStrings.trips_title || "Trips"}</span>
          </div>
          <h1 className="text-2xl font-bold text-white">
            {uiStrings.trip_list_title || "Trip Management"}
          </h1>
        </div>
        <Link
          to="/admin/trips/add"
          className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
        >
          <PlusIcon className="w-4 h-4" />
          {uiStrings.btn_add_trip || "Add Trip"}
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
}: {
  columns: { key: string; label: string; icon: string; trips: Trip[] }[];
  loading: boolean;
  keyword: string;
  onKeywordChange: (v: string) => void;
  formatDate: (d: string) => string;
  totalCount: number;
}) {
  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-white font-semibold">Board View</h2>
        <span className="text-xs text-text-muted">{totalCount} trips</span>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-20 text-text-muted">
          <ArrowPathIcon className="w-6 h-6 animate-spin mr-2" />
          Loading…
        </div>
      ) : (
        <div className="flex gap-4 overflow-x-auto pb-2">
          {columns.map((col) => (
            <div
              key={col.key}
              className="min-w-[260px] w-[260px] flex-shrink-0 flex flex-col"
            >
              <div className="flex items-center gap-2 mb-3 px-1">
                <span className="material-symbols-outlined text-text-muted text-[18px]">
                  {col.icon}
                </span>
                <span className="text-sm font-semibold text-white">{col.label}</span>
                <span className="bg-surface-dark border border-border-color text-text-muted text-xs px-2 py-0.5 rounded-full font-bold">
                  {col.trips.length}
                </span>
              </div>

              <div className="space-y-2">
                {col.trips.length === 0 ? (
                  <div className="text-center py-8 text-text-muted text-xs border border-dashed border-border-color rounded-lg">
                    No trips
                  </div>
                ) : (
                  col.trips.map((trip) => (
                    <Link
                      key={trip.id}
                      to={`/admin/trips/${trip.id}`}
                      className={`block bg-surface-dark border rounded-xl p-3 hover:border-primary/40 transition-colors ${
                        statusColors[trip.status] || ""
                      }`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-mono text-primary text-sm font-semibold">
                          {trip.trip_id}
                        </span>
                        <span className={`w-2 h-2 rounded-full ${statusDot[trip.status] || ""}`} />
                      </div>
                      <div className="text-xs text-text-muted mb-2 truncate">
                        {trip.start_location} → {trip.end_location}
                      </div>
                      <div className="flex items-center justify-between text-[11px] text-text-muted">
                        <span className="flex items-center gap-1">
                          <TruckIcon className="w-3 h-3" />
                          {trip.vehicle_id?.slice(0, 8)}…
                        </span>
                        <span className="font-mono">{trip.distance?.toFixed(1) || "—"} km</span>
                      </div>
                      {trip.customer_name && (
                        <div className="text-[11px] text-text-muted mt-1 truncate">
                          {trip.customer_name}
                        </div>
                      )}
                      <div className="text-[11px] text-text-muted mt-1">
                        {formatDate(trip.start_time)}
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
}) {
  return (
    <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
      {/* Table Header with Search */}
      <div className="p-4 border-b border-border-color flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 bg-surface-darker">
        <div className="flex items-center gap-3">
          <h3 className="text-white font-semibold">
            {uiStrings.title_active_trips || "Dispatch Operations Log"}
          </h3>
          <span className="bg-primary/20 text-primary text-xs px-2 py-0.5 rounded-full font-bold">
            {trips.filter((t) => t.status === "in_progress").length} Live
          </span>
        </div>
        <div className="flex items-center gap-2 w-full sm:w-auto">
          {/* Simple keyword search */}
          <div className="relative flex-1 sm:flex-none sm:min-w-[240px]">
            <span className="material-symbols-outlined absolute left-2.5 top-2 text-text-muted text-sm">
              search
            </span>
            <input
              type="text"
              value={keyword}
              onChange={(e) => onKeywordChange(e.target.value)}
              placeholder="Search trips…"
              className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-8 pr-3 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary placeholder-text-muted"
            />
          </div>
          <button onClick={onRefresh} className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white">
            <ArrowPathIcon className="w-4 h-4" />
          </button>
          <button className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white">
            <ArrowDownTrayIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left">
          <thead className="border-b border-border-color text-xs text-text-muted uppercase bg-surface-darker">
            <tr>
              <th className="px-4 py-3 font-semibold">Trip ID</th>
              <th className="px-4 py-3 font-semibold">Vehicle</th>
              <th className="px-4 py-3 font-semibold">Route</th>
              <th className="px-4 py-3 font-semibold">Start Time</th>
              <th className="px-4 py-3 font-semibold">Distance</th>
              <th className="px-4 py-3 font-semibold">Customer</th>
              <th className="px-4 py-3 font-semibold text-right">Status</th>
            </tr>
          </thead>
          <tbody className="text-sm">
            {loading ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-text-muted">Loading…</td>
              </tr>
            ) : trips.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-text-muted">No trips found</td>
              </tr>
            ) : (
              trips.map((trip) => (
                <tr
                  key={trip.id}
                  className="border-b border-border-color hover:bg-surface-darker/50 transition-colors"
                >
                  <td className="px-4 py-3">
                    <Link to={`/admin/trips/${trip.id}`} className="font-mono text-primary hover:underline">
                      {trip.trip_id}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-white flex items-center gap-2">
                    <TruckIcon className="w-4 h-4 text-text-muted" />
                    {trip.vehicle_id?.slice(0, 8)}…
                  </td>
                  <td className="px-4 py-3 text-text-muted max-w-[200px] truncate">
                    {trip.start_location} → {trip.end_location}
                  </td>
                  <td className="px-4 py-3 text-text-muted whitespace-nowrap">
                    {formatDate(trip.start_time)}
                  </td>
                  <td className="px-4 py-3 text-text-muted font-mono">
                    {trip.distance?.toFixed(1) || "---"} km
                  </td>
                  <td className="px-4 py-3 text-text-muted max-w-[140px] truncate">
                    {trip.customer_name || "—"}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <span
                      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-bold border ${
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

      {/* Pagination */}
      <div className="p-3 border-t border-border-color bg-surface-darker flex items-center justify-between">
        <span className="text-xs text-text-muted">
          Showing {(page - 1) * 20 + 1} – {Math.min(page * 20, total)} of {total}
        </span>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setPage((p: number) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="p-1 rounded text-text-muted hover:text-white hover:bg-border-color disabled:opacity-50"
          >
            <ChevronLeftIcon className="w-4 h-4" />
          </button>
          <span className="px-2 py-1 bg-primary text-white text-xs font-bold rounded">{page}</span>
          <button
            onClick={() => setPage((p: number) => p + 1)}
            disabled={page * 20 >= total}
            className="p-1 rounded text-text-muted hover:text-white hover:bg-border-color disabled:opacity-50"
          >
            <ChevronRightIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
