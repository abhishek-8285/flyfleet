import React from "react";
import { Link } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import { Alert } from "../../types";

interface CriticalVehiclesTableProps {
  alerts: Alert[];
}

export const CriticalVehiclesTable: React.FC<CriticalVehiclesTableProps> = ({
  alerts,
}) => {
  const { uiStrings } = useUIStrings("comp_dashboard");
  return (
    <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col">
      <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
        <h3 className="text-white font-semibold">
          {uiStrings?.title_critical_table || "Critical Vehicle Status"}
        </h3>
        <div className="flex gap-3">
          <button className="text-xs font-medium text-text-muted hover:text-white flex items-center gap-1 bg-border-color/30 px-2 py-1 rounded transition-colors">
            <span className="material-symbols-outlined text-[14px]">
              download
            </span>
            {uiStrings?.btn_export || "Export Selected"}
          </button>
          <button className="text-xs font-medium text-text-muted hover:text-white flex items-center gap-1 bg-border-color/30 px-2 py-1 rounded transition-colors">
            <span className="material-symbols-outlined text-[14px]">
              edit_note
            </span>
            {uiStrings?.btn_batch_edit || "Batch Status Change"}
          </button>
        </div>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm text-text-muted">
          <thead className="bg-[#1d2936] text-xs uppercase font-medium">
            <tr>
              <th className="px-6 py-3 w-10">
                <input
                  type="checkbox"
                  className="rounded bg-surface-dark border-border-color text-primary focus:ring-primary w-4 h-4"
                />
              </th>
              <th className="px-6 py-3 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  {uiStrings?.th_vehicle_id || "Vehicle ID"}{" "}
                  <span className="material-symbols-outlined text-[12px] opacity-100 text-primary">
                    arrow_upward
                  </span>
                </div>
              </th>
              <th className="px-6 py-3 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  {uiStrings?.th_driver || "Driver"}{" "}
                  <span className="material-symbols-outlined text-[12px] opacity-0 group-hover:opacity-100">
                    sort
                  </span>
                </div>
              </th>
              <th className="px-6 py-3">{uiStrings?.th_status || "Status"}</th>
              <th className="px-6 py-3">
                {uiStrings?.th_location || "Location"}
              </th>
              <th className="px-6 py-3 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  Last Update{" "}
                  <span className="material-symbols-outlined text-[12px] opacity-0 group-hover:opacity-100">
                    arrow_downward
                  </span>
                </div>
              </th>
              <th className="px-6 py-3">SLA Aging</th>
              <th className="px-6 py-3 text-right">
                {uiStrings?.th_actions || "Action"}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-color">
            {alerts.length === 0 ? (
              <tr>
                <td
                  colSpan={8}
                  className="px-6 py-10 text-center text-text-muted italic"
                >
                  No critical vehicle alerts detected in the last 24 hours.
                </td>
              </tr>
            ) : (
              alerts.map((alert) => (
                <tr
                  key={alert.id}
                  className="hover:bg-[#1d2936] transition-colors"
                >
                  <td className="px-6 py-3">
                    <input
                      type="checkbox"
                      className="rounded bg-surface-dark border-border-color text-primary focus:ring-primary w-4 h-4"
                    />
                  </td>
                  <td className="px-6 py-3 font-medium text-white">
                    {alert.vehicle_id?.slice(-4) || "N/A"}
                  </td>
                  <td className="px-6 py-3">
                    {alert.driver_id?.slice(-4) || "System"}
                  </td>
                  <td className="px-6 py-3">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${
                        alert.severity === "critical"
                          ? "bg-red-500/10 text-red-500 border-red-500/20"
                          : "bg-orange-500/10 text-orange-500 border-orange-500/20"
                      }`}
                    >
                      {alert.type}
                    </span>
                  </td>
                  <td className="px-6 py-3 text-xs">
                    Route {alert.vehicle_id?.slice(0, 2) || "A"}
                  </td>
                  <td className="px-6 py-3 text-xs text-white">
                    {new Date(alert.timestamp).toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </td>
                  <td className="px-6 py-3 text-xs text-red-400 font-medium">
                    --
                  </td>
                  <td className="px-6 py-3 text-right">
                    <Link
                      to={`/dashboard/vehicles/${alert.vehicle_id}`}
                      className="text-primary hover:text-white text-xs"
                    >
                      Detail
                    </Link>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <div className="bg-surface-darker px-4 py-3 border-t border-border-color text-[11px] text-text-muted flex justify-between items-center">
        <span>
          Showing {alerts.length} of {alerts.length} records
        </span>
        <div className="flex gap-2">
          <button className="p-1 hover:text-white disabled:opacity-50" disabled>
            <span className="material-symbols-outlined text-sm">
              chevron_left
            </span>
          </button>
          <span className="text-white">1</span>
          <button
            className="p-1 hover:text-white"
            disabled={alerts.length < 50}
          >
            <span className="material-symbols-outlined text-sm">
              chevron_right
            </span>
          </button>
        </div>
      </div>
    </div>
  );
};
