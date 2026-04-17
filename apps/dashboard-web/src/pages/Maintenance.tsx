import React from "react";
import { MaintenanceFilters } from "../components/maintenance/MaintenanceFilters";
import { MaintenanceStats } from "../components/maintenance/MaintenanceStats";
import { ServiceLogTable } from "../components/maintenance/ServiceLogTable";
import { useUIStrings } from "../hooks/useUIStrings";

const Maintenance: React.FC = () => {
  const { uiStrings } = useUIStrings("maintenance_page");
  return (
    <div className="flex-1 flex flex-col min-w-0 bg-surface-dim overflow-hidden h-full font-sans antialiased">
      {/* Fleet Readiness Control Center */}
      <div className="w-full bg-surface-container/60 backdrop-blur-3xl border-b border-outline px-10 py-6 flex flex-col md:flex-row items-start md:items-center justify-between gap-6 flex-shrink-0 z-20 shadow-2xl">
        <div className="flex flex-wrap items-center gap-4">
          <div className="relative group">
            <select className="appearance-none bg-surface-dim border border-outline text-white text-[10px] font-black uppercase tracking-widest rounded-xl pl-4 pr-12 py-2.5 focus:border-primary cursor-pointer hover:bg-surface-container-highest/20 transition-all outline-none shadow-lg">
              <option>{uiStrings?.filter_this_week || "Current Cycle"}</option>
              <option>{uiStrings?.filter_next_7 || "Next 7 Cycles"}</option>
              <option>{uiStrings?.filter_next_30 || "Strategic: 30D"}</option>
              <option className="text-rose-500">{uiStrings?.filter_overdue || "CRITICAL OVERDUE"}</option>
            </select>
            <span className="material-symbols-outlined absolute right-4 top-2.5 pointer-events-none text-on-surface-variant text-base opacity-40 group-hover:opacity-100 transition-opacity">
              calendar_month
            </span>
          </div>
          <div className="relative group">
            <select className="appearance-none bg-surface-dim border border-outline text-white text-[10px] font-black uppercase tracking-widest rounded-xl pl-4 pr-12 py-2.5 focus:border-primary cursor-pointer hover:bg-surface-container-highest/20 transition-all outline-none shadow-lg">
              <option>
                {uiStrings?.filter_all_divisions || "Global Assets"}
              </option>
              <option>Division: NA</option>
              <option>Division: EU</option>
            </select>
            <span className="material-symbols-outlined absolute right-4 top-2.5 pointer-events-none text-on-surface-variant text-base opacity-40 group-hover:opacity-100 transition-opacity">
              domain
            </span>
          </div>
        </div>
        <div className="flex items-center gap-6 ml-auto">
          <div className="text-[9px] font-black uppercase tracking-[0.2em] text-on-surface-variant flex items-center gap-3">
            <span className="h-1 w-1 rounded-full bg-primary animate-pulse" />
            {uiStrings?.label_last_sync || "Health Sync: "}{" "}
            <span className="text-white opacity-40">
              {uiStrings?.label_just_now || "Live"}
            </span>
          </div>
          <div className="h-6 w-[1px] bg-outline opacity-30"></div>
          <button className="h-12 px-8 bg-primary hover:bg-primary-dark text-white text-[10px] font-black uppercase tracking-widest rounded-2xl shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-3">
            <span className="material-symbols-outlined text-lg">engineering</span>
            {uiStrings?.btn_log_service || "Log Maintenance"}
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="max-w-[1600px] mx-auto pb-6 space-y-6">
          <MaintenanceStats />
          <div className="grid grid-cols-12 gap-6">
            <div className="col-span-12 lg:col-span-3">
              <MaintenanceFilters />
            </div>
            <div className="col-span-12 lg:col-span-9">
              <ServiceLogTable />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Maintenance;
