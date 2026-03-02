import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../hooks/useUIStrings";
import FuelService from "../services/FuelService";
import { CreateFuelLogRequest } from "../types/fuel";

const AddFuelLog: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("fuel_log");
  const uiStrings = rawUiStrings || {};
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState<CreateFuelLogRequest>({
    vehicle_id: "",
    driver_id: "",
    station_name: "",
    fuel_type: "Diesel",
    quantity: 0,
    unit_price: 0,
    total_cost: 0,
    odometer: 0,
    transaction_date: new Date().toISOString().split("T")[0],
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await FuelService.createLog({
        ...formData,
        transaction_date: new Date(formData.transaction_date).toISOString(),
      });
      navigate("/fuel");
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to create fuel log");
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (
    e: React.ChangeEvent<
      HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
    >,
  ) => {
    const { name, value, type } = e.target;
    const val = type === "number" ? parseFloat(value) : value;

    setFormData((prev) => {
      const next = { ...prev, [name]: val };
      if (name === "quantity" || name === "unit_price") {
        const qty = name === "quantity" ? (val as number) : prev.quantity;
        const price = name === "unit_price" ? (val as number) : prev.unit_price;
        next.total_cost = parseFloat((qty * price).toFixed(2));
      }
      return next;
    });
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex items-center justify-between flex-shrink-0 z-10">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate("/fuel")}
            className="p-1 hover:bg-surface-dark rounded text-text-muted transition-colors"
          >
            <span className="material-symbols-outlined">arrow_back</span>
          </button>
          <h2 className="text-xl font-semibold text-white tracking-tight">
            {uiStrings?.title || "Log New Fuel Transaction"}
          </h2>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-8 custom-scrollbar bg-[radial-gradient(circle_at_top_right,rgba(19,127,236,0.03),transparent)]">
        <div className="max-w-3xl mx-auto">
          <form
            onSubmit={handleSubmit}
            className="bg-surface-dark/40 backdrop-blur-xl border border-border-color rounded-3xl p-8 shadow-2xl space-y-8 animate-slide-up"
          >
            {error && (
              <div className="bg-red-500/10 border border-red-500/20 text-red-500 px-4 py-3 rounded-xl text-sm flex items-center gap-3">
                <span className="material-symbols-outlined">error</span>
                {error}
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <div className="space-y-4">
                <h3 className="text-primary text-xs font-bold uppercase tracking-widest px-1">
                  {uiStrings?.logistics_info || "Logistics Info"}
                </h3>

                <div className="space-y-2">
                  <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                    {uiStrings?.vehicle_id || "Vehicle ID"}
                  </label>
                  <input
                    type="text"
                    name="vehicle_id"
                    value={formData.vehicle_id}
                    onChange={handleChange}
                    required
                    className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all"
                    placeholder="Enter Vehicle UUID"
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                    {uiStrings?.driver_id || "Driver ID"}
                  </label>
                  <input
                    type="text"
                    name="driver_id"
                    value={formData.driver_id}
                    onChange={handleChange}
                    required
                    className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all"
                    placeholder="Enter Driver UUID"
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                    {uiStrings?.current_odometer || "Current Odometer"}
                  </label>
                  <div className="relative">
                    <input
                      type="number"
                      name="odometer"
                      value={formData.odometer}
                      onChange={handleChange}
                      required
                      className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all pr-12"
                    />
                    <span className="absolute right-4 top-2.5 text-text-muted text-xs">
                      {uiStrings?.km_label || "km"}
                    </span>
                  </div>
                </div>
              </div>

              <div className="space-y-4">
                <h3 className="text-accent-yellow text-xs font-bold uppercase tracking-widest px-1">
                  {uiStrings?.transaction_details || "Transaction Details"}
                </h3>

                <div className="space-y-2">
                  <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                    {uiStrings?.station_name || "Station Name"}
                  </label>
                  <input
                    type="text"
                    name="station_name"
                    value={formData.station_name}
                    onChange={handleChange}
                    required
                    className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all"
                    placeholder="e.g., BP Service Station"
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                    {uiStrings?.fuel_type || "Fuel Type"}
                  </label>
                  <select
                    name="fuel_type"
                    value={formData.fuel_type}
                    onChange={handleChange}
                    className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all appearance-none"
                  >
                    <option value="Diesel">
                      {uiStrings?.opt_diesel || "Diesel"}
                    </option>
                    <option value="Petrol 95">
                      {uiStrings?.opt_petrol95 || "Petrol 95"}
                    </option>
                    <option value="Petrol 98">
                      {uiStrings?.opt_petrol98 || "Petrol 98"}
                    </option>
                    <option value="AdBlue">
                      {uiStrings?.opt_adblue || "AdBlue"}
                    </option>
                    <option value="EV Charge">
                      {uiStrings?.opt_ev || "EV Charge"}
                    </option>
                  </select>
                </div>

                <div className="space-y-2">
                  <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                    {uiStrings?.transaction_date || "Transaction Date"}
                  </label>
                  <input
                    type="date"
                    name="transaction_date"
                    value={formData.transaction_date}
                    onChange={handleChange}
                    required
                    className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all"
                  />
                </div>
              </div>
            </div>

            <div className="h-px bg-gradient-to-r from-transparent via-border-color to-transparent"></div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="space-y-2">
                <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                  {uiStrings?.quantity_l || "Quantity (L)"}
                </label>
                <input
                  type="number"
                  step="0.01"
                  name="quantity"
                  value={formData.quantity}
                  onChange={handleChange}
                  required
                  className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all"
                />
              </div>
              <div className="space-y-2">
                <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                  {uiStrings?.unit_price || "Unit Price ($/L)"}
                </label>
                <input
                  type="number"
                  step="0.001"
                  name="unit_price"
                  value={formData.unit_price}
                  onChange={handleChange}
                  required
                  className="w-full bg-background-dark border border-border-color rounded-xl px-4 py-2.5 text-white outline-none focus:ring-1 focus:ring-primary transition-all"
                />
              </div>
              <div className="space-y-2">
                <label className="text-text-muted text-[11px] font-bold uppercase tracking-wider block px-1">
                  {uiStrings?.total_cost || "Total Cost ($)"}
                </label>
                <input
                  type="number"
                  step="0.01"
                  name="total_cost"
                  value={formData.total_cost}
                  readOnly
                  className="w-full bg-surface-dark border border-accent-blue/30 rounded-xl px-4 py-2.5 text-accent-blue font-bold outline-none cursor-not-allowed"
                />
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-4">
              <button
                type="button"
                onClick={() => navigate("/fuel")}
                className="px-6 py-2.5 text-sm text-text-muted hover:text-white transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                className="bg-primary hover:bg-primary/90 text-white text-sm font-semibold py-2.5 px-8 rounded-full shadow-[0_4px_14px_rgba(19,127,236,0.39)] transition-all flex items-center gap-2 hover:-translate-y-0.5"
              >
                {loading ? (
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                ) : (
                  <span className="material-symbols-outlined text-[20px]">
                    check
                  </span>
                )}
                {loading
                  ? uiStrings?.saving_btn || "Saving..."
                  : uiStrings?.submit_btn || "Confirm Entry"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default AddFuelLog;
