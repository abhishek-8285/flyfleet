import React, { useState, useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import { getSidebarCollapsed, setSidebarCollapsed } from '../utils/storage';

const SecuritySidebar: React.FC = () => {
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

  const groups = [
    {
      title: 'Monitoring',
      items: [
        { name: 'Identity Feed', path: '/security', icon: 'dashboard' },
        { name: 'Login Logs', path: '/security/logins', icon: 'history' },
        { name: 'Active Threats', path: '/security/threats', icon: 'warning', badge: 'Active' },
        { name: 'Live Sessions', path: '/security/sessions', icon: 'devices' },
      ]
    },
    {
      title: 'Governance',
      items: [
        { name: 'MFA Center', path: '/security/mfa', icon: 'fingerprint' },
        { name: 'IP Access', path: '/security/ip-access', icon: 'public' },
        { name: 'Access Review', path: '/security/access-review', icon: 'fact_check' },
      ]
    },
    {
      title: 'Configuration',
      items: [
        { name: 'MFA Policy', path: '/security/policy', icon: 'policy' },
        { name: 'Allowances', path: '/security/exceptions', icon: 'verified_user' },
        { name: 'Force Overrides', path: '/security/overrides', icon: 'security_update_warning' },
      ]
    }
  ];

  return (
    <aside className={`h-full bg-slate-900/60 backdrop-blur-xl border-r border-white/5 flex flex-col z-20 transition-all duration-300 relative ${collapsed ? 'w-20' : 'w-64'}`}>
      <div className={`p-6 ${collapsed ? 'px-2' : ''}`}>
        <div className={`flex items-center gap-3 px-2 mb-8 ${collapsed ? 'justify-center px-0' : ''}`}>
           <div className="w-8 h-8 rounded-lg bg-vibrant-blue flex items-center justify-center shadow-lg shadow-blue-500/20 flex-shrink-0">
              <span className="material-symbols-outlined text-white text-xl">shield</span>
           </div>
           {!collapsed && <h1 className="text-xl font-black text-white tracking-tighter uppercase">FleetSecurity</h1>}
        </div>

        {/* Toggle Button */}
        <button 
          onClick={toggleSidebar}
          className="absolute -right-3 top-20 bg-vibrant-blue text-white rounded-full size-6 flex items-center justify-center shadow-lg hover:opacity-90 transition-transform z-40 border border-white/10"
          title={collapsed ? "Expand Sidebar" : "Collapse Sidebar"}
        >
          <span className={`material-symbols-outlined text-sm transition-transform duration-300 ${collapsed ? 'rotate-180' : 'rotate-0'}`}>
            chevron_left
          </span>
        </button>

        <nav className={`space-y-8 ${collapsed ? 'px-2' : ''}`}>
           {groups.map((group) => (
             <div key={group.title}>
                <h3 className={`text-[10px] font-bold uppercase tracking-[0.2em] text-slate-500 mb-4 ${collapsed ? 'text-center px-2 text-[8px]' : 'px-4'}`}>
                  {collapsed ? '•••' : group.title}
                </h3>
                <div className="space-y-1">
                   {group.items.map((item) => (
                     <NavLink
                       key={item.path}
                       to={item.path}
                       title={collapsed ? item.name : ""}
                       className={({ isActive }) => `
                         flex items-center justify-between py-2.5 rounded-xl transition-all duration-200 group
                         ${collapsed 
                           ? 'justify-center px-0' 
                           : 'px-4'}
                         ${isActive 
                           ? 'bg-primary/10 text-white shadow-inner border border-white/5' 
                           : 'text-slate-400 hover:text-white hover:bg-white/5 border border-transparent'}
                       `}
                     >
                        <div className="flex items-center gap-3">
                           <span className={`material-symbols-outlined text-[20px] transition-colors flex-shrink-0
                             ${item.badge === 'Active' ? 'text-vibrant-red' : 'group-hover:text-primary'}
                           `}>
                              {item.icon}
                           </span>
                           {!collapsed && <span className="text-sm font-semibold animate-in fade-in slide-in-from-left-2 duration-300">{item.name}</span>}
                        </div>
                        {!collapsed && item.badge && (
                           <span className="bg-red-500/10 text-vibrant-red text-[8px] font-black uppercase px-1.5 py-0.5 rounded border border-red-500/20">
                              {item.badge}
                           </span>
                        )}
                     </NavLink>
                   ))}
                </div>
             </div>
           ))}
        </nav>
      </div>

      <div className="mt-auto p-6 border-t border-white/5 bg-slate-900/40">
         <div className="glass-card p-4 rounded-xl flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center">
               <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            </div>
            <div>
               <p className="text-[10px] font-bold text-white uppercase leading-none">Firewall Online</p>
               <p className="text-[9px] text-slate-500 mt-1">Nodes stable: 12/12</p>
            </div>
         </div>
      </div>
    </aside>
  );
};

export default SecuritySidebar;
