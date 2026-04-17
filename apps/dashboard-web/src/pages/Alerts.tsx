import React, { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useUIStrings } from "../hooks/useUIStrings";
import { AppDispatch, RootState } from "../store";
import { acknowledgeAlert, Alert, fetchAlerts } from "../store/alertsSlice";

const Alerts: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { alerts, loading } = useSelector((state: RootState) => state.alerts);
  const { uiStrings } = useUIStrings("alerts");

  useEffect(() => {
    dispatch(fetchAlerts({}) as any);
  }, [dispatch]);

  const handleAcknowledge = (id: string) => {
    dispatch(acknowledgeAlert(id));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-surface-dim">
        <div className="relative">
          <div className="w-16 h-16 rounded-full border-4 border-primary/20 border-t-primary animate-spin" />
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="material-symbols-outlined text-primary text-xl animate-pulse">notifications_active</span>
          </div>
        </div>
      </div>
    );
  }

  const getSeverityStyle = (severity: string) => {
    switch (severity) {
      case "critical":
        return "bg-rose-500/10 text-rose-400 border-rose-500/20";
      case "warning":
        return "bg-amber-500/10 text-amber-400 border-amber-500/20";
      default:
        return "bg-primary/10 text-primary border-primary/20";
    }
  };

  return (
    <div className="p-8 space-y-10 bg-surface-dim min-h-screen font-sans antialiased">
      {/* Incident Management Header */}
      <div className="relative bg-surface-container/60 backdrop-blur-3xl border border-outline rounded-[3rem] p-10 shadow-2xl overflow-hidden group">
        <div className="absolute top-0 right-0 p-12 opacity-[0.03] group-hover:opacity-[0.05] transition-opacity">
          <span className="material-symbols-outlined text-[180px] rotate-12">warning</span>
        </div>
        
        <div className="relative flex justify-between items-center">
          <div className="space-y-4">
            <nav className="flex items-center gap-3 text-[10px] font-black uppercase tracking-[0.3em] text-on-surface-variant">
              <span className="h-1.5 w-1.5 rounded-full bg-rose-500 animate-pulse" />
              Live Incident Triage Center
            </nav>
            <div>
              <h1 className="text-5xl font-black tracking-tighter uppercase text-white mb-2">
                {uiStrings?.title_system_alerts || "Incident Registry"}
              </h1>
              <p className="text-on-surface-variant text-sm font-medium tracking-wide">
                {uiStrings?.sub_title || "Monitoring real-time telemetry anomalies and mission-critical hardware failures"}
              </p>
            </div>
          </div>
          <div className="h-14 px-8 bg-surface-container-highest border border-outline rounded-2xl flex items-center gap-4 text-[10px] font-black uppercase tracking-widest text-white shadow-lg">
            <span className="material-symbols-outlined text-primary">sync</span>
            Signal Link: Active
          </div>
        </div>
      </div>

      <div className="space-y-6 max-w-6xl mx-auto">
        {alerts.length === 0 ? (
          <div className="bg-surface-container rounded-[2rem] border border-outline p-20 text-center shadow-2xl">
            <span className="material-symbols-outlined text-6xl text-on-surface-variant opacity-20 mb-6 flex justify-center">shield_check</span>
            <p className="text-[12px] font-black uppercase tracking-[0.4em] text-on-surface-variant opacity-40">
              {uiStrings?.msg_no_alerts || "Fleet Shield Integrity: 100% (No Anomalies)"}
            </p>
          </div>
        ) : (
          alerts.map((alert: Alert) => (
            <div
              key={alert.id}
              className={`bg-surface-container rounded-[2rem] border border-outline p-8 group hover:bg-surface-container-highest transition-all duration-500 shadow-xl relative overflow-hidden ${
                !alert.acknowledged ? "border-l-4 border-l-primary" : ""
              }`}
            >
              <div className="relative flex items-start justify-between">
                <div className="flex-1 space-y-4">
                  <div className="flex items-center gap-4">
                    <span className={`px-4 py-1 text-[10px] font-black uppercase tracking-widest rounded-lg border ${getSeverityStyle(alert.severity)}`}>
                      {alert.severity}
                    </span>
                    <span className="text-xl font-bold text-white tracking-tight uppercase">
                      {alert.type}
                    </span>
                  </div>
                  <p className="text-on-surface-variant text-lg font-medium leading-relaxed max-w-3xl">
                    {alert.message}
                  </p>
                  <div className="flex items-center gap-6">
                    <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-on-surface-variant opacity-50">
                        <span className="material-symbols-outlined text-sm">schedule</span>
                        {new Date(alert.timestamp).toLocaleString()}
                    </div>
                    {alert.acknowledged && (
                        <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-emerald-500">
                            <span className="material-symbols-outlined text-sm">check_circle</span>
                            Neutralized
                        </div>
                    )}
                  </div>
                </div>
                {!alert.acknowledged && (
                  <button
                    onClick={() => handleAcknowledge(alert.id)}
                    className="h-14 px-8 bg-primary hover:bg-primary-dark text-white text-[10px] font-black uppercase tracking-widest rounded-2xl shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-3"
                  >
                    <span className="material-symbols-outlined text-lg">done_all</span>
                    {uiStrings?.btn_acknowledge || "Acknowledge"}
                  </button>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Alerts;
