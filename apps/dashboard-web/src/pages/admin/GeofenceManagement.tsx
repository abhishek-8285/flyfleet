import React, { useState, useEffect } from 'react';
import { 
  Plus, 
  Search, 
  Map as MapIcon, 
  Circle, 
  Pentagon, 
  MoreVertical, 
  Trash2, 
  Edit,
  Shield,
  Activity,
  Navigation
} from 'lucide-react';
import { api } from '../../services/api';
import { Geofence } from '../../modules/dashboard/components/LiveOperationsMap';
import { LoadingSpinner } from '../../components/common/LoadingSpinner';

const GeofenceManagement: React.FC = () => {
  const [geofences, setGeofences] = useState<Geofence[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchGeofences = async () => {
    try {
      setLoading(true);
      const res = await api.getGeofences();
      const data = res?.data || res;
      setGeofences(Array.isArray(data) ? data : data?.geofences || []);
    } catch (err) {
      console.error('Failed to fetch geofences:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchGeofences();
  }, []);

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this geofence?')) return;
    try {
      await api.deleteGeofence(id);
      fetchGeofences();
    } catch (err) {
      console.error('Failed to delete geofence:', err);
    }
  };

  const filteredGeofences = geofences.filter(g => 
    g.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="p-6 space-y-6 bg-[#f8fafc] dark:bg-black min-h-screen">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-gray-900 dark:text-white uppercase">
            Geofence Management
          </h1>
          <p className="text-sm text-gray-500 font-medium">
            Monitor and configure spatial boundaries for automated alerts.
          </p>
        </div>
        <button 
          onClick={() => window.location.href = '/admin/geofences/new'}
          className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2.5 rounded-xl font-bold transition-all shadow-lg shadow-emerald-600/20 text-sm uppercase tracking-tighter"
        >
          <Plus size={18} />
          Create New Geofence
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { label: 'Total Zones', value: geofences.length, icon: MapIcon, color: 'text-blue-500' },
          { label: 'Active Alerts', value: geofences.filter(g => g.status === 'active').length, icon: Activity, color: 'text-emerald-500' },
          { label: 'Detection Types', value: 'Circle / Polygon', icon: Shield, color: 'text-purple-500' }
        ].map((stat, i) => (
          <div key={i} className="bg-white dark:bg-white/[0.03] p-4 rounded-2xl border border-gray-100 dark:border-white/5 shadow-sm">
            <div className="flex items-center justify-between mb-1">
              <p className="text-[10px] font-black uppercase tracking-widest text-gray-400">{stat.label}</p>
              <stat.icon size={16} className={stat.color} />
            </div>
            <p className="text-2xl font-black text-gray-900 dark:text-white">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Toolbar */}
      <div className="flex flex-col md:flex-row gap-4 items-center bg-white dark:bg-white/[0.02] p-2 rounded-2xl border border-gray-100 dark:border-white/5">
        <div className="relative flex-1 w-full">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
          <input
            type="text"
            placeholder="SEARCH GEOFENCES..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-transparent pl-10 pr-4 py-2 text-sm font-bold uppercase tracking-tighter text-gray-700 dark:text-white focus:outline-none"
          />
        </div>
      </div>

      {/* List */}
      {loading ? (
        <div className="flex justify-center py-20">
          <LoadingSpinner />
        </div>
      ) : filteredGeofences.length > 0 ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {filteredGeofences.map((gf) => (
            <div key={gf.id} className="bg-white dark:bg-white/[0.03] border border-gray-100 dark:border-white/5 rounded-2xl overflow-hidden hover:shadow-xl transition-all group">
              <div className="p-5 flex items-start justify-between">
                <div className="flex gap-4">
                  <div className={`w-12 h-12 rounded-2xl flex items-center justify-center ${gf.type === 'circle' ? 'bg-blue-50 text-blue-600' : 'bg-purple-50 text-purple-600'} dark:bg-white/5`}>
                    {gf.type === 'circle' ? <Circle size={24} /> : <Pentagon size={24} />}
                  </div>
                  <div>
                    <h3 className="font-black text-gray-900 dark:text-white uppercase tracking-tighter text-lg leading-none mb-1">
                      {gf.name}
                    </h3>
                    <div className="flex items-center gap-3">
                      <span className={`text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded ${gf.status === 'active' ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-600'}`}>
                        {gf.status}
                      </span>
                      <span className="text-[10px] font-black uppercase tracking-widest text-gray-400 flex items-center gap-1">
                        <Navigation size={10} />
                        {gf.type === 'circle' ? `${gf.radius_m}m Radius` : 'Polygonal Boundary'}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button className="p-2 hover:bg-gray-100 dark:hover:bg-white/5 rounded-lg text-gray-400 hover:text-blue-500 transition-colors">
                    <Edit size={18} />
                  </button>
                  <button 
                    onClick={() => handleDelete(gf.id)}
                    className="p-2 hover:bg-gray-100 dark:hover:bg-white/5 rounded-lg text-gray-400 hover:text-red-500 transition-colors"
                  >
                    <Trash2 size={18} />
                  </button>
                </div>
              </div>
              
              {/* Mini Preview Placeholder / Info */}
              <div className="px-5 py-3 bg-gray-50 dark:bg-white/[0.01] border-t border-gray-100 dark:border-white/5 flex justify-between items-center text-[11px] font-bold uppercase tracking-tighter text-gray-400">
                <div className="flex gap-4">
                  <span>ID: {gf.id.substring(0, 8)}</span>
                  {gf.center_lat && (
                    <span>{gf.center_lat.toFixed(4)}, {gf.center_lng?.toFixed(4)}</span>
                  )}
                </div>
                <div className="flex items-center gap-1 text-emerald-500">
                  <Activity size={12} />
                  Live Monitoring
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="bg-white dark:bg-white/[0.02] border-2 border-dashed border-gray-200 dark:border-white/5 rounded-3xl py-20 flex flex-col items-center justify-center text-center">
          <div className="w-16 h-16 bg-gray-100 dark:bg-white/5 rounded-2xl flex items-center justify-center text-gray-400 mb-4">
            <MapIcon size={32} />
          </div>
          <h3 className="text-xl font-black text-gray-900 dark:text-white uppercase tracking-tighter mb-1">No Geofences Found</h3>
          <p className="text-gray-400 max-w-xs mb-6">Start by creating a new spatial boundary to monitor asset entry and exit events.</p>
          <button 
            onClick={() => window.location.href = '/admin/geofences/new'}
            className="bg-emerald-600 hover:bg-emerald-500 text-white px-6 py-3 rounded-xl font-bold transition-all uppercase tracking-tighter"
          >
            Create Your First Geofence
          </button>
        </div>
      )}
    </div>
  );
};

export default GeofenceManagement;
