import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import NewTicketModal from '../../components/support/NewTicketModal';

const SupportDashboard: React.FC = () => {
  const [stats, setStats] = useState<any>(null);
  const [trends, setTrends] = useState<number[]>([]);
  const [activity, setActivity] = useState<any[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const fetchData = async () => {
    const token = localStorage.getItem('token');
    const headers = {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };

    try {
      const statsRes = await fetch('http://localhost:8081/api/support/dashboard/stats', { headers });
      if (statsRes.ok) setStats(await statsRes.json());
      
      const trendsRes = await fetch('http://localhost:8081/api/support/dashboard/trends', { headers });
      if (trendsRes.ok) setTrends(await trendsRes.json());

      const activityRes = await fetch('http://localhost:8081/api/support/dashboard/recent-activity', { headers });
      if (activityRes.ok) setActivity(await activityRes.json());
    } catch (err) {
      console.error('Failed to fetch dashboard data', err);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-white text-lg font-bold">Support Dashboard</h2>
            <div className="flex items-center gap-2 mt-0.5">
              <span className="text-[11px] text-[#92adc9] uppercase tracking-wider font-semibold">Support</span>
              <span className="text-[11px] text-[#92adc9]">/</span>
              <span className="text-[11px] text-white uppercase tracking-wider font-semibold">Dashboard</span>
            </div>
          </div>
          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-emerald-500/10 text-emerald-500 border border-emerald-500/30">
            Live Data
          </span>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3 px-3 py-1.5 rounded bg-[#16202a] border border-[#233648]">
            <span className="text-xs text-[#92adc9] font-medium">Auto-Refresh</span>
            <label className="relative inline-flex items-center cursor-pointer">
              <input defaultChecked className="sr-only peer" type="checkbox" />
              <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#137fec]"></div>
            </label>
          </div>
          <button 
            onClick={() => setIsModalOpen(true)}
            className="px-4 py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <span className="material-symbols-outlined text-[16px]">add</span>
            New Ticket
          </button>
        </div>
      </header>

      <NewTicketModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        onCreated={() => fetchData()} 
      />
      
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <Link to="/support/incidents" className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-primary/50 transition-all cursor-pointer">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-white">inbox</span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Open Tickets</div>
            <div className="text-2xl font-bold text-white group-hover:text-primary transition-colors">{stats?.open_tickets?.count || '...'}</div>
            <div className={`text-[10px] ${stats?.open_tickets?.trend < 0 ? 'text-emerald-500' : 'text-rose-500'} flex items-center gap-1 mt-1`}>
              <span className="material-symbols-outlined text-[12px]">{stats?.open_tickets?.trend < 0 ? 'trending_down' : 'trending_up'}</span>
              {stats?.open_tickets?.trend}% vs yesterday
            </div>
          </Link>
          
          <Link to="/support/reports" className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-rose-500/50 transition-all cursor-pointer">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-rose-500">timer_off</span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">SLA Breaches (24h)</div>
            <div className="text-2xl font-bold text-rose-500 group-hover:scale-105 transition-transform origin-left">{stats?.sla_breaches?.count || '...'}</div>
            <div className="text-[10px] text-rose-500 mt-1">{stats?.sla_breaches?.priority} Priority</div>
          </Link>

          <Link to="/support/logs" className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group hover:border-amber-500/50 transition-all cursor-pointer">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-amber-500">error</span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Active Device Errors</div>
            <div className="text-2xl font-bold text-amber-500 group-hover:scale-105 transition-transform origin-left">{stats?.active_errors?.count || '...'}</div>
            <div className="text-[10px] text-[#92adc9] mt-1">{stats?.active_errors?.scope}</div>
          </Link>

          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-emerald-500">av_timer</span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Avg. Resolution Time</div>
            <div className="text-2xl font-bold text-emerald-500">{stats?.avg_resolution?.time || '...'}</div>
            <div className="text-[10px] text-emerald-500 flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">trending_down</span>
              {stats?.avg_resolution?.trend}% faster this week
            </div>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-6 mb-6">
          <div className="col-span-12 lg:col-span-8">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 h-80 flex flex-col">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-[#137fec] text-[18px]">show_chart</span>
                  Incident Trends (Last 7 Days)
                </h3>
              </div>
              <div className="flex-1 relative w-full h-full">
                <svg className="w-full h-full" preserveAspectRatio="none" viewBox="0 0 800 200">
                  <defs>
                    <linearGradient id="chartGradient" x1="0" x2="0" y1="0" y2="1">
                      <stop offset="0%" stopColor="#137fec" stopOpacity="0.3"></stop>
                      <stop offset="100%" stopColor="#137fec" stopOpacity="0"></stop>
                    </linearGradient>
                  </defs>
                  <line stroke="#233648" strokeWidth="1" x1="0" x2="800" y1="199" y2="199"></line>
                  <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="150" y2="150"></line>
                  <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="100" y2="100"></line>
                  <line stroke="#233648" strokeDasharray="4" strokeWidth="1" x1="0" x2="800" y1="50" y2="50"></line>
                  {/* Dynamic path would go here based on trends state */}
                  <path d="M0,140 C100,160 150,80 250,90 S350,120 450,70 S600,60 800,40 V200 H0 Z" fill="url(#chartGradient)"></path>
                  <path d="M0,140 C100,160 150,80 250,90 S350,120 450,70 S600,60 800,40" fill="none" stroke="#137fec" strokeWidth="2"></path>
                </svg>
                <div className="absolute bottom-2 left-0 right-0 flex justify-between text-[10px] text-[#92adc9] px-2">
                  <span>Mon</span><span>Tue</span><span>Wed</span><span>Thu</span><span>Fri</span><span>Sat</span><span>Sun</span>
                </div>
              </div>
            </div>
          </div>
          
          <div className="col-span-12 lg:col-span-4">
             <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col h-80">
                <div className="p-4 border-b border-[#233648] flex justify-between items-center">
                  <h3 className="text-white text-sm font-semibold">Top System Errors</h3>
                </div>
                <div className="flex-1 p-4 flex flex-col justify-center gap-4">
                   {/* Hardcoded for now as per design mockup */}
                   <div className="w-full">
                      <div className="flex justify-between text-xs mb-1">
                        <span className="text-white">Connection Timeout (E-504)</span>
                        <span className="text-rose-500 font-bold">45</span>
                      </div>
                      <div className="w-full bg-[#111a22] h-2 rounded-full overflow-hidden">
                        <div className="bg-rose-500 h-full rounded-full" style={{width: '75%'}}></div>
                      </div>
                   </div>
                   <div className="w-full">
                      <div className="flex justify-between text-xs mb-1">
                        <span className="text-white">GPS Signal Lost (E-102)</span>
                        <span className="text-amber-500 font-bold">28</span>
                      </div>
                      <div className="w-full bg-[#111a22] h-2 rounded-full overflow-hidden">
                        <div className="bg-amber-500 h-full rounded-full" style={{width: '45%'}}></div>
                      </div>
                   </div>
                   {/* ... other errors ... */}
                </div>
             </div>
          </div>
        </div>

        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden">
          <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-[#92adc9] text-[18px]">history</span>
              Recent Support Activity
            </h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                  <th className="px-4 py-3 font-semibold">Ticket ID</th>
                  <th className="px-4 py-3 font-semibold">Subject</th>
                  <th className="px-4 py-3 font-semibold text-right">Status</th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {activity.map((item) => (
                  <tr 
                    key={item.id} 
                    className="border-b border-[#233648] hover:bg-[#111a22] transition-colors group cursor-pointer"
                  >
                    <td className="px-4 py-3 text-white font-mono text-xs">
                      <Link to={`/support/incidents/${item.id}`} className="hover:text-primary transition-colors">
                        #{item.id}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-white group-hover:text-primary transition-colors">
                      <Link to={`/support/incidents/${item.id}`} className="block w-full">
                        {item.subject}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Link to={`/support/incidents/${item.id}`}>
                        <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold ${
                          item.status === 'IN PROGRESS' ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20' : 'bg-amber-500/10 text-amber-500 border border-amber-500/20'
                        }`}>
                          {item.status}
                        </span>
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </main>
  );
};

export default SupportDashboard;
