import {
    ArrowPathIcon,
    CheckCircleIcon,
    CloudIcon,
    Cog6ToothIcon,
    ExclamationCircleIcon,
    PlusIcon
} from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import { api } from '../../../services/api';

interface Connector {
    id: string;
    name: string;
    type: 'SAP' | 'SALESFORCE' | 'ORACLE' | 'DYNAMICS';
    status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
    lastSync: string;
}

interface IntegrationLog {
    id: string;
    integration_name?: string;
    integration_id?: string;
    entity?: string;
    direction?: string;
    status?: string;
    created_at?: string;
}

const IntegrationHub: React.FC = () => {
    const [connectors, setConnectors] = useState<Connector[]>([]);
    const [logs, setLogs] = useState<IntegrationLog[]>([]);
    const [loading, setLoading] = useState(true);
    const [logsLoading, setLogsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const orgId = localStorage.getItem('org_id') || 'default';

    const fetchIntegrations = async () => {
        try {
            setLoading(true);
            setError(null);
            const response = await api.getIntegrations(orgId);
            const data = (response as any)?.data ?? response;
            const items = data?.items || data?.integrations || data?.data || data || [];
            const normalized = Array.isArray(items)
                ? items.map((item: any) => ({
                    id: item.id,
                    name: item.name || item.integration_name || item.provider || 'Integration',
                    type: (item.type || item.integration_type || 'ORACLE').toUpperCase(),
                    status: (item.status || item.connection_status || 'DISCONNECTED').toUpperCase(),
                    lastSync: item.last_sync_at || item.lastSync || item.updated_at || '—'
                }))
                : [];
            setConnectors(normalized);
        } catch (err) {
            setError('Failed to load integrations.');
            setConnectors([]);
        } finally {
            setLoading(false);
        }
    };

    const fetchLogs = async () => {
        try {
            setLogsLoading(true);
            const response = await api.getIntegrationLogs({ org_id: orgId, limit: 10, offset: 0 });
            const data = (response as any)?.data ?? response;
            const items = data?.items || data?.logs || data?.data || data || [];
            setLogs(Array.isArray(items) ? items : []);
        } catch (err) {
            setLogs([]);
        } finally {
            setLogsLoading(false);
        }
    };

    useEffect(() => {
        fetchIntegrations();
        fetchLogs();
    }, []);

    return (
        <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-white">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-2">
                        <CloudIcon className="h-8 w-8 text-blue-500" />
                        Enterprise Integration Hub
                    </h1>
                    <p className="text-slate-400">Manage ERP connectors and data synchronization</p>
                </div>
                <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition flex items-center gap-2">
                    <PlusIcon className="h-5 w-5" />
                    Add Connector
                </button>
            </div>

            {/* Connectors Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {loading ? (
                    <div className="col-span-full text-slate-400">Loading integrations...</div>
                ) : error ? (
                    <div className="col-span-full text-red-400">{error}</div>
                ) : connectors.length === 0 ? (
                    <div className="col-span-full text-slate-400">No integrations connected.</div>
                ) : connectors.map((connector) => (
                    <div key={connector.id} className="bg-slate-800 p-6 rounded-xl border border-slate-700 relative group hover:border-blue-500/50 transition">
                        <div className="absolute top-4 right-4">
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${connector.status === 'CONNECTED' ? 'bg-green-100 text-green-800' :
                                    connector.status === 'ERROR' ? 'bg-red-100 text-red-800' : 'bg-slate-100 text-slate-800'
                                }`}>
                                {connector.status}
                            </span>
                        </div>

                        <div className="flex items-center gap-4 mb-4">
                            <div className={`h-12 w-12 rounded-lg flex items-center justify-center text-xl font-bold ${connector.type === 'SAP' ? 'bg-blue-900 text-blue-200' :
                                    connector.type === 'SALESFORCE' ? 'bg-sky-900 text-sky-200' :
                                        'bg-red-900 text-red-200'
                                }`}>
                                {connector.type.substring(0, 2)}
                            </div>
                            <div>
                                <h3 className="font-bold text-white">{connector.name}</h3>
                                <p className="text-xs text-slate-400">ID: {connector.id}</p>
                            </div>
                        </div>

                        <div className="flex justify-between items-center text-sm text-slate-400 mt-4 border-t border-slate-700 pt-4">
                            <span className="flex items-center gap-1">
                                <ArrowPathIcon className="h-4 w-4" />
                                {connector.lastSync}
                            </span>
                            <button className="text-blue-400 hover:text-blue-300 flex items-center gap-1">
                                <Cog6ToothIcon className="h-4 w-4" />
                                Configure
                            </button>
                        </div>
                    </div>
                ))}
            </div>

            {/* Sync Logs */}
            <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
                <div className="p-6 border-b border-slate-700 flex justify-between items-center">
                    <h3 className="text-lg font-bold text-white">Recent Sync Jobs</h3>
                    <button onClick={fetchLogs} className="text-sm text-blue-400 hover:text-blue-300">Refresh Logs</button>
                </div>
                <table className="w-full text-left text-sm text-slate-400">
                    <thead className="bg-slate-900/50 uppercase font-medium">
                        <tr>
                            <th className="px-6 py-4">Job ID</th>
                            <th className="px-6 py-4">Connector</th>
                            <th className="px-6 py-4">Entity</th>
                            <th className="px-6 py-4">Direction</th>
                            <th className="px-6 py-4">Status</th>
                            <th className="px-6 py-4">Time</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700">
                        {logsLoading ? (
                            <tr>
                                <td colSpan={6} className="px-6 py-6 text-slate-500">Loading logs...</td>
                            </tr>
                        ) : logs.length === 0 ? (
                            <tr>
                                <td colSpan={6} className="px-6 py-6 text-slate-500">No recent sync jobs.</td>
                            </tr>
                        ) : logs.map((log) => (
                            <tr key={log.id}>
                                <td className="px-6 py-4 font-mono text-xs">{log.id}</td>
                                <td className="px-6 py-4 text-white">{log.integration_name || log.integration_id || 'Unknown'}</td>
                                <td className="px-6 py-4">{log.entity || '—'}</td>
                                <td className="px-6 py-4">
                                    <span className={log.direction === 'OUTBOUND' ? 'text-green-400' : 'text-blue-400'}>
                                        {log.direction || 'UNKNOWN'}
                                    </span>
                                </td>
                                <td className="px-6 py-4">
                                    {log.status === 'SUCCESS' || log.status === 'COMPLETED' ? (
                                        <span className="flex items-center gap-1 text-green-500"><CheckCircleIcon className="h-4 w-4" /> Success</span>
                                    ) : (
                                        <span className="flex items-center gap-1 text-red-500"><ExclamationCircleIcon className="h-4 w-4" /> Failed</span>
                                    )}
                                </td>
                                <td className="px-6 py-4">{log.created_at ? new Date(log.created_at).toLocaleString() : '—'}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default IntegrationHub;
