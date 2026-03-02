import React, { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import { api } from "../../services/api";
import { Organization } from "../../types";

export const Organizations: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // State
  const [loading, setLoading] = useState(true);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [error, setError] = useState<{
    message: string;
    code?: string;
    traceId?: string;
  } | null>(null);
  const [dropdownOpen, setDropdownOpen] = useState<string | null>(null);
  const { uiStrings } = useUIStrings("admin_orgs");

  // Data State
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [stats, setStats] = useState({
    total_orgs: 0,
    active_orgs: 0,
    pending_orgs: 0,
    inactive_orgs: 0,
    total_users: 0,
    total_vehicles: 0,
    total_devices: 0,
    total_trips: 0,
    active_alerts: 0,
  });

  // Pagination State
  const [page, setPage] = useState(parseInt(searchParams.get("page") || "1"));
  const [pageSize, setPageSize] = useState(
    parseInt(searchParams.get("size") || "10"),
  );
  const [totalItems, setTotalItems] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");

  const dropdownRef = useRef<HTMLDivElement>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // Fetch Stats and Orgs in parallel
      const [orgStats, orgsResponse, globalStats] = await Promise.all([
        api.getOrganizationStats().catch((err) => {
          console.warn("Failed to fetch org stats", err);
          return {}; // Fallback
        }),
        api.getOrganizations({ page, limit: pageSize, search: searchQuery }),
        api.getStatistics().catch((err) => {
          console.warn("Failed to fetch global stats", err);
          return {};
        }),
      ]);

      // Normalize Stats
      // Both getOrganizationStats and getStatistics now return { data: ... }
      const statsData = (orgStats as any)?.data || (globalStats as any)?.data;
      const orgsMeta = (orgsResponse as any)?.meta;
      const orgsList =
        (orgsResponse as any)?.data ||
        (Array.isArray(orgsResponse) ? orgsResponse : []);

      setStats({
        total_orgs: statsData?.total_orgs || orgsMeta?.total || 0,
        active_orgs: statsData?.active_orgs || 0,
        pending_orgs: statsData?.pending_orgs || 0,
        inactive_orgs: statsData?.inactive_orgs || 0,
        total_users: statsData?.total_users || statsData?.total_drivers || 0,
        total_vehicles: statsData?.total_vehicles || 0,
        total_devices: statsData?.total_devices || 0,
        total_trips: statsData?.total_trips || statsData?.active_trips || 0,
        active_alerts: statsData?.active_alerts || statsData?.alerts_count || 0,
      });

      // Normalize Organizations Data
      const total =
        orgsMeta?.total || (orgsResponse as any)?.total || orgsList.length;

      setOrganizations(Array.isArray(orgsList) ? orgsList : []);
      setTotalItems(total);
    } catch (err: any) {
      console.error("Failed to fetch organization data", err);
      setError({
        message: err.message || "Failed to load organization data",
        code: err.code || "UNKNOWN_ERROR",
        traceId: err.traceId || "N/A",
      });
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, searchQuery]);

  // Effects
  useEffect(() => {
    fetchData();
  }, [fetchData, refreshTrigger]);

  useEffect(() => {
    const params: any = { page: page.toString(), size: pageSize.toString() };
    if (searchQuery) params.search = searchQuery;
    setSearchParams(params);
  }, [page, pageSize, searchQuery, setSearchParams]);

  // Handlers
  const handleDelete = async (id: string) => {
    if (window.confirm("Are you sure you want to delete this organization?")) {
      try {
        await api.deleteOrganization(id);
        setRefreshTrigger((prev) => prev + 1);
      } catch (error) {
        alert("Failed to delete organization");
      }
    }
  };

  const toggleDropdown = (orgId: string) => {
    setDropdownOpen(dropdownOpen === orgId ? null : orgId);
  };

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= Math.ceil(totalItems / pageSize)) {
      setPage(newPage);
    }
  };

  const totalPages = Math.ceil(totalItems / pageSize);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setDropdownOpen(null);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  if (loading && !organizations.length) {
    return (
      <div className="p-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-white text-2xl font-bold mb-2">
              {uiStrings?.title || "Organizations"}
            </h1>
            <p className="text-text-muted">
              {uiStrings?.loading || "Loading organization data..."}
            </p>
          </div>
        </div>
        <div className="flex items-center justify-center h-64">
          <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 max-w-4xl mx-auto">
        <div className="bg-red-900/20 border border-red-500/50 rounded-xl p-6">
          <h3 className="text-white text-lg font-bold mb-2">
            Failed to Load Organizations
          </h3>
          <p className="text-text-muted mb-4">{error.message}</p>
          <button
            onClick={() => setRefreshTrigger((prev) => prev + 1)}
            className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg font-medium transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full">
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">
          {uiStrings?.title || "Organization Management"}
        </h1>
        <p className="text-text-muted text-sm">
          {uiStrings?.subtitle ||
            "Manage tenant organizations and their configurations"}
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4 mb-6">
        <div
          onClick={() => navigate("/admin/organizations?status=all")}
          className="bg-surface-dark rounded-xl border border-border-color p-4 hover:border-blue-500/50 transition-all cursor-pointer group"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="material-symbols-outlined text-blue-400 text-xl">
              corporate_fare
            </span>
            <span className="text-xs font-bold text-blue-400 bg-blue-400/20 px-2 py-1 rounded">
              {uiStrings?.stat_total || "ORGANIZATIONS"}
            </span>
          </div>
          <p className="text-2xl font-bold text-white">
            {stats.total_orgs?.toLocaleString()}
          </p>
        </div>
        <div
          onClick={() => navigate("/admin/users")}
          className="bg-surface-dark rounded-xl border border-border-color p-4 hover:border-indigo-500/50 transition-all cursor-pointer group"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="material-symbols-outlined text-indigo-400 text-xl">
              group
            </span>
            <span className="text-xs font-bold text-indigo-400 bg-indigo-400/20 px-2 py-1 rounded">
              {uiStrings?.stat_users || "USERS"}
            </span>
          </div>
          <p className="text-2xl font-bold text-white">
            {stats.total_users?.toLocaleString()}
          </p>
        </div>
        <div
          onClick={() => navigate("/admin/vehicles")}
          className="bg-surface-dark rounded-xl border border-border-color p-4 hover:border-emerald-500/50 transition-all cursor-pointer group"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="material-symbols-outlined text-emerald-400 text-xl">
              directions_car
            </span>
            <span className="text-xs font-bold text-emerald-400 bg-emerald-400/20 px-2 py-1 rounded">
              {uiStrings?.stat_vehicles || "VEHICLES"}
            </span>
          </div>
          <p className="text-2xl font-bold text-white">
            {stats.total_vehicles?.toLocaleString()}
          </p>
        </div>
        <div
          onClick={() => navigate("/admin/devices")}
          className="bg-surface-dark rounded-xl border border-border-color p-4 hover:border-amber-500/50 transition-all cursor-pointer group"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="material-symbols-outlined text-amber-400 text-xl">
              devices
            </span>
            <span className="text-xs font-bold text-amber-400 bg-amber-400/20 px-2 py-1 rounded">
              {uiStrings?.stat_devices || "DEVICES"}
            </span>
          </div>
          <p className="text-2xl font-bold text-white">
            {stats.total_devices?.toLocaleString()}
          </p>
        </div>
        <div
          onClick={() => navigate("/admin/alerts")}
          className="bg-surface-dark rounded-xl border border-border-color p-4 hover:border-red-500/50 transition-all cursor-pointer group"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="material-symbols-outlined text-red-400 text-xl">
              warning
            </span>
            <span className="text-xs font-bold text-red-400 bg-red-400/20 px-2 py-1 rounded">
              {uiStrings?.stat_alerts || "ALERTS"}
            </span>
          </div>
          <p className="text-2xl font-bold text-white">
            {stats.active_alerts?.toLocaleString()}
          </p>
        </div>
      </div>

      {/* Toolbar */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-4 border-b border-border-color bg-surface-darker">
          <div className="flex justify-between items-center mb-4">
            <div className="flex items-center gap-6 whitespace-nowrap">
              <button className="text-sm font-semibold text-white border-b-2 border-primary pb-2">
                {uiStrings?.tab_all || "All Organizations"}
              </button>
            </div>
            <button
              onClick={() => navigate("/admin/organizations/new")}
              className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-semibold rounded-lg flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
            >
              <span className="material-symbols-outlined text-base">add</span>
              {uiStrings?.btn_new_org || "New Organization"}
            </button>
          </div>

          <div className="flex items-center gap-3">
            <div className="relative flex-1 max-w-md">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 material-symbols-outlined text-text-muted">
                search
              </span>
              <input
                type="text"
                placeholder={
                  uiStrings?.placeholder_search || "Search organizations..."
                }
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && setPage(1)} // Reset to page 1 on new search
                className="w-full bg-surface-dark border border-border-color text-sm text-white rounded-lg pl-10 pr-4 py-2 focus:ring-2 focus:ring-primary focus:border-primary outline-none placeholder:text-text-muted/50"
              />
            </div>
            <button
              onClick={() => fetchData()}
              className="p-2 hover:bg-border-color rounded-lg text-text-muted hover:text-white transition-colors"
            >
              <span className="material-symbols-outlined text-lg">refresh</span>
            </button>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden mt-4">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-surface-darker border-b border-border-color">
              <tr>
                <th className="p-4 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_name || "Org Name"}
                </th>
                <th className="p-4 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_domain || "Domain"}
                </th>
                <th className="p-4 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_status || "Status"}
                </th>
                <th className="p-4 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_created || "Created At"}
                </th>
                <th className="p-4 text-left text-xs font-semibold text-text-muted uppercase tracking-wider min-w-[120px]">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {organizations.length === 0 ? (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-text-muted">
                    <div className="flex flex-col items-center gap-2">
                      <span className="material-symbols-outlined text-4xl text-border-color">
                        corporate_fare
                      </span>
                      <p>{uiStrings?.no_orgs || "No organizations found"}</p>
                    </div>
                  </td>
                </tr>
              ) : (
                organizations.map((org) => (
                  <tr
                    key={org.id}
                    className="hover:text-amber-200 hover:bg-surface-dark/10 transition-colors"
                  >
                    <td className="p-4">
                      <div className="font-medium text-white">{org.name}</div>
                    </td>
                    <td className="p-4 text-text-muted">{org.domain}</td>
                    <td className="p-4">
                      <span
                        className={`px-2 py-1 rounded-full text-xs font-semibold border 
                        ${
                          org.status?.toLowerCase() === "active"
                            ? "bg-success/20 text-success border-success/30"
                            : org.status?.toLowerCase() === "suspended"
                              ? "bg-red-500/20 text-red-400 border-red-500/30"
                              : "bg-gray-500/20 text-gray-400 border-gray-500/30"
                        }`}
                      >
                        {(org.status || "UNKNOWN").toUpperCase()}
                      </span>
                    </td>
                    <td className="p-4 text-text-muted">
                      {org.created_at
                        ? new Date(org.created_at).toLocaleDateString()
                        : "N/A"}
                    </td>
                    <td className="p-4 relative">
                      <div ref={dropdownRef}>
                        <button
                          onClick={() => toggleDropdown(org.id)}
                          className="p-2 hover:bg-border-color rounded text-text-muted hover:text-white transition-colors"
                        >
                          <span className="material-symbols-outlined text-[16px]">
                            more_vert
                          </span>
                        </button>

                        {dropdownOpen === org.id && (
                          <div className="absolute right-0 mt-2 w-48 bg-surface-dark border border-border-color rounded-lg shadow-lg z-50">
                            <button
                              onClick={() =>
                                navigate(`/admin/organizations/${org.id}`)
                              }
                              className="w-full px-4 py-3 text-left text-white hover:text-amber-200 hover:bg-surface-dark flex items-center gap-3 text-sm font-medium transition-colors"
                            >
                              <span className="material-symbols-outlined text-base">
                                visibility
                              </span>
                              {uiStrings?.btn_view || "View Details"}
                            </button>
                            <button
                              onClick={() => handleDelete(org.id)}
                              className="w-full px-4 py-3 text-left text-red-400 hover:bg-red-500/20 flex items-center gap-3 text-sm font-medium transition-colors"
                            >
                              <span className="material-symbols-outlined text-base">
                                block
                              </span>
                              {uiStrings?.btn_suspend || "Suspend"}
                            </button>
                          </div>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Dynamic Pagination Controls */}
        <div className="p-4 border-t border-border-color bg-surface-darker flex items-center justify-between">
          <div className="text-sm text-text-muted">
            {uiStrings?.pagination_showing || "Showing"}{" "}
            {(page - 1) * pageSize + 1} {uiStrings?.pagination_to || "to"}{" "}
            {Math.min(page * pageSize, totalItems)}{" "}
            {uiStrings?.pagination_of || "of"} {totalItems}{" "}
            {uiStrings?.pagination_results || "results"}
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <select
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
                  setPage(1); // Reset to first page
                }}
                className="bg-surface-dark border border-border-color text-text-muted text-sm rounded px-2 py-1 outline-none focus:ring-1 focus:ring-primary"
              >
                <option value={10}>
                  10 {uiStrings?.page_size_label || "/ page"}
                </option>
                <option value={20}>
                  20 {uiStrings?.page_size_label || "/ page"}
                </option>
                <option value={50}>
                  50 {uiStrings?.page_size_label || "/ page"}
                </option>
              </select>
            </div>
            <div className="flex items-center gap-1">
              <button
                disabled={page === 1}
                onClick={() => handlePageChange(page - 1)}
                className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <span className="material-symbols-outlined text-sm">
                  chevron_left
                </span>
              </button>

              {/* Simplified pagination numbers */}
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let p = i + 1;
                // Adjust logic to show relevant pages if totalPages > 5
                if (totalPages > 5 && page > 3) {
                  p = page - 2 + i;
                }
                if (p > totalPages) return null;

                return (
                  <button
                    key={p}
                    onClick={() => handlePageChange(p)}
                    className={`px-3 py-1 text-sm rounded transition-colors ${page === p ? "bg-primary text-white font-medium" : "hover:bg-border-color text-white"}`}
                  >
                    {p}
                  </button>
                );
              })}

              <button
                disabled={page === totalPages || totalPages === 0}
                onClick={() => handlePageChange(page + 1)}
                className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <span className="material-symbols-outlined text-sm">
                  chevron_right
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
