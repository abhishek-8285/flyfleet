import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const IncidentManagement: React.FC = () => {
    const navigate = useNavigate();
    const [metrics, setMetrics] = useState<any>(null);
    const [feed, setFeed] = useState<any[]>([]);
    const [incidents, setIncidents] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [activeTab, setActiveTab] = useState('daily');

    const fetchData = useCallback(async () => {
        const token = localStorage.getItem('access_token') || localStorage.getItem('token');
        const headers = { 'Authorization': `Bearer ${token}` };

        try {
            const [metricsRes, trendsRes, feedRes, listRes] = await Promise.all([
                fetch('/bff/support/incidents/metrics', { headers }),
                fetch('/bff/support/incidents/trends', { headers }),
                fetch('/bff/support/incidents/feed', { headers }),
                fetch('/bff/support/incidents', { headers })
            ]);

            if (metricsRes.ok) setMetrics(await metricsRes.json());
            if (trendsRes.ok) {
                // trends are currently mocked in SVG, but we fetch them to be ready
                await trendsRes.json();
            }
            if (feedRes.ok) setFeed(await feedRes.json());
            if (listRes.ok) {
                const data = await listRes.json();
                setIncidents(data.items || []);
            }
        } catch (err) {
            console.error('Failed to fetch incident data', err);
        } finally {
            setIsLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 30000);
        return () => clearInterval(interval);
    }, [fetchData]);

    if (isLoading) {
        return (
            <div className="flex-1 flex items-center justify-center bg-[#0b1219]">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                    <span className="text-[#92adc9] text-xs font-bold uppercase tracking-widest animate-pulse">Syncing Incident Nexus...</span>
                </div>
            </div>
        );
    }

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden font-display">
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <div className="flex items-center gap-2 text-[10px] text-[#92adc9] uppercase font-bold tracking-wider mb-0.5">
                            <span>Support</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-white">Incidents</span>
                        </div>
                        <h2 className="text-white text-lg font-bold uppercase tracking-tight">Incident Management</h2>
                    </div>
                </div>
                <div className="flex items-center gap-4">
                    <div className="relative">
                        <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[#92adc9] text-[18px]">search</span>
                        <input 
                            className="bg-[#16202a] border border-[#233648] text-xs text-white rounded pl-9 pr-3 py-2 w-64 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-[#92adc9]/60 font-medium" 
                            placeholder="Search incidents (ID, Subject, Org)..." 
                            type="text"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
                    <button className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20">
                        <span className="material-symbols-outlined text-[16px]">add</span>
                        Create Incident
                    </button>
                </div>
            </header>

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                {/* Metrics Grid */}
                <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-8">
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-[#137fec]/50 transition-all">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-white">inbox</span>
                        </div>
                        <div className="text-[#92adc9] text-[10px] font-bold uppercase mb-1 tracking-widest">Total Incidents</div>
                        <div className="text-2xl font-bold text-white">{metrics?.total || 0}</div>
                        <div className="text-[10px] text-emerald-500 flex items-center gap-1 mt-1 font-bold">
                            <span className="material-symbols-outlined text-[12px]">trending_up</span>
                            +{metrics?.newToday || 0} today
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-danger/50 transition-all">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-danger">priority_high</span>
                        </div>
                        <div className="text-[#92adc9] text-[10px] font-bold uppercase mb-1 tracking-widest">Critical Open</div>
                        <div className="text-2xl font-bold text-danger">{metrics?.criticalOpen || 0}</div>
                        <div className="text-[10px] text-[#92adc9] mt-1 font-medium">Avg resolution: 2h</div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-danger h-full shadow-[0_0_5px_rgba(239,68,68,0.5)]" style={{ width: '25%' }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-warning/50 transition-all">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-warning">warning</span>
                        </div>
                        <div className="text-[#92adc9] text-[10px] font-bold uppercase mb-1 tracking-widest">High Priority</div>
                        <div className="text-2xl font-bold text-warning">{metrics?.highPriority || 0}</div>
                        <div className="text-[10px] text-[#92adc9] mt-1 font-medium">Needs attention</div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-warning h-full shadow-[0_0_5px_rgba(245,158,11,0.5)]" style={{ width: '45%' }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-primary/50 transition-all">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-primary">engineering</span>
                        </div>
                        <div className="text-[#92adc9] text-[10px] font-bold uppercase mb-1 tracking-widest">In Progress</div>
                        <div className="text-2xl font-bold text-primary">{metrics?.inProgress || 0}</div>
                        <div className="text-[10px] text-[#92adc9] mt-1 font-medium">Active investigations</div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-primary h-full shadow-[0_0_5px_rgba(19,127,236,0.5)]" style={{ width: '60%' }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-emerald-500/50 transition-all">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-emerald-500">check_circle</span>
                        </div>
                        <div className="text-[#92adc9] text-[10px] font-bold uppercase mb-1 tracking-widest">Resolved (24h)</div>
                        <div className="text-2xl font-bold text-emerald-500">{metrics?.resolved24h || 0}</div>
                        <div className="text-[10px] text-emerald-500 flex items-center gap-1 mt-1 font-bold">
                            <span className="material-symbols-outlined text-[12px]">trending_up</span>
                            94% SLA met
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-[#92adc9]">schedule</span>
                        </div>
                        <div className="text-[#92adc9] text-[10px] font-bold uppercase mb-1 tracking-widest">Avg MTTR</div>
                        <div className="text-2xl font-bold text-white">{metrics?.mttr || '0h 0m'}</div>
                        <div className="text-[10px] text-emerald-500 mt-1 font-bold">-15m vs last week</div>
                    </div>
                </div>

                <div className="grid grid-cols-12 gap-6 mb-8">
                    {/* Chart Section */}
                    <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 h-[340px] flex flex-col relative shadow-lg">
                            <div className="flex justify-between items-center mb-6">
                                <h3 className="text-white text-sm font-bold flex items-center gap-2 uppercase tracking-wide">
                                    <span className="material-symbols-outlined text-primary text-[18px]">show_chart</span>
                                    Incident Volume (Last 30 Days)
                                </h3>
                                <div className="flex bg-[#111a22] p-0.5 rounded border border-[#233648]">
                                    <button 
                                        onClick={() => setActiveTab('daily')}
                                        className={`px-3 py-1 rounded text-[10px] font-bold uppercase tracking-widest transition-all ${activeTab === 'daily' ? 'bg-[#16202a] text-white shadow-sm' : 'text-[#92adc9] hover:text-white'}`}
                                    >
                                        Daily
                                    </button>
                                    <button 
                                        onClick={() => setActiveTab('weekly')}
                                        className={`px-3 py-1 rounded text-[10px] font-bold uppercase tracking-widest transition-all ${activeTab === 'weekly' ? 'bg-[#16202a] text-white shadow-sm' : 'text-[#92adc9] hover:text-white'}`}
                                    >
                                        Weekly
                                    </button>
                                </div>
                            </div>
                            <div className="flex-1 relative w-full h-full overflow-hidden">
                                <svg className="w-full h-full" preserveAspectRatio="none" viewBox="0 0 800 200">
                                    <defs>
                                        <linearGradient id="chartGradient" x1="0" x2="0" y1="0" y2="1">
                                            <stop offset="0%" stopColor="#137fec" stopOpacity="0.25"></stop>
                                            <stop offset="100%" stopColor="#137fec" stopOpacity="0"></stop>
                                        </linearGradient>
                                    </defs>
                                    {/* Grid Lines */}
                                    <line stroke="#233648" strokeWidth="1" x1="0" x2="800" y1="199" y2="199"></line>
                                    <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="150" y2="150"></line>
                                    <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="100" y2="100"></line>
                                    <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="50" y2="50"></line>
                                    
                                    {/* Area Path */}
                                    <path d="M0,140 C100,160 200,100 300,120 S500,60 600,40 S700,80 800,50 V200 H0 Z" fill="url(#chartGradient)"></path>
                                    {/* Line Path */}
                                    <path d="M0,140 C100,160 200,100 300,120 S500,60 600,40 S700,80 800,50" fill="none" stroke="#137fec" strokeWidth="2.5" strokeLinecap="round"></path>
                                    
                                    {/* Point */}
                                    <circle cx="600" cy="40" fill="#fff" r="4.5" stroke="#137fec" strokeWidth="2.5" className="animate-pulse"></circle>
                                </svg>
                                <div className="absolute bottom-1 left-0 right-0 flex justify-between text-[9px] text-[#92adc9] px-2 font-bold uppercase tracking-widest opacity-60">
                                    <span>Oct 1</span>
                                    <span>Oct 8</span>
                                    <span>Oct 15</span>
                                    <span>Oct 22</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Feed Section */}
                    <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col h-[340px] shadow-lg">
                            <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]/50">
                                <h3 className="text-white text-[11px] font-bold uppercase tracking-widest">Incident Feed</h3>
                                <button className="text-primary text-[10px] font-bold uppercase tracking-widest hover:underline transition-all">View All</button>
                            </div>
                            <div className="flex-1 overflow-y-auto custom-scrollbar p-0">
                                <div className="divide-y divide-[#233648]">
                                    {feed.map((item, idx) => (
                                        <div key={idx} className="flex items-start gap-3 p-4 hover:bg-[#111a22] transition-colors cursor-pointer group">
                                            <div className={`mt-1 w-2 h-2 rounded-full flex-shrink-0 shadow-[0_0_8px] ${
                                                item.priority === 'CRITICAL' ? 'bg-danger shadow-danger/40' : 
                                                item.priority === 'HIGH' ? 'bg-warning shadow-warning/40' : 'bg-primary shadow-primary/40'
                                            }`}></div>
                                            <div className="flex flex-col min-w-0 flex-1">
                                                <div className="flex justify-between items-center mb-1">
                                                    <span className="text-[10px] font-bold text-white group-hover:text-primary transition-colors">{item.incidentId}</span>
                                                    <span className="text-[10px] text-[#92adc9] font-medium">{item.timeAgo} ago</span>
                                                </div>
                                                <p className="text-xs text-[#92adc9] truncate font-medium leading-tight mb-2 group-hover:text-white transition-colors">{item.subject}</p>
                                                <div className="flex items-center gap-2">
                                                    <span className="text-[9px] bg-[#111a22] px-1.5 py-0.5 rounded text-[#92adc9] border border-[#233648] font-bold uppercase tracking-tighter">
                                                        {item.priority}
                                                    </span>
                                                    <span className="text-[9px] text-primary font-bold uppercase tracking-tighter">{item.status}</span>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Table Section */}
                <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden shadow-2xl">
                    <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                        <h3 className="text-white font-bold text-sm flex items-center gap-2 uppercase tracking-wide">
                            <span className="material-symbols-outlined text-[#92adc9] text-[18px]">list_alt</span>
                            Active Incidents
                        </h3>
                        <div className="flex items-center gap-2">
                            <button className="flex items-center gap-1.5 px-3 py-1.5 bg-[#16202a] border border-[#233648] rounded text-[10px] font-bold text-[#92adc9] uppercase tracking-widest hover:text-white hover:border-[#92adc9] transition-all">
                                <span className="material-symbols-outlined text-[16px]">filter_list</span>
                                Filter
                            </button>
                            <button className="flex items-center gap-1.5 px-3 py-1.5 bg-[#16202a] border border-[#233648] rounded text-[10px] font-bold text-[#92adc9] uppercase tracking-widest hover:text-white hover:border-[#92adc9] transition-all">
                                <span className="material-symbols-outlined text-[16px]">download</span>
                                Export
                            </button>
                        </div>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-left border-collapse">
                            <thead>
                                <tr className="border-b border-[#233648] text-[10px] text-[#92adc9] uppercase tracking-widest bg-[#111a22]/50 font-bold">
                                    <th className="px-4 py-4 w-28">Incident ID</th>
                                    <th className="px-4 py-4 w-24 text-center">Priority</th>
                                    <th className="px-4 py-4 w-28">Status</th>
                                    <th className="px-4 py-4">Subject</th>
                                    <th className="px-4 py-4">Assigned To</th>
                                    <th className="px-4 py-4">Organization</th>
                                    <th className="px-4 py-4 text-right">Created</th>
                                </tr>
                            </thead>
                            <tbody className="text-xs">
                                {incidents.filter(inc => inc.incidentId.toLowerCase().includes(searchQuery.toLowerCase()) || inc.subject.toLowerCase().includes(searchQuery.toLowerCase())).map((incident) => (
                                    <tr 
                                        key={incident.incidentId} 
                                        onClick={() => navigate(`/support/incidents/${incident.incidentId}`)}
                                        className="border-b border-[#233648] hover:bg-[#111a22]/70 transition-all cursor-pointer group"
                                    >
                                        <td className="px-4 py-4 font-mono text-[11px] text-primary font-bold group-hover:underline transition-all">
                                            {incident.incidentId}
                                        </td>
                                        <td className="px-4 py-4 text-center">
                                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-[9px] font-black border uppercase tracking-tighter ${
                                                incident.priority === 'CRITICAL' ? 'bg-danger/10 text-danger border-danger/20' : 
                                                incident.priority === 'HIGH' ? 'bg-warning/10 text-warning border-warning/20' : 
                                                'bg-[#92adc9]/10 text-[#92adc9] border-[#233648]'
                                            }`}>
                                                {incident.priority}
                                            </span>
                                        </td>
                                        <td className="px-4 py-4 font-bold">
                                            <span className="flex items-center gap-1.5 text-white text-[10px] uppercase font-black">
                                                <span className={`w-1.5 h-1.5 rounded-full ${incident.status === 'OPEN' ? 'bg-danger animate-pulse' : incident.status === 'RESOLVED' ? 'bg-emerald-500' : 'bg-primary'}`}></span>
                                                {incident.status}
                                            </span>
                                        </td>
                                        <td className="px-4 py-4 text-white font-semibold group-hover:text-primary transition-colors text-xs leading-snug">
                                            {incident.subject}
                                        </td>
                                        <td className="px-4 py-4 text-[#92adc9] text-[10px] font-bold uppercase tracking-tighter">
                                            <div className="flex items-center gap-2">
                                                <div className={`w-6 h-6 rounded-full flex items-center justify-center text-[8px] text-white font-black ${incident.assigneeInitials === 'UN' ? 'bg-[#233648]' : 'bg-indigo-600'}`}>
                                                    {incident.assigneeInitials}
                                                </div>
                                                <span>{incident.assignedTo}</span>
                                            </div>
                                        </td>
                                        <td className="px-4 py-4 text-[#92adc9] text-[11px] font-medium">{incident.organization}</td>
                                        <td className="px-4 py-4 text-right text-[#92adc9] text-[10px] font-mono font-bold uppercase italic opacity-80">
                                            {new Date(incident.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            <footer className="h-10 bg-[#111a22] border-t border-[#233648] px-6 flex items-center justify-between text-[10px] font-bold uppercase tracking-widest text-[#92adc9]/60 z-20">
                <div className="flex items-center gap-4">
                    <span>© 2026 FleetSupport Incident Nexus. v2.4-STABLE</span>
                    <span className="w-1 h-1 rounded-full bg-[#233648]"></span>
                    <span>System Status: <span className="text-emerald-500">OPERATIONAL</span></span>
                </div>
                <div className="flex items-center gap-6">
                    <span>API: <span className="text-white">Active</span></span>
                    <span>MTTR Latency: <span className="text-white">42ms</span></span>
                    <span>Gateway: <span className="text-white">US-East-Primary</span></span>
                </div>
            </footer>
        </main>
    );
};

export default IncidentManagement;
