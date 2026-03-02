import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface NotificationChannel {
  id: string;
  channel_type: string;
  provider: string;
  config: Record<string, any>;
  enabled: boolean;
  updated_at: string;
}

const NotificationChannels: React.FC = () => {
  const [channels, setChannels] = useState<NotificationChannel[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { uiStrings } = useUIStrings("admin_settings");

  useEffect(() => {
    fetchChannels();
  }, []);

  const fetchChannels = async () => {
    setError(null);
    try {
      const response = (await apiClient.get("/settings/channels")) as any;
      const items = Array.isArray(response)
        ? response
        : Array.isArray(response?.data)
          ? response.data
          : [];
      setChannels(items);
    } catch (error) {
      console.error("Failed to fetch channels", error);
      setError(
        uiStrings?.error_load_channels ||
          "Failed to load notification channels",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleUpdate = async (channel: NotificationChannel) => {
    setSaving(channel.id);
    try {
      await apiClient.put("/settings/channels", channel);
      setChannels(channels.map((c) => (c.id === channel.id ? channel : c)));
      alert(uiStrings?.msg_save_success || "Channel configuration saved!");
    } catch (error) {
      console.error("Failed to update channel", error);
      alert(
        uiStrings?.msg_error_save || "Failed to update channel configuration",
      );
    } finally {
      setSaving(null);
    }
  };

  const handleConfigChange = (id: string, key: string, value: string) => {
    setChannels(
      channels.map((c) => {
        if (c.id === id) {
          return { ...c, config: { ...c.config, [key]: value } };
        }
        return c;
      }),
    );
  };

  if (loading)
    return (
      <div className="p-6 text-white">{uiStrings?.loading || "Loading..."}</div>
    );

  const emailChannel = channels.find((c) => c.channel_type === "EMAIL") || {
    id: "temp-email",
    channel_type: "EMAIL",
    provider: "SendGrid",
    config: { api_key: "", from_email: "" },
    enabled: false,
    updated_at: new Date().toISOString(),
  };

  const smsChannel = channels.find((c) => c.channel_type === "SMS") || {
    id: "temp-sms",
    channel_type: "SMS",
    provider: "Twilio",
    config: { account_sid: "", auth_token: "", from_number: "" },
    enabled: false,
    updated_at: new Date().toISOString(),
  };

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-white text-lg font-bold">
              {uiStrings?.title_gateways || "Notification Gateways"}
            </h2>
            <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5 uppercase tracking-wide font-medium">
              <span>{uiStrings?.breadcrumb_system || "System Settings"}</span>
              <span className="material-symbols-outlined text-[10px]">
                chevron_right
              </span>
              <span className="text-white">
                {uiStrings?.breadcrumb_channels || "Channels"}
              </span>
            </div>
          </div>
        </div>
      </header>
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {error && (
          <div className="mb-4 rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">
            {error}
          </div>
        )}
        <div className="max-w-5xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-8">
          {/* Email Channel */}
          <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]/50 flex justify-between items-center">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-[#137fec] text-[20px]">
                  mail
                </span>
                {uiStrings?.card_email || "Email Configuration"}
              </h3>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={emailChannel.enabled}
                  onChange={(e) => {
                    const updated = {
                      ...emailChannel,
                      enabled: e.target.checked,
                    };
                    setChannels(
                      channels.map((c) =>
                        c.id === emailChannel.id ? updated : c,
                      ),
                    ); // Optimistic UI update for toggle
                    if (emailChannel.id !== "temp-email") handleUpdate(updated);
                  }}
                  className="sr-only peer"
                />
                <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#10b981]"></div>
              </label>
            </div>
            <div className="p-6 space-y-5">
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  Provider
                </label>
                <div className="relative">
                  <select
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] appearance-none"
                    value={emailChannel.provider}
                    onChange={(e) => {
                      const updated = {
                        ...emailChannel,
                        provider: e.target.value,
                      };
                      setChannels(
                        channels.map((c) =>
                          c.id === emailChannel.id ? updated : c,
                        ),
                      );
                    }}
                  >
                    <option value="SendGrid">SendGrid</option>
                    <option value="SES">Amazon SES</option>
                    <option value="SMTP">Custom SMTP</option>
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-[#92adc9]">
                    <span className="material-symbols-outlined text-[16px]">
                      expand_more
                    </span>
                  </div>
                </div>
              </div>

              {emailChannel.provider === "SendGrid" && (
                <>
                  <div>
                    <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                      {uiStrings?.label_api_key || "API Key"}
                    </label>
                    <input
                      className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] placeholder-gray-600 font-mono"
                      type="password"
                      placeholder="SG.xxxxxxxx"
                      value={emailChannel.config["api_key"] || ""}
                      onChange={(e) =>
                        handleConfigChange(
                          emailChannel.id,
                          "api_key",
                          e.target.value,
                        )
                      }
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                      {uiStrings?.label_from_email || "From Email"}
                    </label>
                    <input
                      className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] placeholder-gray-600"
                      type="email"
                      placeholder="notifications@yourdomain.com"
                      value={emailChannel.config["from_email"] || ""}
                      onChange={(e) =>
                        handleConfigChange(
                          emailChannel.id,
                          "from_email",
                          e.target.value,
                        )
                      }
                    />
                  </div>
                </>
              )}

              <div className="pt-2">
                <button
                  onClick={() => handleUpdate(emailChannel)}
                  disabled={saving === emailChannel.id}
                  className="w-full py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded transition-colors shadow-lg shadow-[#137fec]/20 disabled:opacity-50"
                >
                  {saving === emailChannel.id
                    ? uiStrings?.btn_saving || "Saving..."
                    : uiStrings?.btn_save_email || "Save Email Configuration"}
                </button>
              </div>
            </div>
          </div>

          {/* SMS Channel */}
          <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]/50 flex justify-between items-center">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-[#137fec] text-[20px]">
                  sms
                </span>
                {uiStrings?.card_sms || "SMS Configuration"}
              </h3>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={smsChannel.enabled}
                  onChange={(e) => {
                    const updated = {
                      ...smsChannel,
                      enabled: e.target.checked,
                    };
                    setChannels(
                      channels.map((c) =>
                        c.id === smsChannel.id ? updated : c,
                      ),
                    );
                    if (smsChannel.id !== "temp-sms") handleUpdate(updated);
                  }}
                  className="sr-only peer"
                />
                <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#10b981]"></div>
              </label>
            </div>
            <div className="p-6 space-y-5">
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  Provider
                </label>
                <div className="relative">
                  <select
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] appearance-none"
                    value={smsChannel.provider}
                    onChange={(e) => {
                      const updated = {
                        ...smsChannel,
                        provider: e.target.value,
                      };
                      setChannels(
                        channels.map((c) =>
                          c.id === smsChannel.id ? updated : c,
                        ),
                      );
                    }}
                  >
                    <option value="Twilio">Twilio</option>
                    <option value="SNS">AWS SNS</option>
                    <option value="Nexmo">Nexmo / Vonage</option>
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-[#92adc9]">
                    <span className="material-symbols-outlined text-[16px]">
                      expand_more
                    </span>
                  </div>
                </div>
              </div>

              {smsChannel.provider === "Twilio" && (
                <>
                  <div>
                    <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                      {uiStrings?.label_account_sid || "Account SID"}
                    </label>
                    <input
                      className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] placeholder-gray-600 font-mono"
                      type="text"
                      placeholder="ACxxxxxxxx"
                      value={smsChannel.config["account_sid"] || ""}
                      onChange={(e) =>
                        handleConfigChange(
                          smsChannel.id,
                          "account_sid",
                          e.target.value,
                        )
                      }
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                      {uiStrings?.label_auth_token || "Auth Token"}
                    </label>
                    <input
                      className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] placeholder-gray-600 font-mono"
                      type="password"
                      value={smsChannel.config["auth_token"] || ""}
                      onChange={(e) =>
                        handleConfigChange(
                          smsChannel.id,
                          "auth_token",
                          e.target.value,
                        )
                      }
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                      {uiStrings?.label_from_number || "From Number"}
                    </label>
                    <input
                      className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] placeholder-gray-600 font-mono"
                      type="text"
                      placeholder="+15551234567"
                      value={smsChannel.config["from_number"] || ""}
                      onChange={(e) =>
                        handleConfigChange(
                          smsChannel.id,
                          "from_number",
                          e.target.value,
                        )
                      }
                    />
                  </div>
                </>
              )}

              <div className="pt-2">
                <button
                  onClick={() => handleUpdate(smsChannel)}
                  disabled={saving === smsChannel.id}
                  className="w-full py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded transition-colors shadow-lg shadow-[#137fec]/20 disabled:opacity-50"
                >
                  {saving === smsChannel.id
                    ? uiStrings?.btn_saving || "Saving..."
                    : uiStrings?.btn_save_sms || "Save SMS Configuration"}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
};

export default NotificationChannels;
