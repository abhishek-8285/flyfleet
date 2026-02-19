import { useEffect, useState } from 'react';
import { ActivityFeed } from '../components/dashboard/ActivityFeed';
import { AddVehicleModal } from '../components/dashboard/AddVehicleModal'; // Added import
import { CriticalVehiclesTable } from '../components/dashboard/CriticalVehiclesTable';
import { LiveOperationsMap } from '../components/dashboard/LiveOperationsMap';
import { QuickActions } from '../components/dashboard/QuickActions';
import { StatsGrid } from '../components/dashboard/StatsGrid';
import { api } from '../services/api';
import { Alert, Trip } from '../types';

const Dashboard = () => {
    const [stats, setStats] = useState<any>({
        total_vehicles: 0,
        active_trips: 0,
        active_alerts: 0,
        total_drivers: 0,
        online_drivers: 0,
        trends: {
            vehicles: 0,
            trips: 0,
            alerts: 0
        }
    });
    const [trips, setTrips] = useState<Trip[]>([]);
    const [alerts, setAlerts] = useState<Alert[]>([]);
    const [loading, setLoading] = useState(true);
    const [isAddVehicleModalOpen, setIsAddVehicleModalOpen] = useState(false);
    const [refreshInterval, setRefreshInterval] = useState<number>(30000); // 30s default
    const [division, setDivision] = useState('Global Fleet Ops');
    const [timezone, setTimezone] = useState('UTC-05:00 (EST)');

    const fetchData = async () => {
        setLoading(true);
        try {
            const [tripsData, alertsData, statsData] = await Promise.all([
                api.getTrips({ page: 1, size: 50, status: 'ongoing' }).catch(() => ({ trips: [], total: 0 })),
                api.getAlerts({ page: 1, size: 5, severity: 'critical' }).catch(() => ({ alerts: [], total: 0 })),
                api.getStatistics().catch((err) => {
                    console.error("Stats fetch failed", err);
                    return null;
                })
            ]);

            setTrips(tripsData.trips || []);
            setAlerts(alertsData.alerts || []);

            if (statsData && (statsData as any).data) {
                const s = (statsData as any).data;
                setStats({
                    total_vehicles: Number(s.total_vehicles || 0),
                    active_trips: Number(s.active_trips || (tripsData as any).total || 0),
                    active_alerts: Number(s.alerts_count || (alertsData as any).total || 0),
                    total_drivers: Number(s.total_drivers || 0),
                    online_drivers: Number(s.active_vehicles || 0),
                    trends: {
                        vehicles: Number(s.vehicles_trend_pct || 0),
                        trips: Number(s.trips_trend_pct || 0),
                        alerts: Number(s.alerts_trend_pct || 0)
                    }
                });
            } else {
                setStats((prev: any) => ({
                    ...prev,
                    active_trips: Number((tripsData as any).total || (tripsData as any).trips?.length || 0),
                    active_alerts: Number((alertsData as any).total || (alertsData as any).alerts?.length || 0)
                }));
            }
        } catch (error) {
            console.error("Failed to load dashboard data", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        if (refreshInterval === 0) return;

        const interval = setInterval(fetchData, refreshInterval);
        return () => clearInterval(interval);
    }, [refreshInterval]); // Re-run when interval changes

    const handleRefreshChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        const val = e.target.value;
        if (val === 'Off') setRefreshInterval(0);
        else if (val === '30s') setRefreshInterval(30000);
        else if (val === '1m') setRefreshInterval(60000);
        else if (val === '5m') setRefreshInterval(300000);
    };

    return (
        <div className="flex flex-col h-full">
            {/* Modal Overlay */}
            <AddVehicleModal
                isOpen={isAddVehicleModalOpen}
                onClose={() => setIsAddVehicleModalOpen(false)}
            />

            {/* Quick Filters Bar */}
            <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10">
                <div className="flex flex-wrap items-center gap-3">
                    <div className="relative">
                        <select className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors">
                            <option>Today: {new Date().toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</option>
                            <option>Last 7 Days</option>
                            <option>Last 30 Days</option>
                            <option>Custom Range</option>
                        </select>
                        <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">calendar_today</span>
                    </div>
                    <div className="relative">
                        <select
                            value={division}
                            onChange={(e) => setDivision(e.target.value)}
                            className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors"
                        >
                            <option>Global Fleet Ops</option>
                            <option>North America Div</option>
                            <option>Europe Logistics</option>
                        </select>
                        <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">domain</span>
                    </div>
                    <div className="relative hidden lg:block">
                        <select
                            value={timezone}
                            onChange={(e) => setTimezone(e.target.value)}
                            className="appearance-none bg-surface-dark border border-border-color text-text-muted text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors"
                        >
                            <option>UTC-05:00 (EST)</option>
                            <option>UTC+00:00 (GMT)</option>
                            <option>UTC-08:00 (PST)</option>
                        </select>
                        <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">schedule</span>
                    </div>
                </div>
                <div className="flex items-center gap-3 ml-auto">
                    <div className="flex items-center gap-2 text-xs text-text-muted bg-border-color/30 px-2 py-1 rounded border border-border-color/50">
                        <span className="material-symbols-outlined text-[14px] text-emerald-500 animate-pulse">radio_button_checked</span>
                        <span>Live</span>
                    </div>
                    <div className="h-4 w-px bg-border-color"></div>
                    <button onClick={fetchData} className="text-text-muted hover:text-white transition-colors" title="Manual Refresh">
                        <span className="material-symbols-outlined text-[18px]">refresh</span>
                    </button>
                    <div className="flex items-center gap-2">
                        <span className="text-xs text-text-muted">Auto:</span>
                        <select
                            onChange={handleRefreshChange}
                            value={refreshInterval === 0 ? 'Off' : refreshInterval === 30000 ? '30s' : refreshInterval === 60000 ? '1m' : '5m'}
                            className="bg-transparent text-xs text-white border-none p-0 pr-4 focus:ring-0 cursor-pointer font-medium"
                        >
                            <option>30s</option>
                            <option>1m</option>
                            <option>5m</option>
                            <option>Off</option>
                        </select>
                    </div>
                </div>
            </div>

            {/* Main Scrollable Content */}
            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar pb-24">
                <div className="grid grid-cols-12 gap-6 max-w-[1600px] mx-auto">
                    {/* Stats Row */}
                    <StatsGrid stats={stats} />

                    {/* Middle Row: Map & Critical Table */}
                    <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
                        <LiveOperationsMap trips={trips} />
                        <CriticalVehiclesTable alerts={alerts} />
                    </div>

                    {/* Right Column: Quick Actions & Activity */}
                    <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
                        <QuickActions onAddVehicle={() => setIsAddVehicleModalOpen(true)} />
                        <ActivityFeed />
                    </div>
                </div>

                <footer className="mt-8 text-center text-xs text-[#64748b] pb-6 flex flex-col items-center gap-1">
                    <span>© 2026 FlyFleet Global Operations. All rights reserved.</span>
                    <span className="text-[10px]">v2.5.0-PRO | Node: {window.location.hostname} | Server Time: <span className="font-mono">{new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })} UTC</span></span>
                </footer>
            </div>
        </div>
    );
};

export default Dashboard;
