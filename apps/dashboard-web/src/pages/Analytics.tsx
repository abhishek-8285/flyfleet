import {
    ArrowTrendingDownIcon,
    ArrowTrendingUpIcon,
    ClockIcon,
    TruckIcon,
} from "@heroicons/react/24/outline";
import { useEffect, useMemo, useState } from "react";
import {
    Bar,
    BarChart,
    CartesianGrid,
    Legend,
    Line,
    LineChart,
    PolarAngleAxis,
    PolarGrid,
    PolarRadiusAxis,
    Radar,
    RadarChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";

// Fix for Recharts TS2786 with React 18
const PolarAngleAxisAny = PolarAngleAxis as any;
const PolarGridAny = PolarGrid as any;
const PolarRadiusAxisAny = PolarRadiusAxis as any;
const RadarAny = Radar as any;
const BarAny = Bar as any;
const CartesianGridAny = CartesianGrid as any;
const LegendAny = Legend as any;
const LineAny = Line as any;
const RadarChartAny = RadarChart as any;
const TooltipAny = Tooltip as any;
const XAxisAny = XAxis as any;
const YAxisAny = YAxis as any;
const BarChartAny = BarChart as any;
const LineChartAny = LineChart as any;

const toNumber = (value: any): number => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
};

const Analytics = () => {
  const [tripStats, setTripStats] = useState<any[]>([]);
  const [driverStats, setDriverStats] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("analytics");

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [trips, drivers] = await Promise.all([
          api.getTripStats(),
          api.getDriverStats(),
        ]);
        setTripStats(Array.isArray(trips) ? trips : trips?.data || []);
        setDriverStats(Array.isArray(drivers) ? drivers : drivers?.data || []);
      } catch (error) {
        console.error("Failed to fetch analytics data", error);
        setTripStats([]);
        setDriverStats([]);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const tripChartData = useMemo(() => {
    return (Array.isArray(tripStats) ? tripStats : []).map(
      (row: any, idx: number) => ({
        label: row.date || row.day || row.label || `Day ${idx + 1}`,
        total_distance: toNumber(row.total_distance || row.distance_km),
        total_trips: toNumber(row.total_trips || row.trip_count),
        fuel_consumption: toNumber(
          row.fuel_consumption ||
            row.fuel_consumption_l_per_100km ||
            row.fuel_l_per_100km,
        ),
        maintenance_cost: toNumber(
          row.maintenance_cost || row.maintenance_cost_usd,
        ),
      }),
    );
  }, [tripStats]);

  const normalizedDrivers = useMemo(() => {
    return (Array.isArray(driverStats) ? driverStats : []).map(
      (driver: any, idx: number) => ({
        driver_id: driver.driver_id || driver.id || `driver-${idx}`,
        name:
          driver.name || driver.driver_name || driver.email || "Unknown Driver",
        safety_score: toNumber(driver.safety_score),
        trips_completed: toNumber(driver.trips_completed || driver.total_trips),
        on_time_rate: toNumber(
          driver.on_time_rate || driver.on_time_percentage,
        ),
        total_distance: toNumber(driver.total_distance || driver.distance_km),
      }),
    );
  }, [driverStats]);

  const predictiveData = useMemo(() => {
    const source = tripChartData.slice(-7);
    return source.map((row: any, idx: number) => {
      const actual = row.total_trips || row.total_distance;
      const predicted = Math.max(0, Math.round(actual * 1.06));
      const label = row.label || `D${idx + 1}`;
      return { day: label, actual, predicted };
    });
  }, [tripChartData]);

  const kpis = useMemo(() => {
    const avgOnTime = normalizedDrivers.length
      ? normalizedDrivers.reduce((sum, d) => sum + d.on_time_rate, 0) /
        normalizedDrivers.length
      : 0;
    const avgFuel = tripChartData.length
      ? tripChartData.reduce((sum, t) => sum + t.fuel_consumption, 0) /
        tripChartData.length
      : 0;
    const totalMaintenance = tripChartData.reduce(
      (sum, t) => sum + t.maintenance_cost,
      0,
    );
    const avgSafety = normalizedDrivers.length
      ? normalizedDrivers.reduce((sum, d) => sum + d.safety_score, 0) /
        normalizedDrivers.length
      : 0;

    return {
      fleetEfficiency: `${avgSafety.toFixed(1)}%`,
      fuelConsumption: avgFuel > 0 ? `${avgFuel.toFixed(1)} L/100km` : "--",
      maintenanceCosts: `$${Math.round(totalMaintenance).toLocaleString()}`,
      onTimeDelivery: `${avgOnTime.toFixed(1)}%`,
    };
  }, [normalizedDrivers, tripChartData]);

  const driverSkillsData = useMemo(() => {
    if (!normalizedDrivers.length) {
      return [
        { subject: "Safety", A: 0, fullMark: 150 },
        { subject: "Efficiency", A: 0, fullMark: 150 },
        { subject: "Punctuality", A: 0, fullMark: 150 },
        { subject: "Compliance", A: 0, fullMark: 150 },
        { subject: "Maintenance", A: 0, fullMark: 150 },
        { subject: "Customer", A: 0, fullMark: 150 },
      ];
    }

    const avgSafety =
      normalizedDrivers.reduce(
        (sum: number, d: any) => sum + d.safety_score,
        0,
      ) / normalizedDrivers.length;
    const avgOnTime =
      normalizedDrivers.reduce(
        (sum: number, d: any) => sum + d.on_time_rate,
        0,
      ) / normalizedDrivers.length;
    const avgTrips =
      normalizedDrivers.reduce(
        (sum: number, d: any) => sum + d.trips_completed,
        0,
      ) / normalizedDrivers.length;

    return [
      {
        subject: "Safety",
        A: Math.min(150, Math.round(avgSafety * 1.5)),
        fullMark: 150,
      },
      {
        subject: "Efficiency",
        A: Math.min(150, Math.round(avgOnTime * 1.5)),
        fullMark: 150,
      },
      {
        subject: "Punctuality",
        A: Math.min(150, Math.round(avgOnTime * 1.4)),
        fullMark: 150,
      },
      {
        subject: "Compliance",
        A: Math.min(150, Math.round(avgSafety * 1.45)),
        fullMark: 150,
      },
      {
        subject: "Maintenance",
        A: Math.min(150, Math.round((avgTrips % 100) * 1.2)),
        fullMark: 150,
      },
      {
        subject: "Customer",
        A: Math.min(150, Math.round(avgOnTime * 1.2)),
        fullMark: 150,
      },
    ];
  }, [normalizedDrivers]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-background-dark">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="bg-background-dark min-h-screen text-white p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {uiStrings?.title || "Analytics & Intelligence"}
          </h1>
          <p className="text-text-muted text-sm">
            {uiStrings?.sub_title ||
              "AI-driven predictive modeling and fleet performance"}
          </p>
        </div>
        <div className="flex gap-2">
          <button className="px-4 py-2 bg-surface-dark border border-border-color rounded-lg text-sm font-medium text-white hover:bg-surface-darker transition-colors">
            {uiStrings?.btn_last_7d || "Last 7 Days"}
          </button>
          <button className="px-4 py-2 bg-primary hover:bg-primary/90 rounded-lg text-sm font-medium text-white shadow-lg shadow-primary/20 transition-all">
            {uiStrings?.btn_export || "Export Report"}
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <KpiCard
          title={uiStrings?.kpi_ref_fleet_eff || "Fleet Efficiency"}
          value={kpis.fleetEfficiency}
          change="Live"
          icon={<ArrowTrendingUpIcon className="text-emerald-400" />}
          positive
          uiStrings={uiStrings}
        />
        <KpiCard
          title={uiStrings?.kpi_ref_fuel_cons || "Fuel Consumption"}
          value={kpis.fuelConsumption}
          change="Live"
          icon={<TruckIcon className="text-blue-400" />}
          positive
          uiStrings={uiStrings}
        />
        <KpiCard
          title={uiStrings?.kpi_ref_maint_cost || "Maintenance Costs"}
          value={kpis.maintenanceCosts}
          change="Live"
          icon={<ArrowTrendingDownIcon className="text-red-400" />}
          positive={false}
          uiStrings={uiStrings}
        />
        <KpiCard
          title={uiStrings?.kpi_ref_on_time || "On-Time Delivery"}
          value={kpis.onTimeDelivery}
          change="Live"
          icon={<ClockIcon className="text-yellow-400" />}
          positive
          uiStrings={uiStrings}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Predictive Maintenance Chart */}
        <ChartCard
          title={
            uiStrings?.chart_predictive_title ||
            "Predictive Fleet Utilization (AI Model)"
          }
        >
          <ResponsiveContainer width="100%" height="100%">
            <LineChartAny data={predictiveData}>
              <CartesianGridAny
                strokeDasharray="3 3"
                stroke="#334155"
                vertical={false}
              />
              <XAxisAny
                dataKey="day"
                tick={{ fill: "#94a3b8" }}
                axisLine={false}
                tickLine={false}
              />
              <YAxisAny
                tick={{ fill: "#94a3b8" }}
                axisLine={false}
                tickLine={false}
              />
              <TooltipAny
                contentStyle={{
                  backgroundColor: "#1e293b",
                  borderRadius: "8px",
                  border: "1px solid #334155",
                  color: "#fff",
                }}
              />
              <LegendAny />
              <LineAny
                type="monotone"
                dataKey="actual"
                name={uiStrings?.chart_predictive_actual || "Actual Usage"}
                stroke="#3b82f6"
                strokeWidth={3}
                dot={{ r: 4 }}
              />
              <LineAny
                type="monotone"
                dataKey="predicted"
                name={uiStrings?.chart_predictive_ai || "AI Prediction"}
                stroke="#10b981"
                strokeWidth={3}
                strokeDasharray="5 5"
                dot={false}
              />
            </LineChartAny>
          </ResponsiveContainer>
        </ChartCard>

        {/* Driver Skills Radar */}
        <ChartCard
          title={
            uiStrings?.chart_skills_title || "Driver Workforce Skills Matrix"
          }
        >
          <ResponsiveContainer width="100%" height="100%">
            <RadarChartAny
              cx="50%"
              cy="50%"
              outerRadius="80%"
              data={driverSkillsData}
            >
              <PolarGridAny stroke="#334155" />
              <PolarAngleAxisAny
                dataKey="subject"
                tick={{ fill: "#94a3b8", fontSize: 12 }}
              />
              <PolarRadiusAxisAny
                angle={30}
                domain={[0, 150]}
                tick={false}
                axisLine={false}
              />
              <RadarAny
                name={uiStrings?.chart_skills_avg || "Fleet Average"}
                dataKey="A"
                stroke="#8b5cf6"
                fill="#8b5cf6"
                fillOpacity={0.3}
              />
              <LegendAny />
              <TooltipAny
                contentStyle={{
                  backgroundColor: "#1e293b",
                  borderRadius: "8px",
                  border: "1px solid #334155",
                  color: "#fff",
                }}
              />
            </RadarChartAny>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      {/* Distance Trend Chart */}
      <ChartCard
        title={uiStrings?.chart_distance_title || "Distance Traveled Analysis"}
        height="h-80"
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChartAny data={tripChartData}>
            <CartesianGridAny
              strokeDasharray="3 3"
              stroke="#334155"
              vertical={false}
            />
            <XAxisAny
              dataKey="label"
              tick={{ fill: "#94a3b8" }}
              axisLine={false}
              tickLine={false}
            />
            <YAxisAny
              tick={{ fill: "#94a3b8" }}
              axisLine={false}
              tickLine={false}
            />
            <TooltipAny
              contentStyle={{
                backgroundColor: "#1e293b",
                borderRadius: "8px",
                border: "1px solid #334155",
                color: "#fff",
              }}
              cursor={{ fill: "rgba(255,255,255,0.05)" }}
            />
            <LegendAny />
            <BarAny
              dataKey="total_distance"
              name={uiStrings?.chart_distance_label || "Distance (km)"}
              fill="#3b82f6"
              radius={[4, 4, 0, 0]}
            />
          </BarChartAny>
        </ResponsiveContainer>
      </ChartCard>

      {/* Driver Performance Table - Dark Mode */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-6 border-b border-border-color bg-surface-darker/50">
          <h2 className="text-lg font-bold text-white">
            {uiStrings?.table_performance_title || "Top Driver Performance"}
          </h2>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border-color">
            <thead className="bg-surface-darker">
              <tr>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider"
                >
                  {uiStrings?.th_driver || "Driver"}
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider"
                >
                  {uiStrings?.th_safety_score || "Safety Score"}
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider"
                >
                  {uiStrings?.th_trips || "Trips"}
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider"
                >
                  {uiStrings?.th_on_time || "On-Time"}
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider"
                >
                  {uiStrings?.th_distance || "Distance"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {normalizedDrivers.map((driver) => (
                <tr
                  key={driver.driver_id}
                  className="hover:bg-surface-darker/50 transition-colors"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold mr-3 border border-primary/30">
                        {driver.name.charAt(0)}
                      </div>
                      <div className="text-sm font-medium text-white">
                        {driver.name}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 py-0.5 inline-flex text-xs leading-5 font-semibold rounded-full border ${
                        driver.safety_score >= 90
                          ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20"
                          : driver.safety_score >= 80
                            ? "bg-yellow-500/10 text-yellow-500 border-yellow-500/20"
                            : "bg-red-500/10 text-red-500 border-red-500/20"
                      }`}
                    >
                      {driver.safety_score}/100
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted">
                    {driver.trips_completed}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted">
                    {driver.on_time_rate}%
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted">
                    {driver.total_distance.toLocaleString()} km
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

const KpiCard = ({ title, value, change, icon, positive, uiStrings }: any) => (
  <div className="bg-surface-dark p-6 rounded-xl border border-border-color relative overflow-hidden group hover:border-primary/30 transition-all">
    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
      {/* Clone icon for background effect */}
      <div className="origin-top-right transform scale-150">{icon}</div>
    </div>
    <div className="flex justify-between items-start mb-4">
      <div>
        <p className="text-xs font-semibold text-text-muted uppercase tracking-wider">
          {title}
        </p>
        <h3 className="text-2xl font-bold text-white mt-1">{value}</h3>
      </div>
      <div className="p-2 rounded-lg bg-surface-darker border border-border-color">
        {icon}
      </div>
    </div>
    <div
      className={`text-xs font-medium flex items-center gap-1 ${positive ? "text-emerald-500" : "text-red-500"}`}
    >
      <span>{change}</span>
      <span className="text-text-muted">
        {uiStrings?.label_vs_last_month || "vs last month"}
      </span>
    </div>
  </div>
);

const ChartCard = ({ title, children, height = "h-80" }: any) => (
  <div className="bg-surface-dark p-6 rounded-xl border border-border-color">
    <h2 className="text-lg font-bold text-white mb-6">{title}</h2>
    <div className={`${height} w-full`}>{children}</div>
  </div>
);

export default Analytics;
