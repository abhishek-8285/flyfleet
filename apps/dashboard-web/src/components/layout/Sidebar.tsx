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
} from "@heroicons/react/24/outline"; // Assuming heroicons is installed, or use alternatives
import React, { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../../context/AuthProvider";
import { useRBAC } from "../../context/RBACProvider";

interface MenuItem {
  name: string;
  path: string;
  icon: React.ElementType;
  resource: string;
  action: string;
}

const MENU_ITEMS: MenuItem[] = [
  {
    name: "Dashboard",
    path: "/admin",
    icon: Squares2X2Icon,
    resource: "dashboard",
    action: "read",
  },
  {
    name: "Command Center",
    path: "/dashboard/command-center",
    icon: MapIcon,
    resource: "trip",
    action: "read",
  },
  {
    name: "Organizations",
    path: "/admin/organizations",
    icon: BuildingOffice2Icon,
    resource: "org",
    action: "read",
  },
  {
    name: "Users",
    path: "/admin/users",
    icon: UsersIcon,
    resource: "user",
    action: "read",
  },
  {
    name: "Roles & Permissions",
    path: "/admin/rbac/roles",
    icon: ShieldCheckIcon,
    resource: "roles",
    action: "read",
  },
  {
    name: "Role Assignments",
    path: "/admin/rbac/assignments",
    icon: UsersIcon,
    resource: "users",
    action: "update",
  },
  {
    name: "Security Logs",
    path: "/admin/rbac/logs",
    icon: ShieldCheckIcon,
    resource: "audit_logs",
    action: "read",
  },
  {
    name: "Vehicles",
    path: "/admin/vehicles",
    icon: TruckIcon,
    resource: "vehicle",
    action: "read",
  },
  {
    name: "Bulk Uploads",
    path: "/admin/bulk-uploads",
    icon: ArrowUpTrayIcon,
    resource: "vehicle",
    action: "create",
  },
  {
    name: "Trips",
    path: "/admin/trips",
    icon: MapIcon,
    resource: "trip",
    action: "read",
  },
  {
    name: "Route Optimization",
    path: "/admin/route-optimization",
    icon: ArrowPathIcon,
    resource: "trip",
    action: "create",
  },
  {
    name: "Video Telematics",
    path: "/admin/video",
    icon: VideoCameraIcon,
    resource: "video",
    action: "read",
  },
  {
    name: "Drivers",
    path: "/admin/drivers",
    icon: UsersIcon,
    resource: "driver",
    action: "read",
  },
  {
    name: "Devices",
    path: "/admin/devices",
    icon: Squares2X2Icon,
    resource: "device",
    action: "read",
  },
  {
    name: "Live Map",
    path: "/admin/map",
    icon: MapIcon,
    resource: "trip",
    action: "read",
  },
  {
    name: "Fuel Management",
    path: "/admin/fuel",
    icon: TruckIcon,
    resource: "fuel",
    action: "read",
  },
  {
    name: "Diagnostics & AI",
    path: "/dashboard/insights",
    icon: CpuChipIcon,
    resource: "dashboard",
    action: "read",
  },
  {
    name: "Governance",
    path: "/admin/governance",
    icon: ShieldCheckIcon,
    resource: "governance",
    action: "read",
  },
  {
    name: "Disaster Recovery",
    path: "/admin/dr",
    icon: ArchiveBoxIcon,
    resource: "disaster_recovery",
    action: "read",
  },
  {
    name: "Restore Operations",
    path: "/admin/dr/restores",
    icon: ArrowPathIcon,
    resource: "disaster_recovery",
    action: "read",
  },
  {
    name: "Global Settings",
    path: "/admin/settings/platform",
    icon: Cog6ToothIcon,
    resource: "org",
    action: "update",
  },
];

export const Sidebar: React.FC = () => {
  const location = useLocation();
  const { can } = useRBAC();
  const { user, logout, toggleSidebar } = useAuth();
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
    window.addEventListener("resize", handleResize);
    handleResize();
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const isCollapsed = forceCollapsed || collapsed;

  // Temporary bypass for development - show all items if no permissions available
  const filteredItems =
    user && user.permissions && user.permissions.length > 0
      ? MENU_ITEMS.filter((item) => can(item.action, item.resource))
      : MENU_ITEMS;

  return (
    <aside
      className={`flex flex-col h-full bg-[#111a22] border-r border-border-color flex-shrink-0 z-30 transition-all duration-300 relative ${isCollapsed ? "w-20" : "w-64"}`}
    >
      <div
        className={`p-6 flex items-center gap-3 ${isCollapsed ? "justify-center px-2" : ""}`}
      >
        <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(19,127,236,0.3)] flex-shrink-0">
          <span className="material-symbols-outlined">
            admin_panel_settings
          </span>
        </div>
        {!collapsed && (
          <h1 className="text-lg font-bold tracking-tight text-white truncate animate-in fade-in duration-300">
            FlyFleet
          </h1>
        )}
      </div>

      {/* Toggle Button */}
      <button
        onClick={handleToggle}
        className="absolute -right-3 top-20 bg-primary text-white rounded-full size-6 flex items-center justify-center shadow-lg hover:bg-primary/90 transition-transform z-40 border border-white/10"
        title={isCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
      >
        <span
          className={`material-symbols-outlined text-sm transition-transform duration-300 ${isCollapsed ? "rotate-180" : "rotate-0"}`}
        >
          chevron_left
        </span>
      </button>

      <nav
        className={`flex-1 space-y-1 overflow-y-auto custom-scrollbar ${isCollapsed ? "px-2" : "px-3"}`}
      >
        <div
          className={`mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted ${isCollapsed ? "text-center text-[10px] scale-90 px-2" : "px-3"}`}
        >
          {isCollapsed ? "•••" : "Core Modules"}
        </div>

        {filteredItems.map((item) => {
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              title={isCollapsed ? item.name : ""}
              className={`flex items-center gap-3 py-2.5 rounded-lg transition-all duration-200 group ${
                isCollapsed ? "justify-center px-0" : "px-3"
              } ${
                isActive
                  ? "bg-primary/10 text-primary border-l-4 border-primary"
                  : "text-text-muted hover:bg-border-color hover:text-white border-l-4 border-transparent"
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
        className={`p-4 border-t border-border-color bg-[#0f151b] transition-all duration-300 ${isCollapsed ? "px-2" : ""}`}
      >
        <button
          onClick={logout}
          className={`flex items-center w-full text-sm text-text-muted hover:text-white hover:bg-border-color rounded-lg transition-colors gap-3 ${
            isCollapsed ? "justify-center px-0 py-2.5" : "px-3 py-2.5"
          }`}
        >
          <ArrowLeftOnRectangleIcon className="h-5 w-5 flex-shrink-0" />
          {!isCollapsed && <span>Logout</span>}
        </button>
      </div>
    </aside>
  );
};
