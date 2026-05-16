import React, { useState } from 'react';
import { useSelector } from 'react-redux';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthProvider';
import { useUIStrings } from '../../hooks/useUIStrings';
import { extractUserRole, isRoleAllowed } from '../../utils/rbac';
import type { RootState } from '../../store';

export const GlobalAdminSidebar: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { toggleSidebar, logout } = useAuth();
  const user = useSelector((state: RootState) => state.auth.user);
  const { uiStrings } = useUIStrings('layout_sidebar');
  const collapsed = user?.sidebar_collapsed ?? false;

  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  const effectiveUser = user || JSON.parse(localStorage.getItem('user_data') || 'null');
  const role = extractUserRole(effectiveUser);

  const canSee = (roles: string[]) => isRoleAllowed(role, roles);

  const [menuItems] = useState<any[]>(
    [
      { title: 'Dashboard', path: '/admin', icon: 'dashboard', roles: ['*'], items: [] },
      {
        title: 'Control Tower',
        path: '/admin/control-tower',
        icon: 'satellite_alt',
        roles: ['super_admin'],
        items: [],
      },
      {
        title: 'Dispatch',
        path: '/admin/dispatch',
        icon: 'support_agent',
        roles: ['*'],
        items: [],
      },
      {
        title: 'Vehicles',
        path: '/admin/vehicles',
        icon: 'local_shipping',
        roles: ['*'],
        items: [],
      },
      { title: 'Drivers', path: '/admin/drivers', icon: 'group', roles: ['*'], items: [] },
      { title: 'Trips', path: '/admin/trips', icon: 'route', roles: ['*'], items: [] },
      { title: 'Geofences', path: '/admin/geofences', icon: 'map', roles: ['*'], items: [] },
      {
        title: 'Customers',
        path: '/admin/customers',
        icon: 'account_balance',
        roles: ['admin', 'super_admin'],
        items: [],
      },
      {
        title: 'Maintenance',
        path: '/admin/maintenance/work-orders',
        icon: 'build',
        roles: ['*'],
        items: [],
      },
      {
        title: 'Finance',
        path: '/admin/finance/fuel',
        icon: 'attach_money',
        roles: ['admin', 'super_admin', 'fleet_manager'],
        items: [],
      },
      {
        title: 'Billing',
        path: '/admin/billing',
        icon: 'receipt_long',
        roles: ['admin', 'super_admin'],
        items: [],
      },
      {
        title: 'Tenants',
        path: '/admin/tenants',
        icon: 'business',
        roles: ['super_admin'],
        items: [],
      },
      {
        title: 'Users',
        path: '/admin/users',
        icon: 'group_add',
        roles: ['admin', 'super_admin'],
        items: [],
      },
      {
        title: 'Reports',
        path: '/admin/reports',
        icon: 'assessment',
        roles: ['admin', 'super_admin', 'fleet_manager'],
        items: [],
      },
      {
        title: 'Monitoring',
        path: '/admin/monitoring',
        icon: 'monitoring',
        roles: ['super_admin'],
        items: [],
      },
      {
        title: 'Settings',
        path: '/admin/settings/platform',
        icon: 'settings',
        roles: ['admin', 'super_admin'],
        items: [],
      },
      {
        title: 'Support',
        path: '/support',
        icon: 'help',
        roles: ['admin', 'super_admin'],
        items: [],
      },
    ].filter((item) => canSee(item.roles))
  );

  return (
    <aside
      className={`bg-surface-container-low border-r border-outline flex flex-col transition-all duration-300 ease-in-out shadow-sm z-[60] ${
        collapsed ? 'w-16' : 'w-64'
      }`}
    >
      {/* Header Branding */}
      <div className="h-16 flex items-center justify-between px-6 border-b border-outline">
        {!collapsed && (
          <div className="flex flex-col">
            <h1 className="text-white font-bold text-lg tracking-tight leading-tight">
              FlyFleet Command
            </h1>
            <p className="text-on-surface-variant text-[10px] uppercase tracking-wider font-bold leading-none mt-1">
              Fleet Management
            </p>
          </div>
        )}
        <button
          onClick={toggleSidebar}
          className={`p-2 rounded-lg hover:bg-surface-container-high text-on-surface-variant hover:text-white transition-colors ${collapsed ? 'mx-auto' : 'ml-auto'}`}
        >
          <span className="material-symbols-outlined text-lg">
            {collapsed ? 'menu_open' : 'menu'}
          </span>
        </button>
      </div>

      {/* Navigation Menu */}
      <nav className="flex-1 overflow-y-auto py-4 no-scrollbar">
        <ul className="space-y-1 px-3">
          {menuItems.map((item) => (
            <li key={item.path}>
              <Link
                to={item.path}
                className={`flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium transition-all group ${
                  isActive(item.path)
                    ? 'bg-surface-container-high text-white border-l-4 border-primary shadow-sm'
                    : 'text-on-surface-variant hover:text-white hover:bg-surface-container-high'
                }`}
              >
                <span
                  className={`material-symbols-outlined text-lg flex-shrink-0 ${isActive(item.path) ? 'text-white' : 'group-hover:text-white'}`}
                >
                  {item.icon}
                </span>
                {!collapsed && <span className="flex-1 tracking-tight">{item.title}</span>}
              </Link>
            </li>
          ))}
        </ul>
      </nav>

      {/* Quick Access Context Items (Support/Settings) */}
      <div className="pt-2 border-t border-outline">
        <button className="w-full flex items-center px-6 py-3 text-on-surface-variant hover:bg-surface-container-high hover:text-white transition-colors">
          <span className="material-symbols-outlined mr-3 text-lg">help</span>
          {!collapsed && <span className="text-sm font-medium">Support</span>}
        </button>
        <button className="w-full flex items-center px-6 py-3 text-on-surface-variant hover:bg-surface-container-high hover:text-white transition-colors">
          <span className="material-symbols-outlined mr-3 text-lg">settings</span>
          {!collapsed && <span className="text-sm font-medium">Settings</span>}
        </button>
      </div>

      {/* Profile Footer */}
      <div className="p-4 border-t border-outline bg-surface-container-lowest/30">
        <div className={`flex items-center gap-3 ${collapsed ? 'justify-center' : 'px-2'}`}>
          <div className="relative size-8 flex-shrink-0">
            <img
              src={
                effectiveUser?.avatar_url ||
                `https://ui-avatars.com/api/?name=${encodeURIComponent(effectiveUser?.name || 'F+M')}&background=137fec&color=fff`
              }
              alt={effectiveUser?.name || 'User'}
              className="size-full rounded-full object-cover border border-outline shadow-sm"
              onError={(e) => {
                // Fallback to a simple data-uri or local asset if CDN fails
                (e.target as HTMLImageElement).src =
                  `https://ui-avatars.com/api/?name=${encodeURIComponent(effectiveUser?.name || 'User')}&background=0D1117&color=FFFFFF`;
              }}
            />
          </div>
          {!collapsed && (
            <div className="flex-1 min-w-0">
              <p className="text-white text-xs font-bold truncate uppercase tracking-tight">
                {effectiveUser?.name || 'Fleet Manager'}
              </p>
              <p className="text-on-surface-variant text-[9px] truncate uppercase font-bold tracking-wider leading-none mt-1">
                {effectiveUser?.role?.replace('_', ' ') || 'Operations Manager'}
              </p>
            </div>
          )}

          <button
            onClick={logout}
            className="p-1.5 bg-surface-container hover:bg-surface-container-high rounded-lg text-on-surface-variant hover:text-error transition-all border border-outline ml-2 group active:scale-95"
            title="Logout"
          >
            <span className="material-symbols-outlined text-[18px] group-hover:rotate-12 transition-transform">
              power_settings_new
            </span>
          </button>
        </div>
      </div>
    </aside>
  );
};
