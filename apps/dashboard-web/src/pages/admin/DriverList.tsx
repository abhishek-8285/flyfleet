import React, { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../services/api";
import { Driver, DriverListResponse } from "../../modules/drivers/types";
import { DriverTable } from "../../components/drivers/DriverTable";
import { Vehicle } from "../../types";

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

      // Fetch drivers
      const res = await api.getDrivers(params as any);
      const data: DriverListResponse = res;
      const rawDrivers = data?.data || [];

      // Fetch vehicles to resolve registration numbers
      // We fetch a larger limit to cover the fleet
      const vehiclesRes = await api.getVehicles({ limit: 1000 });
      const vehicles: Vehicle[] = vehiclesRes?.data || vehiclesRes?.vehicles || vehiclesRes || [];
      const vehicleMap = new Map(vehicles.map(v => [v.id, v.registration_number]));

      // Enrich drivers with vehicle registration
      const enrichedDrivers = rawDrivers.map(d => ({
        ...d,
        current_vehicle: d.current_vehicle_id ? {
          id: d.current_vehicle_id,
          registration_number: vehicleMap.get(d.current_vehicle_id) || d.current_vehicle_id
        } : undefined
      }));

      setDrivers(enrichedDrivers);
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

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to terminate this operational node?")) return;
    try {
      await api.deleteDriver(id);
      fetchDrivers();
    } catch (error) {
      console.error("Failed to delete driver:", error);
      alert("Termination failed. Node has active asset assignments.");
    }
  };

  return (
    <div className="w-full h-full flex flex-col bg-surface-dim -mx-6 -mt-6 relative overflow-hidden">
      {/* Professional Header Section */}
      <div className="w-full bg-surface-container/20 border-b border-outline px-8 py-6 flex flex-col gap-6 z-10 relative">
        <div className="flex justify-between items-end">
          <div>
            <div className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-wider text-on-surface-variant mb-2">
              <span className="material-symbols-outlined text-[10px]">home</span>
              <span>Dashboard</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">Drivers</span>
            </div>
            <h1 className="text-2xl font-bold text-white leading-none tracking-tight">
              Driver Directory
            </h1>
          </div>
          
          <div className="flex items-center gap-3">
             <button className="hidden md:flex items-center gap-2 px-4 py-2.5 bg-surface-container border border-outline text-on-surface-variant text-xs font-semibold uppercase tracking-wide rounded-xl hover:text-white hover:border-primary transition-all">
              <span className="material-symbols-outlined text-[18px]">download</span>
              Export Registry
            </button>
            <button
              onClick={() => navigate("/admin/drivers/new")}
              className="flex items-center gap-2 px-6 py-2.5 bg-primary text-white text-xs font-bold uppercase tracking-wide rounded-xl hover:bg-primary-dark transition-all shadow-lg shadow-primary/20"
            >
              <span className="material-symbols-outlined text-[18px]">person_add</span>
              Add Driver
            </button>
          </div>
        </div>

        {/* Filters & Control Station */}
        <div className="flex flex-wrap items-center gap-4">
          <div className="relative group">
            <span className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-on-surface-variant">
              <span className="material-symbols-outlined text-[20px] group-focus-within:text-primary transition-colors">search</span>
            </span>
            <input
              type="text"
              placeholder="Search by name, license or ID..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="bg-surface-container border border-outline text-white text-sm rounded-xl pl-10 pr-4 py-2.5 w-72 focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder:text-on-surface-variant/40 transition-all font-body"
            />
          </div>

          <div className="flex items-center gap-3">
            <div className="relative">
              <select
                value={filters.status}
                onChange={(e) => setFilters((f) => ({ ...f, status: e.target.value }))}
                className="appearance-none bg-surface-container border border-outline text-white text-[11px] font-bold uppercase tracking-wider rounded-xl pl-4 pr-10 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-surface-container-high transition-colors min-w-[160px]"
              >
                <option value="">Filter Status</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="suspended">Suspended</option>
                <option value="resigned">Resigned</option>
              </select>
              <span className="material-symbols-outlined absolute right-3 top-2.5 pointer-events-none text-on-surface-variant text-sm">filter_list</span>
            </div>
            
            <div className="relative">
              <select
                value={filters.licenseType}
                onChange={(e) => setFilters((f) => ({ ...f, licenseType: e.target.value }))}
                className="appearance-none bg-surface-container border border-outline text-white text-[11px] font-bold uppercase tracking-wider rounded-xl pl-4 pr-10 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-surface-container-high transition-colors min-w-[160px]"
              >
                <option value="">License Type</option>
                <option value="LMV">Light (LMV)</option>
                <option value="HMV">Heavy (HMV)</option>
                <option value="Transport">Commercial</option>
              </select>
              <span className="material-symbols-outlined absolute right-3 top-2.5 pointer-events-none text-on-surface-variant text-sm">workspace_premium</span>
            </div>
          </div>
        </div>
      </div>

      {/* Main Registry Terminal */}
      <div className="flex-1 overflow-hidden p-8 z-10 relative">
        <DriverTable 
          drivers={drivers}
          loading={loading}
          onDriverClick={(d) => navigate(`/admin/drivers/${d.id}`)}
          selectedIds={selectedDrivers}
          onSelect={handleSelect}
          onSelectAll={handleSelectAll}
          onEdit={(id) => navigate(`/admin/drivers/${id}/edit`)}
          onDelete={handleDelete}
        />
      </div>

      {/* Selected Driver Overlay (Bulk Actions) */}
      {selectedDrivers.size > 0 && (
        <div className="fixed bottom-10 left-1/2 -translate-x-1/2 bg-surface-container-highest/95 backdrop-blur-xl border border-primary/40 rounded-2xl px-8 py-4 flex items-center gap-6 shadow-2xl z-[100] animate-bounce-in">
          <div className="flex items-center gap-3">
            <div className="flex h-6 w-6 items-center justify-center rounded-lg bg-primary text-white text-xs font-bold">
              {selectedDrivers.size}
            </div>
            <span className="text-white text-xs font-bold uppercase tracking-wider">Drivers Selected</span>
          </div>
          <div className="h-6 w-px bg-outline"></div>
          <div className="flex items-center gap-4">
            <button className="text-on-surface-variant text-[10px] font-bold uppercase tracking-wide hover:text-white transition-colors">Export CSV</button>
            <button className="text-on-surface-variant text-[10px] font-bold uppercase tracking-wide hover:text-white transition-colors">Update Status</button>
            <button
              onClick={() => setSelectedDrivers(new Set())}
              className="ml-4 bg-white/5 hover:bg-white/10 text-on-surface-variant hover:text-white rounded-lg p-1.5 transition-all flex items-center justify-center"
            >
              <span className="material-symbols-outlined text-[18px]">close</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
