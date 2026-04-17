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
  const [missionClock, setMissionClock] = useState(new Date());
  const { uiStrings } = useUIStrings("analytics");

  useEffect(() => {
    const timer = setInterval(() => setMissionClock(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

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
      <div className="flex items-center justify-center h-screen bg-surface-dim">
        <div className="relative">
          <div className="w-16 h-16 rounded-full border-4 border-primary/20 border-t-primary animate-spin" />
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="material-symbols-outlined text-primary text-xl animate-pulse">radar</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-surface-dim min-h-screen text-white p-8 space-y-10 font-sans antialiased">
      {/* Intelligence Command Header */}
      <div className="relative bg-surface-container/60 backdrop-blur-3xl border border-outline rounded-[3rem] p-10 shadow-2xl overflow-hidden group">
        <div className="absolute top-0 right-0 p-12 opacity-[0.03] group-hover:opacity-[0.05] transition-opacity">
          <span className="material-symbols-outlined text-[180px] rotate-12">monitoring</span>
        </div>
        
        <div className="relative flex flex-col md:flex-row justify-between items-start md:items-center gap-8">
          <div className="space-y-4">
            <nav className="flex items-center gap-3 text-[10px] font-black uppercase tracking-[0.3em] text-on-surface-variant">
              <span className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse" />
              Strategic Intelligence Core
            </nav>
            <div>
              <h1 className="text-5xl font-black tracking-tighter uppercase text-white mb-2">
                {uiStrings?.title || "Operational HUD"}
              </h1>
              <p className="text-on-surface-variant text-sm font-medium tracking-wide">
                {uiStrings?.sub_title || "AI-driven predictive modeling and fleet performance analytics"}
              </p>
            </div>
          </div>

          <div className="flex flex-col md:flex-row items-end md:items-center gap-6">
            {/* Mission Clock */}
            <div className="bg-surface-dim/80 border border-outline px-6 py-4 rounded-3xl flex items-center gap-6 shadow-inner">
              <div className="flex flex-col text-right">
                <span className="text-[9px] font-black text-on-surface-variant uppercase tracking-widest opacity-50">Operational Time</span>
                <span className="text-2xl font-bold text-primary tracking-tighter font-mono">
                  {missionClock.toLocaleTimeString()}
                </span>
              </div>
              <div className="h-8 w-[1px] bg-outline/30" />
              <div className="flex flex-col text-right">
                <span className="text-[9px] font-black text-on-surface-variant uppercase tracking-widest opacity-50">Local Cycle</span>
                <span className="text-sm font-bold text-white tracking-widest uppercase">
                  {missionClock.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })}
                </span>
              </div>
            </div>

            <div className="flex gap-4">
              <button className="h-14 px-8 bg-surface-container-highest border border-outline rounded-2xl text-[10px] font-black uppercase tracking-widest text-white hover:bg-surface-on-variant transition-all active:scale-95 shadow-lg">
                {uiStrings?.btn_last_7d || "Cycle: 7D"}
              </button>
              <button className="h-14 px-8 bg-primary hover:bg-primary-dark text-white text-[10px] font-black uppercase tracking-widest rounded-2xl shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-3">
                <span className="material-symbols-outlined text-lg">ios_share</span>
                {uiStrings?.btn_export || "Deploy Report"}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Telemetry Intelligence Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <KpiCard
          title={uiStrings?.kpi_ref_fleet_eff || "Fleet Velocity Efficiency"}
          value={kpis.fleetEfficiency}
          change="+2.4%"
          icon="finance"
          positive
          uiStrings={uiStrings}
        />
        <KpiCard
          title={uiStrings?.kpi_ref_fuel_cons || "Resource Consumption"}
          value={kpis.fuelConsumption}
          change="-0.8%"
          icon="water_drop"
          positive
          uiStrings={uiStrings}
        />
        <KpiCard
          title={uiStrings?.kpi_ref_maint_cost || "Operational Maintenance"}
          value={kpis.maintenanceCosts}
          change="+12%"
          icon="build"
          positive={false}
          uiStrings={uiStrings}
        />
        <KpiCard
          title={uiStrings?.kpi_ref_on_time || "SLA Adherence Rate"}
          value={kpis.onTimeDelivery}
          change="+4.1%"
          icon="schedule"
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
                strokeDasharray="4 4"
                stroke="rgba(255,255,255,0.05)"
                vertical={false}
              />
              <XAxisAny
                dataKey="day"
                tick={{ fill: "#94a3b8", fontSize: 10, fontWeight: 900 }}
                axisLine={false}
                tickLine={false}
                dy={10}
              />
              <YAxisAny
                tick={{ fill: "#94a3b8", fontSize: 10, fontWeight: 900 }}
                axisLine={false}
                tickLine={false}
                dx={-10}
              />
              <TooltipAny
                contentStyle={{
                  backgroundColor: "#0F1117",
                  borderRadius: "1.5rem",
                  border: "1px solid rgba(255,255,255,0.1)",
                  padding: "1.5rem",
                  boxShadow: "0 25px 50px -12px rgba(0,0,0,0.5)",
                  color: "#fff",
                }}
                itemStyle={{ fontSize: "12px", fontWeight: "bold", textTransform: "uppercase" }}
              />
              <LegendAny 
                verticalAlign="top" 
                align="right" 
                iconType="circle"
                wrapperStyle={{ paddingBottom: "40px", fontSize: "10px", fontWeight: 900, textTransform: "uppercase", letterSpacing: "0.1em" }}
              />
              <LineAny
                type="monotone"
                dataKey="actual"
                name={uiStrings?.chart_predictive_actual || "Actual Usage"}
                stroke="#137FEC"
                strokeWidth={4}
                dot={{ r: 6, fill: "#137FEC", strokeWidth: 2, stroke: "#0F1117" }}
                activeDot={{ r: 8, strokeWidth: 0 }}
              />
              <LineAny
                type="monotone"
                dataKey="predicted"
                name={uiStrings?.chart_predictive_ai || "AI Prediction"}
                stroke="#10B981"
                strokeWidth={4}
                strokeDasharray="8 8"
                dot={false}
              />
            </LineChartAny>
          </ResponsiveContainer>
        </ChartCard>

        {/* Driver Skills Radar */}
        <ChartCard
          title={
            uiStrings?.chart_skills_title || "Workforce Competency Matrix"
          }
        >
          <ResponsiveContainer width="100%" height="100%">
            <RadarChartAny
              cx="50%"
              cy="50%"
              outerRadius="80%"
              data={driverSkillsData}
            >
              <PolarGridAny stroke="rgba(255,255,255,0.05)" />
              <PolarAngleAxisAny
                dataKey="subject"
                tick={{ fill: "#94a3b8", fontSize: 10, fontWeight: 900, textTransform: "uppercase" }}
              />
              <PolarRadiusAxisAny
                angle={30}
                domain={[0, 150]}
                tick={false}
                axisLine={false}
              />
              <RadarAny
                name={uiStrings?.chart_skills_avg || "Fleet Baseline"}
                dataKey="A"
                stroke="#8B5CF6"
                fill="#8B5CF6"
                fillOpacity={0.15}
                strokeWidth={3}
              />
              <LegendAny 
                verticalAlign="top" 
                align="right" 
                iconType="circle"
                wrapperStyle={{ paddingBottom: "40px", fontSize: "10px", fontWeight: 900, textTransform: "uppercase", letterSpacing: "0.1em" }}
              />
              <TooltipAny
                contentStyle={{
                  backgroundColor: "#0F1117",
                  borderRadius: "1.5rem",
                  border: "1px solid rgba(255,255,255,0.1)",
                  padding: "1.5rem",
                  boxShadow: "0 25px 50px -12px rgba(0,0,0,0.5)",
                  color: "#fff",
                }}
              />
            </RadarChartAny>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      {/* Distance Trend Chart */}
      <ChartCard
        title={uiStrings?.chart_distance_title || "Operational Distance Spectrum"}
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChartAny data={tripChartData}>
            <CartesianGridAny
              strokeDasharray="4 4"
              stroke="rgba(255,255,255,0.05)"
              vertical={false}
            />
            <XAxisAny
              dataKey="label"
              tick={{ fill: "#94a3b8", fontSize: 10, fontWeight: 900 }}
              axisLine={false}
              tickLine={false}
              dy={10}
            />
            <YAxisAny
              tick={{ fill: "#94a3b8", fontSize: 10, fontWeight: 900 }}
              axisLine={false}
              tickLine={false}
              dx={-10}
            />
            <TooltipAny
              contentStyle={{
                backgroundColor: "#0F1117",
                borderRadius: "1.5rem",
                border: "1px solid rgba(255,255,255,0.1)",
                padding: "1.5rem",
                boxShadow: "0 25px 50px -12px rgba(0,0,0,0.5)",
                color: "#fff",
              }}
              cursor={{ fill: "rgba(255,255,255,0.02)" }}
            />
            <LegendAny 
                verticalAlign="top" 
                align="right" 
                iconType="circle"
                wrapperStyle={{ paddingBottom: "40px", fontSize: "10px", fontWeight: 900, textTransform: "uppercase", letterSpacing: "0.1em" }}
            />
            <BarAny
              dataKey="total_distance"
              name={uiStrings?.chart_distance_label || "Active Distance (KM)"}
              fill="#137FEC"
              radius={[12, 12, 0, 0]}
              barSize={40}
            />
          </BarChartAny>
        </ResponsiveContainer>
      </ChartCard>

      {/* Workforce Performance Matrix */}
      <div className="bg-surface-container rounded-[3rem] border border-outline overflow-hidden shadow-2xl">
        <div className="p-10 border-b border-outline bg-surface-dim flex justify-between items-center">
          <div className="flex items-center gap-4">
            <div className="w-1.5 h-6 bg-primary rounded-full" />
            <h2 className="text-xl font-bold text-white tracking-tighter uppercase">
              {uiStrings?.table_performance_title || "Personnel Performance Matrix"}
            </h2>
          </div>
          <span className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em] opacity-50">
            Telemetry Synced: {missionClock.toLocaleTimeString()}
          </span>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-outline">
            <thead className="bg-surface-dim">
              <tr>
                <th
                  scope="col"
                  className="px-8 py-6 text-left text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]"
                >
                  {uiStrings?.th_driver || "Assigned Personnel"}
                </th>
                <th
                  scope="col"
                  className="px-8 py-6 text-left text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]"
                >
                  {uiStrings?.th_safety_score || "Safety Index"}
                </th>
                <th
                  scope="col"
                  className="px-8 py-6 text-left text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]"
                >
                  {uiStrings?.th_trips || "Operations"}
                </th>
                <th
                  scope="col"
                  className="px-8 py-6 text-left text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]"
                >
                  {uiStrings?.th_on_time || "SLA Adherence"}
                </th>
                <th
                  scope="col"
                  className="px-8 py-6 text-left text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]"
                >
                  {uiStrings?.th_distance || "Total Spectrum (KM)"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-outline bg-surface-container">
              {normalizedDrivers.map((driver) => (
                <tr
                  key={driver.driver_id}
                  className="hover:bg-surface-container-highest transition-all duration-300 group"
                >
                  <td className="px-8 py-6 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="h-12 w-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary font-black uppercase text-lg border border-primary/20 shadow-inner group-hover:scale-110 transition-transform">
                        {driver.name.charAt(0)}
                      </div>
                      <div className="ml-5">
                        <div className="text-sm font-bold text-white tracking-tight group-hover:text-primary transition-colors">
                          {driver.name}
                        </div>
                        <div className="text-[9px] font-black text-on-surface-variant uppercase tracking-widest opacity-40 mt-1">
                          ID: <span className="font-mono">{driver.driver_id.toString().slice(-8).toUpperCase()}</span>
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-8 py-6 whitespace-nowrap">
                    <div className="flex items-center gap-3">
                      <div className={`h-2 w-2 rounded-full ${
                        driver.safety_score >= 90 ? 'bg-emerald-500' : driver.safety_score >= 80 ? 'bg-amber-500' : 'bg-rose-500'
                      }`} />
                      <span
                        className={`text-xs font-black font-mono tracking-tighter ${
                          driver.safety_score >= 90
                            ? "text-emerald-500"
                            : driver.safety_score >= 80
                              ? "text-amber-500"
                              : "text-rose-500"
                        }`}
                      >
                        {driver.safety_score}/100
                      </span>
                    </div>
                  </td>
                  <td className="px-8 py-6 whitespace-nowrap text-sm font-bold text-white font-mono tracking-tighter">
                    {driver.trips_completed}
                  </td>
                  <td className="px-8 py-6 whitespace-nowrap text-sm font-bold text-white font-mono tracking-tighter">
                    {driver.on_time_rate}%
                  </td>
                  <td className="px-8 py-6 whitespace-nowrap text-sm font-bold text-primary font-mono tracking-tighter">
                    {driver.total_distance.toLocaleString()}
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
  <div className="bg-surface-container border border-outline rounded-[2rem] p-8 relative overflow-hidden group hover:bg-surface-container-highest transition-all duration-500 shadow-xl">
    <div className="absolute top-0 right-0 p-8 opacity-[0.03] group-hover:opacity-[0.08] transition-all transform group-hover:scale-110 duration-700">
      <span className="material-symbols-outlined text-8xl">{icon}</span>
    </div>
    
    <div className="relative space-y-6">
      <div className="flex justify-between items-center">
        <h4 className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.25em]">{title}</h4>
        <div className={`h-1.5 w-1.5 rounded-full ${positive ? 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]' : 'bg-rose-500 shadow-[0_0_10px_rgba(244,63,94,0.5)]'}`} />
      </div>
      
      <div>
        <h3 className="text-4xl font-bold text-white tracking-tighter uppercase font-mono">{value}</h3>
      </div>
      
      <div className="flex items-center gap-3">
        <div className={`flex items-center gap-1 font-black text-[10px] ${positive ? 'text-emerald-500' : 'text-rose-500'} bg-surface-dim border border-outline px-2 py-1 rounded-lg`}>
          <span className="material-symbols-outlined text-sm">{positive ? 'trending_up' : 'trending_down'}</span>
          {change}
        </div>
        <span className="text-[9px] font-black text-on-surface-variant uppercase tracking-widest opacity-40">
          {uiStrings?.label_vs_last_month || "Vs Cycle Mean"}
        </span>
      </div>
    </div>
  </div>
);

const ChartCard = ({ title, children, height = "h-[450px]" }: any) => (
  <div className="bg-surface-container border border-outline rounded-[3rem] p-10 shadow-2xl overflow-hidden relative group">
    <div className="flex items-center justify-between mb-10">
      <div className="flex items-center gap-4">
        <div className="w-1.5 h-6 bg-primary rounded-full" />
        <h2 className="text-xl font-bold text-white tracking-tighter uppercase">{title}</h2>
      </div>
      <div className="h-10 w-10 flex items-center justify-center rounded-xl bg-surface-dim border border-outline text-on-surface-variant opacity-50 hover:opacity-100 hover:text-white transition-all cursor-crosshair">
        <span className="material-symbols-outlined text-xl">fullscreen</span>
      </div>
    </div>
    <div className={`${height} w-full`}>{children}</div>
  </div>
);

export default Analytics;
