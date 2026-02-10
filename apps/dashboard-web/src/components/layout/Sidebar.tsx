import {
    ArchiveBoxIcon,
    ArrowLeftOnRectangleIcon,
    ArrowPathIcon,
    ArrowUpTrayIcon,
    BuildingOffice2Icon,
    Cog6ToothIcon,
    MapIcon,
    ShieldCheckIcon,
    Squares2X2Icon,
    TruckIcon,
    UsersIcon
} from '@heroicons/react/24/outline'; // Assuming heroicons is installed, or use alternatives
import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthProvider';
import { useRBAC } from '../../context/RBACProvider';

interface MenuItem {
  name: string;
  path: string;
  icon: React.ElementType;
  resource: string;
  action: string;
}

const MENU_ITEMS: MenuItem[] = [
  { name: 'Dashboard', path: '/admin', icon: Squares2X2Icon, resource: 'dashboard', action: 'read' },
  { name: 'Organizations', path: '/admin/organizations', icon: BuildingOffice2Icon, resource: 'org', action: 'read' },
  { name: 'Users', path: '/admin/users', icon: UsersIcon, resource: 'user', action: 'read' },
  { name: 'Roles & Permissions', path: '/admin/rbac/roles', icon: ShieldCheckIcon, resource: 'roles', action: 'read' },
  { name: 'Role Assignments', path: '/admin/rbac/assignments', icon: UsersIcon, resource: 'users', action: 'update' },
  { name: 'Security Logs', path: '/admin/rbac/logs', icon: ShieldCheckIcon, resource: 'audit_logs', action: 'read' },
  { name: 'Vehicles', path: '/admin/vehicles', icon: TruckIcon, resource: 'vehicle', action: 'read' },
  { name: 'Bulk Uploads', path: '/admin/bulk-uploads', icon: ArrowUpTrayIcon, resource: 'vehicle', action: 'create' },
  { name: 'Trips', path: '/admin/trips', icon: MapIcon, resource: 'trip', action: 'read' },
  { name: 'Drivers', path: '/admin/drivers', icon: UsersIcon, resource: 'driver', action: 'read' },
  { name: 'Devices', path: '/admin/devices', icon: Squares2X2Icon, resource: 'device', action: 'read' },
  { name: 'Live Map', path: '/admin/map', icon: MapIcon, resource: 'trip', action: 'read' },
  { name: 'Fuel Management', path: '/admin/fuel', icon: TruckIcon, resource: 'fuel', action: 'read' },
  { name: 'Governance', path: '/admin/governance', icon: ShieldCheckIcon, resource: 'governance', action: 'read' },
  { name: 'Disaster Recovery', path: '/admin/dr', icon: ArchiveBoxIcon, resource: 'disaster_recovery', action: 'read' },
  { name: 'Restore Operations', path: '/admin/dr/restores', icon: ArrowPathIcon, resource: 'disaster_recovery', action: 'read' },
  { name: 'Global Settings', path: '/admin/settings/platform', icon: Cog6ToothIcon, resource: 'org', action: 'update' },
];

export const Sidebar: React.FC = () => {
  const location = useLocation();
  const { can } = useRBAC();
  const { user, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);

  // Temporary bypass for development - show all items if no permissions available
  const filteredItems = (user && user.permissions && user.permissions.length > 0)
    ? MENU_ITEMS.filter(item => can(item.action, item.resource))
    : MENU_ITEMS;

  return (
    <aside className="flex flex-col w-64 h-full bg-[#111a22] border-r border-border-color flex-shrink-0 z-30">
      <div className="p-6 flex items-center gap-3">
        <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(19,127,236,0.3)]">
          <span className="material-symbols-outlined">admin_panel_settings</span>
        </div>
        <h1 className="text-lg font-bold tracking-tight text-white">FlyFleet</h1>
      </div>

      <nav className="flex-1 px-3 space-y-1 overflow-y-auto custom-scrollbar">
        <div className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted">Core Modules</div>
        
        {filteredItems.map((item) => {
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors ${
                isActive 
                  ? 'bg-primary/10 text-primary border-l-4 border-primary' 
                  : 'text-text-muted hover:bg-border-color hover:text-white'
              }`}
            >
              <item.icon className="h-5 w-5" />
              <span className="text-sm font-medium">{item.name}</span>
            </Link>
          );
        })}
      </nav>

      <div className="p-4 border-t border-border-color">
        <button 
          onClick={logout}
          className="flex items-center w-full px-3 py-2.5 text-sm text-text-muted hover:text-white hover:bg-border-color rounded-lg transition-colors gap-3"
        >
          <ArrowLeftOnRectangleIcon className="h-5 w-5" />
          <span>Logout</span>
        </button>
      </div>
    </aside>
  );
};
