import {
    ArchiveBoxIcon,
    ArrowLeftOnRectangleIcon,
    ArrowPathIcon,
    ArrowUpTrayIcon,
    BuildingOffice2Icon,
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
];

export const Sidebar: React.FC = () => {
  const location = useLocation();
  const { can } = useRBAC();
  const { logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);

  const filteredItems = MENU_ITEMS.filter(item => can(item.action, item.resource));

  return (
    <div className={`flex flex-col h-screen bg-slate-900 text-white transition-all duration-300 ${collapsed ? 'w-20' : 'w-64'}`}>
      {/* Header */}
      <div className="flex items-center justify-center h-16 border-b border-slate-800">
        <h1 className={`font-bold text-xl tracking-wider ${collapsed ? 'hidden' : 'block'}`}>FLYFLEET</h1>
        {collapsed && <span className="font-bold text-xl">FF</span>}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-4">
        <ul className="space-y-1 px-2">
          {filteredItems.map((item) => {
            const isActive = location.pathname === item.path;
            return (
              <li key={item.path}>
                <Link
                  to={item.path}
                  className={`flex items-center px-4 py-3 rounded-lg transition-colors ${
                    isActive 
                      ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/50' 
                      : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                  }`}
                >
                  <item.icon className="h-6 w-6 shrink-0" />
                  {!collapsed && <span className="ml-3 font-medium">{item.name}</span>}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* Footer / User Profile */}
      <div className="border-t border-slate-800 p-4">
        <button 
          onClick={logout}
          className="flex items-center w-full px-4 py-2 text-sm text-red-400 hover:bg-red-900/20 rounded-lg transition-colors"
        >
          <ArrowLeftOnRectangleIcon className="h-5 w-5 shrink-0" />
          {!collapsed && <span className="ml-3">Logout</span>}
        </button>
      </div>
    </div>
  );
};
