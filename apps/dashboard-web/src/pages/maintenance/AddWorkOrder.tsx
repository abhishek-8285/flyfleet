import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

interface CreateWorkOrderRequest {
  vehicle_id: string;
  title: string;
  description: string;
  maintenance_type: string;
  priority: string;
  assigned_to?: string;
  scheduled_date?: string;
  estimated_cost: number;
}

interface Vehicle {
  id: string;
  make: string;
  model: string;
  license_plate: string;
}

const AddWorkOrder: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("maintenance");
  const { uiStrings: commonStrings } = useUIStrings("common");
  const uiStrings = rawUiStrings || {};

  const [loading, setLoading] = useState(false);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [formData, setFormData] = useState<CreateWorkOrderRequest>({
    vehicle_id: "",
    title: "",
    description: "",
    maintenance_type: "corrective",
    priority: "MEDIUM",
    estimated_cost: 0,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    // Fetch vehicles for dropdown
    const fetchVehicles = async () => {
      try {
        const data = (await apiClient.get("/admin/vehicles", {
          params: { limit: 100 },
        })) as any;
        setVehicles(data?.vehicles || []);
      } catch (error) {
        console.error("Failed to fetch vehicles:", error);
      }
    };
    fetchVehicles();
  }, []);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!formData.vehicle_id)
      newErrors.vehicle_id = uiStrings.err_vehicle_req || "Vehicle is required";
    if (!formData.title.trim())
      newErrors.title = uiStrings.err_title_req || "Title is required";
    if (!formData.description.trim())
      newErrors.description =
        uiStrings.err_desc_req || "Description is required";
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    setLoading(true);
    try {
      const payload = (await apiClient.post(
        "/admin/work-orders",
        formData,
      )) as any;
      navigate(`/maintenance/work-orders/${payload?.id}`);
    } catch (error: any) {
      setErrors({
        submit:
          error?.message ||
          uiStrings.err_create_wo ||
          "Failed to create work order",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate("/maintenance/work-orders")}
            className="p-2 hover:bg-surface-dark rounded-lg transition-colors"
          >
            <span className="material-symbols-outlined text-text-muted">
              arrow_back
            </span>
          </button>
          <div>
            <h1 className="text-2xl font-bold text-white">
              {uiStrings.title_create_wo || "Create Work Order"}
            </h1>
            <p className="text-sm text-text-muted mt-1">
              {uiStrings.sub_create_wo || "Add a new maintenance work order"}
            </p>
          </div>
        </div>
      </div>

      {/* Form */}
      <div className="flex-1 overflow-y-auto p-6">
        <form onSubmit={handleSubmit} className="max-w-2xl mx-auto space-y-6">
          {errors.submit && (
            <div className="bg-red-500/20 border border-red-500/30 text-red-400 px-4 py-3 rounded-lg text-sm">
              {errors.submit}
            </div>
          )}

          <div className="bg-surface-dark border border-border-color rounded-xl p-6 space-y-5">
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">
                assignment
              </span>
              {uiStrings.label_wo_details || "Work Order Details"}
            </h2>

            {/* Vehicle */}
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                {uiStrings.label_vehicle_req || "Vehicle *"}
              </label>
              <select
                value={formData.vehicle_id}
                onChange={(e) =>
                  setFormData({ ...formData, vehicle_id: e.target.value })
                }
                className={`w-full bg-surface-darker border ${errors.vehicle_id ? "border-red-500" : "border-border-color"} text-white rounded-lg px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary`}
              >
                <option value="">
                  {uiStrings.placeholder_select_vehicle || "Select a vehicle"}
                </option>
                {vehicles.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.make} {v.model} - {v.license_plate}
                  </option>
                ))}
              </select>
              {errors.vehicle_id && (
                <p className="text-xs text-red-400 mt-1">{errors.vehicle_id}</p>
              )}
            </div>

            {/* Title */}
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                {uiStrings.label_title_req || "Title *"}
              </label>
              <input
                type="text"
                placeholder={
                  uiStrings.placeholder_wo_title ||
                  "e.g., Oil Change, Brake Inspection"
                }
                value={formData.title}
                onChange={(e) =>
                  setFormData({ ...formData, title: e.target.value })
                }
                className={`w-full bg-surface-darker border ${errors.title ? "border-red-500" : "border-border-color"} text-white rounded-lg px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary`}
              />
              {errors.title && (
                <p className="text-xs text-red-400 mt-1">{errors.title}</p>
              )}
            </div>

            {/* Description */}
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                {uiStrings.label_desc_req || "Description *"}
              </label>
              <textarea
                rows={4}
                placeholder={
                  uiStrings.placeholder_wo_desc ||
                  "Describe the work to be performed..."
                }
                value={formData.description}
                onChange={(e) =>
                  setFormData({ ...formData, description: e.target.value })
                }
                className={`w-full bg-surface-darker border ${errors.description ? "border-red-500" : "border-border-color"} text-white rounded-lg px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary resize-none`}
              />
              {errors.description && (
                <p className="text-xs text-red-400 mt-1">
                  {errors.description}
                </p>
              )}
            </div>

            {/* Type & Priority */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_maintenance_type || "Type"}
                </label>
                <select
                  value={formData.maintenance_type}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      maintenance_type: e.target.value,
                    })
                  }
                  className="w-full bg-surface-darker border border-border-color text-white rounded-lg px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary"
                >
                  <option value="preventive">
                    {uiStrings.option_type_preventive || "Preventive"}
                  </option>
                  <option value="corrective">
                    {uiStrings.option_type_corrective || "Corrective"}
                  </option>
                  <option value="inspection">
                    {uiStrings.option_type_inspection || "Inspection"}
                  </option>
                  <option value="breakdown">
                    {uiStrings.option_type_breakdown || "Breakdown"}
                  </option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {commonStrings?.label_priority ||
                    uiStrings?.th_priority ||
                    "Priority"}
                </label>
                <select
                  value={formData.priority}
                  onChange={(e) =>
                    setFormData({ ...formData, priority: e.target.value })
                  }
                  className="w-full bg-surface-darker border border-border-color text-white rounded-lg px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary"
                >
                  <option value="LOW">
                    {commonStrings?.option_priority_low || "Low"}
                  </option>
                  <option value="MEDIUM">
                    {commonStrings?.option_priority_medium || "Medium"}
                  </option>
                  <option value="HIGH">
                    {commonStrings?.option_priority_high || "High"}
                  </option>
                  <option value="CRITICAL">
                    {commonStrings?.option_priority_critical || "Critical"}
                  </option>
                </select>
              </div>
            </div>

            {/* Scheduled Date & Cost */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_scheduled_date || "Scheduled Date"}
                </label>
                <input
                  type="date"
                  value={formData.scheduled_date || ""}
                  onChange={(e) =>
                    setFormData({ ...formData, scheduled_date: e.target.value })
                  }
                  className="w-full bg-surface-darker border border-border-color text-white rounded-lg px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_estimated_cost_usd || "Estimated Cost ($)"}
                </label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  placeholder="0.00"
                  value={formData.estimated_cost || ""}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      estimated_cost: parseFloat(e.target.value) || 0,
                    })
                  }
                  className="w-full bg-surface-darker border border-border-color text-white rounded-lg px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary"
                />
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-4">
            <button
              type="button"
              onClick={() => navigate("/maintenance/vendors")}
              className="px-4 py-2 mr-3 text-sm font-medium text-text-muted hover:text-white transition-colors"
            >
              {uiStrings.btn_cancel || commonStrings?.cancel || "Cancel"}
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-6 py-2.5 text-sm font-medium bg-primary hover:bg-primary/90 text-white rounded-lg shadow-[0_0_15px_rgba(19,127,236,0.3)] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {loading && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              )}
              {uiStrings.btn_create_wo || "Create Work Order"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default AddWorkOrder;
