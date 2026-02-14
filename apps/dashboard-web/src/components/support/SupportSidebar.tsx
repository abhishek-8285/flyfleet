import React, { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { getSidebarCollapsed, setSidebarCollapsed } from '../../utils/storage';

const SupportSidebar: React.FC = () => {
    const location = useLocation();
    const [collapsed, setCollapsed] = useState(false);

    // Load persisted sidebar state on mount
    useEffect(() => {
        const isCollapsed = getSidebarCollapsed();
        setCollapsed(isCollapsed);
    }, []);

    const toggleSidebar = () => {
        const newCollapsed = !collapsed;
        setCollapsed(newCollapsed);
        setSidebarCollapsed(newCollapsed);
    };

    // Handle responsive behavior
    useEffect(() => {
        const handleResize = () => {
            if (window.innerWidth < 1024) {
                setCollapsed(true);
            } else {
                const isPersistedCollapsed = getSidebarCollapsed();
                setCollapsed(isPersistedCollapsed);
            }
        };

        window.addEventListener('resize', handleResize);
        handleResize();

        return () => window.removeEventListener('resize', handleResize);
    }, []);

    const NavItem = ({ to, icon, label, exact = false }: { to: string, icon: string, label: string, exact?: boolean }) => {
        const active = exact ? location.pathname === to : location.pathname.startsWith(to);
        return (
            <Link 
                to={to} 
                title={collapsed ? label : ""}
                className={`flex items-center gap-3 py-2.5 rounded-lg transition-all duration-200 group ${
                    collapsed 
                        ? 'justify-center px-0' 
                        : 'px-3'
                } ${
                    active 
                        ? "bg-primary/10 text-primary border-l-4 border-primary" 
                        : "text-[#92adc9] hover:bg-[#233648] hover:text-white border-l-4 border-transparent"
                }`}
            >
                <span className="material-symbols-outlined flex-shrink-0">{icon}</span>
                {!collapsed && <span className="text-sm font-medium animate-in fade-in slide-in-from-left-2 duration-300">{label}</span>}
            </Link>
        );
    };

    return (
        <aside className={`flex flex-col h-full bg-[#111a22] border-r border-[#233648] flex-shrink-0 z-30 transition-all duration-300 relative ${collapsed ? 'w-20' : 'w-64'}`}>
            <div className={`p-6 flex items-center gap-3 ${collapsed ? 'justify-center px-2' : ''}`}>
                <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(19,127,236,0.3)] flex-shrink-0">
                    <span className="material-symbols-outlined">support_agent</span>
                </div>
                {!collapsed && <h1 className="text-lg font-bold tracking-tight text-white truncate animate-in fade-in duration-300">FleetSupport</h1>}
            </div>

            {/* Toggle Button */}
            <button 
                onClick={toggleSidebar}
                className="absolute -right-3 top-20 bg-primary text-white rounded-full size-6 flex items-center justify-center shadow-lg hover:bg-primary/90 transition-transform z-40 border border-white/10"
                title={collapsed ? "Expand Sidebar" : "Collapse Sidebar"}
            >
                <span className={`material-symbols-outlined text-sm transition-transform duration-300 ${collapsed ? 'rotate-180' : 'rotate-0'}`}>
                    chevron_left
                </span>
            </button>
            <nav className={`flex-1 space-y-1 overflow-y-auto custom-scrollbar ${collapsed ? 'px-2' : 'px-3'}`}>
                <div className={`px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-[#92adc9] ${collapsed ? 'text-center text-[10px] scale-90 px-2' : ''}`}>
                    {collapsed ? '•••' : 'Support Tools'}
                </div>
                <NavItem to="/support/dashboard" icon="dashboard" label="Overview" />
                <NavItem to="/support/incidents" icon="emergency" label="Incidents" />
                <NavItem to="/support/logs" icon="bug_report" label="Troubleshooting Logs" />
                <NavItem to="/support/tickets" icon="confirmation_number" label="Tickets" />
                
                <div className={`px-3 mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-[#92adc9] ${collapsed ? 'text-center text-[10px] scale-90 px-2' : ''}`}>
                    {collapsed ? '•••' : 'Platform Health'}
                </div>
                <NavItem to="/support/status" icon="monitor_heart" label="System Status" />
                <NavItem to="/support/alerts" icon="notifications_active" label="Alerts" />

                <div className="px-3 mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-[#92adc9]">Knowledge</div>
                <NavItem to="/support/articles" icon="library_books" label="Documentation" />
            </nav>
            <div className="p-4 border-t border-[#233648]">
                <div className="flex items-center gap-3">
                    <div className="bg-center bg-no-repeat bg-cover rounded-full size-10 border-2 border-[#233648]" style={{backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuB4VhzOFW3Dq3K6lVra-IfzYBB25STyIEccTv5dhkHOP-UG8PC324sKwt0X1gtpfVRy_03a0VGH-UM_jfzD8OF3GLXMtXl60OV-x50rfLua-I3ee-Nk_-1tnyB5ojs_4_Rf4e7prPUNNcbTTTaOdgf-iF7L4WRC12e2czLr8OoPSWOUU72a9RqffeHmgMAAjNmh6zath8vPdZFF2cTHTDZsV0HhJd5EvSuqdR--8rd_9Jhb4JpqD_gbm_xI5Hha6JYZPAXuKlTUBm4")'}}></div>
                    <div className="flex flex-col overflow-hidden">
                        <span className="text-sm font-medium text-white truncate">Alex Morgan</span>
                        <span className="text-xs text-[#92adc9] truncate">Support Lead</span>
                    </div>
                    <button className="ml-auto text-[#92adc9] hover:text-white">
                        <span className="material-symbols-outlined text-[20px]">logout</span>
                    </button>
                </div>
            </div>
        </aside>
    );
};

export default SupportSidebar;
