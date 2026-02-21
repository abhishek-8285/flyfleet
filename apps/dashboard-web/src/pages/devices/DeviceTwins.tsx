import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface DeviceTwin {
  device_id: string;
  org_id: string;
  desired_state?: Record<string, unknown>;
  reported_state?: Record<string, unknown>;
  drift_status: 'in_sync' | 'drifted' | 'unknown';
  drift_reason?: string;
  last_seen?: string;
  last_command_id?: string;
  last_command_status?: string;
  updated_at: string;
}

const DeviceTwins: React.FC = () => {
  const navigate = useNavigate();
  const [twins, setTwins] = useState<DeviceTwin[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<DeviceTwin | null>(null);
  const [desiredEditor, setDesiredEditor] = useState<string>('{}');
  const [reportedEditor, setReportedEditor] = useState<string>('{}');
  const [saving, setSaving] = useState(false);

  const headers = (): HeadersInit => {
    const token = localStorage.getItem('access_token');
    const orgId = localStorage.getItem('org_id');
    const h: Record<string, string> = { 'Content-Type': 'application/json' };
    if (token) h.Authorization = `Bearer ${token}`;
    if (orgId) h['X-Organization-ID'] = orgId;
    return h;
  };

  const parseError = async (res: Response): Promise<string> => {
    try {
      const text = await res.text();
      return text || `Request failed (${res.status})`;
    } catch {
      return `Request failed (${res.status})`;
    }
  };

  const loadTwins = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch('/api/v1/devices/twins?limit=200', { headers: headers() });
      if (!res.ok) throw new Error(await parseError(res));
      const data = await res.json();
      setTwins(Array.isArray(data) ? data : []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load device twins');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadTwins();
  }, [loadTwins]);

  const openEditor = (twin: DeviceTwin) => {
    setSelected(twin);
    setDesiredEditor(JSON.stringify(twin.desired_state || {}, null, 2));
    setReportedEditor(JSON.stringify(twin.reported_state || {}, null, 2));
  };

  const saveDesired = async () => {
    if (!selected) return;
    setSaving(true);
    setError(null);
    try {
      const desiredState = JSON.parse(desiredEditor);
      const res = await fetch(`/api/v1/devices/twins/${selected.device_id}/desired`, {
        method: 'PUT',
        headers: headers(),
        body: JSON.stringify({ desired_state: desiredState }),
      });
      if (!res.ok) throw new Error(await parseError(res));
      await loadTwins();
    } catch (err: any) {
      setError(err?.message || 'Failed to save desired state');
    } finally {
      setSaving(false);
    }
  };

  const saveReported = async () => {
    if (!selected) return;
    setSaving(true);
    setError(null);
    try {
      const reportedState = JSON.parse(reportedEditor);
      const res = await fetch(`/api/v1/devices/twins/${selected.device_id}/reported`, {
        method: 'PUT',
        headers: headers(),
        body: JSON.stringify({ reported_state: reportedState }),
      });
      if (!res.ok) throw new Error(await parseError(res));
      await loadTwins();
    } catch (err: any) {
      setError(err?.message || 'Failed to save reported state');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/admin/devices/diagnostics')} className="p-2 rounded-lg bg-slate-800/50 text-slate-400 hover:text-white">← Back</button>
          <div>
            <h1 className="text-2xl font-bold text-white">Device Twins</h1>
            <p className="text-slate-400">Desired vs reported state convergence</p>
          </div>
        </div>
      </div>

      {error && <div className="mb-4 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300">{error}</div>}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-slate-800/50 rounded-xl border border-slate-700/50 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-700/50">
                <th className="text-left py-3 px-4 text-xs text-slate-400 uppercase">Device</th>
                <th className="text-left py-3 px-4 text-xs text-slate-400 uppercase">Drift</th>
                <th className="text-left py-3 px-4 text-xs text-slate-400 uppercase">Command</th>
                <th className="py-3 px-4" />
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={4} className="py-10 text-center text-slate-400">Loading...</td></tr>
              ) : twins.length === 0 ? (
                <tr><td colSpan={4} className="py-10 text-center text-slate-400">No twin records</td></tr>
              ) : twins.map((t) => (
                <tr key={t.device_id} className="border-b border-slate-700/20 hover:bg-slate-700/20">
                  <td className="py-3 px-4 text-slate-200 font-mono text-xs">{t.device_id.slice(0, 8)}</td>
                  <td className="py-3 px-4">
                    <span className={`px-2 py-1 rounded text-xs ${
                      t.drift_status === 'in_sync' ? 'bg-emerald-500/20 text-emerald-400' :
                      t.drift_status === 'drifted' ? 'bg-red-500/20 text-red-400' : 'bg-slate-500/20 text-slate-300'
                    }`}>{t.drift_status}</span>
                  </td>
                  <td className="py-3 px-4 text-xs text-slate-400">{t.last_command_status || '-'}</td>
                  <td className="py-3 px-4 text-right">
                    <button onClick={() => openEditor(t)} className="text-violet-300 text-xs hover:text-violet-200">Edit</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="bg-slate-800/50 rounded-xl border border-slate-700/50 p-4">
          {!selected ? (
            <div className="text-slate-400 text-sm">Select a twin to inspect/edit state.</div>
          ) : (
            <div className="space-y-4">
              <div className="text-xs text-slate-400">Device: <span className="text-slate-200 font-mono">{selected.device_id}</span></div>
              <div>
                <label className="block text-xs uppercase text-slate-400 mb-2">Desired State</label>
                <textarea value={desiredEditor} onChange={(e) => setDesiredEditor(e.target.value)} rows={10} className="w-full bg-slate-950 border border-slate-700 rounded p-3 text-xs text-slate-100 font-mono" />
                <button onClick={() => void saveDesired()} disabled={saving} className="mt-2 px-3 py-2 bg-violet-600 text-white rounded text-xs disabled:opacity-60">Save Desired</button>
              </div>
              <div>
                <label className="block text-xs uppercase text-slate-400 mb-2">Reported State</label>
                <textarea value={reportedEditor} onChange={(e) => setReportedEditor(e.target.value)} rows={10} className="w-full bg-slate-950 border border-slate-700 rounded p-3 text-xs text-slate-100 font-mono" />
                <button onClick={() => void saveReported()} disabled={saving} className="mt-2 px-3 py-2 bg-blue-600 text-white rounded text-xs disabled:opacity-60">Save Reported</button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default DeviceTwins;
