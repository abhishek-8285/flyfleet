import React, { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useSelector } from "react-redux";
import type { RootState } from "../../store";
import { useAuth } from "../../context/AuthProvider";
import { useUIStrings } from "../../hooks/useUIStrings";

export const GlobalAdminSidebar: React.FC = () => {
  const location = useLocation();
  const { toggleSidebar } = useAuth();
  const user = useSelector((state: RootState) => state.auth.user);
  const { uiStrings } = useUIStrings("layout_sidebar");
  const collapsed = user?.sidebar_collapsed ?? false;

  const isActive = (path: string) => {
    return (
      location.pathname === path || location.pathname.startsWith(path + "/")
    );
  };

  const effectiveUser = user || JSON.parse(localStorage.getItem('user_data') || 'null');
  const role = effectiveUser?.role?.toLowerCase() || '';

  /** Role-based menu visibility: which roles can see each menu item */
  const canSee = (roles: string[]) => roles.includes('*') || roles.includes(role);

  const [menuItems, setMenuitems] = useState<any[]>([
    { title: "Dashboard", path: "/admin", icon: "dashboard", roles: ["*"], items: [] },
    { title: "Control Tower", path: "/admin/control-tower", icon: "satellite_alt", roles: ["*"], items: [] },
    { title: "Dispatch", path: "/admin/dispatch", icon: "support_agent", roles: ["*"], items: [] },
    { title: "Vehicles", path: "/admin/vehicles", icon: "directions_car", roles: ["*"], items: [] },
    { title: "Drivers", path: "/admin/drivers", icon: "badge", roles: ["*"], items: [] },
    { title: "Trips", path: "/admin/trips", icon: "route", roles: ["*"], items: [] },
    { title: "Maintenance", path: "/admin/maintenance/work-orders", icon: "build", roles: ["*"], items: [] },
    { title: "Finance", path: "/admin/finance/fuel", icon: "attach_money", roles: ["admin", "super_admin", "fleet_manager"], items: [] },
    { title: "Billing", path: "/admin/billing", icon: "receipt_long", roles: ["admin", "super_admin"], items: [] },
    { title: "Organizations", path: "/admin/organizations", icon: "business", roles: ["admin", "super_admin"], items: [] },
    { title: "Users", path: "/admin/users", icon: "group", roles: ["admin", "super_admin"], items: [] },
    { title: "Reports", path: "/admin/reports", icon: "assessment", roles: ["admin", "super_admin", "fleet_manager"], items: [] },
    { title: "Monitoring", path: "/admin/monitoring", icon: "monitoring", roles: ["admin", "super_admin", "fleet_manager"], items: [] },
    { title: "Settings", path: "/admin/settings/platform", icon: "settings", roles: ["admin", "super_admin"], items: [] },
    { title: "Support", path: "/support", icon: "headset_mic", roles: ["admin", "super_admin"], items: [] },
  ].filter(item => canSee(item.roles)));

  return (
    <aside
      className={`bg-surface-dark border-r border-border-color flex flex-col transition-all duration-300 ease-in-out ${
        collapsed ? "w-16" : "w-64"
      }`}
    >
      {/* Header */}
      <div className="h-16 flex items-center justify-between px-4 border-b border-border-color">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <div className="bg-primary rounded-lg p-1">
              <span className="material-symbols-outlined text-white text-lg">
                admin_panel_settings
              </span>
            </div>
            <div>
              <h1 className="text-white font-bold text-sm">
                {uiStrings?.global_admin || "Global Admin"}
              </h1>
              <p className="text-text-muted text-[10px] uppercase tracking-wider">
                {uiStrings?.enterprise_portal || "Enterprise Portal"}
              </p>
            </div>
          </div>
        )}
        <button
          onClick={toggleSidebar}
          className="p-2 rounded-lg hover:bg-border-color text-text-muted hover:text-white transition-colors ml-auto"
        >
          <span className="material-symbols-outlined text-lg">
            {collapsed ? "menu_open" : "menu"}
          </span>
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-4">
        <ul className="space-y-1 px-2">
          {menuItems.map((item) => (
            <li key={item.path}>
              <Link
                to={item.path}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all ${
                  isActive(item.path)
                    ? "bg-primary text-white shadow-lg shadow-primary/20"
                    : "text-text-muted hover:text-white hover:bg-border-color"
                }`}
              >
                <span className="material-symbols-outlined text-lg flex-shrink-0">
                  {item.icon}
                </span>
                {!collapsed && (
                  <>
                    <span className="flex-1">{item.title}</span>
                    {item.items.length > 0 && (
                      <span className="material-symbols-outlined text-xs">
                        expand_more
                      </span>
                    )}
                  </>
                )}
              </Link>

              {/* Sub-items */}
              {!collapsed && item.items.length > 0 && isActive(item.path) && (
                <ul className="mt-1 ml-8 space-y-1">
                  {item.items.map((subItem: any) => (
                    <li key={subItem.path}>
                      <Link
                        to={subItem.path}
                        className={`block px-3 py-1.5 text-xs rounded transition-colors ${
                          isActive(subItem.path)
                            ? "text-primary font-medium"
                            : "text-text-muted hover:text-white"
                        }`}
                      >
                        {subItem.title}
                      </Link>
                    </li>
                  ))}
                </ul>
              )}
            </li>
          ))}
        </ul>
      </nav>

      {/* Footer */}
      <div className="p-4 border-t border-border-color">
        <div
          className={`flex items-center gap-3 ${collapsed ? "justify-center" : ""}`}
        >
          <div
            className="bg-center bg-no-repeat bg-cover rounded-full size-8 border-2 border-border-color"
            style={{
              backgroundImage:
                'url("https://ui-avatars.com/api/?name=A&background=137fec&color=fff")',
            }}
          ></div>
          {!collapsed && (
            <div className="flex-1 min-w-0">
              <p className="text-white text-sm font-medium truncate">
                {uiStrings?.user_admin || "Admin User"}
              </p>
              <p className="text-text-muted text-xs truncate">
                {uiStrings?.system_admin || "System Administrator"}
              </p>
            </div>
          )}
        </div>
      </div>
    </aside>
  );
};
