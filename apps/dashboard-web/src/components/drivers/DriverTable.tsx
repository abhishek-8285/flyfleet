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
}

const formatRelativeTime = (value?: string) => {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  const diffMs = Date.now() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  if (diffMins < 1) return "Just now";
  if (diffMins < 60) return `${diffMins} mins ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  return date.toLocaleDateString();
};

const getScoreColor = (score: number) => {
  if (score >= 90) return "text-emerald-500";
  if (score >= 80) return "text-emerald-400";
  if (score >= 60) return "text-yellow-500";
  return "text-red-500";
};

export const DriverTable: React.FC<DriverTableProps> = ({
  drivers,
  loading,
}) => {
  const { uiStrings: rawUiStrings } = useUIStrings("drivers");
  const { uiStrings: commonStrings } = useUIStrings("common");
  const uiStrings = rawUiStrings || {};

  return (
    <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col h-full shadow-sm overflow-hidden">
      <div className="overflow-auto custom-scrollbar flex-1">
        <table className="w-full text-left text-sm text-text-muted">
          <thead className="bg-[#1d2936] text-xs uppercase font-medium sticky top-0 z-10 shadow-sm">
            <tr>
              <th className="px-6 py-4 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  {uiStrings.th_name_id || "Name / ID"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">
                    sort
                  </span>
                </div>
              </th>
              <th className="px-6 py-4 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  {uiStrings.th_license_status || "License Status"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">
                    filter_list
                  </span>
                </div>
              </th>
              <th className="px-6 py-4">
                {uiStrings.th_current_vehicle || "Current Vehicle"}
              </th>
              <th className="px-6 py-4 cursor-pointer hover:text-white group text-center">
                <div className="flex items-center justify-center gap-1">
                  {uiStrings.th_safety_score || "Safety Score"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">
                    arrow_downward
                  </span>
                </div>
              </th>
              <th className="px-6 py-4 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  {uiStrings.th_last_active || "Last Active"}{" "}
                  <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">
                    sort
                  </span>
                </div>
              </th>
              <th className="px-6 py-4 text-right">
                {uiStrings.th_actions || "Actions"}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-color">
            {loading ? (
              <tr>
                <td
                  colSpan={6}
                  className="px-6 py-8 text-center text-text-muted"
                >
                  Loading drivers...
                </td>
              </tr>
            ) : drivers.length === 0 ? (
              <tr>
                <td
                  colSpan={6}
                  className="px-6 py-8 text-center text-text-muted"
                >
                  No drivers found.
                </td>
              </tr>
            ) : (
              drivers.map((driver) => {
                const initials = driver.name
                  .split(" ")
                  .map((part) => part[0])
                  .slice(0, 2)
                  .join("")
                  .toUpperCase();
                const score = Math.max(
                  0,
                  Math.min(
                    100,
                    Math.round(driver.behavior_score ?? driver.rating ?? 0),
                  ),
                );
                const statusLabel = driver.status.replace("_", " ");
                const licenseStatus = driver.license_status;
                const licenseExpiry = driver.license_expiry;
                const vehicleLabel =
                  driver.current_vehicle?.name ||
                  driver.current_vehicle?.registration_number ||
                  driver.current_vehicle?.license_plate ||
                  driver.current_vehicle?.id;

                return (
                  <tr
                    key={driver.id}
                    className="hover:bg-[#1d2936] transition-colors group"
                  >
                    <td className="px-6 py-3">
                      <div className="flex items-center gap-3">
                        <div className="size-8 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs">
                          {initials || "DR"}
                        </div>
                        <div>
                          <div className="font-medium text-white">
                            {driver.name}
                          </div>
                          <div className="text-[10px] text-text-muted">
                            ID: {driver.id}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-3">
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 capitalize">
                        {licenseStatus || statusLabel}
                      </span>
                      {licenseExpiry ? (
                        <div className="text-[10px] mt-0.5">
                          Exp: {licenseExpiry}
                        </div>
                      ) : (
                        <div className="text-[10px] mt-0.5 capitalize">
                          {statusLabel}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-3">
                      {vehicleLabel ? (
                        <div className="flex items-center gap-2">
                          <span className="material-symbols-outlined text-[16px] text-text-muted">
                            local_shipping
                          </span>
                          <span className="text-white">{vehicleLabel}</span>
                        </div>
                      ) : (
                        <span className="text-text-muted/50 italic text-xs">
                          Unassigned
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-3 text-center">
                      <div className="inline-flex flex-col items-center">
                        <span className={`${getScoreColor(score)} font-bold`}>
                          {score}
                        </span>
                        <div className="w-16 h-1 bg-border-color rounded-full mt-1 overflow-hidden">
                          <div
                            className={`h-full ${score >= 90 ? "bg-emerald-500" : score >= 80 ? "bg-emerald-400" : score >= 60 ? "bg-yellow-500" : "bg-red-500"}`}
                            style={{ width: `${score}%` }}
                          ></div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-3 text-xs">
                      <div className="text-white">
                        {formatRelativeTime(
                          driver.last_active || driver.last_seen,
                        )}
                      </div>
                      <div className="text-[10px] text-text-muted capitalize">
                        {statusLabel}
                      </div>
                    </td>
                    <td className="px-6 py-3 text-right">
                      <button className="p-1 text-text-muted hover:text-white hover:bg-border-color rounded transition-colors">
                        <span className="material-symbols-outlined text-[18px]">
                          more_vert
                        </span>
                      </button>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
      <div className="bg-surface-darker px-6 py-3 border-t border-border-color flex items-center justify-between flex-shrink-0">
        <div className="text-xs text-text-muted">
          {(uiStrings.msg_showing_results || "Showing {count} drivers").replace(
            "{count}",
            String(drivers.length),
          )}
        </div>
        <div className="flex gap-2">
          <button
            className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors disabled:opacity-50"
            disabled
          >
            {commonStrings?.btn_prev || "Previous"}
          </button>
          <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">
            {commonStrings?.btn_next || "Next"}
          </button>
        </div>
      </div>
    </div>
  );
};
