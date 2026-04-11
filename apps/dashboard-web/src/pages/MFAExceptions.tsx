import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import type { RootState } from "../store";
import SecurityFooter from "../components/SecurityFooter";
import SecuritySidebar from "../components/SecuritySidebar";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";

const MFAExceptions: React.FC = () => {
  const navigate = useNavigate();
  const user = useSelector((state: RootState) => state.auth.user);
  const [exceptions, setExceptions] = useState<any[]>([]);
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const { uiStrings } = useUIStrings("mfa_exceptions");
  const [newException, setNewException] = useState({
    userId: "",
    reason: "",
    days: 7,
  });
  const [isSaving, setIsSaving] = useState(false);

  const orgId = user?.org_id || localStorage.getItem("org_id") || "";

  useEffect(() => {
    if (!orgId) return;
    const fetchData = async () => {
      try {
        const [exData, userData] = await Promise.all([
          api.getMFAExceptions(orgId),
          api.getUsers(orgId),
        ]);
        setExceptions(exData || []);
        setUsers(userData.users || []);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [orgId]);

  const handleDelete = async (id: string) => {
    if (
      !window.confirm(
        uiStrings?.confirm_revoke ||
          "Are you sure you want to revoke this security exception? This will immediately re-enforce MFA for the target user.",
      )
    )
      return;
    try {
      await api.deleteMFAException(id);
      setExceptions((prev) => prev.filter((e) => e.id !== id));
    } catch (err) {
      console.error(err);
    }
  };

  const handleAddException = async () => {
    if (!newException.userId || !newException.reason) return;
    setIsSaving(true);
    try {
      await api.addMFAException({
        userId: newException.userId,
        orgId: orgId,
        reason: newException.reason,
        days: newException.days,
      });
      const updated = await api.getMFAExceptions(orgId);
      setExceptions(updated || []);
      setShowModal(false);
      setNewException({ userId: "", reason: "", days: 7 });
    } catch (err) {
      console.error(err);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="flex h-screen w-full bg-background-dark text-white overflow-hidden">
      <SecuritySidebar />
      <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden relative">
        <div className="absolute top-0 left-0 w-[500px] h-[500px] bg-warning/5 rounded-full blur-[120px] -z-10"></div>

        <header className="h-16 flex items-center justify-between px-6 border-b border-white/5 bg-surface-dark/40 backdrop-blur-md flex-shrink-0 z-20">
          <div className="flex flex-col justify-center">
            <nav className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-widest mb-0.5 font-bold">
              <span
                onClick={() => navigate("/security")}
                className="hover:text-primary cursor-pointer transition-colors"
              >
                {uiStrings?.breadcrumb_parent || "Security Center"}
              </span>
              <span className="material-symbols-outlined text-[10px]">
                chevron_right
              </span>
              <span className="text-white">
                {uiStrings?.subtitle || "Exception registry"}
              </span>
            </nav>
            <h2 className="text-white text-lg font-bold tracking-tight">
              {uiStrings?.title || "MFA Bypass Protocols"}
            </h2>
          </div>
          <button
            onClick={() => setShowModal(true)}
            className="px-4 py-2 bg-warning/20 hover:bg-warning/30 text-warning text-xs font-bold rounded-lg flex items-center gap-2 transition-all border border-warning/30"
          >
            <span className="material-symbols-outlined text-[18px]">
              add_moderator
            </span>
            {uiStrings?.btn_grant || "Grant Temporary Exception"}
          </button>
        </header>

        <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
          {loading ? (
            <div className="flex flex-col items-center justify-center h-full gap-4">
              <div className="w-10 h-10 border-2 border-warning/20 border-t-warning rounded-full animate-spin"></div>
              <p className="text-xs text-text-muted font-mono tracking-widest">
                {uiStrings?.loading || "QUERYING EXCEPTION REGISTRY..."}
              </p>
            </div>
          ) : (
            <div className="max-w-6xl mx-auto space-y-6">
              <div className="bg-surface-dark/40 backdrop-blur-xl rounded-3xl border border-white/5 shadow-2xl overflow-hidden">
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse">
                    <thead>
                      <tr className="border-b border-white/5 text-[10px] text-text-muted uppercase tracking-[0.2em] bg-white/5 font-black">
                        <th className="px-8 py-4">
                          {uiStrings?.th_principal || "Security Principal"}
                        </th>
                        <th className="px-8 py-4">
                          {uiStrings?.th_justification || "Justification"}
                        </th>
                        <th className="px-8 py-4">
                          {uiStrings?.th_authorizer || "Authorizer"}
                        </th>
                        <th className="px-8 py-4">
                          {uiStrings?.th_limit || "Temporal Limit"}
                        </th>
                        <th className="px-8 py-4 text-right">
                          {uiStrings?.th_protocol || "Protocol"}
                        </th>
                      </tr>
                    </thead>
                    <tbody className="text-sm divide-y divide-white/5">
                      {exceptions.map((ex) => {
                        const targetUser = users.find(
                          (u) => u.id === ex.user_id,
                        );
                        const isExpired = new Date(ex.expires_at) < new Date();
                        return (
                          <tr
                            key={ex.id}
                            className="hover:bg-white/[0.02] transition-colors group"
                          >
                            <td className="px-8 py-5">
                              <div className="flex items-center gap-3">
                                <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold text-xs">
                                  {targetUser?.name?.charAt(0) || "?"}
                                </div>
                                <div>
                                  <div className="text-xs font-bold text-white uppercase tracking-wide">
                                    {targetUser?.name || "Unknown User"}
                                  </div>
                                  <div className="text-[10px] text-text-muted mt-0.5">
                                    {targetUser?.email || ex.user_id}
                                  </div>
                                </div>
                              </div>
                            </td>
                            <td className="px-8 py-5">
                              <p className="text-xs text-text-muted max-w-xs truncate group-hover:text-white transition-colors">
                                {ex.reason}
                              </p>
                            </td>
                            <td className="px-8 py-5">
                              <span className="text-[10px] font-mono text-text-muted">
                                {ex.granted_by}
                              </span>
                            </td>
                            <td className="px-8 py-5">
                              <div
                                className={`text-xs font-bold ${isExpired ? "text-danger" : "text-success"}`}
                              >
                                {new Date(ex.expires_at).toLocaleDateString()}
                                {isExpired && (
                                  <span className="ml-2 text-[9px] bg-danger/10 px-1.5 py-0.5 rounded uppercase">
                                    Expired
                                  </span>
                                )}
                              </div>
                            </td>
                            <td className="px-8 py-5 text-right">
                              <button
                                onClick={() => handleDelete(ex.id)}
                                className="text-[10px] uppercase font-black tracking-widest text-danger hover:text-white border border-danger/30 hover:bg-danger/80 px-3 py-1.5 rounded-lg transition-all"
                              >
                                {uiStrings?.btn_revoke || "Revoke"}
                              </button>
                            </td>
                          </tr>
                        );
                      })}
                      {exceptions.length === 0 && (
                        <tr>
                          <td colSpan={5} className="px-8 py-20 text-center">
                            <span className="material-symbols-outlined text-4xl text-white/5 mb-3 block">
                              verified_user
                            </span>
                            <p className="text-text-muted text-sm font-medium">
                              {uiStrings?.no_exceptions ||
                                "No active security exceptions found in the registry."}
                            </p>
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Grant Exception Modal */}
        {showModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-6 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-surface-dark border border-white/10 rounded-3xl w-full max-w-md shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
              <div className="px-8 py-6 border-b border-white/5 flex items-center justify-between bg-white/5">
                <h3 className="text-lg font-bold">
                  {uiStrings?.modal_title || "Grant Security Exception"}
                </h3>
                <button
                  onClick={() => setShowModal(false)}
                  className="text-text-muted hover:text-white"
                >
                  <span className="material-symbols-outlined">close</span>
                </button>
              </div>
              <div className="p-8 space-y-6">
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase tracking-widest text-text-muted">
                    {uiStrings?.label_target || "Target Security Principal"}
                  </label>
                  <select
                    className="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm focus:ring-1 focus:ring-warning/50 outline-none transition-all"
                    value={newException.userId}
                    onChange={(e) =>
                      setNewException((prev) => ({
                        ...prev,
                        userId: e.target.value,
                      }))
                    }
                  >
                    <option value="" className="bg-surface-dark">
                      {uiStrings?.opt_select_user || "Select User..."}
                    </option>
                    {users.map((u) => (
                      <option
                        key={u.id}
                        value={u.id}
                        className="bg-surface-dark text-black"
                      >
                        {u.name} ({u.email})
                      </option>
                    ))}
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase tracking-widest text-text-muted">
                    {uiStrings?.label_duration || "Temporal Duration (Days)"}
                  </label>
                  <input
                    type="number"
                    className="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm font-mono focus:ring-1 focus:ring-warning/50 outline-none transition-all"
                    value={newException.days}
                    onChange={(e) =>
                      setNewException((prev) => ({
                        ...prev,
                        days: parseInt(e.target.value),
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase tracking-widest text-text-muted">
                    {uiStrings?.label_justification ||
                      "Justification for Audit"}
                  </label>
                  <textarea
                    className="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm min-h-[100px] focus:ring-1 focus:ring-warning/50 outline-none transition-all resize-none"
                    placeholder={
                      uiStrings?.placeholder_justification ||
                      "Reason for temporary MFA bypass..."
                    }
                    value={newException.reason}
                    onChange={(e) =>
                      setNewException((prev) => ({
                        ...prev,
                        reason: e.target.value,
                      }))
                    }
                  ></textarea>
                </div>

                <button
                  onClick={handleAddException}
                  disabled={
                    isSaving || !newException.userId || !newException.reason
                  }
                  className="w-full py-3.5 bg-warning hover:bg-warning-dark text-black font-black uppercase tracking-widest text-xs rounded-xl transition-all shadow-xl shadow-warning/20 disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {isSaving ? (
                    <span className="material-symbols-outlined animate-spin">
                      sync
                    </span>
                  ) : (
                    uiStrings?.btn_authorize || "Authorize Exception"
                  )}
                </button>
              </div>
            </div>
          </div>
        )}

        <SecurityFooter />
      </main>
    </div>
  );
};

export default MFAExceptions;
