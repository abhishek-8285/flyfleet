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
     <div className="w-full bg-surface-dim border-b border-outline px-6 py-4 flex items-center gap-4 flex-shrink-0 z-10 shadow-lg">
        {/* Status */}
        <div className="flex flex-col">
          <label className="text-[9px] font-black text-on-surface-variant uppercase tracking-[0.2em] mb-1.5 ml-1">Fleet Status</label>
          <select
            value={filters.status}
            onChange={(e) => updateFilter("status", e.target.value)}
            className="bg-surface-container border border-outline text-white text-xs font-bold rounded-xl px-4 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-surface-container-highest transition-all min-w-[160px] outline-none shadow-sm"
          >
            <option value="">{uiStrings.filter_all_statuses || "All Trips"}</option>
            <option value="scheduled">Scheduled</option>
            <option value="in_progress">In Transit</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>

        {/* Date Range Button */}
        <div ref={datePopupRef} className="relative flex flex-col">
          <label className="text-[9px] font-black text-on-surface-variant uppercase tracking-[0.2em] mb-1.5 ml-1">Operational Window</label>
          <button
            onClick={() => setShowDatePopup(!showDatePopup)}
            className={`flex items-center gap-3 bg-surface-container border text-xs font-bold rounded-xl px-4 py-2.5 hover:bg-surface-container-highest transition-all min-w-[220px] justify-between shadow-sm outline-none ${
              filters.startDate || filters.endDate
                ? "border-primary text-primary"
                : "border-outline text-on-surface-variant"
            }`}
          >
            <span className="flex items-center gap-2">
              <span className="material-symbols-outlined text-[18px]">calendar_today</span>
              <span className="truncate">{dateRangeLabel()}</span>
            </span>
            {(filters.startDate || filters.endDate) && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  updateFilter("startDate", "");
                  updateFilter("endDate", "");
                }}
                className="text-on-surface-variant hover:text-white transition-colors"
              >
                <span className="material-symbols-outlined text-sm font-black">close</span>
              </button>
            )}
          </button>

          {showDatePopup && (
            <div className="absolute top-[calc(100%+8px)] left-0 bg-surface-container border border-outline rounded-2xl shadow-2xl z-50 w-[440px] overflow-hidden">
              <div className="px-5 py-4 border-b border-outline bg-surface-dim">
                <h4 className="text-white text-[11px] font-black uppercase tracking-[0.15em]">Timeframe Selection</h4>
              </div>

              <div className="flex">
                {/* Sidebar Presets */}
                <div className="w-[150px] bg-surface-dim border-r border-outline py-3 flex-shrink-0">
                  {PRESETS.map((p) => (
                    <button
                      key={p.key}
                      onClick={() => applyPreset(p.key)}
                      className={`w-full text-left px-4 py-3 text-[10px] font-black uppercase tracking-widest flex items-center gap-3 transition-all ${
                        datePreset === p.key
                          ? "bg-primary text-white"
                          : "text-on-surface-variant hover:bg-surface-container hover:text-white"
                      }`}
                    >
                      <span className="material-symbols-outlined text-[18px]">{p.icon}</span>
                      <span className="truncate">{p.label}</span>
                    </button>
                  ))}
                </div>

                {/* Custom Range */}
                <div className="flex-1 p-5">
                  {datePreset === "custom" || !datePreset ? (
                    <>
                      <p className="text-[10px] text-on-surface-variant font-black uppercase tracking-widest mb-4">Manual Selection</p>
                      <div className="space-y-4">
                        <div>
                          <label className="text-[9px] font-black text-on-surface-variant uppercase tracking-widest mb-1.5 block">Start Date</label>
                          <input
                            type="date"
                            value={tempStart}
                            onChange={(e) => { setTempStart(e.target.value); setDateError(""); }}
                            className="w-full bg-surface-dim border border-outline text-white text-xs font-bold rounded-xl px-4 py-3 focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                          />
                        </div>
                        <div>
                          <label className="text-[9px] font-black text-on-surface-variant uppercase tracking-widest mb-1.5 block">End Date</label>
                          <input
                            type="date"
                            value={tempEnd}
                            onChange={(e) => { setTempEnd(e.target.value); setDateError(""); }}
                            className="w-full bg-surface-dim border border-outline text-white text-xs font-bold rounded-xl px-4 py-3 focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                          />
                        </div>
                        {dateError && (
                          <p className="text-[10px] text-rose-400 font-bold flex items-center gap-2 mt-2">
                            <span className="material-symbols-outlined text-[14px]">error</span>
                            {dateError}
                          </p>
                        )}
                      </div>
                    </>
                  ) : (
                    <div className="flex flex-col items-center justify-center h-full py-8 text-center">
                      <span className="material-symbols-outlined text-4xl text-primary/40 mb-3 animate-pulse">history</span>
                      <p className="text-[11px] text-white font-black uppercase tracking-widest">
                        {PRESETS.find((p) => p.key === datePreset)?.label}
                      </p>
                      <p className="text-[10px] text-on-surface-variant font-bold mt-2 font-mono">{tempStart} / {tempEnd}</p>
                    </div>
                  )}
                </div>
              </div>

              {/* Footer */}
              <div className="border-t border-outline px-5 py-4 flex items-center justify-between bg-surface-dim">
                <button
                  onClick={() => { clearDateSelection(); updateFilter("startDate", ""); updateFilter("endDate", ""); }}
                  className="text-[10px] font-black uppercase tracking-widest text-on-surface-variant hover:text-white transition-all px-4 py-2 hover:bg-surface-container rounded-xl"
                >
                  Clear Selection
                </button>
                <button
                  onClick={applyDateRange}
                  className="bg-primary hover:bg-primary-dark text-white text-[10px] font-black uppercase tracking-widest rounded-xl px-6 py-2.5 transition-all shadow-lg shadow-primary/20 active:scale-95"
                >
                  Apply Filter
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Sort */}
        <div className="flex flex-col">
          <label className="text-[9px] font-black text-on-surface-variant uppercase tracking-[0.2em] mb-1.5 ml-1">Order By</label>
          <div className="flex items-center gap-2">
            <select
              value={filters.orderBy}
              onChange={(e) => updateFilter("orderBy", e.target.value)}
              className="bg-surface-container border border-outline text-white text-xs font-bold rounded-xl px-4 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-surface-container-highest transition-all min-w-[140px] outline-none shadow-sm"
            >
              <option value="">{uiStrings.sort_default || "Newest Entry"}</option>
              <option value="created_at">Booking Date</option>
              <option value="scheduled_start">Trip Start</option>
              <option value="status">Status</option>
            </select>
            {filters.orderBy && (
              <button
                onClick={() => updateFilter("orderDir", filters.orderDir === "ASC" ? "DESC" : "ASC")}
                className="p-2.5 bg-surface-container border border-outline text-on-surface-variant hover:text-white hover:bg-white/5 rounded-xl transition-all active:scale-95 shadow-sm"
              >
                <span className="material-symbols-outlined text-[18px]">
                  {filters.orderDir === "ASC" ? "stat_1" : "stat_minus_1"}
                </span>
              </button>
            )}
          </div>
        </div>

        <div className="flex-1 mt-5" />

        {/* Global Action Area */}
        <div className="flex items-center gap-3 mt-5">
          {/* Advanced Filters */}
          <button
            onClick={openDrawer}
            className="flex items-center gap-2 px-5 py-2.5 bg-surface-container border border-outline text-xs font-black uppercase tracking-widest text-on-surface-variant rounded-xl hover:text-white hover:bg-surface-container-highest transition-all shadow-sm active:scale-95"
          >
            <span className="material-symbols-outlined text-[18px]">tune</span>
            Detailed Filters
            {drawerAdvancedCount > 0 && (
              <span className="bg-primary text-white text-[10px] px-2 py-0.5 rounded-full font-black ml-1">
                {drawerAdvancedCount}
              </span>
            )}
          </button>

          {/* Clear All */}
          {totalActiveCount > 0 && (
            <button
              onClick={onReset}
              className="flex items-center gap-2 px-4 py-2.5 text-[10px] font-black uppercase tracking-[0.15em] text-on-surface-variant hover:text-rose-400 transition-all active:scale-95"
            >
              <span className="material-symbols-outlined text-[18px]">close</span>
              Reset All ({totalActiveCount})
            </button>
          )}
        </div>
      </div>

          {/* Advanced Filters Side Drawer */}
          {showDrawer && (
            <>
              <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 transition-all" onClick={closeDrawer} />
              <div className="fixed top-0 right-0 h-full w-[360px] bg-surface-dim border-l border-outline shadow-[0_0_50px_rgba(0,0,0,0.5)] z-50 flex flex-col">
                <div className="flex items-center justify-between px-6 py-5 border-b border-outline bg-surface-container/30">
                  <h3 className="text-white font-black text-xs uppercase tracking-[.2em]">Detailed Analytics</h3>
                  <button onClick={closeDrawer} className="text-on-surface-variant hover:text-white p-2 hover:bg-surface-container rounded-xl transition-all">
                    <span className="material-symbols-outlined text-xl font-black">close</span>
                  </button>
                </div>

                <div className="flex-1 overflow-y-auto px-6 py-6 space-y-8">
                  <div>
                    <label className="text-[10px] text-on-surface-variant uppercase tracking-[.25em] font-black mb-3 block">
                      Internal Asset Type
                    </label>
                    <select
                      value={filters.tripType}
                      onChange={(e) => updateFilter("tripType", e.target.value)}
                      className="w-full bg-surface-container border border-outline text-white text-xs font-bold rounded-xl px-4 py-3.5 focus:ring-1 focus:ring-primary outline-none hover:bg-surface-container-highest cursor-pointer transition-all"
                    >
                      <option value="">Consolidated Types</option>
                      <option value="delivery">Primary Delivery</option>
                      <option value="service">Technical Service</option>
                      <option value="long_haul">Long Haul Corridor</option>
                      <option value="local">Local Distribution</option>
                      <option value="return">Return Material</option>
                    </select>
                  </div>

                  <div>
                    <label className="text-[10px] text-on-surface-variant uppercase tracking-[.25em] font-black mb-3 block">
                      Partner Organization
                    </label>
                    <select
                      value={filters.customerName}
                      onChange={(e) => updateFilter("customerName", e.target.value)}
                      className="w-full bg-surface-container border border-outline text-white text-xs font-bold rounded-xl px-4 py-3.5 focus:ring-1 focus:ring-primary outline-none hover:bg-surface-container-highest cursor-pointer transition-all"
                    >
                      <option value="">All Partners</option>
                      {customers.map((c) => (
                        <option key={c} value={c}>{c}</option>
                      ))}
                    </select>
                  </div>

                  {totalActiveCount > 0 && (
                    <div className="bg-surface-container border border-outline rounded-2xl p-4 shadow-inner">
                      <h4 className="text-[9px] text-on-surface-variant uppercase font-black tracking-widest mb-3 border-b border-outline/30 pb-2">Active Operational Constraints</h4>
                      <div className="flex flex-wrap gap-2">
                        {filters.status && (
                          <span className="bg-primary/20 text-primary text-[9px] px-3 py-1 rounded-full font-black uppercase tracking-wider border border-primary/20">
                             {filters.status.replace("_", " ")}
                          </span>
                        )}
                        {filters.tripType && (
                          <span className="bg-primary/20 text-primary text-[9px] px-3 py-1 rounded-full font-black uppercase tracking-wider border border-primary/20">
                            {filters.tripType.replace("_", " ")}
                          </span>
                        )}
                        {filters.customerName && (
                          <span className="bg-primary/20 text-primary text-[9px] px-3 py-1 rounded-full font-black uppercase tracking-wider border border-primary/20 truncate max-w-[180px]">
                            {filters.customerName}
                          </span>
                        )}
                        {(filters.startDate || filters.endDate) && (
                          <span className="bg-amber-500/10 text-amber-500 text-[9px] px-3 py-1 rounded-full font-black uppercase tracking-wider border border-amber-500/20">
                            {dateRangeLabel()}
                          </span>
                        )}
                      </div>
                    </div>
                  )}
                </div>

                <div className="border-t border-outline px-6 py-5 flex gap-3 bg-surface-container/20">
                  <button
                    onClick={() => { resetDrawerFilters(); closeDrawer(); }}
                    className="flex-1 text-[10px] font-black uppercase tracking-widest text-on-surface-variant hover:text-white py-3.5 border border-outline rounded-xl hover:bg-surface-container transition-all active:scale-95"
                  >
                    Clear All
                  </button>
                  <button
                    onClick={closeDrawer}
                    className="flex-1 bg-primary hover:bg-primary-dark text-white text-[10px] font-black uppercase tracking-widest rounded-xl py-3.5 shadow-lg shadow-primary/20 transition-all active:scale-95"
                  >
                    Confirm Filters
                  </button>
                </div>
              </div>
            </>
          )}
    </>
  );
};
