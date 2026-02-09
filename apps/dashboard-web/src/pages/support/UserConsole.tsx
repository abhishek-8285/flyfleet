import React, { useEffect, useState } from 'react';

const UserConsole: React.FC = () => {
  const [profile, setProfile] = useState<any>(null);
  const [sessions, setSessions] = useState<any[]>([]);

  useEffect(() => {
    const token = localStorage.getItem('token');
    const headers = { 'Authorization': `Bearer ${token}` };

    fetch('http://localhost:8081/api/support/users/profile', { headers })
      .then(res => res.ok ? res.json() : null)
      .then(data => setProfile(data));
    
    fetch('http://localhost:8081/api/support/users/sessions', { headers })
      .then(res => res.ok ? res.json() : [])
      .then(data => Array.isArray(data) ? setSessions(data) : setSessions([]))
      .catch(() => setSessions([]));
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <div className="flex items-center gap-2 text-[10px] text-[#92adc9] uppercase tracking-wider font-semibold mb-0.5">
              <span>Support</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">User Console</span>
            </div>
            <h2 className="text-white text-lg font-bold">User Troubleshooting: {profile?.name || '...'}</h2>
          </div>
        </div>
      </header>
      
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-12 gap-6">
          <div className="col-span-12 lg:col-span-4">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-0 overflow-hidden">
              <div className="p-5 border-b border-[#233648] bg-[#111a22]/50">
                <div className="flex items-start justify-between">
                  <div className="w-16 h-16 rounded-full bg-primary/20 border-2 border-primary text-primary flex items-center justify-center text-2xl font-bold">SJ</div>
                  <div className="flex flex-col items-end">
                    <span className="text-xs text-[#92adc9] uppercase font-bold tracking-wider">User ID</span>
                    <span className="text-white font-mono text-xs mt-1">{profile?.id}</span>
                  </div>
                </div>
                <div className="mt-4">
                  <h3 className="text-lg font-bold text-white">{profile?.name}</h3>
                  <p className="text-sm text-[#92adc9]">{profile?.email}</p>
                  <div className="flex items-center gap-2 mt-2">
                    <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-[10px] font-medium uppercase">{profile?.role}</span>
                  </div>
                </div>
              </div>
              <div className="p-5 space-y-4">
                <div>
                  <div className="text-[10px] uppercase text-[#92adc9] font-bold mb-2">Account Status</div>
                  <div className="flex items-center gap-2 text-sm text-white">
                    <span className="material-symbols-outlined text-emerald-500 text-[18px]">check_circle</span>
                    {profile?.status}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="col-span-12 lg:col-span-8">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden">
              <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                <h3 className="text-white font-semibold text-sm">Active Sessions</h3>
              </div>
              <div className="divide-y divide-[#233648]">
                {sessions.map((session, idx) => (
                  <div key={idx} className="p-4 hover:bg-[#111a22] transition-colors cursor-pointer">
                    <div className="flex items-center gap-3 mb-2">
                      <span className="material-symbols-outlined text-[#92adc9]">laptop_mac</span>
                      <span className="text-sm text-white font-medium">{session.device}</span>
                    </div>
                    <div className="flex flex-col gap-1 pl-9">
                      <div className="flex items-center gap-2 text-xs text-[#92adc9]">
                        <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                        <span>{session.location} • {session.ip}</span>
                      </div>
                      <div className="text-[10px] text-[#92adc9]/70">{session.browser} • {session.current ? 'Current Session' : 'Active'}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
};

export default UserConsole;
