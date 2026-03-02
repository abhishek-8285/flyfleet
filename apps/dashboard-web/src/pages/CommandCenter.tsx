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
import { Alert, Vehicle } from "../types";

const CommandCenter: React.FC = () => {
  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = `${wsProtocol}://${window.location.host}/api/v1/ws`;
  const { lastMessage, isConnected } = useWebSocket(wsUrl);
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
    <div className="flex flex-col h-[calc(100vh-64px)] overflow-hidden bg-gray-50 dark:bg-gray-900">
      {/* Top Header / Stats */}
      <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-4 shrink-0 shadow-sm z-10">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <MapIcon className="w-6 h-6 text-blue-600" />
            {uiStrings?.title || "Fleet Command Center"}
            {!isConnected && (
              <span className="flex items-center gap-1 text-xs font-normal text-red-500 bg-red-50 dark:bg-red-900/20 px-2 py-0.5 rounded-full animate-pulse">
                <ArrowPathIcon className="w-3 h-3 animate-spin" />{" "}
                {uiStrings?.status_disconnected || "Disconnected"}
              </span>
            )}
          </h1>

          <div className="flex gap-4">
            <StatCard
              label={uiStrings?.stat_total || "Total"}
              value={vehicleStats.total}
              color="blue"
            />
            <StatCard
              label={uiStrings?.stat_moving || "Moving"}
              value={vehicleStats.moving}
              color="green"
            />
            <StatCard
              label={uiStrings?.stat_idle || "Idle"}
              value={vehicleStats.idle}
              color="yellow"
            />
          </div>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* Left: Alert Feed */}
        <div className="w-80 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col shrink-0">
          <div className="p-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between bg-gray-50 dark:bg-gray-800/50">
            <div className="flex items-center gap-2 font-semibold dark:text-white">
              <BellIcon className="w-5 h-5 text-gray-500" />
              {uiStrings?.feed_title || "Live Alerts"}
            </div>
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value as any)}
              className="text-xs border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded py-1"
            >
              <option value="all">{uiStrings?.filter_all || "All"}</option>
              <option value="critical">
                {uiStrings?.filter_critical || "Critical"}
              </option>
              <option value="warning">
                {uiStrings?.filter_warning || "Warning"}
              </option>
            </select>
          </div>
          <div className="flex-1 overflow-y-auto">
            {filteredAlerts.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">
                {uiStrings?.no_alerts || "No active alerts"}
              </div>
            ) : (
              filteredAlerts.map((alert) => (
                <div
                  key={alert.id}
                  className="p-4 border-b border-gray-100 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors cursor-pointer"
                >
                  <div className="flex justify-between items-start mb-1">
                    <span
                      className={`text-[10px] font-bold uppercase px-1.5 py-0.5 rounded ${
                        alert.severity === "critical"
                          ? "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"
                          : alert.severity === "warning"
                            ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400"
                            : "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
                      }`}
                    >
                      {alert.severity}
                    </span>
                    <span className="text-[10px] text-gray-400">
                      {format(new Date(alert.timestamp), "HH:mm:ss")}
                    </span>
                  </div>
                  <p className="text-xs font-medium text-gray-900 dark:text-gray-200">
                    {alert.message}
                  </p>
                  <p className="text-[10px] text-gray-500 mt-1">
                    {uiStrings?.th_vehicle || "Vehicle"}:{" "}
                    {alert.vehicle_id?.substring(0, 8)}
                  </p>
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

        {/* Right: Info Panel */}
        {selectedVehicle && (
          <div className="w-80 border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-6 overflow-y-auto shrink-0 animate-in slide-in-from-right duration-300">
            <div className="flex justify-between items-start mb-6">
              <div>
                <h2 className="text-lg font-bold dark:text-white">
                  {selectedVehicle.registration_number || "---"}
                </h2>
                <span className="text-xs text-gray-500">
                  {selectedVehicle.id}
                </span>
              </div>
              <button
                onClick={() => setSelectedVehicleId(null)}
                className="text-gray-400 hover:text-gray-600"
              >
                &times;
              </button>
            </div>

            <div className="space-y-6">
              <DetailItem
                icon={<TruckIcon className="w-4 h-4" />}
                label={uiStrings?.panel_status || "Status"}
                value={
                  (selectedVehicle.current_location?.speed || 0) > 5
                    ? uiStrings?.stat_moving || "Moving"
                    : uiStrings?.stat_idle || "Idle"
                }
                color={
                  (selectedVehicle.current_location?.speed || 0) > 5
                    ? "text-green-500"
                    : "text-yellow-500"
                }
              />
              <DetailItem
                icon={<ArrowPathIcon className="w-4 h-4" />}
                label={uiStrings?.panel_speed || "Speed"}
                value={`${Math.round(selectedVehicle.current_location?.speed || 0)} km/h`}
              />
              <DetailItem
                icon={<ExclamationTriangleIcon className="w-4 h-4" />}
                label={uiStrings?.panel_alerts || "Open Alerts"}
                value={alerts
                  .filter((a) => a.vehicle_id === selectedVehicle.id)
                  .length.toString()}
              />
            </div>

            <div className="mt-8 pt-8 border-t border-gray-100 dark:border-gray-700">
              <h3 className="text-sm font-semibold mb-4 dark:text-white text-gray-900">
                {uiStrings?.label_dispatcher || "Dispatcher Actions"}
              </h3>
              <div className="grid grid-cols-2 gap-2">
                <ActionButton
                  icon={<ChatBubbleLeftRightIcon className="w-4 h-4" />}
                  label={uiStrings?.btn_message || "Message"}
                />
                <ActionButton
                  icon={<ArrowPathIcon className="w-4 h-4" />}
                  label={uiStrings?.btn_reroute || "Reroute"}
                />
                <ActionButton
                  icon={
                    <ExclamationTriangleIcon className="w-4 h-4 text-red-500" />
                  }
                  label={uiStrings?.btn_incident || "Log Incident"}
                  color="hover:bg-red-50 dark:hover:bg-red-900/20"
                />
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
  color: "blue" | "green" | "yellow";
}) => {
  const colors = {
    blue: "text-blue-600 bg-blue-50 dark:bg-blue-900/20",
    green: "text-green-600 bg-green-50 dark:bg-green-900/20",
    yellow: "text-yellow-600 bg-yellow-50 dark:bg-yellow-900/20",
  };
  return (
    <div
      className={`px-4 py-1.5 rounded-lg border border-transparent dark:border-gray-700 shadow-sm flex items-center gap-3 ${colors[color]}`}
    >
      <span className="text-xs font-medium uppercase opacity-80">{label}</span>
      <span className="text-lg font-bold">{value}</span>
    </div>
  );
};

const DetailItem = ({
  icon,
  label,
  value,
  color = "text-gray-900 dark:text-gray-200",
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  color?: string;
}) => (
  <div className="flex items-center gap-3">
    <div className="w-8 h-8 rounded bg-gray-50 dark:bg-gray-700/50 flex items-center justify-center text-gray-400">
      {icon}
    </div>
    <div>
      <p className="text-[10px] text-gray-500 uppercase font-bold tracking-wider">
        {label}
      </p>
      <p className={`text-sm font-semibold ${color}`}>{value}</p>
    </div>
  </div>
);

const ActionButton = ({
  icon,
  label,
  color = "hover:bg-gray-50 dark:hover:bg-gray-700/50",
}: {
  icon: React.ReactNode;
  label: string;
  color?: string;
}) => (
  <button
    className={`flex flex-col items-center justify-center gap-2 p-3 rounded-lg border border-gray-100 dark:border-gray-700 transition-all ${color}`}
  >
    {icon}
    <span className="text-[10px] font-bold text-gray-700 dark:text-gray-300">
      {label}
    </span>
  </button>
);

export default CommandCenter;
