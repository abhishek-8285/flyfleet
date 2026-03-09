import axios from "axios";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

const NewBooking: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("portal_booking");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    request_type: "standard",
    origin: { address: "", latitude: 0, longitude: 0 },
    destination: { address: "", latitude: 0, longitude: 0 },
    requested_pickup_time: "",
    cargo_details: { weight: 0, volume: 0, description: "" },
    notes: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem("customer_token");
      const config = { headers: { Authorization: `Bearer ${token}` } };

      const payload = {
        ...formData,
        requested_pickup_time: formData.requested_pickup_time
          ? new Date(formData.requested_pickup_time).toISOString()
          : null,
      };

      await axios.post("/api/v1/portal/bookings", payload, config);
      navigate("/portal/dashboard");
    } catch (err: any) {
      setError(err.response?.data || "Failed to create booking request");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate(-1)}
          className="p-2 hover:bg-surface-light rounded-full text-text-muted transition-colors"
        >
          <span className="material-symbols-outlined">arrow_back</span>
        </button>
        <h1 className="text-2xl font-bold text-white">
          {uiStrings?.title || "New Booking Request"}
        </h1>
      </div>

      <form
        onSubmit={handleSubmit}
        className="bg-surface-dark border border-border-color rounded-xl p-6 space-y-6"
      >
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 text-red-500 text-sm rounded-lg flex items-center gap-2">
            <span className="material-symbols-outlined text-[18px]">error</span>
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-2">
            <label className="text-sm font-medium text-text-muted">
              {uiStrings?.label_type || "Request Type"}
            </label>
            <select
              className="w-full bg-surface-darker border border-border-color p-2.5 rounded-lg text-white outline-none focus:border-primary transition-colors"
              value={formData.request_type}
              onChange={(e) =>
                setFormData({ ...formData, request_type: e.target.value })
              }
            >
              <option value="standard">Standard</option>
              <option value="express">Express</option>
              <option value="hazardous">Hazardous</option>
            </select>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-text-muted">
              {uiStrings?.label_pickup || "Requested Pickup"}
            </label>
            <input
              type="datetime-local"
              className="w-full bg-surface-darker border border-border-color p-2 rounded-lg text-white outline-none focus:border-primary transition-colors"
              required
              value={formData.requested_pickup_time}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  requested_pickup_time: e.target.value,
                })
              }
            />
          </div>
        </div>

        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-text-muted">
              {uiStrings?.label_origin || "Origin Address"}
            </label>
            <input
              type="text"
              placeholder="Full pickup address"
              className="w-full bg-surface-darker border border-border-color p-2.5 rounded-lg text-white outline-none focus:border-primary transition-colors"
              required
              value={formData.origin.address}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  origin: { ...formData.origin, address: e.target.value },
                })
              }
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-text-muted">
              {uiStrings?.label_destination || "Destination Address"}
            </label>
            <input
              type="text"
              placeholder="Full delivery address"
              className="w-full bg-surface-darker border border-border-color p-2.5 rounded-lg text-white outline-none focus:border-primary transition-colors"
              required
              value={formData.destination.address}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  destination: {
                    ...formData.destination,
                    address: e.target.value,
                  },
                })
              }
            />
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-text-muted">
            {uiStrings?.label_cargo || "Cargo Description"}
          </label>
          <textarea
            className="w-full bg-surface-darker border border-border-color p-2.5 rounded-lg text-white outline-none focus:border-primary transition-colors h-24"
            placeholder="Describe your goods, weight, etc."
            required
            value={formData.cargo_details.description}
            onChange={(e) =>
              setFormData({
                ...formData,
                cargo_details: {
                  ...formData.cargo_details,
                  description: e.target.value,
                },
              })
            }
          />
        </div>

        <div className="pt-4">
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary hover:bg-primary-dark text-white font-bold py-3 rounded-xl transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 disabled:opacity-50"
          >
            {loading ? (
              <span className="material-symbols-outlined animate-spin">
                progress_activity
              </span>
            ) : (
              <span className="material-symbols-outlined">send</span>
            )}
            {uiStrings?.btn_submit || "Submit Booking Request"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default NewBooking;
