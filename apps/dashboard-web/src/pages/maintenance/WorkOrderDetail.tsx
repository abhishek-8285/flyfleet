import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

interface WorkOrder {
  id: string;
  vehicle_id: string;
  title: string;
  description: string;
  maintenance_type: string;
  priority: string;
  status: string;
  assigned_to?: string;
  scheduled_date?: string;
  completed_date?: string;
  estimated_cost: number;
  actual_cost: number;
  created_at: string;
  updated_at: string;
  parts: WorkOrderPart[];
  vehicle?: { make: string; model: string; license_plate: string };
  assignee?: { name: string; initials: string };
  total_parts_cost: number;
}

interface WorkOrderPart {
  id: string;
  part_id: string;
  quantity: number;
  unit_cost: number;
  total_cost: number;
  part?: { part_number: string; name: string; description: string };
}

const statusColors: Record<string, string> = {
  OPEN: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  IN_PROGRESS: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  ON_HOLD: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  COMPLETED: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
};

const priorityColors: Record<string, string> = {
  LOW: 'bg-gray-500/20 text-gray-400',
  MEDIUM: 'bg-blue-500/20 text-blue-400',
  HIGH: 'bg-amber-500/20 text-amber-400',
  CRITICAL: 'bg-red-500/20 text-red-400',
};

const WorkOrderDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [workOrder, setWorkOrder] = useState<WorkOrder | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

  const fetchWorkOrder = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/work-orders/${id}`);
      if (response.ok) {
        const data = await response.json();
        setWorkOrder(data);
      } else if (response.status === 404) {
        navigate('/maintenance/work-orders');
      }
    } catch (error) {
      console.error('Failed to fetch work order:', error);
    } finally {
      setLoading(false);
    }
  }, [id, navigate]);

  useEffect(() => {
    fetchWorkOrder();
  }, [fetchWorkOrder]);

  const changeStatus = async (newStatus: string) => {
    if (!workOrder) return;
    setActionLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/work-orders/${id}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      if (response.ok) {
        fetchWorkOrder();
      }
    } catch (error) {
      console.error('Failed to change status:', error);
    } finally {
      setActionLoading(false);
    }
  };

  const closeWorkOrder = async () => {
    if (!workOrder || workOrder.status === 'COMPLETED') return;
    setActionLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/work-orders/${id}/close`, {
        method: 'POST',
      });
      if (response.ok) {
        fetchWorkOrder();
      }
    } catch (error) {
      console.error('Failed to close work order:', error);
    } finally {
      setActionLoading(false);
    }
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-background-dark">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!workOrder) {
    return (
      <div className="flex-1 flex items-center justify-center bg-background-dark">
        <p className="text-text-muted">Work order not found</p>
      </div>
    );
  }

  const isCompleted = workOrder.status === 'COMPLETED';

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/maintenance/work-orders')}
              className="p-2 hover:bg-surface-dark rounded-lg transition-colors"
            >
              <span className="material-symbols-outlined text-text-muted">arrow_back</span>
            </button>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-2xl font-bold text-white">{workOrder.title}</h1>
                <span className={`px-2 py-1 text-xs font-medium rounded border ${statusColors[workOrder.status]}`}>
                  {workOrder.status.replace('_', ' ')}
                </span>
                <span className={`px-2 py-1 text-xs font-medium rounded ${priorityColors[workOrder.priority]}`}>
                  {workOrder.priority}
                </span>
              </div>
              <p className="text-sm text-text-muted mt-1">
                Created {formatDate(workOrder.created_at)}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            {!isCompleted && (
              <>
                {workOrder.status === 'OPEN' && (
                  <button
                    onClick={() => changeStatus('IN_PROGRESS')}
                    disabled={actionLoading}
                    className="px-4 py-2 text-sm font-medium bg-amber-500 hover:bg-amber-600 text-white rounded-lg transition-colors disabled:opacity-50"
                  >
                    Start Work
                  </button>
                )}
                {workOrder.status === 'IN_PROGRESS' && (
                  <button
                    onClick={closeWorkOrder}
                    disabled={actionLoading}
                    className="px-4 py-2 text-sm font-medium bg-emerald-500 hover:bg-emerald-600 text-white rounded-lg transition-colors disabled:opacity-50"
                  >
                    Complete
                  </button>
                )}
              </>
            )}
            <button
              onClick={() => navigate(`/maintenance/work-orders/${id}/edit`)}
              disabled={isCompleted}
              className="px-4 py-2 text-sm font-medium border border-border-color text-text-muted hover:text-white rounded-lg transition-colors disabled:opacity-50"
            >
              Edit
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-5xl mx-auto grid grid-cols-3 gap-6">
          {/* Main Info */}
          <div className="col-span-2 space-y-6">
            {/* Description */}
            <div className="bg-surface-dark border border-border-color rounded-xl p-6">
              <h3 className="text-lg font-semibold text-white mb-4">Description</h3>
              <p className="text-text-muted whitespace-pre-wrap">{workOrder.description}</p>
            </div>

            {/* Parts */}
            <div className="bg-surface-dark border border-border-color rounded-xl p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-white">Parts Used</h3>
                {!isCompleted && (
                  <button
                    onClick={() => navigate(`/maintenance/work-orders/${id}/parts`)}
                    className="text-sm text-primary hover:text-primary/80"
                  >
                    + Add Parts
                  </button>
                )}
              </div>
              {workOrder.parts.length === 0 ? (
                <p className="text-text-muted text-sm">No parts assigned yet</p>
              ) : (
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-text-muted uppercase">
                      <th className="pb-3">Part</th>
                      <th className="pb-3">Quantity</th>
                      <th className="pb-3 text-right">Unit Cost</th>
                      <th className="pb-3 text-right">Total</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border-color">
                    {workOrder.parts.map((part) => (
                      <tr key={part.id}>
                        <td className="py-3">
                          <div className="text-sm text-white">{part.part?.name || 'Unknown'}</div>
                          <div className="text-xs text-text-muted">{part.part?.part_number}</div>
                        </td>
                        <td className="py-3 text-sm text-text-muted">{part.quantity}</td>
                        <td className="py-3 text-sm text-text-muted text-right">${part.unit_cost.toFixed(2)}</td>
                        <td className="py-3 text-sm text-white text-right">${part.total_cost.toFixed(2)}</td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot>
                    <tr className="border-t border-border-color">
                      <td colSpan={3} className="pt-3 text-sm font-medium text-white text-right">Total Parts Cost</td>
                      <td className="pt-3 text-sm font-bold text-primary text-right">${workOrder.total_parts_cost.toFixed(2)}</td>
                    </tr>
                  </tfoot>
                </table>
              )}
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Vehicle */}
            <div className="bg-surface-dark border border-border-color rounded-xl p-6">
              <h3 className="text-sm font-medium text-text-muted mb-3">Vehicle</h3>
              {workOrder.vehicle ? (
                <div>
                  <p className="text-white font-medium">{workOrder.vehicle.make} {workOrder.vehicle.model}</p>
                  <p className="text-sm text-text-muted">{workOrder.vehicle.license_plate}</p>
                </div>
              ) : (
                <p className="text-sm text-text-muted">{workOrder.vehicle_id.slice(0, 8)}</p>
              )}
            </div>

            {/* Assignee */}
            <div className="bg-surface-dark border border-border-color rounded-xl p-6">
              <h3 className="text-sm font-medium text-text-muted mb-3">Assignee</h3>
              {workOrder.assignee ? (
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center text-sm font-medium text-primary">
                    {workOrder.assignee.initials}
                  </div>
                  <span className="text-white">{workOrder.assignee.name}</span>
                </div>
              ) : (
                <p className="text-sm text-text-muted">Unassigned</p>
              )}
            </div>

            {/* Dates */}
            <div className="bg-surface-dark border border-border-color rounded-xl p-6 space-y-4">
              <div>
                <h4 className="text-sm font-medium text-text-muted">Scheduled</h4>
                <p className="text-white">{workOrder.scheduled_date ? formatDate(workOrder.scheduled_date) : '-'}</p>
              </div>
              {workOrder.completed_date && (
                <div>
                  <h4 className="text-sm font-medium text-text-muted">Completed</h4>
                  <p className="text-emerald-400">{formatDate(workOrder.completed_date)}</p>
                </div>
              )}
            </div>

            {/* Costs */}
            <div className="bg-surface-dark border border-border-color rounded-xl p-6 space-y-4">
              <div>
                <h4 className="text-sm font-medium text-text-muted">Estimated Cost</h4>
                <p className="text-2xl font-bold text-white">${workOrder.estimated_cost.toFixed(2)}</p>
              </div>
              <div className="border-t border-border-color pt-4">
                <h4 className="text-sm font-medium text-text-muted">Actual Cost</h4>
                <p className="text-2xl font-bold text-primary">${workOrder.actual_cost.toFixed(2)}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkOrderDetail;
