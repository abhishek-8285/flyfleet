import React from "react";
import { Driver, SafetyScore } from "../types";

interface Props {
  driver: Driver;
  score: SafetyScore;
}

/**
 * DriverPerformance: A premium dashboard component for driver safety analytics.
 */
export const DriverPerformance: React.FC<Props> = ({ driver, score }) => {
  return (
    <div className="bg-surface-dark rounded-2xl border border-border-color overflow-hidden shadow-xl">
      {/* Header */}
      <div className="p-6 bg-gradient-to-r from-surface-darker to-surface-dark border-b border-border-color">
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-primary/10 border-2 border-primary/20 flex items-center justify-center text-primary text-2xl font-bold">
              {driver.name
                .split(" ")
                .map((n) => n[0])
                .join("")}
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white tracking-tight">
                {driver.name}
              </h1>
              <p className="text-text-muted text-sm font-medium">
                License: {driver.licenseNumber}
              </p>
            </div>
          </div>
          <div className="text-right">
            <div
              className={`text-4xl font-black ${
                score.score >= 90
                  ? "text-emerald-400"
                  : score.score >= 80
                    ? "text-yellow-400"
                    : "text-red-400"
              }`}
            >
              {score.score.toFixed(1)}
            </div>
            <div className="text-[10px] uppercase font-bold tracking-widest text-text-muted mt-1">
              Safety Index
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Score Breakdown */}
        <div className="col-span-1 space-y-4">
          <h3 className="text-white font-bold text-sm uppercase tracking-wider mb-2">
            Event Breakdown
          </h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center p-3 bg-surface-darker rounded-xl border border-border-color/50">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-emerald-400">
                  speed
                </span>
                <span className="text-sm font-medium">Overspeeding</span>
              </div>
              <span className="font-bold">{score.events.overspeeding}</span>
            </div>
            <div className="flex justify-between items-center p-3 bg-surface-darker rounded-xl border border-border-color/50">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-yellow-400">
                  warning
                </span>
                <span className="text-sm font-medium">Harsh Braking</span>
              </div>
              <span className="font-bold">{score.events.harshBraking}</span>
            </div>
            <div className="flex justify-between items-center p-3 bg-surface-darker rounded-xl border border-border-color/50">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-red-400">
                  dangerous
                </span>
                <span className="text-sm font-medium">Harsh Turning</span>
              </div>
              <span className="font-bold">{score.events.harshTurning}</span>
            </div>
          </div>
        </div>

        {/* Real-time Status */}
        <div className="col-span-2 space-y-4">
          <h3 className="text-white font-bold text-sm uppercase tracking-wider mb-2">
            Current Session
          </h3>
          <div className="bg-surface-darker rounded-2xl p-6 border border-border-color">
            <div className="grid grid-cols-2 gap-8">
              <div>
                <div className="text-text-muted text-xs uppercase mb-1">
                  Status
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
                  <span className="text-white font-bold">
                    {driver.status.replace("_", " ")}
                  </span>
                </div>
              </div>
              <div>
                <div className="text-text-muted text-xs uppercase mb-1">
                  Active Vehicle
                </div>
                <div className="text-white font-bold">
                  {driver.activeVehicleId || "Not Assigned"}
                </div>
              </div>
              <div>
                <div className="text-text-muted text-xs uppercase mb-1">
                  Shift Duration
                </div>
                <div className="text-white font-bold">04:32:15</div>
              </div>
              <div>
                <div className="text-text-muted text-xs uppercase mb-1">
                  Rating
                </div>
                <div className="flex items-center gap-1 text-yellow-400 font-bold">
                  <span className="material-symbols-outlined text-sm">
                    star
                  </span>
                  {driver.rating}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
