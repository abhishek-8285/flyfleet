import {
    CheckCircleIcon,
    Cog6ToothIcon,
    FolderIcon,
} from "@heroicons/react/24/outline";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import api from "../../../services/api";

interface Settings {
  retention_period: number;
  policies: Record<string, any>;
}

export default function TripSettings() {
  const { uiStrings: rawUiStrings } = useUIStrings("tracking");
  const uiStrings = rawUiStrings || {};
  const [settings, setSettings] = useState<Settings>({
    retention_period: 36,
    policies: {},
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const data = await api.getTripSettings();
      setSettings(data);
    } catch (error) {
      console.error("Failed to load settings:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setSaved(false);
    try {
      await api.updateTripSettings(settings);
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (error) {
      console.error("Failed to save settings:", error);
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
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <Link to="/admin/trips" className="hover:text-white">
              {uiStrings.trips_title || "Trips"}
            </Link>
            <span>→</span>
            <span className="text-white">
              {uiStrings.breadcrumb_settings || "Settings"}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-white">
            {uiStrings.title_trip_config || "Trip Configuration"}
          </h1>
        </div>
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded flex items-center gap-2 transition-colors disabled:opacity-50"
        >
          {saved ? (
            <CheckCircleIcon className="w-4 h-4" />
          ) : (
            <Cog6ToothIcon className="w-4 h-4" />
          )}
          {saving
            ? uiStrings.msg_saving || "Saving..."
            : saved
              ? uiStrings.msg_saved || "Saved!"
              : uiStrings.btn_save_changes || "Save Changes"}
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_active_policies || "Total Active Policies"}
          </div>
          <div className="text-2xl font-bold text-white">
            {Object.keys(settings.policies).length || 1}
          </div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_retention || "Data Retention"}
          </div>
          <div className="text-2xl font-bold text-white">
            {settings.retention_period}{" "}
            <span className="text-sm font-normal text-text-muted">
              {uiStrings.unit_months || "months"}
            </span>
          </div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_compliance || "Compliance Status"}
          </div>
          <div className="text-2xl font-bold text-success">
            {uiStrings.val_compliant || "Compliant"}
          </div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">
            {uiStrings.stat_last_updated || "Last Updated"}
          </div>
          <div className="text-lg font-medium text-white">
            {uiStrings.val_just_now || "Just now"}
          </div>
        </div>
      </div>

      {/* Settings Form */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
          <FolderIcon className="w-5 h-5 text-text-muted" />
          <h3 className="text-white font-semibold">
            {uiStrings.section_retention || "Default Data Retention"}
          </h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">
                {uiStrings.label_retention_period || "Retention Period"}
              </label>
              <select
                value={settings.retention_period}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    retention_period: parseInt(e.target.value),
                  })
                }
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
              >
                <option value={12}>
                  {uiStrings.option_retention_12 || "12 Months (1 Year)"}
                </option>
                <option value={24}>
                  {uiStrings.option_retention_24 || "24 Months (2 Years)"}
                </option>
                <option value={36}>
                  {uiStrings.option_retention_36 ||
                    "36 Months (3 Years) - Recommended"}
                </option>
                <option value={60}>
                  {uiStrings.option_retention_60 || "60 Months (5 Years)"}
                </option>
                <option value={84}>
                  {uiStrings.option_retention_84 || "84 Months (7 Years)"}
                </option>
              </select>
              <p className="text-xs text-text-muted">
                {uiStrings.desc_retention ||
                  "Trip data will be archived after this period"}
              </p>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">
                {uiStrings.label_auto_archive || "Auto-Archive"}
              </label>
              <select
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
                defaultValue="enabled"
              >
                <option value="enabled">
                  {uiStrings.option_enabled || "Enabled"}
                </option>
                <option value="disabled">
                  {uiStrings.option_disabled || "Disabled"}
                </option>
              </select>
              <p className="text-xs text-text-muted">
                {uiStrings.desc_auto_archive ||
                  "Automatically archive old trip data"}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Telemetry Settings */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-4 border-b border-border-color bg-surface-darker">
          <h3 className="text-white font-semibold">
            {uiStrings.section_telemetry_config || "Telemetry Configuration"}
          </h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">
                {uiStrings.label_update_frequency || "Update Frequency"}
              </label>
              <select
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
                defaultValue="10"
              >
                <option value="5">
                  {uiStrings.option_freq_5 || "Every 5 seconds"}
                </option>
                <option value="10">
                  {uiStrings.option_freq_10 || "Every 10 seconds"}
                </option>
                <option value="30">
                  {uiStrings.option_freq_30 || "Every 30 seconds"}
                </option>
                <option value="60">
                  {uiStrings.option_freq_60 || "Every 60 seconds"}
                </option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">
                {uiStrings.label_max_events || "Max Events per Trip"}
              </label>
              <input
                type="number"
                defaultValue={10000}
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
