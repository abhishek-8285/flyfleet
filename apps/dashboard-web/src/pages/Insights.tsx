import {
    ChartBarIcon,
    CheckCircleIcon,
    CpuChipIcon,
    ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useUIStrings } from "../hooks/useUIStrings";
import { AppDispatch, RootState } from "../store";
import { acknowledgeAlert, fetchAlerts } from "../store/alertsSlice";
import { Alert } from "../types";

const Insights: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { alerts, loading } = useSelector((state: RootState) => state.alerts);
  const [systemMeta, setSystemMeta] = React.useState<any>(null);
  const { uiStrings } = useUIStrings("insights");

  useEffect(() => {
    dispatch(fetchAlerts({}) as any);
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) => setSystemMeta(data.system_metadata))
      .catch((err) => console.error("Global config fetch failed", err));
  }, [dispatch]);

  const handleAcknowledge = (id: string) => {
    dispatch(acknowledgeAlert(id));
  };

  // Filter for AI-generated predictive maintenance alerts
  const aiAlerts = alerts.filter(
    (alert: Alert) =>
      alert.type?.includes("Predictive Maintenance") ||
      alert.message?.includes("AI detected"),
  );

  const activeAiAlerts = aiAlerts.filter((a: Alert) => !a.acknowledged);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-lg text-text-muted">
          {uiStrings?.loading || "Loading AI Insights..."}
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center gap-3 mb-6">
        <CpuChipIcon className="w-8 h-8 text-primary" />
        <h1 className="text-3xl font-bold text-white">
          {uiStrings?.title || "Diagnostics & AI Insights"}
        </h1>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-card-bg p-6 rounded-xl border border-border-color shadow-sm">
          <div className="flex justify-between items-start">
            <div>
              <p className="text-sm text-text-muted mb-1">
                {uiStrings?.card_warnings || "Active AI Warnings"}
              </p>
              <h3 className="text-3xl font-bold text-white">
                {activeAiAlerts.length}
              </h3>
            </div>
            <div
              className={`p-3 rounded-lg ${activeAiAlerts.length > 0 ? "bg-red-500/10 text-red-500" : "bg-green-500/10 text-green-500"}`}
            >
              {activeAiAlerts.length > 0 ? (
                <ExclamationTriangleIcon className="w-6 h-6" />
              ) : (
                <CheckCircleIcon className="w-6 h-6" />
              )}
            </div>
          </div>
        </div>

        <div className="bg-card-bg p-6 rounded-xl border border-border-color shadow-sm">
          <div className="flex justify-between items-start">
            <div>
              <p className="text-sm text-text-muted mb-1">
                {uiStrings?.card_predictions || "Total Predictions (30d)"}
              </p>
              <h3 className="text-3xl font-bold text-white">1,245</h3>
            </div>
            <div className="p-3 rounded-lg bg-primary/10 text-primary">
              <ChartBarIcon className="w-6 h-6" />
            </div>
          </div>
        </div>

        <div className="bg-card-bg p-6 rounded-xl border border-border-color shadow-sm">
          <div className="flex justify-between items-start">
            <div>
              <p className="text-sm text-text-muted mb-1">
                {uiStrings?.card_accuracy || "Model Accuracy"}
              </p>
              <h3 className="text-3xl font-bold text-white">
                {systemMeta?.success_rate || "--%"}
              </h3>
            </div>
            <div className="p-3 rounded-lg bg-green-500/10 text-green-500">
              <CheckCircleIcon className="w-6 h-6" />
            </div>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-white mb-4 border-b border-border-color pb-2">
          {uiStrings?.title_alerts || "Predictive Maintenance Alerts"}
        </h2>
        {aiAlerts.length === 0 ? (
          <div className="bg-card-bg rounded-lg border border-border-color p-8 text-center text-text-muted">
            <CheckCircleIcon className="w-12 h-12 mx-auto mb-4 text-green-500 opacity-80" />
            <p className="text-lg text-white font-medium">
              {uiStrings?.empty_title || "All systems normal"}
            </p>
            <p>
              {uiStrings?.empty_desc ||
                "The AI models predict normal operations for all tracked vehicles."}
            </p>
          </div>
        ) : (
          aiAlerts.map((alert: Alert) => {
            // Extract probability if available like "(98%):"
            const probMatch = alert.message.match(/\((\d+)%\)/);
            const probability = probMatch ? parseInt(probMatch[1]) : 0;

            return (
              <div
                key={alert.id}
                className={`flex flex-col md:flex-row gap-6 bg-card-bg rounded-xl border border-border-color p-6 transition-all ${
                  !alert.acknowledged
                    ? "ring-1 ring-red-500/50 shadow-[0_0_15px_rgba(239,68,68,0.1)]"
                    : "opacity-70"
                }`}
              >
                <div className="flex-1">
                  <div className="flex items-center space-x-3 mb-3">
                    <span className="px-3 py-1 bg-red-500/20 text-red-400 text-xs font-bold rounded-full uppercase tracking-wider">
                      {uiStrings?.badge_prediction || "AI PREDICTION"}
                    </span>
                    <span className="text-lg font-bold text-white tracking-tight">
                      {alert.type}
                    </span>
                    <span className="text-sm text-text-muted">
                      {new Date(alert.timestamp).toLocaleString()}
                    </span>
                  </div>

                  <p className="text-gray-300 mb-4">{alert.message}</p>

                  <div className="bg-body-bg rounded-lg p-4 border border-border-color/50">
                    <div className="flex justify-between text-sm mb-2">
                      <span className="text-text-muted font-medium">
                        {uiStrings?.label_probability || "Failure Probability"}
                      </span>
                      <span
                        className={
                          probability > 80
                            ? "text-red-400 font-bold"
                            : "text-yellow-400 font-bold"
                        }
                      >
                        {probability}%
                      </span>
                    </div>
                    {/* Progress bar */}
                    <div className="w-full bg-border-color rounded-full h-2.5 overflow-hidden">
                      <div
                        className={`h-2.5 rounded-full ${probability > 80 ? "bg-red-500" : "bg-yellow-500"}`}
                        style={{ width: `${probability}%` }}
                      ></div>
                    </div>
                  </div>
                </div>

                <div className="flex flex-col justify-center items-end border-t md:border-t-0 md:border-l border-border-color pt-4 md:pt-0 md:pl-6 min-w-[200px]">
                  <p className="text-sm text-text-muted mb-2">
                    {uiStrings?.label_vehicle || "Vehicle ID"}
                  </p>
                  <p className="font-mono text-white bg-[#0f151b] px-3 py-1.5 rounded text-sm mb-4 truncate max-w-[150px]">
                    {alert.vehicle_id || "Unknown"}
                  </p>

                  {!alert.acknowledged ? (
                    <button
                      onClick={() => handleAcknowledge(alert.id)}
                      className="w-full px-4 py-2.5 bg-primary/10 hover:bg-primary/20 text-primary border border-primary/30 font-semibold text-sm rounded-lg transition-colors"
                    >
                      {uiStrings?.btn_acknowledge || "Acknowledge Alert"}
                    </button>
                  ) : (
                    <div className="w-full px-4 py-2.5 bg-green-500/10 text-green-500 font-semibold text-sm rounded-lg text-center flex items-center justify-center gap-2">
                      <CheckCircleIcon className="w-4 h-4" />
                      {uiStrings?.status_viewed || "Viewed"}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};

export default Insights;
