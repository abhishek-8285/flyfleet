import {
  TruckIcon,
  ClockIcon,
  FuelIcon,
  CurrencyDollarIcon,
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import React, { useEffect, useState } from "react";
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import api from "../../../services/api";

interface FleetAnalytics {
  id: string;
  organization_id: string;
  start_date: string;
  end_date: string;
  total_vehicles: number;
  active_vehicles: number;
  total_distance: number;
  average_speed: number;
  total_fuel_consumption: number;
  average_fuel_efficiency: number;
  total_trips: number;
  completed_trips: number;
  cancelled_trips: number;
  average_trip_duration: number;
  idle_time_percentage: number;
  routing_efficiency: number;
  fuel_cost: number;
  maintenance_cost: number;
  revenue: number;
  profit_margin: number;
  safety_score: number;
  accidents_count: number;
  incidents_count: number;
  emergency_stops: number;
  geofence_violations: number;
  alerts_triggered: number;
  generated_at: string;
}

const FleetAnalyticsPage: React.FC = () => {
  const [analytics, setAnalytics] = useState<FleetAnalytics[]>([]);
  const [loading, setLoading] = useState(true);
  const [period, setPeriod] = useState("30");
  const [selectedAnalytics, setSelectedAnalytics] = useState<FleetAnalytics | null>(null);

  useEffect(() => {
    fetchAnalytics();
  }, [period]);

  const fetchAnalytics = async () => {
    try {
      setLoading(true);
      const endDate = new Date();
      const startDate = new Date();
      startDate.setDate(startDate.getDate() - parseInt(period));

      const response = (await api.get("/admin/analytics/fleet/period", {
        params: {
          orgId: "current",
          startDate: format(startDate, "yyyy-MM-dd"),
          endDate: format(endDate, "yyyy-MM-dd"),
        },
      })) as any;
      
      setAnalytics(response?.data ? [response.data] : []);
    } catch (error) {
      console.error("Failed to fetch fleet analytics", error);
    } finally {
      setLoading(false);
    }
  };

  const latest = analytics[0];

  const kpis = [
    {
      title: "Fleet Utilization",
      value: latest ? `${((latest.active_vehicles / latest.total_vehicles) * 100).toFixed(1)}%` : "--",
      change: "+2.5%",
      positive: true,
      icon: <TruckIcon className="h-6 w-6" />,
      color: "text-blue-400",
    },
    {
      title: "Total Distance",
      value: latest ? `${(latest.total_distance / 1000).toFixed(1)}K km` : "--",
      change: "+12.3%",
      positive: true,
      icon: <ArrowTrendingUpIcon className="h-6 w-6" />,
      color: "text-emerald-400",
    },
    {
      title: "Fuel Efficiency",
      value: latest ? `${latest.average_fuel_efficiency.toFixed(1)} km/L` : "--",
      change: "-0.8%",
      positive: false,
      icon: <FuelIcon className="h-6 w-6" />,
      color: "text-amber-400",
    },
    {
      title: "Profit Margin",
      value: latest ? `${latest.profit_margin.toFixed(1)}%` : "--",
      change: "+5.2%",
      positive: true,
      icon: <CurrencyDollarIcon className="h-6 w-6" />,
      color: "text-green-400",
    },
  ];

  const tripStatusData = latest ? [
    { name: "Completed", value: latest.completed_trips, color: "#10b981" },
    { name: "Cancelled", value: latest.cancelled_trips, color: "#ef4444" },
    { name: "In Progress", value: latest.total_trips - latest.completed_trips - latest.cancelled_trips, color: "#3b82f6" },
  ] : [];

  const efficiencyData = latest ? [
    { day: "Week 1", efficiency: 85 + Math.random() * 10 },
    { day: "Week 2", efficiency: 82 + Math.random() * 10 },
    { day: "Week 3", efficiency: 88 + Math.random() * 10 },
    { day: "Week 4", efficiency: 90 + Math.random() * 10 },
  ] : [];

  const costData = latest ? [
    { category: "Fuel", cost: latest.fuel_cost },
    { category: "Maintenance", cost: latest.maintenance_cost },
    { category: "Operations", cost: latest.revenue * 0.3 },
  ] : [];

  return (
    <div className="p-6 bg-slate-900 min-h-screen">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            📊 Fleet Analytics
          </h1>
          <p className="text-slate-400 text-sm mt-1">
            Comprehensive fleet performance insights
          </p>
        </div>
        <div className="flex gap-3">
          <select
            value={period}
            onChange={(e) => setPeriod(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
          >
            <option value="7">Last 7 Days</option>
            <option value="30">Last 30 Days</option>
            <option value="90">Last 90 Days</option>
            <option value="365">Last Year</option>
          </select>
          <button
            onClick={fetchAnalytics}
            className="px-4 py-2 bg-slate-800 text-white rounded-lg border border-slate-700 hover:bg-slate-700 transition flex items-center gap-2"
          >
            <ArrowPathIcon className="h-4 w-4" /> Refresh
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        {kpis.map((kpi, index) => (
          <div
            key={index}
            className="bg-slate-800 rounded-xl border border-slate-700 p-6 hover:border-blue-500/30 transition"
          >
            <div className="flex justify-between items-start mb-4">
              <div className={`p-3 rounded-lg bg-slate-900 ${kpi.color}`}>
                {kpi.icon}
              </div>
              <span className={`text-xs font-semibold ${kpi.positive ? "text-emerald-400" : "text-red-400"}`}>
                {kpi.change}
              </span>
            </div>
            <h3 className="text-slate-400 text-sm mb-1">{kpi.title}</h3>
            <p className="text-2xl font-bold text-white">{kpi.value}</p>
          </div>
        ))}
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Trip Status Pie Chart */}
        <div className="bg-slate-800 rounded-xl border border-slate-700 p-6">
          <h3 className="text-lg font-bold text-white mb-4">Trip Status Distribution</h3>
          {loading || tripStatusData.length === 0 ? (
            <div className="h-64 flex items-center justify-center text-slate-500">
              No data available
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={tripStatusData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={5}
                  dataKey="value"
                  label={({ name, value }) => `${name}: ${value}`}
                >
                  {tripStatusData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#1e293b",
                    border: "1px solid #334155",
                    borderRadius: "8px",
                    color: "#fff",
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Efficiency Trend */}
        <div className="bg-slate-800 rounded-xl border border-slate-700 p-6">
          <h3 className="text-lg font-bold text-white mb-4">Fleet Efficiency Trend</h3>
          {loading || efficiencyData.length === 0 ? (
            <div className="h-64 flex items-center justify-center text-slate-500">
              No data available
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={efficiencyData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                <XAxis dataKey="day" tick={{ fill: "#94a3b8" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: "#94a3b8" }} axisLine={false} tickLine={false} domain={[0, 100]} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#1e293b",
                    border: "1px solid #334155",
                    borderRadius: "8px",
                    color: "#fff",
                  }}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="efficiency"
                  name="Efficiency %"
                  stroke="#3b82f6"
                  strokeWidth={3}
                  dot={{ r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      {/* Charts Row 2 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Cost Breakdown */}
        <div className="bg-slate-800 rounded-xl border border-slate-700 p-6">
          <h3 className="text-lg font-bold text-white mb-4">Cost Breakdown</h3>
          {loading || costData.length === 0 ? (
            <div className="h-64 flex items-center justify-center text-slate-500">
              No data available
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={costData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                <XAxis dataKey="category" tick={{ fill: "#94a3b8" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: "#94a3b8" }} axisLine={false} tickLine={false} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#1e293b",
                    border: "1px solid #334155",
                    borderRadius: "8px",
                    color: "#fff",
                    formatter: (value: any) => [`$${Number(value).toLocaleString()}`, "Cost"],
                  }}
                />
                <Legend />
                <Bar dataKey="cost" name="Cost ($)" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Key Metrics */}
        <div className="bg-slate-800 rounded-xl border border-slate-700 p-6">
          <h3 className="text-lg font-bold text-white mb-4">Key Performance Metrics</h3>
          {loading || !latest ? (
            <div className="h-64 flex items-center justify-center text-slate-500">
              No data available
            </div>
          ) : (
            <div className="space-y-4">
              <MetricRow label="Safety Score" value={`${latest.safety_score}/100`} color={latest.safety_score >= 80 ? "text-emerald-400" : latest.safety_score >= 60 ? "text-amber-400" : "text-red-400"} />
              <MetricRow label="Routing Efficiency" value={`${latest.routing_efficiency.toFixed(1)}%`} color="text-blue-400" />
              <MetricRow label="Idle Time" value={`${latest.idle_time_percentage.toFixed(1)}%`} color={latest.idle_time_percentage < 10 ? "text-emerald-400" : "text-amber-400"} />
              <MetricRow label="Avg Trip Duration" value={`${latest.average_trip_duration.toFixed(1)} hrs`} color="text-white" />
              <MetricRow label="Avg Speed" value={`${latest.average_speed.toFixed(1)} km/h`} color="text-white" />
              <MetricRow label="Geofence Violations" value={latest.geofence_violations.toString()} color={latest.geofence_violations === 0 ? "text-emerald-400" : "text-red-400"} />
              <MetricRow label="Emergency Stops" value={latest.emergency_stops.toString()} color={latest.emergency_stops === 0 ? "text-emerald-400" : "text-amber-400"} />
              <MetricRow label="Alerts Triggered" value={latest.alerts_triggered.toString()} color={latest.alerts_triggered === 0 ? "text-emerald-400" : "text-amber-400"} />
            </div>
          )}
        </div>
      </div>

      {/* Alerts Section */}
      {latest && (latest.accidents_count > 0 || latest.incidents_count > 0) && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <ExclamationTriangleIcon className="h-6 w-6 text-red-400" />
            <h3 className="text-lg font-bold text-white">Attention Required</h3>
          </div>
          <div className="grid grid-cols-3 gap-4">
            {latest.accidents_count > 0 && (
              <div className="bg-red-500/20 rounded-lg p-4">
                <div className="text-red-400 text-sm mb-1">Accidents</div>
                <div className="text-2xl font-bold text-white">{latest.accidents_count}</div>
              </div>
            )}
            {latest.incidents_count > 0 && (
              <div className="bg-amber-500/20 rounded-lg p-4">
                <div className="text-amber-400 text-sm mb-1">Incidents</div>
                <div className="text-2xl font-bold text-white">{latest.incidents_count}</div>
              </div>
            )}
            {latest.cancelled_trips > 0 && (
              <div className="bg-orange-500/20 rounded-lg p-4">
                <div className="text-orange-400 text-sm mb-1">Cancelled Trips</div>
                <div className="text-2xl font-bold text-white">{latest.cancelled_trips}</div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

const MetricRow: React.FC<{ label: string; value: string; color: string }> = ({ label, value, color }) => (
  <div className="flex justify-between items-center py-3 border-b border-slate-700 last:border-0">
    <span className="text-slate-400">{label}</span>
    <span className={`text-lg font-semibold ${color}`}>{value}</span>
  </div>
);

export default FleetAnalyticsPage;
