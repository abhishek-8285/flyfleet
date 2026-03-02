import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface PlatformSettings {
  id?: string;
  system_name: string;
  support_email: string;
  default_timezone: string;
  default_currency: string;
  default_language: string;
  white_label_enabled: boolean;
}

const PlatformConfig: React.FC = () => {
  const [settings, setSettings] = useState<PlatformSettings>({
    system_name: "Enterprise Fleet Manager",
    support_email: "support@fleet-system.com",
    default_timezone: "UTC",
    default_currency: "USD",
    default_language: "en-US",
    white_label_enabled: false,
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [globalConfig, setGlobalConfig] = useState<any>(null);
  const { uiStrings } = useUIStrings("admin_settings");

  useEffect(() => {
    fetchSettings();
    fetchGlobalConfig();
  }, []);

  const fetchGlobalConfig = async () => {
    try {
      const response = (await apiClient.get("/ui/global_config")) as any;
      setGlobalConfig(response?.data || response);
    } catch (err) {
      console.error("Failed to fetch global config", err);
    }
  };

  const fetchSettings = async () => {
    setError(null);
    try {
      const response = (await apiClient.get("/settings/platform")) as any;
      const data = response?.data ?? response;
      if (data) {
        setSettings((prev) => ({
          system_name: data.system_name || prev.system_name,
          support_email: data.support_email || prev.support_email,
          default_timezone: data.default_timezone || prev.default_timezone,
          default_currency: data.default_currency || prev.default_currency,
          default_language: data.default_language || prev.default_language,
          white_label_enabled: Boolean(data.white_label_enabled),
        }));
      }
    } catch (error) {
      console.error("Failed to fetch platform settings", error);
      setError(
        uiStrings?.error_load_platform ||
          "Failed to load platform configuration",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiClient.put("/settings/platform", settings);
      alert(uiStrings?.msg_save_success || "Settings saved successfully!");
    } catch (error) {
      console.error("Failed to save settings", error);
      alert(uiStrings?.msg_error_save || "Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  if (loading)
    return (
      <div className="p-6 text-white">{uiStrings?.loading || "Loading..."}</div>
    );

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex flex-col">
          <div className="flex items-center gap-2 text-[10px] text-[#92adc9] uppercase tracking-wider font-semibold mb-0.5">
            <span>{uiStrings?.breadcrumb_system || "System Settings"}</span>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span className="text-[#137fec]">
              {uiStrings?.breadcrumb_platform || "Platform Config"}
            </span>
          </div>
          <h2 className="text-white text-lg font-bold">
            {uiStrings?.title_global_platform ||
              "Global Platform Configuration"}
          </h2>
        </div>
        <div className="flex items-center gap-3">
          <button className="px-4 py-2 bg-[#16202a] hover:bg-[#233648] text-[#92adc9] hover:text-white border border-[#233648] text-xs font-semibold rounded transition-colors">
            {uiStrings?.btn_cancel || "Cancel"}
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-[#137fec]/20 disabled:opacity-50"
          >
            <span className="material-symbols-outlined text-[16px]">save</span>
            {saving
              ? uiStrings?.btn_saving || "Saving..."
              : uiStrings?.btn_save_config || "Save Configuration"}
          </button>
        </div>
      </header>
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {error && (
          <div className="mb-4 rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">
            {error}
          </div>
        )}
        <div className="max-w-5xl mx-auto space-y-6">
          <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]/50 flex justify-between items-center">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-[#137fec] text-[18px]">
                  branding_watermark
                </span>
                {uiStrings?.card_branding || "General Branding"}
              </h3>
              <span className="text-[10px] text-[#92adc9] uppercase tracking-wide">
                {uiStrings?.label_public_facing || "Public Facing"}
              </span>
            </div>
            <div className="p-6 grid grid-cols-1 md:grid-cols-12 gap-8">
              <div className="md:col-span-4">
                <label className="block text-xs font-medium text-[#92adc9] mb-2">
                  {uiStrings?.label_logo || "Platform Logo"}
                </label>
                <div className="border-2 border-dashed border-[#233648] rounded-lg p-6 flex flex-col items-center justify-center text-center hover:bg-[#111a22] transition-colors cursor-pointer group">
                  <div className="w-16 h-16 bg-[#111a22] rounded mb-3 flex items-center justify-center border border-[#233648] group-hover:border-[#137fec]/50 transition-colors">
                    <span className="material-symbols-outlined text-3xl text-[#92adc9] group-hover:text-[#137fec]">
                      upload_file
                    </span>
                  </div>
                  <p className="text-xs text-white font-medium">
                    {uiStrings?.label_upload || "Click to upload"}
                  </p>
                  <p className="text-[10px] text-[#92adc9] mt-1">
                    {uiStrings?.help_upload || "SVG, PNG or JPG (max 2MB)"}
                  </p>
                </div>
              </div>
              <div className="md:col-span-8 space-y-5">
                <div>
                  <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                    {uiStrings?.label_system_name || "System Name"}
                  </label>
                  <input
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] focus:ring-1 focus:ring-[#137fec] placeholder-gray-600"
                    type="text"
                    value={settings.system_name}
                    onChange={(e) =>
                      setSettings({ ...settings, system_name: e.target.value })
                    }
                  />
                  <p className="text-[10px] text-[#92adc9] mt-1">
                    {uiStrings?.help_system_name ||
                      "This name will appear in emails and browser tabs."}
                  </p>
                </div>
                <div>
                  <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                    {uiStrings?.label_support_email || "Support Email Contact"}
                  </label>
                  <input
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] focus:ring-1 focus:ring-[#137fec] placeholder-gray-600"
                    type="email"
                    value={settings.support_email}
                    onChange={(e) =>
                      setSettings({
                        ...settings,
                        support_email: e.target.value,
                      })
                    }
                  />
                </div>
                <div className="flex items-center gap-3 mt-4">
                  <div className="flex items-center h-5">
                    <input
                      className="w-4 h-4 rounded border-[#233648] bg-[#111a22] text-[#137fec] focus:ring-offset-[#16202a] focus:ring-[#137fec]"
                      id="whitelabel"
                      type="checkbox"
                      checked={settings.white_label_enabled}
                      onChange={(e) =>
                        setSettings({
                          ...settings,
                          white_label_enabled: e.target.checked,
                        })
                      }
                    />
                  </div>
                  <div className="text-xs">
                    <label
                      className="font-medium text-white"
                      htmlFor="whitelabel"
                    >
                      {uiStrings?.label_whitelabel || "Enable White-label Mode"}
                    </label>
                    <p className="text-[#92adc9]">
                      {uiStrings?.help_whitelabel ||
                        "Removes all vendor branding from the footer."}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]/50 flex justify-between items-center">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-[#137fec] text-[18px]">
                  public
                </span>
                {uiStrings?.card_regional || "Regional Defaults"}
              </h3>
              <span className="text-[10px] text-[#92adc9] uppercase tracking-wide">
                {uiStrings?.label_system_wide || "System Wide"}
              </span>
            </div>
            <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  {uiStrings?.label_timezone || "Default Timezone"}
                </label>
                <div className="relative">
                  <select
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] focus:ring-1 focus:ring-[#137fec] appearance-none"
                    value={settings.default_timezone}
                    onChange={(e) =>
                      setSettings({
                        ...settings,
                        default_timezone: e.target.value,
                      })
                    }
                  >
                    {globalConfig?.timezones ? (
                      globalConfig.timezones.map((t: any) => (
                        <option key={t.id} value={t.id}>
                          {t.name}
                        </option>
                      ))
                    ) : (
                      <option value="UTC">UTC</option>
                    )}
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-[#92adc9]">
                    <span className="material-symbols-outlined text-[16px]">
                      expand_more
                    </span>
                  </div>
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  {uiStrings?.label_currency || "Default Currency"}
                </label>
                <div className="relative">
                  <select
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] focus:ring-1 focus:ring-[#137fec] appearance-none"
                    value={settings.default_currency}
                    onChange={(e) =>
                      setSettings({
                        ...settings,
                        default_currency: e.target.value,
                      })
                    }
                  >
                    {globalConfig?.currencies ? (
                      globalConfig.currencies.map((c: any) => (
                        <option key={c.id} value={c.id}>
                          {c.name}
                        </option>
                      ))
                    ) : (
                      <option value="USD">USD ($)</option>
                    )}
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-[#92adc9]">
                    <span className="material-symbols-outlined text-[16px]">
                      expand_more
                    </span>
                  </div>
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  {uiStrings?.label_language || "System Language"}
                </label>
                <div className="relative">
                  <select
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] focus:ring-1 focus:ring-[#137fec] appearance-none"
                    value={settings.default_language}
                    onChange={(e) =>
                      setSettings({
                        ...settings,
                        default_language: e.target.value,
                      })
                    }
                  >
                    {globalConfig?.languages ? (
                      globalConfig.languages.map((l: any) => (
                        <option key={l.id} value={l.id}>
                          {l.name}
                        </option>
                      ))
                    ) : (
                      <option value="en-US">English (US)</option>
                    )}
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-[#92adc9]">
                    <span className="material-symbols-outlined text-[16px]">
                      expand_more
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
};

export default PlatformConfig;
