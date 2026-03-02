import React from "react";
import { useUIStrings } from "../../hooks/useUIStrings";

export const MaintenanceFilters: React.FC = () => {
  const { uiStrings } = useUIStrings("maintenance_dashboard");

  return (
    <div className="bg-surface-dark rounded-xl border border-border-color h-full flex flex-col overflow-hidden">
      <div className="p-4 border-b border-border-color bg-surface-darker">
        <h3 className="text-white font-semibold flex items-center gap-2">
          <span className="material-symbols-outlined text-primary text-[20px]">
            filter_list
          </span>
          {uiStrings?.title_filters || "Filters"}
        </h3>
      </div>
      <div className="p-5 space-y-6 overflow-y-auto flex-1 custom-scrollbar">
        <div>
          <label className="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">
            {uiStrings?.label_search_vehicle || "Search Vehicle"}
          </label>
          <div className="relative">
            <input
              className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary placeholder-text-muted/50"
              placeholder={
                uiStrings?.filter_placeholder || "ID, Model or Plate"
              }
              type="text"
            />
            <span className="material-symbols-outlined absolute right-2.5 top-2.5 text-text-muted text-sm">
              search
            </span>
          </div>
        </div>
        <div>
          <label className="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-3">
            {uiStrings?.label_service_type || "Service Type"}
          </label>
          <div className="space-y-2.5">
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                defaultChecked
                className="form-checkbox rounded border-border-color bg-surface-darker text-primary focus:ring-primary h-4 w-4 custom-checkbox"
                type="checkbox"
              />
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.type_oil_change || "Oil Change"}
              </span>
              <span className="ml-auto text-xs text-text-muted bg-border-color/30 px-1.5 rounded">
                12
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                defaultChecked
                className="form-checkbox rounded border-border-color bg-surface-darker text-primary focus:ring-primary h-4 w-4 custom-checkbox"
                type="checkbox"
              />
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.type_tire_rotation || "Tire Rotation"}
              </span>
              <span className="ml-auto text-xs text-text-muted bg-border-color/30 px-1.5 rounded">
                8
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                className="form-checkbox rounded border-border-color bg-surface-darker text-primary focus:ring-primary h-4 w-4 custom-checkbox"
                type="checkbox"
              />
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.type_brake_inspection || "Brake Inspection"}
              </span>
              <span className="ml-auto text-xs text-text-muted bg-border-color/30 px-1.5 rounded">
                3
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                className="form-checkbox rounded border-border-color bg-surface-darker text-primary focus:ring-primary h-4 w-4 custom-checkbox"
                type="checkbox"
              />
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.type_engine_tuneup || "Engine Tune-up"}
              </span>
              <span className="ml-auto text-xs text-text-muted bg-border-color/30 px-1.5 rounded">
                5
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                className="form-checkbox rounded border-border-color bg-surface-darker text-primary focus:ring-primary h-4 w-4 custom-checkbox"
                type="checkbox"
              />
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.type_safety_inspection || "Safety Inspection"}
              </span>
              <span className="ml-auto text-xs text-text-muted bg-border-color/30 px-1.5 rounded">
                2
              </span>
            </label>
          </div>
        </div>
        <div className="w-full h-px bg-border-color"></div>
        <div>
          <label className="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-3">
            {uiStrings?.label_health_status || "Health Status"}
          </label>
          <div className="space-y-2.5">
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                className="form-checkbox rounded border-border-color bg-surface-darker text-red-500 focus:ring-red-500 h-4 w-4"
                type="checkbox"
              />
              <span className="w-2 h-2 rounded-full bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)]"></span>
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.status_critical || "Critical"}
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                className="form-checkbox rounded border-border-color bg-surface-darker text-yellow-500 focus:ring-yellow-500 h-4 w-4"
                type="checkbox"
              />
              <span className="w-2 h-2 rounded-full bg-yellow-500"></span>
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.status_maintenance_due || "Maintenance Due"}
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer group">
              <input
                defaultChecked
                className="form-checkbox rounded border-border-color bg-surface-darker text-emerald-500 focus:ring-emerald-500 h-4 w-4"
                type="checkbox"
              />
              <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
              <span className="text-sm text-text-muted group-hover:text-white transition-colors">
                {uiStrings?.status_healthy || "Healthy"}
              </span>
            </label>
          </div>
        </div>
        <div className="pt-4 mt-auto">
          <button className="w-full py-2 border border-border-color rounded text-xs text-text-muted hover:text-white hover:bg-border-color transition-colors">
            {uiStrings?.btn_reset_filters || "Reset Filters"}
          </button>
        </div>
      </div>
    </div>
  );
};
