import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

export const EditVehicle: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("fleet");
  const { uiStrings: commonStrings } = useUIStrings("common");
  const uiStrings = rawUiStrings || {};
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    license_plate: "",
    vehicle_no: "",
    vin: "",
    make: "",
    model: "",
    year: new Date().getFullYear(),
    fuel_type: "petrol",
    odometer: 0,
    notes: "",
    status: "active",
  });

  useEffect(() => {
    const fetchVehicle = async () => {
      try {
        const res = (await apiClient.get(`/admin/vehicles?id=${id}`)) as any;
        const v = res.data;
        setFormData({
          name: v.name || "",
          license_plate: v.license_plate || "",
          vehicle_no: v.vehicle_no || "",
          vin: v.vin || "",
          make: v.make || "",
          model: v.model || "",
          year: v.year || new Date().getFullYear(),
          fuel_type: v.fuel_type || "petrol",
          odometer: v.odometer || 0,
          notes: v.notes || "",
          status: v.status || "active",
        });
      } catch (error) {
        console.error("Failed to fetch vehicle:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchVehicle();
  }, [id]);

  const handleChange = (
    e: React.ChangeEvent<
      HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
    >,
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]:
        name === "year" || name === "odometer" ? parseInt(value) || 0 : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await apiClient.put(`/admin/vehicles?id=${id}`, formData);
      navigate(`/admin/vehicles/${id}`);
    } catch (error) {
      console.error("Failed to update vehicle:", error);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm("Are you sure you want to delete this vehicle?"))
      return;
    try {
      await apiClient.delete(`/admin/vehicles?id=${id}`);
      navigate("/admin/vehicles");
    } catch (error) {
      console.error("Failed to delete vehicle:", error);
    }
  };

  if (loading)
    return (
      <div className="text-white p-10">
        {uiStrings.msg_loading || "Loading..."}
      </div>
    );

  return (
    <div className="flex-1 flex flex-col min-w-0 overflow-y-auto bg-[#0f111a]">
      <header className="h-20 flex items-center justify-between px-8 border-b border-white/5 bg-[#161b22] sticky top-0 z-10">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate(`/admin/vehicles/${id}`)}
              className="text-slate-400 hover:text-white transition-colors"
            >
              <span className="material-symbols-outlined">arrow_back</span>
            </button>
            <h1 className="text-xl font-bold text-white">
              {uiStrings.title_edit_vehicle || "Edit Vehicle"}
            </h1>
          </div>
          <button
            type="button"
            onClick={handleDelete}
            className="text-red-400 hover:text-red-300 hover:bg-red-500/10 px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">delete</span>
            {uiStrings.btn_delete_vehicle || "Delete Vehicle"}
          </button>
        </header>

        <div className="p-8 max-w-4xl mx-auto w-full">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Section: Identification */}
            <div className="bg-[#1c222b] rounded-xl border border-white/5 p-6">
              <h3 className="text-lg font-medium text-white mb-4 flex items-center gap-2">
                <span className="material-symbols-outlined text-blue-500">
                  badge
                </span>
                {uiStrings.section_identification || "Identification"}
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_internal_name || "Vehicle Name (Internal)"}
                  </label>
                  <input
                    required
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                    placeholder={
                      uiStrings.placeholder_internal_name || "e.g. Truck 001"
                    }
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_license_plate || "License Plate"}
                  </label>
                  <input
                    required
                    name="license_plate"
                    value={formData.license_plate}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none uppercase"
                    placeholder={
                      uiStrings.placeholder_license_plate ||
                      "e.g. KA-01-AB-1234"
                    }
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_vin_chassis || "VIN (Chassis No)"}
                  </label>
                  <input
                    name="vin"
                    value={formData.vin}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none uppercase"
                    placeholder={
                      uiStrings.placeholder_vin || "17-character VIN"
                    }
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_fleet_id || "Vehicle No / Fleet ID"}
                  </label>
                  <input
                    name="vehicle_no"
                    value={formData.vehicle_no}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                    placeholder={
                      uiStrings.placeholder_fleet_id || "Vehicle Number"
                    }
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.th_status || "Status"}
                  </label>
                  <select
                    name="status"
                    value={formData.status}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                  >
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
              </div>
            </div>

            {/* Section: Specs */}
            <div className="bg-[#1c222b] rounded-xl border border-white/5 p-6">
              <h3 className="text-lg font-medium text-white mb-4 flex items-center gap-2">
                <span className="material-symbols-outlined text-purple-500">
                  directions_car
                </span>
                {uiStrings.section_specifications || "Specifications"}
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_make || "Make"}
                  </label>
                  <input
                    name="make"
                    value={formData.make}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                    placeholder={uiStrings.placeholder_make || "e.g. Tata"}
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_model || "Model"}
                  </label>
                  <input
                    name="model"
                    value={formData.model}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                    placeholder={uiStrings.placeholder_model || "e.g. Ace"}
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_year || "Year"}
                  </label>
                  <input
                    type="number"
                    name="year"
                    value={formData.year}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_fuel_type || "Fuel Type"}
                  </label>
                  <select
                    name="fuel_type"
                    value={formData.fuel_type}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                  >
                    <option value="diesel">Diesel</option>
                    <option value="petrol">Petrol</option>
                    <option value="electric">Electric</option>
                    <option value="cng">CNG</option>
                    <option value="hybrid">Hybrid</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5">
                    {uiStrings.label_current_odometer ||
                      "Current Odometer (km)"}
                  </label>
                  <input
                    type="number"
                    name="odometer"
                    value={formData.odometer}
                    onChange={handleChange}
                    className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                  />
                </div>
              </div>
            </div>

            {/* Section: Notes */}
            <div className="bg-[#1c222b] rounded-xl border border-white/5 p-6">
              <h3 className="text-lg font-medium text-white mb-4 flex items-center gap-2">
                <span className="material-symbols-outlined text-slate-500">
                  notes
                </span>
                {uiStrings.section_notes || "Notes"}
              </h3>
              <textarea
                name="notes"
                value={formData.notes}
                onChange={handleChange}
                rows={3}
                className="w-full bg-[#161b22] border border-white/10 rounded-lg px-4 py-2.5 text-white focus:ring-2 focus:ring-blue-500 outline-none"
                placeholder={
                  uiStrings.placeholder_notes || "Additional details..."
                }
              />
            </div>

            <div className="flex justify-end gap-4 pt-4">
              <button
                type="button"
                onClick={() => navigate(`/admin/vehicles/${id}`)}
                className="px-6 py-2.5 text-slate-300 hover:text-white font-medium transition-colors"
              >
                {commonStrings?.btn_cancel || uiStrings.btn_cancel || "Cancel"}
              </button>
              <button
                type="submit"
                disabled={saving}
                className="bg-blue-600 hover:bg-blue-500 text-white px-8 py-2.5 rounded-lg font-medium shadow-lg shadow-blue-500/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
              >
                {saving ? (
                  <span className="animate-spin material-symbols-outlined text-sm">
                    progress_activity
                  </span>
                ) : null}
                {uiStrings.btn_save_changes || "Save Changes"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};
