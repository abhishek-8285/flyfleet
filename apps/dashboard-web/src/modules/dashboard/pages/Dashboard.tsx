import { AlertCircle, RefreshCcw } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { dashboardApi } from "../../../services/dashboardApi";
import telemetryService from "../../../services/telemetryService";
import { RootState } from "../../../store";
import { setAlerts } from "../../../store/alertsSlice";
import {
    fetchStatsFailure,
    fetchStatsStart,
    fetchStatsSuccess,
} from "../../../store/dashboardSlice";
import { setTimelineEvents } from "../../../store/timelineSlice";
import ActivityFeed from "../components/ActivityFeed";
import CriticalVehiclesTable from "../components/CriticalVehiclesTable";
import { DashboardSkeleton } from "../components/DashboardSkeleton";
import LiveOperationsMap from "../components/LiveOperationsMap";
import QuickActions from "../components/QuickActions";
import StatsGrid from "../components/StatsGrid";

const Dashboard: React.FC = () => {
  const dispatch = useDispatch();
  const { stats, loading, error } = useSelector(
    (state: RootState) => state.dashboard,
  );
  const { events } = useSelector((state: RootState) => state.timeline);
  const { vehicles } = useSelector((state: RootState) => state.telemetry);
  const { alerts } = useSelector((state: RootState) => state.alerts);
  const [isRetrying, setIsRetrying] = useState(false);

  const fetchDashboardData = async () => {
    dispatch(fetchStatsStart());
    setIsRetrying(true);
    try {
      const response = (await dashboardApi.getOverview()) as any;
      const data = response.data || response;

      dispatch(fetchStatsSuccess(data.stats));
      dispatch(setTimelineEvents(data.timeline || []));
      dispatch(setAlerts(data.alerts || []));
    } catch (err: any) {
      dispatch(
        fetchStatsFailure(err.message || "Failed to load dashboard overview"),
      );
    } finally {
      setIsRetrying(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
    telemetryService.connect();

    return () => {
      telemetryService.disconnect();
    };
  }, [dispatch]);

  const mapMarkers = Object.values(vehicles).map((v) => ({
    id: v.id,
    lat: v.latitude,
    lng: v.longitude,
    status: (v as any).status || "active",
    label: v.id,
  }));

  const criticalAlerts = alerts
    .filter(
      (a) =>
        !a.acknowledged &&
        (a.severity === "critical" || a.severity === "warning"),
    )
    .map((a) => ({
      id: a.vehicleId,
      driverId: (a.metadata as any)?.driverId || "Unknown",
      severity: a.severity as "critical" | "warning",
      location: (a.metadata as any)?.location || "Remote",
      lastUpdate: "Just now",
      slaAging: (a.metadata as any)?.sla || "+0m",
    }));

  if (loading && !isRetrying) {
    return <DashboardSkeleton />;
  }

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center bg-black">
        <div className="max-w-md p-8 bg-gray-900 border border-red-900/50 rounded-2xl text-center">
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-bold text-white mb-2">
            Connection Failure
          </h2>
          <p className="text-gray-400 mb-6">{error}</p>
          <button
            onClick={fetchDashboardData}
            className="px-6 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors flex items-center gap-2 mx-auto"
          >
            <RefreshCcw size={18} />
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6 bg-black min-h-screen text-white">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Fleet Command</h1>
          <p className="text-gray-400 mt-1">
            Real-time operational overview of your global fleet.
          </p>
        </div>
        <div className="flex gap-3">
          <div className="text-right">
            <p className="text-sm font-semibold text-green-400">
              System Healthy
            </p>
            <p className="text-xs text-gray-500">
              Last sync: {new Date().toLocaleTimeString()}
            </p>
          </div>
        </div>
      </div>

      <StatsGrid stats={stats} />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <LiveOperationsMap vehicles={mapMarkers} />
          <CriticalVehiclesTable vehicles={criticalAlerts} />
        </div>
        <div className="space-y-6">
          <QuickActions />
          <ActivityFeed events={events} />
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
