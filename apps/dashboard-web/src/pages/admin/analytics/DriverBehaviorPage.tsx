import {
  UserGroupIcon,
  TrophyIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  ArrowPathIcon,
  StarIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import React, { useEffect, useState } from "react";
import {
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import api from "../../../services/api";

interface DriverBehavior {
  id: string;
  driver_id: string;
  driver_name?: string;
  organization_id: string;
  analysis_period_start: string;
  analysis_period_end: string;
  total_driving_hours: number;
  aggressive_acceleration_count: number;
  hard_braking_count: number;
  sharp_turn_count: number;
  over_speeding_count: number;
  idling_time: number;
  average_speed: number;
  max_speed: number;
  fuel_efficiency: number;
  safety_score: number;
  efficiency_score: number;
  comfort_score: number;
  ranking: number;
  improvement_suggestions: string[];
  behavior_trend: string;
  generated_at: string;
}

const DriverBehaviorPage: React.FC = () => {
  const [drivers, setDrivers] = useState<DriverBehavior[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDriver, setSelectedDriver] = useState<DriverBehavior | null>(null);
  const [sortBy, setSortBy] = useState<"safety" | "efficiency" | "ranking">("safety");
  const [filter, setFilter] = useState<string>("all");

  useEffect(() => {
    fetchDrivers();
  }, []);

  const fetchDrivers = async () => {
    try {
      setLoading(true);
      const response = (await api.get("/admin/analytics/driver-behavior/org", {
        params: { orgId: "current" },
      })) as any;
      setDrivers(response?.data || []);
    } catch (error) {
      console.error("Failed to fetch driver behavior", error);
    } finally {
      setLoading(false);
    }
  };

  const sortedDrivers = [...drivers].sort((a, b) => {
    switch (sortBy) {
      case "safety":
        return b.safety_score - a.safety_score;
      case "efficiency":
        return b.efficiency_score - a.efficiency_score;
      case "ranking":
        return a.ranking - b.ranking;
      default:
        return 0;
    }
  });

  const filteredDrivers = filter === "all" 
    ? sortedDrivers 
    : sortedDrivers.filter(d => {
        if (filter === "excellent") return d.safety_score >= 90;
        if (filter === "good") return d.safety_score >= 70 && d.safety_score < 90;
        if (filter === "needs-improvement") return d.safety_score < 70;
        return true;
      });

  const topPerformers = sortedDrivers.slice(0, 3);
  const needsImprovement = sortedDrivers.filter(d => d.safety_score < 70);

  const getScoreColor = (score: number) => {
    if (score >= 90) return "text-emerald-400";
    if (score >= 70) return "text-amber-400";
    return "text-red-400";
  };

  const getScoreBadge = (score: number) => {
    if (score >= 90) return "bg-emerald-500/10 text-emerald-500 border-emerald-500/20";
    if (score >= 70) return "bg-amber-500/10 text-amber-500 border-amber-500/20";
    return "bg-red-500/10 text-red-500 border-red-500/20";
  };

  const radarData = selectedDriver ? [
    { subject: "Safety", score: selectedDriver.safety_score, fullMark: 100 },
    { subject: "Efficiency", score: selectedDriver.efficiency_score, fullMark: 100 },
    { subject: "Comfort", score: selectedDriver.comfort_score, fullMark: 100 },
    { subject: "Fuel", score: Math.min(100, selectedDriver.fuel_efficiency * 10), fullMark: 100 },
    { subject: "Punctuality", score: Math.max(0, 100 - (selectedDriver.aggressive_acceleration_count + selectedDriver.hard_braking_count)), fullMark: 100 },
    { subject: "Compliance", score: Math.max(0, 100 - selectedDriver.over_speeding_count * 5), fullMark: 100 },
  ] : [];

  const eventsData = selectedDriver ? [
    { event: "Hard Braking", count: selectedDriver.hard_braking_count, color: "#ef4444" },
    { event: "Aggressive Accel", count: selectedDriver.aggressive_acceleration_count, color: "#f59e0b" },
    { event: "Sharp Turn", count: selectedDriver.sharp_turn_count, color: "#3b82f6" },
    { event: "Over Speeding", count: selectedDriver.over_speeding_count, color: "#8b5cf6" },
  ] : [];

  return (
    <div className="p-6 bg-slate-900 min-h-screen">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            👥 Driver Behavior Analytics
          </h1>
          <p className="text-slate-400 text-sm mt-1">
            Monitor and improve driver performance
          </p>
        </div>
        <div className="flex gap-3">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
          >
            <option value="safety">Sort by Safety</option>
            <option value="efficiency">Sort by Efficiency</option>
            <option value="ranking">Sort by Ranking</option>
          </select>
          <button
            onClick={fetchDrivers}
            className="px-4 py-2 bg-slate-800 text-white rounded-lg border border-slate-700 hover:bg-slate-700 transition flex items-center gap-2"
          >
            <ArrowPathIcon className="h-4 w-4" /> Refresh
          </button>
        </div>
      </div>

      {/* Top Performers */}
      <div className="mb-8">
        <h2 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
          <TrophyIcon className="h-5 w-5 text-yellow-400" /> Top Performers
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {topPerformers.map((driver, index) => (
            <div
              key={driver.id}
              className={`bg-gradient-to-br rounded-xl p-6 border ${
                index === 0
                  ? "from-yellow-500/20 to-yellow-600/5 border-yellow-500/30"
                  : index === 1
                  ? "from-slate-400/20 to-slate-500/5 border-slate-400/30"
                  : "from-amber-600/20 to-amber-700/5 border-amber-600/30"
              }`}
            >
              <div className="flex items-center gap-3 mb-4">
                <div className="h-12 w-12 rounded-full bg-slate-800 flex items-center justify-center text-xl font-bold text-white border-2 border-current">
                  {index === 0 ? "🥇" : index === 1 ? "🥈" : "🥉"}
                </div>
                <div>
                  <h3 className="text-white font-semibold">{driver.driver_name || `Driver ${driver.driver_id.slice(0, 8)}`}</h3>
                  <p className="text-slate-400 text-sm">Rank #{driver.ranking}</p>
                </div>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-slate-400 text-sm">Safety Score</span>
                <span className={`text-2xl font-bold ${getScoreColor(driver.safety_score)}`}>
                  {driver.safety_score}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-2 mb-6">
        {[
          { value: "all", label: "All Drivers" },
          { value: "excellent", label: "Excellent (90+)" },
          { value: "good", label: "Good (70-89)" },
          { value: "needs-improvement", label: "Needs Improvement (<70)" },
        ].map((f) => (
          <button
            key={f.value}
            onClick={() => setFilter(f.value)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition ${
              filter === f.value
                ? "bg-blue-600 text-white"
                : "bg-slate-800 text-slate-400 hover:bg-slate-700"
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Drivers Table */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      ) : filteredDrivers.length === 0 ? (
        <div className="text-center py-20">
          <UserGroupIcon className="h-16 w-16 mx-auto text-slate-600 mb-4" />
          <h3 className="text-lg font-medium text-white">No drivers found</h3>
          <p className="text-slate-400 mt-2">No drivers match the selected filter</p>
        </div>
      ) : (
        <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
          <table className="min-w-full divide-y divide-slate-700">
            <thead className="bg-slate-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase">Driver</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase">Safety</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase">Efficiency</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase">Comfort</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase">Events</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase">Trend</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-slate-400 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-700">
              {filteredDrivers.map((driver) => (
                <tr
                  key={driver.id}
                  className="hover:bg-slate-700/50 transition cursor-pointer"
                  onClick={() => setSelectedDriver(driver)}
                >
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 rounded-full bg-blue-500/20 flex items-center justify-center text-blue-400 font-bold">
                        {(driver.driver_name || "D").charAt(0)}
                      </div>
                      <div>
                        <div className="text-sm font-medium text-white">
                          {driver.driver_name || `Driver ${driver.driver_id.slice(0, 8)}`}
                        </div>
                        <div className="text-xs text-slate-400">
                          {driver.total_driving_hours.toFixed(1)} hrs driving
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <div className="flex-1 h-2 bg-slate-700 rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${
                            driver.safety_score >= 90 ? "bg-emerald-500" : driver.safety_score >= 70 ? "bg-amber-500" : "bg-red-500"
                          }`}
                          style={{ width: `${driver.safety_score}%` }}
                        />
                      </div>
                      <span className={`text-sm font-semibold ${getScoreColor(driver.safety_score)}`}>
                        {driver.safety_score}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm text-white">{driver.efficiency_score}</span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm text-white">{driver.comfort_score}</span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1">
                      {(driver.hard_braking_count + driver.aggressive_acceleration_count + driver.over_speeding_count) > 0 ? (
                        <ExclamationTriangleIcon className="h-4 w-4 text-amber-400" />
                      ) : (
                        <CheckCircleIcon className="h-4 w-4 text-emerald-400" />
                      )}
                      <span className="text-sm text-slate-400">
                        {driver.hard_braking_count + driver.aggressive_acceleration_count + driver.over_speeding_count}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`text-xs px-2 py-1 rounded-full border ${
                      driver.behavior_trend === "improving"
                        ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                        : driver.behavior_trend === "declining"
                        ? "bg-red-500/10 text-red-400 border-red-500/20"
                        : "bg-slate-500/10 text-slate-400 border-slate-500/20"
                    }`}>
                      {driver.behavior_trend || "stable"}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedDriver(driver);
                      }}
                      className="text-blue-400 hover:text-blue-300 text-sm font-medium"
                    >
                      View Details
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Needs Improvement Alert */}
      {needsImprovement.length > 0 && (
        <div className="mt-6 bg-red-500/10 border border-red-500/20 rounded-xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <ExclamationTriangleIcon className="h-6 w-6 text-red-400" />
            <h3 className="text-lg font-bold text-white">Drivers Needing Attention</h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {needsImprovement.slice(0, 6).map((driver) => (
              <div key={driver.id} className="bg-red-500/20 rounded-lg p-4">
                <div className="flex justify-between items-center mb-2">
                  <span className="text-white font-medium">
                    {driver.driver_name || `Driver ${driver.driver_id.slice(0, 8)}`}
                  </span>
                  <span className="text-red-400 font-bold">{driver.safety_score}</span>
                </div>
                <div className="text-xs text-slate-400">
                  {driver.improvement_suggestions?.[0] || "Review driving patterns"}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Driver Detail Modal */}
      {selectedDriver && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-slate-800 rounded-xl border border-slate-700 w-full max-w-4xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-slate-700 flex justify-between items-center sticky top-0 bg-slate-800">
              <div>
                <h2 className="text-xl font-bold text-white">
                  {selectedDriver.driver_name || `Driver ${selectedDriver.driver_id.slice(0, 8)}`}
                </h2>
                <p className="text-slate-400 text-sm">Performance Analysis</p>
              </div>
              <button
                onClick={() => setSelectedDriver(null)}
                className="text-slate-400 hover:text-white"
              >
                ✕
              </button>
            </div>

            <div className="p-6">
              {/* Score Cards */}
              <div className="grid grid-cols-3 gap-4 mb-6">
                <ScoreCard label="Safety Score" score={selectedDriver.safety_score} />
                <ScoreCard label="Efficiency Score" score={selectedDriver.efficiency_score} />
                <ScoreCard label="Comfort Score" score={selectedDriver.comfort_score} />
              </div>

              {/* Charts */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
                {/* Radar Chart */}
                <div className="bg-slate-900 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-4">Performance Radar</h3>
                  <ResponsiveContainer width="100%" height={300}>
                    <RadarChart data={radarData}>
                      <PolarGrid stroke="#334155" />
                      <PolarAngleAxis dataKey="subject" tick={{ fill: "#94a3b8", fontSize: 12 }} />
                      <PolarRadiusAxis angle={30} domain={[0, 100]} tick={false} axisLine={false} />
                      <Radar
                        name="Driver Score"
                        dataKey="score"
                        stroke="#3b82f6"
                        fill="#3b82f6"
                        fillOpacity={0.5}
                      />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: "#1e293b",
                          border: "1px solid #334155",
                          borderRadius: "8px",
                          color: "#fff",
                        }}
                      />
                    </RadarChart>
                  </ResponsiveContainer>
                </div>

                {/* Events Bar Chart */}
                <div className="bg-slate-900 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-4">Safety Events</h3>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={eventsData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                      <XAxis dataKey="event" tick={{ fill: "#94a3b8" }} axisLine={false} tickLine={false} />
                      <YAxis tick={{ fill: "#94a3b8" }} axisLine={false} tickLine={false} />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: "#1e293b",
                          border: "1px solid #334155",
                          borderRadius: "8px",
                          color: "#fff",
                        }}
                      />
                      <Legend />
                      <Bar dataKey="count" name="Events" fill="#ef4444" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>

              {/* Stats Grid */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                <StatBox label="Driving Hours" value={selectedDriver.total_driving_hours.toFixed(1)} />
                <StatBox label="Avg Speed" value={`${selectedDriver.average_speed.toFixed(1)} km/h`} />
                <StatBox label="Max Speed" value={`${selectedDriver.max_speed.toFixed(1)} km/h`} />
                <StatBox label="Fuel Efficiency" value={`${selectedDriver.fuel_efficiency.toFixed(1)} km/L`} />
                <StatBox label="Idling Time" value={`${selectedDriver.idling_time.toFixed(1)} hrs`} />
                <StatBox label="Hard Braking" value={selectedDriver.hard_braking_count.toString()} danger={selectedDriver.hard_braking_count > 5} />
                <StatBox label="Over Speeding" value={selectedDriver.over_speeding_count.toString()} danger={selectedDriver.over_speeding_count > 3} />
                <StatBox label="Ranking" value={`#${selectedDriver.ranking}`} />
              </div>

              {/* Improvement Suggestions */}
              {selectedDriver.improvement_suggestions && selectedDriver.improvement_suggestions.length > 0 && (
                <div className="bg-blue-500/10 border border-blue-500/20 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-3 flex items-center gap-2">
                    <StarIcon className="h-5 w-5 text-blue-400" /> Improvement Suggestions
                  </h3>
                  <ul className="space-y-2">
                    {selectedDriver.improvement_suggestions.map((suggestion, index) => (
                      <li key={index} className="text-slate-300 text-sm flex items-start gap-2">
                        <span className="text-blue-400 mt-1">•</span>
                        {suggestion}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const ScoreCard: React.FC<{ label: string; score: number }> = ({ label, score }) => {
  const color = score >= 90 ? "text-emerald-400" : score >= 70 ? "text-amber-400" : "text-red-400";
  const bgColor = score >= 90 ? "bg-emerald-500" : score >= 70 ? "bg-amber-500" : "bg-red-500";
  
  return (
    <div className="bg-slate-900 rounded-xl p-4 text-center">
      <div className="text-slate-400 text-sm mb-2">{label}</div>
      <div className="relative inline-flex items-center justify-center">
        <svg className="w-24 h-24 transform -rotate-90">
          <circle cx="48" cy="48" r="40" stroke="#1e293b" strokeWidth="8" fill="none" />
          <circle
            cx="48"
            cy="48"
            r="40"
            stroke="currentColor"
            strokeWidth="8"
            fill="none"
            strokeDasharray={`${(score / 100) * 251.2} 251.2`}
            className={color}
          />
        </svg>
        <span className={`absolute text-2xl font-bold ${color}`}>{score}</span>
      </div>
    </div>
  );
};

const StatBox: React.FC<{ label: string; value: string; danger?: boolean }> = ({ label, value, danger }) => (
  <div className={`bg-slate-900 rounded-xl p-4 ${danger ? "border border-red-500/30" : ""}`}>
    <div className="text-slate-400 text-xs mb-1">{label}</div>
    <div className={`text-xl font-bold ${danger ? "text-red-400" : "text-white"}`}>{value}</div>
  </div>
);

export default DriverBehaviorPage;
