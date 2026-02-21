import React, { useCallback, useEffect, useState } from 'react';
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
  sent_at?: string;
  executed_at?: string;
}

interface Device {
  id: string;
  serial_number: string;
  status?: string;
}

const COMMAND_PRESETS = [
  { command: 'reboot', category: 'power', label: 'Reboot Device', icon: 'REBOOT' },
  { command: 'reset_modem', category: 'network', label: 'Reset Modem', icon: 'MODEM' },
  { command: 'clear_cache', category: 'config', label: 'Clear Cache', icon: 'CACHE' },
  { command: 'update_config', category: 'config', label: 'Update Config', icon: 'CFG' },
  { command: 'request_position', category: 'diagnostic', label: 'Request Position', icon: 'GPS' },
  { command: 'run_diagnostics', category: 'diagnostic', label: 'Run Diagnostics', icon: 'DIAG' },
];

const RemoteCommands: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [devices, setDevices] = useState<Device[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<string>(searchParams.get('device') || '');
  const [commands, setCommands] = useState<DeviceCommand[]>([]);
  const [loading, setLoading] = useState(false);
  const [sendingCommand, setSendingCommand] = useState<string | null>(null);
  const [transitioningCommand, setTransitioningCommand] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const getAuthHeaders = (): HeadersInit => {
    const token = localStorage.getItem('access_token');
    const orgId = localStorage.getItem('org_id');
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    if (orgId) {
      headers['X-Organization-ID'] = orgId;
    }
    return headers;
  };

  const parseError = async (res: Response): Promise<string> => {
    try {
      const text = await res.text();
      if (text) return text;
    } catch (err) {
      // no-op
    }
    return `Request failed (${res.status})`;
  };

  const fetchDevices = useCallback(async () => {
    try {
      const res = await fetch('/api/v1/devices?limit=100', {
        headers: getAuthHeaders(),
      });
      if (!res.ok) {
        throw new Error(await parseError(res));
      }
      const list = (await res.json()) as Device[];
      setDevices(Array.isArray(list) ? list : []);
    } catch (err: any) {
      setError(err?.message || 'Failed to fetch devices');
    }
  }, []);

  const fetchCommandHistory = useCallback(async () => {
    if (!selectedDevice) return;
    setLoading(true);
    try {
      const res = await fetch(`/api/v1/devices/${selectedDevice}/commands?limit=50`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) {
        throw new Error(await parseError(res));
      }
      const list = (await res.json()) as DeviceCommand[];
      setCommands(Array.isArray(list) ? list : []);
    } catch (err: any) {
      setError(err?.message || 'Failed to fetch command history');
    } finally {
      setLoading(false);
    }
  }, [selectedDevice]);

  useEffect(() => {
    void fetchDevices();
  }, [fetchDevices]);

  useEffect(() => {
    if (!selectedDevice) {
      setCommands([]);
      return;
    }

    void fetchCommandHistory();
    const poll = window.setInterval(() => {
      void fetchCommandHistory();
    }, 5000);

    return () => {
      window.clearInterval(poll);
    };
  }, [selectedDevice, fetchCommandHistory]);

  const sendCommand = async (preset: typeof COMMAND_PRESETS[0]) => {
    if (!selectedDevice) return;
    setError(null);
    setSuccess(null);
    setSendingCommand(preset.command);

    try {
      const res = await fetch(`/api/v1/devices/${selectedDevice}/commands`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
          command: preset.command,
          category: preset.category,
          params: {},
        }),
      });

      if (!res.ok) {
        throw new Error(await parseError(res));
      }

      setSuccess(`Command queued: ${preset.label}`);
      await fetchCommandHistory();
    } catch (err: any) {
      setError(err?.message || `Failed to send command: ${preset.label}`);
    } finally {
      setSendingCommand(null);
    }
  };

  const normalizeStatus = (raw: string): string => {
    const status = raw.toLowerCase();
    if (status === 'pending') return 'queued';
    if (status === 'sent') return 'delivered';
    if (status === 'acknowledged') return 'acked';
    if (status === 'executed' || status === 'completed') return 'applied';
    return status;
  };

  const getAllowedTransitions = (status: string): string[] => {
    switch (normalizeStatus(status)) {
      case 'queued':
        return ['delivered', 'failed'];
      case 'delivered':
        return ['acked', 'timed_out'];
      case 'acked':
        return ['applied', 'failed'];
      default:
        return [];
    }
  };

  const transitionCommand = async (commandId: string, nextStatus: string) => {
    if (!selectedDevice) return;
    setError(null);
    setSuccess(null);
    setTransitioningCommand(commandId);
    try {
      const res = await fetch(`/api/v1/devices/${selectedDevice}/commands/${commandId}/status`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
        body: JSON.stringify({
          status: nextStatus,
          result: { updated_from: 'remote-commands-ui' },
        }),
      });
      if (!res.ok) {
        throw new Error(await parseError(res));
      }
      setSuccess(`Command transitioned to ${nextStatus}`);
      await fetchCommandHistory();
    } catch (err: any) {
      setError(err?.message || `Failed to transition command to ${nextStatus}`);
    } finally {
      setTransitioningCommand(null);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'executed':
      case 'completed':
        return 'bg-emerald-500/20 text-emerald-400';
      case 'pending':
        return 'bg-yellow-500/20 text-yellow-400';
      case 'sent':
      case 'delivered':
      case 'acknowledged':
        return 'bg-blue-500/20 text-blue-400';
      case 'failed':
      case 'cancelled':
        return 'bg-red-500/20 text-red-400';
      default:
        return 'bg-slate-500/20 text-slate-400';
    }
  };

  const getLifecycleStage = (cmd: DeviceCommand): string => {
    if (cmd.executed_at) {
      return cmd.status.toLowerCase() === 'failed' ? 'Executed (failed)' : 'Executed';
    }
    if (cmd.sent_at) {
      return 'Dispatched';
    }
    return 'Queued';
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
      <div className="flex items-center justify-between mb-8 gap-4">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/admin/devices')}
            className="p-2 rounded-lg bg-slate-800/50 text-slate-400 hover:text-white transition-colors"
          >
            ← Back
          </button>
          <div>
            <h1 className="text-2xl font-bold text-white">Remote Commands</h1>
            <p className="text-slate-400">Dispatch commands and track ACK/execute lifecycle</p>
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

      {error && (
        <div className="mb-4 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300">
          {success}
        </div>
      )}

      {!selectedDevice ? (
        <div className="flex items-center justify-center h-64 bg-slate-800/50 rounded-xl border border-slate-700/50">
          <p className="text-slate-400">Select a device to send commands</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Quick Commands</h2>
            <div className="grid grid-cols-2 gap-4">
              {COMMAND_PRESETS.map((preset) => (
                <button
                  key={preset.command}
                  onClick={() => void sendCommand(preset)}
                  disabled={sendingCommand !== null}
                  className="flex items-center gap-3 p-4 bg-slate-900/50 rounded-xl border border-slate-700/50 hover:border-violet-500/50 hover:bg-slate-800/50 transition-all disabled:opacity-50"
                >
                  <span className="text-[10px] text-violet-300 font-semibold px-2 py-1 rounded bg-violet-500/20">{preset.icon}</span>
                  <div className="text-left">
                    <span className="text-white font-medium block">{preset.label}</span>
                    <span className="text-slate-400 text-xs">{preset.category}</span>
                  </div>
                </button>
              ))}
            </div>
          </div>

          <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Command History</h2>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {loading ? (
                <p className="text-slate-400 text-sm animate-pulse">Loading history...</p>
              ) : commands.length === 0 ? (
                <p className="text-slate-400 text-sm">No commands sent yet</p>
              ) : (
                commands.map((cmd) => (
                  <div key={cmd.id} className="p-3 bg-slate-900/50 rounded-lg border border-slate-800">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-white font-medium">{cmd.command}</span>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getStatusColor(cmd.status)}`}>
                        {cmd.status.toUpperCase()}
                      </span>
                    </div>
                    <div className="text-xs text-slate-300 mb-1">Stage: {getLifecycleStage(cmd)}</div>
                    <div className="text-xs text-slate-400">
                      Queued: {new Date(cmd.created_at).toLocaleString()}
                      {cmd.sent_at && <span> • Sent: {new Date(cmd.sent_at).toLocaleString()}</span>}
                      {cmd.executed_at && <span> • Executed: {new Date(cmd.executed_at).toLocaleString()}</span>}
                    </div>
                    {cmd.error_message && (
                      <p className="mt-2 text-xs text-red-400">{cmd.error_message}</p>
                    )}
                    {cmd.result && Object.keys(cmd.result).length > 0 && (
                      <pre className="mt-2 text-[11px] text-slate-300 bg-slate-950/70 rounded p-2 overflow-x-auto">{JSON.stringify(cmd.result, null, 2)}</pre>
                    )}
                    {getAllowedTransitions(cmd.status).length > 0 && (
                      <div className="mt-3 flex flex-wrap gap-2">
                        {getAllowedTransitions(cmd.status).map((next) => (
                          <button
                            key={`${cmd.id}-${next}`}
                            onClick={() => void transitionCommand(cmd.id, next)}
                            disabled={transitioningCommand === cmd.id}
                            className="px-2 py-1 rounded text-[11px] font-medium bg-slate-800 text-slate-200 hover:bg-slate-700 disabled:opacity-60"
                          >
                            {transitioningCommand === cmd.id ? 'Updating...' : `Set ${next}`}
                          </button>
                        ))}
                      </div>
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
