import React, { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import apiClient from '../../services/apiClient';

interface Device {
    id: string;
    serial_number: string;
    type: string;
    status: 'active' | 'offline' | 'warning';
    vehicle_id?: string;
    last_heartbeat: string;
    firmware_version: string;
    signal_dbm: number;
    battery_pct: number;
    last_ping_at?: string;
    created_at: string;
}

interface DeviceMetrics {
    total: number;
    online: number;
    offline: number;
    low_battery: number;
    signal_warnings: number;
    system_errors: number;
}

export const Devices: React.FC = () => {
    const [devices, setDevices] = useState<Device[]>([]);
    const [metrics, setMetrics] = useState<DeviceMetrics | null>(null);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [autoRefresh, setAutoRefresh] = useState(true);

    const fetchData = useCallback(async () => {
        try {
            const res = await apiClient.get('/admin/devices?page=1&size=100') as any;
            setDevices(res.data || []);
            setMetrics(res.meta || null);
        } catch (error) {
            console.error('Failed to fetch devices:', error);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchData();
        let interval: any;
        if (autoRefresh) {
            interval = setInterval(fetchData, 30000);
        }
        return () => clearInterval(interval);
    }, [fetchData, autoRefresh]);

    const filteredDevices = devices.filter(dev => 
        dev.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
        dev.serial_number.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (dev.vehicle_id && dev.vehicle_id.toLowerCase().includes(searchTerm.toLowerCase()))
    );

    const getSignalIcon = (dbm: number) => {
        if (dbm === 0) return <span className="material-symbols-outlined text-[16px] text-[#92adc9]">signal_cellular_off</span>;
        if (dbm > -70) return <span className="material-symbols-outlined text-[16px] text-[#10b981]">signal_cellular_alt</span>;
        if (dbm > -90) return <span className="material-symbols-outlined text-[16px] text-[#f59e0b]">signal_cellular_alt_2_bar</span>;
        return <span className="material-symbols-outlined text-[16px] text-[#ef4444]">signal_cellular_alt_1_bar</span>;
    };

    const getBatteryColor = (pct: number) => {
        if (pct > 50) return 'bg-[#10b981]';
        if (pct > 20) return 'bg-[#f59e0b]';
        return 'bg-[#ef4444]';
    };

    return (
        <div className="flex flex-col h-full bg-[#0b1219] text-white font-display overflow-hidden">
            {/* Sub-Header */}
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <h2 className="text-white text-lg font-bold">Device Health Dashboard</h2>
                        <p className="text-[11px] text-[#92adc9] mt-0.5">Device Diagnostics Module • Enterprise View</p>
                    </div>
                    <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-[#10b981]/20 text-[#10b981] border border-[#10b981]/30">
                        System Healthy
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
                            <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#137fec]"></div>
                        </label>
                    </div>
                    <button 
                        onClick={fetchData}
                        className="px-4 py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-[#137fec]/20"
                    >
                        <span className="material-symbols-outlined text-[16px]">refresh</span>
                        Run Diagnostics
                    </button>
                </div>
            </header>

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                {/* Metrics Grid */}
                <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-6">
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-white">developer_board</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Total Devices</div>
                        <div className="text-2xl font-bold text-white">{metrics?.total.toLocaleString() || '---'}</div>
                        <div className="text-[10px] text-[#10b981] flex items-center gap-1 mt-1">
                            <span className="material-symbols-outlined text-[12px]">add_circle</span>
                            +15 new
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-[#10b981]">wifi</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Online</div>
                        <div className="text-2xl font-bold text-[#10b981]">{metrics?.online.toLocaleString() || '---'}</div>
                        <div className="text-[10px] text-[#92adc9] mt-1">95.7% uptime</div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-[#10b981] h-full" style={{ width: '95.7%' }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-[#92adc9]">wifi_off</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Offline</div>
                        <div className="text-2xl font-bold text-[#92adc9]">{metrics?.offline.toLocaleString() || '---'}</div>
                        <div className="text-[10px] text-[#f59e0b] mt-1">Check connectivity</div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-[#92adc9] h-full" style={{ width: '4.3%' }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-[#f59e0b]">battery_alert</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Low Battery</div>
                        <div className="text-2xl font-bold text-[#f59e0b]">{metrics?.low_battery.toLocaleString() || '---'}</div>
                        <div className="text-[10px] text-[#f59e0b] flex items-center gap-1 mt-1">
                            <span className="material-symbols-outlined text-[12px]">warning</span>
                            Below 20%
                        </div>
                        <div className="w-full bg-[#233648] h-1 rounded-full mt-2 overflow-hidden">
                            <div className="bg-[#f59e0b] h-full" style={{ width: '15%' }}></div>
                        </div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-orange-500">signal_cellular_alt_2_bar</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Signal Warnings</div>
                        <div className="text-2xl font-bold text-orange-500">{metrics?.signal_warnings.toLocaleString() || '---'}</div>
                        <div className="text-[10px] text-[#92adc9] mt-1">Weak reception area</div>
                    </div>

                    <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                            <span className="material-symbols-outlined text-4xl text-[#ef4444]">error</span>
                        </div>
                        <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">System Errors</div>
                        <div className="text-2xl font-bold text-[#ef4444]">{metrics?.system_errors.toLocaleString() || '---'}</div>
                        <div className="text-[10px] text-[#ef4444] mt-1">Critical faults</div>
                    </div>
                </div>

                {/* Device Table */}
                <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden min-h-[500px]">
                    <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                        <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                            <span className="material-symbols-outlined text-[#137fec] text-[18px]">list_alt</span>
                            Device Status & Health
                        </h3>
                        <div className="flex items-center gap-2">
                            <div className="relative">
                                <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[18px] text-[#92adc9]">search</span>
                                <input 
                                    type="text" 
                                    placeholder="Search Device ID or Vehicle..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="bg-[#16202a] border border-[#233648] text-xs text-white rounded px-3 py-1.5 pl-10 w-64 focus:ring-1 focus:ring-[#137fec] focus:border-[#137fec] outline-none"
                                />
                            </div>
                            <button className="p-1.5 hover:bg-[#233648] rounded text-[#92adc9] hover:text-white transition-colors">
                                <span className="material-symbols-outlined text-[18px]">filter_list</span>
                            </button>
                            <button className="p-1.5 hover:bg-[#233648] rounded text-[#92adc9] hover:text-white transition-colors">
                                <span className="material-symbols-outlined text-[18px]">download</span>
                            </button>
                        </div>
                    </div>

                    <div className="overflow-x-auto">
                        <table className="w-full text-left border-collapse">
                            <thead>
                                <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                                    <th className="px-4 py-3 font-semibold">Device ID</th>
                                    <th className="px-4 py-3 font-semibold">Assigned Vehicle</th>
                                    <th className="px-4 py-3 font-semibold">Status</th>
                                    <th className="px-4 py-3 font-semibold">Last Heartbeat</th>
                                    <th className="px-4 py-3 font-semibold">Firmware</th>
                                    <th className="px-4 py-3 font-semibold text-center">Signal (dBm)</th>
                                    <th className="px-4 py-3 font-semibold">Battery %</th>
                                    <th className="px-4 py-3 text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="text-sm">
                                {loading ? (
                                    <tr>
                                        <td colSpan={8} className="px-4 py-20 text-center text-[#92adc9]">
                                            <div className="flex flex-col items-center gap-3">
                                                <div className="w-8 h-8 border-2 border-[#137fec] border-t-transparent rounded-full animate-spin"></div>
                                                Initializing hardware diagnostics...
                                            </div>
                                        </td>
                                    </tr>
                                ) : filteredDevices.length > 0 ? (
                                    filteredDevices.map((device) => (
                                        <tr key={device.id} className="hover:bg-[#16202a]/50 transition-colors text-xs border-b border-[#233648] last:border-0 group">
                                            <td className="px-4 py-4 font-mono text-[#137fec]">
                                                <Link to={`/admin/devices/${device.id}`} className="hover:underline">{device.id}</Link>
                                            </td>
                                            <td className="px-4 py-4 text-white">{device.vehicle_id || 'Unassigned'}</td>
                                            <td className="px-4 py-4">
                                                <span className={`px-2 py-0.5 rounded-full bg-opacity-10 text-[10px] uppercase font-bold border ${
                                                    device.status === 'active' ? 'bg-[#10b981] text-[#10b981] border-[#10b981]/20' : 
                                                    device.status === 'warning' ? 'bg-[#f59e0b] text-[#f59e0b] border-[#f59e0b]/20' : 
                                                    'bg-[#ef4444] text-[#ef4444] border-[#ef4444]/20'
                                                }`}>
                                                    {device.status}
                                                </span>
                                            </td>
                                            <td className="px-4 py-4 text-[#92adc9]">{device.last_heartbeat}</td>
                                            <td className="px-4 py-4 text-[#92adc9]">{device.firmware_version}</td>
                                            <td className={`px-4 py-4 ${device.signal_dbm < -80 ? 'text-[#ef4444]' : 'text-white'}`}>
                                                <div className="flex items-center gap-2">
                                                    {getSignalIcon(device.signal_dbm)}
                                                    {device.signal_dbm} dBm
                                                </div>
                                            </td>
                                            <td className="px-4 py-4">
                                                <div className="flex items-center gap-2">
                                                    <div className="w-12 h-1.5 bg-[#111a22] rounded-full overflow-hidden">
                                                        <div 
                                                            className={`h-full ${getBatteryColor(device.battery_pct)}`} 
                                                            style={{ width: `${device.battery_pct}%` }}
                                                        ></div>
                                                    </div>
                                                    <span className={device.battery_pct < 20 ? 'text-[#ef4444]' : 'text-white'}>
                                                        {device.battery_pct}%
                                                    </span>
                                                </div>
                                            </td>
                                            <td className="px-4 py-4 text-right">
                                                <Link 
                                                    to={`/admin/devices/${device.id}`}
                                                    className="text-[#92adc9] hover:text-[#137fec] transition-colors p-1"
                                                >
                                                    <span className="material-symbols-outlined text-[18px]">more_vert</span>
                                                </Link>
                                            </td>
                                        </tr>
                                    ))
                                ) : (
                                    <tr>
                                        <td colSpan={8} className="px-4 py-20 text-center text-[#92adc9]">
                                            No devices matching "{searchTerm}" found.
                                        </td>
                                    </tr>
                                )}
                            </tbody>
                        </table>
                    </div>

                    <div className="mt-auto border-t border-[#233648] p-3 bg-[#111a22] flex items-center justify-between text-xs text-[#92adc9]">
                        <span>Showing {filteredDevices.length} of {metrics?.total || 0} devices</span>
                        <div className="flex items-center gap-2">
                            <button className="p-1 hover:text-white disabled:opacity-50"><span className="material-symbols-outlined">chevron_left</span></button>
                            <span className="text-white">Page 1 of {Math.ceil((metrics?.total || 1) / 100)}</span>
                            <button className="p-1 hover:text-white"><span className="material-symbols-outlined">chevron_right</span></button>
                        </div>
                    </div>
                </div>
            </div>

            {/* Footer */}
            <footer className="bg-[#111a22] border-t border-[#233648] px-6 py-2 flex items-center justify-between text-[10px] text-[#92adc9] z-20">
                <span>© 2026 FleetPulse Enterprise. Device Diagnostics Console V2.5</span>
                <div className="flex items-center gap-4">
                    <span>Protocol: <span className="text-[#10b981]">MQTT/TLS</span></span>
                    <span>System Status: <span className="text-[#10b981]">Operational</span></span>
                    <span>Server Time: <span className="font-mono text-white">{new Date().toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit' })} UTC</span></span>
                </div>
            </footer>
        </div>
    );
};
