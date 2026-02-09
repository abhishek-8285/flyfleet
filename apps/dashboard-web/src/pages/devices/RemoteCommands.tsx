import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

interface DeviceCommand {
  id: string;
  device_id: string;
  command: string;
  category: string;
  status: string;
  result?: Record<string, unknown>;
  error_message?: string;
  created_at: string;
  executed_at?: string;
}

interface Device {
  id: string;
  serial_number: string;
}

const COMMAND_PRESETS = [
  { command: 'reboot', category: 'power', label: 'Reboot Device', icon: '🔄' },
  { command: 'reset_modem', category: 'network', label: 'Reset Modem', icon: '📶' },
  { command: 'clear_cache', category: 'config', label: 'Clear Cache', icon: '🗑️' },
  { command: 'update_config', category: 'config', label: 'Update Config', icon: '⚙️' },
  { command: 'request_position', category: 'diagnostic', label: 'Request Position', icon: '📍' },
  { command: 'run_diagnostics', category: 'diagnostic', label: 'Run Diagnostics', icon: '🔍' },
];

const RemoteCommands: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [devices, setDevices] = useState<Device[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<string>(searchParams.get('device') || '');
  const [commands, setCommands] = useState<DeviceCommand[]>([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);

  useEffect(() => {
    fetchDevices();
  }, []);

  useEffect(() => {
    if (selectedDevice) {
      fetchCommandHistory();
    }
  }, [selectedDevice]);

  const fetchDevices = async () => {
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch('/api/devices/status?limit=100', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        setDevices(await res.json() || []);
      }
    } catch (error) {
      console.error('Failed to fetch devices:', error);
    }
  };

  const fetchCommandHistory = async () => {
    if (!selectedDevice) return;
    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch(`/api/devices/${selectedDevice}/commands?limit=50`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        setCommands(await res.json() || []);
      }
    } catch (error) {
      console.error('Failed to fetch commands:', error);
    } finally {
      setLoading(false);
    }
  };

  const sendCommand = async (preset: typeof COMMAND_PRESETS[0]) => {
    if (!selectedDevice) return;
    setSending(true);
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch(`/api/devices/${selectedDevice}/commands`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          command: preset.command,
          category: preset.category,
          params: {}
        })
      });
      if (res.ok) {
        fetchCommandHistory();
      }
    } catch (error) {
      console.error('Failed to send command:', error);
    } finally {
      setSending(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'bg-emerald-500/20 text-emerald-400';
      case 'pending': return 'bg-yellow-500/20 text-yellow-400';
      case 'sent': return 'bg-blue-500/20 text-blue-400';
      case 'failed': return 'bg-red-500/20 text-red-400';
      default: return 'bg-slate-500/20 text-slate-400';
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/admin/devices')}
            className="p-2 rounded-lg bg-slate-800/50 text-slate-400 hover:text-white transition-colors"
          >
            ← Back
          </button>
          <div>
            <h1 className="text-2xl font-bold text-white">Remote Commands</h1>
            <p className="text-slate-400">Send commands to devices</p>
          </div>
        </div>

        <select
          value={selectedDevice}
          onChange={(e) => setSelectedDevice(e.target.value)}
          className="bg-slate-800/50 border border-slate-700/50 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
        >
          <option value="">Select Device</option>
          {devices.map((d) => (
            <option key={d.id} value={d.id}>{d.serial_number}</option>
          ))}
        </select>
      </div>

      {!selectedDevice ? (
        <div className="flex items-center justify-center h-64 bg-slate-800/50 rounded-xl border border-slate-700/50">
          <p className="text-slate-400">Select a device to send commands</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
              <span>⚡</span> Quick Commands
            </h2>
            <div className="grid grid-cols-2 gap-4">
              {COMMAND_PRESETS.map((preset) => (
                <button
                  key={preset.command}
                  onClick={() => sendCommand(preset)}
                  disabled={sending}
                  className="flex items-center gap-3 p-4 bg-slate-900/50 rounded-xl border border-slate-700/50 hover:border-violet-500/50 hover:bg-slate-800/50 transition-all disabled:opacity-50"
                >
                  <span className="text-2xl">{preset.icon}</span>
                  <div className="text-left">
                    <span className="text-white font-medium block">{preset.label}</span>
                    <span className="text-slate-400 text-xs">{preset.category}</span>
                  </div>
                </button>
              ))}
            </div>
          </div>

          <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
              <span>📋</span> Command History
            </h2>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {loading ? (
                <p className="text-slate-400 text-sm animate-pulse">Loading history...</p>
              ) : commands.length === 0 ? (
                <p className="text-slate-400 text-sm">No commands sent yet</p>
              ) : (
                commands.map((cmd) => (
                  <div key={cmd.id} className="p-3 bg-slate-900/50 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-white font-medium">{cmd.command}</span>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getStatusColor(cmd.status)}`}>
                        {cmd.status.toUpperCase()}
                      </span>
                    </div>
                    <div className="text-xs text-slate-400">
                      Sent: {new Date(cmd.created_at).toLocaleString()}
                      {cmd.executed_at && (
                        <span> • Executed: {new Date(cmd.executed_at).toLocaleString()}</span>
                      )}
                    </div>
                    {cmd.error_message && (
                      <p className="mt-2 text-xs text-red-400">{cmd.error_message}</p>
                    )}
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default RemoteCommands;
