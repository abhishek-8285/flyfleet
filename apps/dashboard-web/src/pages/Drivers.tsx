import React, { useEffect, useMemo, useState } from "react";
import {
    DriverFilters,
    DriverStatusFilter,
    SafetyScoreFilter,
} from "../components/drivers/DriverFilters";
import { DriverTable } from "../components/drivers/DriverTable";
import { useUIStrings } from "../hooks/useUIStrings";
import { DriverPerformance } from "../modules/drivers/components/DriverPerformance";
import { driverService } from "../modules/drivers/DriverService";
import {
    Driver as EnterpriseDriver,
    SafetyScore,
} from "../modules/drivers/types";
import { api } from "../services/api";
import { Driver } from "../types";

const Drivers: React.FC = () => {
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<DriverStatusFilter>("all");
  const [safetyFilter, setSafetyFilter] = useState<SafetyScoreFilter>("all");
  const { uiStrings } = useUIStrings("drivers");

  const [selectedDriver, setSelectedDriver] = useState<EnterpriseDriver | null>(
    null,
  );
  const [safetyScore, setSafetyScore] = useState<SafetyScore | null>(null);

  useEffect(() => {
    const fetchDrivers = async () => {
      setLoading(true);
      try {
        const data = await api.getDrivers();
        // Handle nested response format: { data: { drivers: [...], total: n } }
        const list = (data?.data?.drivers || 
                      data?.drivers ||
                      data?.data ||
                      data?.items ||
                      data) as Driver[];
        setDrivers(Array.isArray(list) ? list : []);
      } catch (error) {
        console.error("Failed to load drivers", error);
        setDrivers([]);
      } finally {
        setLoading(false);
      }
    };

    fetchDrivers();
  }, []);

  const handleDriverClick = async (id: string) => {
    const driver = await driverService.getDriver(id);
    const score = await driverService.getSafetyScore(id);
    setSelectedDriver(driver);
    setSafetyScore(score);
  };

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
      {/* Quick View Overlay if selected */}
      {selectedDriver && safetyScore && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-6 bg-black/60 backdrop-blur-sm"
          onClick={() => setSelectedDriver(null)}
        >
          <div
            className="w-full max-w-4xl max-h-[90vh] overflow-auto"
            onClick={(e) => e.stopPropagation()}
          >
            <DriverPerformance driver={selectedDriver as any} analytics={safetyScore} />
          </div>
        </div>
      )}

      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10">
        <h1 className="text-xl font-bold text-white">
          Driver Fleet Management
        </h1>
        <div className="flex items-center gap-3 ml-auto">
          <button className="bg-primary hover:bg-primary/90 text-white text-xs font-medium py-1.5 px-3 rounded flex items-center gap-2 shadow-[0_0_10px_rgba(19,127,236,0.3)] transition-all">
            <span className="material-symbols-outlined text-[16px]">
              person_add
            </span>
            {uiStrings?.btn_add_driver || "Add New Pilot"}
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
        <DriverTable
          drivers={filteredDrivers}
          loading={loading}
          onDriverClick={(d: Driver) => handleDriverClick(d.id)}
        />
      </div>
    </div>
  );
};

export default Drivers;
