import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import type { RootState } from "../store";
import {
    Area,
    AreaChart,
    CartesianGrid,
    Cell,
    Pie,
    PieChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import SecurityFooter from "../components/SecurityFooter";
import SecuritySidebar from "../components/SecuritySidebar";
import api from "../services/api";

// Fix for Recharts TS2786 with React 18
const AreaAny = Area as any;
const AreaChartAny = AreaChart as any;
const CartesianGridAny = CartesianGrid as any;
const CellAny = Cell as any;
const PieAny = Pie as any;
const PieChartAny = PieChart as any;
const TooltipAny = Tooltip as any;
const XAxisAny = XAxis as any;
const YAxisAny = YAxis as any;

const MFACenter: React.FC = () => {
  const navigate = useNavigate();
  const user = useSelector((state: RootState) => state.auth.user);
  const [policy, setPolicy] = useState<any>(null);
  const [summary, setSummary] = useState<any>(null);
  const [trend, setTrend] = useState<any[]>([]);
  const [methods, setMethods] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [mfaMeta, setMfaMeta] = useState<any>(null);

  const orgId = user?.org_id || localStorage.getItem("org_id") || "";

  useEffect(() => {
    if (!orgId) return;

    const fetchData = async () => {
      try {
        const [policyRes, summaryRes, trendRes, methodsRes] = await Promise.all(
          [
            api.getMFAPolicy(orgId),
            api.getMFADashboard(orgId),
            api.getMFAAdoptionTrend(orgId, 14),
            api.getMFAMethodDistribution(orgId),
          ],
        );

        setPolicy(policyRes);
        setSummary(summaryRes);
        setTrend(trendRes);

        // Transform methods distribution for PieChart
        const methodData = Object.entries(methodsRes).map(([name, value]) => ({
          name: name.toUpperCase(),
          value,
        }));
        setMethods(methodData);
      } catch (err) {
        console.error("Failed to fetch security data", err);
      } finally {
        setLoading(false);
        fetch("/api/v1/ui/global_config")
          .then((res) => res.json())
          .then((data) => setMfaMeta(data.mfa_metadata))
          .catch((err) => console.error("Failed to fetch MFA config", err));
      }
    };
    fetchData();
  }, [orgId]);

  const [COLORS, setColors] = useState<any[]>([]);
  useEffect(() => {
    fetch("/api/v1/ui/misc/COLORS")
      .then((res) => res.json())
      .then((data) => setColors(data))
      .catch((err) => console.error(err));
  }, []);

  return (
    <div className="flex h-screen w-full bg-background-dark text-white font-sans">
      <SecuritySidebar />
      <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden relative">
        {/* Background decorative elements */}
        <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] -z-10 pointer-events-none"></div>
        <div className="absolute bottom-0 left-0 w-[300px] h-[300px] bg-secondary/5 rounded-full blur-[100px] -z-10 pointer-events-none"></div>

        <header className="h-16 flex items-center justify-between px-6 border-b border-white/5 bg-surface-dark/40 backdrop-blur-md flex-shrink-0 z-20">
          <div className="flex items-center gap-4">
            <div>
              <h2 className="text-white text-lg font-bold tracking-tight">
                Security Command Center
              </h2>
              <p className="text-[11px] text-text-muted mt-0.5 flex items-center gap-1">
                <span className="material-symbols-outlined text-[12px]">
                  security
                </span>
                MFA Intelligence & Infrastructure Control
              </p>
            </div>
            <div className="h-6 w-px bg-white/10 mx-2"></div>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-widest bg-primary/20 text-primary border border-primary/30 animate-pulse">
              Live Status
            </span>
          </div>
          <div className="flex items-center gap-6">
            <div className="hidden lg:flex flex-col items-end">
              <span className="text-[10px] text-text-muted uppercase tracking-tighter">
                System Pulse
              </span>
              <div className="flex items-center gap-1.5">
                <div className="w-1.5 h-1.5 rounded-full bg-success animate-ping"></div>
                <span className="text-xs font-mono text-success">Healthy</span>
              </div>
            </div>
            <button
              onClick={() => navigate("/security/policy")}
              className="px-4 py-2 bg-gradient-to-r from-primary to-primary-dark hover:from-primary-dark hover:to-primary text-white text-xs font-bold rounded-lg flex items-center gap-2 transition-all transform hover:scale-[1.02] active:scale-[0.98] shadow-lg shadow-primary/25 border border-white/10"
            >
              <span className="material-symbols-outlined text-[18px]">
                tune
              </span>
              Edit Security Policy
            </button>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto p-6 custom-scrollbar relative">
          {loading ? (
            <div className="flex flex-col items-center justify-center h-full gap-4">
              <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
              <p className="text-sm text-text-muted font-medium animate-pulse">
                Synchronizing Cryptographic Assets...
              </p>
            </div>
          ) : (
            <div className="space-y-6 max-w-[1600px] mx-auto">
              {/* Stats Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
                <div className="bg-surface-dark/60 backdrop-blur-sm rounded-2xl border border-white/5 p-5 relative overflow-hidden group hover:border-white/10 transition-all shadow-xl">
                  <div className="absolute -right-4 -top-4 w-24 h-24 bg-primary/5 rounded-full blur-2xl group-hover:bg-primary/10 transition-colors"></div>
                  <div className="flex items-center justify-between mb-3">
                    <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
                      <span className="material-symbols-outlined">
                        verified_user
                      </span>
                    </div>
                    <span className="text-[10px] font-bold text-success bg-success/10 px-2 py-0.5 rounded-full">
                      {mfaMeta?.adoption_growth || "+--%"}
                    </span>
                  </div>
                  <div className="text-text-muted text-xs font-bold uppercase tracking-widest mb-1">
                    MFA Adoption
                  </div>
                  <div className="flex items-baseline gap-2">
                    <div className="text-3xl font-black text-white">
                      {Math.round(
                        (summary?.mfa_enabled / summary?.total_users) * 100,
                      ) || 0}
                      %
                    </div>
                    <div className="text-sm text-text-muted font-medium">
                      of {summary?.total_users || 0} users
                    </div>
                  </div>
                </div>

                <div className="bg-surface-dark/60 backdrop-blur-sm rounded-2xl border border-white/5 p-5 relative overflow-hidden group hover:border-white/10 transition-all shadow-xl">
                  <div className="absolute -right-4 -top-4 w-24 h-24 bg-warning/5 rounded-full blur-2xl group-hover:bg-warning/10 transition-colors"></div>
                  <div className="flex items-center justify-between mb-3">
                    <div className="w-10 h-10 rounded-xl bg-warning/10 flex items-center justify-center text-warning border border-warning/20">
                      <span className="material-symbols-outlined">
                        emergency_home
                      </span>
                    </div>
                  </div>
                  <div className="text-text-muted text-xs font-bold uppercase tracking-widest mb-1">
                    Pending setup
                  </div>
                  <div className="text-3xl font-black text-warning">
                    {summary?.mfa_pending || 0}
                  </div>
                  <div className="text-[10px] text-text-muted mt-2 flex items-center gap-1">
                    <span className="material-symbols-outlined text-[14px]">
                      info
                    </span>
                    Required action within {policy?.grace_period_days || 7} days
                  </div>
                </div>

                <div className="bg-surface-dark/60 backdrop-blur-sm rounded-2xl border border-white/5 p-5 relative overflow-hidden group hover:border-white/10 transition-all shadow-xl">
                  <div className="absolute -right-4 -top-4 w-24 h-24 bg-danger/5 rounded-full blur-2xl group-hover:bg-danger/10 transition-colors"></div>
                  <div className="flex items-center justify-between mb-3">
                    <div className="w-10 h-10 rounded-xl bg-danger/10 flex items-center justify-center text-danger border border-danger/20">
                      <span className="material-symbols-outlined">
                        lock_reset
                      </span>
                    </div>
                    <span className="text-[10px] font-bold text-danger bg-danger/10 px-2 py-0.5 rounded-full">
                      ACTION REQ
                    </span>
                  </div>
                  <div className="text-text-muted text-xs font-bold uppercase tracking-widest mb-1">
                    Locked Accounts
                  </div>
                  <div className="text-3xl font-black text-danger">
                    {summary?.locked_accounts || 0}
                  </div>
                  <div className="text-[10px] text-text-muted mt-2 flex items-center gap-1">
                    Due to multiple MFA failures
                  </div>
                </div>

                <div className="bg-surface-dark/60 backdrop-blur-sm rounded-2xl border border-white/5 p-5 relative overflow-hidden group hover:border-white/10 transition-all shadow-xl">
                  <div className="absolute -right-4 -top-4 w-24 h-24 bg-info/5 rounded-full blur-2xl group-hover:bg-info/10 transition-colors"></div>
                  <div className="flex items-center justify-between mb-3">
                    <div className="w-10 h-10 rounded-xl bg-info/10 flex items-center justify-center text-info border border-info/20">
                      <span className="material-symbols-outlined">
                        history_toggle_off
                      </span>
                    </div>
                  </div>
                  <div className="text-text-muted text-xs font-bold uppercase tracking-widest mb-1">
                    Failed (24h)
                  </div>
                  <div className="text-3xl font-black text-white">
                    {summary?.failed_attempts_24h || 0}
                  </div>
                  <div className="text-[10px] text-text-muted mt-2">
                    Unique IP signatures detected
                  </div>
                </div>
              </div>

              {/* Charts Area */}
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2 bg-surface-dark/60 backdrop-blur-sm rounded-2xl border border-white/5 p-6 shadow-2xl">
                  <div className="flex items-center justify-between mb-6">
                    <div>
                      <h3 className="text-white font-bold text-base">
                        Adoption Trendline
                      </h3>
                      <p className="text-[11px] text-text-muted mt-1">
                        Daily enrollment and verification success rates
                      </p>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-primary shadow-[0_0_8px_rgba(19,127,236,0.5)]"></div>
                        <span className="text-[10px] text-text-muted font-bold uppercase tracking-widest">
                          Enrolled
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-success shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>
                        <span className="text-[10px] text-text-muted font-bold uppercase tracking-widest">
                          Active
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="h-[300px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChartAny data={trend}>
                        <defs>
                          <linearGradient
                            id="colorValue"
                            x1="0"
                            y1="0"
                            x2="0"
                            y2="1"
                          >
                            <stop
                              offset="5%"
                              stopColor="#137fec"
                              stopOpacity={0.3}
                            />
                            <stop
                              offset="95%"
                              stopColor="#137fec"
                              stopOpacity={0}
                            />
                          </linearGradient>
                          <linearGradient
                            id="colorActive"
                            x1="0"
                            y1="0"
                            x2="0"
                            y2="1"
                          >
                            <stop
                              offset="5%"
                              stopColor="#10b981"
                              stopOpacity={0.2}
                            />
                            <stop
                              offset="95%"
                              stopColor="#10b981"
                              stopOpacity={0}
                            />
                          </linearGradient>
                        </defs>
                        <CartesianGridAny
                          strokeDasharray="3 3"
                          vertical={false}
                          stroke="rgba(255,255,255,0.05)"
                        />
                        <XAxisAny
                          dataKey="date"
                          axisLine={false}
                          tickLine={false}
                          tick={{ fill: "rgba(255,255,255,0.4)", fontSize: 10 }}
                          dy={10}
                        />
                        <YAxisAny
                          axisLine={false}
                          tickLine={false}
                          tick={{ fill: "rgba(255,255,255,0.4)", fontSize: 10 }}
                        />
                        <TooltipAny
                          contentStyle={{
                            backgroundColor: "#1a2633",
                            border: "1px solid rgba(255,255,255,0.1)",
                            borderRadius: "12px",
                            fontSize: "11px",
                          }}
                          itemStyle={{ fontSize: "11px", fontWeight: "bold" }}
                        />
                        <AreaAny
                          type="monotone"
                          dataKey="enrolled"
                          stroke="#137fec"
                          strokeWidth={3}
                          fillOpacity={1}
                          fill="url(#colorValue)"
                        />
                        <AreaAny
                          type="monotone"
                          dataKey="active"
                          stroke="#10b981"
                          strokeWidth={2}
                          fillOpacity={1}
                          fill="url(#colorActive)"
                          strokeDasharray="5 5"
                        />
                      </AreaChartAny>
                    </ResponsiveContainer>
                  </div>
                </div>

                <div className="bg-surface-dark/60 backdrop-blur-sm rounded-2xl border border-white/5 p-6 shadow-2xl flex flex-col">
                  <h3 className="text-white font-bold text-base mb-1">
                    MFA Methods
                  </h3>
                  <p className="text-[11px] text-text-muted mb-6">
                    Distribution of authentication factors
                  </p>
                  <div className="flex-1 min-h-[200px]">
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChartAny>
                        <PieAny
                          data={methods}
                          cx="50%"
                          cy="50%"
                          innerRadius={60}
                          outerRadius={80}
                          paddingAngle={8}
                          dataKey="value"
                        >
                          {methods.map((entry, index) => (
                            <CellAny
                              key={`cell-${index}`}
                              fill={COLORS[index % COLORS.length]}
                              stroke="none"
                            />
                          ))}
                        </PieAny>
                        <TooltipAny
                          contentStyle={{
                            backgroundColor: "#1a2633",
                            border: "1px solid rgba(255,255,255,0.1)",
                            borderRadius: "12px",
                            fontSize: "11px",
                          }}
                        />
                      </PieChartAny>
                    </ResponsiveContainer>
                  </div>
                  <div className="space-y-2.5 mt-4">
                    {methods.map((m, idx) => (
                      <div
                        key={m.name}
                        className="flex items-center justify-between group"
                      >
                        <div className="flex items-center gap-2">
                          <div
                            className="size-2 rounded-full"
                            style={{
                              backgroundColor: COLORS[idx % COLORS.length],
                            }}
                          ></div>
                          <span className="text-xs text-text-muted group-hover:text-white transition-colors">
                            {m.name}
                          </span>
                        </div>
                        <span className="text-xs font-mono font-bold text-white">
                          {m.value}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Secondary View Links */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div
                  onClick={() => navigate("/security/overrides")}
                  className="group bg-gradient-to-br from-surface-dark/60 to-surface-darker/60 p-6 rounded-2xl border border-white/5 hover:border-primary/40 cursor-pointer transition-all shadow-xl"
                >
                  <div className="flex items-start justify-between">
                    <div className="p-3 rounded-xl bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white transition-all">
                      <span className="material-symbols-outlined">domain</span>
                    </div>
                    <span className="material-symbols-outlined text-white/20 group-hover:text-primary group-hover:translate-x-1 transition-all">
                      arrow_forward_ios
                    </span>
                  </div>
                  <h4 className="mt-4 text-white font-bold">
                    Organization Overrides
                  </h4>
                  <p className="text-xs text-text-muted mt-1 leading-relaxed">
                    Customize security protocols for specific sub-fleets or
                    regional offices without altering global policy.
                  </p>
                </div>

                <div
                  onClick={() => navigate("/security/exceptions")}
                  className="group bg-gradient-to-br from-surface-dark/60 to-surface-darker/60 p-6 rounded-2xl border border-white/5 hover:border-warning/40 cursor-pointer transition-all shadow-xl"
                >
                  <div className="flex items-start justify-between">
                    <div className="p-3 rounded-xl bg-warning/10 text-warning group-hover:bg-warning group-hover:text-white transition-all">
                      <span className="material-symbols-outlined">
                        rule_folder
                      </span>
                    </div>
                    <span className="material-symbols-outlined text-white/20 group-hover:text-warning group-hover:translate-x-1 transition-all">
                      arrow_forward_ios
                    </span>
                  </div>
                  <h4 className="mt-4 text-white font-bold">
                    Exception Registry
                  </h4>
                  <p className="text-xs text-text-muted mt-1 leading-relaxed">
                    Manage temporary MFA bypasses for hardware failures, VIP
                    accounts, or specialized service terminals.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
        <SecurityFooter />
      </main>
    </div>
  );
};

export default MFACenter;
