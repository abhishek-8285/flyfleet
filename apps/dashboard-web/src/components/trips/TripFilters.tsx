import React, { useEffect, useRef, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";

// ── Date Helpers ──
const fmt = (d: Date): string => d.toISOString().split("T")[0];
const today = () => new Date();
const daysAgo = (n: number) => { const d = new Date(); d.setDate(d.getDate() - n); return d; };

type DatePreset = "today" | "yesterday" | "last7" | "last30" | "custom";

const PRESETS: { key: DatePreset; label: string; icon: string }[] = [
  { key: "today", label: "Today", icon: "today" },
  { key: "yesterday", label: "Yesterday", icon: "calendar_today" },
  { key: "last7", label: "Last 7 Days", icon: "date_range" },
  { key: "last30", label: "Last 1 Month", icon: "calendar_month" },
  { key: "custom", label: "Custom", icon: "tune" },
];

export interface TripFilterState {
  status: string;
  tripType: string;
  customerName: string;
  vehicleId: string;
  driverId: string;
  startDate: string;
  endDate: string;
  orderBy: string;
  orderDir: "ASC" | "DESC";
}

interface TripFiltersProps {
  filters: TripFilterState;
  onFilterChange: (filters: TripFilterState) => void;
  onReset: () => void;
}

interface VehicleOption {
  id: string;
  registration_number?: string;
  plate_number?: string;
  name?: string;
}

interface DriverOption {
  id: string;
  name: string;
}

export const TripFilters: React.FC<TripFiltersProps> = ({
  filters,
  onFilterChange,
  onReset,
}) => {
  const { uiStrings: rawUiStrings } = useUIStrings("trips");
  const uiStrings = rawUiStrings || {};

  const [vehicles, setVehicles] = useState<VehicleOption[]>([]);
  const [drivers, setDrivers] = useState<DriverOption[]>([]);
  const [customers, setCustomers] = useState<string[]>([]);

  // Date picker popup
  const [showDatePopup, setShowDatePopup] = useState(false);
  const [datePreset, setDatePreset] = useState<DatePreset | null>(null);
  const [tempStart, setTempStart] = useState(filters.startDate);
  const [tempEnd, setTempEnd] = useState(filters.endDate);
  const [dateError, setDateError] = useState("");
  const datePopupRef = useRef<HTMLDivElement>(null);

  // Advanced side drawer
  const [showDrawer, setShowDrawer] = useState(false);

  const drawerAdvancedCount =
    (filters.tripType ? 1 : 0) + (filters.customerName ? 1 : 0);

  const totalActiveCount =
    (filters.status ? 1 : 0) +
    (filters.startDate ? 1 : 0) +
    (filters.endDate ? 1 : 0) +
    drawerAdvancedCount +
    (filters.driverId || filters.vehicleId ? 1 : 0);

  useEffect(() => {
    loadFilterOptions();
  }, []);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        datePopupRef.current &&
        !datePopupRef.current.contains(e.target as Node)
      ) {
        setShowDatePopup(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    if (showDatePopup) {
      setTempStart(filters.startDate);
      setTempEnd(filters.endDate);
      setDateError("");
      setDatePreset(filters.startDate && filters.endDate ? "custom" : null);
    }
  }, [showDatePopup, filters.startDate, filters.endDate]);

  const loadFilterOptions = async () => {
    try {
      const [vehiclesRes, driversRes, customersRes] = await Promise.all([
        api.getVehicles({ limit: 500 }),
        api.getDrivers({ limit: 500 }),
        api.getTripCustomers(),
      ]);
      setVehicles(vehiclesRes.vehicles || vehiclesRes.data || []);
      setDrivers(driversRes.drivers || driversRes.data || []);
      setCustomers(customersRes.customers || []);
    } catch (error) {
      console.error("Failed to load filter options:", error);
    }
  };

  const updateFilter = <K extends keyof TripFilterState>(
    key: K,
    value: TripFilterState[K],
  ) => {
    onFilterChange({ ...filters, [key]: value });
  };

  const getVehicleLabel = (v: VehicleOption) =>
    v.registration_number || v.plate_number || v.name || v.id.slice(0, 8);

  // --- Date Range ---
  const applyPreset = (key: DatePreset) => {
    setDatePreset(key);
    setDateError("");
    let from = "", to = "";
    switch (key) {
      case "today":
        from = to = fmt(today());
        break;
      case "yesterday":
        from = to = fmt(daysAgo(1));
        break;
      case "last7":
        from = fmt(daysAgo(7));
        to = fmt(today());
        break;
      case "last30":
        from = fmt(daysAgo(30));
        to = fmt(today());
        break;
      case "custom":
        from = tempStart;
        to = tempEnd;
        break;
    }
    setTempStart(from);
    setTempEnd(to);
  };

  const validateDates = (): boolean => {
    if (tempStart && tempEnd && tempEnd < tempStart) {
      setDateError("End date cannot be before start date");
      return false;
    }
    setDateError("");
    return true;
  };

  const applyDateRange = () => {
    if (!validateDates()) return;
    updateFilter("startDate", tempStart);
    updateFilter("endDate", tempEnd);
    setShowDatePopup(false);
  };

  const clearDateSelection = () => {
    setTempStart("");
    setTempEnd("");
    setDatePreset(null);
    setDateError("");
  };

  const dateRangeLabel = () => {
    if (filters.startDate && filters.endDate)
      return `${filters.startDate} → ${filters.endDate}`;
    if (filters.startDate) return `From ${filters.startDate}`;
    if (filters.endDate) return `Until ${filters.endDate}`;
    return "Date Range";
  };

  // --- Drawer ---
  const openDrawer = () => setShowDrawer(true);
  const closeDrawer = () => setShowDrawer(false);
  const resetDrawerFilters = () => {
    updateFilter("tripType", "");
    updateFilter("customerName", "");
  };

  return (
    <>
      {/* Primary Filter Bar */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex items-center gap-3 flex-shrink-0 z-10">
        {/* Status */}
        <select
          value={filters.status}
          onChange={(e) => updateFilter("status", e.target.value)}
          className="bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-border-color min-w-[140px]"
        >
          <option value="">{uiStrings.filter_all_statuses || "All Statuses"}</option>
          <option value="scheduled">Scheduled</option>
          <option value="in_progress">In Progress</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>

        {/* Date Range Button */}
        <div ref={datePopupRef} className="relative">
          <button
            onClick={() => setShowDatePopup(!showDatePopup)}
            className={`flex items-center gap-2 bg-surface-dark border text-sm rounded-lg px-3 py-2 hover:bg-border-color transition-colors min-w-[180px] justify-between ${
              filters.startDate || filters.endDate
                ? "border-primary/40 text-primary"
                : "border-border-color text-text-muted"
            }`}
          >
            <span className="flex items-center gap-2">
              <span className="material-symbols-outlined text-sm">calendar_today</span>
              <span className="truncate">{dateRangeLabel()}</span>
            </span>
            {(filters.startDate || filters.endDate) && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  updateFilter("startDate", "");
                  updateFilter("endDate", "");
                }}
                className="text-text-muted hover:text-white"
              >
                <span className="material-symbols-outlined text-sm">close</span>
              </button>
            )}
          </button>

          {showDatePopup && (
            <div className="absolute top-full left-0 mt-2 bg-[#111827] border border-border-color rounded-lg shadow-2xl z-50 w-[420px] overflow-hidden">
              <div className="px-4 py-3 border-b border-border-color">
                <h4 className="text-white text-sm font-semibold">Filter by Date</h4>
              </div>

              <div className="flex">
                {/* Sidebar Presets */}
                <div className="w-[140px] bg-[#0f1419] border-r border-border-color py-2 flex-shrink-0">
                  {PRESETS.map((p) => (
                    <button
                      key={p.key}
                      onClick={() => applyPreset(p.key)}
                      className={`w-full text-left px-3 py-2 text-sm flex items-center gap-2 transition-colors ${
                        datePreset === p.key
                          ? "bg-primary/10 text-primary border-r-2 border-primary"
                          : "text-text-muted hover:bg-border-color hover:text-white"
                      }`}
                    >
                      <span className="material-symbols-outlined text-[16px]">{p.icon}</span>
                      <span className="truncate">{p.label}</span>
                    </button>
                  ))}
                </div>

                {/* Custom Range */}
                <div className="flex-1 p-4">
                  {datePreset === "custom" || !datePreset ? (
                    <>
                      <p className="text-xs text-text-muted mb-3">Select a custom date range</p>
                      <div className="space-y-3">
                        <div>
                          <label className="text-xs text-text-muted mb-1 block font-medium">From</label>
                          <input
                            type="date"
                            value={tempStart}
                            onChange={(e) => { setTempStart(e.target.value); setDateError(""); }}
                            className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary"
                          />
                        </div>
                        <div>
                          <label className="text-xs text-text-muted mb-1 block font-medium">To</label>
                          <input
                            type="date"
                            value={tempEnd}
                            onChange={(e) => { setTempEnd(e.target.value); setDateError(""); }}
                            className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary"
                          />
                        </div>
                        {dateError && (
                          <p className="text-xs text-red-400 flex items-center gap-1">
                            <span className="material-symbols-outlined text-sm">error</span>
                            {dateError}
                          </p>
                        )}
                      </div>
                    </>
                  ) : (
                    <div className="flex flex-col items-center justify-center h-full py-6 text-center">
                      <span className="material-symbols-outlined text-3xl text-text-muted mb-2">check_circle</span>
                      <p className="text-sm text-white font-medium">
                        {PRESETS.find((p) => p.key === datePreset)?.label}
                      </p>
                      <p className="text-xs text-text-muted mt-1">{tempStart} → {tempEnd}</p>
                    </div>
                  )}
                </div>
              </div>

              {/* Footer */}
              <div className="border-t border-border-color px-4 py-3 flex items-center justify-between bg-[#0f1419]">
                <button
                  onClick={() => { clearDateSelection(); updateFilter("startDate", ""); updateFilter("endDate", ""); }}
                  className="text-sm text-text-muted hover:text-white transition-colors px-3 py-1.5 rounded"
                >
                  Clear
                </button>
                <button
                  onClick={applyDateRange}
                  className="bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded-lg px-5 py-2 transition-colors shadow-lg shadow-primary/20"
                >
                  Apply
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Sort */}
        <select
          value={filters.orderBy}
          onChange={(e) => updateFilter("orderBy", e.target.value)}
          className="bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-border-color min-w-[140px]"
        >
          <option value="">{uiStrings.sort_default || "Sort: Newest"}</option>
          <option value="created_at">Created Date</option>
          <option value="scheduled_start">Scheduled Start</option>
          <option value="status">Status</option>
        </select>

        {filters.orderBy && (
          <button
            onClick={() => updateFilter("orderDir", filters.orderDir === "ASC" ? "DESC" : "ASC")}
            className="px-2 py-2 bg-surface-dark border border-border-color text-sm text-text-muted hover:text-white hover:bg-border-color rounded-lg"
          >
            {filters.orderDir === "ASC" ? "↑" : "↓"}
          </button>
        )}

        <div className="flex-1" />

        {/* Advanced Filters */}
        <button
          onClick={openDrawer}
          className="flex items-center gap-2 px-3 py-2 bg-surface-dark border border-border-color text-sm text-text-muted font-medium rounded-lg hover:text-white hover:bg-border-color transition-colors"
        >
          <span className="material-symbols-outlined text-[18px]">tune</span>
          Filters
          {drawerAdvancedCount > 0 && (
            <span className="bg-primary/20 text-primary text-xs px-1.5 py-0.5 rounded-full font-bold">
              {drawerAdvancedCount}
            </span>
          )}
        </button>

        {/* Clear All */}
        {totalActiveCount > 0 && (
          <button
            onClick={onReset}
            className="flex items-center gap-2 px-3 py-2 text-sm text-text-muted hover:text-white transition-colors"
          >
            <span className="material-symbols-outlined text-[16px]">close</span>
            Clear ({totalActiveCount})
          </button>
        )}
      </div>

      {/* Advanced Filters Side Drawer */}
      {showDrawer && (
        <>
          <div className="fixed inset-0 bg-black/50 z-40" onClick={closeDrawer} />
          <div className="fixed top-0 right-0 h-full w-80 bg-surface-dark border-l border-border-color shadow-2xl z-50 flex flex-col">
            <div className="flex items-center justify-between px-5 py-4 border-b border-border-color">
              <h3 className="text-white font-semibold text-base">Advanced Filters</h3>
              <button onClick={closeDrawer} className="text-text-muted hover:text-white p-1">
                <span className="material-symbols-outlined text-xl">close</span>
              </button>
            </div>

            <div className="flex-1 overflow-y-auto px-5 py-4 space-y-5">
              <div>
                <label className="text-xs text-text-muted uppercase tracking-wider font-semibold mb-2 block">
                  Trip Type
                </label>
                <select
                  value={filters.tripType}
                  onChange={(e) => updateFilter("tripType", e.target.value)}
                  className="w-full bg-surface-darker border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer"
                >
                  <option value="">All Types</option>
                  <option value="delivery">Delivery</option>
                  <option value="service">Service</option>
                  <option value="long_haul">Long Haul</option>
                  <option value="local">Local</option>
                  <option value="return">Return</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-text-muted uppercase tracking-wider font-semibold mb-2 block">
                  Organization
                </label>
                <select
                  value={filters.customerName}
                  onChange={(e) => updateFilter("customerName", e.target.value)}
                  className="w-full bg-surface-darker border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer"
                >
                  <option value="">All Organizations</option>
                  {customers.map((c) => (
                    <option key={c} value={c}>{c}</option>
                  ))}
                </select>
              </div>

              {totalActiveCount > 0 && (
                <div className="bg-surface-darker border border-border-color rounded-lg p-3">
                  <h4 className="text-xs text-text-muted uppercase font-semibold mb-2">Active Filters</h4>
                  <div className="flex flex-wrap gap-1.5">
                    {filters.status && (
                      <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full">
                        Status: {filters.status.replace("_", " ")}
                      </span>
                    )}
                    {filters.tripType && (
                      <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full">
                        Type: {filters.tripType.replace("_", " ")}
                      </span>
                    )}
                    {filters.customerName && (
                      <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full truncate max-w-[200px]">
                        Org: {filters.customerName}
                      </span>
                    )}
                    {filters.driverId && (
                      <span className="bg-warning/10 text-warning text-xs px-2 py-0.5 rounded-full">Driver</span>
                    )}
                    {filters.vehicleId && (
                      <span className="bg-warning/10 text-warning text-xs px-2 py-0.5 rounded-full">Vehicle</span>
                    )}
                    {(filters.startDate || filters.endDate) && (
                      <span className="bg-warning/10 text-warning text-xs px-2 py-0.5 rounded-full">
                        {dateRangeLabel()}
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>

            <div className="border-t border-border-color px-5 py-4 flex gap-2">
              <button
                onClick={() => { resetDrawerFilters(); closeDrawer(); }}
                className="flex-1 text-sm text-text-muted hover:text-white py-2.5 border border-border-color rounded-lg"
              >
                Reset
              </button>
              <button
                onClick={closeDrawer}
                className="flex-1 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded-lg py-2.5"
              >
                Apply
              </button>
            </div>
          </div>
        </>
      )}
    </>
  );
};
