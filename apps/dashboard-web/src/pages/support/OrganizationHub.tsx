import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

const OrganizationHub: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const [summary, setSummary] = useState<any>(null);
    const [stats, setStats] = useState<any>(null);
    const [health, setHealth] = useState<any>(null);
    const [fleet, setFleet] = useState<any>(null);
    const [diagnostics, setDiagnostics] = useState<any[]>([]);
    const [alerts, setAlerts] = useState<any[]>([]);
    const [contacts, setContacts] = useState<any[]>([]);
    const [activeTab, setActiveTab] = useState('fleet');
    const [isLoading, setIsLoading] = useState(true);

    const fetchData = React.useCallback(async () => {
        const token = localStorage.getItem('access_token');
        const headers = { 'Authorization': `Bearer ${token}` };

        try {
            const [sumRes, statRes, healthRes, fleetRes, diagRes, alertRes, contactRes] = await Promise.all([
                fetch(`/api/v1/support/orgs/${id}/summary`, { headers }),
                fetch(`/api/v1/support/orgs/${id}/devices/stats`, { headers }),
                fetch(`/api/v1/support/orgs/${id}/integration/health`, { headers }),
                fetch(`/api/v1/support/orgs/${id}/fleet/locations`, { headers }),
                fetch(`/api/v1/support/orgs/${id}/diagnostics/recent`, { headers }),
                fetch(`/api/v1/support/orgs/${id}/alerts`, { headers }),
                fetch(`/api/v1/support/orgs/${id}/contacts`, { headers })
            ]);

            if (sumRes.ok) setSummary(await sumRes.json());
            if (statRes.ok) setStats(await statRes.json());
            if (healthRes.ok) setHealth(await healthRes.json());
            if (fleetRes.ok) setFleet(await fleetRes.json());
            if (diagRes.ok) setDiagnostics(await diagRes.json());
            if (alertRes.ok) setAlerts(await alertRes.json());
            if (contactRes.ok) setContacts(await contactRes.json());
        } catch (err) {
            console.error('Failed to fetch org hub data', err);
        } finally {
            setIsLoading(false);
        }
    }, [id]);

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 30000);
        return () => clearInterval(interval);
    }, [id]);

    const handleImpersonate = async () => {
        const token = localStorage.getItem('access_token');
        try {
            const res = await fetch(`/api/v1/support/orgs/${id}/impersonate`, {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (res.ok) {
                const data = await res.json();
                alert(`Impersonation active. Token: ${data.sessionToken} (Read-only mode)`);
            }
        } catch (err) {
            alert('Failed to initialize impersonation protocol.');
        }
    };

    if (isLoading) {
        return (
            <div className="flex-1 flex items-center justify-center bg-[#0b1219]">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                    <span className="text-[#92adc9] text-xs font-bold uppercase tracking-widest animate-pulse">Initializing Diagnostics...</span>
                </div>
            </div>
        );
    }

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden font-display">
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <h2 className="text-white text-lg font-bold">Organization Support Console</h2>
                        <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5 uppercase tracking-wide">
                            <span>Support</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span>Organization Console</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-white">{summary?.name || 'Loading...'}</span>
                        </div>
                    </div>
                    <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-emerald-500/10 text-emerald-500 border border-emerald-500/30">
                        {summary?.status || 'Active Account'}
                    </span>
                </div>
                <div className="flex items-center gap-3">
                    <button className="px-4 py-2 bg-[#16202a] hover:bg-[#111a22] border border-[#233648] text-[#92adc9] hover:text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors">
                        <span className="material-symbols-outlined text-[16px]">build</span>
                        Troubleshoot Configs
                    </button>
                    <button 
                        onClick={handleImpersonate}
                        className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
                    >
                        <span className="material-symbols-outlined text-[16px]">admin_panel_settings</span>
                        Impersonate (Read-Only)
                    </button>
                </div>
            </header>

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    {/* Subscription Card */}
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-5xl text-primary">verified</span>
                        </div>
                        <div className="flex flex-col h-full justify-between">
                            <div>
                                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Subscription Plan</div>
                                <div className="text-2xl font-bold text-white mb-1">{summary?.plan || 'Enterprise Gold'}</div>
                                <div className="text-xs text-[#92adc9]">Next renewal: <span className="text-white">{summary?.renewalDate || 'Nov 15, 2024'}</span></div>
                            </div>
                            <div className="mt-4 pt-4 border-t border-[#233648] flex justify-between items-center">
                                <span className="text-[10px] text-emerald-500 bg-emerald-500/10 px-2 py-1 rounded border border-emerald-500/20 font-mono font-bold tracking-widest">PAID</span>
                                <span className="text-xs text-[#92adc9] font-mono">{summary?.orgId || 'ORG-8922'}</span>
                            </div>
                        </div>
                    </div>

                    {/* Devices Card */}
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-5xl text-white">devices</span>
                        </div>
                        <div className="flex flex-col h-full justify-between">
                            <div>
                                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Total Devices</div>
                                <div className="text-2xl font-bold text-white mb-1">{stats?.total?.toLocaleString() || '2,450'}</div>
                                <div className="flex items-center gap-3 mt-1">
                                    <span className="text-xs text-emerald-500 flex items-center gap-1">
                                        <span className="w-2 h-2 rounded-full bg-emerald-500"></span> {stats?.online?.toLocaleString() || '2,100'} Online
                                    </span>
                                    <span className="text-xs text-[#92adc9] flex items-center gap-1">
                                        <span className="w-2 h-2 rounded-full bg-[#92adc9]"></span> {stats?.offline?.toLocaleString() || '350'} Offline
                                    </span>
                                </div>
                            </div>
                            <div className="w-full bg-[#111a22] h-1.5 rounded-full mt-4 overflow-hidden flex">
                                <div className="bg-emerald-500 h-full" style={{ width: `${(stats?.online / stats?.total) * 100}%` }}></div>
                                <div className="bg-[#92adc9]/30 h-full" style={{ width: `${(stats?.offline / stats?.total) * 100}%` }}></div>
                            </div>
                        </div>
                    </div>

                    {/* Integration Health */}
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-5xl text-emerald-500">hub</span>
                        </div>
                        <div className="flex flex-col h-full justify-between">
                            <div>
                                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Integration Health</div>
                                <div className="text-2xl font-bold text-emerald-500 mb-1">{health?.successRate}%</div>
                                <div className="text-xs text-[#92adc9]">API Success Rate (24h)</div>
                            </div>
                            <div className="mt-4 flex gap-2">
                                <div className="flex-1 bg-[#111a22] rounded p-2 border border-[#233648]">
                                    <div className="text-[10px] text-[#92adc9] uppercase font-bold tracking-tighter">Latency</div>
                                    <div className="text-sm font-bold text-white">{health?.latencyMs}ms</div>
                                </div>
                                <div className="flex-1 bg-[#111a22] rounded p-2 border border-[#233648]">
                                    <div className="text-[10px] text-[#92adc9] uppercase font-bold tracking-tighter">Errors</div>
                                    <div className="text-sm font-bold text-rose-500">{health?.errors24h}</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Tabs */}
                <div className="mb-6">
                    <div className="border-b border-[#233648] flex gap-8 px-1">
                        {[
                            { id: 'fleet', icon: 'local_shipping', label: 'Fleet Overview' },
                            { id: 'logs', icon: 'history', label: 'Activity Logs' },
                            { id: 'config', icon: 'settings_applications', label: 'Technical Config' }
                        ].map(tab => (
                            <button 
                                key={tab.id}
                                onClick={() => setActiveTab(tab.id)}
                                className={`pb-3 px-1 text-sm font-semibold flex items-center gap-2 transition-all relative ${
                                    activeTab === tab.id ? 'text-white' : 'text-[#92adc9] hover:text-white'
                                }`}
                            >
                                <span className={`material-symbols-outlined text-[18px] ${activeTab === tab.id ? 'text-primary' : 'text-[#92adc9]'}`}>{tab.icon}</span>
                                {tab.label}
                                {activeTab === tab.id && <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary"></div>}
                            </button>
                        ))}
                    </div>
                </div>

                <div className="grid grid-cols-12 gap-6">
                    <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
                        {activeTab === 'fleet' && (
                            <div className="bg-[#16202a] rounded-xl border border-[#233648] h-[400px] flex flex-col overflow-hidden shadow-inner">
                                <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                                    <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                        <span className="material-symbols-outlined text-primary text-[18px]">map</span>
                                        Real-time Fleet Distribution
                                    </h3>
                                    <div className="flex gap-2 text-[#92adc9] text-[10px] font-bold">
                                         <span>LOAD: <span className="text-white">NOMINAL</span></span>
                                         <span className="w-px h-3 bg-[#233648] mx-1"></span>
                                         <span>NODES: <span className="text-white">{fleet?.regions?.length || 0}</span></span>
                                    </div>
                                </div>
                                <div className="relative flex-1 bg-[#0b1219] w-full h-full overflow-hidden">
                                     {/* Grid background for fleet distribution canvas */}
                                    <div className="absolute inset-0 opacity-10" style={{ backgroundImage: 'radial-gradient(#233648 1px, transparent 1px)', backgroundSize: '20px 20px' }}></div>
                                    
                                    {/* Simulation points */}
                                    <div className="absolute top-1/4 left-1/4 w-3 h-3 bg-emerald-500 rounded-full shadow-[0_0_15px_rgba(16,185,129,0.4)] animate-pulse"></div>
                                    <div className="absolute top-1/3 left-1/2 w-3 h-3 bg-emerald-500 rounded-full shadow-[0_0_15px_rgba(16,185,129,0.4)]"></div>
                                    <div className="absolute bottom-1/3 right-1/4 w-3 h-3 bg-amber-500 rounded-full shadow-[0_0_15px_rgba(245,158,11,0.4)]"></div>
                                    <div className="absolute top-1/2 right-1/3 w-3 h-3 bg-rose-500 rounded-full shadow-[0_0_15px_rgba(239,68,68,0.5)] animate-ping"></div>
                                    <div className="absolute top-1/2 right-1/3 w-3 h-3 bg-rose-500 rounded-full"></div>

                                    <div className="absolute bottom-4 left-4 bg-[#111a22]/90 backdrop-blur border border-[#233648] p-4 rounded-lg z-10">
                                        <div className="text-[10px] font-bold text-white mb-3 uppercase tracking-widest">Region Status</div>
                                        <div className="space-y-2">
                                            {fleet?.regions?.map((r: any) => (
                                                <div key={r.name} className="flex items-center justify-between gap-6 text-[10px]">
                                                    <div className="flex items-center gap-2 text-[#92adc9]">
                                                        <span className={`w-1.5 h-1.5 rounded-full ${r.activePercent > 90 ? 'bg-emerald-500' : 'bg-amber-500'}`}></span>
                                                        {r.name}
                                                    </div>
                                                    <span className="text-white font-bold">{r.activePercent}% Active</span>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}

                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden">
                            <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                                <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[#92adc9] text-[18px]">medical_services</span>
                                    Recent Diagnostic Events
                                </h3>
                                <button className="text-[10px] text-primary font-bold uppercase tracking-widest hover:underline transition-all">View Full Logs</button>
                            </div>
                            <div className="overflow-x-auto">
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="border-b border-[#233648] text-[10px] text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                                            <th className="px-4 py-3 font-bold">Time</th>
                                            <th className="px-4 py-3 font-bold">Device ID</th>
                                            <th className="px-4 py-3 font-bold">Issue Type</th>
                                            <th className="px-4 py-3 font-bold">Firmware</th>
                                            <th className="px-4 py-3 font-bold text-right">Status</th>
                                        </tr>
                                    </thead>
                                    <tbody className="text-xs">
                                        {diagnostics.map((event, idx) => (
                                            <tr key={idx} className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                                                <td className="px-4 py-4 text-[#92adc9] font-mono">{new Date(event.time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</td>
                                                <td className="px-4 py-4 font-bold text-white font-mono">{event.deviceId}</td>
                                                <td className="px-4 py-4 text-white font-medium">{event.issue}</td>
                                                <td className="px-4 py-4 text-[#92adc9]">{event.firmware}</td>
                                                <td className="px-4 py-4 text-right">
                                                    <span className={`px-2 py-0.5 rounded text-[9px] font-bold border ${
                                                        event.status === 'CRITICAL' ? 'bg-rose-500/10 text-rose-500 border-rose-500/20' :
                                                        event.status === 'INVESTIGATING' ? 'bg-amber-500/10 text-amber-500 border-amber-500/20' :
                                                        'bg-emerald-500/10 text-emerald-500 border-emerald-500/20'
                                                    }`}>
                                                        {event.status}
                                                    </span>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>

                    {/* Sidebar components */}
                    <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
                        {/* Alerts panel */}
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col max-h-[400px]">
                            <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                                <h3 className="text-white text-sm font-semibold">Active Alerts</h3>
                                <span className="bg-rose-500 text-white text-[10px] px-2 py-0.5 rounded-full font-bold shadow-lg shadow-rose-500/20 animate-pulse">{alerts.length} New</span>
                            </div>
                            <div className="flex-1 overflow-y-auto custom-scrollbar p-3 space-y-3">
                                {alerts.map(alert => (
                                    <div key={alert.alertId} className="bg-[#111a22] border border-[#233648] rounded p-3 flex gap-3 hover:border-primary/50 transition-all cursor-pointer group">
                                        <div className={`mt-0.5 ${alert.severity === 'CRITICAL' ? 'text-rose-500' : alert.severity === 'WARNING' ? 'text-amber-500' : 'text-primary'}`}>
                                            <span className="material-symbols-outlined text-[20px]">{alert.severity === 'CRITICAL' ? 'error' : 'warning_amber'}</span>
                                        </div>
                                        <div>
                                            <div className="text-sm font-semibold text-white group-hover:text-primary transition-colors">{alert.title}</div>
                                            <div className="text-[11px] text-[#92adc9] mt-1 leading-relaxed">{alert.desc}</div>
                                            <div className="text-[9px] text-[#92adc9]/60 mt-2 font-mono uppercase font-bold">{alert.alertId} • {alert.timeAgo}</div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Contacts */}
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col">
                            <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                                <h3 className="text-white text-sm font-semibold">Technical Org Contacts</h3>
                            </div>
                            <div className="p-2 space-y-1">
                                {contacts.map(contact => (
                                    <div key={contact.email} className="flex items-center justify-between p-3 hover:bg-[#111a22] rounded-lg transition-colors cursor-pointer group">
                                        <div className="flex items-center gap-3">
                                            <div className="w-9 h-9 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center text-xs font-bold text-primary group-hover:bg-primary group-hover:text-white transition-all">
                                                {contact.initials}
                                            </div>
                                            <div className="flex flex-col">
                                                <span className="text-sm font-semibold text-white">{contact.name}</span>
                                                <span className="text-[10px] text-[#92adc9] uppercase font-bold tracking-tighter">{contact.role} • {contact.email}</span>
                                            </div>
                                        </div>
                                        <button className="p-1.5 text-[#92adc9] hover:text-white hover:bg-[#233648] rounded transition-all">
                                            <span className="material-symbols-outlined text-[18px]">mail</span>
                                        </button>
                                    </div>
                                ))}
                            </div>
                            <div className="p-3 border-t border-[#233648] bg-[#111a22]/50">
                                <button className="w-full py-2 text-[10px] text-center text-primary hover:text-white font-bold uppercase tracking-widest border border-primary/20 hover:bg-primary/10 rounded transition-all">
                                    View Full Directory
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <footer className="h-10 bg-[#111a22] border-t border-[#233648] px-6 py-2 flex items-center justify-between text-[10px] text-[#92adc9] z-20 font-medium">
                <div className="flex items-center gap-4">
                    <span>© 2026 FleetAnalytics Premium. Support Hub v4.1</span>
                    <span className="w-1 h-1 rounded-full bg-[#233648]"></span>
                    <span>System Protocol: <span className="text-emerald-500 font-bold tracking-widest">STABLE</span></span>
                </div>
                <div className="flex items-center gap-6 font-bold uppercase tracking-widest">
                    <span>Latency: <span className="text-emerald-500">24ms</span></span>
                    <span>Gateway: <span className="text-white">US-East-Primary</span></span>
                    <span>Internal Time: <span className="font-mono text-white">{new Date().toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit' })} UTC</span></span>
                </div>
            </footer>
        </main>
    );
};

export default OrganizationHub;
