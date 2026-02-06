import React, { useState } from 'react';
import { User, Organization } from '../types';

interface OrgSelectionScreenProps {
  user: User;
  onSignOut: () => void;
}

const OrgSelectionScreen: React.FC<OrgSelectionScreenProps> = ({ user, onSignOut }) => {
  const [selectedOrgId, setSelectedOrgId] = useState<string>('org-1');
  // Toggle for demo purposes to switch between empty and populated states
  const [isEmptyState, setIsEmptyState] = useState(false);

  const orgs: Organization[] = [
    { id: 'org-1', name: 'Acme Logistics Corp', role: 'Fleet Manager', lastUsed: true },
    { id: 'org-2', name: 'Global Freight Inc.', role: 'Viewer' },
    { id: 'org-3', name: 'Speedy Delivery Co.', role: 'Admin' },
    { id: 'org-4', name: 'Maritime Transport Ltd.', role: 'Fleet Analyst' },
  ];

  return (
    <div className="flex-1 flex flex-col min-h-screen">
      {/* Header */}
      <header className="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-6 py-4 lg:px-10 sticky top-0 z-10">
        <div className="flex items-center gap-4">
          <div className="text-primary size-8 flex items-center justify-center">
            <svg className="w-full h-full" fill="none" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
              <path d="M4 4H17.3334V17.3334H30.6666V30.6666H44V44H4V4Z" fill="currentColor"></path>
            </svg>
          </div>
          <h2 className="text-slate-900 dark:text-white text-lg font-bold leading-tight tracking-tight">Enterprise Fleet</h2>
        </div>
        <div className="flex items-center gap-4">
           {/* Demo Toggle */}
           <button 
             onClick={() => setIsEmptyState(!isEmptyState)}
             className="hidden lg:block text-xs bg-slate-800 text-slate-300 px-2 py-1 rounded border border-slate-700"
           >
             Toggle Empty State
           </button>
           
          <div className="hidden md:flex flex-col items-end mr-2">
            <span className="text-sm font-medium text-slate-900 dark:text-white">{user.name}</span>
            <button onClick={onSignOut} className="text-xs text-slate-500 dark:text-slate-400 hover:text-primary transition-colors">Sign Out</button>
          </div>
          <div 
            className="bg-center bg-no-repeat bg-cover rounded-full size-10 border border-slate-200 dark:border-slate-700"
            style={{ backgroundImage: `url("${user.avatarUrl}")` }}
            title={user.name}
          ></div>
        </div>
      </header>

      <main className="flex-1 flex justify-center py-8 px-4 sm:px-6">
        <div className="w-full max-w-[800px] flex flex-col">
          <div className="bg-white dark:bg-slate-900 rounded-xl shadow-sm border border-slate-200 dark:border-slate-800 flex flex-col overflow-hidden">
            
            {isEmptyState ? (
              // Empty State
              <div className="flex flex-col items-center justify-center py-16 px-6 text-center animate-in fade-in zoom-in duration-300">
                <div className="bg-slate-50 dark:bg-slate-800 p-6 rounded-full mb-6 ring-1 ring-slate-100 dark:ring-slate-700">
                  <span className="material-symbols-outlined text-slate-400 dark:text-slate-500 text-[64px]">domain_disabled</span>
                </div>
                <h1 className="text-2xl sm:text-3xl font-bold text-slate-900 dark:text-white mb-3">No organizations found</h1>
                <p className="text-slate-500 dark:text-slate-400 max-w-md mb-8 leading-relaxed">
                  It looks like you aren't associated with any fleets yet. Please contact your administrator or support.
                </p>
                <button className="bg-primary hover:bg-blue-600 text-white font-medium py-2.5 px-6 rounded-lg shadow-sm hover:shadow-md transition-all flex items-center gap-2">
                  Contact Support
                </button>
              </div>
            ) : (
              // Populated List State
              <div className="animate-in fade-in slide-in-from-bottom-2 duration-300">
                <div className="p-6 sm:p-8 pb-4">
                  <h1 className="text-2xl sm:text-[32px] font-bold text-slate-900 dark:text-white leading-tight mb-2">Select Organization</h1>
                  <p className="text-slate-500 dark:text-slate-400 text-sm sm:text-base">You are associated with multiple organizations. Choose one to proceed to your dashboard.</p>
                </div>

                <div className="px-6 sm:px-8 py-2">
                  <label className="relative flex w-full items-center">
                    <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-slate-400">
                      <span className="material-symbols-outlined">search</span>
                    </div>
                    <input 
                      type="text" 
                      placeholder="Search organizations..." 
                      className="w-full h-12 rounded-lg bg-slate-50 dark:bg-slate-800 border-none pl-10 pr-4 text-slate-900 dark:text-white placeholder:text-slate-400 focus:ring-2 focus:ring-primary transition-all text-base"
                    />
                  </label>
                </div>

                <div className="flex flex-col gap-3 p-6 sm:p-8 pt-4">
                  {orgs.map((org) => (
                    <label 
                      key={org.id}
                      className={`group relative flex cursor-pointer items-center gap-4 rounded-xl border p-4 transition-all hover:border-primary/50 hover:bg-slate-50 dark:hover:bg-slate-800/50 
                        ${selectedOrgId === org.id 
                          ? 'border-primary bg-blue-50/30 dark:bg-blue-900/10 ring-1 ring-primary' 
                          : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800'}`}
                    >
                      <input 
                        type="radio" 
                        name="org_select" 
                        className="h-5 w-5 border-slate-300 text-primary focus:ring-primary dark:border-slate-600 dark:bg-slate-700"
                        checked={selectedOrgId === org.id}
                        onChange={() => setSelectedOrgId(org.id)}
                      />
                      <div className="flex grow flex-col">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-slate-900 dark:text-white text-base font-semibold">{org.name}</span>
                          {org.lastUsed && (
                            <span className="inline-flex items-center rounded-md bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary ring-1 ring-inset ring-primary/20">Last Used</span>
                          )}
                        </div>
                        <p className="text-slate-500 dark:text-slate-400 text-sm">{org.role}</p>
                      </div>
                      <div className={`hidden sm:block ${selectedOrgId === org.id ? 'text-primary' : 'text-slate-400'}`}>
                        <span className="material-symbols-outlined">chevron_right</span>
                      </div>
                    </label>
                  ))}
                </div>

                {/* Pagination */}
                <div className="flex items-center justify-between border-t border-slate-100 dark:border-slate-800 p-4 sm:px-8">
                  <div className="hidden sm:flex flex-1 items-center justify-between">
                    <div>
                      <p className="text-sm text-slate-500 dark:text-slate-400">
                        Showing <span className="font-medium text-slate-900 dark:text-white">1</span> to <span className="font-medium text-slate-900 dark:text-white">4</span> of <span className="font-medium text-slate-900 dark:text-white">12</span> results
                      </p>
                    </div>
                    <div>
                      <nav className="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                        <a href="#" className="relative inline-flex items-center rounded-l-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-200 dark:ring-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 focus:z-20 focus:outline-offset-0">
                          <span className="sr-only">Previous</span>
                          <span className="material-symbols-outlined text-lg">chevron_left</span>
                        </a>
                        <a href="#" aria-current="page" className="relative z-10 inline-flex items-center bg-primary px-4 py-2 text-sm font-semibold text-white focus:z-20 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">1</a>
                        <a href="#" className="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-slate-900 dark:text-white ring-1 ring-inset ring-slate-200 dark:ring-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 focus:z-20 focus:outline-offset-0">2</a>
                        <a href="#" className="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-slate-900 dark:text-white ring-1 ring-inset ring-slate-200 dark:ring-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 focus:z-20 focus:outline-offset-0">3</a>
                        <a href="#" className="relative inline-flex items-center rounded-r-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-200 dark:ring-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 focus:z-20 focus:outline-offset-0">
                          <span className="sr-only">Next</span>
                          <span className="material-symbols-outlined text-lg">chevron_right</span>
                        </a>
                      </nav>
                    </div>
                  </div>
                  {/* Mobile Pagination */}
                  <div className="flex sm:hidden w-full justify-between">
                    <a href="#" className="relative inline-flex items-center rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Previous</a>
                    <a href="#" className="relative ml-3 inline-flex items-center rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Next</a>
                  </div>
                </div>

                {/* Footer Action */}
                <div className="bg-slate-50 dark:bg-slate-800/50 p-6 sm:px-8 border-t border-slate-200 dark:border-slate-800 flex justify-end">
                  <button className="bg-primary hover:bg-blue-600 text-white font-medium py-3 px-8 rounded-lg shadow-sm hover:shadow-md transition-all flex items-center gap-2 w-full sm:w-auto justify-center">
                    Enter Dashboard
                    <span className="material-symbols-outlined text-lg">arrow_forward</span>
                  </button>
                </div>
              </div>
            )}
          </div>

          <div className="mt-6 text-center">
            <p className="text-sm text-slate-400 dark:text-slate-500">Not seeing your organization? <a href="#" className="text-primary hover:underline">Contact Support</a></p>
          </div>
        </div>
      </main>
    </div>
  );
};

export default OrgSelectionScreen;