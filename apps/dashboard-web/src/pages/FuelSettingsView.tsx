import {
    CurrencyDollarIcon,
    ShieldCheckIcon
} from "@heroicons/react/24/outline";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../hooks/useUIStrings";
import FuelService from "../services/FuelService";
import { FuelSettings } from "../types/fuel";

export default function FuelSettingsView() {
  const navigate = useNavigate();
  const [settings, setSettings] = useState<FuelSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const { uiStrings: rawUiStrings } = useUIStrings("fuel");
  const uiStrings = rawUiStrings || {};

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const data = await FuelService.getSettings();
        setSettings(data);
      } catch (error) {
        console.error("Failed to fetch fuel settings:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchSettings();
  }, []);

  const handleUpdate = async (update: Partial<FuelSettings>) => {
    if (!settings) return;
    setSaving(true);
    try {
      const updated = await FuelService.updateSettings(update);
      setSettings(updated);
    } catch (error) {
      console.error("Failed to update settings:", error);
    } finally {
      setSaving(false);
    }
  };

  if (loading) return null;

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex items-center justify-between flex-shrink-0 z-10 font-poppins">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate("/fuel")}
            className="p-1 hover:bg-surface-dark rounded text-text-muted transition-colors"
          >
            <span className="material-symbols-outlined">arrow_back</span>
          </button>
          <h1 className="text-xl font-semibold text-white tracking-tight">
            {uiStrings?.title_settings || "Fuel Management Settings"}
          </h1>
        </div>
        {saving && (
          <div className="text-xs text-primary animate-pulse flex items-center gap-2">
            <div className="w-3 h-3 border border-primary border-t-transparent rounded-full animate-spin"></div>
            {uiStrings?.saving_changes || "Saving changes..."}
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
        <div className="max-w-3xl mx-auto space-y-8 animate-fade-in">
          {/* General Settings */}
          <div className="bg-surface-dark rounded-2xl border border-border-color overflow-hidden shadow-xl mb-6">
            <div className="p-4 border-b border-border-color bg-surface-darker/50 flex items-center gap-3">
              <CurrencyDollarIcon className="w-5 h-5 text-primary" />
              <h3 className="text-white font-semibold">
                {uiStrings?.section_regional || "Regional & Currency"}
              </h3>
            </div>

            <div className="p-6">
              <div className="max-w-xs space-y-2">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wider">
                  {uiStrings?.label_currency || "Default Currency"}
                </label>
                <select
                  value={settings?.default_currency}
                  onChange={(e) =>
                    handleUpdate({ default_currency: e.target.value })
                  }
                  className="w-full bg-background-dark border border-border-color rounded-xl py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                >
                  <option value="USD">USD ($)</option>
                  <option value="EUR">EUR (€)</option>
                  <option value="GBP">GBP (£)</option>
                  <option value="INR">INR (₹)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Validation Rules */}
          <div className="bg-surface-dark rounded-2xl border border-border-color overflow-hidden shadow-xl mb-6">
            <div className="p-4 border-b border-border-color bg-surface-darker/50 flex items-center gap-3">
              <ShieldCheckIcon className="w-5 h-5 text-accent-yellow" />
              <h3 className="text-white font-semibold">
                {uiStrings?.section_validation || "Entry Validation Rules"}
              </h3>
            </div>

            <div className="p-6 space-y-6">
              <div className="flex items-center justify-between p-4 bg-background-dark border border-border-color/50 rounded-xl group hover:border-primary/30 transition-colors">
                <div>
                  <h4 className="text-sm font-medium text-white">
                    {uiStrings?.label_odometer_check ||
                      "Previous Odometer Check"}
                  </h4>
                  <p className="text-xs text-text-muted mt-1">
                    {uiStrings?.desc_odometer_check ||
                      "Prevent fuel entries with odometer readings lower than preceding logs."}
                  </p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    className="sr-only peer"
                    checked={settings?.validation_rules.odometer_check}
                    onChange={(e) =>
                      handleUpdate({
                        validation_rules: {
                          ...settings!.validation_rules,
                          odometer_check: e.target.checked,
                        },
                      })
                    }
                  />
                  <div className="w-11 h-6 bg-surface-dark peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                </label>
              </div>

              <div className="flex items-center justify-between p-4 bg-background-dark border border-border-color/50 rounded-xl group hover:border-primary/30 transition-colors">
                <div>
                  <h4 className="text-sm font-medium text-white">
                    {uiStrings?.label_cost_verify || "Cost Verification"}
                  </h4>
                  <p className="text-xs text-text-muted mt-1">
                    {uiStrings?.desc_cost_verify ||
                      "Automatically verify that Total Cost = Quantity × Unit Price."}
                  </p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    className="sr-only peer"
                    checked={settings?.validation_rules.cost_calculation_verify}
                    onChange={(e) =>
                      handleUpdate({
                        validation_rules: {
                          ...settings!.validation_rules,
                          cost_calculation_verify: e.target.checked,
                        },
                      })
                    }
                  />
                  <div className="w-11 h-6 bg-surface-dark peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                </label>
              </div>
            </div>
          </div>

          {/* Fuel Types */}
          <div className="bg-surface-dark rounded-2xl border border-border-color overflow-hidden shadow-xl">
            <div className="p-4 border-b border-border-color bg-surface-darker/50 flex items-center gap-3">
              <div className="w-5 h-5 rounded-full bg-orange-400/20 flex items-center justify-center text-orange-400 text-[10px] font-bold">
                !
              </div>
              <h3 className="text-white font-semibold">
                {uiStrings?.section_fuel_categories ||
                  "Allowed Fuel Categories"}
              </h3>
            </div>

            <div className="p-6 flex flex-wrap gap-2">
              {[
                "Diesel",
                "Petrol 95",
                "Petrol 98",
                "AdBlue",
                "EV Charge",
                "LPG",
              ].map((type) => (
                <button
                  key={type}
                  className={`px-4 py-2 rounded-full text-xs font-medium border transition-all ${
                    (settings?.allowed_fuel_types || []).includes(type)
                      ? "bg-primary/20 border-primary text-primary"
                      : "bg-surface-dark border-border-color text-text-muted hover:border-text-muted/50"
                  }`}
                >
                  {type}
                </button>
              ))}
              <button className="w-8 h-8 rounded-full bg-surface-dark border border-dashed border-border-color flex items-center justify-center text-text-muted hover:text-white hover:border-white transition-all">
                <span className="material-symbols-outlined text-[18px]">
                  add
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
