import { useEffect, useState } from 'react';

// Types
interface TimelineEvent {
  id: string;
  type: string;
  title: string;
  description: string;
  severity: string;
  timestamp: string;
  actor: string;
  actor_type: string;
  metadata: Record<string, any>;
}

interface TimelineStat {
  label: string;
  value: string;
  trend: string;
  status: string;
  icon: string;
}

const TimelinePage = () => {
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [stats, setStats] = useState<TimelineStat[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [eventsRes, statsRes] = await Promise.all([
            fetch('http://localhost:8081/api/timeline'),
            fetch('http://localhost:8081/api/timeline/stats')
        ]);
        
        const eventsData = await eventsRes.json();
        const statsData = await statsRes.json();

        if (eventsData.data) setEvents(eventsData.data);
        if (statsData) setStats(statsData);
        
      } catch (error) {
        console.error("Failed to fetch timeline data:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const getSeverityIcon = (severity: string) => {
      switch(severity) {
          case 'critical': return 'warning';
          case 'warning': return 'sensors';
          case 'success': return 'route';
          case 'info': return 'info';
          default: return 'circle';
      }
  }

  const getSeverityColor = (severity: string) => {
      switch(severity) {
          case 'critical': return 'text-danger border-danger/30';
          case 'warning': return 'text-warning border-warning/30';
          case 'success': return 'text-success border-success/30';
          case 'info': return 'text-primary border-primary/30';
          default: return 'text-text-muted border-border-color';
      }
  }

  return (
    <div className="flex h-screen w-full">
      <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden">
        <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
          <div className="flex items-center gap-4">
            <div>
              <div className="flex items-center gap-2 text-[10px] text-text-muted font-bold uppercase tracking-wider mb-0.5">
                <span>Platform</span>
                <span className="text-border-color">/</span>
                <span className="text-primary">Activity Timeline</span>
              </div>
              <h2 className="text-white text-lg font-bold">Unified Activity Timeline</h2>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="relative">
              <span className="absolute inset-y-0 left-0 flex items-center pl-3">
                <span className="material-symbols-outlined text-text-muted text-[18px]">search</span>
              </span>
              <input className="bg-surface-dark border border-border-color text-sm text-white rounded-lg pl-10 pr-4 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary outline-none w-64 placeholder-text-muted/60" placeholder="Global Search..." type="text"/>
            </div>
            <button className="px-3 py-1.5 bg-surface-dark border border-border-color hover:bg-surface-darker text-text-muted hover:text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors">
              <span className="material-symbols-outlined text-[16px]">refresh</span>
              Refresh Feed
            </button>
          </div>
        </header>
        
        <div className="flex-1 overflow-hidden flex flex-col md:flex-row">
          <div className="flex-1 flex flex-col overflow-hidden">
            {/* Stats Check */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 p-6 pb-2 flex-shrink-0">
                {stats.map((stat, idx) => (
                    <div key={idx} className="bg-surface-dark rounded-xl border border-border-color p-4 relative overflow-hidden group">
                        <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                        <span className={`material-symbols-outlined text-4xl ${stat.status === 'danger' ? 'text-danger' : 'text-primary'}`}>{stat.icon}</span>
                        </div>
                        <div className="text-text-muted text-xs font-medium uppercase mb-1">{stat.label}</div>
                        <div className={`text-2xl font-bold ${stat.status === 'danger' ? 'text-danger' : 'text-white'}`}>{stat.value}</div>
                        <div className={`text-[10px] flex items-center gap-1 mt-1 ${stat.status === 'danger' ? 'text-danger' : 'text-text-muted'}`}>
                        {stat.trend}
                        </div>
                    </div>
                ))}
            </div>
            
            <div className="flex-1 overflow-y-auto custom-scrollbar p-6 pt-2">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-white font-semibold text-sm">Today</h3>
                <span className="text-xs text-text-muted">October 24, 2023</span>
              </div>
              <div className="space-y-0 relative pb-10">
                {loading ? (
                    <div className="text-center text-text-muted">Loading events...</div>
                ) : (
                    events.map((event) => (
                        <div key={event.id} className="timeline-item relative pl-16 py-3 group">
                            <div className="absolute top-3.5 bottom-[-1rem] left-6 w-[2px] bg-border-color z-0 last:hidden"></div>
                            <div className={`absolute left-0 top-3 w-12 h-12 rounded-full bg-surface-darker border flex items-center justify-center z-10 ${getSeverityColor(event.severity)}`}>
                                <span className="material-symbols-outlined">{getSeverityIcon(event.severity)}</span>
                            </div>
                            <div className={`bg-surface-dark rounded-xl border border-border-color p-4 hover:border-opacity-100 transition-colors ${getSeverityColor(event.severity).replace('text-', 'hover:border-')}`}>
                                <div className="flex justify-between items-start mb-2">
                                <div className="flex items-center gap-2">
                                    <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-opacity-10 border border-opacity-20 ${getSeverityColor(event.severity).replace('border-', 'bg-').replace('border-', 'border-')}`}>
                                        {event.type}
                                    </span>
                                    <span className="text-xs text-text-muted">{new Date(event.timestamp).toLocaleTimeString()}</span>
                                </div>
                                <button className="text-primary text-xs hover:underline flex items-center gap-1">
                                    View Details <span className="material-symbols-outlined text-[12px]">open_in_new</span>
                                </button>
                                </div>
                                <h4 className="text-white text-sm font-semibold mb-1">{event.title}</h4>
                                <p className="text-text-muted text-xs leading-relaxed mb-3">{event.description}</p>
                                <div className="flex items-center gap-3 mt-2 text-[11px] text-text-muted bg-surface-darker/50 p-2 rounded border border-border-color/50">
                                <span className="flex items-center gap-1"><span className="material-symbols-outlined text-[12px]">person</span> {event.actor}</span>
                                {event.metadata && event.metadata.location && (
                                    <span className="flex items-center gap-1"><span className="material-symbols-outlined text-[12px]">location_on</span> {event.metadata.location}</span>
                                )}
                                </div>
                            </div>
                        </div>
                    ))
                )}
                
                <div className="flex justify-center py-6">
                  <div className="flex items-center gap-2 text-text-muted text-xs">
                    <span className="material-symbols-outlined animate-spin text-[16px]">progress_activity</span>
                    Loading older events...
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <aside className="w-full md:w-80 bg-surface-dark border-l border-border-color flex flex-col flex-shrink-0 z-20 shadow-xl">
            <div className="p-4 border-b border-border-color flex items-center justify-between">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-[18px]">filter_alt</span>
                Timeline Filters
              </h3>
              <button className="text-[10px] text-primary hover:text-primary-dark font-medium uppercase tracking-wide">Reset All</button>
            </div>
            <div className="p-4 flex-1 overflow-y-auto custom-scrollbar space-y-6">
                {/* Mock Filters - Visual Only */}
              <div>
                <label className="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">Event Types</label>
                <div className="space-y-2">
                  <label className="flex items-center gap-2 cursor-pointer group">
                    <input defaultChecked className="form-checkbox rounded bg-surface-darker border-border-color text-primary focus:ring-primary focus:ring-offset-surface-dark h-4 w-4" type="checkbox"/>
                    <span className="text-sm text-text-muted group-hover:text-white transition-colors">Critical Alerts</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer group">
                    <input defaultChecked className="form-checkbox rounded bg-surface-darker border-border-color text-primary focus:ring-primary focus:ring-offset-surface-dark h-4 w-4" type="checkbox"/>
                    <span className="text-sm text-text-muted group-hover:text-white transition-colors">Trip Events</span>
                  </label>
                </div>
              </div>
            </div>
            <div className="p-4 border-t border-border-color bg-surface-darker">
              <button className="w-full py-2 bg-primary hover:bg-primary-dark text-white text-sm font-semibold rounded flex items-center justify-center gap-2 transition-colors shadow-lg shadow-primary/20">
                Apply Filters
              </button>
            </div>
          </aside>
        </div>
        
        <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
          <span>© 2023 Enterprise Fleet Manager. Platform V2.4</span>
          <div className="flex items-center gap-4">
            <span>System Status: <span className="text-success">Operational</span></span>
            <span>Data Lag: <span className="font-mono text-white">~45ms</span></span>
          </div>
        </footer>
      </main>
    </div>
  );
};

export default TimelinePage;
