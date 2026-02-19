import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../services/apiClient';

// Types
interface WorkOrder {
  id: string;
  vehicle_id: string;
  title: string;
  description: string;
  maintenance_type: 'preventive' | 'corrective' | 'inspection' | 'breakdown';
  priority: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  status: 'OPEN' | 'IN_PROGRESS' | 'ON_HOLD' | 'COMPLETED';
  assigned_to?: string;
  scheduled_date?: string;
  estimated_cost: number;
  actual_cost: number;
  approval_status: 'NOT_REQUIRED' | 'PENDING_MANAGER' | 'PENDING_FLEET_HEAD' | 'PENDING_FINANCE' | 'APPROVED' | 'REJECTED';
  vendor_id?: string;
  vendor?: {
    name: string;
  };
  created_at: string;
  vehicle?: {
    make: string;
    model: string;
    license_plate: string;
  };
  assignee?: {
    name: string;
    initials: string;
  };
}

interface WorkOrderFilter {
  status: string[];
  priority: string[];
  search: string;
  page: number;
  pageSize: number;
}

const statusColors: Record<string, string> = {
  OPEN: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  IN_PROGRESS: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  ON_HOLD: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  COMPLETED: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
};

const approvalStatusColors: Record<string, string> = {
  NOT_REQUIRED: 'bg-gray-500/10 text-gray-400 border-transparent',
  PENDING_MANAGER: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  PENDING_FLEET_HEAD: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  PENDING_FINANCE: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  APPROVED: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  REJECTED: 'bg-red-500/20 text-red-400 border-red-500/30',
};

const priorityColors: Record<string, string> = {
  LOW: 'bg-gray-500/20 text-gray-400',
  MEDIUM: 'bg-blue-500/20 text-blue-400',
  HIGH: 'bg-amber-500/20 text-amber-400',
  CRITICAL: 'bg-red-500/20 text-red-400',
};

const WorkOrderList: React.FC = () => {
  const navigate = useNavigate();
  const [workOrders, setWorkOrders] = useState<WorkOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [filter, setFilter] = useState<WorkOrderFilter>({
    status: [],
    priority: [],
    search: '',
    page: 1,
    pageSize: 20,
  });
  const [stats, setStats] = useState({
    activeWorkOrders: 0,
    avgResolutionTime: 0,
    slaBreachesMTD: 0,
    lowStockAlerts: 0,
  });

  const fetchWorkOrders = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, any> = {
        page: filter.page,
        page_size: filter.pageSize,
      };
      if (filter.search) params.search = filter.search;
      if (filter.status.length) params.status = filter.status;
      if (filter.priority.length) params.priority = filter.priority;
      const data = await apiClient.get('/admin/work-orders', { params }) as any;
      setWorkOrders(data.work_orders || data.data || []);
      setTotal(data.total || 0);
    } catch (error) {
      console.error('Failed to fetch work orders:', error);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  const fetchStats = useCallback(async () => {
    try {
      const data = await apiClient.get('/admin/work-orders/stats') as any;
      setStats({
        activeWorkOrders: data.active_work_orders || 0,
        avgResolutionTime: data.avg_resolution_time_hours || 0,
        slaBreachesMTD: data.sla_breaches_mtd || 0,
        lowStockAlerts: data.low_stock_alerts || 0,
      });
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  }, []);

  useEffect(() => {
    fetchWorkOrders();
    fetchStats();
  }, [fetchWorkOrders, fetchStats]);

  const handleStatusFilter = (status: string) => {
    setFilter(prev => ({
      ...prev,
      status: prev.status.includes(status)
        ? prev.status.filter(s => s !== status)
        : [...prev.status, status],
      page: 1,
    }));
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white">Work Orders</h1>
            <p className="text-sm text-text-muted mt-1">Manage maintenance work orders</p>
          </div>
          <button
            onClick={() => navigate('/maintenance/work-orders/new')}
            className="bg-primary hover:bg-primary/90 text-white text-sm font-medium py-2 px-4 rounded-lg flex items-center gap-2 shadow-[0_0_15px_rgba(19,127,236,0.3)] transition-all"
          >
            <span className="material-symbols-outlined text-lg">add</span>
            New Work Order
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="px-6 py-4 grid grid-cols-1 md:grid-cols-4 gap-4 flex-shrink-0">
        <div className="bg-surface-dark border border-border-color rounded-xl p-4">
          <div className="flex items-center justify-between">
            <span className="material-symbols-outlined text-primary text-2xl">engineering</span>
            <span className="text-xs text-emerald-400 bg-emerald-500/20 px-2 py-0.5 rounded">Active</span>
          </div>
          <p className="text-3xl font-bold text-white mt-3">{stats.activeWorkOrders}</p>
          <p className="text-sm text-text-muted">Active Work Orders</p>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-xl p-4">
          <div className="flex items-center justify-between">
            <span className="material-symbols-outlined text-amber-400 text-2xl">schedule</span>
          </div>
          <p className="text-3xl font-bold text-white mt-3">{stats.avgResolutionTime.toFixed(1)}h</p>
          <p className="text-sm text-text-muted">Avg Resolution Time</p>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-xl p-4">
          <div className="flex items-center justify-between">
            <span className="material-symbols-outlined text-red-400 text-2xl">warning</span>
          </div>
          <p className="text-3xl font-bold text-white mt-3">{stats.slaBreachesMTD}</p>
          <p className="text-sm text-text-muted">SLA Breaches MTD</p>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-xl p-4">
          <div className="flex items-center justify-between">
            <span className="material-symbols-outlined text-blue-400 text-2xl">inventory_2</span>
          </div>
          <p className="text-3xl font-bold text-white mt-3">{stats.lowStockAlerts}</p>
          <p className="text-sm text-text-muted">Low Stock Alerts</p>
        </div>
      </div>

      {/* Filters */}
      <div className="px-6 py-3 flex flex-wrap items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-lg">search</span>
          <input
            type="text"
            placeholder="Search work orders..."
            value={filter.search}
            onChange={(e) => setFilter(prev => ({ ...prev, search: e.target.value, page: 1 }))}
            className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-10 pr-4 py-2 focus:ring-1 focus:ring-primary focus:border-primary"
          />
        </div>
        <div className="flex items-center gap-2">
          {['OPEN', 'IN_PROGRESS', 'ON_HOLD', 'COMPLETED'].map(status => (
            <button
              key={status}
              onClick={() => handleStatusFilter(status)}
              className={`px-3 py-1.5 text-xs font-medium rounded-full border transition-all ${filter.status.includes(status)
                ? statusColors[status]
                : 'bg-surface-dark border-border-color text-text-muted hover:border-primary/50'
                }`}
            >
              {status.replace('_', ' ')}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <div className="flex-1 overflow-y-auto px-6 pb-6">
        <div className="bg-surface-dark border border-border-color rounded-xl overflow-hidden">
          {loading ? (
            <div className="flex items-center justify-center py-20">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : workOrders.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-text-muted">
              <span className="material-symbols-outlined text-4xl mb-3">assignment</span>
              <p>No work orders found</p>
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-surface-darker">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Title</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Vehicle</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Type</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Priority</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Status</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Approval</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Scheduled</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Assignee</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-color">
                {workOrders.map((wo) => (
                  <tr
                    key={wo.id}
                    className="hover:bg-surface-darker/50 cursor-pointer transition-colors"
                    onClick={() => navigate(`/maintenance/work-orders/${wo.id}`)}
                  >
                    <td className="px-4 py-4">
                      <div className="text-sm font-medium text-white">{wo.title}</div>
                      <div className="text-xs text-text-muted line-clamp-1">{wo.description}</div>
                    </td>
                    <td className="px-4 py-4">
                      <div className="text-sm text-white">
                        {wo.vehicle ? `${wo.vehicle.make} ${wo.vehicle.model}` : wo.vehicle_id.slice(0, 8)}
                      </div>
                      {wo.vehicle && <div className="text-xs text-text-muted">{wo.vehicle.license_plate}</div>}
                    </td>
                    <td className="px-4 py-4">
                      <span className="text-xs text-text-muted capitalize">{wo.maintenance_type}</span>
                    </td>
                    <td className="px-4 py-4">
                      <span className={`px-2 py-1 text-xs font-medium rounded ${priorityColors[wo.priority]}`}>
                        {wo.priority}
                      </span>
                    </td>
                    <td className="px-4 py-4">
                      <span className={`px-2 py-1 text-xs font-medium rounded border ${statusColors[wo.status]}`}>
                        {wo.status.replace('_', ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-4">
                      <span className={`px-2 py-1 text-xs font-medium rounded border ${approvalStatusColors[wo.approval_status]}`}>
                        {wo.approval_status.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-sm text-text-muted">
                      {wo.scheduled_date ? formatDate(wo.scheduled_date) : '-'}
                    </td>
                    <td className="px-4 py-4">
                      {wo.assignee ? (
                        <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-xs font-medium text-primary">
                          {wo.assignee.initials}
                        </div>
                      ) : (
                        <span className="text-xs text-text-muted">Unassigned</span>
                      )}
                    </td>
                    <td className="px-4 py-4 text-right">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          navigate(`/maintenance/work-orders/${wo.id}`);
                        }}
                        className="text-primary hover:text-primary/80 text-sm font-medium"
                      >
                        View
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        {!loading && total > filter.pageSize && (
          <div className="flex items-center justify-between mt-4">
            <p className="text-sm text-text-muted">
              Showing {(filter.page - 1) * filter.pageSize + 1} to {Math.min(filter.page * filter.pageSize, total)} of {total} results
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setFilter(prev => ({ ...prev, page: prev.page - 1 }))}
                disabled={filter.page === 1}
                className="px-3 py-1 text-sm bg-surface-dark border border-border-color rounded text-text-muted hover:text-white disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => setFilter(prev => ({ ...prev, page: prev.page + 1 }))}
                disabled={filter.page * filter.pageSize >= total}
                className="px-3 py-1 text-sm bg-surface-dark border border-border-color rounded text-text-muted hover:text-white disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default WorkOrderList;
