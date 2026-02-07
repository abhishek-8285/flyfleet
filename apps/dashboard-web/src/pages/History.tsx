import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { AppDispatch, RootState } from '../store';
import { fetchTrips } from '../store/tripsSlice';
import { Trip } from '../types';

const History: React.FC = () => {
    const dispatch = useDispatch<AppDispatch>();
    const { trips, loading } = useSelector((state: RootState) => state.trips);
    const [viewMode, setViewMode] = useState<'list' | 'timeline'>('list');

    useEffect(() => {
        dispatch(fetchTrips({}) as any);
    }, [dispatch]);

    return (
        <div className="flex-1 flex flex-col overflow-hidden bg-background-dark h-full">
            {/* Control Bar */}
            <div className="p-6 border-b border-border-color bg-surface-darker/50 flex flex-wrap gap-4 items-center">
                <div className="flex bg-surface-dark rounded border border-border-color p-1 shadow-inner">
                    <button 
                        onClick={() => setViewMode('list')} 
                        className={`px-4 py-1.5 text-xs font-medium rounded transition-all duration-200 ${viewMode === 'list' ? 'bg-primary text-white shadow-md' : 'text-text-muted hover:text-white'}`}
                    >
                        List View
                    </button>
                    <button 
                        onClick={() => setViewMode('timeline')} 
                        className={`px-4 py-1.5 text-xs font-medium rounded transition-all duration-200 ${viewMode === 'timeline' ? 'bg-primary text-white shadow-md' : 'text-text-muted hover:text-white'}`}
                    >
                        Timeline
                    </button>
                </div>
                
                <div className="ml-auto flex gap-3">
                    <div className="relative group">
                        <div className="absolute inset-y-0 left-3 flex items-center pointer-events-none">
                            <span className="material-symbols-outlined text-[14px] text-text-muted group-hover:text-primary transition-colors">calendar_month</span>
                        </div>
                        <input 
                            type="date" 
                            className="bg-surface-dark border border-border-color text-white text-xs rounded-lg pl-9 pr-3 py-2 focus:ring-1 focus:ring-primary/50 focus:border-primary [color-scheme:dark] shadow-sm outline-none transition-all hover:border-primary/30" 
                        />
                    </div>
                    <button className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white hover:bg-border-color rounded-lg text-xs font-medium transition-all flex items-center gap-2">
                        <span className="material-symbols-outlined text-sm">filter_list</span>
                        Filter
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden shadow-xl shadow-black/30 max-w-[1600px] mx-auto">
                    {loading ? (
                        <div className="p-12 text-center">
                            <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary mx-auto mb-4"></div>
                            <p className="text-text-muted text-sm font-medium">Fetching trip records...</p>
                        </div>
                    ) : (
                        <table className="w-full text-left text-sm text-text-muted border-collapse">
                            <thead className="bg-[#1d2936] text-xs uppercase font-semibold tracking-wider">
                                <tr>
                                    <th className="px-6 py-4 text-white">Date / Time</th>
                                    <th className="px-6 py-4">Vehicle</th>
                                    <th className="px-6 py-4">Driver</th>
                                    <th className="px-6 py-4">Duration</th>
                                    <th className="px-6 py-4">Distance</th>
                                    <th className="px-6 py-4 text-right">Route</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-border-color/50">
                                {trips.length === 0 ? (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-12 text-center">
                                            <div className="flex flex-col items-center gap-3 opacity-50">
                                                <span className="material-symbols-outlined text-4xl">history</span>
                                                <p className="text-sm">No historical trip logs found.</p>
                                            </div>
                                        </td>
                                    </tr>
                                ) : (
                                    trips.map((trip: Trip) => (
                                        <tr key={trip.id} className="hover:bg-primary/5 transition-colors group cursor-pointer border-l-2 border-transparent hover:border-primary">
                                            <td className="px-6 py-4">
                                                <div className="text-white font-medium">{new Date(trip.start_time).toLocaleDateString([], { month: 'short', day: 'numeric' })}</div>
                                                <div className="text-[10px] text-text-muted font-mono">{new Date(trip.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <span className="material-symbols-outlined text-sm text-primary">directions_car</span>
                                                    <span className="text-white font-medium">#{trip.vehicle_id.slice(0, 4).toUpperCase()}</span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <div className="w-6 h-6 rounded-full bg-border-color flex items-center justify-center text-[10px] font-bold text-text-muted">
                                                        {trip.driver_id.slice(0, 1).toUpperCase()}
                                                    </div>
                                                    <span>{trip.driver_id.slice(0, 8)}</span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">4h 20m</td> {/* Mocked as db doesn't have duration yet */}
                                            <td className="px-6 py-4 font-mono">{trip.distance_km ? `${trip.distance_km.toFixed(1)} km` : '-'}</td>
                                            <td className="px-6 py-4 text-right">
                                                <button className="text-primary text-xs font-semibold hover:text-white transition-colors flex items-center gap-1 justify-end ml-auto group-hover:underline">
                                                    Replay Route
                                                    <span className="material-symbols-outlined text-[14px]">play_arrow</span>
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
                    <span>© 2026 FlyFleet Enterprise. All rights reserved.</span>
                    <span className="text-[10px]">v2.4.1 | Server Time: <span className="font-mono">{new Date().getUTCHours()}:{new Date().getUTCMinutes().toString().padStart(2, '0')} UTC</span></span>
                </footer>
            </div>
        </div>
    );
};

export default History;
