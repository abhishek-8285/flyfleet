import {
    ArrowPathIcon,
    ExclamationTriangleIcon,
    PencilIcon,
    PlusIcon,
    TrashIcon,
    XCircleIcon
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import React, { useEffect, useState } from "react";
import api from "../../../services/api";

interface PredictiveModel {
  id: string;
  name: string;
  version: string;
  model_type: string;
  algorithm: string;
  status: string;
  accuracy: number;
  trained_at: string;
  expires_at: string;
  description?: string;
}

const PredictiveModelsPage: React.FC = () => {
  const [models, setModels] = useState<PredictiveModel[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedModel, setSelectedModel] = useState<PredictiveModel | null>(
    null,
  );
  const [filter, setFilter] = useState<string>("all");

  useEffect(() => {
    fetchModels();
  }, []);

  const fetchModels = async () => {
    try {
      setLoading(true);
      const response = (await api.get("/admin/analytics/models")) as any;
      setModels(response?.data || response?.models || []);
    } catch (error) {
      console.error("Failed to fetch models", error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to delete this model?")) return;
    try {
      await api.delete(`/admin/analytics/models/${id}`);
      fetchModels();
    } catch (error) {
      console.error("Failed to delete model", error);
      alert("Failed to delete model");
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "active":
        return "bg-emerald-500/10 text-emerald-500 border-emerald-500/20";
      case "inactive":
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
      case "training":
        return "bg-blue-500/10 text-blue-500 border-blue-500/20";
      case "error":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      default:
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
    }
  };

  const getModelTypeIcon = (type: string) => {
    switch (type) {
      case "regression":
        return "📈";
      case "classification":
        return "🏷️";
      case "clustering":
        return "🔵";
      default:
        return "🤖";
    }
  };

  const filteredModels =
    filter === "all" ? models : models.filter((m) => m.status === filter);

  return (
    <div className="p-6 bg-slate-900 min-h-screen">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            🤖 Predictive Models
          </h1>
          <p className="text-slate-400 text-sm mt-1">
            Manage ML models for predictive analytics
          </p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={fetchModels}
            className="px-4 py-2 bg-slate-800 text-white rounded-lg border border-slate-700 hover:bg-slate-700 transition flex items-center gap-2"
          >
            <ArrowPathIcon className="h-4 w-4" /> Refresh
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition flex items-center gap-2 shadow-lg shadow-blue-900/20"
          >
            <PlusIcon className="h-4 w-4" /> New Model
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-2 mb-6">
        {["all", "active", "training", "inactive", "error"].map((status) => (
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

      {/* Models Grid */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      ) : filteredModels.length === 0 ? (
        <div className="text-center py-20">
          <ExclamationTriangleIcon className="h-16 w-16 mx-auto text-slate-600 mb-4" />
          <h3 className="text-lg font-medium text-white">No models found</h3>
          <p className="text-slate-400 mt-2">
            Create your first predictive model to get started
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredModels.map((model) => (
            <div
              key={model.id}
              className="bg-slate-800 rounded-xl border border-slate-700 p-6 hover:border-blue-500/30 transition group"
            >
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-3">
                  <span className="text-3xl">
                    {getModelTypeIcon(model.model_type)}
                  </span>
                  <div>
                    <h3 className="text-lg font-bold text-white">
                      {model.name}
                    </h3>
                    <p className="text-sm text-slate-400">v{model.version}</p>
                  </div>
                </div>
                <span
                  className={`px-2 py-1 rounded-full text-xs font-semibold border ${getStatusColor(
                    model.status,
                  )}`}
                >
                  {model.status.toUpperCase()}
                </span>
              </div>

              <div className="space-y-3 mb-4">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Algorithm</span>
                  <span className="text-white">
                    {model.algorithm.replace("_", " ")}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Accuracy</span>
                  <span className="text-emerald-400 font-semibold">
                    {(model.accuracy * 100).toFixed(1)}%
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Trained</span>
                  <span className="text-white">
                    {format(new Date(model.trained_at), "MMM d, yyyy")}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Expires</span>
                  <span className="text-amber-400">
                    {format(new Date(model.expires_at), "MMM d, yyyy")}
                  </span>
                </div>
              </div>

              <div className="flex gap-2 pt-4 border-t border-slate-700">
                <button
                  onClick={() => setSelectedModel(model)}
                  className="flex-1 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg text-sm font-medium transition flex items-center justify-center gap-2"
                >
                  <PencilIcon className="h-4 w-4" /> View
                </button>
                <button
                  onClick={() => handleDelete(model.id)}
                  className="px-3 py-2 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded-lg transition"
                >
                  <TrashIcon className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <CreateModelModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={fetchModels}
        />
      )}

      {/* Detail Modal */}
      {selectedModel && (
        <ModelDetailModal
          model={selectedModel}
          onClose={() => setSelectedModel(null)}
        />
      )}
    </div>
  );
};

// Create Model Modal Component
interface CreateModelModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

const CreateModelModal: React.FC<CreateModelModalProps> = ({
  onClose,
  onSuccess,
}) => {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    version: "",
    model_type: "regression",
    algorithm: "random_forest",
    description: "",
    output_variable: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await api.post("/admin/analytics/models", formData);
      onSuccess();
      onClose();
    } catch (error) {
      console.error("Failed to create model", error);
      alert("Failed to create model");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-xl border border-slate-700 w-full max-w-lg p-6">
        <h2 className="text-xl font-bold text-white mb-6">
          Create Predictive Model
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1">
              Model Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              placeholder="e.g., Demand Forecasting"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-1">
                Version *
              </label>
              <input
                type="text"
                required
                value={formData.version}
                onChange={(e) =>
                  setFormData({ ...formData, version: e.target.value })
                }
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
                placeholder="1.0.0"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-1">
                Model Type *
              </label>
              <select
                required
                value={formData.model_type}
                onChange={(e) =>
                  setFormData({ ...formData, model_type: e.target.value })
                }
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              >
                <option value="regression">Regression</option>
                <option value="classification">Classification</option>
                <option value="clustering">Clustering</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1">
              Algorithm *
            </label>
            <select
              required
              value={formData.algorithm}
              onChange={(e) =>
                setFormData({ ...formData, algorithm: e.target.value })
              }
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
            >
              <option value="random_forest">Random Forest</option>
              <option value="neural_network">Neural Network</option>
              <option value="svm">Support Vector Machine</option>
              <option value="xgboost">XGBoost</option>
              <option value="linear_regression">Linear Regression</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1">
              Output Variable *
            </label>
            <input
              type="text"
              required
              value={formData.output_variable}
              onChange={(e) =>
                setFormData({ ...formData, output_variable: e.target.value })
              }
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              placeholder="e.g., predicted_demand"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) =>
                setFormData({ ...formData, description: e.target.value })
              }
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
              rows={3}
              placeholder="Model description..."
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
              {loading ? "Creating..." : "Create Model"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// Model Detail Modal
interface ModelDetailModalProps {
  model: PredictiveModel;
  onClose: () => void;
}

const ModelDetailModal: React.FC<ModelDetailModalProps> = ({
  model,
  onClose,
}) => {
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-xl border border-slate-700 w-full max-w-2xl p-6 max-h-[80vh] overflow-y-auto">
        <div className="flex justify-between items-start mb-6">
          <div className="flex items-center gap-3">
            <span className="text-4xl">
              {model.model_type === "regression"
                ? "📈"
                : model.model_type === "classification"
                  ? "🏷️"
                  : "🔵"}
            </span>
            <div>
              <h2 className="text-2xl font-bold text-white">{model.name}</h2>
              <p className="text-slate-400">Version {model.version}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-white transition"
          >
            <XCircleIcon className="h-6 w-6" />
          </button>
        </div>

        <div className="grid grid-cols-2 gap-6">
          <div>
            <h3 className="text-sm font-semibold text-slate-300 mb-3">
              Model Info
            </h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-slate-400">Algorithm</span>
                <span className="text-white">
                  {model.algorithm.replace("_", " ")}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Type</span>
                <span className="text-white capitalize">
                  {model.model_type}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Status</span>
                <span className="text-white capitalize">{model.status}</span>
              </div>
            </div>
          </div>
          <div>
            <h3 className="text-sm font-semibold text-slate-300 mb-3">
              Performance
            </h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-slate-400">Accuracy</span>
                <span className="text-emerald-400">
                  {(model.accuracy * 100).toFixed(1)}%
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Trained</span>
                <span className="text-white">
                  {format(new Date(model.trained_at), "MMM d, yyyy")}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Expires</span>
                <span className="text-amber-400">
                  {format(new Date(model.expires_at), "MMM d, yyyy")}
                </span>
              </div>
            </div>
          </div>
        </div>

        {model.description && (
          <div className="mt-6">
            <h3 className="text-sm font-semibold text-slate-300 mb-2">
              Description
            </h3>
            <p className="text-slate-400 text-sm">{model.description}</p>
          </div>
        )}

        <div className="flex gap-3 mt-6 pt-6 border-t border-slate-700">
          <button
            onClick={onClose}
            className="flex-1 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition"
          >
            Close
          </button>
          <button className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition">
            Retrain Model
          </button>
        </div>
      </div>
    </div>
  );
};

export default PredictiveModelsPage;
