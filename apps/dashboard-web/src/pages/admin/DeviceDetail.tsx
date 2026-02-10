import React, { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import apiClient from '../../services/apiClient';

interface Telemetry {
    device_id: string;
    model: string;
    firmware: string;
    imei: string;
    iccid: string;
    voltage: number;
    temp: number;
    ignition: boolean;
    gps: {
        status: string;
        sats: number;
        lat: number;
        lon: number;
    };
    accelerometer: {
        x: number;
        y: number;
        z: number;
    };
    storage_used_pct: number;
}

interface LogEntry {
    ts: string;
    level: string;
    msg: string;
    color: string;
}

export const DeviceDetail: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const [telemetry, setTelemetry] = useState<Telemetry | null>(null);
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [diagProgress, setDiagProgress] = useState(72);

    const fetchData = useCallback(async () => {
        try {
            const [telRes, logRes] = await Promise.all([
                apiClient.get(`/admin/devices/${id}/telemetry`),
                apiClient.get(`/admin/devices/${id}/logs`)
            ]) as any[];
            setTelemetry(telRes);
            setLogs(logRes || []);
        } catch (error) {
            console.error('Failed to fetch device details:', error);
        } finally {
            setLoading(false);
        }
    }, [id]);

    useEffect(() => {
        fetchData();
        const interval = setInterval(() => {
            setDiagProgress(prev => (prev >= 100 ? 0 : prev + 1));
        }, 500);
        return () => clearInterval(interval);
    }, [fetchData]);

    const handleAction = async (action: string) => {
        try {
            await apiClient.post(`/admin/devices/${id}/commands`, { action });
            alert(`Action ${action} enqueued successfully`);
        } catch (error) {
            alert('Failed to execute action');
        }
    };

    if (loading) {
        return (
            <div className="flex h-full items-center justify-center bg-[#0b1219]">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-12 h-12 border-4 border-[#137fec] border-t-transparent rounded-full animate-spin"></div>
                    <span className="text-[#92adc9] font-mono">ESTABLISHING SECURE HANDSHAKE...</span>
                </div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full bg-[#0b1219] text-white font-display overflow-hidden">
            {/* Header */}
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <nav className="flex items-center text-[11px] text-[#92adc9] mb-0.5 space-x-1">
                            <Link to="/admin/devices" className="hover:text-white cursor-pointer transition-colors uppercase">DIAGNOSTICS</Link>
                            <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                            <span className="hover:text-white cursor-pointer transition-colors uppercase">DEVICES</span>
                            <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                            <span className="text-[#137fec] font-mono font-medium">{id}</span>
                        </nav>
                        <h2 className="text-white text-lg font-bold flex items-center gap-2">
                            Telematics Control Unit #{id?.split('-').pop()}
                            <span className="w-2 h-2 rounded-full bg-[#10b981] animate-pulse"></span>
                        </h2>
                    </div>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded bg-[#16202a] border border-[#233648]">
                        <span className="w-2 h-2 rounded-full bg-[#10b981]"></span>
                        <span className="text-xs text-white font-medium">Online</span>
                        <span className="text-[10px] text-[#92adc9] ml-1 border-l border-[#233648] pl-2">Last ping: 2s ago</span>
                    </div>
                    <button onClick={fetchData} className="w-8 h-8 flex items-center justify-center rounded hover:bg-[#233648] text-[#92adc9] transition-colors">
                        <span className="material-symbols-outlined text-[20px]">refresh</span>
                    </button>
                    <button className="w-8 h-8 flex items-center justify-center rounded hover:bg-[#233648] text-[#92adc9] transition-colors">
                        <span className="material-symbols-outlined text-[20px]">more_vert</span>
                    </button>
                </div>
            </header>

            <div className="flex-1 overflow-hidden p-6 custom-scrollbar">
                <div className="grid grid-cols-12 gap-6 h-full">
                    {/* Left Panel: Info & Telemetry */}
                    <div className="col-span-3 flex flex-col gap-6 h-full overflow-y-auto custom-scrollbar pr-2">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
                            <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[#137fec] text-[18px]">memory</span>
                                    Hardware Information
                                </h3>
                            </div>
                            <div className="p-4 space-y-4">
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <div className="text-[10px] text-[#92adc9] uppercase">Model</div>
                                        <div className="text-sm text-white font-medium">{telemetry?.model}</div>
                                    </div>
                                    <div>
                                        <div className="text-[10px] text-[#92adc9] uppercase">Firmware</div>
                                        <div className="text-sm text-white font-medium">v4.2.1-stable</div>
                                    </div>
                                    <div>
                                        <div className="text-[10px] text-[#92adc9] uppercase">IMEI</div>
                                        <div className="text-sm text-[#92adc9] font-mono truncate">{telemetry?.imei}</div>
                                    </div>
                                    <div>
                                        <div className="text-[10px] text-[#92adc9] uppercase">SIM ICCID</div>
                                        <div className="text-sm text-[#92adc9] font-mono truncate">{telemetry?.iccid}</div>
                                    </div>
                                </div>
                                <div className="pt-4 border-t border-[#233648]">
                                    <div className="flex justify-between items-center mb-1">
                                        <span className="text-xs text-[#92adc9]">Internal Battery</span>
                                        <span className="text-xs text-[#10b981] font-bold">94%</span>
                                    </div>
                                    <div className="w-full bg-[#111a22] h-1.5 rounded-full overflow-hidden">
                                        <div className="bg-[#10b981] h-full" style={{ width: '94%' }}></div>
                                    </div>
                                </div>
                                <div>
                                    <div className="flex justify-between items-center mb-1">
                                        <span className="text-xs text-[#92adc9]">Storage</span>
                                        <span className="text-xs text-[#137fec] font-bold">{telemetry?.storage_used_pct}%</span>
                                    </div>
                                    <div className="w-full bg-[#111a22] h-1.5 rounded-full overflow-hidden">
                                        <div className="bg-[#137fec] h-full" style={{ width: `${telemetry?.storage_used_pct}%` }}></div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden flex-1">
                            <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[#f59e0b] text-[18px]">sensors</span>
                                    Sensor Telemetry
                                </h3>
                            </div>
                            <div className="p-4 space-y-4">
                                <div className="flex items-center justify-between p-3 bg-[#111a22] rounded border border-[#233648]">
                                    <div className="flex items-center gap-3">
                                        <span className="material-symbols-outlined text-[#92adc9]">key</span>
                                        <div>
                                            <div className="text-xs text-[#92adc9]">Ignition</div>
                                            <div className="text-sm font-bold text-white uppercase">{telemetry?.ignition ? 'ON' : 'OFF'}</div>
                                        </div>
                                    </div>
                                    <div className="text-[10px] text-[#10b981] bg-[#10b981]/10 px-2 py-0.5 rounded border border-[#10b981]/20">Active</div>
                                </div>
                                <div className="flex items-center justify-between p-3 bg-[#111a22] rounded border border-[#233648]">
                                    <div className="flex items-center gap-3">
                                        <span className="material-symbols-outlined text-[#92adc9]">satellite_alt</span>
                                        <div>
                                            <div className="text-xs text-[#92adc9]">GPS Fix</div>
                                            <div className="text-sm font-bold text-white">{telemetry?.gps.status} ({telemetry?.gps.sats} Sats)</div>
                                        </div>
                                    </div>
                                    <span className="material-symbols-outlined text-[#10b981] text-[18px]">check_circle</span>
                                </div>
                                <div className="p-3 bg-[#111a22] rounded border border-[#233648]">
                                    <div className="flex items-center gap-3 mb-3">
                                        <span className="material-symbols-outlined text-[#92adc9]">vibration</span>
                                        <div>
                                            <div className="text-xs text-[#92adc9]">Accelerometer</div>
                                            <div className="text-xs text-[#92adc9] font-mono">X: {telemetry?.accelerometer.x} | Y: {telemetry?.accelerometer.y} | Z: {telemetry?.accelerometer.z}</div>
                                        </div>
                                    </div>
                                    <div className="flex gap-1 h-8 items-end justify-center">
                                        {[20, 40, 30, 60, 25, 45, 20, 35, 10].map((h, i) => (
                                            <div key={i} className={`w-1 bg-[#137fec] rounded-t`} style={{ height: `${h}%`, opacity: h/100 + 0.3 }}></div>
                                        ))}
                                    </div>
                                </div>
                                <div className="flex items-center justify-between p-3 bg-[#111a22] rounded border border-[#233648]">
                                    <div className="flex items-center gap-3">
                                        <span className="material-symbols-outlined text-[#92adc9]">thermostat</span>
                                        <div>
                                            <div className="text-xs text-[#92adc9]">Unit Temp</div>
                                            <div className="text-sm font-bold text-white">{telemetry?.temp}°C</div>
                                        </div>
                                    </div>
                                    <span className="text-[10px] text-[#10b981]">Normal</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Middle Panel: Diagnostics & Logs */}
                    <div className="col-span-6 flex flex-col gap-6 h-full">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                            <div className="flex justify-between items-center mb-4">
                                <div>
                                    <h3 className="text-white text-sm font-semibold">System Diagnostics</h3>
                                    <p className="text-[11px] text-[#92adc9]">Running full system check sequence...</p>
                                </div>
                                <span className="text-xs font-mono text-[#137fec] animate-pulse">{diagProgress}% Running...</span>
                            </div>
                            <div className="space-y-2">
                                <div className="flex justify-between text-xs mb-1">
                                    <span className="text-white">Progress</span>
                                    <span className="text-[#92adc9]">{diagProgress}%</span>
                                </div>
                                <div className="w-full bg-[#111a22] h-2 rounded-full overflow-hidden">
                                    <div className="bg-[#137fec] h-full relative transition-all duration-500" style={{ width: `${diagProgress}%` }}>
                                        <div className="absolute top-0 bottom-0 right-0 w-full bg-gradient-to-l from-white/20 to-transparent"></div>
                                    </div>
                                </div>
                                <div className="flex gap-2 mt-3">
                                    {['Memory', 'I/O Ports', 'Network', 'CAN Bus'].map((label, i) => {
                                        const isActive = (i + 1) * 25 <= diagProgress;
                                        return (
                                            <div key={label} className={`flex-1 bg-[#111a22] rounded p-2 border ${isActive ? 'border-[#137fec]/40 bg-[#137fec]/5' : 'border-[#233648]'} flex items-center justify-center gap-2`}>
                                                <span className={`w-2 h-2 rounded-full ${isActive ? 'bg-[#10b981]' : 'bg-[#233648]'}`}></span>
                                                <span className={`text-[10px] uppercase ${isActive ? 'text-white font-bold' : 'text-[#92adc9]'}`}>{label}</span>
                                            </div>
                                        );
                                    })}
                                </div>
                            </div>
                        </div>

                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col flex-1 overflow-hidden">
                            <div className="p-3 px-4 border-b border-[#233648] bg-[#111a22] flex justify-between items-center">
                                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[#92adc9] text-[16px]">terminal</span>
                                    Live Connection Logs
                                </h3>
                                <div className="flex gap-2">
                                    <span className="flex items-center gap-1 px-2 py-0.5 rounded bg-[#16202a] border border-[#233648] text-[10px] text-[#92adc9]">
                                        <span className="w-1.5 h-1.5 rounded-full bg-[#10b981]"></span> WebSocket Open
                                    </span>
                                    <button className="text-[#92adc9] hover:text-white"><span className="material-symbols-outlined text-[16px]">pause</span></button>
                                    <button onClick={() => setLogs([])} className="text-[#92adc9] hover:text-white"><span className="material-symbols-outlined text-[16px]">delete</span></button>
                                </div>
                            </div>
                            <div className="flex-1 bg-[#0b1219] p-4 overflow-y-auto custom-scrollbar font-mono text-[11px] leading-relaxed">
                                {logs.map((log, i) => (
                                    <div key={i} className="text-[#92adc9] mb-0.5">
                                        <span className="text-[#569cd6]">{log.ts}</span>{' '}
                                        <span style={{ color: log.color }}>[{log.level}]</span>{' '}
                                        <span>{log.msg}</span>
                                    </div>
                                ))}
                                <div className="text-white animate-pulse">_</div>
                            </div>
                        </div>
                    </div>

                    {/* Right Panel: Actions & Errors */}
                    <div className="col-span-3 flex flex-col gap-6 h-full">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
                            <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[#92adc9] text-[18px]">build</span>
                                    Remote Actions
                                </h3>
                            </div>
                            <div className="p-4 space-y-3">
                                {[
                                    { label: 'Restart Device', icon: 'restart_alt', action: 'RESTART', hover: 'hover:border-[#137fec]/50 hover:text-[#137fec]' },
                                    { label: 'Clear Cache', icon: 'delete_sweep', action: 'CLEAR_CACHE', hover: 'hover:border-[#f59e0b]/50 hover:text-[#f59e0b]' },
                                    { label: 'Ping Location', icon: 'share_location', action: 'PING_GPS', hover: 'hover:border-[#10b981]/50 hover:text-[#10b981]' },
                                    { label: 'Factory Reset', icon: 'factory', action: 'FACTORY_RESET', hover: 'hover:border-[#ef4444]/50 hover:text-[#ef4444]' },
                                ].map((item) => (
                                    <button 
                                        key={item.label}
                                        onClick={() => handleAction(item.action)}
                                        className={`w-full py-2 px-3 bg-[#111a22] border border-[#233648] ${item.hover} text-white rounded flex items-center justify-between group transition-all`}
                                    >
                                        <span className="flex items-center gap-2 text-sm">
                                            <span className={`material-symbols-outlined text-[#92adc9] group-hover:text-inherit text-[18px]`}>{item.icon}</span> 
                                            {item.label}
                                        </span>
                                        <span className="material-symbols-outlined text-[#92adc9] text-[14px]">chevron_right</span>
                                    </button>
                                ))}
                            </div>
                        </div>

                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex-1 overflow-hidden flex flex-col">
                            <div className="p-4 border-b border-[#233648] bg-[#111a22] flex justify-between items-center">
                                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[#ef4444] text-[18px]">history_toggle_off</span>
                                    Error History
                                </h3>
                                <span className="bg-[#ef4444]/10 text-[#ef4444] text-[10px] px-1.5 py-0.5 rounded border border-[#ef4444]/20">3 new</span>
                            </div>
                            <div className="flex-1 overflow-y-auto custom-scrollbar p-2">
                                <div className="space-y-1">
                                    {[
                                        { title: 'Connection Timeout', time: '10m ago', msg: 'Failed to reach server endpoint after 3 retries. Error Code: 504.', color: 'border-[#ef4444]' },
                                        { title: 'High Latency', time: '2h ago', msg: 'Network round-trip time > 2000ms. Check signal quality.', color: 'border-[#f59e0b]' },
                                        { title: 'GPS Signal Lost', time: 'Yesterday', msg: 'Tunnel detected. Lost satellite fix for 120s.', color: 'border-[#ef4444]' },
                                        { title: 'Update Failed', time: 'Oct 20', msg: 'Firmware CRC check mismatch. Aborting update.', color: 'border-[#233648]' },
                                        { title: 'Power Warning', time: 'Oct 18', msg: 'Voltage dropped below 11.5V momentarily.', color: 'border-[#233648]' },
                                    ].map((err, i) => (
                                        <div key={i} className={`p-3 hover:bg-[#111a22] rounded cursor-pointer border-l-2 ${err.color} ${i===0 ? 'bg-[#111a22]/30' : ''}`}>
                                            <div className="flex justify-between items-start mb-1">
                                                <span className={`text-xs font-bold ${i < 3 ? 'text-white' : 'text-[#92adc9]'}`}>{err.title}</span>
                                                <span className="text-[10px] text-[#92adc9]">{err.time}</span>
                                            </div>
                                            <p className="text-[11px] text-[#92adc9] line-clamp-2">{err.msg}</p>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Footer */}
            <footer className="bg-[#111a22] border-t border-[#233648] px-6 py-2 flex items-center justify-between text-[10px] text-[#92adc9] z-20">
                <span>© 2026 FleetPulse Enterprise. Device Module V2.5</span>
                <div className="flex items-center gap-4">
                    <span>Protocol: <span className="text-[#137fec]">MQTT/TLS</span></span>
                    <span>Gateway: <span className="text-white">us-east-gw-04</span></span>
                    <span>Last Sync: <span className="font-mono text-white">{new Date().toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit' })} UTC</span></span>
                </div>
            </footer>
        </div>
    );
};
