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
        total_vehicles: 842, // Default mock values to match design initially
        active_trips: 125,
        active_alerts: 3,
        online_drivers: 0
    });
    const [trips, setTrips] = useState<Trip[]>([]);
    const [alerts, setAlerts] = useState<Alert[]>([]); // eslint-disable-line @typescript-eslint/no-unused-vars
    const [loading, setLoading] = useState(true); // eslint-disable-line @typescript-eslint/no-unused-vars
    const [isAddVehicleModalOpen, setIsAddVehicleModalOpen] = useState(false); // Added state

    const fetchData = async () => {
        setLoading(true);
        try {
            // Fetch real data but fallback to mock if fails (for dev/demo robustnes)
            const [tripsData, alertsData, statsData] = await Promise.all([
                api.getTrips({ page: 1, size: 50, status: 'ongoing' }).catch(() => ({ trips: [], total: 0 })),
                api.getAlerts({ page: 1, size: 5, severity: 'critical' }).catch(() => ({ alerts: [], total: 0 })),
                api.getStatistics().catch(() => ({ total_vehicles: 842, active_trips: 125 })) 
            ]);

            setTrips(tripsData.trips || []);
            setAlerts(alertsData.alerts || []);
            setStats({
                ...statsData,
                total_vehicles: statsData.total_vehicles || 842,
                active_trips: tripsData.total || 125,
                active_alerts: alertsData.total || 3
            });
        } catch (error) {
            console.error("Failed to load dashboard data", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 30000); 
        return () => clearInterval(interval);
    }, []);

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
                        <select className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors">
                            <option>Global Fleet Ops</option>
                            <option>North America Div</option>
                            <option>Europe Logistics</option>
                        </select>
                        <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">domain</span>
                    </div>
                    <div className="relative hidden lg:block">
                        <select className="appearance-none bg-surface-dark border border-border-color text-text-muted text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors">
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
                        <select className="bg-transparent text-xs text-white border-none p-0 pr-4 focus:ring-0 cursor-pointer font-medium">
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
                         <CriticalVehiclesTable />
                    </div>

                    {/* Right Column: Quick Actions & Activity */}
                    <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
                        <QuickActions onAddVehicle={() => setIsAddVehicleModalOpen(true)} />
                        <ActivityFeed />
                    </div>
                </div>
                
                <footer className="mt-8 text-center text-xs text-[#64748b] pb-6 flex flex-col items-center gap-1">
                    <span>© 2026 Enterprise Fleet Manager. All rights reserved.</span>
                    <span className="text-[10px]">v2.4.1 | Server Time: <span className="font-mono">{new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })} UTC</span></span>
                </footer>
            </div>
        </div>
    );
};

export default Dashboard;
