import { ChevronRightIcon, XMarkIcon } from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import api from "../../../services/api";

export default function EditTrip() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("tracking");
  const uiStrings = rawUiStrings || {};
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [formData, setFormData] = useState({
    status: "",
    start_location: "",
    end_location: "",
    distance: "",
  });
  const [tripId, setTripId] = useState("");

  useEffect(() => {
    loadTrip();
  }, [id]);

  const loadTrip = async () => {
    try {
      const data = await api.getTrip(id!);
      setTripId(data.trip_id);
      
      const startLoc = typeof data.start_location === "object" ? data.start_location.address : data.start_location;
      const endLoc = typeof data.end_location === "object" ? data.end_location.address : data.end_location;

      setFormData({
        status: data.status || "",
        start_location: startLoc || "",
        end_location: endLoc || "",
        distance: data.distance?.toString() || "",
      });
    } catch (error) {
      console.error("Failed to load trip:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>,
  ) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSaving(true);

    try {
      await api.updateTrip(id!, {
        status: formData.status,
        start_location: formData.start_location,
        end_location: formData.end_location,
        distance: formData.distance ? parseFloat(formData.distance) : undefined,
      });
      navigate(`/admin/trips/${id}`);
    } catch (err: any) {
      setError(
        err.response?.data?.error ||
          uiStrings.error_update_failed ||
          "Failed to update trip",
      );
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6 text-center text-text-muted">
        {uiStrings.msg_loading || "Loading..."}
      </div>
    );
  }

  return (
    <div className="p-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <Link to="/admin/trips" className="hover:text-white">
              {uiStrings.trips_title || "Trips"}
            </Link>
            <ChevronRightIcon className="w-3 h-3" />
            <Link to={`/admin/trips/${id}`} className="hover:text-white">
              {tripId}
            </Link>
            <ChevronRightIcon className="w-3 h-3" />
            <span className="text-white">
              {uiStrings.breadcrumb_edit || "Edit"}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-white">
            {uiStrings.title_edit_trip || "Edit Trip Details"}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to={`/admin/trips/${id}`}
            className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white text-sm rounded transition-colors"
          >
            {uiStrings.btn_cancel || "Cancel"}
          </Link>
          <button
            onClick={handleSubmit}
            disabled={saving}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded transition-colors disabled:opacity-50"
          >
            {saving
              ? uiStrings.msg_saving || "Saving..."
              : uiStrings.btn_save_changes || "Save Changes"}
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-danger/10 border border-danger/20 rounded-lg flex items-center gap-2 text-danger">
          <XMarkIcon className="w-5 h-5" />
          {error}
        </div>
      )}

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="p-4 border-b border-border-color bg-surface-darker">
            <h3 className="text-white font-semibold">
              {uiStrings.section_trip_info || "Trip Information"}
            </h3>
          </div>
          <div className="p-6 space-y-6">
            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted">
                  {uiStrings.th_trip_id || "Trip ID"}
                </label>
                <input
                  type="text"
                  value={tripId}
                  disabled
                  className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-text-muted opacity-70 cursor-not-allowed"
                />
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted">
                  {uiStrings.th_status || "Status"}
                </label>
                <select
                  name="status"
                  value={formData.status}
                  onChange={handleChange}
                  className="w-full bg-background-dark border border-border-color rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                >
                  <option value="scheduled">
                    {uiStrings.filter_scheduled || "Scheduled"}
                  </option>
                  <option value="in_progress">
                    {uiStrings.filter_in_progress || "In Progress"}
                  </option>
                  <option value="completed">
                    {uiStrings.filter_completed || "Completed"}
                  </option>
                  <option value="cancelled">
                    {uiStrings.filter_cancelled || "Cancelled"}
                  </option>
                </select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted">
                  {uiStrings.th_start_location || "Start Location"}
                </label>
                <input
                  type="text"
                  name="start_location"
                  value={formData.start_location}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                />
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted">
                  {uiStrings.th_end_location || "End Location"}
                </label>
                <input
                  type="text"
                  name="end_location"
                  value={formData.end_location}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                />
              </div>
            </div>
            <div className="max-w-xs space-y-1.5">
              <label className="text-xs font-medium text-text-muted">
                {uiStrings.th_distance || "Distance"} (
                {uiStrings.miles_label || "km"})
              </label>
              <input
                type="number"
                name="distance"
                value={formData.distance}
                onChange={handleChange}
                step="0.1"
                className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
              />
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}
