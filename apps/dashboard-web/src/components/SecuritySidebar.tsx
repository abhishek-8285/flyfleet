import React, { useEffect, useState } from "react";
import { NavLink } from "react-router-dom";
import { useUIStrings } from "../hooks/useUIStrings";
import { getSidebarCollapsed, setSidebarCollapsed } from "../utils/storage";

const SecuritySidebar: React.FC = () => {
  const { uiStrings } = useUIStrings("layout_security");
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

    window.addEventListener("resize", handleResize);
    handleResize();

    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const [groups, setGroups] = useState<any[]>([]);
  useEffect(() => {
    fetch("/api/v1/ui/menus/groups")
      .then((res) => res.json())
      .then((data) => setGroups(data))
      .catch((err) => console.error(err));
  }, []);

  return (
    <aside
      className={`h-full bg-slate-900/60 backdrop-blur-xl border-r border-white/5 flex flex-col z-20 transition-all duration-300 relative ${collapsed ? "w-20" : "w-64"}`}
    >
      <div className={`p-6 ${collapsed ? "px-2" : ""}`}>
        <div
          className={`flex items-center gap-3 px-2 mb-8 ${collapsed ? "justify-center px-0" : ""}`}
        >
          <div className="w-8 h-8 rounded-lg bg-vibrant-blue flex items-center justify-center shadow-lg shadow-blue-500/20 flex-shrink-0">
            <span className="material-symbols-outlined text-white text-xl">
              shield
            </span>
          </div>
          {!collapsed && (
            <h1 className="text-xl font-black text-white tracking-tighter uppercase">
              {uiStrings?.brand_security || "FleetSecurity"}
            </h1>
          )}
        </div>

        {/* Toggle Button */}
        <button
          onClick={toggleSidebar}
          className="absolute -right-3 top-20 bg-vibrant-blue text-white rounded-full size-6 flex items-center justify-center shadow-lg hover:opacity-90 transition-transform z-40 border border-white/10"
          title={
            collapsed
              ? uiStrings?.toggle_expand || "Expand Sidebar"
              : uiStrings?.toggle_collapse || "Collapse Sidebar"
          }
        >
          <span
            className={`material-symbols-outlined text-sm transition-transform duration-300 ${collapsed ? "rotate-180" : "rotate-0"}`}
          >
            chevron_left
          </span>
        </button>

        <nav className={`space-y-8 ${collapsed ? "px-2" : ""}`}>
          {groups.map((group) => (
            <div key={group.title}>
              <h3
                className={`text-[10px] font-bold uppercase tracking-[0.2em] text-slate-500 mb-4 ${collapsed ? "text-center px-2 text-[8px]" : "px-4"}`}
              >
                {collapsed ? "•••" : group.title}
              </h3>
              <div className="space-y-1">
                {group.items.map((item: any) => (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    title={collapsed ? item.name : ""}
                    className={({ isActive }) => `
                         flex items-center justify-between py-2.5 rounded-xl transition-all duration-200 group
                         ${collapsed ? "justify-center px-0" : "px-4"}
                         ${
                           isActive
                             ? "bg-primary/10 text-white shadow-inner border border-white/5"
                             : "text-slate-400 hover:text-white hover:bg-white/5 border border-transparent"
                         }
                       `}
                  >
                    <div className="flex items-center gap-3">
                      <span
                        className={`material-symbols-outlined text-[20px] transition-colors flex-shrink-0
                             ${item.badge === "Active" ? "text-vibrant-red" : "group-hover:text-primary"}
                           `}
                      >
                        {item.icon}
                      </span>
                      {!collapsed && (
                        <span className="text-sm font-semibold animate-in fade-in slide-in-from-left-2 duration-300">
                          {item.name}
                        </span>
                      )}
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
            <p className="text-[10px] font-bold text-white uppercase leading-none">
              {uiStrings?.status_firewall || "Firewall Online"}
            </p>
            <p className="text-[9px] text-slate-500 mt-1">
              {(uiStrings?.label_nodes || "Nodes stable: {{count}}/{{total}}")
                .replace("{{count}}", "12")
                .replace("{{total}}", "12")}
            </p>
          </div>
        </div>
      </div>
    </aside>
  );
};

export default SecuritySidebar;
