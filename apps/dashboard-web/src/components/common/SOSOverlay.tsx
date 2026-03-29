import { ExclamationTriangleIcon, XMarkIcon } from "@heroicons/react/24/solid";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

interface SOSAlert {
  id: string;
  vehicle_reg_number: string;
  driver_name: string;
  timestamp: string;
  severity: "LOW" | "MEDIUM" | "CRITICAL";
}

export const SOSOverlay: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("comp_sos");
  const [activeAlert, setActiveAlert] = useState<SOSAlert | null>(null);

  if (!activeAlert) return null;

  return (
    <div className="fixed inset-0 z-[9999] flex items-end justify-center pointer-events-none pb-12 px-4">
      <div className="bg-red-600 text-white rounded-lg shadow-2xl max-w-lg w-full pointer-events-auto ring-4 ring-red-500 ring-opacity-50 animate-bounce-subtle">
        <div className="p-4 flex items-start gap-4">
          <div className="bg-white/20 p-2 rounded-full animate-pulse">
            <ExclamationTriangleIcon className="h-8 w-8 text-white" />
          </div>
          <div className="flex-1">
            <h3 className="text-lg font-bold">
              {uiStrings?.title_sos || "EMERGENCY SOS TRIGGERED"}
            </h3>
            <p className="text-red-100 mt-1">
              {(
                uiStrings?.msg_report ||
                "Vehicle {{vehicle}} ({{driver}}) reported a {{severity}} incident."
              )
                .replace("{{vehicle}}", activeAlert.vehicle_reg_number)
                .replace("{{driver}}", activeAlert.driver_name)
                .replace("{{severity}}", activeAlert.severity)}
            </p>
            <div className="mt-4 flex gap-3">
              <button
                onClick={() => {
                  navigate("/admin/emergency");
                  setActiveAlert(null);
                }}
                className="bg-white text-red-700 px-4 py-2 rounded font-bold hover:bg-gray-100 transition flex-1"
              >
                {uiStrings?.btn_details || "View Details"}
              </button>
              <button
                onClick={() => setActiveAlert(null)}
                className="border border-white/30 text-white px-4 py-2 rounded hover:bg-white/10 transition"
              >
                {uiStrings?.btn_dismiss || "Dismiss"}
              </button>
            </div>
          </div>
          <button
            onClick={() => setActiveAlert(null)}
            className="text-white/60 hover:text-white"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>
      </div>

      {/* Red overlay flash effect */}
      <div className="fixed inset-0 bg-red-500/10 z-[-1] animate-pulse pointer-events-none"></div>
    </div>
  );
};
