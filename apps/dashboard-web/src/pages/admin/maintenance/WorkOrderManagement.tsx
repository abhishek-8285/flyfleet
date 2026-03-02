import {
    ClockIcon,
    FunnelIcon,
    PlusIcon,
    WrenchScrewdriverIcon,
} from "@heroicons/react/24/outline";
import React, { useState } from "react";
import apiClient from "../../../services/apiClient";

interface WorkOrder {
  id: string;
  vehicle_id: string;
  vehicle_name?: string;
  title: string;
  priority: "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
  status: "OPEN" | "IN_PROGRESS" | "ON_HOLD" | "COMPLETED";
  assigned_to: string;
  due_date: string;
}

const WorkOrderManagement: React.FC = () => {
  const [workOrders, setWorkOrders] = useState<WorkOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadWorkOrders = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = (await apiClient.get("/admin/work-orders", {
        params: { page: 1, page_size: 100 },
      })) as any;
      const rows = Array.isArray(response?.work_orders)
        ? response.work_orders
        : [];
      setWorkOrders(
        rows.map((wo: any) => ({
          id: wo.id,
          vehicle_id: wo.vehicle_id,
          vehicle_name: wo.vehicle?.license_plate || wo.vehicle_id,
          title: wo.title,
          priority: wo.priority,
          status: wo.status,
          assigned_to: wo.assignee?.name || "Unassigned",
          due_date: wo.scheduled_date || wo.created_at,
        })),
      );
    } catch (err: any) {
      setError(err?.message || "Failed to load work orders");
      setWorkOrders([]);
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    loadWorkOrders();
  }, []);

  const moveToInProgress = async (id: string) => {
    try {
      await apiClient.patch(`/admin/work-orders/${id}/status`, {
        status: "IN_PROGRESS",
      });
      await loadWorkOrders();
    } catch (err: any) {
      setError(err?.message || "Failed to update status");
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case "CRITICAL":
        return "bg-red-500/20 text-red-400 border-red-500/50";
      case "HIGH":
        return "bg-orange-500/20 text-orange-400 border-orange-500/50";
      case "MEDIUM":
        return "bg-yellow-500/20 text-yellow-400 border-yellow-500/50";
      default:
        return "bg-blue-500/20 text-blue-400 border-blue-500/50";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "COMPLETED":
        return "text-green-400";
      case "IN_PROGRESS":
        return "text-blue-400";
      case "ON_HOLD":
        return "text-yellow-400";
      default:
        return "text-slate-400";
    }
  };

  return (
    <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-white">
      {/* Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <WrenchScrewdriverIcon className="h-8 w-8 text-blue-500" />
            Maintenance Work Orders
          </h1>
          <p className="text-slate-400">
            Manage repair tasks, assign technicians, and track progress.
          </p>
        </div>
        <div className="flex gap-3">
          <button className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg border border-slate-700 flex items-center gap-2 transition">
            <FunnelIcon className="h-5 w-5" />
            Filter
          </button>
          <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg flex items-center gap-2 transition shadow-lg shadow-blue-500/20">
            <PlusIcon className="h-5 w-5" />
            Create Work Order
          </button>
        </div>
      </div>

      {/* Kanban / List Board */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Column: OPEN */}
        <div className="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
          <h3 className="font-semibold text-slate-300 mb-4 flex items-center justify-between">
            To Do / Open
            <span className="bg-slate-700 text-xs px-2 py-1 rounded-full">
              {workOrders.filter((w) => w.status === "OPEN").length}
            </span>
          </h3>
          <div className="space-y-3">
            {loading && (
              <div className="text-xs text-slate-500">Loading...</div>
            )}
            {!loading &&
              workOrders.filter((w) => w.status === "OPEN").length === 0 && (
                <div className="text-xs text-slate-500">
                  No open work orders
                </div>
              )}
            {workOrders
              .filter((w) => w.status === "OPEN")
              .map((wo) => (
                <WorkOrderCard
                  key={wo.id}
                  wo={wo}
                  getPriorityColor={getPriorityColor}
                  onMove={moveToInProgress}
                />
              ))}
          </div>
        </div>

        {/* Column: IN PROGRESS */}
        <div className="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
          <h3 className="font-semibold text-blue-300 mb-4 flex items-center justify-between">
            In Progress
            <span className="bg-blue-900/50 text-blue-200 text-xs px-2 py-1 rounded-full">
              {workOrders.filter((w) => w.status === "IN_PROGRESS").length}
            </span>
          </h3>
          <div className="space-y-3">
            {!loading &&
              workOrders.filter((w) => w.status === "IN_PROGRESS").length ===
                0 && (
                <div className="text-xs text-slate-500">
                  No in-progress work orders
                </div>
              )}
            {workOrders
              .filter((w) => w.status === "IN_PROGRESS")
              .map((wo) => (
                <WorkOrderCard
                  key={wo.id}
                  wo={wo}
                  getPriorityColor={getPriorityColor}
                  onMove={moveToInProgress}
                />
              ))}
          </div>
        </div>

        {/* Column: ON HOLD / BLOCKED */}
        <div className="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
          <h3 className="font-semibold text-yellow-300 mb-4 flex items-center justify-between">
            On Hold / Blocked
            <span className="bg-yellow-900/50 text-yellow-200 text-xs px-2 py-1 rounded-full">
              {workOrders.filter((w) => w.status === "ON_HOLD").length}
            </span>
          </h3>
          <div className="space-y-3">
            {!loading &&
              workOrders.filter((w) => w.status === "ON_HOLD").length === 0 && (
                <div className="text-xs text-slate-500">
                  No blocked work orders
                </div>
              )}
            {workOrders
              .filter((w) => w.status === "ON_HOLD")
              .map((wo) => (
                <WorkOrderCard
                  key={wo.id}
                  wo={wo}
                  getPriorityColor={getPriorityColor}
                  onMove={moveToInProgress}
                />
              ))}
          </div>
        </div>
      </div>
      {error && (
        <div className="text-sm text-red-300 bg-red-500/10 border border-red-500/30 rounded px-3 py-2">
          {error}
        </div>
      )}
    </div>
  );
};

// Sub-component for individual cards
const WorkOrderCard = ({
  wo,
  getPriorityColor,
  onMove,
}: {
  wo: WorkOrder;
  getPriorityColor: (p: string) => string;
  onMove: (id: string) => void;
}) => (
  <div className="bg-slate-800 p-4 rounded-lg border border-slate-700 hover:border-blue-500/50 transition group cursor-pointer shadow-sm">
    <div className="flex justify-between items-start mb-2">
      <span
        className={`text-[10px] font-bold px-2 py-0.5 rounded border ${getPriorityColor(wo.priority)}`}
      >
        {wo.priority}
      </span>
      <span className="text-xs text-slate-500 font-mono">{wo.id}</span>
    </div>
    <h4 className="font-bold text-white mb-1 group-hover:text-blue-400 transition">
      {wo.title}
    </h4>
    <p className="text-sm text-slate-400 mb-3">
      {wo.vehicle_name}{" "}
      <span className="text-slate-600">({wo.vehicle_id})</span>
    </p>

    <div className="flex items-center justify-between text-xs text-slate-500 border-t border-slate-700/50 pt-3 mt-2">
      <div className="flex items-center gap-1">
        <div className="h-5 w-5 rounded-full bg-slate-700 flex items-center justify-center text-[9px] text-white">
          {wo.assigned_to.charAt(0)}
        </div>
        <span>
          {wo.assigned_to === "Unassigned"
            ? "Unassigned"
            : wo.assigned_to.split(" ")[0]}
        </span>
      </div>
      <div className="flex items-center gap-1">
        <ClockIcon className="h-3 w-3" />
        {new Date(wo.due_date).toLocaleDateString()}
      </div>
    </div>
    {wo.status === "OPEN" && (
      <button
        onClick={() => onMove(wo.id)}
        className="mt-3 w-full rounded bg-blue-600/20 text-blue-300 hover:bg-blue-600/30 text-xs px-2 py-1"
      >
        Move To In Progress
      </button>
    )}
  </div>
);

export default WorkOrderManagement;
