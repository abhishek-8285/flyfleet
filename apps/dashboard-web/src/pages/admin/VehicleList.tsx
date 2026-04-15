import React, { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import apiClient from "../../services/apiClient";
import { VehicleFilters, VehicleFilterState } from "../../components/vehicles/VehicleFilters";

// Types
interface Vehicle {
  id: string;
  org_id: string;
  registration_number: string;
  type: string;
  make: string;
  model: string;
  year: number;
  status: "active" | "inactive" | "maintenance";
  current_driver?: {
    id: string;
    name: string;
  };
  odometer?: number;
  fuel_type?: string;
  metadata?: Record<string, any>;
}

interface PaginationInfo {
  page: number;
  size: number;
  total: number;
  totalPages: number;
}

/**
 * VehicleList: Enterprise-grade vehicle list with table view, filter bar, and pagination.
 * Uses TripFilters-style horizontal filter bar with advanced drawer.
 */
export const VehicleList: React.FC = () => {
  const navigate = useNavigate();

  // State
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState<PaginationInfo>({
    page: 1,
    size: 20,
    total: 0,
    totalPages: 0,
  });
  const [searchQuery, setSearchQuery] = useState("");
  const [filters, setFilters] = useState<VehicleFilterState>({
    status: "",
    vehicleType: "",
    fuelType: "",
    driverId: "",
    organizationId: "",
    make: "",
    search: "",
    orderBy: "created_at",
    orderDir: "DESC",
  });
  const [selectedVehicles, setSelectedVehicles] = useState<Set<string>>(
    new Set()
  );

  // Fetch vehicles
  const fetchVehicles = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: pagination.page.toString(),
        size: pagination.size.toString(),
        order_by: filters.orderBy,
        order_dir: filters.orderDir,
      });

      if (searchQuery) params.append("search", searchQuery);
      if (filters.status) params.append("status", filters.status);
      if (filters.vehicleType) params.append("type", filters.vehicleType);
      if (filters.fuelType) params.append("fuel_type", filters.fuelType);
      if (filters.driverId) params.append("driver_id", filters.driverId);
      if (filters.organizationId) params.append("org_id", filters.organizationId);
      if (filters.make) params.append("make", filters.make);

      const res = await apiClient.get(`/admin/vehicles?${params.toString()}`);
      const data = res.data;
      setVehicles(data?.vehicles || []);
      setPagination((prev) => ({
        ...prev,
        total: data?.total || data?.vehicles?.length || 0,
        totalPages: data?.total_pages || Math.ceil((data?.total || data?.vehicles?.length || 0) / prev.size),
      }));
    } catch (error) {
      console.error("Failed to fetch vehicles:", error);
      setVehicles([]);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.size, searchQuery, filters]);

  useEffect(() => {
    fetchVehicles();
  }, [fetchVehicles]);

  // Handlers
  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleFilterChange = (newFilters: VehicleFilterState) => {
    setFilters(newFilters);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleResetFilters = () => {
    setFilters({
      status: "",
      vehicleType: "",
      fuelType: "",
      driverId: "",
      organizationId: "",
      make: "",
      search: "",
      orderBy: "created_at",
      orderDir: "DESC",
    });
    setSearchQuery("");
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const toggleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedVehicles(new Set(vehicles.map((v) => v.id)));
    } else {
      setSelectedVehicles(new Set());
    }
  };

  const toggleSelectVehicle = (id: string, checked: boolean) => {
    setSelectedVehicles((prev) => {
      const next = new Set(prev);
      if (checked) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  };

  const getStatusBadge = (status: string) => {
    const config = {
      active: { color: "success", label: "ACTIVE" },
      inactive: { color: "danger", label: "INACTIVE" },
      maintenance: { color: "warning", label: "MAINTENANCE" },
    } as const;

    const { color, label } = config[status as keyof typeof config] || {
      color: "text-muted",
      label: status.toUpperCase(),
    };

    const colorClasses = {
      success: "bg-success/10 text-success border-success/20",
      warning: "bg-warning/10 text-warning border-warning/20",
      danger: "bg-danger/10 text-danger border-danger/20",
      "text-muted": "bg-border-color/20 text-text-muted border-border-color/30",
    } as const;

    const dotClasses = {
      success: "bg-success",
      warning: "bg-warning",
      danger: "bg-danger",
      "text-muted": "bg-text-muted",
    } as const;

    return (
      <span
        className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold border ${colorClasses[color as keyof typeof colorClasses]}`}
      >
        <span className={`w-1.5 h-1.5 rounded-full ${dotClasses[color as keyof typeof dotClasses]}`} />
        {label}
      </span>
    );
  };

  const getVehicleIcon = (type: string | undefined) => {
    if (!type) return "directions_car";
    const iconMap: Record<string, string> = {
      truck: "local_shipping",
      van: "airport_shuttle",
      sedan: "directions_car",
      suv: "directions_car",
      heavy_haul: "local_shipping",
    };
    return iconMap[type.toLowerCase()] || "directions_car";
  };

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .substring(0, 2);
  };

  const avatarColors = [
    "bg-primary/20 text-primary",
    "bg-purple-500/20 text-purple-500",
    "bg-orange-500/20 text-orange-500",
    "bg-blue-500/20 text-blue-500",
    "bg-pink-500/20 text-pink-500",
    "bg-green-500/20 text-green-500",
  ];

  return (
    <div className="flex flex-col h-full w-full bg-background-dark">
      {/* Filter Bar */}
      <VehicleFilters
        filters={filters}
        onFilterChange={handleFilterChange}
        onReset={handleResetFilters}
        totalCount={pagination.total}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0 bg-background-dark">
        {/* Table Header */}
        <div className="px-6 py-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
          <div className="flex items-center gap-3">
            <span className="text-white text-sm font-semibold">All Vehicles</span>
            <span className="px-2 py-0.5 rounded-full bg-surface-dark border border-border-color text-xs text-text-muted">
              Page {pagination.page} of {pagination.totalPages || 1}
            </span>
          </div>
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate("/admin/vehicles/new")}
              className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
            >
              <span className="material-symbols-outlined text-[16px]">add</span>
              Add Vehicle
            </button>
            <div className="h-4 w-px bg-border-color" />
            <div className="flex items-center gap-2">
              <button
                onClick={fetchVehicles}
                className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white transition-colors"
                title="Refresh"
              >
                <span className="material-symbols-outlined text-[18px]">
                  refresh
                </span>
              </button>
              <button
                className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white transition-colors"
                title="Export"
              >
                <span className="material-symbols-outlined text-[18px]">
                  download
                </span>
              </button>
              <div className="h-4 w-px bg-border-color mx-1" />
              <button
                className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white transition-colors"
                title="Columns"
              >
                <span className="material-symbols-outlined text-[18px]">
                  view_column
                </span>
              </button>
            </div>
          </div>
        </div>

        {/* Table */}
        <div className="flex-1 overflow-auto custom-scrollbar">
          {loading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : (
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-surface-darker">
                <tr className="border-b border-border-color text-xs text-text-muted uppercase tracking-wider">
                  <th className="px-6 py-3 font-semibold w-12">
                    <input
                      type="checkbox"
                      checked={
                        vehicles.length > 0 &&
                        selectedVehicles.size === vehicles.length
                      }
                      onChange={(e) => toggleSelectAll(e.target.checked)}
                      className="rounded border-border-color bg-surface-dark text-primary focus:ring-offset-surface-darker focus:ring-primary w-3.5 h-3.5"
                    />
                  </th>
                  <th className="px-6 py-3 font-semibold">Registration No / ID</th>
                  <th className="px-6 py-3 font-semibold">Make / Model</th>
                  <th className="px-6 py-3 font-semibold">Assigned Driver</th>
                  <th className="px-6 py-3 font-semibold">Odometer</th>
                  <th className="px-6 py-3 font-semibold text-right">Status</th>
                  <th className="px-6 py-3 font-semibold w-10" />
                </tr>
              </thead>
              <tbody className="text-sm divide-y divide-border-color bg-surface-dark">
                {vehicles.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="px-6 py-20 text-center text-text-muted">
                      <div className="flex flex-col items-center">
                        <span className="material-symbols-outlined text-6xl opacity-20 mb-4">
                          inventory_2
                        </span>
                        <p className="text-lg font-medium">No vehicles found</p>
                        <p className="text-sm mt-1">
                          Try adjusting your filters or add a new vehicle
                        </p>
                      </div>
                    </td>
                  </tr>
                ) : (
                  vehicles.map((vehicle) => (
                    <tr
                      key={vehicle.id}
                      className="hover:bg-surface-darker/50 transition-colors group"
                    >
                      <td className="px-6 py-3">
                        <input
                          type="checkbox"
                          checked={selectedVehicles.has(vehicle.id)}
                          onChange={(e) =>
                            toggleSelectVehicle(vehicle.id, e.target.checked)
                          }
                          className="rounded border-border-color bg-surface-darker text-primary focus:ring-offset-surface-darker focus:ring-primary w-3.5 h-3.5"
                        />
                      </td>
                      <td className="px-6 py-3">
                        <div>
                          <div className="text-white font-medium font-mono text-sm">
                            {vehicle.registration_number || "N/A"}
                          </div>
                          <div className="text-[10px] text-text-muted font-mono">
                            #{vehicle.id.slice(-6).toUpperCase()}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-3 text-white">
                        <div className="flex items-center gap-2">
                          <span className="material-symbols-outlined text-text-muted text-lg">
                            {getVehicleIcon(vehicle.type)}
                          </span>
                          <div>
                            <div>
                              {vehicle.make} {vehicle.model}
                            </div>
                            <div className="text-[10px] text-text-muted">
                              {vehicle.year} • {vehicle.metadata?.fuelType || vehicle.fuel_type || "Unknown"}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-3">
                        {vehicle.current_driver ? (
                          <div className="flex items-center gap-2">
                            <div
                              className={`w-6 h-6 rounded-full ${
                                avatarColors[
                                  vehicle.current_driver.name.charCodeAt(0) %
                                    avatarColors.length
                                ]
                              } flex items-center justify-center text-[10px] font-bold`}
                            >
                              {getInitials(vehicle.current_driver.name)}
                            </div>
                            <span className="text-text-muted">
                              {vehicle.current_driver.name}
                            </span>
                          </div>
                        ) : (
                          <span className="text-text-muted italic text-xs">
                            Unassigned
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-3 text-text-muted font-mono">
                        {vehicle.odometer
                          ? `${vehicle.odometer.toLocaleString()} km`
                          : "—"}
                      </td>
                      <td className="px-6 py-3 text-right">
                        {getStatusBadge(vehicle.status)}
                      </td>
                      <td className="px-6 py-3 text-right">
                        <button
                          onClick={() =>
                            navigate(`/admin/vehicles/${vehicle.id}`)
                          }
                          className="text-text-muted hover:text-white opacity-0 group-hover:opacity-100 transition-opacity"
                          title="View Details"
                        >
                          <span className="material-symbols-outlined text-[18px]">
                            more_vert
                          </span>
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        <div className="px-6 py-3 border-t border-border-color bg-surface-darker flex items-center justify-between">
          <span className="text-xs text-text-muted">
            Showing {(pagination.page - 1) * pagination.size + 1} to{" "}
            {Math.min(pagination.page * pagination.size, pagination.total)} of{" "}
            {pagination.total} entries
          </span>
          <div className="flex items-center gap-2">
            <button
              onClick={() =>
                setPagination((prev) => ({ ...prev, page: prev.page - 1 }))
              }
              disabled={pagination.page === 1}
              className="p-1 rounded hover:bg-surface-dark text-text-muted disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span className="material-symbols-outlined text-[18px]">
                chevron_left
              </span>
            </button>
            {Array.from({ length: Math.min(5, pagination.totalPages) }, (_, i) => {
              const pageNum = Math.max(1, pagination.page - 2) + i;
              if (pageNum > pagination.totalPages) return null;
              return (
                <button
                  key={pageNum}
                  onClick={() =>
                    setPagination((prev) => ({ ...prev, page: pageNum }))
                  }
                  className={`px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                    pagination.page === pageNum
                      ? "bg-primary text-white"
                      : "hover:bg-surface-dark text-text-muted"
                  }`}
                >
                  {pageNum}
                </button>
              );
            })}
            {pagination.totalPages > 5 && (
              <>
                <span className="text-text-muted text-xs">...</span>
                <button
                  onClick={() =>
                    setPagination((prev) => ({
                      ...prev,
                      page: pagination.totalPages,
                    }))
                  }
                  className="px-2.5 py-1 rounded hover:bg-surface-dark text-text-muted text-xs font-medium"
                >
                  {pagination.totalPages}
                </button>
              </>
            )}
            <button
              onClick={() =>
                setPagination((prev) => ({ ...prev, page: prev.page + 1 }))
              }
              disabled={pagination.page === pagination.totalPages}
              className="p-1 rounded hover:bg-surface-dark text-text-muted disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span className="material-symbols-outlined text-[18px]">
                chevron_right
              </span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VehicleList;
