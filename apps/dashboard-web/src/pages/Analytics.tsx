import {
  ArrowTrendingDownIcon,
  ArrowTrendingUpIcon,
  ClockIcon,
  TruckIcon
} from '@heroicons/react/24/outline';
import { useEffect, useState } from 'react';
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
} from 'recharts';
import api from '../services/api';

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

// Mock Predictive Data
const predictiveData = [
  { day: 'Mon', actual: 45, predicted: 48 },
  { day: 'Tue', actual: 52, predicted: 50 },
  { day: 'Wed', actual: 49, predicted: 55 },
  { day: 'Thu', actual: 60, predicted: 58 },
  { day: 'Fri', actual: 55, predicted: 62 },
  { day: 'Sat', actual: 30, predicted: 35 },
  { day: 'Sun', actual: 25, predicted: 28 },
];

const driverSkillsData = [
  { subject: 'Safety', A: 120, fullMark: 150 },
  { subject: 'Efficiency', A: 98, fullMark: 150 },
  { subject: 'Punctuality', A: 86, fullMark: 150 },
  { subject: 'Compliance', A: 99, fullMark: 150 },
  { subject: 'Maintenance', A: 85, fullMark: 150 },
  { subject: 'Customer', A: 65, fullMark: 150 },
];

const Analytics = () => {
  const [tripStats, setTripStats] = useState<any[]>([]);
  const [driverStats, setDriverStats] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [trips, drivers] = await Promise.all([
          api.getTripStats(),
          api.getDriverStats(),
        ]);
        setTripStats(trips || []);
        setDriverStats(drivers || []);
      } catch (error) {
        console.error('Failed to fetch analytics data', error);
        setTripStats([]);
        setDriverStats([]);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

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
          <h1 className="text-2xl font-bold tracking-tight">Analytics & Intelligence</h1>
          <p className="text-text-muted text-sm">AI-driven predictive modeling and fleet performance</p>
        </div>
        <div className="flex gap-2">
          <button className="px-4 py-2 bg-surface-dark border border-border-color rounded-lg text-sm font-medium text-white hover:bg-surface-darker transition-colors">
            Last 7 Days
          </button>
          <button className="px-4 py-2 bg-primary hover:bg-primary/90 rounded-lg text-sm font-medium text-white shadow-lg shadow-primary/20 transition-all">
            Export Report
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <KpiCard title="Fleet Efficiency" value="94.2%" change="+2.4%" icon={<ArrowTrendingUpIcon className="text-emerald-400" />} positive />
        <KpiCard title="Fuel Consumption" value="12.8 L/100km" change="-1.2%" icon={<TruckIcon className="text-blue-400" />} positive />
        <KpiCard title="Maintenance Costs" value="$4,250" change="+5.1%" icon={<ArrowTrendingDownIcon className="text-red-400" />} positive={false} />
        <KpiCard title="On-Time Delivery" value="98.5%" change="+0.5%" icon={<ClockIcon className="text-yellow-400" />} positive />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Predictive Maintenance Chart */}
        <ChartCard title="Predictive Fleet Utilization (AI Model)">
          <ResponsiveContainer width="100%" height="100%">
            <LineChartAny data={predictiveData}>
              <CartesianGridAny strokeDasharray="3 3" stroke="#334155" vertical={false} />
              <XAxisAny dataKey="day" tick={{ fill: '#94a3b8' }} axisLine={false} tickLine={false} />
              <YAxisAny tick={{ fill: '#94a3b8' }} axisLine={false} tickLine={false} />
              <TooltipAny
                contentStyle={{ backgroundColor: '#1e293b', borderRadius: '8px', border: '1px solid #334155', color: '#fff' }}
              />
              <LegendAny />
              <LineAny type="monotone" dataKey="actual" name="Actual Usage" stroke="#3b82f6" strokeWidth={3} dot={{ r: 4 }} />
              <LineAny type="monotone" dataKey="predicted" name="AI Prediction" stroke="#10b981" strokeWidth={3} strokeDasharray="5 5" dot={false} />
            </LineChartAny>
          </ResponsiveContainer>
        </ChartCard>

        {/* Driver Skills Radar */}
        <ChartCard title="Driver Workforce Skills Matrix">
          <ResponsiveContainer width="100%" height="100%">
            <RadarChartAny cx="50%" cy="50%" outerRadius="80%" data={driverSkillsData}>
              <PolarGridAny stroke="#334155" />
              <PolarAngleAxisAny dataKey="subject" tick={{ fill: '#94a3b8', fontSize: 12 }} />
              <PolarRadiusAxisAny angle={30} domain={[0, 150]} tick={false} axisLine={false} />
              <RadarAny name="Fleet Average" dataKey="A" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.3} />
              <LegendAny />
              <TooltipAny contentStyle={{ backgroundColor: '#1e293b', borderRadius: '8px', border: '1px solid #334155', color: '#fff' }} />
            </RadarChartAny>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      {/* Distance Trend Chart */}
      <ChartCard title="Distance Traveled Analysis" height="h-80">
        <ResponsiveContainer width="100%" height="100%">
          <BarChartAny data={tripStats}>
            <CartesianGridAny strokeDasharray="3 3" stroke="#334155" vertical={false} />
            <XAxisAny dataKey="date" tick={{ fill: '#94a3b8' }} axisLine={false} tickLine={false} />
            <YAxisAny tick={{ fill: '#94a3b8' }} axisLine={false} tickLine={false} />
            <TooltipAny
              contentStyle={{ backgroundColor: '#1e293b', borderRadius: '8px', border: '1px solid #334155', color: '#fff' }}
              cursor={{ fill: 'rgba(255,255,255,0.05)' }}
            />
            <LegendAny />
            <BarAny dataKey="total_distance" name="Distance (km)" fill="#3b82f6" radius={[4, 4, 0, 0]} />
          </BarChartAny>
        </ResponsiveContainer>
      </ChartCard>

      {/* Driver Performance Table - Dark Mode */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-6 border-b border-border-color bg-surface-darker/50">
          <h2 className="text-lg font-bold text-white">Top Driver Performance</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border-color">
            <thead className="bg-surface-darker">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Driver</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Safety Score</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Trips</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">On-Time</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Distance</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {driverStats.map((driver) => (
                <tr key={driver.driver_id} className="hover:bg-surface-darker/50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold mr-3 border border-primary/30">
                        {driver.name.charAt(0)}
                      </div>
                      <div className="text-sm font-medium text-white">{driver.name}</div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-0.5 inline-flex text-xs leading-5 font-semibold rounded-full border ${driver.safety_score >= 90 ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' :
                      driver.safety_score >= 80 ? 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20' : 'bg-red-500/10 text-red-500 border-red-500/20'
                      }`}>
                      {driver.safety_score}/100
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted">{driver.trips_completed}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted">{driver.on_time_rate}%</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted">{driver.total_distance.toLocaleString()} km</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

const KpiCard = ({ title, value, change, icon, positive }: any) => (
  <div className="bg-surface-dark p-6 rounded-xl border border-border-color relative overflow-hidden group hover:border-primary/30 transition-all">
    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
      {/* Clone icon for background effect */}
      <div className="origin-top-right transform scale-150">{icon}</div>
    </div>
    <div className="flex justify-between items-start mb-4">
      <div>
        <p className="text-xs font-semibold text-text-muted uppercase tracking-wider">{title}</p>
        <h3 className="text-2xl font-bold text-white mt-1">{value}</h3>
      </div>
      <div className="p-2 rounded-lg bg-surface-darker border border-border-color">
        {icon}
      </div>
    </div>
    <div className={`text-xs font-medium flex items-center gap-1 ${positive ? 'text-emerald-500' : 'text-red-500'}`}>
      <span>{change}</span>
      <span className="text-text-muted">vs last month</span>
    </div>
  </div>
);

const ChartCard = ({ title, children, height = "h-80" }: any) => (
  <div className="bg-surface-dark p-6 rounded-xl border border-border-color">
    <h2 className="text-lg font-bold text-white mb-6">{title}</h2>
    <div className={`${height} w-full`}>
      {children}
    </div>
  </div>
);

export default Analytics;
