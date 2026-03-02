import React from "react";

import { useUIStrings } from "../../hooks/useUIStrings";
import { Vehicle } from "../../types";

interface VehicleSpecsPanelProps {
  vehicle?: Vehicle;
}

export const VehicleSpecsPanel: React.FC<VehicleSpecsPanelProps> = ({
  vehicle,
}) => {
  const { uiStrings } = useUIStrings("comp_vehicles");
  if (!vehicle) return null;

  return (
    <div className="bg-surface-dark rounded-xl border border-border-color flex-1 flex flex-col overflow-hidden">
      <div className="flex border-b border-border-color bg-surface-darker">
        <button className="px-6 py-3 text-sm font-medium text-primary border-b-2 border-primary bg-primary/5">
          {uiStrings?.tab_specs || "Specs & Overview"}
        </button>
        <button className="px-6 py-3 text-sm font-medium text-text-muted hover:text-white hover:bg-border-color/30 transition-colors">
          {uiStrings?.tab_telemetry || "Telemetry Feed"}
        </button>
        <button className="px-6 py-3 text-sm font-medium text-text-muted hover:text-white hover:bg-border-color/30 transition-colors">
          {uiStrings?.tab_maintenance || "Maintenance Logs"}
        </button>
        <button className="px-6 py-3 text-sm font-medium text-text-muted hover:text-white hover:bg-border-color/30 transition-colors">
          {uiStrings?.tab_driver || "Assigned Driver"}
        </button>
      </div>
      <div className="p-6 overflow-y-auto custom-scrollbar flex-1">
        <h4 className="text-white font-semibold mb-4 flex items-center gap-2">
          <span className="material-symbols-outlined text-primary text-lg">
            info
          </span>
          {uiStrings?.title_specs || "Vehicle Specifications"}
        </h4>
        <div className="grid grid-cols-2 gap-x-8 gap-y-4 mb-8">
          <div className="flex justify-between py-2 border-b border-border-color/50">
            <span className="text-sm text-text-muted">
              {uiStrings?.label_model || "Make / Model"}
            </span>
            <span className="text-sm text-white font-medium">
              {vehicle.make} {vehicle.model}
            </span>
          </div>
          <div className="flex justify-between py-2 border-b border-border-color/50">
            <span className="text-sm text-text-muted">
              {uiStrings?.label_year || "Year"}
            </span>
          </div>
          <div className="flex justify-between py-2 border-b border-border-color/50">
            <span className="text-sm text-text-muted">
              {uiStrings?.label_plate || "License Plate"}
            </span>
            <span className="text-sm text-white font-medium">
              {vehicle.registration_number}
            </span>
          </div>
          <div className="flex justify-between py-2 border-b border-border-color/50">
            <span className="text-sm text-text-muted">
              {uiStrings?.label_type || "Type"}
            </span>
          </div>
          {/* Additional metadata specs if available */}
          <div className="flex justify-between py-2 border-b border-border-color/50">
            <span className="text-sm text-text-muted">
              {uiStrings?.label_fuel_type || "Fuel Type"}
            </span>
          </div>
          <div className="flex justify-between py-2 border-b border-border-color/50">
            <span className="text-sm text-text-muted">
              {uiStrings?.label_weight || "Gross Weight"}
            </span>
          </div>
        </div>
        <h4 className="text-white font-semibold mb-4 flex items-center gap-2">
          <span className="material-symbols-outlined text-primary text-lg">
            assignment_ind
          </span>
          {uiStrings?.title_assignment || "Current Assignment"}
        </h4>
        <div className="bg-surface-darker rounded-lg p-4 border border-border-color flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-full bg-border-color flex items-center justify-center">
              <span className="material-symbols-outlined text-text-muted text-2xl">
                person
              </span>
            </div>
            <div>
              <div className="text-white font-medium">
                {vehicle.current_driver?.name ?? "Unassigned"}
              </div>
              <div className="text-xs text-text-muted">
                ID: {vehicle.current_driver?.id ?? "N/A"} •{" "}
                {vehicle.current_driver?.status ?? "Unknown"}
              </div>
            </div>
          </div>
          <div className="flex flex-col items-end gap-2">
            {vehicle.current_driver && (
              <span
                className={`px-2 py-1 ${vehicle.current_driver.status === "on_trip" ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20" : "bg-yellow-500/10 text-yellow-500 border-yellow-500/20"} text-xs rounded border`}
              >
                {vehicle.current_driver.status.replace("_", " ")}
              </span>
            )}
            <button className="text-xs text-primary hover:text-white hover:underline">
              {uiStrings?.btn_view_profile || "View Driver Profile"}
            </button>
          </div>
        </div>
      </div>
      <div className="p-4 border-t border-border-color bg-surface-darker flex justify-end gap-3">
        <button className="px-4 py-2 text-xs font-medium text-white bg-primary hover:bg-blue-600 rounded transition-colors shadow-lg shadow-primary/20">
          {uiStrings?.btn_report || "Download Full Report"}
        </button>
      </div>
    </div>
  );
};
