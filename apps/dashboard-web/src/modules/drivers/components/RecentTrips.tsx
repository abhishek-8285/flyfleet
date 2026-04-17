import React, { useEffect, useState, useCallback } from "react";
import { api } from "../../../services/api";

interface RecentTripsProps {
  driverId: string;
}

export const RecentTrips: React.FC<RecentTripsProps> = ({ driverId }) => {
  const [trips, setTrips] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const loadTrips = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getTrips({ driver_id: driverId, limit: 10 });
      setTrips(res.data || res.trips || []);
    } catch (err) {
      console.error("Failed to load recent trips:", err);
    } finally {
      setLoading(false);
    }
  }, [driverId]);

  useEffect(() => {
    loadTrips();
  }, [loadTrips]);

  const getStatusStyle = (status: string) => {
    const s = status.toLowerCase();
    if (s === "completed") return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
    if (s === "in_progress" || s === "active") return "bg-blue-500/10 text-blue-400 border-blue-500/20";
    return "bg-amber-500/10 text-amber-400 border-amber-500/20";
  };

  if (loading) return <div className="p-8 text-center text-on-surface-variant font-mono text-[10px] animate-pulse">RECONSTRUCTING TRIP LOGS...</div>;

  return (
    <section className="bg-surface-container rounded-xl border border-outline overflow-hidden shadow-sm">
      <div className="p-6 border-b border-outline flex justify-between items-center bg-surface-container-high/30">
        <h3 className="font-bold text-white flex items-center gap-2">
          <span className="material-symbols-outlined text-primary">history</span>
          Industrial Trip Log
        </h3>
        <div className="flex gap-2">
          <button className="p-1 hover:bg-surface-container-highest rounded transition-colors text-on-surface-variant hover:text-white">
            <span className="material-symbols-outlined text-sm">filter_list</span>
          </button>
          <button className="p-1 hover:bg-surface-container-highest rounded transition-colors text-on-surface-variant hover:text-white">
            <span className="material-symbols-outlined text-sm">download</span>
          </button>
        </div>
      </div>
      
      <div className="overflow-x-auto custom-scrollbar">
        <table className="w-full text-left">
          <thead className="bg-surface-container-high text-[10px] font-bold text-on-surface-variant uppercase tracking-widest border-b border-outline">
            <tr>
              <th className="px-6 py-4">Node / ID</th>
              <th className="px-6 py-4">Terminal Route</th>
              <th className="px-6 py-4">Duration</th>
              <th className="px-6 py-4">Pulse Status</th>
              <th className="px-6 py-4 text-right">Registry</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-outline">
            {trips.length > 0 ? trips.map((trip) => (
              <tr key={trip.id} className="hover:bg-surface-container-high/50 transition-colors group cursor-default">
                <td className="px-6 py-4">
                  <p className="text-sm text-white font-medium">{new Date(trip.start_time).toLocaleDateString()}</p>
                  <p className="text-[10px] text-on-surface-variant font-mono uppercase tracking-tighter">ID: TRP-{trip.id.slice(0, 4).toUpperCase()}</p>
                </td>
                <td className="px-6 py-4">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-white font-medium">{trip.origin || "Terminal Alpha"}</span>
                    <span className="material-symbols-outlined text-xs text-on-surface-variant group-hover:text-primary transition-all">trending_flat</span>
                    <span className="text-sm text-white font-medium">{trip.destination || "Target Bravo"}</span>
                  </div>
                </td>
                <td className="px-6 py-4 text-sm text-on-surface-variant font-medium italic">
                  {trip.duration_minutes ? `${Math.floor(trip.duration_minutes / 60)}h ${trip.duration_minutes % 60}m` : "LIVE TRACK"}
                </td>
                <td className="px-6 py-4">
                  <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold border uppercase tracking-widest ${getStatusStyle(trip.status)}`}>
                    {trip.status.replace(/_/g, ' ')}
                  </span>
                </td>
                <td className="px-6 py-4 text-right">
                  <button className="material-symbols-outlined text-on-surface-variant hover:text-white transition-colors opacity-50 group-hover:opacity-100 italic">more_vert</button>
                </td>
              </tr>
            )) : (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center text-on-surface-variant italic text-xs opacity-50">
                  <span className="material-symbols-outlined text-2xl mb-2">inbox</span>
                  <p>No trip records found for this operational node.</p>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      
      <div className="p-4 text-center border-t border-outline bg-surface-container/50">
        <button className="text-[10px] font-black tracking-[0.2em] text-on-surface-variant hover:text-white transition-all uppercase italic">
          DEEP DIVE TRIP ARCHIVE
        </button>
      </div>
    </section>
  );
};
