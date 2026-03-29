import React, { useEffect } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { StatCard } from "../animated/AnimatedCard";
import { AnimatedNumber } from "../animated/AnimatedNumber";
import { motion } from "framer-motion";

interface StatsGridProps {
  stats: {
    total_vehicles: number;
    active_trips: number;
    active_alerts: number;
    total_drivers?: number;
    online_drivers?: number;
    trends?: {
      vehicles: number;
      trips: number;
      alerts: number;
      drivers?: number;
      [key: string]: number | undefined;
    };
    metadata?: {
      stats_sources?: {
        total_vehicles?: string;
        active_trips?: string;
        online_drivers?: string;
        active_alerts?: string;
        [key: string]: string | undefined;
      };
    };
  };
}

const containerVariants = {
  hidden: {},
  visible: {
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.2
    }
  }
};

export const StatsGrid: React.FC<StatsGridProps> = ({ stats }) => {
  const { uiStrings } = useUIStrings("comp_dashboard");
  const getTrendColor = (val: number) =>
    val >= 0 ? "text-emerald-500" : "text-error";
  const getTrendIcon = (val: number) =>
    val >= 0 ? "trending_up" : "trending_down";

  // Check for reduced motion preference
  const [shouldAnimate, setShouldAnimate] = React.useState(true);
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setShouldAnimate(!mediaQuery.matches);
    const handleChange = (e: MediaQueryListEvent) => {
      setShouldAnimate(!e.matches);
    };
    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  const StatCardContent = ({ titleKey, value, trendKey, sourceKey, icon, badge }: any) => (
    <>
      <div className="flex justify-between items-start z-10 relative">
        <div className="tooltip-trigger relative">
          <div className="flex items-center gap-1 mb-1">
            <p className="text-text-muted text-sm font-medium">
              {titleKey}
            </p>
            <span className="material-symbols-outlined text-[14px] text-text-muted/50 hover:text-primary cursor-help">
              info
            </span>
          </div>
          <h3 className="text-3xl font-bold text-white tracking-tight">
            {shouldAnimate ? (
              <AnimatedNumber value={value} duration={1500} />
            ) : (
              value.toLocaleString()
            )}
          </h3>
        </div>
        <div className={`p-2 ${badge} rounded-lg`}>
          <span className="material-symbols-outlined">{icon}</span>
        </div>
      </div>
      <div className="mt-4 flex items-center text-sm">
        <span
          className={`${getTrendColor(stats.trends?.[trendKey] ?? 0)} font-medium flex items-center gap-1`}
        >
          <span className="material-symbols-outlined text-[16px]">
            {getTrendIcon(stats.trends?.[trendKey] ?? 0)}
          </span>
          {Math.abs(stats.trends?.[trendKey] ?? 0)}%
        </span>
        <span className="text-[#64748b] ml-2">vs last week</span>
      </div>
      <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
        <span>
          Source:{" "}
          {stats.metadata?.stats_sources?.[sourceKey] || "Live System"}
        </span>
        <span className="text-emerald-500/80">Active</span>
      </div>
    </>
  );

  return (
    <motion.div
      className="col-span-12 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {/* Live Vehicles Card */}
      <StatCard data-testid="stat-card-vehicles">
        <StatCardContent
          titleKey={uiStrings?.label_active_vehicles || "Live Vehicles"}
          value={stats.total_vehicles}
          trendKey="vehicles"
          sourceKey="total_vehicles"
          icon="directions_bus"
          badge="bg-primary/10 text-primary"
        />
      </StatCard>

      {/* Active Trips Card */}
      <StatCard data-testid="stat-card-trips">
        <StatCardContent
          titleKey={uiStrings?.label_active_trips || "Active Trips"}
          value={stats.active_trips}
          trendKey="trips"
          sourceKey="active_trips"
          icon="route"
          badge="bg-purple-500/10 text-purple-500"
        />
      </StatCard>

      {/* Drivers Card */}
      <StatCard data-testid="stat-card-drivers">
        <StatCardContent
          titleKey={uiStrings?.label_active_drivers || "Active Drivers"}
          value={stats.online_drivers ?? 0}
          trendKey="drivers"
          sourceKey="online_drivers"
          icon="person"
          badge="bg-blue-500/10 text-blue-500"
        />
        <div className="mt-4 flex items-center text-sm">
          <span className="text-white font-medium">
            {stats.total_drivers ?? 0}
          </span>
          <span className="text-[#64748b] ml-2">Total Registered</span>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>
            Source:{" "}
            {stats.metadata?.stats_sources?.online_drivers || "Driver Service"}
          </span>
          <span className="text-text-muted">Synced</span>
        </div>
      </StatCard>

      {/* Active Alerts Card */}
      <StatCard
        data-testid="stat-card-alerts"
        whileHover={{
          y: -6,
          boxShadow: '0 25px 50px -12px rgba(239, 68, 68, 0.2)',
          borderColor: 'rgba(239, 68, 68, 0.4)'
        }}
      >
        <StatCardContent
          titleKey={uiStrings?.label_alerts || "Critical Alerts"}
          value={stats.active_alerts}
          trendKey="alerts"
          sourceKey="active_alerts"
          icon="warning"
          badge="bg-red-500/10 text-red-500"
        >
          <div className="flex items-baseline gap-2 mt-1">
            <h3 className="text-3xl font-bold text-white tracking-tight">
              {shouldAnimate ? (
                <AnimatedNumber value={stats.active_alerts} duration={1500} />
              ) : (
                stats.active_alerts
              )}
            </h3>
            <span
              className={`${getTrendColor(-(stats.trends?.alerts ?? 0))} text-xs font-medium`}
            >
              {(stats.trends?.alerts ?? 0) < 0 ? "↓" : "↑"}{" "}
              {Math.abs(stats.trends?.alerts ?? 0)}%
            </span>
          </div>
        </StatCardContent>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>
            Source:{" "}
            {stats.metadata?.stats_sources?.active_alerts || "Radar-VNet"}
          </span>
          <a className="text-primary hover:underline" href="/dashboard/alerts">
            Manage →
          </a>
        </div>
      </StatCard>
    </motion.div>
  );
};

export default StatsGrid;

