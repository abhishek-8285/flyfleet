import { motion } from "framer-motion";
import {
    Activity,
    AlertTriangle,
    TrendingDown,
    TrendingUp,
    Truck,
    Users,
} from "lucide-react";
import React from "react";
import AnimatedNumber from "../../../components/animated/AnimatedNumber";

interface Stat {
  label: string;
  value: string | number;
  trend: number;
  icon: React.ElementType;
  color: string;
}

const StatsGrid: React.FC<{ stats: any }> = ({ stats }) => {
  const [shouldReduceMotion, setShouldReduceMotion] = React.useState(false);

  React.useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setShouldReduceMotion(mediaQuery.matches);

    const handler = (e: MediaQueryListEvent) => {
      setShouldReduceMotion(e.matches);
    };

    mediaQuery.addEventListener('change', handler);
    return () => mediaQuery.removeEventListener('change', handler);
  }, []);

  const items: Stat[] = [
    {
      label: "Live Vehicles",
      value: stats?.total_vehicles || 0,
      trend: 12.5,
      icon: Truck,
      color: "text-blue-400",
    },
    {
      label: "Active Trips",
      value: stats?.active_trips || 0,
      trend: 8.2,
      icon: Activity,
      color: "text-green-400",
    },
    {
      label: "Active Drivers",
      value: stats?.active_drivers || 0,
      trend: -2.4,
      icon: Users,
      color: "text-purple-400",
    },
    {
      label: "Critical Alerts",
      value: stats?.alerts_count || 0,
      trend: 1.2,
      icon: AlertTriangle,
      color: "text-red-400",
    },
    {
      label: "Fleet Health",
      value: `${stats?.fleet_health_score || 0}%`,
      trend: 2.1,
      icon: Activity,
      color: "text-emerald-400",
    },
    {
      label: "Offline Units",
      value: stats?.vehicles_offline || 0,
      trend: -5.4,
      icon: Truck,
      color: "text-gray-400",
    },
    {
      label: "SLA Violations",
      value: stats?.sla_violations || 0,
      trend: 0,
      icon: AlertTriangle,
      color: "text-orange-400",
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-4 gap-4">
      {items.map((stat, idx) => (
        <motion.div
          key={idx}
          initial={shouldReduceMotion ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: shouldReduceMotion ? 0 : idx * 0.1 }}
          className="bg-gray-900 border border-gray-800 p-6 rounded-xl shadow-lg hover:border-gray-700 transition-colors"
        >
          <div className="flex justify-between items-start mb-4">
            <div className={`p-3 rounded-lg bg-gray-800 ${stat.color}`}>
              <stat.icon size={24} />
            </div>
            <div
              className={`flex items-center text-sm ${stat.trend > 0 ? "text-green-400" : "text-red-400"}`}
            >
              {stat.trend > 0 ? (
                <TrendingUp size={16} className="mr-1" />
              ) : (
                <TrendingDown size={16} className="mr-1" />
              )}
              {Math.abs(stat.trend)}%
            </div>
          </div>
          <h3 className="text-gray-400 text-sm font-medium">{stat.label}</h3>
          <p className="text-2xl font-bold text-white mt-1">
            <AnimatedNumber value={stat.value as number} />
          </p>
        </motion.div>
      ))}
    </div>
  );
};

export default StatsGrid;
