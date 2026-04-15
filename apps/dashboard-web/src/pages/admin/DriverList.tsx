import React, { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../services/api";
import { Driver, DriverListResponse } from "../../modules/drivers/types";

interface PaginationInfo {
  page: number;
  size: number;
  total: number;
  totalPages: number;
}

interface DriverFilterState {
  status: string;
  licenseType: string;
  employmentType: string;
  hasVehicle: boolean;
  search: string;
  orderBy: string;
  orderDir: string;
}

const initialFilters: DriverFilterState = {
  status: "",
  licenseType: "",
  employmentType: "",
  hasVehicle: false,
  search: "",
  orderBy: "created_at",
  orderDir: "DESC",
};

/**
 * DriverList: Enterprise-grade driver list with table view, filters, and pagination.
 * Follows the VehicleList pattern.
 */
export const DriverList: React.FC = () => {
  const navigate = useNavigate();

  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState<PaginationInfo>({
    page: 1,
    size: 20,
    total: 0,
    totalPages: 0,
  });
  const [searchQuery, setSearchQuery] = useState("");
  const [filters, setFilters] = useState<DriverFilterState>(initialFilters);
  const [selectedDrivers, setSelectedDrivers] = useState<Set<string>>(new Set());
  const [showFilterDrawer, setShowFilterDrawer] = useState(false);

  // Fetch drivers
  const fetchDrivers = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string | number> = {
        page: pagination.page,
        limit: pagination.size,
        order_by: filters.orderBy,
        order_dir: filters.orderDir,
      };

      if (searchQuery) params.name = searchQuery;
      if (filters.status) params.status = filters.status;
      if (filters.licenseType) params.license_type = filters.licenseType;
      if (filters.employmentType) params.employment_type = filters.employmentType;

      const res = await api.getDrivers(params as any);
      const data: DriverListResponse = res;
      setDrivers(data?.data || []);
      setPagination((prev) => ({
        ...prev,
        total: data?.total || 0,
        totalPages: Math.ceil((data?.total || 0) / prev.size),
      }));
    } catch (error) {
      console.error("Failed to fetch drivers:", error);
      setDrivers([]);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.size, searchQuery, filters]);

  useEffect(() => {
    fetchDrivers();
  }, [fetchDrivers]);

  const handleSelectAll = () => {
    if (selectedDrivers.size === drivers.length) {
      setSelectedDrivers(new Set());
    } else {
      setSelectedDrivers(new Set(drivers.map((d) => d.id)));
    }
  };

  const handleSelect = (id: string) => {
    const next = new Set(selectedDrivers);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelectedDrivers(next);
  };

  const handlePageChange = (newPage: number) => {
    setPagination((prev) => ({ ...prev, page: Math.max(1, Math.min(newPage, prev.totalPages || 1)) }));
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to delete this driver?")) return;
    try {
      await api.deleteDriver(id);
      fetchDrivers();
    } catch (error) {
      console.error("Failed to delete driver:", error);
      alert("Failed to delete driver. They may have an active assignment or trip.");
    }
  };

  const handleStatusChange = async (id: string, status: string) => {
    try {
      await api.updateDriverStatus(id, status);
      fetchDrivers();
    } catch (error) {
      console.error("Failed to update status:", error);
    }
  };

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      active: "bg-green-500/20 text-green-400",
      inactive: "bg-gray-500/20 text-gray-400",
      suspended: "bg-red-500/20 text-red-400",
      resigned: "bg-yellow-500/20 text-yellow-400",
    };
    return colors[status] || "bg-gray-500/20 text-gray-400";
  };

  const getScoreColor = (score: number) => {
    if (score >= 90) return "text-green-400";
    if (score >= 70) return "text-yellow-400";
    if (score >= 50) return "text-orange-400";
    return "text-red-400";
  };

  return (
    <div className="w-full">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">Drivers</h1>
        <p className="text-text-muted text-sm">Manage driver profiles, licenses, and assignments</p>
      </div>

      {/* Toolbar */}
      <div className="flex flex-col md:flex-row items-start md:items-center gap-3 mb-4">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <span className="material-symbols-outlined absolute left-3 top-2.5 text-text-muted text-sm">search</span>
          <input
            type="text"
            placeholder="Search by name, license, employee code..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-9 pr-4 py-2 focus:ring-1 focus:ring-primary focus:border-primary"
          />
        </div>

        {/* Filter Toggle */}
        <button
          onClick={() => setShowFilterDrawer(!showFilterDrawer)}
          className="flex items-center gap-2 px-3 py-2 bg-surface-dark border border-border-color text-text-muted text-sm font-medium rounded-lg hover:text-white hover:bg-border-color transition-colors"
        >
          <span className="material-symbols-outlined text-[18px]">filter_list</span>
          Filters
        </button>

        <div className="ml-auto flex items-center gap-2">
          <button
            onClick={() => navigate("/admin/drivers/new")}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <span className="material-symbols-outlined text-[16px]">add</span>
            Add Driver
          </button>
        </div>
      </div>

      {/* Filter Drawer */}
      {showFilterDrawer && (
        <div className="bg-surface-dark border border-border-color rounded-lg p-4 mb-4 grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="text-text-muted text-xs mb-1 block">Status</label>
            <select
              value={filters.status}
              onChange={(e) => setFilters((f) => ({ ...f, status: e.target.value }))}
              className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2"
            >
              <option value="">All</option>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
              <option value="suspended">Suspended</option>
              <option value="resigned">Resigned</option>
            </select>
          </div>
          <div>
            <label className="text-text-muted text-xs mb-1 block">License Type</label>
            <select
              value={filters.licenseType}
              onChange={(e) => setFilters((f) => ({ ...f, licenseType: e.target.value }))}
              className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2"
            >
              <option value="">All</option>
              <option value="LMV">LMV</option>
              <option value="HMV">HMV</option>
              <option value="Transport">Transport</option>
            </select>
          </div>
          <div>
            <label className="text-text-muted text-xs mb-1 block">Employment</label>
            <select
              value={filters.employmentType}
              onChange={(e) => setFilters((f) => ({ ...f, employmentType: e.target.value }))}
              className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2"
            >
              <option value="">All</option>
              <option value="FULL_TIME">Full Time</option>
              <option value="CONTRACT">Contract</option>
            </select>
          </div>
          <div className="md:col-span-3 flex justify-end gap-2">
            <button
              onClick={() => setFilters(initialFilters)}
              className="px-3 py-1.5 text-text-muted text-sm hover:text-white transition-colors"
            >
              Clear
            </button>
            <button
              onClick={() => { setShowFilterDrawer(false); fetchDrivers(); }}
              className="px-4 py-1.5 bg-primary text-white text-sm rounded-lg hover:bg-primary-dark transition-colors"
            >
              Apply
            </button>
          </div>
        </div>
      )}

      {/* Table */}
      <div className="bg-surface-dark border border-border-color rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border-color bg-surface-darker">
                <th className="px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    checked={drivers.length > 0 && selectedDrivers.size === drivers.length}
                    onChange={handleSelectAll}
                    className="rounded border-border-color"
                  />
                </th>
                <th className="px-4 py-3 text-left text-text-muted font-medium">Driver</th>
                <th className="px-4 py-3 text-left text-text-muted font-medium">License</th>
                <th className="px-4 py-3 text-left text-text-muted font-medium">Vehicle</th>
                <th className="px-4 py-3 text-left text-text-muted font-medium">Score</th>
                <th className="px-4 py-3 text-left text-text-muted font-medium">Status</th>
                <th className="px-4 py-3 text-left text-text-muted font-medium">Last Active</th>
                <th className="px-4 py-3 text-right text-text-muted font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center text-text-muted">
                    <span className="material-symbols-outlined animate-spin mr-2">progress_activity</span>
                    Loading drivers...
                  </td>
                </tr>
              ) : drivers.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center text-text-muted">
                    No drivers found.{" "}
                    <button
                      onClick={() => navigate("/admin/drivers/new")}
                      className="text-primary hover:underline"
                    >
                      Add your first driver
                    </button>
                  </td>
                </tr>
              ) : (
                drivers.map((driver) => (
                  <tr
                    key={driver.id}
                    className="border-b border-border-color hover:bg-surface-dark/50 transition-colors"
                  >
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedDrivers.has(driver.id)}
                        onChange={() => handleSelect(driver.id)}
                        className="rounded border-border-color"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xs font-semibold">
                          {driver.first_name[0]}{driver.last_name[0]}
                        </div>
                        <div>
                          <div className="text-white font-medium">{driver.first_name} {driver.last_name}</div>
                          <div className="text-text-muted text-xs">{driver.email}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-white">{driver.license_number}</div>
                      <div className="text-text-muted text-xs">{driver.license_type}</div>
                    </td>
                    <td className="px-4 py-3">
                      {driver.current_vehicle_id ? (
                        <span className="text-primary text-xs">Assigned</span>
                      ) : (
                        <span className="text-text-muted text-xs">None</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`font-semibold ${getScoreColor(driver.behavior_score)}`}>
                        {driver.behavior_score}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getStatusBadge(driver.status)}`}>
                        {driver.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-text-muted text-xs">
                      {new Date(driver.updated_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={() => navigate(`/admin/drivers/${driver.id}`)}
                          className="p-1.5 rounded hover:bg-surface-darker text-text-muted hover:text-white transition-colors"
                          title="View"
                        >
                          <span className="material-symbols-outlined text-[18px]">visibility</span>
                        </button>
                        <button
                          onClick={() => navigate(`/admin/drivers/${driver.id}/edit`)}
                          className="p-1.5 rounded hover:bg-surface-darker text-text-muted hover:text-white transition-colors"
                          title="Edit"
                        >
                          <span className="material-symbols-outlined text-[18px]">edit</span>
                        </button>
                        <button
                          onClick={() => handleDelete(driver.id)}
                          className="p-1.5 rounded hover:bg-red-500/20 text-text-muted hover:text-red-400 transition-colors"
                          title="Delete"
                        >
                          <span className="material-symbols-outlined text-[18px]">delete</span>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {pagination.totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-border-color">
            <span className="text-text-muted text-sm">
              {((pagination.page - 1) * pagination.size) + 1}–{Math.min(pagination.page * pagination.size, pagination.total)} of {pagination.total}
            </span>
            <div className="flex items-center gap-1">
              <button
                onClick={() => handlePageChange(pagination.page - 1)}
                disabled={pagination.page <= 1}
                className="p-1.5 rounded hover:bg-surface-darker disabled:opacity-30 disabled:cursor-not-allowed text-text-muted hover:text-white"
              >
                <span className="material-symbols-outlined text-[18px]">chevron_left</span>
              </button>
              {Array.from({ length: Math.min(5, pagination.totalPages) }, (_, i) => {
                const pageNum = i + 1;
                return (
                  <button
                    key={pageNum}
                    onClick={() => handlePageChange(pageNum)}
                    className={`w-8 h-8 rounded text-sm font-medium transition-colors ${
                      pagination.page === pageNum
                        ? "bg-primary text-white"
                        : "text-text-muted hover:bg-surface-darker hover:text-white"
                    }`}
                  >
                    {pageNum}
                  </button>
                );
              })}
              <button
                onClick={() => handlePageChange(pagination.page + 1)}
                disabled={pagination.page >= pagination.totalPages}
                className="p-1.5 rounded hover:bg-surface-darker disabled:opacity-30 disabled:cursor-not-allowed text-text-muted hover:text-white"
              >
                <span className="material-symbols-outlined text-[18px]">chevron_right</span>
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Selected count */}
      {selectedDrivers.size > 0 && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 bg-surface-dark border border-border-color rounded-lg px-4 py-2 flex items-center gap-3 shadow-xl z-50">
          <span className="text-white text-sm font-medium">{selectedDrivers.size} selected</span>
          <button className="text-text-muted text-sm hover:text-white">Export</button>
          <button className="text-text-muted text-sm hover:text-white">Bulk Status</button>
          <button
            onClick={() => setSelectedDrivers(new Set())}
            className="text-text-muted hover:text-white"
          >
            <span className="material-symbols-outlined text-[18px]">close</span>
          </button>
        </div>
      )}
    </div>
  );
};
