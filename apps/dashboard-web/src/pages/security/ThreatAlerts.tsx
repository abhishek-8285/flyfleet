import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

interface ThreatAlert {
  id: string;
  org_id: string;
  type: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  title?: string;
  description?: string;
  status: 'NEW' | 'ACKNOWLEDGED' | 'RESOLVED' | 'ESCALATED';
  source_ip?: string;
  target_resource?: string;
  timestamp: string;
  acknowledged_by?: string;
  acknowledged_at?: string;
  resolved_by?: string;
  resolved_at?: string;
}

const API_BASE = process.env.REACT_APP_API_URL || '/api';

export default function ThreatAlerts() {
  const [alerts, setAlerts] = useState<ThreatAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState({ severity: '', status: '' });
  const [searchParams] = useSearchParams();
  const orgId = searchParams.get('org_id') || '';

  useEffect(() => {
    fetchAlerts();
  }, [orgId, filter]);

  const fetchAlerts = async () => {
    setLoading(true);
    try {
      let url = `${API_BASE}/mfa/events?org_id=${orgId}&page_size=50`;
      if (filter.severity) url += `&severity=${filter.severity}`;
      if (filter.status) url += `&status=${filter.status}`;
      
      const res = await fetch(url);
      const data = await res.json();
      // Transform MFA events to threat alerts format
      const transformedAlerts = (data.events || []).map((event: any) => ({
        id: event.id,
        org_id: event.org_id,
        type: event.event_type,
        severity: event.status === 'failure' ? 'HIGH' : 'LOW',
        title: event.event_type.replace(/_/g, ' '),
        description: event.details,
        status: 'NEW',
        source_ip: event.ip_address,
        target_resource: `User: ${event.user_id}`,
        timestamp: event.created_at,
      }));
      setAlerts(transformedAlerts);
    } catch (err) {
      console.error('Failed to fetch alerts:', err);
    }
    setLoading(false);
  };

  const handleAcknowledge = async (alertId: string) => {
    // This would be a real API call in production
    setAlerts(alerts.map(a => 
      a.id === alertId ? { ...a, status: 'ACKNOWLEDGED' as const } : a
    ));
  };

  const handleResolve = async (alertId: string) => {
    // This would be a real API call in production
    setAlerts(alerts.map(a => 
      a.id === alertId ? { ...a, status: 'RESOLVED' as const } : a
    ));
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString();
  };

  const getSeverityStyles = (severity: string) => {
    const styles: Record<string, string> = {
      LOW: 'bg-blue-600/20 text-blue-400 border-blue-500',
      MEDIUM: 'bg-yellow-600/20 text-yellow-400 border-yellow-500',
      HIGH: 'bg-orange-600/20 text-orange-400 border-orange-500',
      CRITICAL: 'bg-red-600/20 text-red-400 border-red-500',
    };
    return styles[severity] || styles.LOW;
  };

  const getStatusStyles = (status: string) => {
    const styles: Record<string, string> = {
      NEW: 'bg-red-600/20 text-red-400',
      ACKNOWLEDGED: 'bg-yellow-600/20 text-yellow-400',
      RESOLVED: 'bg-green-600/20 text-green-400',
      ESCALATED: 'bg-purple-600/20 text-purple-400',
    };
    return styles[status] || '';
  };

  const criticalCount = alerts.filter(a => a.severity === 'CRITICAL').length;
  const highCount = alerts.filter(a => a.severity === 'HIGH').length;
  const newCount = alerts.filter(a => a.status === 'NEW').length;

  return (
    <div className="p-6 bg-gray-900 min-h-screen text-white">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">Threat Alerts</h1>
          <p className="text-gray-400">Monitor and respond to security threats in real-time</p>
        </div>
        <div className="flex gap-3">
          <select
            value={filter.severity}
            onChange={(e) => setFilter({ ...filter, severity: e.target.value })}
            className="px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none"
          >
            <option value="">All Severities</option>
            <option value="CRITICAL">Critical</option>
            <option value="HIGH">High</option>
            <option value="MEDIUM">Medium</option>
            <option value="LOW">Low</option>
          </select>
          <select
            value={filter.status}
            onChange={(e) => setFilter({ ...filter, status: e.target.value })}
            className="px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none"
          >
            <option value="">All Status</option>
            <option value="NEW">New</option>
            <option value="ACKNOWLEDGED">Acknowledged</option>
            <option value="RESOLVED">Resolved</option>
          </select>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-white">{alerts.length}</div>
          <div className="text-gray-400 text-sm">Total Alerts</div>
        </div>
        <div className="bg-red-900/30 backdrop-blur-lg rounded-xl p-4 border border-red-800/50">
          <div className="text-3xl font-bold text-red-400">{criticalCount}</div>
          <div className="text-gray-400 text-sm">Critical</div>
        </div>
        <div className="bg-orange-900/30 backdrop-blur-lg rounded-xl p-4 border border-orange-800/50">
          <div className="text-3xl font-bold text-orange-400">{highCount}</div>
          <div className="text-gray-400 text-sm">High Severity</div>
        </div>
        <div className="bg-yellow-900/30 backdrop-blur-lg rounded-xl p-4 border border-yellow-800/50">
          <div className="text-3xl font-bold text-yellow-400">{newCount}</div>
          <div className="text-gray-400 text-sm">Unacknowledged</div>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      ) : (
        <div className="space-y-4">
          {alerts.map((alert) => (
            <div
              key={alert.id}
              className={`bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border-l-4 ${getSeverityStyles(alert.severity)}`}
            >
              <div className="flex justify-between items-start mb-2">
                <div className="flex items-center gap-3">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${getSeverityStyles(alert.severity)}`}>
                    {alert.severity}
                  </span>
                  <h3 className="font-semibold text-lg">{alert.title || alert.type}</h3>
                </div>
                <span className={`px-2 py-1 rounded text-xs ${getStatusStyles(alert.status)}`}>
                  {alert.status}
                </span>
              </div>
              <p className="text-gray-400 mb-3">{alert.description || 'No description available'}</p>
              <div className="flex flex-wrap gap-4 text-sm text-gray-400 mb-3">
                <span>Source: <span className="text-blue-400 font-mono">{alert.source_ip || 'Unknown'}</span></span>
                <span>Target: <span className="text-purple-400">{alert.target_resource || 'N/A'}</span></span>
                <span>Time: <span className="text-gray-300">{formatDate(alert.timestamp)}</span></span>
              </div>
              {alert.status === 'NEW' && (
                <div className="flex gap-2">
                  <button
                    onClick={() => handleAcknowledge(alert.id)}
                    className="px-4 py-1.5 bg-yellow-600 hover:bg-yellow-700 rounded text-sm"
                  >
                    Acknowledge
                  </button>
                  <button
                    onClick={() => handleResolve(alert.id)}
                    className="px-4 py-1.5 bg-green-600 hover:bg-green-700 rounded text-sm"
                  >
                    Resolve
                  </button>
                </div>
              )}
              {alert.status === 'ACKNOWLEDGED' && (
                <button
                  onClick={() => handleResolve(alert.id)}
                  className="px-4 py-1.5 bg-green-600 hover:bg-green-700 rounded text-sm"
                >
                  Mark Resolved
                </button>
              )}
            </div>
          ))}
          {alerts.length === 0 && (
            <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-12 text-center border border-gray-700">
              <div className="text-4xl mb-4">🛡️</div>
              <h3 className="text-xl font-semibold mb-2">No Threat Alerts</h3>
              <p className="text-gray-400">Your system is currently secure with no active threats.</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
