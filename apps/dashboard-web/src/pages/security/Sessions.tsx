import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

interface Session {
  id: string;
  user_id: string;
  user_email?: string;
  user_name?: string;
  device_id?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
  expires_at: string;
  is_current?: boolean;
}

const API_BASE = import.meta.env.VITE_API_URL || "/api/v1";

export default function Sessions() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [uiStrings, setUiStrings] = useState<any>(null);
  const [searchParams] = useSearchParams();

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.security_v2 ||
            data.system_metadata?.ui_strings?.security_v2,
        ),
      )
      .catch((err) => console.error("Failed to fetch UI strings", err));
  }, []);
  const orgId = searchParams.get("org_id") || "";

  useEffect(() => {
    fetchSessions();
  }, [orgId]);

  const fetchSessions = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/security/sessions?org_id=${orgId}`);
      const data = await res.json();
      setSessions(data.sessions || []);
    } catch (err) {
      console.error("Failed to fetch sessions:", err);
    }
    setLoading(false);
  };

  const handleRevoke = async (sessionId: string) => {
    const previousSessions = [...sessions];
    // Optimistic Update
    setSessions((prev) => prev.filter((s) => s.id !== sessionId));

    try {
      const res = await fetch(`${API_BASE}/security/sessions/revoke`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: sessionId }),
      });
      if (!res.ok) {
        throw new Error("Failed to revoke session");
      }
      fetchSessions(); // Refresh to ensure strict sync
    } catch (err) {
      console.error("Failed to revoke session:", err);
      // Restore on failure
      setSessions(previousSessions);
      alert(uiStrings?.revoke_failed || "Revocation Failed. Please try again.");
    }
  };

  const handleRevokeAll = async () => {
    if (
      !window.confirm(
        uiStrings?.confirm_revoke_all ||
          "Are you sure you want to revoke all sessions?",
      )
    )
      return;
    try {
      await fetch(`${API_BASE}/security/sessions/revoke-all?org_id=${orgId}`, {
        method: "POST",
      });
      fetchSessions();
    } catch (err) {
      console.error("Failed to revoke all sessions:", err);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString();
  };

  const getDeviceInfo = (userAgent?: string) => {
    if (!userAgent) return uiStrings?.dev_unknown || "Unknown Device";
    if (userAgent.includes("Chrome"))
      return uiStrings?.dev_chrome || "Chrome Browser";
    if (userAgent.includes("Firefox"))
      return uiStrings?.dev_firefox || "Firefox Browser";
    if (userAgent.includes("Safari"))
      return uiStrings?.dev_safari || "Safari Browser";
    if (userAgent.includes("Mobile"))
      return uiStrings?.dev_mobile || "Mobile Device";
    return uiStrings?.dev_desktop || "Desktop";
  };

  return (
    <div className="p-6 bg-gray-900 min-h-screen text-white">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">
            {uiStrings?.sessions_title || "Active Sessions"}
          </h1>
          <p className="text-gray-400">
            {uiStrings?.sessions_subtitle ||
              "Manage user sessions across your organization"}
          </p>
        </div>
        <button
          onClick={handleRevokeAll}
          className="px-4 py-2 bg-red-600 hover:bg-red-700 rounded-lg transition-colors"
        >
          {uiStrings?.revoke_all_btn || "Revoke All Sessions"}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-green-400">
            {sessions.length}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_active_sessions || "Active Sessions"}
          </div>
        </div>
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-blue-400">
            {new Set(sessions.map((s) => s.user_id)).size}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_unique_users || "Unique Users"}
          </div>
        </div>
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-purple-400">
            {sessions.filter((s) => s.device_id?.includes("mobile")).length}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_mobile_sessions || "Mobile Sessions"}
          </div>
        </div>
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-amber-400">
            {
              sessions.filter((s) => {
                const exp = new Date(s.expires_at);
                const now = new Date();
                const diff = (exp.getTime() - now.getTime()) / (1000 * 60 * 60);
                return diff < 1;
              }).length
            }
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_expiring_soon || "Expiring Soon"}
          </div>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      ) : (
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl border border-gray-700 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-700/50 text-left">
                <th className="px-4 py-3 text-gray-300">
                  {uiStrings?.th_user || "User"}
                </th>
                <th className="px-4 py-3 text-gray-300">
                  {uiStrings?.th_device || "Device"}
                </th>
                <th className="px-4 py-3 text-gray-300">
                  {uiStrings?.th_ip_address || "IP Address"}
                </th>
                <th className="px-4 py-3 text-gray-300">
                  {uiStrings?.th_started || "Started"}
                </th>
                <th className="px-4 py-3 text-gray-300">
                  {uiStrings?.th_expires || "Expires"}
                </th>
                <th className="px-4 py-3 text-gray-300">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody>
              {sessions.map((session) => (
                <tr
                  key={session.id}
                  className="border-t border-gray-700 hover:bg-gray-700/30 transition-colors"
                >
                  <td className="px-4 py-3">
                    <div className="font-medium">
                      {session.user_name ||
                        uiStrings?.unknown_label ||
                        "Unknown"}
                    </div>
                    <div className="text-sm text-gray-400">
                      {session.user_email}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-gray-300">
                    {getDeviceInfo(session.user_agent)}
                  </td>
                  <td className="px-4 py-3 text-gray-300 font-mono text-sm">
                    {session.ip_address || "-"}
                  </td>
                  <td className="px-4 py-3 text-gray-300">
                    {formatDate(session.created_at)}
                  </td>
                  <td className="px-4 py-3 text-gray-300">
                    {formatDate(session.expires_at)}
                  </td>
                  <td className="px-4 py-3">
                    {session.is_current ? (
                      <span className="px-2 py-1 bg-green-600/20 text-green-400 rounded text-sm">
                        {uiStrings?.current_label || "Current"}
                      </span>
                    ) : (
                      <button
                        onClick={() => handleRevoke(session.id)}
                        className="px-3 py-1 bg-red-600/20 text-red-400 hover:bg-red-600/40 rounded transition-colors"
                      >
                        {uiStrings?.revoke_btn || "Revoke"}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
              {sessions.length === 0 && (
                <tr>
                  <td
                    colSpan={6}
                    className="px-4 py-8 text-center text-gray-400"
                  >
                    {uiStrings?.no_sessions_found || "No active sessions found"}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
