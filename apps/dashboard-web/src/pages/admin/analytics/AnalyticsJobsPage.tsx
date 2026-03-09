import {
  PlayIcon,
  PauseIcon,
  StopIcon,
  ArrowPathIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ExclamationTriangleIcon,
  ChartBarIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import React, { useEffect, useState } from "react";
import api from "../../../services/api";

interface AnalyticsJob {
  id: string;
  name: string;
  type: string;
  job_type: string;
  status: string;
  progress: number;
  organization_id: string;
  created_by: string;
  created_at: string;
  start_time?: string;
  end_time?: string;
  error_details?: string;
}

const AnalyticsJobsPage: React.FC = () => {
  const [jobs, setJobs] = useState<AnalyticsJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [filter, setFilter] = useState<string>("all");
  const [selectedJob, setSelectedJob] = useState<AnalyticsJob | null>(null);

  useEffect(() => {
    fetchJobs();
    // Poll for updates every 5 seconds
    const interval = setInterval(fetchJobs, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchJobs = async () => {
    try {
      setLoading(true);
      const response = (await api.get("/admin/analytics/jobs")) as any;
      setJobs(response?.data || response?.jobs || []);
    } catch (error) {
      console.error("Failed to fetch jobs", error);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (id: string) => {
    if (!confirm("Are you sure you want to cancel this job?")) return;
    try {
      await api.put(`/admin/analytics/jobs/${id}`, { status: "cancelled" });
      fetchJobs();
    } catch (error) {
      console.error("Failed to cancel job", error);
      alert("Failed to cancel job");
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "completed":
        return "bg-emerald-500/10 text-emerald-500 border-emerald-500/20";
      case "running":
        return "bg-blue-500/10 text-blue-500 border-blue-500/20 animate-pulse";
      case "pending":
        return "bg-yellow-500/10 text-yellow-500 border-yellow-500/20";
      case "failed":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      case "cancelled":
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
      default:
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "completed":
        return <CheckCircleIcon className="h-5 w-5" />;
      case "running":
        return <ArrowPathIcon className="h-5 w-5 animate-spin" />;
      case "failed":
        return <XCircleIcon className="h-5 w-5" />;
      default:
        return <ClockIcon className="h-5 w-5" />;
    }
  };

  const getJobTypeIcon = (type: string) => {
    switch (type) {
      case "prediction":
        return "🔮";
      case "training":
        return "🎯";
      case "evaluation":
        return "📊";
      case "inference":
        return "⚡";
      default:
        return "📋";
    }
  };

  const filteredJobs = filter === "all" 
    ? jobs 
    : jobs.filter(j => j.status === filter);

  const stats = {
    total: jobs.length,
    running: jobs.filter(j => j.status === "running").length,
    completed: jobs.filter(j => j.status === "completed").length,
    failed: jobs.filter(j => j.status === "failed").length,
  };

  return (
    <div className="p-6 bg-slate-900 min-h-screen">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            📊 Analytics Jobs
          </h1>
          <p className="text-slate-400 text-sm mt-1">
            Monitor and manage analytics job execution
          </p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={fetchJobs}
            className="px-4 py-2 bg-slate-800 text-white rounded-lg border border-slate-700 hover:bg-slate-700 transition flex items-center gap-2"
          >
            <ArrowPathIcon className="h-4 w-4" /> Refresh
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition flex items-center gap-2 shadow-lg shadow-blue-900/20"
          >
            <PlayIcon className="h-4 w-4" /> New Job
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <StatCard label="Total Jobs" value={stats.total} color="text-white" />
        <StatCard label="Running" value={stats.running} color="text-blue-400" />
        <StatCard label="Completed" value={stats.completed} color="text-emerald-400" />
        <StatCard label="Failed" value={stats.failed} color="text-red-400" />
      </div>

      {/* Filters */}
      <div className="flex gap-2 mb-6">
        {["all", "pending", "running", "completed", "failed", "cancelled"].map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition ${
              filter === status
                ? "bg-blue-600 text-white"
                : "bg-slate-800 text-slate-400 hover:bg-slate-700"
            }`}
          >
            {status.charAt(0).toUpperCase() + status.slice(1)}
          </button>
        ))}
      </div>

      {/* Jobs Table */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      ) : filteredJobs.length === 0 ? (
        <div className="text-center py-20">
          <ExclamationTriangleIcon className="h-16 w-16 mx-auto text-slate-600 mb-4" />
          <h3 className="text-lg font-medium text-white">No jobs found</h3>
          <p className="text-slate-400 mt-2">
            Create your first analytics job to get started
          </p>
        </div>
      ) : (
        <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
          <table className="min-w-full divide-y divide-slate-700">
            <thead className="bg-slate-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                  Job
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                  Progress
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                  Started
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-slate-400 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-700">
              {filteredJobs.map((job) => (
                <tr
                  key={job.id}
                  className="hover:bg-slate-700/50 transition cursor-pointer"
                  onClick={() => setSelectedJob(job)}
                >
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <ChartBarIcon className="h-5 w-5 text-slate-500" />
                      <div>
                        <div className="text-sm font-medium text-white">
                          {job.name}
                        </div>
                        <div className="text-xs text-slate-400">
                          {job.id.slice(0, 8)}...
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-white">
                      {getJobTypeIcon(job.job_type)} {job.job_type}
                    </div>
                    <div className="text-xs text-slate-400">{job.type}</div>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-semibold border ${getStatusColor(
                        job.status
                      )}`}
                    >
                      {getStatusIcon(job.status)}
                      {job.status}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <div className="flex-1 h-2 bg-slate-700 rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full transition-all ${
                            job.status === "failed"
                              ? "bg-red-500"
                              : job.status === "completed"
                              ? "bg-emerald-500"
                              : "bg-blue-500"
                          }`}
                          style={{ width: `${job.progress}%` }}
                        />
                      </div>
                      <span className="text-xs text-slate-400 w-12 text-right">
                        {job.progress.toFixed(0)}%
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm text-slate-400">
                    {job.start_time
                      ? format(new Date(job.start_time), "MMM d, HH:mm")
                      : "-"}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      {job.status === "running" && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCancel(job.id);
                          }}
                          className="p-2 text-red-400 hover:bg-red-500/20 rounded-lg transition"
                          title="Cancel job"
                        >
                          <StopIcon className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <CreateJobModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={fetchJobs}
        />
      )}

      {/* Detail Modal */}
      {selectedJob && (
        <JobDetailModal
          job={selectedJob}
          onClose={() => setSelectedJob(null)}
        />
      )}
    </div>
  );
};

// Stat Card Component
const StatCard: React.FC<{ label: string; value: number; color: string }> = ({
  label,
  value,
  color,
}) => (
  <div className="bg-slate-800 rounded-xl border border-slate-700 p-4">
    <div className="text-2xl font-bold text-white">{value}</div>
    <div className={`text-sm ${color}`}>{label}</div>
  </div>
);

// Create Job Modal
interface CreateJobModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

const CreateJobModal: React.FC<CreateJobModalProps> = ({ onClose, onSuccess }) => {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    type: "batch",
    job_type: "prediction",
    description: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await api.post("/admin/analytics/jobs", formData);
      onSuccess();
      onClose();
    } catch (error) {
      console.error("Failed to create job", error);
      alert("Failed to create job");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-xl border border-slate-700 w-full max-w-lg p-6">
        <h2 className="text-xl font-bold text-white mb-6">Create Analytics Job</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1">
              Job Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              placeholder="e.g., Weekly Demand Prediction"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-1">
                Type *
              </label>
              <select
                required
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              >
                <option value="batch">Batch</option>
                <option value="real_time">Real-time</option>
                <option value="streaming">Streaming</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-1">
                Job Type *
              </label>
              <select
                required
                value={formData.job_type}
                onChange={(e) => setFormData({ ...formData, job_type: e.target.value })}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              >
                <option value="prediction">Prediction</option>
                <option value="training">Training</option>
                <option value="evaluation">Evaluation</option>
                <option value="inference">Inference</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              rows={3}
              placeholder="Job description..."
            />
          </div>
          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition disabled:opacity-50"
            >
              {loading ? "Creating..." : "Create Job"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// Job Detail Modal
interface JobDetailModalProps {
  job: AnalyticsJob;
  onClose: () => void;
}

const JobDetailModal: React.FC<JobDetailModalProps> = ({ job, onClose }) => {
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-xl border border-slate-700 w-full max-w-2xl p-6">
        <div className="flex justify-between items-start mb-6">
          <div>
            <h2 className="text-xl font-bold text-white">{job.name}</h2>
            <p className="text-slate-400 text-sm">ID: {job.id}</p>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white">
            <XCircleIcon className="h-6 w-6" />
          </button>
        </div>

        <div className="mb-6">
          <div className="flex justify-between text-sm mb-2">
            <span className="text-slate-400">Progress</span>
            <span className="text-white">{job.progress.toFixed(0)}%</span>
          </div>
          <div className="h-3 bg-slate-700 rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full transition-all ${
                job.status === "failed"
                  ? "bg-red-500"
                  : job.status === "completed"
                  ? "bg-emerald-500"
                  : "bg-blue-500"
              }`}
              style={{ width: `${job.progress}%` }}
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-6 text-sm">
          <div>
            <div className="text-slate-400 mb-1">Status</div>
            <div className="text-white capitalize">{job.status}</div>
          </div>
          <div>
            <div className="text-slate-400 mb-1">Type</div>
            <div className="text-white">{job.type} / {job.job_type}</div>
          </div>
          <div>
            <div className="text-slate-400 mb-1">Created</div>
            <div className="text-white">
              {format(new Date(job.created_at), "MMM d, yyyy HH:mm")}
            </div>
          </div>
          <div>
            <div className="text-slate-400 mb-1">Started</div>
            <div className="text-white">
              {job.start_time
                ? format(new Date(job.start_time), "MMM d, yyyy HH:mm")
                : "-"}
            </div>
          </div>
        </div>

        {job.error_details && (
          <div className="mt-6 p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
            <div className="text-red-400 text-sm font-medium mb-1">Error</div>
            <div className="text-red-300 text-sm">{job.error_details}</div>
          </div>
        )}

        <div className="flex gap-3 mt-6 pt-6 border-t border-slate-700">
          <button
            onClick={onClose}
            className="flex-1 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition"
          >
            Close
          </button>
          {job.status === "completed" && (
            <button className="flex-1 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg transition">
              View Results
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default AnalyticsJobsPage;
