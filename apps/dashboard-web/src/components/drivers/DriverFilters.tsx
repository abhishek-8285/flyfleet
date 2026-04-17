import React from "react";
import { useUIStrings } from "../../hooks/useUIStrings";

export type DriverStatusFilter = "all" | "available" | "on_trip" | "off_duty";
export type SafetyScoreFilter =
  | "all"
  | "excellent"
  | "good"
  | "needs_review"
  | "critical";

interface DriverFiltersProps {
  statusFilter: DriverStatusFilter;
  safetyFilter: SafetyScoreFilter;
  onStatusChange: (value: DriverStatusFilter) => void;
  onSafetyChange: (value: SafetyScoreFilter) => void;
}

export const DriverFilters: React.FC<DriverFiltersProps> = ({
  statusFilter,
  safetyFilter,
  onStatusChange,
  onSafetyChange,
}) => {
  const { uiStrings: rawUiStrings } = useUIStrings("drivers");
  const uiStrings = rawUiStrings || {};

  return (
    <div className="w-full bg-surface-container border-b border-outline px-8 py-4 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10 relative">
      <div className="flex flex-wrap items-center gap-4 w-full md:w-auto">
        <div className="relative group">
          <select
            value={statusFilter}
            onChange={(e) =>
              onStatusChange(e.target.value as DriverStatusFilter)
            }
            className="appearance-none bg-surface-container border border-outline text-white text-[11px] font-bold uppercase tracking-wider rounded-xl pl-4 pr-10 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-surface-container-high transition-all min-w-[170px] shadow-sm"
          >
            <option value="all" className="bg-surface-container-highest">All Driver Status</option>
            <option value="available" className="bg-surface-container-highest">Available</option>
            <option value="on_trip" className="bg-surface-container-highest">On Duty</option>
            <option value="off_duty" className="bg-surface-container-highest">Off Duty</option>
          </select>
          <span className="material-symbols-outlined absolute right-3 top-2.5 pointer-events-none text-on-surface-variant text-sm">
            filter_list
          </span>
        </div>

        <div className="relative group">
          <select
            value={safetyFilter}
            onChange={(e) =>
              onSafetyChange(e.target.value as SafetyScoreFilter)
            }
            className="appearance-none bg-surface-container border border-outline text-white text-[11px] font-bold uppercase tracking-wider rounded-xl pl-4 pr-10 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-surface-container-high transition-all min-w-[170px] shadow-sm"
          >
            <option value="all" className="bg-surface-container-highest">
              {uiStrings.filter_all_scores || "Safety Score: All"}
            </option>
            <option value="excellent" className="bg-surface-container-highest">
              {uiStrings.filter_excellent || "Excellent (90-100)"}
            </option>
            <option value="good" className="bg-surface-container-highest">
              {uiStrings.filter_good || "Standard (80-89)"}
            </option>
            <option value="needs_review" className="bg-surface-container-highest">
              {uiStrings.filter_needs_review || "At Risk (< 80)"}
            </option>
            <option value="critical" className="bg-surface-container-highest">
              {uiStrings.filter_critical || "Critical (< 60)"}
            </option>
          </select>
          <span className="material-symbols-outlined absolute right-3 top-2.5 pointer-events-none text-on-surface-variant text-sm">
            health_and_safety
          </span>
        </div>
      </div>

      <div className="flex items-center gap-4 ml-auto w-full md:w-auto justify-end">
        <button className="hidden md:flex items-center gap-2 px-5 py-2.5 bg-surface-container border border-outline text-on-surface-variant text-[10px] font-bold uppercase tracking-wide rounded-xl hover:text-white hover:border-primary transition-all shadow-sm">
          <span className="material-symbols-outlined text-[18px]">
            download
          </span>
          {uiStrings.btn_download_report || "Export Directory"}
        </button>
        <button className="flex items-center gap-2 px-6 py-2.5 bg-primary text-white text-[10px] font-bold uppercase tracking-wide rounded-xl hover:bg-primary-dark transition-all shadow-lg shadow-primary/20">
          <span className="material-symbols-outlined text-[18px]">
            person_add
          </span>
          {uiStrings.btn_add_driver || "Add Driver"}
        </button>
      </div>
    </div>
  );
};
