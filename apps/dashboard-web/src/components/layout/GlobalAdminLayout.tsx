import React from 'react';
import { Outlet } from 'react-router-dom';
import { useUIStrings } from '../../hooks/useUIStrings';
import { GlobalAdminSidebar } from './GlobalAdminSidebar';

export const GlobalAdminLayout: React.FC = () => {
  const { uiStrings } = useUIStrings('layout_common');

  return (
    <div className="flex h-screen w-full bg-surface-dim text-on-surface font-body overflow-hidden">
      <GlobalAdminSidebar />

      <main className="flex-1 flex flex-col min-w-0 bg-surface-dim overflow-hidden relative">
        {/* Background Aura removed for professional aesthetic */}

        <header className="sticky top-0 z-[50] h-16 flex items-center justify-between px-8 border-b border-outline bg-surface/80 backdrop-blur-md flex-shrink-0 shadow-lg shadow-black/5">
          <div className="flex items-center gap-6 z-10">
            <div className="relative group">
              <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant text-[18px] group-hover:text-primary transition-colors">
                search
              </span>
              <input
                className="bg-surface-container border border-outline text-sm text-white rounded-lg pl-10 pr-4 py-1.5 w-80 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-on-surface-variant/50 shadow-inner"
                placeholder={
                  uiStrings?.search_global_placeholder || 'Search drivers, vehicles, or trips...'
                }
                type="text"
              />
            </div>
          </div>

          <div className="flex items-center gap-4 z-10">
            <button className="p-2 text-on-surface-variant hover:text-white transition-all relative">
              <span className="material-symbols-outlined">notifications</span>
              <span className="absolute top-2 right-2 w-1.5 h-1.5 bg-primary rounded-full ring-2 ring-surface"></span>
            </button>
          </div>
        </header>

        {/* Dynamic Content Reservoir */}
        <div className="flex-1 overflow-auto custom-scrollbar bg-surface-lowest">
          <Outlet />
        </div>
      </main>
    </div>
  );
};
