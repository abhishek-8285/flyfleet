import {
  WrenchIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  ClockIcon,
  ArrowPathIcon,
  TruckIcon,
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

interface PredictiveMaintenance {
  id: string;
  vehicle_id: string;
  vehicle_name?: string;
  organization_id: string;
  analysis_date: string;
  next_service_due: string;
  service_priority: string;
  component_risks: Record<string, number>;
  failure_prediction?: {
    component: string;
    predicted_failure_date: string;
    confidence_score: number;
    urgency_level: string;
    expected_downtime: number;
    expected_cost: number;
    description: string;
  };
  maintenance_schedule: Array<{
    id: string;
    task: string;
    due_date: string;
    priority: string;
    estimated_hours: number;
    cost_estimate: number;
    status: string;
  }>;
  parts_needed: Array<{
    part_number: string;
    description: string;
    quantity: number;
    estimated_cost: number;
    availability: string;
  }>;
  cost_estimation: number;
  uptime_prediction: number;
  reliability_score: number;
  recommendations: string[];
  generated_at: string;
}

const PredictiveMaintenancePage: React.FC = () => {
  const [predictions, setPredictions] = useState<PredictiveMaintenance[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedPrediction, setSelectedPrediction] = useState<PredictiveMaintenance | null>(null);
  const [filter, setFilter] = useState<string>("all");

  useEffect(() => {
    fetchPredictions();
  }, []);

  const fetchPredictions = async () => {
    try {
      setLoading(true);
      const response = (await api.get("/admin/analytics/predictive-maintenance", {
        params: { orgId: "current" },
      })) as any;
      setPredictions(response?.data || []);
    } catch (error) {
      console.error("Failed to fetch predictions", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredPredictions = filter === "all"
    ? predictions
    : predictions.filter(p => p.service_priority === filter);

  const highPriority = predictions.filter(p => p.service_priority === "high");
  const upcomingServices = predictions.filter(p => 
    new Date(p.next_service_due) <= new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
  );

  const priorityData = [
    { name: "High", value: highPriority.length, color: "#ef4444" },
    { name: "Medium", value: predictions.filter(p => p.service_priority === "medium").length, color: "#f59e0b" },
    { name: "Low", value: predictions.filter(p => p.service_priority === "low").length, color: "#10b981" },
  ];

  const reliabilityData = predictions.map(p => ({
    vehicle: p.vehicle_name || p.vehicle_id.slice(0, 8),
    reliability: p.reliability_score,
    uptime: p.uptime_prediction,
  }));

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case "high": return "bg-red-500/10 text-red-400 border-red-500/20";
      case "medium": return "bg-amber-500/10 text-amber-400 border-amber-500/20";
      case "low": return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
      default: return "bg-slate-500/10 text-slate-400 border-slate-500/20";
    }
  };

  const getUrgencyColor = (urgency: string) => {
    switch (urgency) {
      case "critical": return "text-red-400 bg-red-500/10";
      case "high": return "text-orange-400 bg-orange-500/10";
      case "medium": return "text-amber-400 bg-amber-500/10";
      case "low": return "text-slate-400 bg-slate-500/10";
      default: return "text-slate-400 bg-slate-500/10";
    }
  };

  return (
    <div className="p-6 bg-slate-900 min-h-screen">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            🔧 Predictive Maintenance
          </h1>
          <p className="text-slate-400 text-sm mt-1">
            AI-powered maintenance predictions and scheduling
          </p>
        </div>
        <div className="flex gap-3">
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
          >
            <option value="all">All Priorities</option>
            <option value="high">High Priority</option>
            <option value="medium">Medium Priority</option>
            <option value="low">Low Priority</option>
          </select>
          <button
            onClick={fetchPredictions}
            className="px-4 py-2 bg-slate-800 text-white rounded-lg border border-slate-700 hover:bg-slate-700 transition flex items-center gap-2"
          >
            <ArrowPathIcon className="h-4 w-4" /> Refresh
          </button>
        </div>
      </div>

      {/* Alert Banner */}
      {highPriority.length > 0 && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4 mb-6">
          <div className="flex items-center gap-3">
            <ExclamationTriangleIcon className="h-6 w-6 text-red-400" />
            <div>
              <h3 className="text-red-400 font-semibold">
                {highPriority.length} Vehicle{highPriority.length > 1 ? "s" : ""} Require Immediate Attention
              </h3>
              <p className="text-red-300/70 text-sm">
                High priority maintenance tasks pending
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <StatCard
          title="Total Vehicles"
          value={predictions.length.toString()}
          icon={<TruckIcon className="h-6 w-6" />}
          color="text-blue-400"
        />
        <StatCard
          title="High Priority"
          value={highPriority.length.toString()}
          icon={<ExclamationTriangleIcon className="h-6 w-6" />}
          color="text-red-400"
        />
        <StatCard
          title="Due This Week"
          value={upcomingServices.length.toString()}
          icon={<ClockIcon className="h-6 w-6" />}
          color="text-amber-400"
        />
        <StatCard
          title="Avg Reliability"
          value={predictions.length > 0 ? `${(predictions.reduce((sum, p) => sum + p.reliability_score, 0) / predictions.length).toFixed(0)}%` : "--"}
          icon={<CheckCircleIcon className="h-6 w-6" />}
          color="text-emerald-400"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Priority Distribution */}
        <div className="bg-slate-800 rounded-xl border border-slate-700 p-6">
          <h3 className="text-lg font-bold text-white mb-4">Priority Distribution</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={priorityData}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={100}
                paddingAngle={5}
                dataKey="value"
                label={({ name, value }) => `${name}: ${value}`}
              >
                {priorityData.map((entry, index) => (
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
        </div>

        {/* Reliability Scores */}
        <div className="bg-slate-800 rounded-xl border border-slate-700 p-6">
          <h3 className="text-lg font-bold text-white mb-4">Vehicle Reliability Scores</h3>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={reliabilityData.slice(0, 10)}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
              <XAxis dataKey="vehicle" tick={{ fill: "#94a3b8" }} axisLine={false} tickLine={false} />
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
              <Bar dataKey="reliability" name="Reliability %" fill="#3b82f6" radius={[4, 4, 0, 0]} />
              <Bar dataKey="uptime" name="Uptime %" fill="#10b981" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Predictions List */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      ) : filteredPredictions.length === 0 ? (
        <div className="text-center py-20">
          <WrenchIcon className="h-16 w-16 mx-auto text-slate-600 mb-4" />
          <h3 className="text-lg font-medium text-white">No predictions found</h3>
          <p className="text-slate-400 mt-2">No maintenance predictions match the selected filter</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredPredictions.map((prediction) => (
            <div
              key={prediction.id}
              onClick={() => setSelectedPrediction(prediction)}
              className="bg-slate-800 rounded-xl border border-slate-700 p-6 hover:border-blue-500/30 transition cursor-pointer group"
            >
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-3">
                  <div className="h-12 w-12 rounded-lg bg-blue-500/20 flex items-center justify-center text-blue-400">
                    <TruckIcon className="h-6 w-6" />
                  </div>
                  <div>
                    <h3 className="text-white font-semibold">
                      {prediction.vehicle_name || `Vehicle ${prediction.vehicle_id.slice(0, 8)}`}
                    </h3>
                    <p className="text-slate-400 text-sm">
                      Due: {format(new Date(prediction.next_service_due), "MMM d, yyyy")}
                    </p>
                  </div>
                </div>
                <span className={`px-2 py-1 rounded-full text-xs font-semibold border ${getPriorityColor(prediction.service_priority)}`}>
                  {prediction.service_priority.toUpperCase()}
                </span>
              </div>

              <div className="space-y-3 mb-4">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Reliability</span>
                  <span className={`font-semibold ${
                    prediction.reliability_score >= 80 ? "text-emerald-400" :
                    prediction.reliability_score >= 60 ? "text-amber-400" : "text-red-400"
                  }`}>
                    {prediction.reliability_score}%
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Uptime Prediction</span>
                  <span className="text-white">{prediction.uptime_prediction}%</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Est. Cost</span>
                  <span className="text-white">${prediction.cost_estimation.toLocaleString()}</span>
                </div>
              </div>

              {prediction.failure_prediction && (
                <div className={`p-3 rounded-lg mb-4 ${getUrgencyColor(prediction.failure_prediction.urgency_level)}`}>
                  <div className="text-xs font-semibold mb-1">
                    ⚠️ {prediction.failure_prediction.component}
                  </div>
                  <div className="text-xs opacity-80">
                    {format(new Date(prediction.failure_prediction.predicted_failure_date), "MMM d, yyyy")}
                  </div>
                </div>
              )}

              <div className="flex gap-2 pt-4 border-t border-slate-700">
                <button className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition">
                  View Details
                </button>
                <button className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition">
                  <WrenchIcon className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Detail Modal */}
      {selectedPrediction && (
        <PredictionDetailModal
          prediction={selectedPrediction}
          onClose={() => setSelectedPrediction(null)}
        />
      )}
    </div>
  );
};

const StatCard: React.FC<{ title: string; value: string; icon: React.ReactNode; color: string }> = ({
  title,
  value,
  icon,
  color,
}) => (
  <div className="bg-slate-800 rounded-xl border border-slate-700 p-6">
    <div className="flex justify-between items-start mb-4">
      <div className={`p-3 rounded-lg bg-slate-900 ${color}`}>{icon}</div>
    </div>
    <h3 className="text-slate-400 text-sm mb-1">{title}</h3>
    <p className="text-2xl font-bold text-white">{value}</p>
  </div>
);

// Detail Modal Component
interface PredictionDetailModalProps {
  prediction: PredictiveMaintenance;
  onClose: () => void;
}

const PredictionDetailModal: React.FC<PredictionDetailModalProps> = ({ prediction, onClose }) => {
  const componentRiskData = prediction.component_risks ? Object.entries(prediction.component_risks).map(([name, risk]) => ({
    name,
    risk,
  })) : [];

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4 overflow-y-auto">
      <div className="bg-slate-800 rounded-xl border border-slate-700 w-full max-w-4xl my-8">
        <div className="p-6 border-b border-slate-700 flex justify-between items-center">
          <div>
            <h2 className="text-xl font-bold text-white">
              {prediction.vehicle_name || `Vehicle ${prediction.vehicle_id.slice(0, 8)}`}
            </h2>
            <p className="text-slate-400 text-sm">Maintenance Prediction Details</p>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white">✕</button>
        </div>

        <div className="p-6 max-h-[70vh] overflow-y-auto">
          {/* Failure Prediction Alert */}
          {prediction.failure_prediction && (
            <div className={`p-4 rounded-xl mb-6 border ${
              prediction.failure_prediction.urgency_level === "critical"
                ? "bg-red-500/10 border-red-500/20"
                : prediction.failure_prediction.urgency_level === "high"
                ? "bg-orange-500/10 border-orange-500/20"
                : "bg-amber-500/10 border-amber-500/20"
            }`}>
              <div className="flex items-start gap-3">
                <ExclamationTriangleIcon className="h-6 w-6 text-red-400 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <h3 className="text-white font-semibold mb-2">
                    Predicted Failure: {prediction.failure_prediction.component}
                  </h3>
                  <p className="text-slate-300 text-sm mb-3">{prediction.failure_prediction.description}</p>
                  <div className="grid grid-cols-3 gap-4 text-sm">
                    <div>
                      <span className="text-slate-400">Predicted Date</span>
                      <div className="text-white font-medium">
                        {format(new Date(prediction.failure_prediction.predicted_failure_date), "MMM d, yyyy")}
                      </div>
                    </div>
                    <div>
                      <span className="text-slate-400">Confidence</span>
                      <div className="text-white font-medium">
                        {(prediction.failure_prediction.confidence_score * 100).toFixed(0)}%
                      </div>
                    </div>
                    <div>
                      <span className="text-slate-400">Expected Downtime</span>
                      <div className="text-white font-medium">
                        {prediction.failure_prediction.expected_downtime} hrs
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Scores */}
          <div className="grid grid-cols-3 gap-4 mb-6">
            <ScoreBox label="Reliability Score" value={prediction.reliability_score} />
            <ScoreBox label="Uptime Prediction" value={prediction.uptime_prediction} suffix="%" />
            <ScoreBox label="Est. Cost" value={prediction.cost_estimation} prefix="$" />
          </div>

          {/* Component Risks */}
          {componentRiskData.length > 0 && (
            <div className="mb-6">
              <h3 className="text-white font-semibold mb-3">Component Risk Scores</h3>
              <div className="space-y-2">
                {componentRiskData.map((component, index) => (
                  <div key={index} className="flex items-center gap-3">
                    <span className="text-slate-400 text-sm w-32">{component.name}</span>
                    <div className="flex-1 h-2 bg-slate-700 rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full ${
                          component.risk >= 0.8 ? "bg-red-500" : component.risk >= 0.5 ? "bg-amber-500" : "bg-emerald-500"
                        }`}
                        style={{ width: `${component.risk * 100}%` }}
                      />
                    </div>
                    <span className="text-white text-sm w-12 text-right">{(component.risk * 100).toFixed(0)}%</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Maintenance Schedule */}
          {prediction.maintenance_schedule && prediction.maintenance_schedule.length > 0 && (
            <div className="mb-6">
              <h3 className="text-white font-semibold mb-3">Maintenance Schedule</h3>
              <div className="space-y-2">
                {prediction.maintenance_schedule.map((task) => (
                  <div key={task.id} className="bg-slate-900 rounded-lg p-3 flex justify-between items-center">
                    <div>
                      <div className="text-white font-medium">{task.task}</div>
                      <div className="text-slate-400 text-sm">
                        Due: {format(new Date(task.due_date), "MMM d, yyyy")} • {task.estimated_hours} hrs
                      </div>
                    </div>
                    <span className={`px-2 py-1 rounded text-xs font-semibold ${
                      task.priority === "high" ? "bg-red-500/20 text-red-400" :
                      task.priority === "medium" ? "bg-amber-500/20 text-amber-400" :
                      "bg-emerald-500/20 text-emerald-400"
                    }`}>
                      {task.priority}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Parts Needed */}
          {prediction.parts_needed && prediction.parts_needed.length > 0 && (
            <div className="mb-6">
              <h3 className="text-white font-semibold mb-3">Parts Needed</h3>
              <div className="bg-slate-900 rounded-lg overflow-hidden">
                <table className="min-w-full divide-y divide-slate-700">
                  <thead className="bg-slate-950">
                    <tr>
                      <th className="px-4 py-2 text-left text-xs font-medium text-slate-400">Part</th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-slate-400">Qty</th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-slate-400">Cost</th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-slate-400">Availability</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-700">
                    {prediction.parts_needed.map((part, index) => (
                      <tr key={index}>
                        <td className="px-4 py-3 text-sm text-white">{part.description}</td>
                        <td className="px-4 py-3 text-sm text-slate-400">{part.quantity}</td>
                        <td className="px-4 py-3 text-sm text-white">${part.estimated_cost.toLocaleString()}</td>
                        <td className="px-4 py-3">
                          <span className={`text-xs px-2 py-1 rounded ${
                            part.availability === "in_stock" ? "bg-emerald-500/20 text-emerald-400" :
                            part.availability === "low_stock" ? "bg-amber-500/20 text-amber-400" :
                            "bg-red-500/20 text-red-400"
                          }`}>
                            {part.availability.replace("_", " ")}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Recommendations */}
          {prediction.recommendations && prediction.recommendations.length > 0 && (
            <div className="bg-blue-500/10 border border-blue-500/20 rounded-xl p-4">
              <h3 className="text-white font-semibold mb-3">Recommendations</h3>
              <ul className="space-y-2">
                {prediction.recommendations.map((rec, index) => (
                  <li key={index} className="text-slate-300 text-sm flex items-start gap-2">
                    <span className="text-blue-400 mt-1">•</span>
                    {rec}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        <div className="p-6 border-t border-slate-700 flex gap-3">
          <button onClick={onClose} className="flex-1 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition">
            Close
          </button>
          <button className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition">
            Schedule Maintenance
          </button>
        </div>
      </div>
    </div>
  );
};

const ScoreBox: React.FC<{ label: string; value: number; prefix?: string; suffix?: string }> = ({
  label,
  value,
  prefix = "",
  suffix = "",
}) => (
  <div className="bg-slate-900 rounded-xl p-4 text-center">
    <div className="text-slate-400 text-sm mb-2">{label}</div>
    <div className="text-2xl font-bold text-white">
      {prefix}{typeof value === "number" ? value.toLocaleString() : value}{suffix}
    </div>
  </div>
);

export default PredictiveMaintenancePage;
