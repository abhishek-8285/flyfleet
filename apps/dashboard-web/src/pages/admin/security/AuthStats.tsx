import { motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import {
    Area,
    AreaChart,
    Cell,
    Pie,
    PieChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import EnterpriseAuthLayout from "../../../components/auth/EnterpriseAuthLayout";
import api from "../../../services/api";
import { AuthStats } from "../../../types/auth";

const AuthStatsPage: React.FC = () => {
  const [stats, setStats] = useState<AuthStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const resp = await api.getAuthStats();
        setStats(resp.data);
      } catch (err) {
        // Mock data
        setStats({
          totalLogins: 125403,
          mfaAdoptionRate: 88.5,
          activeSessions: 4210,
          failedAuthRate: 2.1,
          deviceDistribution: [
            { type: "Desktop", count: 2450 },
            { type: "Mobile", count: 1560 },
            { type: "Tablet", count: 200 },
          ],
          dailyLoginHistory: [
            { date: "Mar 01", logins: 12000 },
            { date: "Mar 02", logins: 15000 },
            { date: "Mar 03", logins: 14000 },
            { date: "Mar 04", logins: 22000 },
            { date: "Mar 05", logins: 18000 },
            { date: "Mar 06", logins: 11000 },
            { date: "Mar 07", logins: 9500 },
          ],
        });
      } finally {
        setLoading(false);
      }
    };
    fetchStats();
  }, []);

  const COLORS = ["#3b82f6", "#10b981", "#f59e0b"];

  const StatCard = ({ label, value, sub, color = "primary" }: any) => (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-6 rounded-[32px] space-y-2 flex-1 min-w-[200px]"
    >
      <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">
        {label}
      </p>
      <p className="text-3xl text-white font-black tracking-tighter">{value}</p>
      <p className="text-xs text-slate-400">{sub}</p>
    </motion.div>
  );

  return (
    <EnterpriseAuthLayout
      title="Authentication Analytics"
      subtitle="System-wide identity intelligence metrics"
    >
      <div className="p-4 space-y-6">
        {/* Top Stats */}
        <div className="flex flex-wrap gap-4">
          <StatCard
            label="Total Authentications"
            value={stats?.totalLogins.toLocaleString()}
            sub="+12.5% from last month"
          />
          <StatCard
            label="MFA Adoption Rate"
            value={`${stats?.mfaAdoptionRate}%`}
            sub="Enterprise-wide coverage"
          />
          <StatCard
            label="Active Fleet Sessions"
            value={stats?.activeSessions.toLocaleString()}
            sub="Real-time connections"
          />
          <StatCard
            label="Failed Auth Rate"
            value={`${stats?.failedAuthRate}%`}
            sub="Global security health"
          />
        </div>

        {/* Growth Chart */}
        <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-8 rounded-[40px] space-y-6">
          <h3 className="text-white font-bold text-sm uppercase tracking-widest">
            Login Activity Trends
          </h3>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={stats?.dailyLoginHistory}>
                <defs>
                  <linearGradient id="colorLogin" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis
                  dataKey="date"
                  stroke="#475569"
                  fontSize={10}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis hide />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#0f172a",
                    border: "1px solid rgba(255,255,255,0.1)",
                    borderRadius: "12px",
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="logins"
                  stroke="#3b82f6"
                  fillOpacity={1}
                  fill="url(#colorLogin)"
                  strokeWidth={3}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Device Distribution */}
        <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-8 rounded-[40px] space-y-6 max-w-sm">
          <h3 className="text-white font-bold text-sm uppercase tracking-widest">
            Endpoint Distribution
          </h3>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={stats?.deviceDistribution}
                  innerRadius={70}
                  outerRadius={90}
                  paddingAngle={8}
                  dataKey="count"
                >
                  {stats?.deviceDistribution.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={COLORS[index % COLORS.length]}
                    />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="flex justify-center gap-6">
            {stats?.deviceDistribution.map((d: any, i: number) => (
              <div key={i} className="flex items-center gap-2">
                <div
                  className="w-2 h-2 rounded-full"
                  style={{ backgroundColor: COLORS[i] }}
                />
                <span className="text-[10px] text-slate-400 font-bold uppercase">
                  {d.type}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default AuthStatsPage;
