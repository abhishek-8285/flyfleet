import React from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';

const MonitoringLayout: React.FC = () => {
    const location = useLocation();
    const navigate = useNavigate();
    
    const isActive = (path: string) => {
        if (path === '/admin/monitoring' && location.pathname === '/admin/monitoring') return true;
        return location.pathname.startsWith(path) && path !== '/admin/monitoring';
    };

    const getLinkClass = (path: string) => {
        const active = isActive(path);
        return `flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors ${
            active 
            ? "bg-monitor-primary/10 text-monitor-primary border-l-4 border-monitor-primary" 
            : "text-monitor-text-muted hover:bg-monitor-border hover:text-white"
        }`;
    };

    return (
        <div className="flex h-screen w-full bg-monitor-bg-dark font-public-sans text-monitor-text-muted overflow-hidden">
            <aside className="flex flex-col w-64 h-full bg-monitor-surface-darker border-r border-monitor-border flex-shrink-0 z-30">
                <div className="p-6 flex items-center gap-3">
                    <div className="w-8 h-8 rounded bg-monitor-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(236,91,19,0.3)]">
                        <span className="material-symbols-outlined">monitoring</span>
                    </div>
                    <h1 className="text-lg font-bold tracking-tight text-white font-public-sans">FleetMonitor</h1>
                </div>
                <nav className="flex-1 px-3 space-y-1 overflow-y-auto custom-scrollbar">
                    <div className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-monitor-text-muted">Operations</div>
                    <Link to="/admin/monitoring" className={getLinkClass("/admin/monitoring")}>
                        <span className="material-symbols-outlined">dashboard</span>
                        <span className="text-sm font-medium">System Health</span>
                    </Link>
                    <Link to="/admin/monitoring/metrics" className={getLinkClass("/admin/monitoring/metrics")}>
                        <span className="material-symbols-outlined">speed</span>
                        <span className="text-sm font-medium">Performance</span>
                    </Link>
                    <Link to="/admin/monitoring/jobs" className={getLinkClass("/admin/monitoring/jobs")}>
                        <span className="material-symbols-outlined">dns</span>
                        <span className="text-sm font-medium">Job Queues</span>
                    </Link>

                    <div className="px-3 mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-monitor-text-muted">Reliability</div>
                    <Link to="/admin/monitoring/sla" className={getLinkClass("/admin/monitoring/sla")}>
                        <span className="material-symbols-outlined">assignment_turned_in</span>
                        <span className="text-sm font-medium">SLA & Uptime</span>
                    </Link>
                        <Link to="/admin/monitoring/logs" className={getLinkClass("/admin/monitoring/logs")}>
                        <span className="material-symbols-outlined">bug_report</span>
                        <span className="text-sm font-medium">Error Logs</span>
                    </Link>

                    <div className="px-3 mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-monitor-text-muted">Applications</div>
                    <a href="#" className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-monitor-text-muted hover:bg-monitor-border hover:text-white transition-colors">
                        <span className="material-symbols-outlined">api</span>
                        <span className="text-sm font-medium">API Gateway</span>
                    </a>
                    <a href="#" className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-monitor-text-muted hover:bg-monitor-border hover:text-white transition-colors">
                        <span className="material-symbols-outlined">database</span>
                        <span className="text-sm font-medium">Database</span>
                    </a>
                </nav>
                <div className="p-4 border-t border-monitor-border">
                    <div className="flex items-center gap-3">
                        <div className="bg-center bg-no-repeat bg-cover rounded-full size-10 border-2 border-monitor-border" style={{backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuB4VhzOFW3Dq3K6lVra-IfzYBB25STyIEccTv5dhkHOP-UG8PC324sKwt0X1gtpfVRy_03a0VGH-UM_jfzD8OF3GLXMtXl60OV-x50rfLua-I3ee-Nk_-1tnyB5ojs_4_Rf4e7prPUNNcbTTTaOdgf-iF7L4WRC12e2czLr8OoPSWOUU72a9RqffeHmgMAAjNmh6zath8vPdZFF2cTHTDZsV0HhJd5EvSuqdR--8rd_9Jhb4JpqD_gbm_xI5Hha6JYZPAXuKlTUBm4")'}}></div>
                        <div className="flex flex-col overflow-hidden">
                            <span className="text-sm font-medium text-white truncate">Alex Morgan</span>
                            <span className="text-xs text-monitor-text-muted truncate">DevOps Lead</span>
                        </div>
                        <button onClick={() => navigate('/admin')} className="ml-auto text-monitor-text-muted hover:text-white" title="Exit Monitoring">
                            <span className="material-symbols-outlined text-[20px]">logout</span>
                        </button>
                    </div>
                </div>
            </aside>
            <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <Outlet />
            </div>
        </div>
    );
};

export default MonitoringLayout;
