import React, { useEffect, useState } from 'react';

const OrgConsole: React.FC = () => {
  const [org, setOrg] = useState<any>(null);

  useEffect(() => {
    const token = localStorage.getItem('token');
    fetch('/api/support/orgs/overview', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
      .then(res => res.ok ? res.json() : null)
      .then(data => setOrg(data));
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-white text-lg font-bold">Organization Console</h2>
            <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5">
              <span>SUPPORT</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">{org?.name}</span>
            </div>
          </div>
        </div>
      </header>
      
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Subscription Plan</div>
            <div className="text-2xl font-bold text-white mb-1">{org?.plan}</div>
            <div className="text-xs text-[#92adc9]">Next renewal: <span className="text-white">{org?.renewal}</span></div>
          </div>
          
          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Total Devices</div>
            <div className="text-2xl font-bold text-white mb-1">{org?.total_devices}</div>
            <div className="text-xs text-emerald-500 flex items-center gap-1 font-semibold">{org?.online_devices} Online</div>
          </div>

          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Integration Health</div>
            <div className="text-2xl font-bold text-emerald-500 mb-1">{org?.api_health}%</div>
            <div className="text-xs text-[#92adc9]">API Success Rate (24h)</div>
          </div>
        </div>

        <div className="bg-[#16202a] rounded-xl border border-[#233648] h-[400px] flex flex-col overflow-hidden relative">
          <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
            <h3 className="text-white text-sm font-semibold flex items-center gap-2">
              <span className="material-symbols-outlined text-primary text-[18px]">map</span>
              Real-time Fleet Distribution
            </h3>
          </div>
          <div className="flex-1 bg-[#1a2632] flex items-center justify-center">
            <span className="text-[#92adc9] text-sm">Interactive Map Load Pending...</span>
          </div>
        </div>
      </div>
    </main>
  );
};

export default OrgConsole;
