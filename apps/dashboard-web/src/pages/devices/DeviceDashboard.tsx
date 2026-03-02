import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

interface DeviceSummary {
  total_devices: number;
  online_devices: number;
  offline_devices: number;
  low_battery: number;
  weak_signal: number;
  system_errors: number;
  new_devices_week: number;
}

interface Device {
  id: string;
  serial_number: string;
  type: string;
  status: string;
  firmware_version?: string;
  battery_level?: number;
  signal_strength?: number;
  last_heartbeat?: string;
}

interface DeviceTwin {
  device_id: string;
  drift_status: "in_sync" | "drifted" | "unknown";
  drift_reason?: string;
  last_seen?: string;
}

const DeviceDashboard: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("devices");
  const uiStrings = rawUiStrings || {};

  const [summary, setSummary] = useState<DeviceSummary | null>(null);
  const [devices, setDevices] = useState<Device[]>([]);
  const [twinsByDevice, setTwinsByDevice] = useState<
    Record<string, DeviceTwin>
  >({});
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("all");

  const fetchDashboardData = useCallback(async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem("access_token");
      const headers = {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      };

      const [summaryRes, devicesRes] = await Promise.all([
        fetch("/api/v1/devices/summary", { headers }),
        fetch(
          `/api/v1/devices/status?status=${filter === "all" ? "" : filter}`,
          { headers },
        ),
      ]);
      const twinsRes = await fetch("/api/v1/devices/twins?limit=200", {
        headers,
      });

      if (summaryRes.ok) setSummary(await summaryRes.json());
      if (devicesRes.ok) setDevices((await devicesRes.json()) || []);
      if (twinsRes.ok) {
        const twins = (await twinsRes.json()) || [];
        const map: Record<string, DeviceTwin> = {};
        (Array.isArray(twins) ? twins : []).forEach((t: DeviceTwin) => {
          map[t.device_id] = t;
        });
        setTwinsByDevice(map);
      }
    } catch (error) {
      console.error("Failed to fetch dashboard data:", error);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    void fetchDashboardData();
  }, [fetchDashboardData]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case "active":
        return "bg-emerald-500/20 text-emerald-400 border-emerald-500/30";
      case "offline":
        return "bg-red-500/20 text-red-400 border-red-500/30";
      case "error":
        return "bg-orange-500/20 text-orange-400 border-orange-500/30";
      default:
        return "bg-slate-500/20 text-slate-400 border-slate-500/30";
    }
  };

  const getBatteryColor = (level?: number) => {
    if (!level) return "text-slate-400";
    if (level < 20) return "text-red-400";
    if (level < 50) return "text-yellow-400";
    return "text-emerald-400";
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">
          {uiStrings.dashboard_title || "Device Diagnostics"}
        </h1>
        <p className="text-slate-400">
          {uiStrings.dashboard_sub ||
            "Monitor device health, telemetry, and firmware status"}
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-4 mb-8">
        <StatCard
          label={uiStrings.stat_total_devices || "Total Devices"}
          value={summary?.total_devices || 0}
          icon="📡"
        />
        <StatCard
          label={uiStrings.stat_online || "Online"}
          value={summary?.online_devices || 0}
          icon="🟢"
          color="text-emerald-400"
        />
        <StatCard
          label={uiStrings.stat_offline || "Offline"}
          value={summary?.offline_devices || 0}
          icon="🔴"
          color="text-red-400"
        />
        <StatCard
          label={uiStrings.stat_low_battery || "Low Battery"}
          value={summary?.low_battery || 0}
          icon="🔋"
          color="text-yellow-400"
        />
        <StatCard
          label={uiStrings.stat_weak_signal || "Weak Signal"}
          value={summary?.weak_signal || 0}
          icon="📶"
          color="text-orange-400"
        />
        <StatCard
          label={uiStrings.stat_errors || "Errors"}
          value={summary?.system_errors || 0}
          icon="⚠️"
          color="text-red-400"
        />
        <StatCard
          label={uiStrings.stat_new_7d || "New (7d)"}
          value={summary?.new_devices_week || 0}
          icon="✨"
          color="text-violet-400"
        />
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <ActionCard
          title={uiStrings.action_telemetry_title || "Live Telemetry"}
          description={
            uiStrings.action_telemetry_desc || "Real-time device data streams"
          }
          icon="📊"
          onClick={() => navigate("/admin/devices/telemetry")}
        />
        <ActionCard
          title={uiStrings.action_commands_title || "Remote Commands"}
          description={
            uiStrings.action_commands_desc || "Send commands to devices"
          }
          icon="⚡"
          onClick={() => navigate("/admin/devices/commands")}
        />
        <ActionCard
          title={uiStrings.action_firmware_title || "Firmware Manager"}
          description={
            uiStrings.action_firmware_desc || "Manage firmware rollouts"
          }
          icon="📦"
          onClick={() => navigate("/admin/devices/firmware")}
        />
        <ActionCard
          title={uiStrings.action_health_title || "Device Health"}
          description={uiStrings.action_health_desc || "Detailed diagnostics"}
          icon="❤️"
          onClick={() => navigate("/admin/devices/diagnostics")}
        />
        <ActionCard
          title={uiStrings.action_twins_title || "Device Twins"}
          description={
            uiStrings.action_twins_desc || "Desired vs reported state"
          }
          icon="🧬"
          onClick={() => navigate("/admin/devices/twins")}
        />
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2 mb-6">
        {["all", "active", "offline", "error"].map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              filter === f
                ? "bg-violet-600 text-white"
                : "bg-slate-800/50 text-slate-400 hover:bg-slate-700/50"
            }`}
          >
            {uiStrings[`filter_${f}`] || f.charAt(0).toUpperCase() + f.slice(1)}
          </button>
        ))}
      </div>

      {/* Devices Table */}
      <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-700/50">
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_device || "Device"}
                </th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_type || "Type"}
                </th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_status || "Status"}
                </th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_twin || "Twin"}
                </th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_firmware || "Firmware"}
                </th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_battery || "Battery"}
                </th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_signal || "Signal"}
                </th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  {uiStrings.th_last_seen || "Last Seen"}
                </th>
                <th className="py-4 px-6"></th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={9} className="text-center py-12 text-slate-400">
                    {uiStrings.msg_loading_devices || "Loading devices..."}
                  </td>
                </tr>
              ) : devices.length === 0 ? (
                <tr>
                  <td colSpan={9} className="text-center py-12 text-slate-400">
                    {uiStrings.msg_no_devices_found || "No devices found"}
                  </td>
                </tr>
              ) : (
                devices.map((device) => (
                  <tr
                    key={device.id}
                    className="border-b border-slate-700/30 hover:bg-slate-700/30 transition-colors"
                  >
                    <td className="py-4 px-6">
                      <span className="text-white font-medium">
                        {device.serial_number}
                      </span>
                    </td>
                    <td className="py-4 px-6">
                      <span className="text-slate-300">
                        {device.type.toUpperCase()}
                      </span>
                    </td>
                    <td className="py-4 px-6">
                      <span
                        className={`px-3 py-1 rounded-full text-xs font-medium border ${getStatusColor(device.status)}`}
                      >
                        {device.status}
                      </span>
                    </td>
                    <td className="py-4 px-6">
                      <span
                        className={`px-3 py-1 rounded-full text-xs font-medium border ${
                          twinsByDevice[device.id]?.drift_status === "in_sync"
                            ? "bg-emerald-500/20 text-emerald-400 border-emerald-500/30"
                            : twinsByDevice[device.id]?.drift_status ===
                                "drifted"
                              ? "bg-red-500/20 text-red-400 border-red-500/30"
                              : "bg-slate-500/20 text-slate-300 border-slate-500/30"
                        }`}
                      >
                        {(
                          twinsByDevice[device.id]?.drift_status || "unknown"
                        ).replace("_", " ")}
                      </span>
                    </td>
                    <td className="py-4 px-6 text-slate-300">
                      {device.firmware_version || "-"}
                    </td>
                    <td className="py-4 px-6">
                      <span className={getBatteryColor(device.battery_level)}>
                        {device.battery_level
                          ? `${device.battery_level}%`
                          : "-"}
                      </span>
                    </td>
                    <td className="py-4 px-6 text-slate-300">
                      {device.signal_strength
                        ? `${device.signal_strength} dBm`
                        : "-"}
                    </td>
                    <td className="py-4 px-6 text-slate-400 text-sm">
                      {device.last_heartbeat
                        ? new Date(device.last_heartbeat).toLocaleString()
                        : "-"}
                    </td>
                    <td className="py-4 px-6">
                      <button
                        onClick={() => navigate(`/admin/devices/${device.id}`)}
                        className="text-violet-400 hover:text-violet-300 text-sm font-medium"
                      >
                        {uiStrings.btn_view || "View →"}
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

const StatCard: React.FC<{
  label: string;
  value: number;
  icon: string;
  color?: string;
}> = ({ label, value, icon, color = "text-white" }) => (
  <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-4">
    <div className="flex items-center gap-2 mb-2">
      <span className="text-lg">{icon}</span>
      <span className="text-xs text-slate-400 uppercase tracking-wider">
        {label}
      </span>
    </div>
    <span className={`text-2xl font-bold ${color}`}>{value}</span>
  </div>
);

const ActionCard: React.FC<{
  title: string;
  description: string;
  icon: string;
  onClick: () => void;
}> = ({ title, description, icon, onClick }) => (
  <button
    onClick={onClick}
    className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-5 text-left hover:bg-slate-700/50 hover:border-violet-500/50 transition-all group"
  >
    <span className="text-2xl mb-3 block">{icon}</span>
    <h3 className="text-white font-semibold group-hover:text-violet-400 transition-colors">
      {title}
    </h3>
    <p className="text-sm text-slate-400">{description}</p>
  </button>
);

export default DeviceDashboard;
