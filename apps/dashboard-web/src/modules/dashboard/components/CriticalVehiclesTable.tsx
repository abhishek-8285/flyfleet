import { AlertCircle, Download, MoreVertical } from "lucide-react";
import React from "react";

interface Vehicle {
  id: string;
  driverId: string;
  severity: "critical" | "warning";
  location: string;
  lastUpdate: string;
  slaAging: string;
}

interface CriticalVehiclesTableProps {
  vehicles: Vehicle[];
}

const CriticalVehiclesTable: React.FC<CriticalVehiclesTableProps> = ({
  vehicles,
}) => {
  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
      <div className="p-4 border-b border-gray-800 flex justify-between items-center">
        <h2 className="text-lg font-bold text-white flex items-center gap-2">
          <AlertCircle className="text-red-500" size={20} />
          Critical Operations
        </h2>
        <button className="flex items-center gap-2 px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg text-sm transition-colors">
          <Download size={16} />
          Export
        </button>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-left">
          <thead>
            <tr className="bg-gray-950/50 text-gray-400 text-xs uppercase tracking-wider">
              <th className="px-6 py-4 font-semibold">Vehicle / Driver</th>
              <th className="px-6 py-4 font-semibold">Severity</th>
              <th className="px-6 py-4 font-semibold">Location</th>
              <th className="px-6 py-4 font-semibold">Last Update</th>
              <th className="px-6 py-4 font-semibold">SLA Aging</th>
              <th className="px-6 py-4 font-semibold text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {vehicles.length === 0 ? (
              <tr>
                <td
                  colSpan={6}
                  className="px-6 py-20 text-center text-gray-500"
                >
                  No critical operational flags found.
                </td>
              </tr>
            ) : (
              vehicles.map((vehicle) => (
                <tr
                  key={vehicle.id}
                  className="hover:bg-gray-800/30 transition-colors"
                >
                  <td className="px-6 py-4">
                    <div className="text-sm font-medium text-white">
                      {vehicle.id}
                    </div>
                    <div className="text-xs text-gray-500">
                      {vehicle.driverId}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                        vehicle.severity === "critical"
                          ? "bg-red-500/10 text-red-500 border border-red-500/20"
                          : "bg-orange-500/10 text-orange-500 border border-orange-500/20"
                      }`}
                    >
                      {vehicle.severity.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-300">
                    {vehicle.location}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-300">
                    {vehicle.lastUpdate}
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-red-400 font-semibold">
                      {vehicle.slaAging}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button className="text-gray-500 hover:text-white transition-colors">
                      <MoreVertical size={20} />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <div className="p-4 border-t border-gray-800 flex justify-center">
        <nav className="flex items-center gap-2">
          <button className="px-3 py-1 bg-gray-800 text-gray-400 rounded-md disabled:opacity-50">
            Prev
          </button>
          <span className="text-gray-500 text-sm">Page 1 of 1</span>
          <button className="px-3 py-1 bg-gray-800 text-gray-400 rounded-md disabled:opacity-50">
            Next
          </button>
        </nav>
      </div>
    </div>
  );
};

export default CriticalVehiclesTable;
