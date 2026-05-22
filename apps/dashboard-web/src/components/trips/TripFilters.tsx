import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";
import { DateRangePicker } from "../common/DateRangePicker";

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

  // --- Drawer ---
  const openDrawer = () => setShowDrawer(true);
  const closeDrawer = () => setShowDrawer(false);
  const resetDrawerFilters = () => {
    updateFilter("tripType", "");
    updateFilter("customerName", "");
  };

  const dateRangeLabel = () => {
    if (filters.startDate && filters.endDate)
      return `${filters.startDate} → ${filters.endDate}`;
    if (filters.startDate) return `From ${filters.startDate}`;
    if (filters.endDate) return `Until ${filters.endDate}`;
    return "Date Range";
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

        {/* Reusable Date Range Picker */}
        <div className="flex flex-col animate-fadeIn">
          <label className="text-[9px] font-black text-on-surface-variant uppercase tracking-[0.2em] mb-1.5 ml-1">Operational Window</label>
          <DateRangePicker
            startDate={filters.startDate}
            endDate={filters.endDate}
            onChange={(start, end) => {
              onFilterChange({ ...filters, startDate: start, endDate: end });
            }}
            triggerLabel="Date Range"
          />
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
