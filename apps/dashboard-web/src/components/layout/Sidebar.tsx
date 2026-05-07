import {
    ArchiveBoxIcon,
    ArrowLeftOnRectangleIcon,
    ArrowPathIcon,
    ArrowUpTrayIcon,
    BuildingOffice2Icon,
    Cog6ToothIcon,
    CpuChipIcon,
    MapIcon,
    ShieldCheckIcon,
    Squares2X2Icon,
    TruckIcon,
    UsersIcon,
    VideoCameraIcon,
} from '@heroicons/react/24/outline'; // Assuming heroicons is installed, or use alternatives
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthProvider';
import { useRBAC } from '../../context/RBACProvider';
import { useUIStrings } from '../../hooks/useUIStrings';
import type { RootState } from '../../store';

interface MenuItem {
  name: string;
  path: string;
  icon: React.ElementType;
  resource: string;
  action: string;
}

const IconMap: Record<string, React.ElementType> = {
  ArchiveBoxIcon,
  ArrowLeftOnRectangleIcon,
  ArrowPathIcon,
  ArrowUpTrayIcon,
  BuildingOffice2Icon,
  Cog6ToothIcon,
  CpuChipIcon,
  MapIcon,
  ShieldCheckIcon,
  Squares2X2Icon,
  TruckIcon,
  UsersIcon,
  VideoCameraIcon,
};

export const Sidebar: React.FC = () => {
  const location = useLocation();
  const { can } = useRBAC();
  const { logout, toggleSidebar } = useAuth();
  const user = useSelector((state: RootState) => state.auth.user);
  const { uiStrings } = useUIStrings('layout_sidebar');
  // Derive collapsed state from global context (DB-backed), fall back to false
  const collapsed = user?.sidebar_collapsed ?? false;

  const handleToggle = () => {
    toggleSidebar();
  };

  // Handle responsive behavior: override context collapsed state on small screens only visually
  const [forceCollapsed, setForceCollapsed] = useState(false);
  useEffect(() => {
    const handleResize = () => {
      setForceCollapsed(window.innerWidth < 1024);
    };
    window.addEventListener('resize', handleResize);
    handleResize();
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const isCollapsed = forceCollapsed || collapsed;

  const [menuItems, setMenuItems] = useState<MenuItem[]>([]);

  useEffect(() => {
    fetch('/api/v1/ui/menus/main_sidebar')
      .then((res) => res.json())
      .then((data) => {
        if (Array.isArray(data)) {
          const mapped = data.map((item: any) => ({
            ...item,
            icon: IconMap[item.icon] || Squares2X2Icon,
          }));
          setMenuItems(mapped);
        }
      })
      .catch((err) => console.error(err));
  }, []);

  // Temporary bypass for development - show all items if no permissions available
  const filteredItems =
    user && user.permissions && user.permissions.length > 0
      ? menuItems.filter((item) => can(item.action, item.resource))
      : menuItems;

  return (
    <aside
      className={`flex flex-col h-full bg-[#111a22] border-r border-border-color flex-shrink-0 z-30 transition-all duration-300 relative ${isCollapsed ? 'w-20' : 'w-64'}`}
    >
      <div className={`p-6 flex items-center gap-3 ${isCollapsed ? 'justify-center px-2' : ''}`}>
        <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(19,127,236,0.3)] flex-shrink-0">
          <span className="material-symbols-outlined">admin_panel_settings</span>
        </div>
        {!collapsed && (
          <h1 className="text-lg font-bold tracking-tight text-white truncate animate-in fade-in duration-300">
            {uiStrings?.brand_name || 'FlyFleet'}
          </h1>
        )}
      </div>

      {/* Toggle Button */}
      <button
        onClick={handleToggle}
        className="absolute -right-3 top-20 bg-primary text-white rounded-full size-6 flex items-center justify-center shadow-lg hover:bg-primary/90 transition-transform z-40 border border-white/10"
        title={
          isCollapsed
            ? uiStrings?.toggle_expand || 'Expand Sidebar'
            : uiStrings?.toggle_collapse || 'Collapse Sidebar'
        }
      >
        <span
          className={`material-symbols-outlined text-sm transition-transform duration-300 ${isCollapsed ? 'rotate-180' : 'rotate-0'}`}
        >
          chevron_left
        </span>
      </button>

      <nav
        className={`flex-1 space-y-1 overflow-y-auto custom-scrollbar ${isCollapsed ? 'px-2' : 'px-3'}`}
      >
        <div
          className={`mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted ${isCollapsed ? 'text-center text-[10px] scale-90 px-2' : 'px-3'}`}
        >
          {isCollapsed ? '•••' : uiStrings?.section_core || 'Core Modules'}
        </div>

        {filteredItems.map((item) => {
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              title={isCollapsed ? item.name : ''}
              className={`flex items-center gap-3 py-2.5 rounded-lg transition-all duration-200 group ${
                isCollapsed ? 'justify-center px-0' : 'px-3'
              } ${
                isActive
                  ? 'bg-primary/10 text-primary border-l-4 border-primary'
                  : 'text-text-muted hover:bg-border-color hover:text-white border-l-4 border-transparent'
              }`}
            >
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {!isCollapsed && (
                <span className="text-sm font-medium animate-in fade-in slide-in-from-left-2 duration-300">
                  {item.name}
                </span>
              )}
            </Link>
          );
        })}
      </nav>

      <div
        className={`p-4 border-t border-border-color bg-[#0f151b] transition-all duration-300 ${isCollapsed ? 'px-2' : ''}`}
      >
        <button
          onClick={logout}
          className={`flex items-center w-full text-sm text-text-muted hover:text-white hover:bg-border-color rounded-lg transition-colors gap-3 ${
            isCollapsed ? 'justify-center px-0 py-2.5' : 'px-3 py-2.5'
          }`}
        >
          <ArrowLeftOnRectangleIcon className="h-5 w-5 flex-shrink-0" />
          {!isCollapsed && <span>{uiStrings?.label_logout || 'Logout'}</span>}
        </button>
      </div>
    </aside>
  );
};
