import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { setSelectedVehicle, updateTelemetryBatch } from '../../../store/telemetrySlice';
import { fleetService } from '../FleetService';
import { Vehicle } from '../types';

/**
 * FleetList: A high-performance reactive grid for the vehicle registry.
 * Leverages FleetService for worker-offloaded search and filtering.
 */
interface FleetListProps {
  variant?: 'full' | 'compact';
  gpsOnly?: boolean;
  hideHeader?: boolean;
}

export const FleetList: React.FC<FleetListProps> = ({ 
  variant = 'full', 
  gpsOnly = false,
  hideHeader = false 
}) => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState<Record<string, string>>(
    gpsOnly ? { gps_enabled: 'true' } : {}
  );
  const [metadata, setMetadata] = useState<{
    orgs: string[];
    statuses: string[];
  }>({ orgs: [], statuses: [] });
  const selectedId = useSelector((state: any) => state.telemetry.selectedVehicleId);
  const user = useSelector((state: any) => state.auth.user);
  const isAdmin = user?.role === 'super_admin' || user?.role === 'global_admin';

  useEffect(() => {
    const init = async () => {
      setLoading(true);
      try {
        const meta = await fleetService.getFilterMetadata();
        setMetadata(meta);
        const results = await fleetService.searchVehicles('', filters);
        setVehicles(results as Vehicle[]);

        // Sync with Map Registry
        dispatch(updateTelemetryBatch(results as any));
      } catch (err) {
        console.error('[FleetList] Initialization failed:', err);
      } finally {
        setLoading(false);
      }
    };
    init();
  }, []);

  const handleSearch = async (query: string) => {
    setSearchQuery(query);
    const results = await fleetService.searchVehicles(query, filters);
    setVehicles(results as Vehicle[]);
    dispatch(updateTelemetryBatch(results as any));
  };

  const handleFilterChange = async (key: string, value: string) => {
    const newFilters = { ...filters, [key]: value };
    if (!value) delete newFilters[key];
    setFilters(newFilters);
    const results = await fleetService.searchVehicles(searchQuery, newFilters);
    setVehicles(results as Vehicle[]);
    dispatch(updateTelemetryBatch(results as any));
  };

  const handleVehicleSelect = (id: string | null) => {
    dispatch(setSelectedVehicle(id));
    if (id && variant !== 'compact') {
      // Navigate to details only in full view
      navigate(`/admin/vehicles/${id}`);
    }
  };

  return (
    <div className="flex flex-col h-full bg-surface text-white">
      {/* Header & Controls */}
      {(variant === 'full' && !hideHeader) && (
        <div className="p-6 border-b border-border-color bg-surface-dark">
          <div className="flex justify-between items-center mb-6">
            <h1 className="text-2xl font-bold tracking-tight">Fleet Registry</h1>
            <div className="flex gap-2">
              <button className="px-4 py-2 bg-primary hover:bg-primary-hover rounded-lg font-medium transition-all">
                Export CSV
              </button>
              <button
                className="px-4 py-2 bg-surface-darker border border-border-color rounded-lg font-medium"
                onClick={() => navigate('/admin/vehicles/new')}
              >
                Add Asset
              </button>
            </div>
          </div>
        </div>
      )}

      <div className={`${variant === 'compact' ? 'p-4' : 'p-6'} space-y-4`}>
        <div className="flex flex-wrap gap-4">
          <div className="relative flex-1 min-w-[300px]">
            <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">
              search
            </span>
            <input
              type="text"
              placeholder="Search VIN, ID, or Fleet Name..."
              className="w-full bg-surface-darker border border-border-color rounded-lg pl-10 pr-4 py-2 focus:ring-1 focus:ring-primary outline-none transition-all"
              value={searchQuery}
              onChange={(e) => handleSearch(e.target.value)}
            />
          </div>

          <select
            className="bg-surface-darker border border-border-color rounded-lg px-4 py-2 text-sm outline-none"
            onChange={(e) => handleFilterChange('status', e.target.value)}
          >
            <option value="">All Statuses</option>
            {metadata?.statuses?.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>

          {isAdmin && (
            <select
              className="bg-surface-darker border border-border-color rounded-lg px-4 py-2 text-sm outline-none"
              onChange={(e) => handleFilterChange('org', e.target.value)}
            >
              <option value="">All Organizations</option>
              {metadata?.orgs?.map((o) => (
                <option key={o} value={o}>
                  {o}
                </option>
              ))}
            </select>
          )}
        </div>
      </div>

      {/* Grid / Table */}
      <div className="flex-1 overflow-auto p-6">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : (
          <div
            className={
              variant === 'compact'
                ? 'flex flex-col gap-2'
                : 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4'
            }
          >
            {vehicles.map((v) => (
              <div
                key={v.id}
                onClick={() => handleVehicleSelect(v.id)}
                className={`bg-surface-darker/50 border transition-all cursor-pointer group ${
                  selectedId === v.id
                    ? 'border-primary bg-primary/5'
                    : 'border-border-color hover:border-primary/30'
                } ${variant === 'compact' ? 'p-3 rounded-lg flex items-center justify-between' : 'p-4 rounded-xl'}`}
              >
                <div className="flex items-center gap-3 overflow-hidden">
                  <div
                    className={`w-8 h-8 rounded-lg flex items-center justify-center transition-transform ${
                      v.status === 'STALLED'
                        ? 'bg-red-500/10 text-red-500 animate-pulse'
                        : 'bg-primary/10 text-primary'
                    }`}
                  >
                    <span className="material-symbols-outlined text-lg">
                      {v.status === 'STALLED' ? 'warning' : 'local_shipping'}
                    </span>
                  </div>
                  <div className="overflow-hidden">
                    <h3 className="font-bold text-sm truncate">
                      {v.registrationNumber || v.name || `Asset ${v.id.substring(0, 8)}`}
                    </h3>
                    <p className="text-[10px] text-text-muted truncate opacity-70">
                      {v.status === 'STALLED'
                        ? 'CRITICAL ALERT'
                        : `Last update: ${new Date(v.lastUpdate).toLocaleTimeString()}`}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  {variant === 'full' && (
                    <div className="text-right hidden sm:block">
                      <p className="text-[10px] text-text-muted uppercase tracking-widest font-bold">
                        Capacity
                      </p>
                      <p className="text-xs font-mono">{v.metadata?.capacity || 'N/A'}T</p>
                    </div>
                  )}
                  <span
                    className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-tighter ${
                      v.status === 'ACTIVE' || (v.telemetry?.speed ?? 0) > 0
                        ? 'bg-emerald-500/10 text-emerald-400'
                        : v.status === 'STALLED'
                          ? 'bg-red-500/10 text-red-400'
                          : 'bg-yellow-500/10 text-yellow-400'
                    }`}
                  >
                    {v.status}
                  </span>
                </div>
              </div>
            ))}
            {vehicles.length === 0 && (
              <div className="col-span-full flex flex-col items-center justify-center py-20 text-text-muted">
                <span className="material-symbols-outlined text-6xl opacity-20 mb-4">
                  inventory_2
                </span>
                <p>No assets found in the current registry.</p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
