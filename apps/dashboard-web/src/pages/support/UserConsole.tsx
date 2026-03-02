import React, { useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

const UserConsole: React.FC = () => {
  const { uiStrings } = useUIStrings("support_v2");

  const [userId, setUserId] = useState("");
  const [userData, setUserData] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [noteText, setNoteText] = useState("");
  const [savingNote, setSavingNote] = useState(false);

  const fetchUserData = async () => {
    if (!userId) return;
    setLoading(true);
    setError(null);
    try {
      const [profile, security, sessions, logins, roles, permissions, notes] =
        await Promise.all([
          apiClient.get(`/api/v1/support/users/${userId}/profile`),
          apiClient.get(`/api/v1/support/users/${userId}/security`),
          apiClient.get(`/api/v1/support/users/${userId}/sessions`),
          apiClient.get(`/api/v1/support/users/${userId}/logins`),
          apiClient.get(`/api/v1/support/users/${userId}/roles`),
          apiClient.get(`/api/v1/support/users/${userId}/permissions`),
          apiClient.get(`/api/v1/support/users/${userId}/notes`),
        ]);

      setUserData({
        profile: profile.data,
        security: security.data,
        sessions: sessions.data.data,
        logins: logins.data.data,
        roles: roles.data.roles,
        permissions: permissions.data.permissions,
        notes: notes.data.data,
      });
    } catch (err) {
      setError(uiStrings?.error_load_user_data || "Failed to load user data");
    } finally {
      setLoading(false);
    }
  };

  const handleSaveNote = async () => {
    if (!noteText.trim() || !userId) return;
    setSavingNote(true);
    try {
      await apiClient.post(`/api/v1/support/users/${userId}/notes`, {
        note: noteText,
        agent: "Global Admin",
      });
      setNoteText("");
      await fetchUserData();
    } catch (err) {
      alert(uiStrings?.error_save_note || "Failed to save note");
    } finally {
      setSavingNote(false);
    }
  };

  const handleAction = async (action: string) => {
    try {
      await apiClient.post(`/api/v1/support/users/${userId}/${action}`);
      alert(
        uiStrings?.msg_action_success?.replace("{action}", action) ||
          `${action} executed successfully`,
      );
    } catch (err) {
      alert(
        uiStrings?.error_action_failed?.replace("{action}", action) ||
          `Failed to execute ${action}`,
      );
    }
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-[#001219] overflow-hidden h-full">
      <header className="h-[72px] bg-[#001219]/95 backdrop-blur-xl border-b border-white/5 flex items-center justify-between px-8 sticky top-0 z-50">
        <div className="flex items-center gap-6">
          <h1 className="text-[20px] font-bold text-white tracking-tight">
            {uiStrings?.title_user_console || "User Support Console"}
          </h1>
          <div className="h-4 w-px bg-white/10" />
          <div className="flex items-center gap-2 bg-primary/20 text-primary px-3 py-1 rounded-full border border-primary/20">
            <span className="material-symbols-outlined text-[14px]">
              settings_input_composite
            </span>
            <span className="text-[11px] font-bold uppercase tracking-wider">
              {uiStrings?.label_system_config || "System Configuration"}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 text-white/50 text-[12px] font-medium mr-4">
            <span className="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
            {uiStrings?.label_api_status || "API: Active"}
          </div>
          <div className="flex items-center gap-2 text-white/50 text-[12px] font-medium">
            <span className="material-symbols-outlined text-[16px]">dns</span>
            {uiStrings?.label_gateway_status || "Gateway: US-East-Primary"}
          </div>
        </div>
      </header>

      <main className="flex-1 overflow-y-auto p-8 custom-scrollbar">
        <div className="max-w-7xl mx-auto space-y-8 animate-fade-in relative z-10">
          <section className="bg-white/[0.03] backdrop-blur-3xl border border-white/5 rounded-3xl p-8 shadow-2xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full blur-[100px] pointer-events-none group-hover:bg-primary/10 transition-colors" />
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8 relative z-10">
              <div>
                <h2 className="text-[22px] font-bold text-white mb-2">
                  {uiStrings?.title_user_troubleshooting ||
                    "User Troubleshooting:"}
                </h2>
                <p className="text-white/40 text-[13px] font-medium max-w-lg">
                  {uiStrings?.subtitle_incident_detail ||
                    "Verified Customer Support"}
                </p>
              </div>
              <div className="flex items-center gap-3">
                <div className="relative">
                  <span className="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-white/30 text-[20px]">
                    search
                  </span>
                  <input
                    type="text"
                    value={userId}
                    onChange={(e) => setUserId(e.target.value)}
                    placeholder={
                      uiStrings?.placeholder_user_search ||
                      "Search by User ID, Email..."
                    }
                    className="h-12 w-80 bg-white/5 border border-white/10 rounded-2xl pl-12 pr-4 text-white text-[13px] outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/20"
                  />
                </div>
                <button
                  onClick={fetchUserData}
                  disabled={loading}
                  className="h-12 px-8 rounded-2xl bg-primary text-white font-bold text-[13px] hover:scale-[1.02] active:scale-[0.98] transition-all shadow-[0_8px_16px_rgba(19,127,236,0.2)]"
                >
                  {loading
                    ? uiStrings?.loading_user_data || "Loading..."
                    : uiStrings?.btn_load || "Load"}
                </button>
              </div>
            </div>

            {error && (
              <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-xl text-rose-500 text-sm mb-6">
                {error}
              </div>
            )}

            {userData && (
              <div className="space-y-12">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 pt-8 border-t border-white/5">
                  <div className="space-y-6">
                    <h3 className="text-white font-bold text-sm uppercase tracking-wider opacity-60 flex items-center gap-2">
                      <span className="material-symbols-outlined text-[18px]">
                        person
                      </span>
                      User Profile
                    </h3>
                    <div className="space-y-4">
                      <div>
                        <label className="text-white/30 text-[11px] font-bold uppercase tracking-wider block mb-1">
                          {uiStrings?.label_user_id || "User ID"}
                        </label>
                        <span className="text-white text-[13px] font-mono">
                          {userData.profile.userId}
                        </span>
                      </div>
                      <div>
                        <label className="text-white/30 text-[11px] font-bold uppercase tracking-wider block mb-1">
                          Name
                        </label>
                        <span className="text-white text-[15px] font-bold">
                          {userData.profile.name}
                        </span>
                      </div>
                      <div>
                        <label className="text-white/30 text-[11px] font-bold uppercase tracking-wider block mb-1">
                          Email
                        </label>
                        <span className="text-white text-[13px]">
                          {userData.profile.email}
                        </span>
                      </div>
                      <div>
                        <label className="text-white/30 text-[11px] font-bold uppercase tracking-wider block mb-1">
                          {uiStrings?.label_account_status || "Account Status"}
                        </label>
                        <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-500 text-[10px] font-bold border border-emerald-500/20 uppercase">
                          {userData.profile.status}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-6">
                    <h3 className="text-white font-bold text-sm uppercase tracking-wider opacity-60 flex items-center gap-2">
                      <span className="material-symbols-outlined text-[18px]">
                        security
                      </span>
                      Security & MFA
                    </h3>
                    <div className="space-y-4">
                      <div>
                        <label className="text-white/30 text-[11px] font-bold uppercase tracking-wider block mb-1">
                          {uiStrings?.label_mfa_status || "MFA Status"}
                        </label>
                        <span className="text-white text-[13px] flex items-center gap-2">
                          {userData.security.mfaEnabled ? (
                            <span className="text-emerald-500 font-bold">
                              {uiStrings?.status_enforced || "Enforced"}
                            </span>
                          ) : (
                            <span className="text-rose-500 font-bold">
                              {uiStrings?.status_disabled || "Disabled"}
                            </span>
                          )}
                        </span>
                      </div>
                      <div>
                        <label className="text-white/30 text-[11px] font-bold uppercase tracking-wider block mb-1">
                          Methods
                        </label>
                        <div className="flex flex-wrap gap-2">
                          {userData.security.mfaMethods.map((m: string) => (
                            <span
                              key={m}
                              className="px-2 py-1 rounded-lg bg-white/5 text-white/60 text-[10px] font-bold border border-white/5"
                            >
                              {m}
                            </span>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-6">
                    <h3 className="text-white font-bold text-sm uppercase tracking-wider opacity-60 flex items-center gap-2">
                      <span className="material-symbols-outlined text-[18px]">
                        bolt
                      </span>
                      {uiStrings?.title_admin_actions || "Admin Actions"}
                    </h3>
                    <div className="grid grid-cols-2 gap-3">
                      <button
                        onClick={() => handleAction("password-reset")}
                        className="h-10 px-4 rounded-xl border border-white/10 hover:border-white/20 text-white/70 hover:text-white font-bold text-[11px] uppercase tracking-wider transition-all bg-white/5"
                      >
                        {uiStrings?.btn_reset_pass || "Reset Pass"}
                      </button>
                      <button
                        onClick={() => handleAction("unlock")}
                        className="h-10 px-4 rounded-xl border border-white/10 hover:border-white/20 text-white/70 hover:text-white font-bold text-[11px] uppercase tracking-wider transition-all bg-white/5"
                      >
                        {uiStrings?.btn_unlock_acc || "Unlock Acc"}
                      </button>
                      <button
                        onClick={() => handleAction("mfa-reset")}
                        className="h-10 px-4 rounded-xl border border-white/10 hover:border-white/20 text-white/70 hover:text-white font-bold text-[11px] uppercase tracking-wider transition-all bg-white/5"
                      >
                        {uiStrings?.btn_reset_mfa || "Reset MFA"}
                      </button>
                      <button
                        onClick={() => handleAction("logout")}
                        className="h-10 px-4 rounded-xl border border-rose-500/20 hover:border-rose-500/40 text-rose-500 hover:text-rose-400 font-bold text-[11px] uppercase tracking-wider transition-all bg-rose-500/5"
                      >
                        {uiStrings?.btn_force_logout || "Force Logout"}
                      </button>
                    </div>
                  </div>
                </div>

                {/* Active Sessions */}
                <div className="space-y-6">
                  <h3 className="text-white font-bold text-sm uppercase tracking-wider opacity-60 flex items-center gap-2">
                    <span className="material-symbols-outlined text-[18px]">
                      devices
                    </span>
                    {uiStrings?.title_active_sessions || "Active Sessions"}
                  </h3>
                  <div className="overflow-hidden rounded-2xl border border-white/5 bg-white/[0.02]">
                    <div className="overflow-x-auto">
                      <table className="w-full text-left border-collapse">
                        <thead>
                          <tr className="border-b border-white/5 bg-white/5">
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider">
                              Device / OS
                            </th>
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider">
                              IP / Location
                            </th>
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider">
                              Status
                            </th>
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider text-right">
                              Actions
                            </th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-white/5">
                          {userData.sessions.map((session: any) => (
                            <tr
                              key={session.sessionId}
                              className="hover:bg-white/[0.02] transition-colors"
                            >
                              <td className="p-4">
                                <div className="text-white text-[13px] font-bold">
                                  {session.device}
                                </div>
                                <div className="text-white/30 text-[11px]">
                                  {session.browser}
                                </div>
                              </td>
                              <td className="p-4">
                                <div className="text-white text-[13px]">
                                  {session.ip}
                                </div>
                                <div className="text-white/30 text-[11px]">
                                  {session.location}
                                </div>
                              </td>
                              <td className="p-4">
                                {session.current ? (
                                  <span className="px-2 py-0.5 rounded bg-primary/20 text-primary text-[10px] font-bold border border-primary/20 uppercase">
                                    Current
                                  </span>
                                ) : (
                                  <span className="px-2 py-0.5 rounded bg-white/5 text-white/40 text-[10px] font-bold border border-white/10 uppercase">
                                    {uiStrings?.status_active || "Active"}
                                  </span>
                                )}
                              </td>
                              <td className="p-4 text-right">
                                {!session.current && (
                                  <button
                                    onClick={() =>
                                      handleAction(
                                        `sessions/${session.sessionId}/revoke`,
                                      )
                                    }
                                    className="text-rose-500 hover:text-rose-400 text-[11px] font-bold uppercase transition-colors"
                                  >
                                    {uiStrings?.btn_revoke || "Revoke"}
                                  </button>
                                )}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>

                {/* Login History */}
                <div className="space-y-6">
                  <h3 className="text-white font-bold text-sm uppercase tracking-wider opacity-60 flex items-center gap-2">
                    <span className="material-symbols-outlined text-[18px]">
                      history
                    </span>
                    {uiStrings?.title_login_history || "Recent Login History"}
                  </h3>
                  <div className="overflow-hidden rounded-2xl border border-white/5 bg-white/[0.02]">
                    <div className="overflow-x-auto">
                      <table className="w-full text-left border-collapse">
                        <thead>
                          <tr className="border-b border-white/5 bg-white/5">
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider">
                              {uiStrings?.th_time || "Time"}
                            </th>
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider">
                              {uiStrings?.th_event || "Event"}
                            </th>
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider">
                              {uiStrings?.th_ip_location || "IP / Location"}
                            </th>
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider">
                              {uiStrings?.th_risk || "Risk"}
                            </th>
                            <th className="p-4 text-[11px] font-bold text-white/40 uppercase tracking-wider text-right">
                              {uiStrings?.th_result || "Result"}
                            </th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-white/5">
                          {userData.logins.map((login: any, idx: number) => (
                            <tr
                              key={idx}
                              className="hover:bg-white/[0.02] transition-colors"
                            >
                              <td className="p-4 text-white text-[13px]">
                                {new Date(login.time).toLocaleString()}
                              </td>
                              <td className="p-4 text-white text-[13px] font-bold">
                                {login.eventType}
                              </td>
                              <td className="p-4">
                                <div className="text-white text-[13px]">
                                  {login.ip}
                                </div>
                                <div className="text-white/30 text-[11px]">
                                  {login.location}
                                </div>
                              </td>
                              <td className="p-4">
                                <span
                                  className={`font-bold text-[12px] ${login.riskScore > 50 ? "text-rose-500" : "text-emerald-500"}`}
                                >
                                  {login.riskScore}%
                                </span>
                              </td>
                              <td className="p-4 text-right">
                                <span
                                  className={`px-2 py-0.5 rounded text-[10px] font-bold border uppercase ${login.result === "SUCCESS" ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20" : "bg-rose-500/10 text-rose-500 border-rose-500/20"}`}
                                >
                                  {login.result}
                                </span>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>

                {/* Internal Notes */}
                <div className="space-y-6">
                  <h3 className="text-white font-bold text-sm uppercase tracking-wider opacity-60 flex items-center gap-2">
                    <span className="material-symbols-outlined text-[18px]">
                      description
                    </span>
                    {uiStrings?.title_internal_notes || "Internal Notes"}
                  </h3>
                  <div className="space-y-4">
                    {userData.notes.map((note: any, idx: number) => (
                      <div
                        key={idx}
                        className="p-4 rounded-2xl bg-white/5 border border-white/5"
                      >
                        <div className="flex justify-between items-center mb-2">
                          <span className="text-primary text-[11px] font-bold uppercase tracking-wider">
                            {note.agent}
                          </span>
                          <span className="text-white/20 text-[11px]">
                            {new Date(note.time).toLocaleString()}
                          </span>
                        </div>
                        <p className="text-white/80 text-[13px] leading-relaxed">
                          {note.note}
                        </p>
                      </div>
                    ))}
                    <div className="pt-4">
                      <textarea
                        placeholder={
                          uiStrings?.placeholder_internal_note ||
                          "Add internal note..."
                        }
                        value={noteText}
                        onChange={(e) => setNoteText(e.target.value)}
                        className="w-full h-24 bg-white/5 border border-white/10 rounded-2xl p-4 text-white text-[13px] outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/20 resize-none mb-3"
                      />
                      <div className="flex justify-end">
                        <button
                          onClick={handleSaveNote}
                          disabled={savingNote || !noteText.trim()}
                          className="h-10 px-6 rounded-xl bg-white/10 hover:bg-white/20 text-white font-bold text-[12px] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          {savingNote
                            ? uiStrings?.loading_saving || "Saving..."
                            : uiStrings?.btn_save_note || "Save Note"}
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </section>
        </div>
      </main>
    </div>
  );
};

export default UserConsole;
