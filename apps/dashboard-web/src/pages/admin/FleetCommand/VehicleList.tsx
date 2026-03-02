import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import { api } from "../../../services/api";
import { Vehicle } from "../../../types";

export const VehicleList: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("fleet_vehicles");
  const uiStrings = rawUiStrings || {};
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [filterStatus, setFilterStatus] = useState<string>("all");

  useEffect(() => {
    fetchVehicles();
  }, []);

  const fetchVehicles = async () => {
    setLoading(true);
    try {
      const data = await api.getVehicles({ page: 1, size: 100 });
      // The API returns an object with 'vehicles' array
      setVehicles(data.vehicles || []);
    } catch (error) {
      console.error("Failed to fetch vehicles:", error);
      // Fallback to empty if error
      setVehicles([]);
    } finally {
      setLoading(false);
    }
  };

  const filteredVehicles = vehicles.filter((v) => {
    const regNum =
      (v as any).registration_number || (v as any).license_plate || "";
    const name = (v as any).name || (v as any).make + " " + (v as any).model;
    const vin = (v as any).vin || v.metadata?.vin || "";

    const matchesSearch =
      name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      regNum.toLowerCase().includes(searchTerm.toLowerCase()) ||
      vin.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesStatus = filterStatus === "all" || v.status === filterStatus;

    return matchesSearch && matchesStatus;
  });

  return (
    <div className="flex flex-col gap-6">
      {/* Page Title & Actions Area */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">
            {uiStrings.fleet_command_title || "Fleet Command"}
          </h1>
          <p className="text-text-muted text-sm mt-1 font-medium">
            {uiStrings.fleet_command_sub ||
              "Real-time asset visibility and control."}
          </p>
        </div>
        <button
          onClick={() => navigate("/admin/vehicles/new")}
          className="flex items-center gap-2 bg-primary hover:bg-blue-600 text-white px-5 py-2.5 rounded-lg font-medium transition-all shadow-lg shadow-primary/20"
        >
          <span className="material-symbols-outlined text-[20px]">add</span>
          {uiStrings.btn_add_vehicle || "Add Vehicle"}
        </button>
      </div>

      {/* Filters & Search */}
      <div className="bg-surface-dark p-4 rounded-xl border border-border-color shadow-sm">
        <div className="flex flex-wrap justify-between items-center gap-4">
          <div className="flex items-center gap-4 flex-1 min-w-[300px]">
            <div className="relative flex-1 max-w-md">
              <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">
                search
              </span>
              <input
                type="text"
                placeholder={
                  uiStrings.search_placeholder || "Search VIN, Plate, Name..."
                }
                className="w-full bg-surface-darker border border-border-color text-white pl-10 pr-4 py-2 rounded-lg focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-text-muted/50 text-sm"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
            <div className="h-8 w-px bg-border-color mx-2"></div>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="bg-surface-darker border border-border-color text-text-muted px-4 py-2 rounded-lg focus:ring-1 focus:ring-primary outline-none text-sm cursor-pointer hover:bg-border-color/30 transition-colors"
            >
              <option value="all">
                {uiStrings.filter_all_status || "All Status"}
              </option>
              <option value="active">
                {uiStrings.filter_active || "Active"}
              </option>
              <option value="maintenance">
                {uiStrings.filter_maintenance || "Maintenance"}
              </option>
              <option value="inactive">
                {uiStrings.filter_inactive || "Inactive"}
              </option>
            </select>
          </div>
          <div className="flex items-center gap-3 text-xs text-text-muted font-medium bg-surface-darker px-3 py-1.5 rounded-full border border-border-color">
            <span className="material-symbols-outlined text-[14px]">dns</span>
            <span>
              {(
                uiStrings.managed_assets_label || "{count} Managed Assets"
              ).replace("{count}", String(filteredVehicles.length))}
            </span>
          </div>
        </div>
      </div>

      {/* Table Area */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden shadow-sm">
        {loading ? (
          <div className="flex flex-col items-center justify-center h-64 text-text-muted gap-4">
            <span className="animate-spin material-symbols-outlined text-3xl text-primary">
              sync
            </span>
            <p className="text-sm font-medium">
              {uiStrings.syncing_data || "Synchronizing Fleet Data..."}
            </p>
          </div>
        ) : (
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-surface-darker border-b border-border-color text-text-muted text-[10px] uppercase font-bold tracking-wider">
                <th className="p-4 pl-6">
                  {uiStrings.th_vehicle_asset || "Vehicle Asset"}
                </th>
                <th className="p-4">
                  {uiStrings.th_vin_id || "VIN / Identification"}
                </th>
                <th className="p-4 text-center">
                  {uiStrings.th_odometer || "Odometer"}
                </th>
                <th className="p-4">
                  {uiStrings.th_assigned_driver || "Assigned Driver"}
                </th>
                <th className="p-4">{uiStrings.th_status || "Status"}</th>
                <th className="p-4 text-right pr-6">
                  {uiStrings.th_management || "Management"}
                </th>
              </tr>
            </thead>
            <tbody className="text-sm border-0">
              {filteredVehicles.map((vehicle) => (
                <tr
                  key={vehicle.id}
                  className="border-b border-border-color/50 hover:bg-white/[0.02] transition-colors group cursor-pointer"
                  onClick={() => navigate(`/admin/vehicles/${vehicle.id}`)}
                >
                  <td className="p-4 pl-6">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
                        <span className="material-symbols-outlined">
                          local_shipping
                        </span>
                      </div>
                      <div>
                        <div className="font-bold text-white tracking-tight">
                          {vehicle.make} {vehicle.model}
                        </div>
                        <div className="text-[11px] text-text-muted font-mono">
                          {vehicle.registration_number ||
                            (vehicle as any).license_plate}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="p-4">
                    <div className="text-white text-xs font-mono tracking-tighter opacity-80">
                      {vehicle.metadata?.vin || (vehicle as any).vin || "N/A"}
                    </div>
                    <div className="text-[10px] text-text-muted mt-0.5 uppercase tracking-widest">
                      {(
                        uiStrings.production_year_label ||
                        "{year} Production Year"
                      ).replace("{year}", String(vehicle.year))}
                    </div>
                  </td>
                  <td className="p-4 text-center">
                    <div className="text-white font-bold">
                      {vehicle.odometer?.toLocaleString() || "0"}
                    </div>
                    <div className="text-[10px] text-text-muted uppercase">
                      {uiStrings.miles_label || "Miles"}
                    </div>
                  </td>
                  <td className="p-4">
                    {vehicle.current_driver ? (
                      <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-full bg-border-color flex items-center justify-center">
                          <span className="material-symbols-outlined text-[14px] text-text-muted">
                            person
                          </span>
                        </div>
                        <span className="text-white text-xs font-medium">
                          {vehicle.current_driver.name}
                        </span>
                      </div>
                    ) : (
                      <span className="text-text-muted text-xs italic">
                        {uiStrings.unassigned || "Unassigned"}
                      </span>
                    )}
                  </td>
                  <td className="p-4">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider border ${
                        vehicle.status === "active"
                          ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                          : vehicle.status === "maintenance"
                            ? "bg-yellow-500/10 text-yellow-400 border-yellow-500/20"
                            : "bg-text-muted/10 text-text-muted border-border-color"
                      }`}
                    >
                      {vehicle.status}
                    </span>
                  </td>
                  <td className="p-4 text-right pr-6">
                    <div className="flex justify-end gap-2">
                      <button
                        className="text-text-muted hover:text-white p-1.5 hover:bg-white/5 rounded transition-colors"
                        onClick={(e) => {
                          e.stopPropagation();
                          navigate(`/admin/vehicles/${vehicle.id}/edit`);
                        }}
                      >
                        <span className="material-symbols-outlined text-[18px]">
                          edit_square
                        </span>
                      </button>
                      <button className="text-text-muted hover:text-red-400 p-1.5 hover:bg-red-500/5 rounded transition-colors">
                        <span className="material-symbols-outlined text-[18px]">
                          delete
                        </span>
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {filteredVehicles.length === 0 && (
                <tr>
                  <td colSpan={6} className="p-20 text-center">
                    <div className="flex flex-col items-center gap-3">
                      <div className="size-16 rounded-full bg-surface-darker border border-border-color flex items-center justify-center mb-2">
                        <span className="material-symbols-outlined text-4xl text-text-muted/20">
                          search_off
                        </span>
                      </div>
                      <p className="text-white font-medium">
                        {uiStrings.no_results_title ||
                          "No assets matches search"}
                      </p>
                      <p className="text-sm text-text-muted">
                        {uiStrings.no_results_sub ||
                          "Try adjusting your filters or search keywords."}
                      </p>
                      <button
                        onClick={() => {
                          setSearchTerm("");
                          setFilterStatus("all");
                        }}
                        className="mt-2 text-primary text-sm font-medium hover:underline"
                      >
                        {uiStrings.btn_clear_filters || "Clear all filters"}
                      </button>
                    </div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
