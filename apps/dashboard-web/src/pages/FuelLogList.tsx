import { MagnifyingGlassIcon } from "@heroicons/react/24/outline";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { FuelLogTable, FuelStats } from "../components/fuel";
import { useUIStrings } from "../hooks/useUIStrings";
import FuelService from "../services/FuelService";
import { FuelLog, FuelLogStats } from "../types/fuel";

export default function FuelLogList() {
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("fleet_fuel");
  const uiStrings = rawUiStrings || {};
  const [logs, setLogs] = useState<FuelLog[]>([]);
  const [stats, setStats] = useState<FuelLogStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [logsData, statsData] = await Promise.all([
          FuelService.listLogs(),
          FuelService.getStats(),
        ]);
        setLogs(logsData.logs);
        setStats(statsData);
      } catch (error) {
        console.error("Failed to fetch fuel data:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header / Toolbar */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10 transition-all duration-300">
        <div className="flex flex-wrap items-center gap-3">
          <div className="text-xl font-semibold text-white tracking-tight flex items-center gap-2">
            <span className="material-symbols-outlined text-primary">
              local_gas_station
            </span>
            {uiStrings?.title || "Fuel Management"}
          </div>
          <div className="h-6 w-px bg-border-color mx-2 hidden md:block"></div>
          <div className="relative">
            <select className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors">
              <option>{uiStrings?.range_30d || "Last 30 Days"}</option>
              <option>{uiStrings?.range_this_month || "This Month"}</option>
              <option>{uiStrings?.range_last_month || "Last Month"}</option>
              <option>{uiStrings?.range_this_year || "This Year"}</option>
            </select>
            <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">
              calendar_month
            </span>
          </div>
        </div>

        <div className="flex items-center gap-3 ml-auto">
          <button
            onClick={() => navigate("/fuel/settings")}
            className="p-1.5 text-text-muted hover:text-white hover:bg-surface-dark rounded transition-colors"
            title={uiStrings?.title_settings || "Fuel Settings"}
          >
            <span className="material-symbols-outlined text-[20px]">
              settings
            </span>
          </button>
          <div className="h-4 w-px bg-border-color"></div>
          <button
            onClick={() => navigate("/fuel/add")}
            className="bg-primary hover:bg-primary/90 text-white text-xs font-medium py-1.5 px-4 rounded-full flex items-center gap-2 shadow-[0_4px_14px_rgba(19,127,236,0.39)] hover:shadow-[0_6px_20px_rgba(19,127,236,0.23)] hover:-translate-y-0.5 transition-all duration-200 active:scale-95"
          >
            <span className="material-symbols-outlined text-[18px]">add</span>
            {uiStrings?.log_btn || "Log Fuel Entry"}
          </button>
        </div>
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar bg-[radial-gradient(circle_at_top_right,rgba(19,127,236,0.05),transparent)]">
        <div className="max-w-[1600px] mx-auto pb-6 space-y-6 animate-fade-in">
          {stats && <FuelStats stats={stats} />}

          <div className="grid grid-cols-12 gap-6">
            <div className="col-span-12">
              <div className="bg-surface-dark/50 backdrop-blur-md border border-border-color rounded-2xl overflow-hidden shadow-xl">
                <div className="px-6 py-4 border-b border-border-color flex items-center justify-between bg-surface-darker/30">
                  <h3 className="text-sm font-medium text-white">
                    {uiStrings?.logs_title || "Fuel Logs"}
                  </h3>
                  <div className="flex items-center gap-2">
                    <div className="relative">
                      <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                      <input
                        type="text"
                        placeholder={
                          uiStrings?.search_placeholder || "Search logs..."
                        }
                        className="bg-background-dark border border-border-color text-white text-xs rounded-full pl-9 pr-4 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary w-64 outline-none transition-all focus:w-80"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                      />
                    </div>
                  </div>
                </div>
                <FuelLogTable logs={logs} loading={loading} />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
