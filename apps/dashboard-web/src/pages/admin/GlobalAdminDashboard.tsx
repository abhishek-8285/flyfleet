import { formatDistanceToNow, parseISO } from "date-fns";
import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../services/api";
import { useUIStrings } from "../../hooks/useUIStrings";

interface AdminMetrics {
  risk_overview: {
    sla_compliance: number;
    integration_failures: number;
    payment_failures: number;
    security_score: number;
  };
  uptime_trend: number[];
  critical_indicators: {
    open_incidents: number;
    devices_offline: number;
    system_health: number;
    billing_risk: number;
    total_devices: number;
    integration_failures: number;
  };
  operational_metrics: {
    active_orgs: number;
    active_fleet: number;
    connected_devices: number;
    active_users_today: number;
    total_orgs: number;
    new_orgs_today: number;
    total_vehicles: number;
    vehicles_offline: number;
    devices_offline: number;
    users_yesterday: number;
  };
  pending_approvals: number;
  recent_activity: Array<{ id: number; type: string; [key: string]: any }>;
  last_updated: string;
}

export const GlobalAdminDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [metrics, setMetrics] = useState<AdminMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [timeRange, setTimeRange] = useState<
    "today" | "7days" | "30days" | "custom"
  >("today");
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings, loading: uiLoading } = useUIStrings("admin_dashboard");

  const fetchMetrics = useCallback(
    async (showRefreshing = false) => {
      if (showRefreshing) setIsRefreshing(true);
      try {
        const data = await api.getAdminDashboardData(timeRange);
        setMetrics(data);

        // Dynamic Meta
        fetch("/api/v1/ui/global_config")
          .then((res) => res.json())
          .then((data) => setSystemMeta(data.system_metadata))
          .catch((err) => console.error("Failed to fetch system config", err));
      } catch (error) {
        console.error("Failed to fetch admin metrics:", error);
      } finally {
        setLoading(false);
        if (showRefreshing) setIsRefreshing(false);
      }
    },
    [timeRange],
  );

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(() => fetchMetrics(true), 30000); // Background refresh
    return () => clearInterval(interval);
  }, [fetchMetrics]);

  const handleManualRefresh = () => {
    fetchMetrics(true);
  };

  const formatLastUpdated = (timestamp?: string) => {
    if (!timestamp) return "Never";
    try {
      return formatDistanceToNow(parseISO(timestamp), { addSuffix: true });
    } catch (e) {
      return "Just now";
    }
  };

  if (loading || uiLoading) {
    return (
      <div className="p-8 text-center text-text-muted">
        {uiStrings?.loading_dashboard || "Loading system dashboard..."}
      </div>
    );
  }

  return (
    <div className="w-full p-6">
      <div className="mb-8">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-white text-3xl font-bold mb-2">
              {uiStrings?.title}
            </h1>
            <p className="text-text-muted text-lg">
              {uiStrings?.subtitle}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-text-muted text-sm">{uiStrings?.label_time_range}</span>
            <div className="flex bg-surface-dark rounded-lg border border-border-color p-1">
              {(["today", "7days", "30days"] as const).map((range) => (
                <button
                  key={range}
                  onClick={() => setTimeRange(range)}
                  className={`px-3 py-1 text-xs font-medium rounded transition-colors ${
                    timeRange === range
                      ? "bg-primary text-white"
                      : "text-text-muted hover:text-white hover:bg-surface-light"
                  }`}
                >
                  {range === "today"
                    ? uiStrings?.opt_today
                    : range === "7days"
                      ? uiStrings?.opt_7days
                      : uiStrings?.opt_30days}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Platform Risk Overview */}
      <div className="mb-6">
        <div className="bg-surface-dark rounded-xl border border-border-color p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-white font-bold flex items-center gap-2">
              <span className="material-symbols-outlined text-red-400">
                warning
              </span>
              {uiStrings?.risk_title}
            </h3>
            <div className="flex items-center gap-3 text-xs text-text-muted">
              <span className="flex items-center gap-1">
                {uiStrings?.label_last_updated}{" "}
                {formatLastUpdated((metrics as any)?.last_updated)}
                {isRefreshing && (
                  <span className="material-symbols-outlined text-xs animate-spin text-primary">
                    sync
                  </span>
                )}
              </span>
              <button
                onClick={handleManualRefresh}
                className={`p-1 hover:text-white transition-colors ${isRefreshing ? "opacity-50 cursor-not-allowed" : ""}`}
                disabled={isRefreshing}
              >
                <span
                  className={`material-symbols-outlined text-sm ${isRefreshing ? "animate-spin" : ""}`}
                >
                  refresh
                </span>
              </button>
            </div>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-red-900/20 border border-red-500/30 rounded-lg p-3 hover:border-l-4 hover:border-l-red-500 transition-all">
              <div className="flex items-center justify-between mb-1">
                <span className="text-red-400 text-sm font-medium">
                  {uiStrings?.label_sla_compliance}
                </span>
                <span className="material-symbols-outlined text-red-400 text-sm">
                  trending_down
                </span>
              </div>
              <p className="text-2xl font-bold text-white">
                {(metrics?.risk_overview.sla_compliance || 0).toFixed(1)}%
              </p>
              <p className="text-red-400 text-xs">
                {systemMeta?.revenue_diff || `↓ --% ${uiStrings?.msg_vs_yesterday}`}
              </p>
            </div>
            <div className="bg-amber-900/20 border border-amber-500/30 rounded-lg p-3 hover:border-l-4 hover:border-l-amber-500 transition-all">
              <div className="flex items-center justify-between mb-1">
                <span className="text-amber-400 text-sm font-medium">
                  {uiStrings?.label_int_failures}
                </span>
                <span className="material-symbols-outlined text-amber-400 text-sm">
                  error
                </span>
              </div>
              <p className="text-2xl font-bold text-white">
                {metrics?.risk_overview.integration_failures}
              </p>
              <p className="text-amber-400 text-xs">{uiStrings?.label_last_24h}</p>
            </div>
            <div className="bg-surface-light/10 border border-border-color rounded-lg p-3 hover:border-l-4 hover:border-l-green-500 transition-all">
              <div className="flex items-center justify-between mb-1">
                <span className="text-text-muted text-sm font-medium">
                  {uiStrings?.label_pay_failures}
                </span>
                <span className="material-symbols-outlined text-green-400 text-sm">
                  check_circle
                </span>
              </div>
              <p className="text-2xl font-bold text-white">
                {metrics?.risk_overview.payment_failures}
              </p>
              <p className="text-green-400 text-xs">{uiStrings?.msg_all_success}</p>
            </div>
            <div className="bg-surface-light/10 border border-border-color rounded-lg p-3 hover:border-l-4 hover:border-l-green-500 transition-all">
              <div className="flex items-center justify-between mb-1">
                <span className="text-text-muted text-sm font-medium">
                  {uiStrings?.label_security_score}
                </span>
                <span className="material-symbols-outlined text-green-400 text-sm">
                  shield
                </span>
              </div>
              <p className="text-2xl font-bold text-white">
                {(metrics?.risk_overview.security_score || 0).toFixed(1)}%
              </p>
              <p className="text-green-400 text-xs">
                {systemMeta?.uptime_diff || `↑ --% improvement`}
              </p>
            </div>
          </div>
          <div className="mt-4 pt-4 border-t border-border-color">
            <div className="flex items-center justify-between mb-2">
              <span className="text-text-muted text-sm">
                {uiStrings?.uptime_trend_title}
              </span>
              <span className="text-xs text-green-400 font-medium">
                {systemMeta?.uptime || "--%"} {uiStrings?.label_average}
              </span>
            </div>
            <div className="h-12 flex items-end gap-1">
              {(metrics?.uptime_trend || []).map((value: number, i: number) => (
                <div
                  key={i}
                  className={`flex-1 rounded-t ${value >= 99 ? "bg-green-500" : value >= 98 ? "bg-amber-500" : "bg-red-500"}`}
                  style={{ height: `${value - 95}%` }}
                ></div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Critical Risk Indicators - Row 1 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div
          onClick={() => navigate("/admin/incidents")}
          className="bg-surface-dark border-2 border-red-500/50 rounded-xl p-5 text-white cursor-pointer hover:border-l-4 hover:border-l-red-500 transition-all group"
        >
          <div className="flex justify-between items-start mb-2">
            <div>
              <h3 className="text-sm font-medium text-text-muted mb-1">
                {uiStrings?.label_open_incidents}
              </h3>
              <p className="text-3xl font-bold text-white">
                {metrics?.critical_indicators.open_incidents}
              </p>
            </div>
            <div className="flex flex-col items-end">
              <span className="material-symbols-outlined text-red-400 text-2xl group-hover:scale-110 transition-transform">
                report
              </span>
              <span className="text-[10px] text-red-400 bg-red-400/20 px-1.5 py-0.5 rounded mt-1 animate-pulse">
                {uiStrings?.badge_critical}
              </span>
            </div>
          </div>
          <p className="text-xs text-text-muted">
            {metrics?.critical_indicators.integration_failures} {uiStrings?.msg_int_failures}
          </p>
        </div>

        <div
          onClick={() => navigate("/admin/devices?status=offline")}
          className="bg-surface-dark border border-border-color rounded-xl p-5 text-white cursor-pointer hover:border-l-4 hover:border-l-amber-500 transition-all group"
        >
          <div className="flex justify-between items-start mb-2">
            <div>
              <h3 className="text-sm font-medium text-text-muted mb-1">
                {uiStrings?.label_devices_offline}
              </h3>
              <p className="text-3xl font-bold text-white">
                {metrics?.critical_indicators.devices_offline}
              </p>
            </div>
            <span className="material-symbols-outlined text-amber-400 text-2xl group-hover:scale-110 transition-transform">
              cloud_off
            </span>
          </div>
          <p className="text-xs text-text-muted">
            {metrics
              ? `${((metrics.critical_indicators.devices_offline / (metrics.critical_indicators.total_devices || 1)) * 100).toFixed(1)}% ${uiStrings?.msg_offline}`
              : `0% ${uiStrings?.msg_offline}`}
            <br />
            <span className="text-green-400">↓ -2 {uiStrings?.msg_vs_yesterday}</span>
          </p>
        </div>

        <div className="bg-surface-dark border border-border-color rounded-xl p-4 text-white hover:border-l-4 hover:border-l-green-500 transition-all">
          <div className="flex justify-between items-start mb-2">
            <div>
              <h3 className="text-xs font-medium text-text-muted mb-1">
                {uiStrings?.label_system_health}
              </h3>
              <p className="text-2xl font-bold text-white">
                {(metrics?.critical_indicators.system_health || 0).toFixed(1)}%
              </p>
            </div>
            <span className="material-symbols-outlined text-green-400 text-2xl pulse-live">
              monitor_heart
            </span>
          </div>
          <p className="text-[10px] text-text-muted">{uiStrings?.msg_uptime_7d}</p>
        </div>

        <div
          onClick={() => navigate("/admin/billing")}
          className="bg-surface-dark border border-border-color rounded-xl p-5 text-white cursor-pointer hover:border-l-4 hover:border-l-red-500 transition-all group"
        >
          <div className="flex justify-between items-start mb-2">
            <div>
              <h3 className="text-sm font-medium text-text-muted mb-1">
                {uiStrings?.label_billing_risk}
              </h3>
              <p className="text-xl font-bold text-white">
                $
                {(
                  metrics?.critical_indicators.billing_risk || 0
                ).toLocaleString(undefined, {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}
              </p>
            </div>
            <span className="material-symbols-outlined text-red-400 text-2xl group-hover:scale-110 transition-transform">
              payments
            </span>
          </div>
          <p className="text-xs text-text-muted">{uiStrings?.msg_unpaid_invoices}</p>
        </div>
      </div>

      {/* Operational Metrics - Row 2 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <div
          onClick={() => navigate("/admin/organizations")}
          className="bg-surface-dark border border-border-color rounded-xl p-4 text-white cursor-pointer hover:border-l-4 hover:border-l-blue-500 transition-all group"
        >
          <div className="flex justify-between items-start mb-1">
            <div>
              <h3 className="text-xs font-medium text-text-muted mb-1">
                {uiStrings?.label_active_orgs}
              </h3>
              <p className="text-2xl font-bold text-white">
                {metrics?.operational_metrics.total_orgs?.toLocaleString()}
              </p>
            </div>
            <span className="material-symbols-outlined text-blue-400 text-xl group-hover:scale-110 transition-transform">
              corporate_fare
            </span>
          </div>
          <p className="text-[10px] text-text-muted">
            {metrics?.operational_metrics.new_orgs_today} {uiStrings?.msg_new_today}
          </p>
        </div>

        <div
          onClick={() => navigate("/admin/vehicles")}
          className="bg-surface-dark border border-border-color rounded-xl p-4 text-white cursor-pointer hover:border-l-4 hover:border-l-green-500 transition-all group"
        >
          <div className="flex justify-between items-start mb-1">
            <div>
              <h3 className="text-xs font-medium text-text-muted mb-1">
                {uiStrings?.label_active_fleet}
              </h3>
              <p className="text-2xl font-bold text-white">
                {metrics?.operational_metrics.active_fleet?.toLocaleString()}
              </p>
            </div>
            <span className="material-symbols-outlined text-green-400 text-xl group-hover:scale-110 transition-transform">
              directions_car
            </span>
          </div>
          <p className="text-[10px] text-text-muted">
            {metrics?.operational_metrics.vehicles_offline} {uiStrings?.msg_offline}
          </p>
        </div>

        <div
          onClick={() => navigate("/admin/devices")}
          className="bg-surface-dark border border-border-color rounded-xl p-4 text-white cursor-pointer hover:border-l-4 hover:border-l-amber-500 transition-all group"
        >
          <div className="flex justify-between items-start mb-1">
            <div>
              <h3 className="text-xs font-medium text-text-muted mb-1">
                {uiStrings?.label_connected_devices}
              </h3>
              <p className="text-2xl font-bold text-white">
                {metrics?.operational_metrics.connected_devices?.toLocaleString()}
              </p>
            </div>
            <span className="material-symbols-outlined text-amber-400 text-xl group-hover:scale-110 transition-transform">
              devices
            </span>
          </div>
          <p className="text-[10px] text-text-muted">
            {metrics?.operational_metrics.devices_offline} {uiStrings?.msg_offline}
          </p>
        </div>

        <div className="bg-surface-dark border border-border-color rounded-xl p-4 text-white hover:border-l-4 hover:border-l-emerald-500 transition-all">
          <div className="flex justify-between items-start mb-1">
            <div>
              <h3 className="text-xs font-medium text-text-muted mb-1">
                {uiStrings?.label_active_users}
              </h3>
              <p className="text-2xl font-bold text-white">
                {metrics?.operational_metrics.active_users_today?.toLocaleString()}
              </p>
            </div>
            <span className="material-symbols-outlined text-emerald-400 text-xl">
              group
            </span>
          </div>
          <p className="text-[10px] text-text-muted">
            {metrics
              ? `↑ ${(metrics.operational_metrics.active_users_today - metrics.operational_metrics.users_yesterday).toLocaleString()} ${uiStrings?.msg_vs_yesterday}`
              : uiStrings?.msg_platform_engagement}
          </p>
        </div>
      </div>

      {/* Critical Admin Actions - Enhanced Hierarchy */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Approval Queue */}
        <div className="bg-surface-dark rounded-xl border border-border-color p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-white text-xl font-bold">{uiStrings?.title_approvals}</h2>
            <span className="material-symbols-outlined text-amber-400">
              hourglass_full
            </span>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between p-3 bg-amber-500/10 rounded-lg border border-amber-500/20">
              <div>
                <p className="text-white font-medium">
                  {systemMeta?.pending_approval_1_name || "---"}
                </p>
                <p className="text-text-muted text-sm">
                  Pending since {systemMeta?.pending_approval_1_since || "---"}
                </p>
              </div>
              <div className="flex gap-2">
                <button className="px-3 py-1 bg-success text-white text-xs rounded hover:bg-green-600 transition-colors">
                  Approve
                </button>
                <button className="px-3 py-1 bg-error text-white text-xs rounded hover:bg-red-600 transition-colors">
                  Reject
                </button>
              </div>
            </div>
            <div className="flex items-center justify-between p-3 bg-amber-500/10 rounded-lg border border-amber-500/20">
              <div>
                <p className="text-white font-medium">
                  {systemMeta?.pending_approval_2_name || "---"}
                </p>
                <p className="text-text-muted text-sm">
                  Pending since {systemMeta?.pending_approval_2_since || "---"}
                </p>
              </div>
              <div className="flex gap-2">
                <button className="px-3 py-1 bg-success text-white text-xs rounded hover:bg-green-600 transition-colors">
                  Approve
                </button>
                <button className="px-3 py-1 bg-error text-white text-xs rounded hover:bg-red-600 transition-colors">
                  Reject
                </button>
              </div>
            </div>
          </div>
          <button
            onClick={() => navigate("/admin/organizations?status=pending")}
            className="w-full mt-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-lg font-medium transition-colors"
          >
            {uiStrings?.btn_view_pending?.replace("{count}", String(metrics?.pending_approvals))}
          </button>
        </div>

        {/* Enhanced Incident Management - Urgent Focus */}
        <div className="bg-red-900/30 border-2 border-red-500 rounded-xl p-6 shadow-2xl shadow-red-500/20">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <span className="material-symbols-outlined text-red-400 text-xl animate-pulse">
                warning
              </span>
              <h2 className="text-white text-xl font-bold">{uiStrings?.title_incidents}</h2>
              <span className="px-2 py-0.5 bg-red-500 text-white text-xs font-bold rounded-full animate-pulse">
                {uiStrings?.badge_urgent}
              </span>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-xs text-text-muted">{uiStrings?.label_refresh_30s}</span>
              <button className="p-1.5 bg-red-500 hover:bg-red-600 rounded text-white transition-colors">
                <span className="material-symbols-outlined text-sm">
                  refresh
                </span>
              </button>
            </div>
          </div>
          <div className="space-y-3 max-h-64 overflow-y-auto mb-4 custom-scrollbar">
            {(
              (metrics?.recent_activity || []).filter(
                (a) => a.type === "incident_opened",
              ) || []
            ).length > 0 ? (
              (metrics?.recent_activity || [])
                .filter((a) => a.type === "incident_opened")
                .map((incident: any) => (
                  <div
                    key={incident.id}
                    className={`p-3 rounded-lg border-2 bg-surface-light/5 hover:bg-surface-light/10 transition-all ${
                      incident.severity === "high"
                        ? "bg-red-500/10 border-red-500/30"
                        : incident.severity === "medium"
                          ? "bg-amber-500/10 border-amber-500/30"
                          : "bg-emerald-500/10 border-emerald-500/30"
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <p className="text-white font-medium">
                            {incident.title}
                          </p>
                          <span
                            className={`text-[10px] px-2 py-0.5 rounded-full font-bold ${
                              incident.severity === "high"
                                ? "bg-red-500 text-white"
                                : incident.severity === "medium"
                                  ? "bg-amber-500 text-white"
                                  : "bg-emerald-500 text-white"
                            }`}
                          >
                            {(incident.severity || "LOW").toUpperCase()}
                          </span>
                        </div>
                        <p className="text-text-muted text-sm">
                          {incident.description || "System generated incident"}
                        </p>
                        <div className="flex items-center gap-3 mt-2 text-xs">
                          <span className="text-text-muted">
                            {incident.timestamp}
                          </span>
                          <span className="text-blue-400">
                            {incident.performer}
                          </span>
                          <span className="text-emerald-400 font-medium">
                            SLA: OK
                          </span>
                        </div>
                      </div>
                      <div className="flex flex-col items-end gap-2 ml-3">
                        <span
                          className={`material-symbols-outlined ${
                            incident.severity === "high"
                              ? "text-red-400 text-xl"
                              : incident.severity === "medium"
                                ? "text-amber-400 text-xl"
                                : "text-emerald-400 text-xl"
                          }`}
                        >
                          {incident.severity === "high"
                            ? "error"
                            : incident.severity === "medium"
                              ? "warning"
                              : "info"}
                        </span>
                      </div>
                    </div>
                  </div>
                ))
            ) : (
              <div className="p-4 text-center text-text-muted italic">
                {uiStrings?.msg_no_incidents}
              </div>
            )}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => navigate("/admin/incidents")}
              className="flex-1 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg font-medium transition-all hover:scale-[1.02] flex items-center justify-center gap-2 active:scale-95"
            >
              <span className="material-symbols-outlined text-sm">
                manage_accounts
              </span>
              {uiStrings?.btn_manage_incidents?.replace("{count}", String(metrics?.critical_indicators.open_incidents))}
            </button>
            <button className="px-4 py-2 bg-surface-dark hover:bg-border-color text-white rounded-lg font-medium transition-colors border border-border-color">
              <span className="material-symbols-outlined text-sm">
                filter_list
              </span>
            </button>
          </div>
        </div>
      </div>

      {/* Interactive Activity Feed */}
      <div className="bg-surface-dark rounded-xl border border-border-color p-6 mb-8">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-white text-xl font-bold">{uiStrings?.title_activity}</h2>
          <div className="flex items-center gap-2">
            <div className="relative">
              <span className="material-symbols-outlined absolute left-2 top-1/2 -translate-y-1/2 text-text-muted text-sm">
                search
              </span>
              <input
                type="text"
                placeholder={uiStrings?.placeholder_search_activity}
                className="bg-surface-dark border border-border-color text-xs text-white rounded pl-8 pr-3 py-1 w-40 focus:ring-2 focus:ring-primary focus:border-primary outline-none placeholder:text-text-muted/50"
              />
            </div>
            <button className="p-1.5 bg-surface-dark hover:bg-border-color rounded text-text-muted hover:text-white transition-colors border border-border-color">
              <span className="material-symbols-outlined text-sm">
                filter_list
              </span>
            </button>
            <span className="material-symbols-outlined text-blue-400">
              timeline
            </span>
          </div>
        </div>
        <div className="space-y-2 max-h-80 overflow-y-auto mb-4">
          {(metrics?.recent_activity || []).map((activity: any) => (
            <div
              key={activity.id}
              className="flex items-start gap-3 p-3 bg-surface-light/5 hover:bg-surface-light/10 rounded-lg border border-border-color/30 hover:border-border-color/60 transition-all cursor-pointer group"
            >
              <div className="mt-1 flex-shrink-0">
                {activity.type === "org_created" && (
                  <span className="material-symbols-outlined text-green-400 text-lg group-hover:scale-110 transition-transform">
                    add_business
                  </span>
                )}
                {activity.type === "user_added" && (
                  <span className="material-symbols-outlined text-blue-400 text-lg group-hover:scale-110 transition-transform">
                    person_add
                  </span>
                )}
                {activity.type === "device_registered" && (
                  <span className="material-symbols-outlined text-amber-400 text-lg group-hover:scale-110 transition-transform">
                    devices
                  </span>
                )}
                {activity.type === "incident_opened" && (
                  <span className="material-symbols-outlined text-red-400 text-lg group-hover:scale-110 transition-transform">
                    report
                  </span>
                )}
                {activity.type === "subscription_expired" && (
                  <span className="material-symbols-outlined text-orange-400 text-lg group-hover:scale-110 transition-transform">
                    event_busy
                  </span>
                )}
                {activity.type === "payment_received" && (
                  <span className="material-symbols-outlined text-emerald-400 text-lg group-hover:scale-110 transition-transform">
                    payments
                  </span>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-white text-sm font-medium group-hover:text-primary transition-colors">
                  {activity.type === "org_created" &&
                    `New organization created: ${activity.org_name}`}
                  {activity.type === "user_added" &&
                    `User added: ${activity.user_name} to ${activity.org}`}
                  {activity.type === "device_registered" &&
                    `Device registered: ${activity.device_id} for ${activity.vehicle}`}
                  {activity.type === "incident_opened" &&
                    `Incident opened: ${activity.title}`}
                  {activity.type === "subscription_expired" &&
                    `Subscription expired: ${activity.org_name}`}
                  {activity.type === "payment_received" &&
                    `Payment received: ${activity.amount} from ${activity.org}`}
                </p>
                <div className="flex flex-wrap items-center gap-2 mt-1 text-xs">
                  <span className="text-text-muted">{activity.timestamp}</span>
                  <span className="text-blue-400 font-medium">
                    by {activity.performer}
                  </span>
                  {activity.user && (
                    <span className="text-amber-400 bg-amber-400/10 px-1.5 py-0.5 rounded">
                      initiated by {activity.user}
                    </span>
                  )}
                  {activity.amount && (
                    <span className="text-emerald-400 font-bold">
                      {activity.amount}
                    </span>
                  )}
                  <button className="text-primary hover:text-primary-dark text-[10px] font-medium transition-colors ml-auto">
                    View details →
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => navigate("/admin/activity")}
            className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">history</span>
            {uiStrings?.btn_full_log}
          </button>
          <button className="px-4 py-2 bg-surface-dark hover:bg-border-color text-white rounded-lg font-medium transition-colors border border-border-color">
            <span className="material-symbols-outlined text-sm">download</span>
          </button>
        </div>
      </div>

      {/* Quick Navigation */}
      <div className="bg-surface-dark rounded-xl border border-border-color p-6">
        <h2 className="text-white text-xl font-bold mb-6">{uiStrings?.title_tools}</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <button
            onClick={() => navigate("/admin/organizations")}
            className="bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/30 rounded-lg p-4 text-center transition-all group"
          >
            <span className="material-symbols-outlined text-blue-400 text-2xl mb-2 block">
              corporate_fare
            </span>
            <p className="text-white font-medium">{uiStrings?.tool_orgs}</p>
            <p className="text-text-muted text-xs mt-1">{uiStrings?.tool_orgs_desc}</p>
          </button>

          <button
            onClick={() => navigate("/admin/users")}
            className="bg-indigo-500/10 hover:bg-indigo-500/20 border border-indigo-500/30 rounded-lg p-4 text-center transition-all group"
          >
            <span className="material-symbols-outlined text-indigo-400 text-2xl mb-2 block">
              group
            </span>
            <p className="text-white font-medium">{uiStrings?.tool_users}</p>
            <p className="text-text-muted text-xs mt-1">{uiStrings?.tool_users_desc}</p>
          </button>

          <button
            onClick={() => navigate("/admin/billing")}
            className="bg-green-500/10 hover:bg-green-500/20 border border-green-500/30 rounded-lg p-4 text-center transition-all group"
          >
            <span className="material-symbols-outlined text-green-400 text-2xl mb-2 block">
              payments
            </span>
            <p className="text-white font-medium">{uiStrings?.tool_billing}</p>
            <p className="text-text-muted text-xs mt-1">{uiStrings?.tool_billing_desc}</p>
          </button>

          <button
            onClick={() => navigate("/admin/security")}
            className="bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 rounded-lg p-4 text-center transition-all group"
          >
            <span className="material-symbols-outlined text-red-400 text-2xl mb-2 block">
              security
            </span>
            <p className="text-white font-medium">{uiStrings?.tool_security}</p>
            <p className="text-text-muted text-xs mt-1">{uiStrings?.tool_security_desc}</p>
          </button>
        </div>
      </div>
    </div>
  );
};
