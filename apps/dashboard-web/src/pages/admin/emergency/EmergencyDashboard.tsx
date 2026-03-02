import {
    ChatBubbleLeftRightIcon,
    CheckCircleIcon,
    ExclamationTriangleIcon,
    MapPinIcon,
    PhoneIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";
import { useWebSocket } from "../../../hooks/useWebSocket";
import apiClient from "../../../services/apiClient";

interface SOSIncident {
  id: string;
  vehicle_id: string;
  vehicle_reg_number: string;
  driver_id?: string;
  driver_name?: string;
  timestamp: string;
  latitude: number;
  longitude: number;
  location_address: string;
  status: "TRIGGERED" | "ACKNOWLEDGED" | "RESOLVED" | "FALSE_ALARM";
  severity: "LOW" | "MEDIUM" | "CRITICAL";
  metadata: any;
}

const EmergencyDashboard: React.FC = () => {
  const [incidents, setIncidents] = useState<SOSIncident[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("emergency_admin");
  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = `${wsProtocol}://${window.location.host}/api/v1/ws`;
  const { lastMessage, isConnected } = useWebSocket(wsUrl);

  useEffect(() => {
    const fetchIncidents = async () => {
      try {
        const response = (await apiClient.get(
          "/admin/emergency/incidents",
        )) as any;
        setIncidents(response?.data || []);
      } catch (error) {
        console.error("Failed to fetch incidents", error);
        setIncidents([]);
      } finally {
        setLoading(false);
      }
    };

    fetchIncidents();
  }, []);

  useEffect(() => {
    if (!lastMessage) return;
    if (!lastMessage.subject?.startsWith("emergency.sos.")) return;

    const payload = lastMessage.data || {};
    const normalized: SOSIncident = {
      id: payload.id,
      vehicle_id: payload.vehicle_id,
      vehicle_reg_number: payload.vehicle_reg_number || payload.vehicle_id,
      driver_id: payload.driver_id,
      driver_name: payload.driver_name,
      timestamp: payload.timestamp || new Date().toISOString(),
      latitude: Number(payload.latitude || 0),
      longitude: Number(payload.longitude || 0),
      location_address: payload.location_address || "Unknown location",
      status: payload.status || "TRIGGERED",
      severity: payload.severity || "MEDIUM",
      metadata: payload.metadata || {},
    };

    setIncidents((prev) => {
      const idx = prev.findIndex((i) => i.id === normalized.id);
      if (idx >= 0) {
        const next = [...prev];
        // DYNAMIC FETCH HOOK PLACEHOLDER
        next[idx] = { ...next[idx], ...normalized };
        return next;
      }
      return [normalized, ...prev];
    });
  }, [lastMessage]);

  const updateIncidentStatus = async (
    incidentId: string,
    action: "acknowledge" | "resolve",
  ) => {
    try {
      await apiClient.post(
        `/admin/emergency/incidents/${incidentId}/${action}`,
        {
          status: action === "resolve" ? "RESOLVED" : "ACKNOWLEDGED",
          notes: `Updated from emergency dashboard`,
        },
      );
      setIncidents((prev) =>
        prev.map((i) =>
          i.id === incidentId
            ? {
                ...i,
                status: action === "resolve" ? "RESOLVED" : "ACKNOWLEDGED",
              }
            : i,
        ),
      );
    } catch (error) {
      console.error(`Failed to ${action} incident`, error);
    }
  };

  const callEmergency = () => {
    window.location.href = "tel:112";
  };

  const openDispatchChat = (incident: SOSIncident) => {
    const text = encodeURIComponent(
      `Emergency incident ${incident.vehicle_reg_number} (${incident.status}) at ${incident.latitude},${incident.longitude}`,
    );
    window.open(
      `mailto:dispatch@flyfleet.com?subject=Emergency Incident&body=${text}`,
      "_blank",
    );
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "CRITICAL":
        return "text-red-500 bg-red-500/10 border-red-500/20";
      case "MEDIUM":
        return "text-orange-500 bg-orange-500/10 border-orange-500/20";
      case "LOW":
        return "text-yellow-500 bg-yellow-500/10 border-yellow-500/20";
      default:
        return "text-gray-400 bg-gray-500/10 border-gray-500/20";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "TRIGGERED":
        return "text-red-400 animate-pulse";
      case "ACKNOWLEDGED":
        return "text-yellow-400";
      case "RESOLVED":
        return "text-green-400";
      case "FALSE_ALARM":
        return "text-gray-400";
      default:
        return "text-white";
    }
  };

  return (
    <div className="space-y-6 p-6 bg-slate-900 min-h-screen">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <ExclamationTriangleIcon className="h-8 w-8 text-red-500" />
            {uiStrings?.title || "Emergency Response Center"}
          </h1>
          <p className="text-slate-400">
            {uiStrings?.subtitle ||
              "Real-time monitoring of active SOS incidents"}
          </p>
          <p
            className={`text-xs mt-1 ${isConnected ? "text-emerald-400" : "text-amber-400"}`}
          >
            WebSocket {isConnected ? "connected" : "reconnecting"}
          </p>
        </div>
        <div className="flex gap-3">
          <button className="px-4 py-2 bg-slate-800 text-white rounded-lg border border-slate-700 hover:bg-slate-700 transition">
            {uiStrings?.btn_manage_contacts || "Manage Contacts"}
          </button>
          <button className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition shadow-lg shadow-red-900/20">
            {uiStrings?.btn_broadcast || "Broadcast Alert"}
          </button>
        </div>
      </div>

      {/* Active Incidents Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {incidents.map((incident) => (
          <div
            key={incident.id}
            className={`bg-slate-800 rounded-xl border border-slate-700 overflow-hidden shadow-lg ${incident.status === "TRIGGERED" ? "ring-2 ring-red-500 ring-opacity-50" : ""}`}
          >
            <div className="p-6">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <div
                    className={`px-3 py-1 rounded-full text-xs font-bold w-fit mb-2 border ${getSeverityColor(incident.severity)}`}
                  >
                    {incident.severity} SEVERITY
                  </div>
                  <h3 className="text-xl font-bold text-white">
                    {incident.vehicle_reg_number}
                  </h3>
                  <p className="text-slate-400 text-sm">
                    Driver: {incident.driver_name || "Unknown"}
                  </p>
                </div>
                <div
                  className={`text-sm font-semibold flex items-center gap-1 ${getStatusColor(incident.status)}`}
                >
                  <span className="h-2 w-2 rounded-full bg-current"></span>
                  {incident.status}
                </div>
              </div>

              <div className="space-y-3 mb-6">
                <div className="flex items-start gap-3 text-slate-300">
                  <MapPinIcon className="h-5 w-5 text-slate-500 mt-0.5" />
                  <div>
                    <p className="text-sm">{incident.location_address}</p>
                    <p className="text-xs text-slate-500 mt-0.5">
                      {incident.latitude.toFixed(4)},{" "}
                      {incident.longitude.toFixed(4)}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3 text-slate-300">
                  <div className="h-5 w-5 flex items-center justify-center">
                    <span className="text-xs font-mono text-slate-500">🕒</span>
                  </div>
                  <p className="text-sm">
                    {format(new Date(incident.timestamp), "PPpp")}
                  </p>
                </div>
              </div>

              <div className="flex gap-3 pt-4 border-t border-slate-700">
                {incident.status === "TRIGGERED" && (
                  <button
                    onClick={() =>
                      updateIncidentStatus(incident.id, "acknowledge")
                    }
                    className="flex-1 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-lg text-sm font-medium transition flex items-center justify-center gap-2"
                  >
                    <CheckCircleIcon className="h-4 w-4" /> Acknowledge
                  </button>
                )}
                {(incident.status === "TRIGGERED" ||
                  incident.status === "ACKNOWLEDGED") && (
                  <button
                    onClick={() => updateIncidentStatus(incident.id, "resolve")}
                    className="flex-1 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg text-sm font-medium transition flex items-center justify-center gap-2"
                  >
                    <CheckCircleIcon className="h-4 w-4" /> Resolve
                  </button>
                )}
                <button
                  onClick={callEmergency}
                  className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg text-sm font-medium transition"
                >
                  <PhoneIcon className="h-4 w-4" />
                </button>
                <button
                  onClick={() => openDispatchChat(incident)}
                  className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg text-sm font-medium transition"
                >
                  <ChatBubbleLeftRightIcon className="h-4 w-4" />
                </button>
              </div>
            </div>
            {/* Mini Map Placeholder */}
            <div className="h-32 bg-slate-900 w-full relative group cursor-pointer overflow-hidden">
              <div className="absolute inset-0 flex items-center justify-center text-slate-600 text-sm font-medium">
                View on Live Map
              </div>
              <div className="absolute bottom-2 right-2 text-[10px] text-slate-400 bg-slate-900/70 px-2 py-1 rounded">
                {incident.latitude.toFixed(4)}, {incident.longitude.toFixed(4)}
              </div>
            </div>
          </div>
        ))}

        {incidents.length === 0 && !loading && (
          <div className="col-span-full py-20 text-center text-slate-500">
            <CheckCircleIcon className="h-16 w-16 mx-auto mb-4 text-emerald-500/20" />
            <h3 className="text-lg font-medium text-white">
              {uiStrings?.msg_normal || "All Systems Normal"}
            </h3>
            <p>
              {uiStrings?.msg_no_incidents ||
                "No active emergency incidents reported."}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default EmergencyDashboard;
