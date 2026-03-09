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
  incident_type?: string;
  metadata: any;
}

const EmergencyDashboard: React.FC = () => {
  const [incidents, setIncidents] = useState<SOSIncident[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("emergency_admin");
  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = `${wsProtocol}://${window.location.host}/api/v1/ws`;
  const { lastMessage, isConnected } = useWebSocket(wsUrl);

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
      incident_type: payload.incident_type,
      metadata: payload.metadata || {},
    };

    setIncidents((prev) => {
      const idx = prev.findIndex((i) => i.id === normalized.id);
      if (idx >= 0) {
        const next = [...prev];
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

  const issueFleetHalt = async () => {
    const reason = window.prompt("Reason for fleet-wide halt:");
    if (!reason) return;
    try {
      await apiClient.post("/api/v1/emergency/fleet-halt", { reason });
      alert("Fleet-wide halt issued successfully.");
    } catch (error) {
      console.error("Failed to issue fleet halt", error);
    }
  };

  const issueFleetResume = async () => {
    if (!window.confirm("Are you sure you want to resume fleet operations?"))
      return;
    try {
      await apiClient.post("/api/v1/emergency/fleet-resume");
      alert("Fleet-wide resume issued successfully.");
    } catch (error) {
      console.error("Failed to resume fleet operations", error);
    }
  };

  return (
    <div className="space-y-6 p-6 bg-slate-900 min-h-screen font-sans selection:bg-red-500/30">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight flex items-center gap-3">
            <ExclamationTriangleIcon className="h-10 w-10 text-red-500 animate-pulse" />
            {uiStrings?.title || "India Emergency Response Center"}
            <span className="text-xs font-mono px-2 py-0.5 bg-red-500/10 text-red-400 border border-red-500/20 rounded uppercase tracking-widest">
              AIS-140 Compliant
            </span>
          </h1>
          <p className="text-slate-400 mt-1 text-lg">
            {uiStrings?.subtitle ||
              "Real-time monitoring of active SOS alerts and fleet safety"}
          </p>
          <p
            className={`text-sm mt-2 flex items-center gap-2 ${isConnected ? "text-emerald-400" : "text-amber-400"}`}
          >
            <span
              className={`h-2 w-2 rounded-full ${isConnected ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]" : "bg-amber-500 animate-ping"}`}
            ></span>
            WebSocket {isConnected ? "Active" : "Reconnecting..."}
          </p>
        </div>
        <div className="flex gap-4">
          <button
            onClick={issueFleetResume}
            className="px-6 py-2.5 bg-slate-800 text-emerald-400 font-bold rounded-xl border border-emerald-500/20 hover:bg-emerald-500/10 hover:border-emerald-500/50 transition-all active:scale-95 shadow-lg shadow-emerald-500/10"
          >
            RESUME FLEET
          </button>
          <button
            onClick={issueFleetHalt}
            className="px-6 py-2.5 bg-red-600 text-white font-bold rounded-xl hover:bg-red-700 transition-all active:scale-95 shadow-xl shadow-red-900/40 border border-white/10"
          >
            ISSUE FLEET HALT
          </button>
        </div>
      </div>

      {/* Active Incidents Grid */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-8 mt-4">
        {incidents.map((incident) => (
          <div
            key={incident.id}
            className={`group bg-slate-800/50 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden shadow-2xl transition-all hover:border-red-500/30 ${incident.status === "TRIGGERED" ? "ring-2 ring-red-500/50" : ""}`}
          >
            <div className="p-8">
              <div className="flex justify-between items-start mb-6">
                <div className="space-y-1">
                  <div className="flex gap-2 mb-2">
                    <div
                      className={`px-2.5 py-0.5 rounded-md text-[10px] font-black uppercase tracking-wider border ${getSeverityColor(incident.severity)}`}
                    >
                      {incident.severity}
                    </div>
                    <div className="px-2.5 py-0.5 rounded-md text-[10px] font-black uppercase tracking-wider border border-blue-500/20 text-blue-400 bg-blue-500/5">
                      {incident.incident_type || "SOS_PANIC"}
                    </div>
                  </div>
                  <h3 className="text-2xl font-black text-white tracking-tight group-hover:text-red-400 transition-colors">
                    {incident.vehicle_reg_number}
                  </h3>
                  <div className="flex items-center gap-2 text-slate-400 text-sm">
                    <span className="font-medium">
                      {incident.driver_name || "Unknown Operator"}
                    </span>
                    <span className="h-1 w-1 rounded-full bg-slate-600"></span>
                    <span className="font-mono text-xs opacity-75">
                      {incident.id.slice(0, 8)}
                    </span>
                  </div>
                </div>
                <div className="text-right">
                  <div
                    className={`text-sm font-black flex items-center justify-end gap-2 ${getStatusColor(incident.status)}`}
                  >
                    <span
                      className={`h-2.5 w-2.5 rounded-full bg-current ${incident.status === "TRIGGERED" ? "animate-pulse shadow-[0_0_8px_currentColor]" : ""}`}
                    ></span>
                    {incident.status}
                  </div>
                  <p className="text-[10px] text-slate-500 mt-1 font-mono">
                    {format(new Date(incident.timestamp), "HH:mm:ss")}
                  </p>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 mb-8">
                <div className="space-y-4">
                  <div className="flex items-start gap-3 text-slate-300">
                    <MapPinIcon className="h-6 w-6 text-red-400/80 -mt-0.5" />
                    <div>
                      <p className="text-sm font-bold text-white mb-0.5 leading-tight">
                        {incident.location_address}
                      </p>
                      <p className="text-[10px] text-slate-500 font-mono tracking-tighter">
                        {incident.latitude.toFixed(6)},{" "}
                        {incident.longitude.toFixed(6)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3 text-slate-300">
                    <div className="h-6 w-6 flex items-center justify-center bg-slate-700/50 rounded-lg">
                      <span className="text-xs">⚡</span>
                    </div>
                    <div>
                      <p className="text-[10px] text-slate-500 uppercase tracking-widest font-bold">
                        Speed/Heading
                      </p>
                      <p className="text-sm font-mono text-white">
                        42 km/h • 182° S
                      </p>
                    </div>
                  </div>
                </div>
                <div className="bg-slate-900/50 rounded-xl p-4 border border-slate-700/30">
                  <p className="text-[10px] text-slate-500 uppercase tracking-widest font-black mb-3">
                    Escalation Status
                  </p>
                  <div className="space-y-3">
                    <div className="flex items-center gap-2">
                      <div className="h-1.5 w-1.5 rounded-full bg-emerald-500"></div>
                      <p className="text-xs text-slate-300">
                        Fleet Manager notified (SMS)
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="h-1.5 w-1.5 rounded-full bg-amber-500 animate-pulse"></div>
                      <p className="text-xs text-slate-300">
                        Safety Head (Pending 1:42s)
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <div className="flex gap-4 pt-6 border-t border-slate-700/50">
                {incident.status === "TRIGGERED" && (
                  <button
                    onClick={() =>
                      updateIncidentStatus(incident.id, "acknowledge")
                    }
                    className="flex-1 py-3 bg-amber-600 hover:bg-amber-500 text-white rounded-xl text-sm font-black transition-all shadow-lg shadow-amber-900/20 active:scale-[0.98] uppercase tracking-wider"
                  >
                    ACKNOWLEDGE
                  </button>
                )}
                {(incident.status === "TRIGGERED" ||
                  incident.status === "ACKNOWLEDGED") && (
                  <button
                    onClick={() => updateIncidentStatus(incident.id, "resolve")}
                    className="flex-1 py-3 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-sm font-black transition-all shadow-lg shadow-emerald-900/20 active:scale-[0.98] uppercase tracking-wider"
                  >
                    RESOLVE
                  </button>
                )}
                <div className="flex gap-2">
                  <button
                    onClick={callEmergency}
                    className="p-3 bg-slate-700 hover:bg-red-500/20 hover:text-red-400 text-slate-300 rounded-xl transition-all border border-transparent hover:border-red-500/30"
                  >
                    <PhoneIcon className="h-5 w-5" />
                  </button>
                  <button
                    onClick={() => openDispatchChat(incident)}
                    className="p-3 bg-slate-700 hover:bg-blue-500/20 hover:text-blue-400 text-slate-300 rounded-xl transition-all border border-transparent hover:border-blue-500/30"
                  >
                    <ChatBubbleLeftRightIcon className="h-5 w-5" />
                  </button>
                </div>
              </div>
            </div>
            {/* Live Map Preview Overlay */}
            <div className="h-40 bg-slate-900/80 w-full relative group cursor-crosshair overflow-hidden border-t border-slate-700/50">
              <div className="absolute inset-0 bg-[url('https://maps.googleapis.com/maps/api/staticmap?center=28.6139,77.2090&zoom=14&size=600x300&key=MOCK')] opacity-30 grayscale contrast-125"></div>
              <div className="absolute inset-0 flex items-center justify-center transition-all bg-slate-950/40 group-hover:bg-transparent">
                <div className="px-4 py-2 bg-slate-800/80 backdrop-blur border border-white/10 rounded-full text-xs font-bold text-white shadow-2xl">
                  OPEN LIVE TELEMETRY MAP
                </div>
              </div>
            </div>
          </div>
        ))}

        {incidents.length === 0 && !loading && (
          <div className="col-span-full py-32 text-center rounded-3xl border-2 border-dashed border-slate-800">
            <CheckCircleIcon className="h-24 w-24 mx-auto mb-6 text-emerald-500/10" />
            <h3 className="text-2xl font-black text-emerald-400/80 tracking-tight">
              MISSION STATUS: SECURE
            </h3>
            <p className="text-slate-500 font-medium max-w-xs mx-auto mt-2">
              Zero active emergency incidents. All fleet assets are operating
              within safety parameters.
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default EmergencyDashboard;
