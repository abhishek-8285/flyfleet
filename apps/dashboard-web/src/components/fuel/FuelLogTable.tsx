import { format } from "date-fns";
import React from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { FuelLog } from "../../types/fuel";
interface FuelLogTableProps {
  logs: FuelLog[];
  loading: boolean;
}

export const FuelLogTable: React.FC<FuelLogTableProps> = ({
  logs,
  loading,
}) => {
  const { uiStrings } = useUIStrings("fuel");
  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-3">
        <div className="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
        <p className="text-text-muted text-xs animate-pulse">
          Loading fuel logs...
        </p>
      </div>
    );
  }

  if (logs.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-text-muted">
        <span className="material-symbols-outlined text-4xl mb-2 opacity-20">
          no_sim
        </span>
        <p className="text-sm">No fuel logs found</p>
        <p className="text-xs opacity-60 mt-1">
          Try adjusting your filters or search query
        </p>
      </div>
    );
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case "APPROVED":
        return "text-green-400 bg-green-400/10 border-green-400/20";
      case "REJECTED":
        return "text-red-400 bg-red-400/10 border-red-400/20";
      default:
        return "text-orange-400 bg-orange-400/10 border-orange-400/20";
    }
  };

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left border-collapse">
        <thead className="bg-surface-darker/50 text-text-muted uppercase text-[10px] font-bold tracking-widest">
          <tr>
            <th className="px-6 py-4">
              {uiStrings?.th_transaction_date || "Transaction Date"}
            </th>
            <th className="px-6 py-4">{uiStrings?.th_vehicle || "Vehicle"}</th>
            <th className="px-6 py-4">
              {uiStrings?.th_station || "Station / Location"}
            </th>
            <th className="px-6 py-4">
              {uiStrings?.th_quantity || "Quantity / Unit"}
            </th>
            <th className="px-6 py-4">
              {uiStrings?.th_total_cost || "Total Cost"}
            </th>
            <th className="px-6 py-4">
              {uiStrings?.th_odometer || "Odometer"}
            </th>
            <th className="px-6 py-4">{uiStrings?.th_status || "Status"}</th>
            <th className="px-6 py-4"></th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border-color/50">
          {logs.map((log) => (
            <tr
              key={log.id}
              className="hover:bg-surface-dark/30 transition-colors group"
            >
              <td className="px-6 py-4">
                <div className="text-white text-sm font-medium">
                  {format(new Date(log.transaction_date), "MMM dd, yyyy")}
                </div>
                <div className="text-text-muted text-[11px] mt-0.5">
                  {format(new Date(log.transaction_date), "HH:mm")}
                </div>
              </td>
              <td className="px-6 py-4">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg bg-surface-dark flex items-center justify-center text-primary group-hover:bg-primary/20 transition-colors">
                    <span className="material-symbols-outlined text-[18px]">
                      local_shipping
                    </span>
                  </div>
                  <div>
                    <div className="text-white text-sm font-medium">
                      {log.vehicle?.name || log.vehicle_id.slice(0, 8)}
                    </div>
                    <div className="text-text-muted text-[11px]">
                      {log.vehicle?.plate || "No Plate"}
                    </div>
                  </div>
                </div>
              </td>
              <td className="px-6 py-4">
                <div className="text-white text-sm">{log.station_name}</div>
                <div className="text-text-muted text-[11px] mt-0.5">
                  {log.fuel_type}
                </div>
              </td>
              <td className="px-6 py-4">
                <div className="text-white text-sm font-medium">
                  {log.quantity.toLocaleString()} L
                </div>
                <div className="text-text-muted text-[11px] mt-0.5">
                  ${log.unit_price}/L
                </div>
              </td>
              <td className="px-6 py-4">
                <div className="text-white text-sm font-bold tracking-tight">
                  ${log.total_cost.toLocaleString()}
                </div>
              </td>
              <td className="px-6 py-4">
                <div className="text-white text-sm">
                  {log.odometer.toLocaleString()} km
                </div>
              </td>
              <td className="px-6 py-4">
                <span
                  className={`px-2 py-0.5 rounded-full text-[10px] font-bold border ${getStatusColor(log.status)}`}
                >
                  {log.status}
                </span>
              </td>
              <td className="px-6 py-4 text-right">
                <button className="p-1 hover:text-white text-text-muted hover:bg-surface-dark rounded transition-all transform hover:scale-110">
                  <span className="material-symbols-outlined text-[20px]">
                    more_vert
                  </span>
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
