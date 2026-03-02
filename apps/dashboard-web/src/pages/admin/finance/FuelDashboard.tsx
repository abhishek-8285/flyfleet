import React, { useEffect, useState } from "react";
import {
    Area,
    AreaChart,
    CartesianGrid,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface FuelRecord {
  id?: string;
  amount?: number;
  liters?: number;
  volume?: number;
  fuel_liters?: number;
  price_per_liter?: number;
  created_at?: string;
  timestamp?: string;
  occurred_at?: string;
  is_anomaly?: boolean;
  anomaly?: boolean;
  status?: string;
  alert_type?: string;
  vehicle_id?: string;
  vehicle_name?: string;
  location_name?: string;
}

const FuelDashboard: React.FC = () => {
  const [stats, setStats] = useState({
    total_spend: 0,
    total_liters: 0,
    avg_price: 0,
    anomalies: 0,
  });
  const [chartData, setChartData] = useState<any[]>([]);
  const [alerts, setAlerts] = useState<FuelRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { uiStrings } = useUIStrings("fleet_fuel");

  useEffect(() => {
    const loadFuelData = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = (await apiClient.get("/finance/fuel", {
          params: { limit: 200 },
        })) as any;
        const records: FuelRecord[] = Array.isArray(response)
          ? response
          : Array.isArray(response?.data)
            ? response.data
            : [];

        const totals = records.reduce(
          (acc, record) => {
            const amount = Number(record.amount || 0);
            const liters = Number(
              record.liters || record.volume || record.fuel_liters || 0,
            );
            if (!Number.isNaN(amount)) acc.total_spend += amount;
            if (!Number.isNaN(liters)) acc.total_liters += liters;
            if (
              record.price_per_liter &&
              !Number.isNaN(Number(record.price_per_liter))
            ) {
              acc.price_points.push(Number(record.price_per_liter));
            }
            return acc;
          },
          { total_spend: 0, total_liters: 0, price_points: [] as number[] },
        );

        const avgFromPoints = totals.price_points.length
          ? totals.price_points.reduce((sum, v) => sum + v, 0) /
            totals.price_points.length
          : 0;

        const anomalyRecords = records.filter(
          (record) =>
            record.is_anomaly ||
            record.anomaly ||
            record.status === "ANOMALY" ||
            record.alert_type === "ANOMALY",
        );

        const chartByDay = new Map<string, number>();
        records.forEach((record, index) => {
          const dateValue =
            record.occurred_at || record.created_at || record.timestamp;
          if (!dateValue) {
            chartByDay.set(
              `Txn ${index + 1}`,
              (chartByDay.get(`Txn ${index + 1}`) || 0) + (record.amount || 0),
            );
            return;
          }
          const key = new Date(dateValue).toISOString().slice(0, 10);
          chartByDay.set(
            key,
            (chartByDay.get(key) || 0) + (record.amount || 0),
          );
        });

        const chartEntries = Array.from(chartByDay.entries())
          .sort(([a], [b]) => a.localeCompare(b))
          .slice(-7)
          .map(([key, spend]) => ({
            name: key.startsWith("Txn")
              ? key
              : new Date(key).toLocaleDateString("en-US", { weekday: "short" }),
            spend,
          }));

        setStats({
          total_spend: totals.total_spend,
          total_liters: totals.total_liters,
          avg_price:
            totals.total_liters > 0
              ? totals.total_spend / totals.total_liters
              : avgFromPoints,
          anomalies: anomalyRecords.length,
        });
        setChartData(chartEntries);
        setAlerts(anomalyRecords.slice(0, 3));
      } catch (err: any) {
        console.error("Failed to fetch fuel data", err);
        setError(err?.message || "Failed to load fuel data");
        setStats({
          total_spend: 0,
          total_liters: 0,
          avg_price: 0,
          anomalies: 0,
        });
        setChartData([]);
        setAlerts([]);
      } finally {
        setLoading(false);
      }
    };

    loadFuelData();
  }, []);

  return (
    <div className="p-6 space-y-6">
      <header className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-white">
            {uiStrings?.fuel_title || "Fuel Management"}
          </h1>
          <p className="text-text-muted">
            {uiStrings?.fuel_subtitle ||
              "Monitor fuel efficiency and detect anomalies."}
          </p>
        </div>
        <button className="bg-primary hover:bg-primary-hover text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2">
          <span className="material-symbols-outlined">upload_file</span>
          {uiStrings?.btn_import || "Import Fuel Data"}
        </button>
      </header>

      {error && (
        <div className="rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">
          {error}
        </div>
      )}

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-surface-dark p-4 rounded-xl border border-border-color">
          <div className="text-text-muted text-sm mb-1">
            {uiStrings?.stat_total_spend || "Total Fuel Spend"}
          </div>
          <div className="text-2xl font-bold text-white">
            ${stats.total_spend.toLocaleString()}
          </div>
          <div className="text-green-500 text-xs mt-1 flex items-center">
            <span className="material-symbols-outlined text-[16px] mr-1">
              trending_down
            </span>
            {loading ? "Loading trend..." : "Rolling 7-day spend"}
          </div>
        </div>

        <div className="bg-surface-dark p-4 rounded-xl border border-border-color">
          <div className="text-text-muted text-sm mb-1">
            {uiStrings?.stat_liters || "Liters Consumed"}
          </div>
          <div className="text-2xl font-bold text-white">
            {stats.total_liters.toLocaleString()} L
          </div>
        </div>

        <div className="bg-surface-dark p-4 rounded-xl border border-border-color">
          <div className="text-text-muted text-sm mb-1">
            {uiStrings?.stat_price || "Avg Price / Liter"}
          </div>
          <div className="text-2xl font-bold text-white">
            ${stats.avg_price.toFixed(2)}
          </div>
        </div>

        <div className="bg-surface-dark p-4 rounded-xl border border-border-color relative overflow-hidden">
          <div className="absolute top-0 right-0 p-4 opacity-10">
            <span className="material-symbols-outlined text-[64px] text-red-500">
              warning
            </span>
          </div>
          <div className="text-text-muted text-sm mb-1">
            {uiStrings?.stat_anomalies || "Anomalies Detected"}
          </div>
          <div className="text-2xl font-bold text-red-500">
            {stats.anomalies}
          </div>
          <div className="text-text-muted text-xs mt-1">
            Requires immediate review
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Chart Section */}
        <div className="lg:col-span-2 bg-surface-dark p-6 rounded-xl border border-border-color">
          <h3 className="text-lg font-bold text-white mb-6">
            {uiStrings?.chart_title || "Fuel Spend Trend"}
          </h3>
          <div className="h-[300px] w-full">
            {loading ? (
              <div className="h-full w-full flex items-center justify-center text-text-muted text-sm">
                Loading chart...
              </div>
            ) : chartData.length === 0 ? (
              <div className="h-full w-full flex items-center justify-center text-text-muted text-sm">
                No fuel transactions available.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData}>
                  <defs>
                    <linearGradient id="colorSpend" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid
                    strokeDasharray="3 3"
                    stroke="#334155"
                    vertical={false}
                  />
                  <XAxis
                    dataKey="name"
                    stroke="#94a3b8"
                    tick={{ fill: "#94a3b8" }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    stroke="#94a3b8"
                    tick={{ fill: "#94a3b8" }}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={(value) => `$${value}`}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "#1e293b",
                      borderColor: "#334155",
                      color: "#fff",
                    }}
                    itemStyle={{ color: "#fff" }}
                    formatter={(value: any) => [`$${value}`, "Spend"]}
                  />
                  <Area
                    type="monotone"
                    dataKey="spend"
                    stroke="#3b82f6"
                    strokeWidth={2}
                    fillOpacity={1}
                    fill="url(#colorSpend)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        {/* Anomalies List */}
        <div className="bg-surface-dark p-6 rounded-xl border border-border-color">
          <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
            {uiStrings?.alerts_title || "Recent Alerts"}
          </h3>
          <div className="space-y-4">
            {loading ? (
              <div className="text-xs text-text-muted">Loading alerts...</div>
            ) : alerts.length === 0 ? (
              <div className="text-xs text-text-muted">
                No anomalies detected.
              </div>
            ) : (
              alerts.map((alert) => {
                const timestamp =
                  alert.occurred_at || alert.created_at || alert.timestamp;
                const vehicleLabel =
                  alert.vehicle_name ||
                  alert.vehicle_id ||
                  alert.id ||
                  "Unknown vehicle";
                return (
                  <div
                    key={alert.id || `${vehicleLabel}-${timestamp}`}
                    className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg"
                  >
                    <div className="flex justify-between items-start mb-1">
                      <span className="text-xs font-bold text-red-400">
                        {alert.alert_type || "ANOMALY"}
                      </span>
                      <span className="text-[10px] text-red-400/70">
                        {timestamp
                          ? new Date(timestamp).toLocaleString()
                          : "Recent"}
                      </span>
                    </div>
                    <p className="text-sm text-white font-medium">
                      {vehicleLabel}
                    </p>
                    <p className="text-xs text-text-muted mt-1">
                      Spend: ${Number(alert.amount || 0).toLocaleString()} ·
                      Volume:{" "}
                      {Number(
                        alert.liters || alert.volume || alert.fuel_liters || 0,
                      ).toLocaleString()}{" "}
                      L
                    </p>
                  </div>
                );
              })
            )}
          </div>
          <button className="w-full mt-4 py-2 text-sm text-primary hover:text-primary-hover font-medium transition-colors">
            {uiStrings?.btn_view_alerts || "View All Alerts"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default FuelDashboard;
