import React, { useEffect, useState } from "react";

export interface Alert {
  id: string;
  vehicleId: string;
  type: "CRITICAL" | "WARNING" | "INFO";
  message: string;
  timestamp: string;
  location?: { lat: number; lng: number };
}

/**
 * LiveAlertsFeed: Real-time notification stream for dispatcher attention.
 */
export const LiveAlertsFeed: React.FC = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);

  useEffect(() => {
    // Simulated alert stream
    const interval = setInterval(() => {
      const newAlert: Alert = {
        id: `alert_${Date.now()}`,
        vehicleId: `V-${Math.floor(Math.random() * 1000)}`,
        type: Math.random() > 0.8 ? "CRITICAL" : "WARNING",
        message:
          Math.random() > 0.5
            ? "Geofence Breach: Cluster B"
            : "Harsh Braking Detected",
        timestamp: new Date().toISOString(),
      };
      setAlerts((prev) => [newAlert, ...prev].slice(0, 50));
    }, 15000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex flex-col h-full bg-surface-darker/50 backdrop-blur-md border-l border-border-color">
      <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-dark">
        <h3 className="text-sm font-bold uppercase tracking-wider text-white flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-red-500 animate-ping"></span>
          Live Incident Feed
        </h3>
        <span className="text-[10px] font-mono text-text-muted">
          {alerts.length} Active
        </span>
      </div>

      <div className="flex-1 overflow-auto p-2 space-y-2 custom-scrollbar">
        {alerts.map((alert) => (
          <div
            key={alert.id}
            className={`p-3 rounded-lg border transition-all cursor-pointer hover:scale-[1.02] ${
              alert.type === "CRITICAL"
                ? "bg-red-500/10 border-red-500/20"
                : "bg-yellow-500/10 border-yellow-500/20"
            }`}
          >
            <div className="flex justify-between items-start mb-1">
              <span
                className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${
                  alert.type === "CRITICAL"
                    ? "bg-red-500 text-white"
                    : "bg-yellow-500 text-black"
                }`}
              >
                {alert.type}
              </span>
              <span className="text-[9px] text-text-muted font-mono">
                {new Date(alert.timestamp).toLocaleTimeString()}
              </span>
            </div>
            <p className="text-xs font-bold text-white mb-1">{alert.message}</p>
            <p className="text-[10px] text-text-muted font-mono">
              Vehicle: {alert.vehicleId}
            </p>
          </div>
        ))}

        {alerts.length === 0 && (
          <div className="h-full flex flex-col items-center justify-center opacity-20 py-20">
            <span className="material-symbols-outlined text-4xl mb-2">
              notifications_off
            </span>
            <p className="text-xs">Watching for incidents...</p>
          </div>
        )}
      </div>
    </div>
  );
};
