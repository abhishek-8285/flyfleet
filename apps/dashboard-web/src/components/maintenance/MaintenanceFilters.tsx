import React from "react";
import { useUIStrings } from "../../hooks/useUIStrings";

export const MaintenanceFilters: React.FC = () => {
  const { uiStrings } = useUIStrings("maintenance_dashboard");

  return (
    <div className="bg-surface-container rounded-[3rem] border border-outline h-full flex flex-col overflow-hidden shadow-2xl relative group">
      <div className="p-8 border-b border-outline bg-surface-container/40 backdrop-blur-xl">
        <h3 className="text-white text-[10px] font-black uppercase tracking-[0.3em] flex items-center gap-3">
          <span className="material-symbols-outlined text-primary text-lg">tune</span>
          {uiStrings?.title_filters || "Triage Selectors"}
        </h3>
      </div>
      <div className="p-8 space-y-10 overflow-y-auto flex-1 custom-scrollbar">
        <div>
          <label className="block text-[9px] font-black text-on-surface-variant uppercase tracking-[0.25em] mb-4">
            {uiStrings?.label_search_vehicle || "Asset Signature"}
          </label>
          <div className="relative group">
            <input
              className="w-full bg-surface-dim border border-outline rounded-xl px-4 py-3 text-[11px] text-white focus:border-primary placeholder-on-surface-variant/40 outline-none transition-all shadow-inner group-hover:bg-surface-container-highest/20"
              placeholder={
                uiStrings?.filter_placeholder || "Search ID / Plate..."
              }
              type="text"
            />
            <span className="material-symbols-outlined absolute right-4 top-3 text-on-surface-variant/40 text-lg group-hover:text-primary transition-colors">
              search
            </span>
          </div>
        </div>
        <div>
          <label className="block text-[9px] font-black text-on-surface-variant uppercase tracking-[0.25em] mb-5">
            {uiStrings?.label_service_type || "Service Categories"}
          </label>
          <div className="space-y-4">
            {[
              { id: 'oil', label: uiStrings?.type_oil_change || "Oil Replacement", count: 12 },
              { id: 'tire', label: uiStrings?.type_tire_rotation || "Dynamic Rotation", count: 8 },
              { id: 'brake', label: uiStrings?.type_brake_inspection || "Brake Integrity", count: 3 },
              { id: 'engine', label: uiStrings?.type_engine_tuneup || "Core Calibration", count: 5 },
              { id: 'safety', label: uiStrings?.type_safety_inspection || "Safety Oversight", count: 2 },
            ].map((type) => (
              <label key={type.id} className="flex items-center gap-4 cursor-pointer group/label">
                <input
                  defaultChecked
                  className="form-checkbox rounded-[6px] border-outline bg-surface-dim text-primary focus:ring-primary h-5 w-5 transition-all cursor-pointer"
                  type="checkbox"
                />
                <span className="text-[11px] font-bold text-on-surface-variant group-hover/label:text-white transition-colors tracking-wide">
                  {type.label}
                </span>
                <span className="ml-auto text-[9px] font-black text-on-surface-variant bg-surface-dim border border-outline px-2 py-0.5 rounded-lg opacity-40">
                  {type.count}
                </span>
              </label>
            ))}
          </div>
        </div>
        <div className="w-full h-[1px] bg-outline opacity-20"></div>
        <div>
          <label className="block text-[9px] font-black text-on-surface-variant uppercase tracking-[0.25em] mb-5">
            {uiStrings?.label_health_status || "Diagnostic Status"}
          </label>
          <div className="space-y-4">
            <label className="flex items-center gap-4 cursor-pointer group/label">
              <input
                className="form-checkbox rounded border-outline bg-surface-dim text-rose-500 focus:ring-rose-500 h-5 w-5"
                type="checkbox"
              />
              <span className="w-2 h-2 rounded-full bg-rose-500 shadow-[0_0_12px_rgba(244,63,94,0.6)] animate-pulse"></span>
              <span className="text-[11px] font-bold text-on-surface-variant group-hover/label:text-white transition-colors uppercase tracking-widest">
                {uiStrings?.status_critical || "Critical Failure"}
              </span>
            </label>
            <label className="flex items-center gap-4 cursor-pointer group/label">
              <input
                className="form-checkbox rounded border-outline bg-surface-dim text-amber-500 focus:ring-amber-500 h-5 w-5"
                type="checkbox"
              />
              <span className="w-2 h-2 rounded-full bg-amber-500"></span>
              <span className="text-[11px] font-bold text-on-surface-variant group-hover/label:text-white transition-colors uppercase tracking-widest">
                {uiStrings?.status_maintenance_due || "Service Due"}
              </span>
            </label>
            <label className="flex items-center gap-4 cursor-pointer group/label">
              <input
                defaultChecked
                className="form-checkbox rounded border-outline bg-surface-dim text-emerald-500 focus:ring-emerald-500 h-5 w-5"
                type="checkbox"
              />
              <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
              <span className="text-[11px] font-bold text-on-surface-variant group-hover/label:text-white transition-colors uppercase tracking-widest">
                {uiStrings?.status_healthy || "Operational"}
              </span>
            </label>
          </div>
        </div>
        <div className="pt-8">
          <button className="h-10 w-full bg-surface-dim border border-outline text-[10px] font-black uppercase tracking-widest text-on-surface-variant hover:text-white hover:bg-surface-container-highest rounded-xl transition-all shadow-md">
            {uiStrings?.btn_reset_filters || "Purge Filters"}
          </button>
        </div>
      </div>
    </div>
  );
};
