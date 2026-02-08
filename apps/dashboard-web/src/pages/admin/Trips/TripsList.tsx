import {
    ArrowDownTrayIcon,
    ArrowPathIcon,
    ChevronLeftIcon, ChevronRightIcon,
    PlusIcon,
    TruckIcon
} from '@heroicons/react/24/outline';
import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import api from '../../../services/api';

interface Trip {
  id: string;
  trip_id: string;
  vehicle_id: string;
  driver_id?: string;
  start_location: string;
  end_location: string;
  start_time: string;
  distance: number;
  status: string;
}

const statusColors: Record<string, string> = {
  scheduled: 'bg-primary/10 text-primary border-primary/20',
  in_progress: 'bg-warning/10 text-warning border-warning/20',
  completed: 'bg-success/10 text-success border-success/20',
  cancelled: 'bg-danger/10 text-danger border-danger/20',
};

export default function TripsList() {
  const navigate = useNavigate();
  const [trips, setTrips] = useState<Trip[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('');

  useEffect(() => {
    loadTrips();
  }, [page, statusFilter]);

  const loadTrips = async () => {
    try {
      setLoading(true);
      const data = await api.getTrips({ page, limit: 20, status: statusFilter || undefined });
      setTrips(data.trips || []);
      setTotal(data.total || 0);
    } catch (error) {
      console.error('Failed to load trips:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', { 
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' 
    });
  };

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <span>Fleet Management</span>
            <ChevronRightIcon className="w-3 h-3" />
            <span className="text-white">Trips</span>
          </div>
          <h1 className="text-2xl font-bold text-white">Trip List View</h1>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-surface-dark border border-border-color text-sm text-white rounded px-3 py-2"
          >
            <option value="">All Statuses</option>
            <option value="scheduled">Scheduled</option>
            <option value="in_progress">In Progress</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <Link
            to="/admin/trips/add"
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <PlusIcon className="w-4 h-4" />
            Add Trip
          </Link>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">Total Trips</div>
          <div className="text-2xl font-bold text-white">{total}</div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">In Progress</div>
          <div className="text-2xl font-bold text-warning">{trips.filter(t => t.status === 'in_progress').length}</div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">Scheduled</div>
          <div className="text-2xl font-bold text-primary">{trips.filter(t => t.status === 'scheduled').length}</div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">Completed</div>
          <div className="text-2xl font-bold text-success">{trips.filter(t => t.status === 'completed').length}</div>
        </div>
      </div>

      {/* Table */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
          <div className="flex items-center gap-3">
            <h3 className="text-white font-semibold">Active Trips</h3>
            <span className="bg-primary/20 text-primary text-xs px-2 py-0.5 rounded-full font-bold">
              {trips.filter(t => t.status === 'in_progress').length} Live
            </span>
          </div>
          <div className="flex items-center gap-2">
            <button onClick={loadTrips} className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white">
              <ArrowPathIcon className="w-4 h-4" />
            </button>
            <button className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white">
              <ArrowDownTrayIcon className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead className="border-b border-border-color text-xs text-text-muted uppercase bg-surface-darker">
              <tr>
                <th className="px-4 py-3 font-semibold">Trip ID</th>
                <th className="px-4 py-3 font-semibold">Vehicle</th>
                <th className="px-4 py-3 font-semibold">Start Location</th>
                <th className="px-4 py-3 font-semibold">End Location</th>
                <th className="px-4 py-3 font-semibold">Start Time</th>
                <th className="px-4 py-3 font-semibold">Distance</th>
                <th className="px-4 py-3 font-semibold text-right">Status</th>
              </tr>
            </thead>
            <tbody className="text-sm">
              {loading ? (
                <tr><td colSpan={7} className="px-4 py-8 text-center text-text-muted">Loading...</td></tr>
              ) : trips.length === 0 ? (
                <tr><td colSpan={7} className="px-4 py-8 text-center text-text-muted">No trips found</td></tr>
              ) : trips.map((trip) => (
                <tr key={trip.id} className="border-b border-border-color hover:bg-surface-darker/50 transition-colors">
                  <td className="px-4 py-3">
                    <Link to={`/admin/trips/${trip.id}`} className="font-mono text-primary hover:underline">
                      {trip.trip_id}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-white flex items-center gap-2">
                    <TruckIcon className="w-4 h-4 text-text-muted" />
                    {trip.vehicle_id?.slice(0, 8)}...
                  </td>
                  <td className="px-4 py-3 text-text-muted">{trip.start_location}</td>
                  <td className="px-4 py-3 text-text-muted">{trip.end_location}</td>
                  <td className="px-4 py-3 text-text-muted whitespace-nowrap">{formatDate(trip.start_time)}</td>
                  <td className="px-4 py-3 text-text-muted font-mono">{trip.distance?.toFixed(1) || '---'} km</td>
                  <td className="px-4 py-3 text-right">
                    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-bold border ${statusColors[trip.status] || statusColors.scheduled}`}>
                      {trip.status === 'in_progress' && <span className="w-1.5 h-1.5 rounded-full bg-warning animate-pulse"></span>}
                      {trip.status?.toUpperCase().replace('_', ' ')}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="p-3 border-t border-border-color bg-surface-darker flex items-center justify-between">
          <span className="text-xs text-text-muted">
            Showing {((page - 1) * 20) + 1}-{Math.min(page * 20, total)} of {total} trips
          </span>
          <div className="flex items-center gap-1">
            <button 
              onClick={() => setPage(p => Math.max(1, p - 1))} 
              disabled={page === 1}
              className="p-1 rounded text-text-muted hover:text-white hover:bg-border-color disabled:opacity-50"
            >
              <ChevronLeftIcon className="w-4 h-4" />
            </button>
            <span className="px-2 py-1 bg-primary text-white text-xs font-bold rounded">{page}</span>
            <button 
              onClick={() => setPage(p => p + 1)} 
              disabled={page * 20 >= total}
              className="p-1 rounded text-text-muted hover:text-white hover:bg-border-color disabled:opacity-50"
            >
              <ChevronRightIcon className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
