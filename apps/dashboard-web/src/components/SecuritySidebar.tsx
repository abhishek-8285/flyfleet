import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthProvider';

const SecuritySidebar: React.FC = () => {
    const location = useLocation();
    const navigate = useNavigate();
    const { user } = useAuth();

    // Helper to determine active state
    const isActive = (path: string) => location.pathname === path ? "bg-primary/10 text-primary active" : "text-text-muted hover:bg-border-color hover:text-white";

    const handleLogout = () => {
        localStorage.removeItem('token');
        navigate('/login');
    };

    return (
        <aside className="flex flex-col w-64 h-full bg-[#111a22] border-r border-border-color flex-shrink-0 z-30">
            <div className="p-6 flex items-center gap-3">
                <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(19,127,236,0.3)]">
                    <span className="material-symbols-outlined">shield_lock</span>
                </div>
                <h1 className="text-lg font-bold tracking-tight text-white">FleetSecurity</h1>
            </div>
            <nav className="flex-1 px-3 space-y-1 overflow-y-auto custom-scrollbar">
                <div className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted">Security Suite</div>
                
                <Link to="/security" className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors nav-link ${isActive('/security')}`}>
                    <span className="material-symbols-outlined">dashboard</span>
                    <span className="text-sm font-medium">Overview</span>
                </Link>
                
                <Link to="/security/policy" className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors nav-link ${isActive('/security/policy')}`}>
                    <span className="material-symbols-outlined">tune</span>
                    <span className="text-sm font-medium">MFA Configuration</span>
                </Link>

                <Link to="/security/overrides" className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors nav-link ${isActive('/security/overrides')}`}>
                    <span className="material-symbols-outlined">domain</span>
                    <span className="text-sm font-medium">Org Overrides</span>
                </Link>
                    <Link to="/security/exceptions" className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors nav-link ${isActive('/security/exceptions')}`}>
                    <span className="material-symbols-outlined">rule_folder</span>
                    <span className="text-sm font-medium">Exceptions</span>
                </Link>

                <div className="px-3 mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted">Monitoring</div>
                {(user?.role_id === 'super_admin' || user?.role_id === 'org_admin') && (
                    <>
                        <Link to="/security/audit" className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors nav-link ${isActive('/security/audit')}`}>
                            <span className="material-symbols-outlined">troubleshoot</span>
                            <span className="text-sm font-medium">Audit Logs</span>
                        </Link>
                        <Link to="/security/audit/settings" className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors nav-link ${isActive('/security/audit/settings')}`}>
                            <span className="material-symbols-outlined">settings_suggest</span>
                            <span className="text-sm font-medium">Governance</span>
                        </Link>
                    </>
                )}
                <a className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-text-muted hover:bg-border-color hover:text-white transition-colors cursor-pointer">
                    <span className="material-symbols-outlined">warning</span>
                    <span className="text-sm font-medium">Threat Alerts</span>
                </a>

                <div className="px-3 mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted">Users</div>
                <a className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-text-muted hover:bg-border-color hover:text-white transition-colors cursor-pointer">
                    <span className="material-symbols-outlined">group</span>
                    <span className="text-sm font-medium">Directory</span>
                </a>
            </nav>
            <div className="p-4 border-t border-border-color">
                <div className="flex items-center gap-3">
                    <div className="bg-center bg-no-repeat bg-cover rounded-full size-10 border-2 border-border-color" style={{backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuB4VhzOFW3Dq3K6lVra-IfzYBB25STyIEccTv5dhkHOP-UG8PC324sKwt0X1gtpfVRy_03a0VGH-UM_jfzD8OF3GLXMtXl60OV-x50rfLua-I3ee-Nk_-1tnyB5ojs_4_Rf4e7prPUNNcbTTTaOdgf-iF7L4WRC12e2czLr8OoPSWOUU72a9RqffeHmgMAAjNmh6zath8vPdZFF2cTHTDZsV0HhJd5EvSuqdR--8rd_9Jhb4JpqD_gbm_xI5Hha6JYZPAXuKlTUBm4")'}}></div>
                    <div className="flex flex-col overflow-hidden">
                        <span className="text-sm font-medium text-white truncate">Abhishek</span>
                        <span className="text-xs text-text-muted truncate">Super Admin</span>
                    </div>
                    <button onClick={handleLogout} className="ml-auto text-text-muted hover:text-white">
                        <span className="material-symbols-outlined text-[20px]">logout</span>
                    </button>
                </div>
            </div>
        </aside>
    );
};

export default SecuritySidebar;
