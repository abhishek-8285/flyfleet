import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface FirmwarePackage {
  id: string;
  version: string;
  hardware_target: string;
  file_url: string;
  file_size?: number;
  checksum?: string;
  release_notes?: string;
  status: string;
  is_critical: boolean;
  released_at?: string;
  created_at: string;
}

interface FirmwareRollout {
  id: string;
  firmware_id: string;
  name?: string;
  device_group?: string;
  total_devices: number;
  completed_devices: number;
  failed_devices: number;
  status: string;
  started_at?: string;
  created_at: string;
}

const FirmwareManager: React.FC = () => {
  const navigate = useNavigate();
  const [firmware, setFirmware] = useState<FirmwarePackage[]>([]);
  const [rollouts, setRollouts] = useState<FirmwareRollout[]>([]);
  const [activeTab, setActiveTab] = useState<'packages' | 'rollouts'>('packages');
  const [loading, setLoading] = useState(true);
  const [showUploadModal, setShowUploadModal] = useState(false);

  useEffect(() => {
    fetchData();
  }, [activeTab]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      const headers = { 'Authorization': `Bearer ${token}` };
      if (activeTab === 'packages') {
        const res = await fetch('/api/firmware', { headers });
        if (res.ok) setFirmware(await res.json() || []);
      } else {
        const res = await fetch('/api/firmware/rollouts', { headers });
        if (res.ok) setRollouts(await res.json() || []);
      }
    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-emerald-500/20 text-emerald-400';
      case 'testing': return 'bg-yellow-500/20 text-yellow-400';
      case 'draft': return 'bg-slate-500/20 text-slate-400';
      case 'deprecated': return 'bg-red-500/20 text-red-400';
      case 'in_progress': return 'bg-blue-500/20 text-blue-400';
      case 'completed': return 'bg-emerald-500/20 text-emerald-400';
      case 'scheduled': return 'bg-violet-500/20 text-violet-400';
      default: return 'bg-slate-500/20 text-slate-400';
    }
  };

  const formatFileSize = (bytes?: number) => {
    if (!bytes) return '-';
    const mb = bytes / (1024 * 1024);
    return `${mb.toFixed(2)} MB`;
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/admin/devices')}
            className="p-2 rounded-lg bg-slate-800/50 text-slate-400 hover:text-white transition-colors"
          >
            ← Back
          </button>
          <div>
            <h1 className="text-2xl font-bold text-white">Firmware Manager</h1>
            <p className="text-slate-400">Manage firmware packages and rollouts</p>
          </div>
        </div>

        <button
          onClick={() => setShowUploadModal(true)}
          className="px-4 py-2 bg-violet-600 text-white rounded-lg hover:bg-violet-500 transition-colors flex items-center gap-2"
        >
          <span>+</span> Upload Firmware
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('packages')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'packages'
              ? 'bg-violet-600 text-white'
              : 'bg-slate-800/50 text-slate-400 hover:bg-slate-700/50'
          }`}
        >
          📦 Packages
        </button>
        <button
          onClick={() => setActiveTab('rollouts')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'rollouts'
              ? 'bg-violet-600 text-white'
              : 'bg-slate-800/50 text-slate-400 hover:bg-slate-700/50'
          }`}
        >
          🚀 Rollouts
        </button>
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex items-center justify-center h-64 bg-slate-800/50 rounded-xl border border-slate-700/50">
          <p className="text-slate-400 animate-pulse">Loading...</p>
        </div>
      ) : activeTab === 'packages' ? (
        <div className="grid gap-4">
          {firmware.length === 0 ? (
            <div className="text-center py-12 bg-slate-800/50 rounded-xl border border-slate-700/50 text-slate-400">
              No firmware packages available
            </div>
          ) : (
            firmware.map((fw) => (
              <div key={fw.id} className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="text-xl font-bold text-white">v{fw.version}</h3>
                      <span className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(fw.status)}`}>
                        {fw.status.toUpperCase()}
                      </span>
                      {fw.is_critical && (
                        <span className="px-3 py-1 rounded-full text-xs font-medium bg-red-500/20 text-red-400">
                          CRITICAL
                        </span>
                      )}
                    </div>
                    <p className="text-slate-400 text-sm mb-4">Target: {fw.hardware_target}</p>
                    {fw.release_notes && (
                      <p className="text-slate-300 text-sm mb-4">{fw.release_notes}</p>
                    )}
                    <div className="flex gap-6 text-sm">
                      <span className="text-slate-400">Size: <span className="text-white">{formatFileSize(fw.file_size)}</span></span>
                      {fw.checksum && <span className="text-slate-400">Checksum: <span className="text-white font-mono">{fw.checksum.slice(0, 12)}...</span></span>}
                      <span className="text-slate-400">Created: <span className="text-white">{new Date(fw.created_at).toLocaleDateString()}</span></span>
                    </div>
                  </div>
                  <button className="px-4 py-2 bg-violet-500/20 text-violet-400 rounded-lg hover:bg-violet-500/30 transition-colors text-sm">
                    Start Rollout
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      ) : (
        <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-700/50">
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Rollout</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Status</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Progress</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Target</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Started</th>
              </tr>
            </thead>
            <tbody>
              {rollouts.length === 0 ? (
                <tr>
                  <td colSpan={5} className="text-center py-12 text-slate-400">
                    No active rollouts
                  </td>
                </tr>
              ) : (
                rollouts.map((r) => (
                  <tr key={r.id} className="border-b border-slate-700/30 hover:bg-slate-700/30">
                    <td className="py-4 px-6 text-white font-medium">{r.name || r.id.slice(0, 8)}</td>
                    <td className="py-4 px-6">
                      <span className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(r.status)}`}>
                        {r.status.replace('_', ' ').toUpperCase()}
                      </span>
                    </td>
                    <td className="py-4 px-6">
                      <div className="flex items-center gap-3">
                        <div className="w-32 h-2 bg-slate-700 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-violet-500"
                            style={{ width: `${r.total_devices ? (r.completed_devices / r.total_devices) * 100 : 0}%` }}
                          ></div>
                        </div>
                        <span className="text-slate-300 text-sm">
                          {r.completed_devices}/{r.total_devices}
                        </span>
                        {r.failed_devices > 0 && (
                          <span className="text-red-400 text-sm">({r.failed_devices} failed)</span>
                        )}
                      </div>
                    </td>
                    <td className="py-4 px-6 text-slate-300">{r.device_group || 'All devices'}</td>
                    <td className="py-4 px-6 text-slate-400 text-sm">
                      {r.started_at ? new Date(r.started_at).toLocaleString() : '-'}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-xl border border-slate-700 p-6 w-full max-w-md">
            <h2 className="text-xl font-bold text-white mb-4">Upload Firmware</h2>
            <form className="space-y-4">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Version</label>
                <input type="text" placeholder="1.0.0" className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white" />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Hardware Target</label>
                <input type="text" placeholder="HW-REV-04" className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white" />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">File URL</label>
                <input type="url" placeholder="https://..." className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white" />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Release Notes</label>
                <textarea rows={3} className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"></textarea>
              </div>
              <div className="flex gap-3 pt-4">
                <button type="button" onClick={() => setShowUploadModal(false)} className="flex-1 py-2 bg-slate-700 text-white rounded-lg">Cancel</button>
                <button type="submit" className="flex-1 py-2 bg-violet-600 text-white rounded-lg">Upload</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default FirmwareManager;
