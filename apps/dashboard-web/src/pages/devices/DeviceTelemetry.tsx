import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

interface TelemetryPacket {
  id: string;
  device_id: string;
  packet_type: string;
  payload: Record<string, unknown>;
  raw_hex?: string;
  timestamp: string;
}

interface Device {
  id: string;
  serial_number: string;
}

const DeviceTelemetry: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [telemetry, setTelemetry] = useState<TelemetryPacket[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<string>(searchParams.get('device') || '');
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(false);

  useEffect(() => {
    fetchDevices();
  }, []);

  useEffect(() => {
    if (selectedDevice) {
      fetchTelemetry();
    }
  }, [selectedDevice]);

  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (autoRefresh && selectedDevice) {
      interval = setInterval(fetchTelemetry, 5000);
    }
    return () => clearInterval(interval);
  }, [autoRefresh, selectedDevice]);

  const fetchDevices = async () => {
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch('/api/v1/devices/status?limit=100', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        setDevices(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch devices:', error);
    }
  };

  const fetchTelemetry = async () => {
    if (!selectedDevice) return;
    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch(`/api/v1/devices/${selectedDevice}/telemetry?limit=100`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        setTelemetry(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch telemetry:', error);
    } finally {
      setLoading(false);
    }
  };

  const getPacketTypeColor = (type: string) => {
    switch (type) {
      case 'gps': return 'bg-emerald-500/20 text-emerald-400';
      case 'obd': return 'bg-blue-500/20 text-blue-400';
      case 'heartbeat': return 'bg-violet-500/20 text-violet-400';
      case 'sensor': return 'bg-yellow-500/20 text-yellow-400';
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
            <h1 className="text-2xl font-bold text-white">Live Telemetry</h1>
            <p className="text-slate-400">Real-time device data streams</p>
          </div>
        </div>

        <div className="flex items-center gap-4">
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

          <label className="flex items-center gap-2 text-slate-300">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="w-4 h-4 rounded accent-violet-500"
            />
            Auto-refresh
          </label>

          <button
            onClick={fetchTelemetry}
            className="px-4 py-2 bg-violet-600 text-white rounded-lg hover:bg-violet-500 transition-colors"
          >
            Refresh
          </button>
        </div>
      </div>

      {!selectedDevice ? (
        <div className="flex items-center justify-center h-64 bg-slate-800/50 rounded-xl border border-slate-700/50">
          <p className="text-slate-400">Select a device to view telemetry</p>
        </div>
      ) : loading ? (
        <div className="flex items-center justify-center h-64 bg-slate-800/50 rounded-xl border border-slate-700/50">
          <p className="text-slate-400 animate-pulse">Loading telemetry...</p>
        </div>
      ) : (
        <div className="space-y-4">
          {telemetry.length === 0 ? (
            <div className="text-center py-12 text-slate-400">
              No telemetry data available for this device
            </div>
          ) : (
            telemetry.map((packet) => (
              <div key={packet.id} className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-4">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${getPacketTypeColor(packet.packet_type)}`}>
                      {packet.packet_type?.toUpperCase() || 'UNKNOWN'}
                    </span>
                    <span className="text-slate-400 text-sm">
                      {new Date(packet.timestamp).toLocaleString()}
                    </span>
                  </div>
                </div>

                <div className="bg-slate-900/50 rounded-lg p-4 font-mono text-sm">
                  <pre className="text-emerald-400 whitespace-pre-wrap overflow-x-auto">
                    {JSON.stringify(packet.payload, null, 2)}
                  </pre>
                </div>

                {packet.raw_hex && (
                  <div className="mt-3">
                    <span className="text-xs text-slate-500 uppercase tracking-wider">Raw Hex</span>
                    <div className="mt-1 bg-slate-900/50 rounded-lg p-3 font-mono text-xs text-slate-400 break-all">
                      {packet.raw_hex}
                    </div>
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export default DeviceTelemetry;
