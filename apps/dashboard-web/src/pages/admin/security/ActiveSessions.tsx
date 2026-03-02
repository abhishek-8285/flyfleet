import {
    ArrowPathIcon,
    ComputerDesktopIcon,
    DevicePhoneMobileIcon,
    GlobeAltIcon,
    ShieldCheckIcon,
    TrashIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import SecurityFooter from "../../../components/SecurityFooter";
import SecuritySidebar from "../../../components/SecuritySidebar";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

const ActiveSessions: React.FC = () => {
  const { uiStrings: rawUiStrings } = useUIStrings("admin_security");
  const uiStrings = rawUiStrings || {};
  const [sessions, setSessions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [revoking, setRevoking] = useState<string | null>(null);

  const orgId = localStorage.getItem("org_id") || "default";

  useEffect(() => {
    fetchSessions();
  }, [orgId]);

  const fetchSessions = async () => {
    try {
      setLoading(true);
      const response = (await apiClient.get("/security/sessions", {
        params: { org_id: orgId },
      })) as any;
      const data = response?.data ?? response;
      setSessions(Array.isArray(data) ? data : data?.items || []);
    } catch (error) {
      console.error("Failed to fetch sessions", error);
      setSessions([]);
    } finally {
      setLoading(false);
    }
  };

  const handleRevoke = async (sessionId: string) => {
    if (
      !window.confirm(
        uiStrings.confirm_terminate ||
          "Are you sure you want to terminate this session? The user will be logged out immediately.",
      )
    )
      return;

    try {
      setRevoking(sessionId);
      await apiClient.delete(`/security/sessions/${sessionId}`);
      setSessions((prev) => prev.filter((s) => (s.id || s._id) !== sessionId));
    } catch (error) {
      alert(uiStrings.error_terminate || "Failed to terminate session.");
    } finally {
      setRevoking(null);
    }
  };

  const handleRevokeAll = async () => {
    if (
      !window.confirm(
        uiStrings.confirm_terminate_all ||
          "Terminate all other active sessions? This will force re-authentication across all devices.",
      )
    )
      return;

    try {
      setLoading(true);
      await apiClient.delete("/security/sessions/all-others", {
        params: { org_id: orgId },
      });
      await fetchSessions();
    } catch (error) {
      alert("Failed to revoke all sessions.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex h-screen bg-[#0b1117] text-slate-200 overflow-hidden font-inter">
      <SecuritySidebar />

      <main className="flex-1 flex flex-col min-w-0 bg-grid-pattern relative">
        <div className="glow-effect" />

        <header className="px-8 py-6 border-b border-white/5 bg-slate-900/40 backdrop-blur-md z-10 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-white flex items-center gap-3">
              <span className="material-symbols-outlined text-primary text-3xl">
                devices
              </span>
              {uiStrings.title_sessions || "Active Auth Sessions"}
            </h2>
            <p className="text-slate-400 text-sm mt-1">
              {uiStrings.subtitle_sessions ||
                "Authorized access tokens and hardware signatures currently active."}
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={handleRevokeAll}
              className="flex items-center gap-2 px-4 py-2 border border-vibrant-red/20 bg-vibrant-red/5 hover:bg-vibrant-red/10 text-vibrant-red rounded-lg text-xs font-black uppercase tracking-widest transition-all"
            >
              <TrashIcon className="w-4 h-4" />
              {uiStrings.btn_revoke_all || "REVOKE ALL OTHER SESSIONS"}
            </button>
            <button
              onClick={fetchSessions}
              className="p-2 text-slate-400 hover:text-white transition-colors"
            >
              <ArrowPathIcon className="w-5 h-5" />
            </button>
          </div>
        </header>

        <div className="p-8 z-10 flex flex-col gap-6 overflow-auto custom-scrollbar flex-1">
          {loading ? (
            <div className="py-20 text-center text-slate-500 italic">
              Decrypting active session signatures...
            </div>
          ) : sessions.length === 0 ? (
            <div className="py-20 text-center text-slate-500">
              No active sessions found.
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4">
              {sessions.map((session) => (
                <div
                  key={session.id || session._id}
                  className="glass-card p-6 rounded-2xl flex items-center gap-6 border-l-4 border-primary/50 group"
                >
                  <div className="w-14 h-14 rounded-xl bg-slate-800 flex items-center justify-center text-primary border border-white/5">
                    {session.device_type === "mobile" ? (
                      <DevicePhoneMobileIcon className="w-8 h-8" />
                    ) : (
                      <ComputerDesktopIcon className="w-8 h-8" />
                    )}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3">
                      <h4 className="font-bold text-white truncate">
                        {session.user_name || session.email}
                      </h4>
                      {session.is_current && (
                        <span className="px-2 py-0.5 rounded bg-vibrant-blue/10 text-vibrant-blue text-[8px] font-black border border-vibrant-blue/20 uppercase tracking-widest">
                          {uiStrings.badge_current_device || "CURRENT DEVICE"}
                        </span>
                      )}
                    </div>
                    <div className="flex flex-wrap gap-x-6 gap-y-1 mt-2">
                      <div className="flex items-center gap-1.5 text-xs text-slate-400">
                        <GlobeAltIcon className="w-3.5 h-3.5" />
                        <span className="font-medium">
                          {session.browser || "Browser"} on {session.os || "OS"}
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5 text-xs text-slate-400">
                        <span className="material-symbols-outlined text-sm">
                          location_on
                        </span>
                        <span className="font-medium">
                          {session.location || "Unknown Location"} ({session.ip}
                          )
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5 text-xs text-slate-400 font-mono">
                        <ShieldCheckIcon className="w-3.5 h-3.5 text-emerald-500" />
                        <span>{session.id?.substring(0, 12)}...</span>
                      </div>
                    </div>
                  </div>

                  <div className="text-right flex flex-col items-end gap-2">
                    <p className="text-[10px] text-slate-500 uppercase font-black tracking-widest">
                      {uiStrings.label_last_active || "Last active:"}
                    </p>
                    <p className="text-sm font-bold text-white">
                      {session.last_active || "Just now"}
                    </p>
                    <button
                      onClick={() => handleRevoke(session.id || session._id)}
                      disabled={revoking === (session.id || session._id)}
                      className="mt-2 text-[10px] font-black text-vibrant-red uppercase tracking-widest opacity-0 group-hover:opacity-100 transition-all hover:underline disabled:opacity-50"
                    >
                      {revoking === (session.id || session._id)
                        ? uiStrings.btn_terminating || "TERMINATING..."
                        : uiStrings.btn_terminate || "TERMINATE SESSION"}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="mt-auto glass-card p-6 rounded-2xl bg-vibrant-blue/5 border-vibrant-blue/10">
            <div className="flex gap-4">
              <div className="w-10 h-10 rounded-lg bg-vibrant-blue/10 flex items-center justify-center text-vibrant-blue">
                <ShieldCheckIcon className="w-6 h-6" />
              </div>
              <div className="flex-1">
                <h5 className="font-bold text-white">
                  {uiStrings.notes_title || "Session Hardening Active"}
                </h5>
                <p className="text-xs text-slate-400 mt-1 max-w-3xl">
                  {uiStrings.notes_desc ||
                    "Each session above is bound to a specific hardware signature (TLS Fingerprinting). Changing devices or network contexts may trigger a re-authentication challenge."}
                </p>
              </div>
              <div className="flex items-center px-4 py-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
                <span className="text-[10px] font-black text-emerald-400 uppercase tracking-widest">
                  {uiStrings.label_secure_primary || "Secure Session Primary"}
                </span>
              </div>
            </div>
          </div>
        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default ActiveSessions;
