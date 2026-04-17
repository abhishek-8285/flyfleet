import React from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { Driver } from "../../types";

type DriverRow = Driver & {
  license_status?: string;
  license_expiry?: string;
  current_vehicle?: {
    id?: string;
    name?: string;
    registration_number?: string;
    license_plate?: string;
  };
  last_active?: string;
  last_seen?: string;
};

interface DriverTableProps {
  drivers: DriverRow[];
  loading: boolean;
  onDriverClick?: (driver: Driver) => void;
  // Multi-selection props
  selectedIds?: Set<string>;
  onSelect?: (id: string) => void;
  onSelectAll?: () => void;
  // Action props
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
}

const formatRelativeTime = (value?: string) => {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  const diffMs = Date.now() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  if (diffMins < 1) return "ACTIVE";
  if (diffMins < 60) return `${diffMins}M AGO`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}H AGO`;
  return date.toLocaleDateString();
};

const getScoreColor = (score: number) => {
  if (score >= 90) return "text-emerald-400";
  if (score >= 80) return "text-primary";
  if (score >= 60) return "text-amber-400";
  return "text-rose-400";
};

const getScoreGradient = (score: number) => {
  if (score >= 90) return "from-emerald-500 to-green-400 shadow-[0_0_8px_rgba(16,185,129,0.3)]";
  if (score >= 80) return "from-primary to-blue-400 shadow-[0_0_8px_rgba(19,127,236,0.3)]";
  if (score >= 60) return "from-amber-500 to-yellow-400";
  return "from-rose-500 to-red-400";
};

export const DriverTable: React.FC<DriverTableProps> = ({
  drivers,
  loading,
  onDriverClick,
  selectedIds = new Set(),
  onSelect,
  onSelectAll,
  onEdit,
  onDelete,
}) => {
  const { uiStrings: rawUiStrings } = useUIStrings("drivers");
  const { uiStrings: commonStrings } = useUIStrings("common");
  const uiStrings = rawUiStrings || {};

  const allSelected = drivers.length > 0 && selectedIds.size === drivers.length;

  return (
    <div className="bg-surface-container rounded-2xl border border-outline flex flex-col h-full shadow-sm overflow-hidden animate-entrance">
      <div className="overflow-auto custom-scrollbar flex-1">
        <table className="w-full text-left font-body">
          <thead className="bg-surface-container-high text-[10px] font-bold uppercase tracking-wider text-on-surface-variant sticky top-0 z-10 border-b border-outline">
            <tr>
              {onSelect && (
                <th className="px-6 py-4 w-12">
                  <input
                    type="checkbox"
                    checked={allSelected}
                    onChange={() => onSelectAll?.()}
                    className="size-4 rounded border-outline bg-surface-container accent-primary transition-all cursor-pointer opacity-50 hover:opacity-100"
                  />
                </th>
              )}
              <th className="px-6 py-4 cursor-pointer hover:text-white transition-colors group">
                <div className="flex items-center gap-2">
                  {uiStrings.th_name_id || "Driver Details"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-20 group-hover:opacity-100 transition-opacity">
                    sort
                  </span>
                </div>
              </th>
              <th className="px-6 py-4 cursor-pointer hover:text-white transition-colors group">
                <div className="flex items-center gap-2">
                  {uiStrings.th_license_status || "License Status"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-20 group-hover:opacity-100 transition-opacity">
                    filter_list
                  </span>
                </div>
              </th>
              <th className="px-6 py-4">
                {uiStrings.th_current_vehicle || "Assigned Vehicle"}
              </th>
              <th className="px-6 py-4 cursor-pointer hover:text-white transition-colors group text-center">
                <div className="flex items-center justify-center gap-2">
                  {uiStrings.th_safety_score || "Safety Rating"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-20 group-hover:opacity-100 transition-opacity">
                    arrow_downward
                  </span>
                </div>
              </th>
              <th className="px-6 py-4 cursor-pointer hover:text-white transition-colors group">
                <div className="flex items-center gap-2">
                  {uiStrings.th_last_active || "Last Activity"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-20 group-hover:opacity-100 transition-opacity">
                    sort
                  </span>
                </div>
              </th>
              <th className="px-6 py-4 text-right">
                {uiStrings.th_actions || "Actions"}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-outline">
            {loading ? (
              [...Array(5)].map((_, i) => (
                <tr key={i} className="animate-pulse">
                  <td colSpan={7} className="px-6 py-5">
                    <div className="h-4 bg-white/5 rounded w-3/4"></div>
                  </td>
                </tr>
              ))
            ) : drivers.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-6 py-24 text-center">
                  <span className="material-symbols-outlined text-4xl text-on-surface-variant opacity-10 mb-4 block">person_off</span>
                  <p className="text-sm font-bold text-on-surface-variant uppercase tracking-wider opacity-40">No Drivers Found</p>
                </td>
              </tr>
            ) : (
              drivers.map((driver) => {
                 const initials = `${driver.name?.[0] || driver.first_name?.[0] || ""}${driver.name?.split(" ")[1]?.[0] || driver.last_name?.[0] || ""}`.toUpperCase();
                const score = Math.max(0, Math.min(100, Math.round(driver.behavior_score ?? driver.rating ?? 0)));
                const statusLabel = driver.status.replace("_", " ");
                const licenseStatus = driver.license_status;
                const vehicleLabel =
                  driver.current_vehicle?.registration_number ||
                  driver.current_vehicle?.name ||
                  driver.current_vehicle?.id;

                return (
                  <tr
                    key={driver.id}
                    className={`hover:bg-surface-container-high/40 transition-all group cursor-pointer border-l-2 active:bg-surface-container-high/60 ${selectedIds.has(driver.id) ? 'border-primary bg-primary/5' : 'border-transparent hover:border-primary'}`}
                    onClick={(e) => {
                      // Prevent trigger if clicking checkbox or action icons
                      const target = e.target as HTMLElement;
                      if (target.tagName === 'INPUT' || target.closest('button')) return;
                      onDriverClick?.(driver);
                    }}
                  >
                    {onSelect && (
                      <td className="px-6 py-4">
                        <input
                          type="checkbox"
                          checked={selectedIds.has(driver.id)}
                          onChange={() => onSelect(driver.id)}
                          className="size-4 rounded border-outline bg-surface-container accent-primary transition-all cursor-pointer"
                          onClick={(e) => e.stopPropagation()}
                        />
                      </td>
                    )}
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-4">
                        <div className="size-10 rounded-xl bg-surface-container-highest border border-outline flex items-center justify-center text-primary font-bold text-xs shadow-sm group-hover:border-primary/50 transition-all">
                          {initials || "DR"}
                        </div>
                        <div>
                          <p className="text-sm font-bold text-white group-hover:text-primary transition-colors tracking-tight">
                            {driver.name || `${driver.first_name} ${driver.last_name}`}
                          </p>
                          <p className="text-[10px] text-on-surface-variant font-mono uppercase tracking-tight mt-0.5 opacity-60">
                            ID: {driver.id.slice(0, 8).toUpperCase()}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-col gap-1.5 items-start">
                        <span className={`px-2 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider border border-emerald-500/10 bg-emerald-500/5 text-emerald-400`}>
                          {licenseStatus || statusLabel}
                        </span>
                        <p className="text-[9px] text-on-surface-variant font-medium uppercase tracking-tight opacity-60">
                          {driver.license_expiry ? `EXP: ${driver.license_expiry}` : 'Verified'}
                        </p>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {vehicleLabel ? (
                        <div className="flex items-center gap-3 group/asset">
                          <div className="p-1.5 bg-white/5 rounded-lg border border-white/5 group-hover/asset:border-primary/30 transition-all">
                            <span className="material-symbols-outlined text-[18px] text-text-muted group-hover/asset:text-primary transition-colors">
                              local_shipping
                            </span>
                          </div>
                          <span className="text-xs font-bold text-white tracking-tight uppercase group-hover/asset:text-primary transition-colors">
                            {vehicleLabel}
                          </span>
                        </div>
                      ) : (
                        <span className="text-[10px] text-on-surface-variant/40 font-bold uppercase tracking-wider">
                          Unassigned
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-col items-center gap-2">
                        <span className={`text-sm font-bold tracking-tight ${getScoreColor(score)}`}>
                          {score}<span className="text-[8px] opacity-40 ml-0.5">/100</span>
                        </span>
                        <div className="w-16 h-1 bg-white/5 rounded-full overflow-hidden border border-white/5">
                          <div
                            className={`h-full bg-gradient-to-r transition-all duration-1000 ${getScoreGradient(score)}`}
                            style={{ width: `${score}%` }}
                          ></div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-[10px] font-bold text-on-surface-variant">
                      <p className="text-white font-medium">
                        {formatRelativeTime(driver.last_active || driver.last_seen || driver.updated_at)}
                      </p>
                      <p className="text-[8px] uppercase tracking-wider font-bold mt-1 opacity-40">
                        Activity Log
                      </p>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        {onEdit && (
                          <button 
                            onClick={(e) => { e.stopPropagation(); onEdit(driver.id); }}
                            className="p-1.5 rounded-lg hover:bg-white/5 text-on-surface-variant hover:text-primary transition-all border border-transparent hover:border-outline"
                            title="Edit Node"
                          >
                            <span className="material-symbols-outlined text-[18px]">edit</span>
                          </button>
                        )}
                        {onDelete && (
                          <button 
                            onClick={(e) => { e.stopPropagation(); onDelete(driver.id); }}
                            className="p-1.5 rounded-lg hover:bg-white/5 text-on-surface-variant hover:text-red-400 transition-all border border-transparent hover:border-outline"
                            title="Delete Driver"
                          >
                            <span className="material-symbols-outlined text-[18px]">delete</span>
                          </button>
                        )}
                        <button className="p-1.5 text-on-surface-variant hover:text-white hover:bg-white/5 rounded-lg transition-all group/btn">
                          <span className="material-symbols-outlined text-[18px]">
                            more_vert
                          </span>
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Professional Pagination Footer */}
      <div className="bg-surface-container-high px-6 py-4 border-t border-outline flex items-center justify-between flex-shrink-0">
        <div className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">
          {(uiStrings.msg_showing_results || "Showing {count} drivers").replace(
            "{count}",
            String(drivers.length),
          )}
        </div>
        <div className="flex gap-4">
          <button
            className="px-4 py-1.5 text-[10px] font-bold uppercase tracking-wide text-on-surface-variant bg-surface-container border border-outline rounded-lg hover:text-white hover:border-primary transition-all disabled:opacity-20 active:scale-95"
            disabled
          >
            {commonStrings?.btn_prev || "Previous"}
          </button>
          <button className="px-4 py-1.5 text-[10px] font-bold uppercase tracking-wide text-white bg-primary border border-primary/20 rounded-lg shadow-lg shadow-primary/20 hover:bg-primary-dark transition-all active:scale-95">
            {commonStrings?.btn_next || "Next"}
          </button>
        </div>
      </div>
    </div>
  );
};
