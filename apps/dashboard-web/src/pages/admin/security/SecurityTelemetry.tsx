import { motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import {
    Area,
    AreaChart,
    Bar,
    BarChart,
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
import { SecurityTelemetry } from "../../../types/auth";

const SecurityTelemetryPage: React.FC = () => {
  const [data, setData] = useState<SecurityTelemetry | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const resp = await api.getSecurityTelemetry();
        setData(resp.data);
      } catch (err) {
        // Mock data
        setData({
          loginFailures: [
            { timestamp: "08:00", count: 12 },
            { timestamp: "10:00", count: 45 },
            { timestamp: "12:00", count: 30 },
            { timestamp: "14:00", count: 85 },
            { timestamp: "16:00", count: 25 },
            { timestamp: "18:00", count: 15 },
          ],
          mfaChallenges: [
            { method: "SMS", count: 450 },
            { method: "TOTP", count: 890 },
            { method: "Passkey", count: 1200 },
            { method: "Email", count: 120 },
          ],
          riskEvents: [
            {
              type: "Credential Stuffing",
              severity: "high",
              timestamp: "10 min ago",
            },
            {
              type: "Impossible Travel",
              severity: "medium",
              timestamp: "45 min ago",
            },
            {
              type: "New Device Login",
              severity: "low",
              timestamp: "2 hours ago",
            },
          ],
          authMethodDistribution: [
            { method: "Password", percentage: 40 },
            { method: "Passkey", percentage: 35 },
            { method: "SSO", percentage: 25 },
          ],
        });
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const COLORS = ["#3b82f6", "#8b5cf6", "#ec4899", "#f59e0b"];

  return (
    <EnterpriseAuthLayout
      title="Security Telemetry"
      subtitle="Real-time threat detection and auth dynamics"
    >
      <div className="p-4 space-y-6">
        {/* Failed Logins Area Chart */}
        <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-6 rounded-[32px] space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-white font-bold text-sm uppercase tracking-widest">
              Failed Login Velocity
            </h3>
            <span className="text-[10px] text-red-500 font-bold bg-red-500/10 px-2 py-1 rounded-full uppercase tracking-tighter">
              Live Monitor
            </span>
          </div>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={data?.loginFailures}>
                <defs>
                  <linearGradient id="colorFail" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis
                  dataKey="timestamp"
                  stroke="#475569"
                  fontSize={10}
                  axisLine={false}
                  tickLine={false}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#0f172a",
                    border: "1px solid rgba(255,255,255,0.1)",
                    borderRadius: "12px",
                  }}
                  itemStyle={{ color: "#fff", fontSize: "12px" }}
                />
                <Area
                  type="monotone"
                  dataKey="count"
                  stroke="#ef4444"
                  fillOpacity={1}
                  fill="url(#colorFail)"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* MFA Method Stats */}
          <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-6 rounded-[32px] space-y-4">
            <h3 className="text-white font-bold text-sm uppercase tracking-widest">
              MFA Challenge Distribution
            </h3>
            <div className="h-64 w-full">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={data?.mfaChallenges} layout="vertical">
                  <XAxis type="number" hide />
                  <YAxis
                    dataKey="method"
                    type="category"
                    stroke="#94a3b8"
                    fontSize={12}
                    axisLine={false}
                    tickLine={false}
                    width={80}
                  />
                  <Tooltip cursor={{ fill: "transparent" }} />
                  <Bar
                    dataKey="count"
                    fill="#3b82f6"
                    radius={[0, 4, 4, 0]}
                    barSize={20}
                  />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* Auth Method Pie */}
          <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-6 rounded-[32px] space-y-4">
            <h3 className="text-white font-bold text-sm uppercase tracking-widest">
              Primary Auth Methods
            </h3>
            <div className="h-64 w-full">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={data?.authMethodDistribution}
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="percentage"
                  >
                    {data?.authMethodDistribution.map(
                      (entry: any, index: number) => (
                        <Cell
                          key={`cell-${index}`}
                          fill={COLORS[index % COLORS.length]}
                        />
                      ),
                    )}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        {/* Risk Events List */}
        <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-6 rounded-[32px] space-y-4">
          <h3 className="text-white font-bold text-sm uppercase tracking-widest">
            Recent Risk Anomalies
          </h3>
          <div className="space-y-3">
            {data?.riskEvents.map((event: any, i: number) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.1 }}
                className="flex items-center justify-between p-4 bg-white/5 rounded-2xl border border-white/5"
              >
                <div className="flex items-center gap-4">
                  <div
                    className={`w-2 h-2 rounded-full ${
                      event.severity === "high"
                        ? "bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)]"
                        : event.severity === "medium"
                          ? "bg-orange-500"
                          : "bg-blue-500"
                    }`}
                  />
                  <div>
                    <p className="text-white font-bold text-sm">{event.type}</p>
                    <p className="text-[10px] text-slate-500 uppercase tracking-widest">
                      {event.timestamp}
                    </p>
                  </div>
                </div>
                <span className="material-symbols-outlined text-slate-500 text-sm">
                  chevron_right
                </span>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default SecurityTelemetryPage;
