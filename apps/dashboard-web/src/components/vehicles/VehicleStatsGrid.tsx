import React from "react";

import { useUIStrings } from "../../hooks/useUIStrings";
import { Vehicle } from "../../types";

interface VehicleStatsGridProps {
  vehicle?: Vehicle;
}

export const VehicleStatsGrid: React.FC<VehicleStatsGridProps> = ({
  vehicle,
}) => {
  const { uiStrings } = useUIStrings("comp_vehicles");
  if (!vehicle) return null;

  return (
    <div className="grid grid-cols-4 gap-4">
      <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
        <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">
          {uiStrings?.label_odometer || "Odometer"}
        </div>
        <div className="text-xl font-bold text-white">
          {vehicle.odometer?.toLocaleString() ?? "0"}{" "}
          <span className="text-xs text-text-muted font-normal">mi</span>
        </div>
      </div>
      <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
        <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">
          {uiStrings?.label_fuel_level || "Fuel Level"}
        </div>
        <div className="text-xl font-bold text-emerald-400">
          {vehicle.fuel_level ?? 0}%
        </div>
        <div className="w-full bg-border-color h-1 rounded-full mt-2">
          <div
            className="bg-emerald-500 h-1 rounded-full"
            style={{ width: `${vehicle.fuel_level ?? 0}%` }}
          ></div>
        </div>
      </div>
      <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
        <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">
          {uiStrings?.label_engine_hours || "Engine Hours"}
        </div>
        <div className="text-xl font-bold text-white">
          {vehicle.engine_hours?.toLocaleString() ?? "0"}{" "}
          <span className="text-xs text-text-muted font-normal">hrs</span>
        </div>
      </div>
      <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
        <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">
          {uiStrings?.label_status || "Status"}
        </div>
        <div
          className={`text-xl font-bold ${vehicle.status === "active" ? "text-emerald-400" : "text-yellow-400"} capitalize`}
        >
          {vehicle.status}
        </div>
      </div>
    </div>
  );
};
