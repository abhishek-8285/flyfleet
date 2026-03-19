import React from "react";
import { Outlet } from "react-router-dom";
import { useAuth } from "../../context/AuthProvider";
import { useUIStrings } from "../../hooks/useUIStrings";
import { GlobalAdminSidebar } from "./GlobalAdminSidebar";

export const GlobalAdminLayout: React.FC = () => {
  const { uiStrings } = useUIStrings("layout_common");
  const { logout } = useAuth();
  return (
    <div className="flex h-screen w-full bg-background-dark text-white font-display overflow-hidden">
      <GlobalAdminSidebar />
      <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden">
        <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
          <div className="flex items-center gap-4">
            <div>
              <h2 className="text-white text-lg font-bold">
                {uiStrings?.global_title || "FlyFleet Global Admin"}
              </h2>
              <div className="flex items-center gap-2 text-[11px] text-text-muted mt-0.5 uppercase tracking-wide font-medium">
                <span>
                  {uiStrings?.command_center || "Enterprise Command Center"}
                </span>
                <span className="material-symbols-outlined text-[10px]">
                  chevron_right
                </span>
                <span className="text-white">
                  {uiStrings?.dashboard || "Dashboard"}
                </span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="relative">
              <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-[18px]">
                search
              </span>
              <input
                className="bg-surface-dark border border-border-color text-xs text-white rounded pl-9 pr-3 py-1.5 w-64 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-text-muted/50"
                placeholder={
                  uiStrings?.search_global_placeholder ||
                  "Search global admin..."
                }
                type="text"
              />
            </div>
            <div className="flex items-center gap-3">
              <div
                className="bg-center bg-no-repeat bg-cover rounded-full size-8 border-2 border-border-color"
                style={{
                  backgroundImage:
                    'url("https://ui-avatars.com/api/?name=A&background=137fec&color=fff")',
                }}
              ></div>
              <div className="flex flex-col overflow-hidden">
                <span className="text-sm font-medium text-white truncate">
                  {uiStrings?.user_global || "Global Admin"}
                </span>
                <span className="text-xs text-text-muted truncate">
                  {uiStrings?.system_admin || "System Administrator"}
                </span>
              </div>
              <button
                onClick={logout}
                className="ml-auto text-text-muted hover:text-white"
              >
                <span className="material-symbols-outlined text-[20px]">
                  logout
                </span>
              </button>
            </div>
          </div>
        </header>
        <div className="flex-1 overflow-auto custom-scrollbar p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
};
