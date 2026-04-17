import React, { useEffect, useRef, useState } from "react";
import apiClient from "../../services/apiClient";

// ── Types ──
export interface VehicleFilterState {
  status: string;
  vehicleType: string;
  fuelType: string;
  driverId: string;
  organizationId: string;
  make: string;
  search: string;
  orderBy: string;
  orderDir: "ASC" | "DESC";
}

interface VehicleFiltersProps {
  filters: VehicleFilterState;
  onFilterChange: (filters: VehicleFilterState) => void;
  onReset: () => void;
  totalCount?: number;
}

interface DriverOption {
  id: string;
  name: string;
}

interface OrgOption {
  id: string;
  name: string;
}

const STATUS_OPTIONS = [
  { value: "active", label: "Active" },
  { value: "maintenance", label: "Maintenance" },
  { value: "inactive", label: "Inactive" },
  { value: "idle", label: "Idle" },
];

const TYPE_OPTIONS = [
  { value: "sedan", label: "Sedan" },
  { value: "suv", label: "SUV" },
  { value: "truck", label: "Truck" },
  { value: "van", label: "Van" },
  { value: "heavy_haul", label: "Heavy Haul" },
];

const FUEL_OPTIONS = [
  { value: "gasoline", label: "Gasoline" },
  { value: "diesel", label: "Diesel" },
  { value: "electric", label: "Electric" },
  { value: "hybrid", label: "Hybrid" },
  { value: "cng", label: "CNG" },
];

export const VehicleFilters: React.FC<VehicleFiltersProps> = ({
  filters,
  onFilterChange,
  onReset,
  totalCount,
}) => {
  const [drivers, setDrivers] = useState<DriverOption[]>([]);
  const [organizations, setOrganizations] = useState<OrgOption[]>([]);
  const [makes, setMakes] = useState<string[]>([]);

  // Advanced side drawer
  const [showDrawer, setShowDrawer] = useState(false);

  const drawerAdvancedCount =
    (filters.vehicleType ? 1 : 0) +
    (filters.fuelType ? 1 : 0) +
    (filters.driverId ? 1 : 0) +
    (filters.organizationId ? 1 : 0) +
    (filters.make ? 1 : 0);

  const totalActiveCount =
    (filters.status ? 1 : 0) + drawerAdvancedCount;

  const sortRef = useRef<HTMLDivElement>(null);
  const drawerRef = useRef<HTMLDivElement>(null);

  // Load filter options on mount
  useEffect(() => {
    loadFilterOptions();
  }, []);

  // Click outside to close drawer
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        drawerRef.current &&
        !drawerRef.current.contains(e.target as Node)
      ) {
        setShowDrawer(false);
      }
    };
    if (showDrawer) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [showDrawer]);

  const loadFilterOptions = async () => {
    try {
      // Load drivers (orgs endpoint doesn't exist yet, skip it)
      const driversRes = await apiClient.get("/admin/drivers?limit=500").catch(() => ({ data: [] }));
      setDrivers(driversRes.data?.drivers || driversRes.data || []);

      // Extract makes from vehicles
      const vehiclesRes = await apiClient.get("/admin/vehicles?limit=1000").catch(() => ({ data: [] }));
      const allVehicles = vehiclesRes.data?.vehicles || vehiclesRes.data || [];
      const uniqueMakes = Array.from(
        new Set(
          allVehicles
            .map((v: any) => v.make)
            .filter(Boolean)
        )
      ).sort();
      setMakes(uniqueMakes as string[]);
    } catch (error) {
      // Silently fail — filters will show empty dropdowns
    }
  };

  const updateFilter = <K extends keyof VehicleFilterState>(
    key: K,
    value: VehicleFilterState[K],
  ) => {
    onFilterChange({ ...filters, [key]: value });
  };

  const openDrawer = () => setShowDrawer(true);
  const closeDrawer = () => setShowDrawer(false);

  const resetDrawerFilters = () => {
    updateFilter("vehicleType", "");
    updateFilter("fuelType", "");
    updateFilter("driverId", "");
    updateFilter("organizationId", "");
    updateFilter("make", "");
  };

  const handleResetAll = () => {
    resetDrawerFilters();
    onReset();
    closeDrawer();
  };

  return (
    <>
      {/* Primary Filter Bar */}
      <div className="w-full bg-surface-container/30 border-b border-outline px-6 py-3 flex items-center gap-3 flex-shrink-0 z-10">
        {/* Status */}
        <select
          value={filters.status}
          onChange={(e) => updateFilter("status", e.target.value)}
          className="bg-surface-container border border-outline text-white text-[11px] font-bold uppercase tracking-wider rounded-lg px-4 py-2 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-surface-container-high transition-all min-w-[160px]"
        >
          <option value="">Operational Status</option>
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label.toUpperCase()}
            </option>
          ))}
        </select>

        {/* Sort Controls */}
        <div ref={sortRef} className="flex items-center gap-1">
          <select
            value={filters.orderBy}
            onChange={(e) => updateFilter("orderBy", e.target.value)}
            className="bg-surface-container border border-outline text-white text-[11px] font-bold uppercase tracking-wider rounded-lg px-4 py-2 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-surface-container-high transition-all min-w-[160px]"
          >
            <option value="">Sort: Default</option>
            <option value="created_at">Join Date</option>
            <option value="registration_number">Reg Number</option>
            <option value="make">Manufacturer</option>
            <option value="status">Current Status</option>
            <option value="odometer">Odometer Reading</option>
          </select>

          {filters.orderBy && (
            <button
              onClick={() =>
                updateFilter("orderDir", filters.orderDir === "ASC" ? "DESC" : "ASC")
              }
              className="px-3 py-2 bg-surface-container border border-outline text-xs text-on-surface-variant hover:text-white hover:bg-surface-container-high rounded-lg transition-all"
              title={filters.orderDir === "ASC" ? "Ascending" : "Descending"}
            >
              {filters.orderDir === "ASC" ? (
                <span className="material-symbols-outlined text-[18px]">arrow_upward</span>
              ) : (
                <span className="material-symbols-outlined text-[18px]">arrow_downward</span>
              )}
            </button>
          )}
        </div>

        <div className="flex-1" />

        {/* Asset Metrics Summary */}
        {totalCount !== undefined && (
          <span className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider px-3 py-1 bg-surface-container/50 rounded-full border border-outline">
            {totalCount} Total Assets
          </span>
        )}

        {/* Advanced Filter Trigger */}
        <button
          onClick={openDrawer}
          className={`flex items-center gap-2 px-4 py-2 border text-[11px] font-bold uppercase tracking-wider rounded-lg transition-all ${
            drawerAdvancedCount > 0
              ? "bg-primary/10 border-primary/40 text-primary hover:bg-primary/20 shadow-lg shadow-primary/10"
              : "bg-surface-container border-outline text-on-surface-variant hover:text-white hover:bg-surface-container-high"
          }`}
        >
          <span className="material-symbols-outlined text-[18px]">tune</span>
          Advanced Filters
          {drawerAdvancedCount > 0 && (
            <span className="ml-1 bg-primary/20 text-primary text-[10px] px-2 py-0.5 rounded-md font-bold">
              {drawerAdvancedCount}
            </span>
          )}
        </button>

        {/* Clear All */}
        {totalActiveCount > 0 && (
          <button
            onClick={handleResetAll}
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
          <div
            className="fixed inset-0 bg-black/50 z-40"
            onClick={closeDrawer}
          />
          <div
            ref={drawerRef}
            className="fixed top-0 right-0 h-full w-80 bg-surface-dark border-l border-border-color shadow-2xl z-50 flex flex-col"
          >
            {/* Drawer Header Area */}
            <div className="flex items-center justify-between px-6 py-5 border-b border-outline bg-surface-container/50">
              <h3 className="text-white font-bold text-sm uppercase tracking-wider">
                Vehicle Intelligence
              </h3>
              <button
                onClick={closeDrawer}
                className="text-on-surface-variant hover:text-white p-1 hover:bg-white/5 rounded-lg transition-all"
              >
                <span className="material-symbols-outlined text-xl">close</span>
              </button>
            </div>

            {/* Drawer Content */}
            <div className="flex-1 overflow-y-auto px-5 py-4 space-y-5 custom-scrollbar">
              {/* Vehicle Type */}
              <div>
                <label className="text-xs text-text-muted uppercase tracking-wider font-semibold mb-2 block">
                  Vehicle Type
                </label>
                <select
                  value={filters.vehicleType}
                  onChange={(e) => updateFilter("vehicleType", e.target.value)}
                  className="w-full bg-surface-darker border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer"
                >
                  <option value="">All Types</option>
                  {TYPE_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Fuel Type */}
              <div>
                <label className="text-xs text-text-muted uppercase tracking-wider font-semibold mb-2 block">
                  Fuel Type
                </label>
                <select
                  value={filters.fuelType}
                  onChange={(e) => updateFilter("fuelType", e.target.value)}
                  className="w-full bg-surface-darker border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer"
                >
                  <option value="">All Fuel Types</option>
                  {FUEL_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Make / Brand */}
              <div>
                <label className="text-xs text-text-muted uppercase tracking-wider font-semibold mb-2 block">
                  Make / Brand
                </label>
                <select
                  value={filters.make}
                  onChange={(e) => updateFilter("make", e.target.value)}
                  className="w-full bg-surface-darker border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer"
                >
                  <option value="">All Makes</option>
                  {makes.map((m) => (
                    <option key={m} value={m}>
                      {m}
                    </option>
                  ))}
                </select>
              </div>

              {/* Assigned Driver */}
              <div>
                <label className="text-xs text-text-muted uppercase tracking-wider font-semibold mb-2 block">
                  Assigned Driver
                </label>
                <select
                  value={filters.driverId}
                  onChange={(e) => updateFilter("driverId", e.target.value)}
                  className="w-full bg-surface-darker border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer"
                >
                  <option value="">All Drivers</option>
                  {drivers.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Organization */}
              <div>
                <label className="text-xs text-text-muted uppercase tracking-wider font-semibold mb-2 block">
                  Organization
                </label>
                <select
                  value={filters.organizationId}
                  onChange={(e) =>
                    updateFilter("organizationId", e.target.value)
                  }
                  className="w-full bg-surface-darker border border-border-color text-white text-sm rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary cursor-pointer"
                >
                  <option value="">All Organizations</option>
                  {organizations.map((o) => (
                    <option key={o.id} value={o.id}>
                      {o.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Active Filters Summary */}
              {totalActiveCount > 0 && (
                <div className="bg-surface-darker border border-border-color rounded-lg p-3">
                  <h4 className="text-xs text-text-muted uppercase font-semibold mb-2">
                    Active Filters
                  </h4>
                  <div className="flex flex-wrap gap-1.5">
                    {filters.status && (
                      <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full">
                        Status:{" "}
                        {STATUS_OPTIONS.find((o) => o.value === filters.status)
                          ?.label || filters.status}
                      </span>
                    )}
                    {filters.vehicleType && (
                      <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full">
                        Type:{" "}
                        {TYPE_OPTIONS.find((o) => o.value === filters.vehicleType)
                          ?.label || filters.vehicleType}
                      </span>
                    )}
                    {filters.fuelType && (
                      <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full">
                        Fuel:{" "}
                        {FUEL_OPTIONS.find((o) => o.value === filters.fuelType)
                          ?.label || filters.fuelType}
                      </span>
                    )}
                    {filters.make && (
                      <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full">
                        Make: {filters.make}
                      </span>
                    )}
                    {filters.driverId && (
                      <span className="bg-warning/10 text-warning text-xs px-2 py-0.5 rounded-full">
                        Driver:{" "}
                        {drivers.find((d) => d.id === filters.driverId)?.name ||
                          "Selected"}
                      </span>
                    )}
                    {filters.organizationId && (
                      <span className="bg-warning/10 text-warning text-xs px-2 py-0.5 rounded-full">
                        Org:{" "}
                        {organizations.find((o) => o.id === filters.organizationId)
                          ?.name || "Selected"}
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Action Bar Footer */}
            <div className="border-t border-outline px-6 py-5 flex gap-3 bg-surface-container/50">
              <button
                onClick={() => {
                  resetDrawerFilters();
                }}
                className="flex-1 text-[11px] font-bold uppercase tracking-wider text-on-surface-variant hover:text-white py-3 border border-outline rounded-xl transition-all hover:bg-white/5"
              >
                Reset All
              </button>
              <button
                onClick={closeDrawer}
                className="flex-1 bg-primary hover:bg-primary-dark text-white text-[11px] font-bold uppercase tracking-wider rounded-xl py-3 transition-all shadow-lg shadow-primary/20 active:scale-95"
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

export default VehicleFilters;
