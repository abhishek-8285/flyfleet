import React, { useEffect, useState } from 'react';
import NewTicketModal from '../../components/support/NewTicketModal';

const toNumber = (value: any): number => {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
};

const SupportDashboard: React.FC = () => {
    const [metrics, setMetrics] = useState<any>(null);
    const [trendPoints, setTrendPoints] = useState<Array<{ label: string; value: number }>>([]);
    const [topErrors, setTopErrors] = useState<any[]>([]);
    const [recentTickets, setRecentTickets] = useState<any[]>([]);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [autoRefresh, setAutoRefresh] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');

    const fetchData = async () => {
        const token = localStorage.getItem('access_token');
        const headers = {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        };

        try {
            const [metricsRes, trendsRes, errorsRes, ticketsRes] = await Promise.all([
                fetch('/bff/support/dashboard/metrics', { headers }),
                fetch('/bff/support/dashboard/trends', { headers }),
                fetch('/bff/support/dashboard/top-errors', { headers }),
                fetch('/bff/support/tickets/recent', { headers }),
                fetch('/bff/support/dashboard/live', { headers })
            ]);

            if (metricsRes.ok) setMetrics(await metricsRes.json());
            if (trendsRes.ok) {
                const trendPayload = await trendsRes.json();
                const points = Array.isArray(trendPayload?.points) ? trendPayload.points : [];
                setTrendPoints(
                    points.map((point: any, index: number) => ({
                        label: String(point?.day ?? point?.label ?? `D${index + 1}`),
                        value: toNumber(point?.count ?? point?.value),
                    }))
                );
            }
            if (errorsRes.ok) setTopErrors(await errorsRes.json());
            if (ticketsRes.ok) {
                const data = await ticketsRes.json();
                setRecentTickets(data.tickets || []);
            }
        } catch (err) {
            console.error('Failed to fetch support dashboard data', err);
        }
    };

    useEffect(() => {
        fetchData();
        let interval: any;
        if (autoRefresh) {
            interval = setInterval(fetchData, 30000);
        }
        return () => clearInterval(interval);
    }, [autoRefresh]);

    const chartPath = (() => {
        if (!trendPoints.length) return '';
        const maxVal = Math.max(...trendPoints.map((p) => p.value), 1);
        const minVal = Math.min(...trendPoints.map((p) => p.value), 0);
        const range = Math.max(maxVal - minVal, 1);
        return trendPoints
            .map((point, index) => {
                const x = (index / Math.max(trendPoints.length - 1, 1)) * 800;
                const y = 190 - ((point.value - minVal) / range) * 150;
                return `${index === 0 ? 'M' : 'L'}${x},${y}`;
            })
            .join(' ');
    })();

    const openTickets = toNumber(metrics?.openTickets ?? metrics?.open_tickets?.count);
    const slaBreaches = toNumber(metrics?.slaBreaches24h ?? metrics?.sla_breaches?.count);
    const activeErrors = toNumber(metrics?.activeDeviceErrors ?? metrics?.active_errors?.count);
    const openTicketsChange = toNumber(metrics?.trends?.openTicketsChange ?? metrics?.open_tickets?.trend);
    const resolutionTrend = toNumber(metrics?.trends?.resolutionTimeChange ?? metrics?.avg_resolution?.trend);

    const filteredTickets = recentTickets.filter((ticket) => {
        const q = searchQuery.toLowerCase();
        if (!q) return true;
        return (
            String(ticket.ticketId || '').toLowerCase().includes(q) ||
            String(ticket.subject || '').toLowerCase().includes(q) ||
            String(ticket.reporter || '').toLowerCase().includes(q)
        );
    });

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden font-display">
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <h2 className="text-white text-lg font-bold">Support Dashboard</h2>
                        <div className="flex items-center gap-2 mt-0.5">
                            <span className="text-[11px] text-[#92adc9] uppercase tracking-wider font-semibold">Support</span>
                            <span className="text-[11px] text-[#92adc9]">/</span>
                            <span className="text-[11px] text-white uppercase tracking-wider font-semibold">Dashboard</span>
                        </div>
                    </div>
                    <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-emerald-500/10 text-emerald-500 border border-emerald-500/30">
                        Live Data
                    </span>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-3 px-3 py-1.5 rounded bg-[#16202a] border border-[#233648]">
                        <span className="text-xs text-[#92adc9] font-medium">Auto-Refresh</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                            <input 
                                type="checkbox" 
                                checked={autoRefresh} 
                                onChange={(e) => setAutoRefresh(e.target.checked)} 
                                className="sr-only peer" 
                            />
                            <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary"></div>
                        </label>
                    </div>
                    <button 
                        onClick={() => setIsModalOpen(true)}
                        className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
                    >
                        <span className="material-symbols-outlined text-[16px]">add</span>
                        New Ticket
                    </button>
                </div>
            </header>

            <NewTicketModal 
                isOpen={isModalOpen} 
                onClose={() => setIsModalOpen(false)} 
                onCreated={() => fetchData()} 
            />

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                {/* Stats Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-white">inbox</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Open Tickets</div>
                        <div className="text-2xl font-bold text-white">{openTickets}</div>
                        <div className="text-[10px] text-emerald-500 flex items-center gap-1 mt-1">
                            <span className="material-symbols-outlined text-[12px]">trending_down</span>
                            {openTicketsChange}% vs yesterday
                        </div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-white h-full" style={{ width: `${Math.max(5, Math.min(100, openTickets))}%` }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-rose-500">timer_off</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">SLA Breaches (24h)</div>
                        <div className="text-2xl font-bold text-rose-500">{slaBreaches}</div>
                        <div className="text-[10px] text-rose-500 mt-1">Critical Priority</div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-rose-500 h-full" style={{ width: `${Math.max(5, Math.min(100, slaBreaches * 8))}%` }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-amber-500">error</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Active Device Errors</div>
                        <div className="text-2xl font-bold text-amber-500">{activeErrors}</div>
                        <div className="text-[10px] text-[#92adc9] mt-1">Across fleet</div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-amber-500 h-full" style={{ width: `${Math.max(5, Math.min(100, activeErrors * 4))}%` }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-emerald-500">av_timer</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Avg. Resolution Time</div>
                        <div className="text-2xl font-bold text-emerald-500">{metrics?.avgResolutionTime || '0.0h'}</div>
                        <div className="text-[10px] text-emerald-500 flex items-center gap-1 mt-1">
                            <span className="material-symbols-outlined text-[12px]">trending_down</span>
                            {Math.abs(resolutionTrend)}% faster this week
                        </div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-emerald-500 h-full" style={{ width: '85%' }}></div>
                        </div>
                    </div>
                </div>

                <div className="grid grid-cols-12 gap-6 mb-6">
                    <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 h-80 flex flex-col">
                            <div className="flex justify-between items-center mb-4">
                                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                    <span className="material-symbols-outlined text-primary text-[18px]">show_chart</span>
                                    Incident Trends (Last 7 Days)
                                </h3>
                                <select className="bg-[#111a22] text-xs text-[#92adc9] border border-[#233648] rounded px-2 py-1">
                                    <option>Last 7 Days</option>
                                    <option>Last 30 Days</option>
                                </select>
                            </div>
                            <div className="flex-1 relative w-full h-full">
                                <svg className="w-full h-full" preserveAspectRatio="none" viewBox="0 0 800 200">
                                    <defs>
                                        <linearGradient id="chartGradient" x1="0" x2="0" y1="0" y2="1">
                                            <stop offset="0%" stopColor="#137fec" stopOpacity="0.3"></stop>
                                            <stop offset="100%" stopColor="#137fec" stopOpacity="0"></stop>
                                        </linearGradient>
                                    </defs>
                                    <line stroke="#233648" strokeWidth="1" x1="0" x2="800" y1="199" y2="199"></line>
                                    <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="150" y2="150"></line>
                                    <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="100" y2="100"></line>
                                    <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="50" y2="50"></line>
                                    {chartPath && (
                                        <>
                                            <path d={`${chartPath} L800,200 L0,200 Z`} fill="url(#chartGradient)"></path>
                                            <path d={chartPath} fill="none" stroke="#137fec" strokeWidth="2"></path>
                                        </>
                                    )}
                                </svg>
                                <div className="absolute bottom-2 left-0 right-0 flex justify-between text-[10px] text-[#92adc9] px-2">
                                    {(trendPoints.length ? trendPoints : [{ label: 'Mon' }, { label: 'Tue' }, { label: 'Wed' }, { label: 'Thu' }, { label: 'Fri' }, { label: 'Sat' }, { label: 'Sun' }]).map((d: any, idx: number) => <span key={`${d.label}-${idx}`}>{d.label}</span>)}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col h-80">
                            <div className="p-4 border-b border-[#233648] flex justify-between items-center">
                                <h3 className="text-white text-sm font-semibold">Top System Errors</h3>
                                <button className="text-primary text-xs hover:underline">View All</button>
                            </div>
                            <div className="flex-1 p-4 flex flex-col justify-center gap-4">
                                {topErrors.slice(0, 4).map((error, idx) => (
                                    <div key={idx} className="w-full">
                                        <div className="flex justify-between text-xs mb-1">
                                            <span className="text-white">{error.title} ({error.errorCode})</span>
                                            <span className={`${error.severity === 'CRITICAL' ? 'text-rose-500' : 'text-amber-500'} font-bold`}>{error.count}</span>
                                        </div>
                                        <div className="w-full bg-[#111a22] h-2 rounded-full overflow-hidden">
                                            <div 
                                                className={`${error.severity === 'CRITICAL' ? 'bg-rose-500' : 'bg-amber-500'} h-full rounded-full`} 
                                                style={{ width: `${(error.count / (topErrors[0]?.count || 1)) * 100}%` }}
                                            ></div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>

                <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden">
                    <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                        <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                            <span className="material-symbols-outlined text-[#92adc9] text-[18px]">history</span>
                            Recent Support Activity
                        </h3>
                        <div className="flex items-center gap-2">
                            <input 
                                className="bg-[#16202a] border border-[#233648] text-xs text-white rounded px-3 py-1.5 w-64 focus:ring-1 focus:ring-primary focus:border-primary outline-none" 
                                placeholder="Search tickets, IDs..." 
                                type="text"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                            <button className="p-1.5 hover:bg-[#233648] rounded text-[#92adc9] hover:text-white">
                                <span className="material-symbols-outlined text-[18px]">filter_list</span>
                            </button>
                            <button className="p-1.5 hover:bg-[#233648] rounded text-[#92adc9] hover:text-white">
                                <span className="material-symbols-outlined text-[18px]">download</span>
                            </button>
                        </div>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-left border-collapse">
                            <thead>
                                <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                                    <th className="px-4 py-3 font-semibold">Ticket ID</th>
                                    <th className="px-4 py-3 font-semibold">Subject</th>
                                    <th className="px-4 py-3 font-semibold">Reporter</th>
                                    <th className="px-4 py-3 font-semibold">Priority</th>
                                    <th className="px-4 py-3 font-semibold">Assigned To</th>
                                    <th className="px-4 py-3 font-semibold">Last Updated</th>
                                    <th className="px-4 py-3 font-semibold text-right">Status</th>
                                </tr>
                            </thead>
                            <tbody className="text-sm">
                                {filteredTickets.map((ticket, idx) => (
                                    <tr key={idx} className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                                        <td className="px-4 py-3 text-white font-mono text-xs">{ticket.ticketId}</td>
                                        <td className="px-4 py-3 text-white">{ticket.subject}</td>
                                        <td className="px-4 py-3 text-[#92adc9]">{ticket.reporter}</td>
                                        <td className="px-4 py-3">
                                            <span className={`font-medium text-xs ${ticket.priority === 'CRITICAL' || ticket.priority === 'HIGH' ? 'text-rose-500' : ticket.priority === 'MEDIUM' ? 'text-amber-500' : 'text-[#92adc9]'}`}>
                                                {ticket.priority || 'Medium'}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3 text-[#92adc9]">
                                            <div className="flex items-center gap-2">
                                                <div className="w-5 h-5 rounded-full bg-primary/20 text-primary flex items-center justify-center text-[10px] font-bold">
                                                    {ticket.assignedTo?.split(' ').map((n: string) => n[0]).join('') || 'JD'}
                                                </div>
                                                <span>{ticket.assignedTo || 'John D.'}</span>
                                            </div>
                                        </td>
                                        <td className="px-4 py-3 text-[#92adc9] text-xs">
                                            {new Date(ticket.updatedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                        </td>
                                        <td className="px-4 py-3 text-right">
                                            <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold border ${
                                                ticket.status === 'RESOLVED' ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' :
                                                ticket.status === 'IN_PROGRESS' ? 'bg-primary/10 text-primary border-primary/20' :
                                                ticket.status === 'PENDING' ? 'bg-amber-500/10 text-amber-500 border-amber-500/20' :
                                                'bg-[#111a22] text-[#92adc9] border border-[#233648]'
                                            }`}>
                                                {ticket.status}
                                            </span>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            <footer className="bg-[#111a22] border-t border-[#233648] px-6 py-2 flex items-center justify-between text-[10px] text-[#92adc9] z-20">
                <span>© 2026 Enterprise Fleet Manager. Support Module V2.1</span>
                <div className="flex items-center gap-4">
                    <span>API Status: <span className="text-emerald-500">Connected</span></span>
                    <span>Support Agent: <span className="text-white">Active</span></span>
                    <span>Server Time: <span className="font-mono text-white">{new Date().toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit' })} UTC</span></span>
                </div>
            </footer>
        </main>
    );
};

export default SupportDashboard;
