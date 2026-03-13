import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { setSelectedVehicle } from "../../../store/telemetrySlice";
import { fleetService } from "../FleetService";
import { Vehicle } from "../types";

/**
 * FleetList: A high-performance reactive grid for the vehicle registry.
 * Leverages FleetService for worker-offloaded search and filtering.
 */
export const FleetList: React.FC = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [metadata, setMetadata] = useState<{
    orgs: string[];
    statuses: string[];
  }>({ orgs: [], statuses: [] });
  const selectedId = useSelector(
    (state: any) => state.telemetry.selectedVehicleId,
  );

  useEffect(() => {
    const init = async () => {
      setLoading(true);
      const meta = await fleetService.getFilterMetadata();
      setMetadata(meta);
      const results = await fleetService.searchVehicles("", {});
      setVehicles(results as Vehicle[]);
      setLoading(false);
    };
    init();
  }, []);

  const handleSearch = async (query: string) => {
    setSearchQuery(query);
    const results = await fleetService.searchVehicles(query, filters);
    setVehicles(results as Vehicle[]);
  };

  const handleFilterChange = async (key: string, value: string) => {
    const newFilters = { ...filters, [key]: value };
    if (!value) delete newFilters[key];
    setFilters(newFilters);
    const results = await fleetService.searchVehicles(searchQuery, newFilters);
    setVehicles(results as Vehicle[]);
  };

  const handleVehicleSelect = (id: string | null) => {
    dispatch(setSelectedVehicle(id));
    if (id) {
      // Navigate to details or just stay on list depending on UX preference
      // For enterprise apps, we often keep the list and update a side-panel,
      // but here we'll support both selection and navigation.
      navigate(`/admin/vehicles/${id}`);
    }
  };

  return (
    <div className="flex flex-col h-full bg-surface text-white">
      {/* Header & Controls */}
      <div className="p-6 border-b border-border-color bg-surface-dark">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold tracking-tight">Fleet Registry</h1>
          <div className="flex gap-2">
            <button className="px-4 py-2 bg-primary hover:bg-primary-hover rounded-lg font-medium transition-all">
              Export CSV
            </button>
            <button
              className="px-4 py-2 bg-surface-darker border border-border-color rounded-lg font-medium"
              onClick={() => navigate("/admin/vehicles/new")}
            >
              Add Asset
            </button>
          </div>
        </div>

        <div className="flex flex-wrap gap-4">
          <div className="relative flex-1 min-w-[300px]">
            <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">
              search
            </span>
            <input
              type="text"
              placeholder="Search VIN, ID, or Fleet Name..."
              className="w-full bg-surface-darker border border-border-color rounded-lg pl-10 pr-4 py-2 focus:ring-1 focus:ring-primary outline-none transition-all"
              value={searchQuery}
              onChange={(e) => handleSearch(e.target.value)}
            />
          </div>

          <select
            className="bg-surface-darker border border-border-color rounded-lg px-4 py-2 text-sm outline-none"
            onChange={(e) => handleFilterChange("status", e.target.value)}
          >
            <option value="">All Statuses</option>
            {metadata.statuses.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>

          <select
            className="bg-surface-darker border border-border-color rounded-lg px-4 py-2 text-sm outline-none"
            onChange={(e) => handleFilterChange("org", e.target.value)}
          >
            <option value="">All Organizations</option>
            {metadata.orgs.map((o) => (
              <option key={o} value={o}>
                {o}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Grid / Table */}
      <div className="flex-1 overflow-auto p-6">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {vehicles.map((v) => (
              <div
                key={v.id}
                onClick={() => handleVehicleSelect(v.id)}
                className={`bg-surface-darker border rounded-xl p-4 hover:border-primary/50 transition-all cursor-pointer group ${
                  selectedId === v.id ? "border-primary" : "border-border-color"
                }`}
              >
                <div className="flex justify-between items-start mb-4">
                  <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 transition-transform">
                    <span className="material-symbols-outlined">
                      local_shipping
                    </span>
                  </div>
                  <span
                    className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                      v.status === "ACTIVE"
                        ? "bg-emerald-500/10 text-emerald-400"
                        : "bg-yellow-500/10 text-yellow-400"
                    }`}
                  >
                    {v.status}
                  </span>
                </div>
                <h3 className="font-bold text-lg mb-1">
                  {v.name || `Asset ${v.id.substring(0, 8)}`}
                </h3>
                <p className="text-text-muted text-xs mb-4 font-mono">{v.id}</p>

                <div className="space-y-2 border-t border-border-color/50 pt-4">
                  <div className="flex justify-between text-xs">
                    <span className="text-text-muted">Capacity</span>
                    <span className="font-medium">
                      {v.metadata?.capacity || "N/A"} Tons
                    </span>
                  </div>
                  <div className="flex justify-between text-xs">
                    <span className="text-text-muted">Last Update</span>
                    <span className="font-medium">
                      {new Date(v.lastUpdate).toLocaleTimeString()}
                    </span>
                  </div>
                </div>
              </div>
            ))}
            {vehicles.length === 0 && (
              <div className="col-span-full flex flex-col items-center justify-center py-20 text-text-muted">
                <span className="material-symbols-outlined text-6xl opacity-20 mb-4">
                  inventory_2
                </span>
                <p>No assets found in the current registry.</p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
