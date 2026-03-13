import React from "react";
import { useSelector } from "react-redux";
import { RootState } from "../../../store";

const Alerts: React.FC = () => {
  const { alerts } = useSelector((state: RootState) => state.alerts);

  return (
    <div className="p-6 space-y-6 bg-black min-h-screen text-white">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Alerts Management</h1>
        <div className="flex gap-2">
          <button className="px-4 py-2 bg-red-600 hover:bg-red-700 rounded-lg text-sm font-semibold">
            Acknowledge All
          </button>
        </div>
      </div>

      <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        <div className="p-4 border-b border-gray-800 flex gap-4">
          <button className="text-sm font-medium text-blue-400 border-b-2 border-blue-400 pb-1">
            All Alerts
          </button>
          <button className="text-sm font-medium text-gray-400 hover:text-white pb-1">
            Critical
          </button>
          <button className="text-sm font-medium text-gray-400 hover:text-white pb-1">
            Warnings
          </button>
        </div>
        <div className="divide-y divide-gray-800">
          {alerts.length === 0 ? (
            <div className="p-20 text-center text-gray-500">
              No active alerts detected.
            </div>
          ) : (
            alerts.map((alert: any) => (
              <div
                key={alert.id}
                className="p-4 flex justify-between items-center hover:bg-gray-800/50 transition-colors"
              >
                <div className="flex gap-4">
                  <div
                    className={`mt-1 h-3 w-3 rounded-full ${alert.severity === "critical" ? "bg-red-500" : "bg-yellow-500"}`}
                  ></div>
                  <div>
                    <p className="text-sm font-semibold text-white">
                      {alert.message}
                    </p>
                    <p className="text-xs text-gray-500 mt-1">
                      {alert.id} • {alert.timestamp}
                    </p>
                  </div>
                </div>
                <button className="text-sm text-blue-400 hover:underline">
                  Acknowledge
                </button>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default Alerts;
