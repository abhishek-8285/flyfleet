import React, { useCallback, useEffect, useState } from 'react';
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
  completed_at?: string;
  created_at: string;
}

interface FirmwareRolloutEvent {
  id: string;
  rollout_id: string;
  event_type: string;
  from_status?: string;
  to_status?: string;
  actor_user_id?: string;
  message: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

interface FirmwareUploadForm {
  version: string;
  hardware_target: string;
  file_url: string;
  file_size: string;
  checksum: string;
  release_notes: string;
  is_critical: boolean;
}

interface RolloutForm {
  firmware_id: string;
  name: string;
  device_group: string;
  total_devices: string;
}

type RolloutEventFilter = 'all' | 'transition' | 'created' | 'note';

const emptyUploadForm: FirmwareUploadForm = {
  version: '',
  hardware_target: '',
  file_url: '',
  file_size: '',
  checksum: '',
  release_notes: '',
  is_critical: false,
};

const emptyRolloutForm: RolloutForm = {
  firmware_id: '',
  name: '',
  device_group: '',
  total_devices: '10',
};

const FirmwareManager: React.FC = () => {
  const navigate = useNavigate();
  const [firmware, setFirmware] = useState<FirmwarePackage[]>([]);
  const [rollouts, setRollouts] = useState<FirmwareRollout[]>([]);
  const [activeTab, setActiveTab] = useState<'packages' | 'rollouts'>('packages');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [updatingRolloutId, setUpdatingRolloutId] = useState<string | null>(null);
  const [selectedRolloutId, setSelectedRolloutId] = useState<string | null>(null);
  const [selectedRollout, setSelectedRollout] = useState<FirmwareRollout | null>(null);
  const [rolloutDetailLoading, setRolloutDetailLoading] = useState(false);
  const [selectedRolloutEvents, setSelectedRolloutEvents] = useState<FirmwareRolloutEvent[]>([]);
  const [rolloutEventsLoading, setRolloutEventsLoading] = useState(false);
  const [rolloutEventFilter, setRolloutEventFilter] = useState<RolloutEventFilter>('all');
  const [rolloutNote, setRolloutNote] = useState('');
  const [rolloutNoteSubmitting, setRolloutNoteSubmitting] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [showUploadModal, setShowUploadModal] = useState(false);
  const [showRolloutModal, setShowRolloutModal] = useState(false);

  const [uploadForm, setUploadForm] = useState<FirmwareUploadForm>(emptyUploadForm);
  const [rolloutForm, setRolloutForm] = useState<RolloutForm>(emptyRolloutForm);

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
      if (text) {
        return text;
      }
    } catch (err) {
      // no-op
    }
    return `Request failed (${res.status})`;
  };

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const headers = getAuthHeaders();
      const [firmwareRes, rolloutsRes] = await Promise.all([
        fetch('/api/v1/firmware?limit=100', { headers }),
        fetch('/api/v1/firmware/rollouts?limit=100', { headers }),
      ]);

      if (!firmwareRes.ok) {
        throw new Error(await parseError(firmwareRes));
      }
      if (!rolloutsRes.ok) {
        throw new Error(await parseError(rolloutsRes));
      }

      const firmwareData = (await firmwareRes.json()) as FirmwarePackage[];
      const rolloutData = (await rolloutsRes.json()) as FirmwareRollout[];
      setFirmware(Array.isArray(firmwareData) ? firmwareData : []);
      setRollouts(Array.isArray(rolloutData) ? rolloutData : []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load firmware data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'completed':
        return 'bg-emerald-500/20 text-emerald-400';
      case 'testing':
      case 'paused':
        return 'bg-yellow-500/20 text-yellow-400';
      case 'draft':
      case 'scheduled':
        return 'bg-violet-500/20 text-violet-400';
      case 'in_progress':
        return 'bg-blue-500/20 text-blue-400';
      case 'deprecated':
      case 'failed':
      case 'cancelled':
        return 'bg-red-500/20 text-red-400';
      default:
        return 'bg-slate-500/20 text-slate-400';
    }
  };

  const formatFileSize = (bytes?: number) => {
    if (!bytes) return '-';
    const mb = bytes / (1024 * 1024);
    return `${mb.toFixed(2)} MB`;
  };

  const handleUploadFirmware = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setSuccess(null);

    try {
      const payload = {
        version: uploadForm.version.trim(),
        hardware_target: uploadForm.hardware_target.trim(),
        file_url: uploadForm.file_url.trim(),
        file_size: Number(uploadForm.file_size),
        checksum: uploadForm.checksum.trim() || undefined,
        release_notes: uploadForm.release_notes.trim() || undefined,
        is_critical: uploadForm.is_critical,
      };

      if (!payload.version || !payload.hardware_target || !payload.file_url || !Number.isFinite(payload.file_size) || payload.file_size <= 0) {
        throw new Error('Version, target, file URL, and file size are required.');
      }

      const res = await fetch('/api/v1/firmware', {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        throw new Error(await parseError(res));
      }

      setShowUploadModal(false);
      setUploadForm(emptyUploadForm);
      setSuccess('Firmware package uploaded successfully.');
      await fetchData();
      setActiveTab('packages');
    } catch (err: any) {
      setError(err?.message || 'Failed to upload firmware package');
    } finally {
      setSubmitting(false);
    }
  };

  const openRolloutModal = (fw: FirmwarePackage) => {
    setRolloutForm({
      firmware_id: fw.id,
      name: `${fw.version} rollout`,
      device_group: fw.hardware_target,
      total_devices: '10',
    });
    setShowRolloutModal(true);
  };

  const handleCreateRollout = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setSuccess(null);

    try {
      const totalDevices = Number(rolloutForm.total_devices);
      if (!rolloutForm.firmware_id || !rolloutForm.name.trim() || !rolloutForm.device_group.trim() || !Number.isInteger(totalDevices) || totalDevices <= 0) {
        throw new Error('Firmware, rollout name, device group, and total devices are required.');
      }

      const res = await fetch('/api/v1/firmware/rollouts', {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
          firmware_id: rolloutForm.firmware_id,
          name: rolloutForm.name.trim(),
          device_group: rolloutForm.device_group.trim(),
          total_devices: totalDevices,
        }),
      });

      if (!res.ok) {
        throw new Error(await parseError(res));
      }

      setShowRolloutModal(false);
      setRolloutForm(emptyRolloutForm);
      setSuccess('Firmware rollout created successfully.');
      await fetchData();
      setActiveTab('rollouts');
    } catch (err: any) {
      setError(err?.message || 'Failed to create rollout');
    } finally {
      setSubmitting(false);
    }
  };

  const transitionRollout = async (rolloutId: string, nextStatus: string) => {
    setUpdatingRolloutId(rolloutId);
    setError(null);
    setSuccess(null);
    try {
      let reason: string | undefined;
      if (nextStatus === 'cancelled' || nextStatus === 'completed') {
        const input = window.prompt(`Optional reason for marking rollout ${nextStatus}:`, '');
        reason = input?.trim() || undefined;
      }
      const res = await fetch(`/api/v1/firmware/rollouts/${rolloutId}/status`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
        body: JSON.stringify({ status: nextStatus, reason }),
      });
      if (!res.ok) {
        throw new Error(await parseError(res));
      }
      setSuccess(`Rollout moved to ${nextStatus}`);
      await fetchData();
      if (selectedRolloutId === rolloutId) {
        await fetchRolloutDetail(rolloutId);
      }
    } catch (err: any) {
      setError(err?.message || `Failed to update rollout to ${nextStatus}`);
    } finally {
      setUpdatingRolloutId(null);
    }
  };

  const fetchRolloutDetail = useCallback(
    async (rolloutId: string) => {
      setRolloutDetailLoading(true);
      try {
        const res = await fetch(`/api/v1/firmware/rollouts/${rolloutId}`, {
          headers: getAuthHeaders(),
        });
        if (!res.ok) {
          throw new Error(await parseError(res));
        }
        const detail = (await res.json()) as FirmwareRollout;
        setSelectedRollout(detail);
      } catch (err: any) {
        setError(err?.message || 'Failed to fetch rollout details');
      } finally {
        setRolloutDetailLoading(false);
      }
    },
    [],
  );

  const fetchRolloutEvents = useCallback(async (rolloutId: string, eventFilter: RolloutEventFilter) => {
    setRolloutEventsLoading(true);
    try {
      const params = new URLSearchParams();
      params.set('limit', '50');
      if (eventFilter !== 'all') {
        params.set('event_type', eventFilter);
      }
      const res = await fetch(`/api/v1/firmware/rollouts/${rolloutId}/events?${params.toString()}`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) {
        throw new Error(await parseError(res));
      }
      const events = (await res.json()) as FirmwareRolloutEvent[];
      setSelectedRolloutEvents(Array.isArray(events) ? events : []);
    } catch (err: any) {
      setError(err?.message || 'Failed to fetch rollout events');
    } finally {
      setRolloutEventsLoading(false);
    }
  }, []);

  const addRolloutNote = async () => {
    if (!selectedRolloutId) return;
    const message = rolloutNote.trim();
    if (!message) return;
    setRolloutNoteSubmitting(true);
    setError(null);
    setSuccess(null);
    try {
      const res = await fetch(`/api/v1/firmware/rollouts/${selectedRolloutId}/events`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
          message,
          metadata: { source: 'dashboard-web' },
        }),
      });
      if (!res.ok) {
        throw new Error(await parseError(res));
      }
      setRolloutNote('');
      setSuccess('Rollout note added');
      await fetchRolloutEvents(selectedRolloutId, rolloutEventFilter);
    } catch (err: any) {
      setError(err?.message || 'Failed to add rollout note');
    } finally {
      setRolloutNoteSubmitting(false);
    }
  };

  useEffect(() => {
    if (!selectedRolloutId) {
      setSelectedRollout(null);
      setSelectedRolloutEvents([]);
      return;
    }
    void fetchRolloutDetail(selectedRolloutId);
    void fetchRolloutEvents(selectedRolloutId, rolloutEventFilter);
  }, [selectedRolloutId, rolloutEventFilter, fetchRolloutDetail, fetchRolloutEvents]);

  useEffect(() => {
    if (!selectedRolloutId || activeTab !== 'rollouts') return;
    const timer = window.setInterval(() => {
      void fetchRolloutDetail(selectedRolloutId);
      void fetchRolloutEvents(selectedRolloutId, rolloutEventFilter);
    }, 15000);
    return () => window.clearInterval(timer);
  }, [selectedRolloutId, rolloutEventFilter, activeTab, fetchRolloutDetail, fetchRolloutEvents]);

  const getRolloutActions = (status: string): string[] => {
    switch (status) {
      case 'scheduled':
        return ['in_progress', 'cancelled'];
      case 'in_progress':
        return ['paused', 'completed', 'cancelled'];
      case 'paused':
        return ['in_progress', 'cancelled'];
      default:
        return [];
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

      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('packages')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'packages'
              ? 'bg-violet-600 text-white'
              : 'bg-slate-800/50 text-slate-400 hover:bg-slate-700/50'
          }`}
        >
          Packages
        </button>
        <button
          onClick={() => setActiveTab('rollouts')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'rollouts'
              ? 'bg-violet-600 text-white'
              : 'bg-slate-800/50 text-slate-400 hover:bg-slate-700/50'
          }`}
        >
          Rollouts
        </button>
      </div>

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
                <div className="flex items-start justify-between gap-4">
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
                    <div className="flex flex-wrap gap-6 text-sm">
                      <span className="text-slate-400">Size: <span className="text-white">{formatFileSize(fw.file_size)}</span></span>
                      {fw.checksum && <span className="text-slate-400">Checksum: <span className="text-white font-mono">{fw.checksum.slice(0, 12)}...</span></span>}
                      <span className="text-slate-400">Created: <span className="text-white">{new Date(fw.created_at).toLocaleDateString()}</span></span>
                    </div>
                  </div>
                  <button
                    onClick={() => openRolloutModal(fw)}
                    className="px-4 py-2 bg-violet-500/20 text-violet-400 rounded-lg hover:bg-violet-500/30 transition-colors text-sm"
                  >
                    Start Rollout
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-4">
          <div className="xl:col-span-2 bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 overflow-hidden">
            <table className="w-full">
            <thead>
              <tr className="border-b border-slate-700/50">
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Rollout</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Status</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Progress</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Target</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Started</th>
                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody>
              {rollouts.length === 0 ? (
                <tr>
                  <td colSpan={6} className="text-center py-12 text-slate-400">
                    No active rollouts
                  </td>
                </tr>
              ) : (
                rollouts.map((r) => (
                  <tr
                    key={r.id}
                    className={`border-b border-slate-700/30 hover:bg-slate-700/30 ${
                      selectedRolloutId === r.id ? 'bg-slate-700/40' : ''
                    }`}
                  >
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
                          />
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
                    <td className="py-4 px-6">
                      <div className="flex flex-wrap gap-2">
                        <button
                          onClick={() => setSelectedRolloutId(r.id)}
                          className="px-2 py-1 rounded text-xs bg-blue-600/20 text-blue-300 hover:bg-blue-600/30"
                        >
                          View
                        </button>
                        {getRolloutActions(r.status).map((action) => (
                          <button
                            key={`${r.id}-${action}`}
                            disabled={updatingRolloutId === r.id}
                            onClick={() => void transitionRollout(r.id, action)}
                            className="px-2 py-1 rounded text-xs bg-slate-700 text-slate-200 hover:bg-slate-600 disabled:opacity-60"
                          >
                            {updatingRolloutId === r.id ? 'Updating...' : action.replace('_', ' ')}
                          </button>
                        ))}
                        {getRolloutActions(r.status).length === 0 && (
                          <span className="text-xs text-slate-500">No actions</span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
            </table>
          </div>
          <div className="bg-slate-800/50 backdrop-blur-xl rounded-xl border border-slate-700/50 p-5">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-semibold">Rollout Details</h3>
              {selectedRolloutId && (
                <button
                  onClick={() => {
                    void fetchRolloutDetail(selectedRolloutId);
                    void fetchRolloutEvents(selectedRolloutId, rolloutEventFilter);
                  }}
                  className="text-xs px-2 py-1 bg-slate-700 rounded text-slate-200 hover:bg-slate-600"
                >
                  Refresh
                </button>
              )}
            </div>
            {!selectedRolloutId ? (
              <p className="text-sm text-slate-400">Select a rollout to inspect live state.</p>
            ) : rolloutDetailLoading && !selectedRollout ? (
              <p className="text-sm text-slate-400 animate-pulse">Loading details...</p>
            ) : selectedRollout ? (
              <div className="space-y-3 text-sm">
                <div>
                  <p className="text-slate-400">Name</p>
                  <p className="text-white">{selectedRollout.name || selectedRollout.id}</p>
                </div>
                <div>
                  <p className="text-slate-400">Status</p>
                  <span className={`px-2 py-1 rounded text-xs font-medium ${getStatusColor(selectedRollout.status)}`}>
                    {selectedRollout.status.replace('_', ' ').toUpperCase()}
                  </span>
                </div>
                <div>
                  <p className="text-slate-400">Device Group</p>
                  <p className="text-white">{selectedRollout.device_group || 'All devices'}</p>
                </div>
                <div>
                  <p className="text-slate-400">Progress</p>
                  <p className="text-white">
                    {selectedRollout.completed_devices}/{selectedRollout.total_devices} completed
                    {selectedRollout.failed_devices > 0 ? `, ${selectedRollout.failed_devices} failed` : ''}
                  </p>
                </div>
                <div>
                  <p className="text-slate-400">Started At</p>
                  <p className="text-white">{selectedRollout.started_at ? new Date(selectedRollout.started_at).toLocaleString() : '-'}</p>
                </div>
                <div>
                  <p className="text-slate-400">Completed At</p>
                  <p className="text-white">{selectedRollout.completed_at ? new Date(selectedRollout.completed_at).toLocaleString() : '-'}</p>
                </div>
                <div>
                  <p className="text-slate-400">Created At</p>
                  <p className="text-white">{new Date(selectedRollout.created_at).toLocaleString()}</p>
                </div>
                <div className="pt-1">
                  <div className="mb-2 flex items-center justify-between gap-3">
                    <p className="text-slate-400">Recent Events</p>
                    <select
                      value={rolloutEventFilter}
                      onChange={(e) => setRolloutEventFilter(e.target.value as RolloutEventFilter)}
                      className="rounded bg-slate-900 border border-slate-700 px-2 py-1 text-xs text-slate-200"
                    >
                      <option value="all">All</option>
                      <option value="transition">Transitions</option>
                      <option value="created">Created</option>
                      <option value="note">Notes</option>
                    </select>
                  </div>
                  <div className="mb-2">
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={rolloutNote}
                        onChange={(e) => setRolloutNote(e.target.value)}
                        placeholder="Add operator note"
                        className="flex-1 rounded bg-slate-900 border border-slate-700 px-2 py-1 text-xs text-white"
                      />
                      <button
                        onClick={() => void addRolloutNote()}
                        disabled={rolloutNoteSubmitting || !rolloutNote.trim()}
                        className="px-2 py-1 rounded text-xs bg-violet-600 text-white disabled:opacity-60"
                      >
                        {rolloutNoteSubmitting ? 'Saving...' : 'Add'}
                      </button>
                    </div>
                  </div>
                  {rolloutEventsLoading ? (
                    <p className="text-xs text-slate-500">Loading events...</p>
                  ) : selectedRolloutEvents.length === 0 ? (
                    <p className="text-xs text-slate-500">No events yet.</p>
                  ) : (
                    <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
                      {selectedRolloutEvents.map((event) => (
                        <div key={event.id} className="rounded border border-slate-700/60 bg-slate-900/40 p-2">
                          <p className="text-xs text-slate-200">{event.message}</p>
                          <p className="text-[11px] text-slate-500 mt-1">
                            {event.event_type}
                            {event.actor_user_id ? ` • actor ${event.actor_user_id.slice(0, 8)}` : ''}
                            {' • '}
                            {new Date(event.created_at).toLocaleString()}
                          </p>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
                {rolloutDetailLoading && (
                  <p className="text-xs text-slate-500">Refreshing...</p>
                )}
              </div>
            ) : (
              <p className="text-sm text-slate-400">No rollout details available.</p>
            )}
          </div>
        </div>
      )}

      {showUploadModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-xl border border-slate-700 p-6 w-full max-w-md">
            <h2 className="text-xl font-bold text-white mb-4">Upload Firmware</h2>
            <form className="space-y-4" onSubmit={handleUploadFirmware}>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Version</label>
                <input
                  type="text"
                  value={uploadForm.version}
                  onChange={(e) => setUploadForm((prev) => ({ ...prev, version: e.target.value }))}
                  placeholder="1.0.0"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Hardware Target</label>
                <input
                  type="text"
                  value={uploadForm.hardware_target}
                  onChange={(e) => setUploadForm((prev) => ({ ...prev, hardware_target: e.target.value }))}
                  placeholder="HW-REV-04"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">File URL</label>
                <input
                  type="url"
                  value={uploadForm.file_url}
                  onChange={(e) => setUploadForm((prev) => ({ ...prev, file_url: e.target.value }))}
                  placeholder="https://..."
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">File Size (bytes)</label>
                <input
                  type="number"
                  min={1}
                  value={uploadForm.file_size}
                  onChange={(e) => setUploadForm((prev) => ({ ...prev, file_size: e.target.value }))}
                  placeholder="1048576"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Checksum (optional)</label>
                <input
                  type="text"
                  value={uploadForm.checksum}
                  onChange={(e) => setUploadForm((prev) => ({ ...prev, checksum: e.target.value }))}
                  placeholder="sha256 checksum"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Release Notes</label>
                <textarea
                  rows={3}
                  value={uploadForm.release_notes}
                  onChange={(e) => setUploadForm((prev) => ({ ...prev, release_notes: e.target.value }))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                />
              </div>
              <label className="flex items-center gap-2 text-sm text-slate-300">
                <input
                  type="checkbox"
                  checked={uploadForm.is_critical}
                  onChange={(e) => setUploadForm((prev) => ({ ...prev, is_critical: e.target.checked }))}
                />
                Mark as critical release
              </label>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  disabled={submitting}
                  onClick={() => setShowUploadModal(false)}
                  className="flex-1 py-2 bg-slate-700 text-white rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="flex-1 py-2 bg-violet-600 text-white rounded-lg disabled:opacity-70"
                >
                  {submitting ? 'Uploading...' : 'Upload'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showRolloutModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-xl border border-slate-700 p-6 w-full max-w-md">
            <h2 className="text-xl font-bold text-white mb-4">Create Rollout</h2>
            <form className="space-y-4" onSubmit={handleCreateRollout}>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Rollout Name</label>
                <input
                  type="text"
                  value={rolloutForm.name}
                  onChange={(e) => setRolloutForm((prev) => ({ ...prev, name: e.target.value }))}
                  placeholder="Jan safety patch rollout"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Device Group</label>
                <input
                  type="text"
                  value={rolloutForm.device_group}
                  onChange={(e) => setRolloutForm((prev) => ({ ...prev, device_group: e.target.value }))}
                  placeholder="region-west"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Total Devices</label>
                <input
                  type="number"
                  min={1}
                  value={rolloutForm.total_devices}
                  onChange={(e) => setRolloutForm((prev) => ({ ...prev, total_devices: e.target.value }))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white"
                  required
                />
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  disabled={submitting}
                  onClick={() => setShowRolloutModal(false)}
                  className="flex-1 py-2 bg-slate-700 text-white rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="flex-1 py-2 bg-violet-600 text-white rounded-lg disabled:opacity-70"
                >
                  {submitting ? 'Creating...' : 'Create Rollout'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default FirmwareManager;
