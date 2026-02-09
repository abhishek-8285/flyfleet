import React from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { AppDispatch, RootState } from '../../store';
import { logout } from '../../store/authSlice';

const Layout: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useSelector((state: RootState) => state.auth);
  const alertsState = useSelector((state: RootState) => state.alerts) as any;
  const unreadCount = alertsState?.unreadCount || 0;

  const handleLogout = async () => {
    await dispatch(logout());
    navigate('/login');
  };

  const menuItems = [
    { path: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
    { path: '/dashboard/map', label: 'Live Map', icon: 'map' },
    { path: '/dashboard/vehicles', label: 'Fleet', icon: 'directions_car' },
    { path: '/dashboard/drivers', label: 'Drivers', icon: 'badge' }, 
    { path: '/dashboard/maintenance', label: 'Maintenance', icon: 'build' },
    { path: '/dashboard/search', label: 'Global Search', icon: 'search' },
    { path: '/dashboard/timeline', label: 'Timeline', icon: 'history' },
  ];

  const analyticsItems = [
    { path: '/dashboard/reports', label: 'Reports', icon: 'analytics' },
    { path: '/dashboard/history', label: 'History', icon: 'history' },
  ];

  const securityItems = [
    { path: '/dashboard/security/mfa', label: 'MFA Center', icon: 'admin_panel_settings' },
  ];

  const [isSidebarExpanded, setIsSidebarExpanded] = React.useState(window.innerWidth >= 1024);
  const [isProfileOpen, setIsProfileOpen] = React.useState(false);
  const [isSearchOpen, setIsSearchOpen] = React.useState(false);
  const [isNotifOpen, setIsNotifOpen] = React.useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = React.useState(false);

  const toggleSidebar = () => setIsSidebarExpanded(!isSidebarExpanded);

  React.useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth < 1024) {
        setIsSidebarExpanded(false);
      } else {
        setIsSidebarExpanded(true);
      }
    };

    window.addEventListener('resize', handleResize);
    // Initial check
    handleResize();

    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return (
    <div className="flex h-screen w-full bg-background-light dark:bg-background-dark text-slate-900 dark:text-white font-display overflow-hidden">
      {/* Sidebar */}
      <aside className={`flex flex-col h-full bg-[#111a22] border-r border-border-color flex-shrink-0 z-30 transition-all duration-300 relative ${isSidebarExpanded ? 'w-64' : 'w-20'}`}>
        <div className={`p-6 flex items-center gap-3 ${!isSidebarExpanded && 'justify-center px-2'}`}>
          <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(19,127,236,0.3)] flex-shrink-0">
            <span className="material-symbols-outlined text-[20px]">local_shipping</span>
          </div>
          {isSidebarExpanded && <h1 className="text-lg font-bold tracking-tight text-white truncate animate-in fade-in duration-300">FleetCommand</h1>}
        </div>

        {/* Toggle Button */}
        <button 
          onClick={toggleSidebar}
          className="absolute -right-3 top-20 bg-primary text-white rounded-full size-6 flex items-center justify-center shadow-lg hover:bg-primary/90 transition-transform z-40 border border-white/10"
          title={isSidebarExpanded ? "Collapse Sidebar" : "Expand Sidebar"}
        >
          <span className={`material-symbols-outlined text-sm transition-transform duration-300 ${isSidebarExpanded ? 'rotate-0' : 'rotate-180'}`}>
            chevron_left
          </span>
        </button>

        <nav className={`flex-1 space-y-1 overflow-y-auto custom-scrollbar ${isSidebarExpanded ? 'px-3' : 'px-2'}`}>
          <div className={`mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted ${isSidebarExpanded ? 'px-3' : 'text-center text-[10px] scale-90'}`}>
            {isSidebarExpanded ? 'Menu' : '•••'}
          </div>
          {menuItems.map((item) => (
            <Link 
              key={item.path}
              to={item.path} 
              title={!isSidebarExpanded ? item.label : ""}
              className={`flex items-center gap-3 py-2.5 rounded-lg transition-all duration-200 group ${isSidebarExpanded ? 'px-3' : 'justify-center px-0'} ${location.pathname === item.path || (item.path === '/dashboard/vehicles' && location.pathname.startsWith('/dashboard/vehicles')) ? 'bg-primary/10 text-primary border-l-4 border-primary' : 'text-text-muted hover:bg-border-color hover:text-white border-l-4 border-transparent'}`}>
              <span className="material-symbols-outlined text-[20px] flex-shrink-0">{item.icon}</span>
              {isSidebarExpanded && <span className="text-sm font-medium animate-in fade-in slide-in-from-left-2 duration-300">{item.label}</span>}
            </Link>
          ))}
          
          <div className={`mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted ${isSidebarExpanded ? 'px-3' : 'text-center text-[10px] scale-90'}`}>
            {isSidebarExpanded ? 'Analytics' : '•••'}
          </div>
          {analyticsItems.map((item) => (
            <Link 
              key={item.path}
              to={item.path} 
              title={!isSidebarExpanded ? item.label : ""}
              className={`flex items-center gap-3 py-2.5 rounded-lg transition-all duration-200 group ${isSidebarExpanded ? 'px-3' : 'justify-center px-0'} ${location.pathname === item.path ? 'bg-primary/10 text-primary border-l-4 border-primary' : 'text-text-muted hover:bg-border-color hover:text-white border-l-4 border-transparent'}`}>
              <span className="material-symbols-outlined text-[20px] flex-shrink-0">{item.icon}</span>
              {isSidebarExpanded && <span className="text-sm font-medium animate-in fade-in slide-in-from-left-2 duration-300">{item.label}</span>}
            </Link>
          ))}

          {/* Security Suite - Only for Admin */}
          {(user?.role?.toLowerCase() === 'admin' || user?.role?.toLowerCase() === 'super admin' || !user?.role) && (
            <>
              <div className={`mt-6 mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted ${isSidebarExpanded ? 'px-3' : 'text-center text-[10px] scale-90'}`}>
                {isSidebarExpanded ? 'Security' : '•••'}
              </div>
              {securityItems.map((item) => (
                <Link 
                  key={item.path}
                  to={item.path} 
                  title={!isSidebarExpanded ? item.label : ""}
                  className={`flex items-center gap-3 py-2.5 rounded-lg transition-all duration-200 group ${isSidebarExpanded ? 'px-3' : 'justify-center px-0'} ${location.pathname === item.path ? 'bg-primary/10 text-primary border-l-4 border-primary' : 'text-text-muted hover:bg-border-color hover:text-white border-l-4 border-transparent'}`}>
                  <span className="material-symbols-outlined text-[20px] flex-shrink-0">{item.icon}</span>
                  {isSidebarExpanded && <span className="text-sm font-medium animate-in fade-in slide-in-from-left-2 duration-300">{item.label}</span>}
                </Link>
              ))}
            </>
          )}
        </nav>

        <div className={`p-4 border-t border-border-color bg-[#0f151b] transition-all duration-300 ${!isSidebarExpanded && 'px-2'}`}>
          <div className={`relative group cursor-pointer ${!isSidebarExpanded && 'flex justify-center'}`} onClick={() => setIsProfileOpen(!isProfileOpen)}>
            <div className={`flex items-center gap-3 w-full text-left ${!isSidebarExpanded && 'justify-center'}`}>
              <div className="w-10 h-10 rounded-full bg-surface-dark border border-border-color flex items-center justify-center text-text-muted font-bold text-lg overflow-hidden whitespace-nowrap flex-shrink-0">
                  {user?.name?.split(' ').map(n => n[0]).join('') || 'U'}
              </div>
              {isSidebarExpanded && (
                <div className="flex flex-col overflow-hidden flex-1 animate-in fade-in slide-in-from-left-2 duration-300">
                    <span className="text-sm font-medium text-white truncate">{user?.name || 'User'}</span>
                    <span className="text-xs text-text-muted truncate capitalize">{user?.role || 'Admin'}</span>
                </div>
              )}
              {isSidebarExpanded && <span className="material-symbols-outlined text-text-muted">expand_less</span>}
            </div>
            
            {/* Profile Dropdown */}
            {isProfileOpen && (
            <div className="absolute bottom-full left-0 mb-2 w-full bg-surface-dark border border-border-color rounded-lg shadow-xl z-50 overflow-hidden animate-in slide-in-from-bottom-2 duration-200">
                <div className="p-2 space-y-1">
                    <a href="#" className="flex items-center gap-2 px-3 py-2 rounded text-sm text-text-muted hover:text-white hover:bg-border-color transition-colors">
                        <span className="material-symbols-outlined text-[18px]">person</span>
                        My Profile
                    </a>
                    <a href="#" className="flex items-center gap-2 px-3 py-2 rounded text-sm text-text-muted hover:text-white hover:bg-border-color transition-colors">
                        <span className="material-symbols-outlined text-[18px]">settings_suggest</span>
                        Organization Settings
                    </a>
                </div>
                <div className="border-t border-border-color p-2">
                    <button onClick={handleLogout} className="flex w-full items-center gap-2 px-3 py-2 rounded text-sm text-red-400 hover:bg-red-500/10 transition-colors">
                        <span className="material-symbols-outlined text-[18px]">logout</span>
                        Logout
                    </button>
                </div>
            </div>
            )}
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden relative">
        {/* Header */}
        <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20 shadow-sm relative">
          <div className="flex items-center gap-4">
            <h2 className="text-white text-lg font-bold tracking-wide">
               {[...menuItems, ...analyticsItems].find(i => i.path === location.pathname)?.label || 'Overview'}
            </h2>
          </div>
          <div className="flex items-center gap-6">
             {/* Search */}
             <div className="relative hidden md:block group z-50">
                <span className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-text-muted">
                   <span className="material-symbols-outlined text-[20px]">search</span>
                </span>
                <input 
                  type="text" 
                  onFocus={() => setIsSearchOpen(true)}
                  onBlur={() => setTimeout(() => setIsSearchOpen(false), 200)}
                  className="bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-10 pr-10 py-2 w-64 focus:ring-1 focus:ring-primary focus:border-primary placeholder-text-muted transition-all group-focus-within:w-96 shadow-inner" 
                  placeholder="Search vehicles, drivers..." 
                />
                <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                   <kbd className="hidden sm:inline-block h-5 px-1.5 border border-border-color rounded font-mono text-[10px] text-text-muted bg-surface-darker">/</kbd>
                </div>
                
                {/* Search Results */}
                {isSearchOpen && (
                <div className="absolute top-full left-0 mt-2 w-96 bg-surface-dark border border-border-color rounded-lg shadow-2xl overflow-hidden animate-in fade-in slide-in-from-top-2 duration-200">
                    <div className="p-2">
                        <div className="text-xs font-semibold text-text-muted px-3 py-2 uppercase tracking-wider">Vehicles (3)</div>
                        <Link to="/dashboard/vehicles/301" className="flex items-center justify-between px-3 py-2 hover:bg-border-color rounded group cursor-pointer">
                            <div className="flex items-center gap-3">
                                <div className="w-8 h-8 rounded bg-primary/20 flex items-center justify-center text-primary">
                                    <span className="material-symbols-outlined text-sm">directions_car</span>
                                </div>
                                <div>
                                    <p className="text-sm font-medium text-white group-hover:text-primary">Vehicle #301</p>
                                    <p className="text-xs text-text-muted">Active • I-95 South</p>
                                </div>
                            </div>
                            <span className="text-[10px] text-text-muted border border-border-color px-1.5 rounded">Enter</span>
                        </Link>
                        <div className="bg-surface-darker px-3 py-2 border-t border-border-color flex justify-between items-center text-[10px] text-text-muted mt-2 -mx-2 -mb-2">
                            <span>Press <kbd className="bg-surface-dark px-1 rounded border border-border-color">Enter</kbd> to select</span>
                            <span>View all 6 results</span>
                        </div>
                    </div>
                </div>
                )}
             </div>
             
             <div className="flex items-center gap-3">
                {/* Notifications */}
                <div className="relative">
                    <button 
                      onClick={() => setIsNotifOpen(!isNotifOpen)}
                      className={`relative p-2 rounded-full transition-colors ${isNotifOpen ? 'bg-primary text-white' : 'text-text-muted hover:text-white hover:bg-border-color'}`}>
                        <span className="material-symbols-outlined">notifications</span>
                        {unreadCount > 0 && <span className="absolute top-2 right-2 size-2.5 bg-red-500 rounded-full border-2 border-surface-darker"></span>}
                    </button>
                    
                    {isNotifOpen && (
                    <div className="absolute right-0 top-full mt-2 w-80 bg-surface-dark border border-border-color rounded-lg shadow-xl z-50 animate-in fade-in slide-in-from-top-2 duration-200">
                        <div className="p-3 border-b border-border-color flex justify-between items-center bg-surface-darker rounded-t-lg">
                            <span className="text-sm font-semibold text-white">Notifications</span>
                            <button className="text-[10px] text-primary hover:text-white transition-colors">Mark all as read</button>
                        </div>
                        <div className="max-h-64 overflow-y-auto custom-scrollbar">
                            <div className="p-3 border-b border-border-color hover:bg-border-color/30 cursor-pointer bg-border-color/10">
                                <div className="flex gap-3">
                                    <div className="flex-shrink-0 mt-1">
                                        <span className="material-symbols-outlined text-red-500 text-sm">error</span>
                                    </div>
                                    <div className="flex-1">
                                        <p className="text-xs font-medium text-white mb-0.5">Engine Overheat Alert</p>
                                        <p className="text-[11px] text-text-muted mb-1">Vehicle #301 - Coolant Temp High</p>
                                        <p className="text-[10px] text-text-muted/70">2 mins ago</p>
                                    </div>
                                    <div className="w-1.5 h-1.5 rounded-full bg-primary mt-2"></div>
                                </div>
                            </div>
                        </div>
                        <div className="p-2 border-t border-border-color bg-surface-darker rounded-b-lg text-center">
                            <Link to="/dashboard/alerts" className="text-xs text-primary hover:text-white font-medium block w-full">View All Notifications</Link>
                        </div>
                    </div>
                    )}
                </div>

                {/* Settings */}
                <div className="relative">
                   <button 
                     onClick={() => setIsSettingsOpen(!isSettingsOpen)}
                     className={`p-2 rounded-full transition-colors ${isSettingsOpen ? 'bg-primary text-white' : 'text-text-muted hover:text-white hover:bg-border-color'}`}>
                       <span className="material-symbols-outlined">settings</span>
                   </button>
                   {isSettingsOpen && (
                    <div className="absolute top-full right-0 mt-2 w-56 bg-surface-dark border border-border-color rounded-lg shadow-xl z-50 overflow-hidden animate-in fade-in slide-in-from-top-2 duration-200">
                        <div className="p-2 space-y-1">
                            <div className="px-3 py-2 text-xs font-semibold text-text-muted uppercase tracking-wider">System</div>
                            <button className="w-full flex items-center gap-3 px-3 py-2 rounded text-sm text-text-muted hover:text-white hover:bg-border-color transition-colors text-left">
                                <span className="material-symbols-outlined text-[18px]">display_settings</span>
                                Display
                            </button>
                            <button className="w-full flex items-center gap-3 px-3 py-2 rounded text-sm text-text-muted hover:text-white hover:bg-border-color transition-colors text-left">
                                <span className="material-symbols-outlined text-[18px]">manage_accounts</span>
                                Manage Users
                            </button>
                        </div>
                    </div>
                   )}
                </div>
             </div>
          </div>
        </header>

        {/* Content Scroll Area */}
        <main className="flex-1 overflow-y-auto custom-scrollbar bg-background-dark relative">
           <Outlet />
        </main>
      </div>
    </div>
  );
};

export default Layout;
