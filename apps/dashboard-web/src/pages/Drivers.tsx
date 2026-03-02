import React, { useEffect, useMemo, useState } from "react";
import {
    DriverFilters,
    DriverStatusFilter,
    SafetyScoreFilter,
} from "../components/drivers/DriverFilters";
import { DriverTable } from "../components/drivers/DriverTable";
import { useUIStrings } from "../hooks/useUIStrings";
import { api } from "../services/api";
import { Driver } from "../types";

const Drivers: React.FC = () => {
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<DriverStatusFilter>("all");
  const [safetyFilter, setSafetyFilter] = useState<SafetyScoreFilter>("all");
  const { uiStrings } = useUIStrings("drivers");
  const [systemMeta, setSystemMeta] = useState<any>(null);

  useEffect(() => {
    const fetchDrivers = async () => {
      setLoading(true);
      try {
        const data = await api.getDrivers();
        const list = (data?.drivers ||
          data?.data ||
          data?.items ||
          data) as Driver[];
        setDrivers(Array.isArray(list) ? list : []);
      } catch (error) {
        console.error("Failed to load drivers", error);
        setDrivers([]);
      } finally {
        setLoading(false);
        fetch("/api/v1/ui/global_config")
          .then((res) => res.json())
          .then((data) => setSystemMeta(data.system_metadata))
          .catch((err) => console.error("Failed to fetch system config", err));
      }
    };

    fetchDrivers();
  }, []);

  const filteredDrivers = useMemo(() => {
    const matchStatus = (driver: Driver) => {
      if (statusFilter === "all") return true;
      return driver.status === statusFilter;
    };

    const matchSafety = (driver: Driver) => {
      if (safetyFilter === "all") return true;
      const score = Math.max(
        0,
        Math.min(100, Math.round(driver.behavior_score ?? driver.rating ?? 0)),
      );
      if (safetyFilter === "excellent") return score >= 90;
      if (safetyFilter === "good") return score >= 80 && score < 90;
      if (safetyFilter === "needs_review") return score < 80;
      if (safetyFilter === "critical") return score < 60;
      return true;
    };

    return drivers.filter(
      (driver) => matchStatus(driver) && matchSafety(driver),
    );
  }, [drivers, statusFilter, safetyFilter]);

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10">
        <div className="flex flex-wrap items-center gap-3">
          <div className="relative">
            <select
              value={statusFilter}
              onChange={(e) =>
                setStatusFilter(e.target.value as DriverStatusFilter)
              }
              className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors"
            >
              <option value="all">
                {uiStrings?.filter_all || "All Drivers"}
              </option>
              <option value="available">
                {uiStrings?.filter_available || "Available"}
              </option>
              <option value="on_trip">
                {uiStrings?.filter_on_trip || "On Trip"}
              </option>
              <option value="off_duty">
                {uiStrings?.filter_off_duty || "Off Duty"}
              </option>
            </select>
            <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">
              badge
            </span>
          </div>
        </div>
        <div className="flex items-center gap-3 ml-auto">
          <button className="bg-primary hover:bg-primary/90 text-white text-xs font-medium py-1.5 px-3 rounded flex items-center gap-2 shadow-[0_0_10px_rgba(19,127,236,0.3)] transition-all">
            <span className="material-symbols-outlined text-[16px]">
              person_add
            </span>
            {uiStrings?.btn_add_driver || "Add Driver"}
          </button>
        </div>
      </div>

      <DriverFilters
        statusFilter={statusFilter}
        safetyFilter={safetyFilter}
        onStatusChange={setStatusFilter}
        onSafetyChange={setSafetyFilter}
      />

      <div className="flex-1 overflow-hidden p-6 bg-background-dark">
        <DriverTable drivers={filteredDrivers} loading={loading} />
      </div>

      <footer className="mt-auto text-center text-xs text-[#64748b] pb-2 flex flex-col items-center gap-1 bg-background-dark">
        <span>
          {(
            uiStrings?.footer_copyright ||
            "© {year} FleetCommand. All rights reserved."
          ).replace("{year}", new Date().getFullYear().toString())}
        </span>
        <span className="text-[10px]">
          {systemMeta?.version || "v--"} | Server Time:{" "}
          <span className="font-mono">{systemMeta?.server_time || "---"}</span>
        </span>
      </footer>
    </div>
  );
};

export default Drivers;
