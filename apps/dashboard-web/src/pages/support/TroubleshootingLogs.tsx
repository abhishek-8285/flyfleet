import React, { useCallback, useEffect, useRef, useState } from 'react';

const TroubleshootingLogs: React.FC = () => {
    const [logs, setLogs] = useState<any[]>([]);
    const [filters, setFilters] = useState<any>(null);
    const [selectedLog, setSelectedLog] = useState<any>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [isLive, setIsLive] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [moduleFilter, setModuleFilter] = useState('All Modules');
    const [severityFilter, setSeverityFilter] = useState('All Severities');
    const logsEndRef = useRef<HTMLDivElement>(null);

    const fetchLogs = useCallback(async () => {
        if (!isLive && logs.length > 0) return;

        const token = localStorage.getItem('access_token');
        const headers = { 'Authorization': `Bearer ${token}` };

        try {
            const [logsRes, filterRes] = await Promise.all([
                fetch('/bff/support/logs', { headers }),
                fetch('/bff/support/logs/filters', { headers })
            ]);

            if (logsRes.ok) {
                const data = await logsRes.json();
                setLogs(data.items || []);
            }
            if (filterRes.ok) setFilters(await filterRes.json());
        } catch (err) {
            console.error('Failed to fetch logs', err);
        } finally {
            setIsLoading(false);
        }
    }, [isLive, logs.length]);

    const fetchLogDetail = async (traceId: string) => {
        const token = localStorage.getItem('access_token');
        try {
            const res = await fetch(`/bff/support/logs/${traceId}`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (res.ok) setSelectedLog(await res.json());
        } catch (err) {
            console.error('Failed to fetch log detail', err);
        }
    };

    useEffect(() => {
        fetchLogs();
        const interval = setInterval(() => {
            if (isLive) fetchLogs();
        }, 5000);
        return () => clearInterval(interval);
    }, [fetchLogs, isLive]);

    if (isLoading) {
        return (
            <div className="flex-1 flex items-center justify-center bg-[#0b1219]">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                    <span className="text-[#92adc9] text-xs font-bold uppercase tracking-widest animate-pulse">Initializing Log Stream...</span>
                </div>
            </div>
        );
    }

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden font-display">
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <div className="flex items-center gap-2 text-[11px] text-[#92adc9] uppercase tracking-wider font-semibold mb-0.5">
                            <span>Support</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-primary">Troubleshooting Logs</span>
                        </div>
                        <h2 className="text-white text-lg font-bold">Consolidated Debugging Logs</h2>
                    </div>
                    <span className="ml-2 px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-[#16202a] text-[#92adc9] border border-[#233648] animate-pulse">
                        {isLive ? 'Live Stream' : 'Paused'}
                    </span>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex items-center bg-[#16202a] border border-[#233648] rounded-lg p-0.5 h-9">
                        <button 
                            onClick={() => setIsLive(true)}
                            className={`px-3 h-full text-xs font-medium rounded-md transition-all ${isLive ? 'bg-[#111a22] text-white shadow-sm border border-[#233648]' : 'text-[#92adc9] hover:text-white'}`}
                        >
                            Live
                        </button>
                        <button 
                            onClick={() => setIsLive(false)}
                            className={`px-3 h-full text-xs font-medium rounded-md transition-all ${!isLive ? 'bg-[#111a22] text-white shadow-sm border border-[#233648]' : 'text-[#92adc9] hover:text-white'}`}
                        >
                            Paused
                        </button>
                    </div>
                    <button className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20">
                        <span className="material-symbols-outlined text-[16px]">file_download</span>
                        Export Evidence
                    </button>
                </div>
            </header>

            <div className="flex flex-1 overflow-hidden">
                <div className="flex-1 flex flex-col min-w-0 border-r border-[#233648]">
                    <div className="p-4 border-b border-[#233648] bg-[#16202a] flex gap-3 items-center">
                        <div className="relative flex-1 max-w-md">
                            <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-[#92adc9]">
                                <span className="material-symbols-outlined text-[18px]">search</span>
                            </span>
                            <input 
                                className="w-full bg-[#111a22] border border-[#233648] text-xs text-white rounded pl-9 pr-3 py-2 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all" 
                                placeholder="Search by message, trace ID, or user..." 
                                type="text"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
                        <div className="h-8 w-[1px] bg-[#233648] mx-1"></div>
                        <select 
                            className="bg-[#111a22] text-xs text-[#92adc9] border border-[#233648] rounded px-3 py-2 outline-none focus:border-primary transition-all cursor-pointer"
                            value={moduleFilter}
                            onChange={(e) => setModuleFilter(e.target.value)}
                        >
                            <option>All Modules</option>
                            {filters?.modules?.map((m: string) => <option key={m}>{m}</option>)}
                        </select>
                        <select 
                            className="bg-[#111a22] text-xs text-[#92adc9] border border-[#233648] rounded px-3 py-2 outline-none focus:border-primary transition-all cursor-pointer"
                            value={severityFilter}
                            onChange={(e) => setSeverityFilter(e.target.value)}
                        >
                            <option>All Severities</option>
                            {filters?.severities?.map((s: string) => <option key={s}>{s}</option>)}
                        </select>
                        <div className="ml-auto flex items-center gap-2">
                            <span className="text-[10px] text-[#92adc9] font-bold uppercase tracking-tighter">Showing {logs.length} Recent Logs</span>
                        </div>
                    </div>

                    <div className="flex-1 overflow-y-auto custom-scrollbar bg-[#111a22]">
                        <table className="w-full text-left border-collapse table-fixed">
                            <thead className="sticky top-0 z-10 bg-[#16202a] shadow-sm">
                                <tr className="border-b border-[#233648] text-[11px] text-[#92adc9] uppercase tracking-wider font-bold">
                                    <th className="px-4 py-2 w-48">Timestamp</th>
                                    <th className="px-4 py-2 w-24">Severity</th>
                                    <th className="px-4 py-2 w-36">Module</th>
                                    <th className="px-4 py-2 w-auto">Message</th>
                                    <th className="px-4 py-2 w-32">Trace ID</th>
                                </tr>
                            </thead>
                            <tbody className="text-[12px] font-mono divide-y divide-[#233648]/40">
                                {logs.filter(log => 
                                    (moduleFilter === 'All Modules' || log.module === moduleFilter) &&
                                    (severityFilter === 'All Severities' || log.severity === severityFilter) &&
                                    (log.message.toLowerCase().includes(searchQuery.toLowerCase()) || log.traceId.toLowerCase().includes(searchQuery.toLowerCase()))
                                ).map((log, idx) => (
                                    <tr 
                                        key={idx} 
                                        onClick={() => fetchLogDetail(log.traceId)}
                                        className={`cursor-pointer transition-colors border-l-2 ${
                                            log.severity === 'ERROR' ? 'bg-rose-500/5 hover:bg-rose-500/10 border-l-rose-500' :
                                            log.severity === 'WARN' ? 'bg-amber-500/5 hover:bg-amber-500/10 border-l-amber-500' :
                                            'hover:bg-[#16202a] border-l-transparent hover:border-l-[#233648]'
                                        }`}
                                    >
                                        <td className="px-4 py-2 text-[#92adc9] whitespace-nowrap">{log.timestamp}</td>
                                        <td className="px-4 py-2">
                                            <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold border ${
                                                log.severity === 'ERROR' ? 'bg-rose-500/20 text-rose-500 border-rose-500/30' :
                                                log.severity === 'WARN' ? 'bg-amber-500/20 text-amber-500 border-amber-500/30' :
                                                'bg-[#0ea5e9]/10 text-[#0ea5e9] border-[#0ea5e9]/20'
                                            }`}>{log.severity}</span>
                                        </td>
                                        <td className="px-4 py-2 text-white font-semibold">{log.module}</td>
                                        <td className="px-4 py-2 text-[#e2e8f0] truncate italic">{log.message}</td>
                                        <td className="px-4 py-2 text-primary hover:underline font-bold">{log.traceId}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>

                {selectedLog && (
                    <div className="w-[450px] bg-[#16202a] border-l border-[#233648] flex flex-col flex-shrink-0 z-20 shadow-2xl animate-in slide-in-from-right duration-300">
                        <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                            <h3 className="font-bold text-white flex items-center gap-2 text-sm uppercase tracking-widest">
                                <span className="material-symbols-outlined text-[#92adc9] text-[18px]">find_in_page</span>
                                Log Detail
                            </h3>
                            <div className="flex gap-2">
                                <button className="text-[#92adc9] hover:text-white p-1 hover:bg-[#233648] rounded transition-all">
                                    <span className="material-symbols-outlined text-[18px]">content_copy</span>
                                </button>
                                <button 
                                    onClick={() => setSelectedLog(null)}
                                    className="text-[#92adc9] hover:text-white p-1 hover:bg-[#233648] rounded transition-all"
                                >
                                    <span className="material-symbols-outlined text-[18px]">close</span>
                                </button>
                            </div>
                        </div>
                        <div className="flex-1 overflow-y-auto custom-scrollbar p-6 space-y-8">
                            <div>
                                <div className="flex items-center justify-between mb-3">
                                    <span className="text-[11px] font-mono text-[#92adc9] font-bold">{selectedLog.traceId}</span>
                                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-black uppercase border ${
                                        selectedLog.severity === 'ERROR' ? 'bg-rose-500/20 text-rose-500 border-rose-500/30' : 'bg-amber-500/20 text-amber-500 border-amber-500/30'
                                    }`}>{selectedLog.severity}</span>
                                </div>
                                <h4 className="text-white font-bold text-base leading-tight">{selectedLog.message}</h4>
                                <div className="mt-3 flex items-center gap-2 text-[11px] text-[#92adc9] font-mono">
                                    <span className="material-symbols-outlined text-[14px]">schedule</span>
                                    {selectedLog.timestamp} UTC
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-4 text-xs">
                                <div className="p-3 rounded bg-[#111a22] border border-[#233648]">
                                    <div className="text-[#92adc9] mb-1 font-bold uppercase text-[9px] tracking-widest">Service</div>
                                    <div className="text-white font-bold">{selectedLog.service}</div>
                                </div>
                                <div className="p-3 rounded bg-[#111a22] border border-[#233648]">
                                    <div className="text-[#92adc9] mb-1 font-bold uppercase text-[9px] tracking-widest">Environment</div>
                                    <div className="text-emerald-500 font-bold">{selectedLog.environment}</div>
                                </div>
                                <div className="p-3 rounded bg-[#111a22] border border-[#233648]">
                                    <div className="text-[#92adc9] mb-1 font-bold uppercase text-[9px] tracking-widest">User ID</div>
                                    <div className="text-primary hover:underline cursor-pointer font-bold">{selectedLog.userId}</div>
                                </div>
                                <div className="p-3 rounded bg-[#111a22] border border-[#233648]">
                                    <div className="text-[#92adc9] mb-1 font-bold uppercase text-[9px] tracking-widest">Client IP</div>
                                    <div className="text-white font-mono font-bold">{selectedLog.ip}</div>
                                </div>
                            </div>

                            <div>
                                <div className="flex items-center justify-between mb-3">
                                    <h5 className="text-[10px] font-black uppercase tracking-widest text-[#92adc9]">Stack Trace</h5>
                                    <button className="text-[10px] text-primary font-bold hover:underline transition-all uppercase">Copy Trace</button>
                                </div>
                                <div className="bg-[#0b1219] rounded border border-[#233648] p-4 overflow-x-auto shadow-inner">
                                    <pre className="text-[11px] font-mono text-[#cbd5e1] whitespace-pre leading-relaxed">{selectedLog.stackTrace}</pre>
                                </div>
                            </div>

                            <div>
                                <h5 className="text-[10px] font-black uppercase tracking-widest text-[#92adc9] mb-3">Request Context</h5>
                                <div className="bg-[#111a22] rounded border border-[#233648] p-4 shadow-inner">
                                    <div className="flex gap-2 mb-3">
                                        <span className="text-[10px] font-black bg-primary/20 text-primary px-2 py-0.5 rounded border border-primary/30 uppercase">{selectedLog.requestContext?.method}</span>
                                        <span className="text-[10px] font-mono text-white font-bold">{selectedLog.requestContext?.url}</span>
                                    </div>
                                    <pre className="text-[10px] font-mono text-[#92adc9] overflow-x-auto whitespace-pre-wrap leading-relaxed">
                                        {JSON.stringify(selectedLog.requestContext?.headers, null, 2)}
                                    </pre>
                                </div>
                            </div>
                        </div>
                        <div className="p-4 border-t border-[#233648] bg-[#111a22]">
                            <button className="w-full py-2.5 bg-[#16202a] hover:bg-[#1a2530] border border-[#233648] text-white text-xs font-bold rounded transition-all flex justify-center items-center gap-2 uppercase tracking-wide group">
                                <span className="material-symbols-outlined text-[18px] text-primary group-hover:scale-110 transition-transform">bug_report</span>
                                Create Jira Protocol
                            </button>
                        </div>
                    </div>
                )}
            </div>

            <footer className="h-10 bg-[#111a22] border-t border-[#233648] px-6 py-2 flex items-center justify-between text-[10px] text-[#92adc9] z-20 font-bold uppercase tracking-widest">
                <div className="flex items-center gap-4">
                    <span>© 2026 FleetOps Debug Console. v3.1-STABLE</span>
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
                    <span>WebSocket: <span className="text-emerald-500">Connected</span></span>
                </div>
                <div className="flex items-center gap-6">
                    <span>Signal: <span className="text-white">Optimal</span></span>
                    <span>Latency: <span className="text-emerald-500">24ms</span></span>
                    <span>Server UTC: <span className="font-mono text-white">{new Date().toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit' })}</span></span>
                </div>
            </footer>
        </main>
    );
};

export default TroubleshootingLogs;
