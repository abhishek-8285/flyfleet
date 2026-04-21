import React, { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { 
  Plus, 
  Search, 
  Download, 
  Filter, 
  ChevronRight,
  MoreVertical,
  User 
} from "lucide-react";
import { api } from "../../services/api";
import { Driver, DriverListResponse } from "../../modules/drivers/types";
import { DriverTable } from "../../components/drivers/DriverTable";

export const DriverList: React.FC = () => {
  const navigate = useNavigate();
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedDrivers, setSelectedDrivers] = useState<Set<string>>(new Set());

  const fetchDrivers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getDrivers({ limit: 100 });
      const data: DriverListResponse = res;
      setDrivers(data?.data || []);
    } catch (error) {
      console.error("Failed to fetch drivers:", error);
      setDrivers([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDrivers();
  }, [fetchDrivers]);

  const handleSelect = (id: string) => {
    const next = new Set(selectedDrivers);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelectedDrivers(next);
  };

  const handleSelectAll = () => {
    if (selectedDrivers.size === drivers.length) {
      setSelectedDrivers(new Set());
    } else {
      setSelectedDrivers(new Set(drivers.map((d) => d.id)));
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to delete this driver?")) return;
    try {
      await api.deleteDriver(id);
      fetchDrivers();
    } catch (error) {
      console.error("Failed to delete driver:", error);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Personnel Management</h1>
          <p className="text-gray-400 text-sm">Manage and track your active personnel</p>
        </div>
        <div className="flex gap-4">
            <button className="flex items-center gap-2 px-4 py-2 bg-white/5 border border-white/10 text-white rounded-lg hover:bg-white/10 transition-all font-semibold">
                <Download className="w-4 h-4" />
                Export
            </button>
            <button
              onClick={() => navigate("/admin/drivers/new")}
              className="flex items-center gap-2 px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all font-semibold shadow-lg shadow-primary/20"
            >
              <Plus className="w-5 h-5" />
              Add Driver
            </button>
        </div>
      </div>

      <div className="bg-surface-dark border border-border-color rounded-2xl overflow-hidden shadow-2xl">
        <div className="p-4 border-b border-border-color flex items-center justify-between gap-4">
            <div className="relative flex-1 max-w-md">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                    type="text"
                    placeholder="Search by name, license or code..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-full bg-white/5 border border-white/10 rounded-xl pl-12 pr-4 py-2 text-white focus:outline-none focus:border-primary transition-all"
                />
            </div>
            <button className="p-2 text-gray-400 hover:text-white transition-colors">
                <Filter className="w-5 h-5" />
            </button>
        </div>

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
    </div>
  );
};

export default DriverList;
