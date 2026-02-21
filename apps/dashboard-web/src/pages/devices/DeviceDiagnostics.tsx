import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

interface DeviceDetails {
  id: string;
  serial_number: string;
  type: string;
  model?: string;
  manufacturer?: string;
  firmware_version?: string;
  hardware_revision?: string;
  battery_level?: number;
  signal_strength?: number;
  status: string;
  last_heartbeat?: string;
  created_at: string;
}

interface SensorReading {
  name: string;
  value: string | number;
  unit?: string;
  status: string;
}

interface DiagnosticsLog {
  id: string;
  log_type: string;
  source: string;
  message: string;
  timestamp: string;
}

const DeviceDiagnostics: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [device, setDevice] = useState<DeviceDetails | null>(null);
  const [sensors, setSensors] = useState<SensorReading[]>([]);
  const [logs, setLogs] = useState<DiagnosticsLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (id) fetchDeviceData();
  }, [id]);

  const fetchDeviceData = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      const [deviceRes, sensorsRes, logsRes] = await Promise.all([
        fetch(`/api/v1/devices/${id}`, { headers }),
        fetch(`/api/v1/devices/${id}/sensors`, { headers }),
        fetch(`/api/v1/devices/${id}/logs?limit=50`, { headers })
      ]);

      if (deviceRes.ok) setDevice(await deviceRes.json());
      if (sensorsRes.ok) setSensors(await sensorsRes.json() || []);
      if (logsRes.ok) setLogs(await logsRes.json() || []);
    } catch (error) {
      console.error('Failed to fetch device data:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'normal': return 'text-emerald-400';
      case 'warning': return 'text-yellow-400';
      case 'critical': return 'text-red-400';
      default: return 'text-slate-400';
    }
  };

  const getLogTypeColor = (type: string) => {
    switch (type) {
      case 'error': return 'bg-red-500/20 text-red-400';
      case 'warning': return 'bg-yellow-500/20 text-yellow-400';
      case 'info': return 'bg-blue-500/20 text-blue-400';
      default: return 'bg-slate-500/20 text-slate-400';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 flex items-center justify-center">
        <div className="animate-pulse text-slate-400">Loading device diagnostics...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
      <div className="flex items-center gap-4 mb-8">
        <button
          onClick={() => navigate('/admin/devices')}
          className="p-2 rounded-lg bg-slate-800/50 text-slate-400 hover:text-white transition-colors"
        >
          ← Back
        </button>
        <div>
          <h1 className="text-2xl font-bold text-white">{device?.serial_number}</h1>
          <p className="text-slate-400">{device?.manufacturer} {device?.model}</p>
        </div>
        <span className={`ml-auto px-4 py-2 rounded-lg text-sm font-medium ${
          device?.status === 'active' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-400'
        }`}>
          {device?.status?.toUpperCase()}
        </span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <span>📱</span> Device Info
          </h2>
          <div className="space-y-3">
            <InfoRow label="Serial" value={device?.serial_number} />
            <InfoRow label="Type" value={device?.type?.toUpperCase()} />
            <InfoRow label="Manufacturer" value={device?.manufacturer} />
            <InfoRow label="Model" value={device?.model} />
            <InfoRow label="Hardware Rev" value={device?.hardware_revision} />
            <InfoRow label="Firmware" value={device?.firmware_version} />
            <InfoRow label="Registered" value={device?.created_at ? new Date(device.created_at).toLocaleDateString() : '-'} />
          </div>
        </div>

        <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <span>📊</span> Sensor Readings
          </h2>
          <div className="space-y-4">
            {sensors.length === 0 ? (
              <p className="text-slate-400 text-sm">No sensor data available</p>
            ) : (
              sensors.map((sensor, idx) => (
                <div key={idx} className="flex items-center justify-between p-3 bg-slate-900/50 rounded-lg">
                  <span className="text-slate-300">{sensor.name}</span>
                  <div className="flex items-center gap-2">
                    <span className={`font-bold ${getStatusColor(sensor.status)}`}>
                      {sensor.value}{sensor.unit}
                    </span>
                    <span className={`w-2 h-2 rounded-full ${
                      sensor.status === 'normal' ? 'bg-emerald-400' :
                      sensor.status === 'warning' ? 'bg-yellow-400' : 'bg-red-400'
                    }`}></span>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <span>⚡</span> Quick Actions
          </h2>
          <div className="space-y-3">
            <ActionButton label="View Telemetry" onClick={() => navigate(`/admin/devices/telemetry?device=${id}`)} />
            <ActionButton label="Send Command" onClick={() => navigate(`/admin/devices/commands?device=${id}`)} />
            <ActionButton label="Update Firmware" onClick={() => navigate(`/admin/devices/firmware`)} />
            <ActionButton label="Reboot Device" variant="warning" onClick={() => {}} />
          </div>
        </div>
      </div>

      <div className="mt-6 bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-6">
        <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <span>📋</span> Diagnostics Logs
        </h2>
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {logs.length === 0 ? (
            <p className="text-slate-400 text-sm">No logs available</p>
          ) : (
            logs.map((log) => (
              <div key={log.id} className="flex items-start gap-3 p-3 bg-slate-900/50 rounded-lg">
                <span className={`px-2 py-1 rounded text-xs font-medium ${getLogTypeColor(log.log_type)}`}>
                  {log.log_type.toUpperCase()}
                </span>
                <div className="flex-1">
                  <p className="text-slate-300 text-sm">{log.message}</p>
                  <p className="text-slate-500 text-xs mt-1">
                    {log.source} • {new Date(log.timestamp).toLocaleString()}
                  </p>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

const InfoRow: React.FC<{ label: string; value?: string }> = ({ label, value }) => (
  <div className="flex justify-between">
    <span className="text-slate-400 text-sm">{label}</span>
    <span className="text-white text-sm font-medium">{value || '-'}</span>
  </div>
);

const ActionButton: React.FC<{
  label: string;
  onClick: () => void;
  variant?: 'default' | 'warning';
}> = ({ label, onClick, variant = 'default' }) => (
  <button
    onClick={onClick}
    className={`w-full py-3 rounded-lg text-sm font-medium transition-all ${
      variant === 'warning'
        ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30 border border-red-500/30'
        : 'bg-violet-500/20 text-violet-400 hover:bg-violet-500/30 border border-violet-500/30'
    }`}
  >
    {label}
  </button>
);

export default DeviceDiagnostics;
