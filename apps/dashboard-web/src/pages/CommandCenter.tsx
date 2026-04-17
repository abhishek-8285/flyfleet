import {
    ArrowPathIcon,
    BellIcon,
    ChatBubbleLeftRightIcon,
    ExclamationTriangleIcon,
    MapIcon,
    TruckIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import React, { useEffect, useMemo, useState } from "react";
import LiveMap from "../components/LiveMap";
import { useUIStrings } from "../hooks/useUIStrings";
import { useWebSocket } from "../hooks/useWebSocket";
import { getWsUrl } from "../services/telemetryService";
import { Alert, Vehicle } from "../types";

const CommandCenter: React.FC = () => {
  const { lastMessage, isConnected } = useWebSocket(getWsUrl());
  const { uiStrings } = useUIStrings("command_center");

  const [vehicles, setVehicles] = useState<Record<string, Vehicle>>({});
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [selectedVehicleId, setSelectedVehicleId] = useState<string | null>(
    null,
  );
  const [filter, setFilter] = useState<"all" | "critical" | "warning">("all");

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.subject.startsWith("fleet.telemetry.")) {
      const data = lastMessage.data;
      setVehicles((prev) => ({
        ...prev,
        [data.vehicle_id]: {
          ...prev[data.vehicle_id],
          id: data.vehicle_id,
          current_location: {
            latitude: data.latitude,
            longitude: data.longitude,
            speed: data.speed,
            heading: data.heading,
            accuracy: data.accuracy,
            timestamp: new Date().toISOString(),
          },
          last_seen: new Date().toISOString(),
        } as Vehicle,
      }));
    } else if (lastMessage.subject.startsWith("fleet.alerts.")) {
      const alert: Alert = lastMessage.data;
      setAlerts((prev) => [alert, ...prev].slice(0, 100)); // Keep last 100
    }
  }, [lastMessage]);

  const filteredAlerts = useMemo(() => {
    if (filter === "all") return alerts;
    return alerts.filter((a) => a.severity === filter);
  }, [alerts, filter]);

  const vehicleStats = useMemo(() => {
    const vals = Object.values(vehicles);
    return {
      total: vals.length,
      moving: vals.filter((v) => (v.current_location?.speed || 0) > 5).length,
      idle: vals.filter((v) => (v.current_location?.speed || 0) <= 5).length,
      offline: 0, // Logic for offline based on heartbeat
    };
  }, [vehicles]);

  const selectedVehicle = selectedVehicleId
    ? vehicles[selectedVehicleId]
    : null;

  return (
    <div className="flex flex-col h-[calc(100vh-64px)] overflow-hidden bg-surface-dim font-sans antialiased">
      {/* Tactical Header Overlay */}
      <div className="bg-surface-container/60 backdrop-blur-3xl border-b border-outline p-8 shrink-0 z-20 shadow-2xl">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <nav className="text-[9px] font-black uppercase tracking-[0.3em] text-on-surface-variant flex items-center gap-2">
              <span className="material-symbols-outlined text-sm">satellite_alt</span>
              Global Operations Theater
            </nav>
            <h1 className="text-3xl font-black text-white tracking-tighter uppercase flex items-center gap-4">
              {uiStrings?.title || "Fleet Command Core"}
              {!isConnected && (
                <div className="flex items-center gap-2 text-[9px] font-black text-rose-500 bg-rose-500/10 border border-rose-500/20 px-3 py-1 rounded-[10px] animate-pulse">
                  <span className="material-symbols-outlined text-xs animate-spin">sync</span>
                  {uiStrings?.status_disconnected || "OFFLINE / RECONNECTING"}
                </div>
              )}
            </h1>
          </div>

          <div className="flex gap-6">
            <StatCard
              label={uiStrings?.stat_total || "Active Assets"}
              value={vehicleStats.total}
              color="primary"
            />
            <StatCard
              label={uiStrings?.stat_moving || "In Motion"}
              value={vehicleStats.moving}
              color="emerald"
            />
            <StatCard
              label={uiStrings?.stat_idle || "Static"}
              value={vehicleStats.idle}
              color="amber"
            />
          </div>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden relative">
        {/* Left: Intelligence Stream */}
        <div className="w-[400px] border-r border-outline bg-surface-container/40 backdrop-blur-2xl flex flex-col shrink-0 z-10 shadow-3xl">
          <div className="p-8 border-b border-outline flex items-center justify-between bg-surface-dim/40">
            <div className="flex items-center gap-4">
              <span className="material-symbols-outlined text-primary">sensors</span>
              <div className="text-[11px] font-black uppercase tracking-[0.2em] text-white">
                {uiStrings?.feed_title || "Operational Intel"}
              </div>
            </div>
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value as any)}
              className="text-[10px] bg-surface-dim border border-outline rounded-xl px-3 py-1.5 text-white font-black uppercase tracking-widest focus:outline-none focus:border-primary transition-all"
            >
              <option value="all">{uiStrings?.filter_all || "Matrix: All"}</option>
              <option value="critical">{uiStrings?.filter_critical || "Critical Only"}</option>
              <option value="warning">{uiStrings?.filter_warning || "Warning Only"}</option>
            </select>
          </div>
          <div className="flex-1 overflow-y-auto custom-scrollbar">
            {filteredAlerts.length === 0 ? (
              <div className="p-12 text-center">
                <p className="text-[10px] font-black uppercase tracking-[0.3em] text-on-surface-variant opacity-30">
                  {uiStrings?.no_alerts || "Zero Signal Interference"}
                </p>
              </div>
            ) : (
              filteredAlerts.map((alert) => (
                <div
                  key={alert.id}
                  className="p-8 border-b border-outline hover:bg-surface-container-highest/50 transition-all cursor-pointer group"
                >
                  <div className="flex justify-between items-start mb-4">
                    <span
                      className={`text-[9px] font-black uppercase tracking-[0.2em] px-3 py-1 rounded-lg border ${
                        alert.severity === "critical"
                          ? "bg-rose-500/10 text-rose-400 border-rose-500/20"
                          : alert.severity === "warning"
                            ? "bg-amber-500/10 text-amber-400 border-amber-500/20"
                            : "bg-primary/10 text-primary border-primary/20"
                      }`}
                    >
                      {alert.severity}
                    </span>
                    <span className="text-[10px] font-mono text-on-surface-variant opacity-40 font-bold">
                      {format(new Date(alert.timestamp), "HH:mm:ss")}
                    </span>
                  </div>
                  <p className="text-xs font-bold text-white tracking-tight leading-relaxed mb-4 group-hover:text-primary transition-colors">
                    {alert.message}
                  </p>
                  <div className="flex items-center gap-2">
                    <span className="material-symbols-outlined text-[14px] text-on-surface-variant">local_shipping</span>
                    <p className="text-[9px] font-black text-on-surface-variant uppercase tracking-widest opacity-40">
                      Asset ID: <span className="text-white font-mono">{alert.vehicle_id?.substring(0, 8).toUpperCase()}</span>
                    </p>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Center: Map */}
        <div className="flex-1 relative">
          <LiveMap
            height="100%"
            vehicles={Object.values(vehicles)}
            onVehicleSelect={(v) => setSelectedVehicleId(v.id)}
          />
        </div>

        {/* Right: Asset Intelligence HUD */}
        {selectedVehicle && (
          <div className="w-[450px] border-l border-outline bg-surface-container border-outline shadow-2xl z-20 flex flex-col overflow-hidden animate-in slide-in-from-right duration-500">
            <div className="p-10 border-b border-outline bg-surface-dim">
                <div className="flex justify-between items-start">
                    <div>
                        <nav className="text-[9px] font-black uppercase tracking-[0.3em] text-on-surface-variant mb-4 flex items-center gap-2">
                            <span className="material-symbols-outlined text-sm">satellite_alt</span>
                            Asset Telemetry Stream
                        </nav>
                        <h2 className="text-3xl font-bold text-white tracking-tighter uppercase">{selectedVehicle.registration_number || "UNIDENTIFIED"}</h2>
                    </div>
                    <button
                        onClick={() => setSelectedVehicleId(null)}
                        className="h-10 w-10 flex items-center justify-center rounded-xl bg-surface-container-highest border border-outline text-white hover:bg-rose-500 transition-all active:scale-95 shadow-lg"
                    >
                        <span className="material-symbols-outlined">close</span>
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-10 space-y-10 custom-scrollbar">
              <DetailItem
                icon="radar"
                label={uiStrings?.panel_status || "Logistics Status"}
                value={
                  (selectedVehicle.current_location?.speed || 0) > 5
                    ? uiStrings?.stat_moving || "In Motion"
                    : uiStrings?.stat_idle || "Static"
                }
                color={
                  (selectedVehicle.current_location?.speed || 0) > 5
                    ? "text-emerald-500"
                    : "text-amber-500"
                }
              />
              <DetailItem
                icon="speed"
                label={uiStrings?.panel_speed || "Current Velocity"}
                value={`${Math.round(selectedVehicle.current_location?.speed || 0)} KM/H`}
              />
              <DetailItem
                icon="report"
                label={uiStrings?.panel_alerts || "Interference Alerts"}
                value={alerts
                  .filter((a) => a.vehicle_id === selectedVehicle.id)
                  .length.toString()}
                color={alerts.filter((a) => a.vehicle_id === selectedVehicle.id).length > 0 ? "text-rose-500" : "text-white"}
              />

              <div className="pt-8 border-t border-outline">
                <h3 className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em] mb-6">
                  {uiStrings?.label_dispatcher || "Operational Directives"}
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  <ActionButton
                    icon="chat"
                    label={uiStrings?.btn_message || "Comms"}
                  />
                  <ActionButton
                    icon="route"
                    label={uiStrings?.btn_reroute || "Tactical Reroute"}
                  />
                  <ActionButton
                    icon="warning"
                    label={uiStrings?.btn_incident || "Log Incident"}
                    variant="danger"
                  />
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// Subcomponents
const StatCard = ({
  label,
  value,
  color,
}: {
  label: string;
  value: number;
  color: "primary" | "emerald" | "amber";
}) => {
  const colors = {
    primary: "text-primary bg-primary/10 border-primary/20",
    emerald: "text-emerald-500 bg-emerald-500/10 border-emerald-500/20",
    amber: "text-amber-500 bg-amber-500/10 border-amber-500/20",
  };
  return (
    <div
      className={`px-6 py-3 rounded-2xl border flex items-center gap-4 shadow-lg ${colors[color]}`}
    >
      <span className="text-[10px] font-black uppercase tracking-widest opacity-60">{label}</span>
      <span className="text-xl font-bold tracking-tighter font-mono">{value}</span>
    </div>
  );
};

const DetailItem = ({
  icon,
  label,
  value,
  color = "text-white",
}: {
  icon: string;
  label: string;
  value: string;
  color?: string;
}) => (
  <div className="flex items-center gap-6 bg-surface-dim border border-outline p-6 rounded-[1.5rem] shadow-inner">
    <div className="w-12 h-12 rounded-2xl bg-surface-container-highest flex items-center justify-center text-primary shadow-lg">
      <span className="material-symbols-outlined text-2xl">{icon}</span>
    </div>
    <div>
      <p className="text-[9px] text-on-surface-variant uppercase font-black tracking-[0.2em] mb-1">
        {label}
      </p>
      <p className={`text-sm font-bold tracking-tight uppercase ${color}`}>{value}</p>
    </div>
  </div>
);

const ActionButton = ({
  icon,
  label,
  variant = "standard",
}: {
  icon: string;
  label: string;
  variant?: "standard" | "danger";
}) => (
  <button
    className={`flex flex-col items-center justify-center gap-3 p-6 rounded-2xl border border-outline transition-all active:scale-95 shadow-lg group ${
        variant === 'danger' ? 'hover:bg-rose-500/20 hover:border-rose-500/40' : 'hover:bg-primary/20 hover:border-primary/40'
    }`}
  >
    <span className={`material-symbols-outlined text-xl transition-transform group-hover:scale-110 ${variant === 'danger' ? 'text-rose-500' : 'text-primary'}`}>{icon}</span>
    <span className="text-[9px] font-black text-white uppercase tracking-widest text-center">
      {label}
    </span>
  </button>
);

export default CommandCenter;
