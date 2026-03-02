import React, { useEffect, useState } from "react";
import NewTicketModal from "../../components/support/NewTicketModal";
import { useUIStrings } from "../../hooks/useUIStrings";

const toNumber = (value: any): number => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
};

const SupportDashboard: React.FC = () => {
  const { uiStrings } = useUIStrings("support_v2");
  const { uiStrings: commonStrings } = useUIStrings("common");

  const [metrics, setMetrics] = useState<any>(null);
  const [trendPoints, setTrendPoints] = useState<
    Array<{ label: string; value: number }>
  >([]);
  const [topErrors, setTopErrors] = useState<any[]>([]);
  const [recentTickets, setRecentTickets] = useState<any[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [systemMeta, setSystemMeta] = useState<any>(null);

  const fetchData = async () => {
    const token = localStorage.getItem("access_token");
    const headers = {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    };

    try {
      const [slasRes, ticketsRes] = await Promise.all([
        fetch("/api/v1/support/slas", { headers }),
        fetch("/api/v1/support/tickets", { headers }),
      ]);

      if (slasRes.ok) {
        const data = await slasRes.json();
        const payload = data.data || data;
        setMetrics({
          openTickets: payload.active_incidents || 0,
          slaBreaches24h: Math.floor(
            (100 - (payload.global_uptime_90d || 100)) * 10,
          ),
          activeDeviceErrors: payload.active_incidents || 0,
          avgResolutionTime: `${payload.mttr_minutes || 0}m`,
          trends: payload.trends || {
            openTicketsChange: 0,
            resolutionTimeChange: 0,
          },
        });
        if (payload.top_errors) {
          setTopErrors(payload.top_errors);
        }
        if (payload.heatmap) {
          setTrendPoints(
            payload.heatmap.map((v: number, i: number) => ({
              label: `D${i + 1}`,
              value: v,
            })),
          );
        }
      }

      if (ticketsRes.ok) {
        const data = await ticketsRes.json();
        const ticketsList = data.data || [];
        setRecentTickets(
          ticketsList.map((t: any) => ({
            ticketId: t.id,
            subject: t.subject,
            reporter: t.vehicle_id || t.driver_id || "System",
            priority: t.priority || "MEDIUM",
            assignedTo: t.assigned_to || "Support Queue",
            status: t.status || "NEW",
            updatedAt: t.created_at || new Date().toISOString(),
          })),
        );
      }
    } catch (err) {
      console.error("Failed to fetch support dashboard data", err);
    }
  };

  useEffect(() => {
    fetchData();
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) => setSystemMeta(data.system_metadata))
      .catch((err) => console.error("Global config fetch failed", err));
    let interval: any;
    if (autoRefresh) {
      interval = setInterval(fetchData, 30000);
    }
    return () => clearInterval(interval);
  }, [autoRefresh]);

  const chartPath = (() => {
    if (!trendPoints.length) return "";
    const maxVal = Math.max(...trendPoints.map((p) => p.value), 1);
    const minVal = Math.min(...trendPoints.map((p) => p.value), 0);
    const range = Math.max(maxVal - minVal, 1);
    return trendPoints
      .map((point, index) => {
        const x = (index / Math.max(trendPoints.length - 1, 1)) * 800;
        const y = 190 - ((point.value - minVal) / range) * 150;
        return `${index === 0 ? "M" : "L"}${x},${y}`;
      })
      .join(" ");
  })();

  const openTickets = toNumber(metrics?.openTickets);
  const slaBreaches = toNumber(metrics?.slaBreaches24h);
  const activeErrors = toNumber(metrics?.activeDeviceErrors);
  const openTicketsChange = toNumber(metrics?.trends?.openTicketsChange);
  const resolutionTrend = toNumber(metrics?.trends?.resolutionTimeChange);

  const filteredTickets = recentTickets.filter((ticket) => {
    const q = searchQuery.toLowerCase();
    if (!q) return true;
    return (
      String(ticket.ticketId || "")
        .toLowerCase()
        .includes(q) ||
      String(ticket.subject || "")
        .toLowerCase()
        .includes(q) ||
      String(ticket.reporter || "")
        .toLowerCase()
        .includes(q)
    );
  });

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden font-display">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-white text-lg font-bold">
              {uiStrings?.dashboard_title || "Support Dashboard"}
            </h2>
            <div className="flex items-center gap-2 mt-0.5">
              <span className="text-[11px] text-[#92adc9] uppercase tracking-wider font-semibold">
                {uiStrings?.breadcrumb_support || "Support"}
              </span>
              <span className="text-[11px] text-[#92adc9]">/</span>
              <span className="text-[11px] text-white uppercase tracking-wider font-semibold">
                {commonStrings?.dashboard || "Dashboard"}
              </span>
            </div>
          </div>
          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-emerald-500/10 text-emerald-500 border border-emerald-500/30">
            {commonStrings?.live_data || "Live Data"}
          </span>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3 px-3 py-1.5 rounded bg-[#16202a] border border-[#233648]">
            <span className="text-xs text-[#92adc9] font-medium">
              {commonStrings?.auto_refresh || "Auto-Refresh"}
            </span>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary"></div>
            </label>
          </div>
          <button
            onClick={() => setIsModalOpen(true)}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <span className="material-symbols-outlined text-[16px]">add</span>
            {uiStrings?.btn_new_ticket || "New Ticket"}
          </button>
        </div>
      </header>

      <NewTicketModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onCreated={() => fetchData()}
      />

      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-white">
                inbox
              </span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">
              {uiStrings?.stat_open_tickets || "Open Tickets"}
            </div>
            <div className="text-2xl font-bold text-white">{openTickets}</div>
            <div className="text-[10px] text-emerald-500 flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">
                trending_down
              </span>
              {openTicketsChange}%{" "}
              {commonStrings?.vs_yesterday || "vs yesterday"}
            </div>
          </div>

          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-rose-500">
                timer_off
              </span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">
              {uiStrings?.stat_sla_breaches || "SLA Breaches (24h)"}
            </div>
            <div className="text-2xl font-bold text-rose-500">
              {slaBreaches}
            </div>
            <div className="text-[10px] text-rose-500 mt-1">
              {commonStrings?.priority_critical || "Critical Priority"}
            </div>
          </div>

          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-amber-500">
                error
              </span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">
              {uiStrings?.stat_device_errors || "Active Device Errors"}
            </div>
            <div className="text-2xl font-bold text-amber-500">
              {activeErrors}
            </div>
            <div className="text-[10px] text-[#92adc9] mt-1">
              {commonStrings?.across_fleet || "Across fleet"}
            </div>
          </div>

          <div className="bg-[#16202a] rounded-xl border border-[#233648] p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-emerald-500">
                av_timer
              </span>
            </div>
            <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">
              {uiStrings?.stat_avg_resolution || "Avg. Resolution Time"}
            </div>
            <div className="text-2xl font-bold text-emerald-500">
              {metrics?.avgResolutionTime || "0.0h"}
            </div>
            <div className="text-[10px] text-emerald-500 flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">
                trending_down
              </span>
              {Math.abs(resolutionTrend)}%{" "}
              {commonStrings?.faster_this_week || "faster this week"}
            </div>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-6 mb-6">
          <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 h-80 flex flex-col">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-primary text-[18px]">
                    show_chart
                  </span>
                  {uiStrings?.trends_title || "Incident Trends (Last 7 Days)"}
                </h3>
                <select className="bg-[#111a22] text-xs text-[#92adc9] border border-[#233648] rounded px-2 py-1">
                  <option>{commonStrings?.last_7d || "Last 7 Days"}</option>
                  <option>{commonStrings?.last_30d || "Last 30 Days"}</option>
                </select>
              </div>
              <div className="flex-1 relative w-full h-full">
                <svg
                  className="w-full h-full"
                  preserveAspectRatio="none"
                  viewBox="0 0 800 200"
                >
                  <defs>
                    <linearGradient
                      id="chartGradient"
                      x1="0"
                      x2="0"
                      y1="0"
                      y2="1"
                    >
                      <stop
                        offset="0%"
                        stopColor="#137fec"
                        stopOpacity="0.3"
                      ></stop>
                      <stop
                        offset="100%"
                        stopColor="#137fec"
                        stopOpacity="0"
                      ></stop>
                    </linearGradient>
                  </defs>
                  {chartPath && (
                    <>
                      <path
                        d={`${chartPath} L800,200 L0,200 Z`}
                        fill="url(#chartGradient)"
                      ></path>
                      <path
                        d={chartPath}
                        fill="none"
                        stroke="#137fec"
                        strokeWidth="2"
                      ></path>
                    </>
                  )}
                </svg>
              </div>
            </div>
          </div>

          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col h-80">
              <div className="p-4 border-b border-[#233648] flex justify-between items-center">
                <h3 className="text-white text-sm font-semibold">
                  {uiStrings?.system_errors_title || "Top System Errors"}
                </h3>
                <button className="text-primary text-xs hover:underline">
                  {commonStrings?.view_all || "View All"}
                </button>
              </div>
              <div className="flex-1 p-4 flex flex-col justify-center gap-4">
                {topErrors.slice(0, 4).map((error, idx) => (
                  <div key={idx} className="w-full">
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-white">
                        {error.title} ({error.errorCode})
                      </span>
                      <span
                        className={`${error.severity === "CRITICAL" ? "text-rose-500" : "text-amber-500"} font-bold`}
                      >
                        {error.count}
                      </span>
                    </div>
                    <div className="w-full bg-[#111a22] h-2 rounded-full overflow-hidden">
                      <div
                        className={`${error.severity === "CRITICAL" ? "bg-rose-500" : "bg-amber-500"} h-full rounded-full`}
                        style={{
                          width: `${(error.count / (topErrors[0]?.count || 1)) * 100}%`,
                        }}
                      ></div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden">
          <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-[#92adc9] text-[18px]">
                history
              </span>
              {uiStrings?.activity_feed_title || "Recent Support Activity"}
            </h3>
            <div className="flex items-center gap-2">
              <input
                className="bg-[#16202a] border border-[#233648] text-xs text-white rounded px-3 py-1.5 w-64 focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                placeholder={
                  uiStrings?.search_placeholder || "Search tickets, IDs..."
                }
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                  <th className="px-4 py-3 font-semibold">
                    {uiStrings?.th_id || "Ticket ID"}
                  </th>
                  <th className="px-4 py-3 font-semibold">
                    {uiStrings?.th_subject || "Subject"}
                  </th>
                  <th className="px-4 py-3 font-semibold">
                    {uiStrings?.th_reporter || "Reporter"}
                  </th>
                  <th className="px-4 py-3 font-semibold">
                    {uiStrings?.th_priority || "Priority"}
                  </th>
                  <th className="px-4 py-3 font-semibold">
                    {uiStrings?.th_assignee || "Assigned To"}
                  </th>
                  <th className="px-4 py-3 font-semibold text-right">
                    {commonStrings?.status || "Status"}
                  </th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {filteredTickets.map((ticket, idx) => (
                  <tr
                    key={idx}
                    className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors"
                  >
                    <td className="px-4 py-3 text-white font-mono text-xs">
                      {ticket.ticketId}
                    </td>
                    <td className="px-4 py-3 text-white">{ticket.subject}</td>
                    <td className="px-4 py-3 text-[#92adc9]">
                      {ticket.reporter}
                    </td>
                    <td className="px-4 py-3 text-[#92adc9]">
                      {ticket.priority}
                    </td>
                    <td className="px-4 py-3 text-[#92adc9]">
                      {ticket.assignedTo}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold border bg-primary/10 text-primary border-primary/20">
                        {ticket.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <footer className="bg-[#111a22] border-t border-[#233648] px-6 py-2 flex items-center justify-between text-[10px] text-[#92adc9] z-20">
        <span>
          © {new Date().getFullYear()}{" "}
          {commonStrings?.company_label || "Enterprise Fleet Manager"}.{" "}
          {uiStrings?.breadcrumb_support || "Support"}{" "}
          {commonStrings?.module_label || "Module"}
        </span>
        <div className="flex items-center gap-4">
          <span>
            {commonStrings?.api_status || "API Status"}:{" "}
            <span className="text-emerald-500">
              {commonStrings?.connected || "Connected"}
            </span>
          </span>
          <span>
            {uiStrings?.agent_status || "Support Agent"}:{" "}
            <span className="text-white">
              {commonStrings?.active || "Active"}
            </span>
          </span>
        </div>
      </footer>
    </main>
  );
};

export default SupportDashboard;
