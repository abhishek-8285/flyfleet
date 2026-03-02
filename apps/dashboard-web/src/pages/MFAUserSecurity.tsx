import React, { useCallback, useEffect, useState } from "react";
import { useAuth } from "../context/AuthProvider";
import { useUIStrings } from "../hooks/useUIStrings";
import { api } from "../services/api";

interface LockedUser {
  id: string;
  email: string;
  name: string;
  locked_until?: string;
  failed_logins: number;
}

interface MFAException {
  id: string;
  user_id: string;
  email?: string;
  name?: string;
  reason: string;
  start_date: string;
  expires_at?: string;
}

const MFAUserSecurity: React.FC = () => {
  const { user } = useAuth();
  const [lockedUsers, setLockedUsers] = useState<LockedUser[]>([]);
  const [exceptions, setExceptions] = useState<MFAException[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("mfa_security");
  const [searchQuery, setSearchQuery] = useState("");
  const [showAddException, setShowAddException] = useState(false);
  const [newException, setNewException] = useState({
    userId: "",
    reason: "",
    days: 7,
  });

  const fetchData = useCallback(async () => {
    if (!user?.org_id) return;
    setLoading(true);
    try {
      const [locked, excs] = await Promise.all([
        api.getLockedUsers(user.org_id),
        api.getMFAExceptions(user.org_id),
      ]);
      setLockedUsers(locked || []);
      setExceptions(excs || []);
    } catch (err) {
      console.error("Failed to fetch user security data", err);
    } finally {
      setLoading(false);
    }
  }, [user?.org_id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleUnlock = async (userId: string) => {
    try {
      await api.unlockUser(userId);
      fetchData();
    } catch (err) {
      alert("Failed to unlock user");
    }
  };

  const handleRemoveException = async (id: string) => {
    if (!window.confirm("Are you sure you want to revoke this MFA exception?"))
      return;
    try {
      await api.deleteMFAException(id);
      fetchData();
    } catch (err) {
      alert("Failed to revoke exception");
    }
  };

  const handleAddException = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user?.org_id) return;
    try {
      await api.addMFAException({
        ...newException,
        orgId: user.org_id,
      });
      setShowAddException(false);
      setNewException({ userId: "", reason: "", days: 7 });
      fetchData();
    } catch (err) {
      alert("Failed to add exception");
    }
  };

  const filteredLocked = lockedUsers.filter(
    (u) =>
      u.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      u.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const filteredExceptions = exceptions.filter(
    (e) =>
      e.email?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      e.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      e.reason.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  return (
    <div className="flex flex-col gap-6 animate-in fade-in duration-500">
      {/* Search & Actions */}
      <div className="flex flex-wrap gap-4 items-center justify-between bg-surface-dark p-4 rounded-xl border border-border-color shadow-lg">
        <div className="relative flex-1 min-w-[300px]">
          <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-text-muted">
            <span className="material-symbols-outlined text-[20px]">
              search
            </span>
          </span>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-surface-darker border border-border-color text-sm text-white rounded-lg pl-10 pr-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all"
            placeholder={
              uiStrings?.search_placeholder ||
              "Search users by name, email, or exception reason..."
            }
          />
        </div>
        <button
          onClick={() => setShowAddException(true)}
          className="px-4 py-2.5 bg-primary text-white text-sm font-semibold rounded-lg hover:bg-blue-600 transition-all flex items-center gap-2 shadow-lg shadow-primary/20"
        >
          <span className="material-symbols-outlined text-[20px]">
            person_add
          </span>
          {uiStrings?.btn_grant_exception || "Grant Exception"}
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Locked Accounts Table */}
        <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col shadow-xl overflow-hidden">
          <div className="p-4 border-b border-border-color bg-surface-darker/50 flex justify-between items-center">
            <h3 className="text-white font-bold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-orange-500 text-[20px]">
                lock_clock
              </span>
              {uiStrings?.title_locked || "Locked Accounts"}
              <span className="bg-orange-500/20 text-orange-500 text-[10px] px-2 py-0.5 rounded-full ml-1">
                {filteredLocked.length}
              </span>
            </h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-border-color text-[10px] text-text-muted uppercase tracking-widest bg-surface-darker/30">
                  <th className="px-5 py-4 font-bold">
                    {uiStrings?.th_user || "User"}
                  </th>
                  <th className="px-5 py-4 font-bold">
                    {uiStrings?.th_attempts || "Attempts"}
                  </th>
                  <th className="px-5 py-4 font-bold">
                    {uiStrings?.th_locked_until || "Locked Until"}
                  </th>
                  <th className="px-5 py-4 font-bold text-right">
                    {uiStrings?.th_actions || "Actions"}
                  </th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {loading ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-5 py-8 text-center text-text-muted"
                    >
                      Loading...
                    </td>
                  </tr>
                ) : filteredLocked.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-5 py-8 text-center text-text-muted"
                    >
                      No locked accounts found
                    </td>
                  </tr>
                ) : (
                  filteredLocked.map((u) => (
                    <tr
                      key={u.id}
                      className="border-b border-border-color hover:bg-white/5 transition-colors"
                    >
                      <td className="px-5 py-4">
                        <div className="flex flex-col">
                          <span className="text-white font-medium">
                            {u.name}
                          </span>
                          <span className="text-xs text-text-muted">
                            {u.email}
                          </span>
                        </div>
                      </td>
                      <td className="px-5 py-4 text-white font-mono">
                        {u.failed_logins}
                      </td>
                      <td className="px-5 py-4 text-text-muted text-xs">
                        {u.locked_until
                          ? new Date(u.locked_until).toLocaleString()
                          : "Indefinite"}
                      </td>
                      <td className="px-5 py-4 text-right">
                        <button
                          onClick={() => handleUnlock(u.id)}
                          className="text-success hover:underline text-xs font-bold"
                        >
                          {uiStrings?.btn_unlock || "Unlock"}
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* MFA Exceptions Table */}
        <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col shadow-xl overflow-hidden">
          <div className="p-4 border-b border-border-color bg-surface-darker/50 flex justify-between items-center">
            <h3 className="text-white font-bold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-primary text-[20px]">
                admin_panel_settings
              </span>
              {uiStrings?.title_exceptions || "Active MFA Exceptions"}
              <span className="bg-primary/20 text-primary text-[10px] px-2 py-0.5 rounded-full ml-1">
                {filteredExceptions.length}
              </span>
            </h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-border-color text-[10px] text-text-muted uppercase tracking-widest bg-surface-darker/30">
                  <th className="px-5 py-4 font-bold">
                    {uiStrings?.th_user || "User"}
                  </th>
                  <th className="px-5 py-4 font-bold">
                    {uiStrings?.th_reason || "Reason"}
                  </th>
                  <th className="px-5 py-4 font-bold">
                    {uiStrings?.th_expiry || "Expiry"}
                  </th>
                  <th className="px-5 py-4 font-bold text-right">
                    {uiStrings?.th_actions || "Actions"}
                  </th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {loading ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-5 py-8 text-center text-text-muted"
                    >
                      Loading...
                    </td>
                  </tr>
                ) : filteredExceptions.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-5 py-8 text-center text-text-muted"
                    >
                      No active exceptions
                    </td>
                  </tr>
                ) : (
                  filteredExceptions.map((e) => (
                    <tr
                      key={e.id}
                      className="border-b border-border-color hover:bg-white/5 transition-colors"
                    >
                      <td className="px-5 py-4">
                        <div className="flex flex-col">
                          <span className="text-white font-medium">
                            {e.name || "User"}
                          </span>
                          <span className="text-xs text-text-muted">
                            {e.email || e.user_id}
                          </span>
                        </div>
                      </td>
                      <td className="px-5 py-4 text-text-muted text-xs italic">
                        "{e.reason}"
                      </td>
                      <td className="px-5 py-4">
                        <span
                          className={`text-[10px] font-bold px-2 py-1 rounded ${
                            e.expires_at && new Date(e.expires_at) < new Date()
                              ? "bg-danger/10 text-danger"
                              : "bg-primary/10 text-primary"
                          }`}
                        >
                          {e.expires_at
                            ? new Date(e.expires_at).toLocaleDateString()
                            : "Permanent"}
                        </span>
                      </td>
                      <td className="px-5 py-4 text-right">
                        <button
                          onClick={() => handleRemoveException(e.id)}
                          className="text-danger hover:underline text-xs font-bold"
                        >
                          {uiStrings?.btn_revoke || "Revoke"}
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Add Exception Modal */}
      {showAddException && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-surface-dark border border-border-color w-full max-w-md rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
            <div className="p-6 border-b border-border-color bg-surface-darker/50 flex justify-between items-center">
              <h3 className="text-white font-bold flex items-center gap-2">
                <span className="material-symbols-outlined text-primary">
                  person_add
                </span>
                {uiStrings?.modal_title || "Grant MFA Exception"}
              </h3>
              <button
                onClick={() => setShowAddException(false)}
                className="text-text-muted hover:text-white transition-colors"
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>
            <form onSubmit={handleAddException} className="p-6 space-y-4">
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-wider">
                  {uiStrings?.label_user_id || "User ID / Email"}
                </label>
                <input
                  type="text"
                  required
                  value={newException.userId}
                  onChange={(e) =>
                    setNewException({ ...newException, userId: e.target.value })
                  }
                  className="w-full bg-surface-darker border border-border-color rounded-lg px-4 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary outline-none transition-all"
                  placeholder={
                    uiStrings?.placeholder_userId ||
                    "Paste User UUID or email..."
                  }
                />
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-wider">
                  {uiStrings?.label_reason || "Reason for Exception"}
                </label>
                <textarea
                  required
                  value={newException.reason}
                  onChange={(e) =>
                    setNewException({ ...newException, reason: e.target.value })
                  }
                  className="w-full bg-surface-darker border border-border-color rounded-lg px-4 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary outline-none transition-all h-24 resize-none"
                  placeholder={
                    uiStrings?.placeholder_reason ||
                    "e.g. Broken hardware key, traveling to restricted area..."
                  }
                />
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-wider">
                  {uiStrings?.label_duration || "Duration (Days)"}
                </label>
                <select
                  value={newException.days}
                  onChange={(e) =>
                    setNewException({
                      ...newException,
                      days: Number(e.target.value),
                    })
                  }
                  className="w-full bg-surface-darker border border-border-color rounded-lg px-4 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary outline-none transition-all"
                >
                  <option value={1}>{uiStrings?.opt_1d || "1 Day"}</option>
                  <option value={3}>{uiStrings?.opt_3d || "3 Days"}</option>
                  <option value={7}>{uiStrings?.opt_7d || "7 Days"}</option>
                  <option value={30}>{uiStrings?.opt_30d || "30 Days"}</option>
                  <option value={0}>
                    {uiStrings?.opt_permanent || "Permanent"}
                  </option>
                </select>
              </div>
              <div className="pt-4 flex gap-3">
                <button
                  type="button"
                  onClick={() => setShowAddException(false)}
                  className="flex-1 px-4 py-2.5 rounded-lg border border-border-color text-sm font-semibold text-white hover:bg-white/5 transition-all"
                >
                  {uiStrings?.btn_cancel || "Cancel"}
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2.5 rounded-lg bg-primary text-sm font-semibold text-white hover:bg-blue-600 transition-all shadow-lg shadow-primary/20"
                >
                  {uiStrings?.btn_grant || "Grant Exemption"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default MFAUserSecurity;
