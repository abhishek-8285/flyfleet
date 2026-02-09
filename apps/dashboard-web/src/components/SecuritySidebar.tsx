import React from 'react';
import { NavLink } from 'react-router-dom';

const SecuritySidebar: React.FC = () => {
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
    <aside className="w-64 h-full bg-slate-900/60 backdrop-blur-xl border-r border-white/5 flex flex-col z-20">
      <div className="p-6">
        <div className="flex items-center gap-3 px-2 mb-8">
           <div className="w-8 h-8 rounded-lg bg-vibrant-blue flex items-center justify-center shadow-lg shadow-blue-500/20">
              <span className="material-symbols-outlined text-white text-xl">shield</span>
           </div>
           <h1 className="text-xl font-black text-white tracking-tighter uppercase">FleetSecurity</h1>
        </div>

        <nav className="space-y-8">
           {groups.map((group) => (
             <div key={group.title}>
                <h3 className="px-4 text-[10px] font-bold uppercase tracking-[0.2em] text-slate-500 mb-4">{group.title}</h3>
                <div className="space-y-1">
                   {group.items.map((item) => (
                     <NavLink
                       key={item.path}
                       to={item.path}
                       className={({ isActive }) => `
                         flex items-center justify-between px-4 py-2.5 rounded-xl transition-all group
                         ${isActive 
                           ? 'bg-primary/10 text-white shadow-inner border border-white/5' 
                           : 'text-slate-400 hover:text-white hover:bg-white/5'}
                       `}
                     >
                        <div className="flex items-center gap-3">
                           <span className={`material-symbols-outlined text-[20px] transition-colors
                             ${item.badge === 'Active' ? 'text-vibrant-red' : 'group-hover:text-primary'}
                           `}>
                              {item.icon}
                           </span>
                           <span className="text-sm font-semibold">{item.name}</span>
                        </div>
                        {item.badge && (
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
