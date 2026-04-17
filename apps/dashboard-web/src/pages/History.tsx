import branding from '@flyfleet/config';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useUIStrings } from '../hooks/useUIStrings';
import { AppDispatch, RootState } from '../store';
import { fetchTrips } from '../store/tripsSlice';
import { Trip } from '../types';

const History: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { trips, loading } = useSelector((state: RootState) => state.trips);
  const [viewMode, setViewMode] = useState<'list' | 'timeline'>('list');
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings } = useUIStrings('history');

  useEffect(() => {
    dispatch(fetchTrips({}) as any);
    fetch('/api/v1/ui/global_config')
      .then((res) => res.json())
      .then((data) => setSystemMeta(data.system_metadata))
      .catch((err) => console.error('Global config fetch failed', err));
  }, [dispatch]);

  const formatDuration = (trip: Trip) => {
    if (!trip.end_time) return '-';
    const start = new Date(trip.start_time).getTime();
    const end = new Date(trip.end_time).getTime();
    if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return '-';
    const totalMinutes = Math.floor((end - start) / 60000);
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };
  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-surface-dim h-full font-sans antialiased">
      {/* Tactical Control Bar */}
      <div className="p-8 border-b border-outline bg-surface-container/60 backdrop-blur-3xl flex flex-wrap gap-6 items-center z-10 shadow-2xl">
        <div className="flex bg-surface-dim rounded-2xl border border-outline p-1.5 shadow-inner">
          <button
            onClick={() => setViewMode('list')}
            className={`px-6 py-2 text-[10px] font-black uppercase tracking-widest rounded-xl transition-all duration-300 ${viewMode === 'list' ? 'bg-primary text-white shadow-lg' : 'text-on-surface-variant hover:text-white'}`}
          >
            {uiStrings?.btn_list || 'Matrix View'}
          </button>
          <button
            onClick={() => setViewMode('timeline')}
            className={`px-6 py-2 text-[10px] font-black uppercase tracking-widest rounded-xl transition-all duration-300 ${viewMode === 'timeline' ? 'bg-primary text-white shadow-lg' : 'text-on-surface-variant hover:text-white'}`}
          >
            {uiStrings?.btn_timeline || 'Timeline HUD'}
          </button>
        </div>

        <div className="ml-auto flex gap-4">
          <div className="relative group">
            <div className="absolute inset-y-0 left-4 flex items-center pointer-events-none">
              <span className="material-symbols-outlined text-sm text-on-surface-variant group-hover:text-primary transition-colors">
                calendar_month
              </span>
            </div>
            <input
              type="date"
              className="bg-surface-dim border border-outline text-white text-[10px] font-black uppercase tracking-widest rounded-xl pl-12 pr-4 py-2.5 focus:border-primary [color-scheme:dark] shadow-sm outline-none transition-all hover:bg-surface-container-highest/20"
            />
          </div>
          <button className="h-11 px-6 bg-surface-container-highest border border-outline text-[10px] font-black uppercase tracking-widest text-white hover:bg-surface-on-variant rounded-xl transition-all active:scale-95 flex items-center gap-3 shadow-lg">
            <span className="material-symbols-outlined text-lg text-primary">filter_list</span>
            {uiStrings?.btn_filter || 'Sector Filter'}
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-10 custom-scrollbar">
        <div className="bg-surface-container rounded-[3rem] border border-outline overflow-hidden shadow-2xl max-w-[1600px] mx-auto relative group">
          {loading ? (
            <div className="p-24 text-center">
              <div className="w-16 h-16 rounded-full border-4 border-primary/20 border-t-primary animate-spin mx-auto mb-6" />
              <p className="text-[10px] font-black uppercase tracking-[0.3em] text-on-surface-variant">
                {uiStrings?.loading || 'Synchronizing Historical Telemetry...'}
              </p>
            </div>
          ) : (
            <table className="w-full text-left text-sm text-on-surface-variant border-collapse">
              <thead className="bg-surface-dim border-b border-outline">
                <tr>
                  <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em] text-white">
                    {uiStrings?.th_datetime || 'Operational Window'}
                  </th>
                  <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">
                    {uiStrings?.th_vehicle || 'Assigned Asset'}
                  </th>
                  <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">
                    {uiStrings?.th_driver || 'Personnel'}
                  </th>
                  <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">
                    {uiStrings?.th_duration || 'Active Duration'}
                  </th>
                  <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">
                    {uiStrings?.th_distance || 'Total Spectrum (KM)'}
                  </th>
                  <th className="px-10 py-6 text-right text-[10px] font-black uppercase tracking-[0.2em]">
                    {uiStrings?.th_route || 'Directives'}
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-outline bg-surface-container/40">
                {trips.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-10 py-32 text-center">
                      <div className="flex flex-col items-center gap-6 opacity-30">
                        <span className="material-symbols-outlined text-7xl text-primary">
                          history
                        </span>
                        <p className="text-[12px] font-black uppercase tracking-[0.4em]">
                          {uiStrings?.no_history || 'No historical log signatures detected.'}
                        </p>
                      </div>
                    </td>
                  </tr>
                ) : (
                  trips.map((trip: Trip) => (
                    <tr
                      key={trip.id}
                      className="hover:bg-surface-container-highest/50 transition-all duration-300 group cursor-pointer"
                    >
                      <td className="px-10 py-6">
                        <div className="text-white font-bold text-base tracking-tight mb-1">
                          {new Date(trip.start_time).toLocaleDateString([], {
                            month: 'short',
                            day: 'numeric',
                            year: 'numeric',
                          })}
                        </div>
                        <div className="text-[10px] text-primary font-black uppercase tracking-widest font-mono">
                          {new Date(trip.start_time).toLocaleTimeString([], {
                            hour: '2-digit',
                            minute: '2-digit',
                            second: '2-digit',
                          })}
                        </div>
                      </td>
                      <td className="px-10 py-6">
                        <div className="flex items-center gap-4">
                          <div className="h-10 w-10 bg-surface-dim border border-outline rounded-xl flex items-center justify-center shadow-inner group-hover:scale-110 transition-transform">
                            <span className="material-symbols-outlined text-lg text-primary">
                              local_shipping
                            </span>
                          </div>
                          <span className="text-white font-bold tracking-tight">
                            ASSET-{trip.vehicle_id.slice(-6).toUpperCase()}
                          </span>
                        </div>
                      </td>
                      <td className="px-10 py-6">
                        <div className="flex items-center gap-4">
                          <div className="w-8 h-8 rounded-[10px] bg-surface-container-highest border border-outline flex items-center justify-center text-[10px] font-black text-white uppercase tracking-widest">
                            {trip.driver_id.slice(0, 1).toUpperCase()}
                          </div>
                          <span className="text-[11px] font-bold text-white tracking-widest uppercase">
                            {trip.driver_id.slice(0, 12)}
                          </span>
                        </div>
                      </td>
                      <td className="px-10 py-6">
                        <span className="text-white font-mono font-bold text-sm tracking-tighter">
                          {formatDuration(trip)}
                        </span>
                      </td>
                      <td className="px-10 py-6">
                        <span className="text-primary font-mono font-bold text-sm tracking-tighter uppercase whitespace-nowrap">
                          {trip.distance_km ? `${trip.distance_km.toFixed(2)} KM` : '---'}
                        </span>
                      </td>
                      <td className="px-10 py-6 text-right">
                        <button className="h-10 px-6 bg-surface-dim border border-outline text-[10px] font-black uppercase tracking-widest text-primary hover:bg-primary hover:text-white rounded-xl transition-all active:scale-95 flex items-center gap-3 justify-end ml-auto shadow-md">
                          {uiStrings?.btn_replay || 'Replay Cycle'}
                          <span className="material-symbols-outlined text-lg">play_circle</span>
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          )}
        </div>

        <footer className="mt-8 text-center text-xs text-[#64748b] pb-6 flex flex-col items-center gap-1">
          <span>
            © {new Date().getFullYear()} {branding.companyName} Command. All rights reserved.
          </span>
          <span className="text-[10px]">
            {systemMeta?.version || 'v--'} | Server Time:{' '}
            <span className="font-mono">{systemMeta?.server_time || '---'}</span>
          </span>
        </footer>
      </div>
    </div>
  );
};

export default History;
