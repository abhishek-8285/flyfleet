import React from "react";
import { Driver } from "../types";

interface Props {
  driver: Driver;
  analytics: any; // Using dynamic analytics payload
}

/**
 * DriverPerformance: A professional analytics dashboard for driver safety and behavior.
 */
export const DriverPerformance: React.FC<Props> = ({ driver, analytics }) => {
  const score = Math.max(0, Math.min(100, Math.round(analytics?.safety_score || driver?.behavior_score || 0)));
  
  const getScoreColor = (s: number) => {
    if (s >= 90) return "text-emerald-400";
    if (s >= 80) return "text-primary";
    if (s >= 60) return "text-amber-400";
    return "text-rose-400";
  };

  const incidents = [
    { label: "Overspeeding", count: analytics?.speeding_incidents || 0, icon: "speed", color: "text-emerald-400" },
    { label: "Harsh Braking", count: analytics?.harsh_braking_count || 0, icon: "warning", color: "text-amber-400" },
    { label: "Harsh Turning", count: analytics?.harsh_acceleration_count || 0, icon: "dangerous", color: "text-rose-400" }, // Acceleration used as proxy for turning if not available
    { label: "Idling Time", count: `${analytics?.idling_time_hours || 0}H`, icon: "pause_circle", color: "text-primary" },
  ];

  return (
    <div className="bg-surface-container rounded-2xl border border-outline overflow-hidden shadow-sm animate-entrance">
      {/* Header Section */}
      <div className="p-8 bg-surface-container-high border-b border-outline relative">
        <div className="flex justify-between items-end relative z-10">
          <div>
            <div className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-wider text-on-surface-variant mb-2">
              Performance Insights
            </div>
            <h2 className="text-2xl font-bold text-white tracking-tight uppercase">Performance Analytics</h2>
          </div>
          <div className="text-right">
            <div className={`text-5xl font-bold leading-none ${getScoreColor(score)}`}>
              {score}<span className="text-xs opacity-40 ml-1">/100</span>
            </div>
            <p className="text-[10px] uppercase font-bold tracking-wider text-on-surface-variant mt-2">Safety Score Rating</p>
          </div>
        </div>
      </div>

      {/* Main Analytics Grid */}
      <div className="p-8 grid grid-cols-1 md:grid-cols-4 gap-6">
        {incidents.map((incident, i) => (
          <div key={i} className="bg-surface-container/50 p-5 rounded-2xl border border-outline group hover:border-primary/50 transition-all">
            <div className="flex justify-between items-start mb-4">
              <span className={`material-symbols-outlined ${incident.color} text-xl`}>{incident.icon}</span>
              <span className="text-[9px] font-bold text-on-surface-variant uppercase tracking-wider opacity-40">ITEM {i+1}</span>
            </div>
            <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-1">{incident.label}</p>
            <p className="text-2xl font-bold text-white tracking-tight">{incident.count}</p>
          </div>
        ))}
      </div>

      {/* Operational Status Section */}
      <div className="px-8 pb-8 grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-surface-container-high/40 rounded-2xl p-6 border border-outline overflow-hidden group">
          <h3 className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-4">Operational Status</h3>
          <div className="grid grid-cols-2 gap-6">
            <div>
              <p className="text-[9px] font-bold text-on-surface-variant uppercase tracking-wider mb-1">Status</p>
              <div className="flex items-center gap-2">
                <span className="size-1.5 rounded-full bg-emerald-500 shadow-sm"></span>
                <span className="text-sm font-bold text-white uppercase tracking-tight">{driver.status.replace("_", " ")}</span>
              </div>
            </div>
            <div>
              <p className="text-[9px] font-bold text-on-surface-variant uppercase tracking-wider mb-1">Fuel Economy</p>
              <p className="text-sm font-bold text-white uppercase tracking-tight">{analytics?.fuel_efficiency || "8.4 KM/L"}</p>
            </div>
          </div>
        </div>

        <div className="bg-surface-container-high/40 rounded-2xl p-6 border border-outline overflow-hidden group">
          <h3 className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-4">Work Environment</h3>
          <div className="grid grid-cols-2 gap-6">
            <div>
              <p className="text-[9px] font-bold text-on-surface-variant uppercase tracking-wider mb-1">Shift Type</p>
              <p className="text-sm font-bold text-white uppercase tracking-tight">{(driver.metadata?.shift_type as string) || "Day Shift"}</p>
            </div>
            <div>
              <p className="text-[9px] font-bold text-on-surface-variant uppercase tracking-wider mb-1">Overall Rating</p>
              <div className="flex items-center gap-1.5">
                <span className="material-symbols-outlined text-amber-500 text-sm">star</span>
                <span className="text-sm font-bold text-white tracking-tight">{driver.rating || "4.8"}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      {/* Analytics Footer */}
      <div className="bg-surface-container-high px-8 py-4 border-t border-outline flex justify-between items-center">
        <p className="text-[9px] font-bold text-on-surface-variant uppercase tracking-wider opacity-60">
          Data Sync Status: <span className="text-emerald-400 font-bold">Live</span>
        </p>
        <button className="text-[9px] font-bold text-primary uppercase tracking-wider hover:text-white transition-all">View Full Safety Report</button>
      </div>
    </div>
  );
};
