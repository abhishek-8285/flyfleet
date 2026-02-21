import React from "react";
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../../context/AuthProvider";

export const GlobalAdminSidebar: React.FC = () => {
  const location = useLocation();
  const { user, toggleSidebar } = useAuth();
  const collapsed = user?.sidebar_collapsed ?? false;

  const isActive = (path: string) => {
    return (
      location.pathname === path || location.pathname.startsWith(path + "/")
    );
  };

  const menuItems = [
    {
      title: "Dashboard",
      icon: "dashboard",
      path: "/admin/dashboard",
      items: [],
    },
    {
      title: "Organizations",
      icon: "corporate_fare",
      path: "/admin/organizations",
      items: [],
    },
    {
      title: "Users",
      icon: "group",
      path: "/admin/users",
      items: [],
    },
    {
      title: "Billing",
      icon: "payments",
      path: "/admin/billing",
      items: [],
    },
    {
      title: "Security",
      icon: "security",
      path: "/admin/security",
      items: [
        { title: "Access Review", path: "/admin/security/access-review" },
        { title: "Active Sessions", path: "/admin/security/sessions" },
        { title: "Login Activity", path: "/admin/security/login-activity" },
        { title: "Threat Alerts", path: "/admin/security/threats" },
      ],
    },
    {
      title: "Monitoring",
      icon: "monitoring",
      path: "/admin/monitoring",
      items: [
        { title: "System Health", path: "/admin/monitoring/health" },
        { title: "Performance", path: "/admin/monitoring/performance" },
        { title: "Logs", path: "/admin/monitoring/logs" },
        { title: "Alerts", path: "/admin/monitoring/alerts" },
      ],
    },
    {
      title: "Settings",
      icon: "settings",
      path: "/admin/settings",
      items: [
        { title: "Platform Config", path: "/admin/settings/platform" },
        { title: "Notifications", path: "/admin/settings/notifications" },
        { title: "Integrations", path: "/admin/settings/integrations" },
        { title: "Audit Settings", path: "/admin/settings/audit" },
      ],
    },
  ];

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
              <h1 className="text-white font-bold text-sm">Global Admin</h1>
              <p className="text-text-muted text-[10px] uppercase tracking-wider">
                Enterprise Portal
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
                  {item.items.map((subItem) => (
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
                Admin User
              </p>
              <p className="text-text-muted text-xs truncate">
                System Administrator
              </p>
            </div>
          )}
        </div>
      </div>
    </aside>
  );
};
