import React, { useEffect, useState } from "react";
import { Link, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

export const CustomerPortalLayout: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("layout_portal");

  useEffect(() => {
    const token = localStorage.getItem("customer_token");
    if (!token) {
      navigate("/portal/login", { replace: true });
    }
  }, [navigate]);

  const [menuItems, setMenuitems] = useState<any[]>([]);
  useEffect(() => {
    fetch("/api/v1/ui/menus/menuItems")
      .then((res) => res.json())
      .then((data) => setMenuitems(data))
      .catch((err) => console.error(err));
  }, []);

  const handleLogout = () => {
    // Clear token
    localStorage.removeItem("customer_token");
    navigate("/portal/login");
  };

  return (
    <div className="flex h-screen w-full bg-background-dark text-white font-display overflow-hidden">
      {/* Sidebar */}
      <aside className="w-64 bg-surface-darker border-r border-border-color flex flex-col flex-shrink-0 z-30">
        <div className="h-16 flex items-center px-6 border-b border-border-color">
          <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white shadow-[0_0_15px_rgba(19,127,236,0.3)] mr-3">
            <span className="material-symbols-outlined text-[20px]">
              local_shipping
            </span>
          </div>
          <h1 className="text-lg font-bold tracking-tight text-white">
            {uiStrings?.portal_name || "FlyFleet Portal"}
          </h1>
        </div>

        <nav className="flex-1 overflow-y-auto p-4 space-y-1">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 px-3">
            {uiStrings?.label_menu || "Menu"}
          </div>
          {menuItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 group ${
                location.pathname === item.path ||
                location.pathname.startsWith(item.path + "/")
                  ? "bg-primary/10 text-primary border-l-4 border-primary"
                  : "text-text-muted hover:bg-border-color hover:text-white border-l-4 border-transparent"
              }`}
            >
              <span className="material-symbols-outlined text-[20px]">
                {item.icon}
              </span>
              <span className="text-sm font-medium">{item.label}</span>
            </Link>
          ))}
        </nav>

        <div className="p-4 border-t border-border-color bg-[#0f151b]">
          <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-border-color/50 transition-colors cursor-pointer group">
            <div className="w-9 h-9 rounded-full bg-surface-dark border border-border-color flex items-center justify-center text-text-muted font-bold">
              C
            </div>
            <div className="flex-1 overflow-hidden">
              <p className="text-sm font-medium text-white truncate group-hover:text-primary transition-colors">
                Acme Corp
              </p>
              <p className="text-xs text-text-muted truncate">
                {uiStrings?.account_customer || "Customer Account"}
              </p>
            </div>
            <button
              onClick={handleLogout}
              className="text-text-muted hover:text-red-400 transition-colors"
            >
              <span className="material-symbols-outlined text-[20px]">
                logout
              </span>
            </button>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden">
        <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
          <div className="flex items-center gap-4">
            <h2 className="text-white text-lg font-bold">
              {menuItems.find((i) => i.path === location.pathname)?.label ||
                uiStrings?.title_portal ||
                "Portal"}
            </h2>
          </div>
          <div className="flex items-center gap-4">
            <button className="p-2 text-text-muted hover:text-white transition-colors relative">
              <span className="material-symbols-outlined">notifications</span>
              <span className="absolute top-2 right-2 w-2 h-2 bg-red-500 rounded-full"></span>
            </button>
          </div>
        </header>

        <div className="flex-1 overflow-auto custom-scrollbar p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
};
