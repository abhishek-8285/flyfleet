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
      const [driversRes, orgsRes] = await Promise.all([
        apiClient.get("/admin/drivers?limit=500"),
        apiClient.get("/admin/organizations?limit=500"),
      ]);
      setDrivers(driversRes.data?.drivers || driversRes.data || []);
      setOrganizations(orgsRes.data?.organizations || orgsRes.data || []);

      // Extract makes from vehicles
      const vehiclesRes = await apiClient.get("/admin/vehicles?limit=1000");
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
      console.error("Failed to load filter options:", error);
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
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex items-center gap-3 flex-shrink-0 z-10">
        {/* Status */}
        <select
          value={filters.status}
          onChange={(e) => updateFilter("status", e.target.value)}
          className="bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-border-color min-w-[140px]"
        >
          <option value="">All Statuses</option>
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>

        {/* Sort */}
        <div ref={sortRef} className="flex items-center gap-1">
          <select
            value={filters.orderBy}
            onChange={(e) => updateFilter("orderBy", e.target.value)}
            className="bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary cursor-pointer hover:bg-border-color min-w-[150px]"
          >
            <option value="">Sort: Newest</option>
            <option value="created_at">Created Date</option>
            <option value="registration_number">Registration No</option>
            <option value="make">Make / Model</option>
            <option value="status">Status</option>
            <option value="odometer">Odometer</option>
          </select>

          {filters.orderBy && (
            <button
              onClick={() =>
                updateFilter("orderDir", filters.orderDir === "ASC" ? "DESC" : "ASC")
              }
              className="px-2 py-2 bg-surface-dark border border-border-color text-sm text-text-muted hover:text-white hover:bg-border-color rounded-lg"
              title={filters.orderDir === "ASC" ? "Ascending" : "Descending"}
            >
              {filters.orderDir === "ASC" ? "↑" : "↓"}
            </button>
          )}
        </div>

        <div className="flex-1" />

        {/* Total Count */}
        {totalCount !== undefined && (
          <span className="text-xs text-text-muted px-2">
            {totalCount} vehicle{totalCount !== 1 ? "s" : ""}
          </span>
        )}

        {/* Advanced Filters Button */}
        <button
          onClick={openDrawer}
          className={`flex items-center gap-2 px-3 py-2 border text-sm font-medium rounded-lg transition-colors ${
            drawerAdvancedCount > 0
              ? "bg-primary/10 border-primary/40 text-primary hover:bg-primary/20"
              : "bg-surface-dark border-border-color text-text-muted hover:text-white hover:bg-border-color"
          }`}
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
            {/* Drawer Header */}
            <div className="flex items-center justify-between px-5 py-4 border-b border-border-color">
              <h3 className="text-white font-semibold text-base">
                Vehicle Filters
              </h3>
              <button
                onClick={closeDrawer}
                className="text-text-muted hover:text-white p-1"
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

            {/* Drawer Footer */}
            <div className="border-t border-border-color px-5 py-4 flex gap-2">
              <button
                onClick={() => {
                  resetDrawerFilters();
                }}
                className="flex-1 text-sm text-text-muted hover:text-white py-2.5 border border-border-color rounded-lg transition-colors"
              >
                Reset Filters
              </button>
              <button
                onClick={closeDrawer}
                className="flex-1 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded-lg py-2.5 transition-colors shadow-lg shadow-primary/20"
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
