import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { useAuth } from "../../../context/AuthProvider";
import apiClient from "../../../services/apiClient";

interface APIKey {
  id: string;
  name: string;
  prefix: string;
  created_at: string;
  expires_at?: string;
  last_used_at?: string;
  is_revoked: boolean;
}

const APIKeys: React.FC = () => {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [newKeySecret, setNewKeySecret] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const { user } = useAuth();

  const fetchKeys = async () => {
    try {
      const data = await apiClient.get("/auth/api-keys");
      setKeys(data as unknown as APIKey[]);
    } catch (err) {
      console.error("Failed to fetch API keys:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchKeys();
  }, []);

  const handleCreateKey = async () => {
    if (!newKeyName) return;
    setIsCreating(true);
    try {
      const resp = await apiClient.post("/auth/api-keys", { name: newKeyName });
      setNewKeySecret((resp as any).api_key);
      fetchKeys();
    } catch (err) {
      console.error("Failed to create API key:", err);
    } finally {
      setIsCreating(false);
    }
  };

  const handleRevokeKey = async (id: string) => {
    if (
      !window.confirm(
        "Are you sure you want to revoke this API key? This action cannot be undone.",
      )
    )
      return;
    try {
      await apiClient.delete(`/auth/api-keys/${id}`);
      fetchKeys();
    } catch (err) {
      console.error("Failed to revoke API key:", err);
    }
  };

  return (
    <div className="p-8 max-w-7xl mx-auto">
      <header className="flex justify-between items-center mb-10">
        <div>
          <h1 className="text-3xl font-black text-white tracking-tight mb-2 bg-gradient-to-r from-white to-slate-500 bg-clip-text text-transparent">
            API Infrastructure
          </h1>
          <p className="text-slate-400 text-sm font-medium uppercase tracking-widest">
            Manage programmatic access for your organization
          </p>
        </div>
        <button
          onClick={() => {
            setNewKeySecret(null);
            setNewKeyName("");
            setShowCreateModal(true);
          }}
          className="bg-primary hover:bg-primary-dark text-white px-6 py-3 rounded-2xl font-bold text-xs uppercase tracking-widest transition-all shadow-lg shadow-primary/20 flex items-center gap-2 group"
        >
          <span className="material-symbols-outlined transition-transform group-hover:rotate-90">
            add
          </span>
          Issue New Key
        </button>
      </header>

      {loading ? (
        <div className="flex flex-col items-center justify-center h-64 gap-4">
          <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
          <span className="text-slate-500 text-xs font-bold uppercase tracking-[0.2em] animate-pulse">
            Initializing Secure Context...
          </span>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6">
          <AnimatePresence mode="popLayout">
            {keys.length === 0 ? (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="bg-slate-900/50 border border-white/5 rounded-3xl p-20 flex flex-col items-center text-center"
              >
                <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center mb-6">
                  <span className="material-symbols-outlined text-4xl text-primary">
                    vpn_key
                  </span>
                </div>
                <h3 className="text-xl font-bold text-white mb-2">
                  No Active Keys
                </h3>
                <p className="text-slate-400 max-w-sm mb-8">
                  Programmatic access allows your external services to interact
                  with FlyFleet securely.
                </p>
              </motion.div>
            ) : (
              keys.map((key) => (
                <motion.div
                  key={key.id}
                  layout
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  className={`group relative bg-slate-900/40 border border-white/5 hover:border-primary/30 rounded-3xl p-6 transition-all ${key.is_revoked ? "opacity-50 grayscale" : ""}`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-6">
                      <div
                        className={`w-14 h-14 rounded-2xl flex items-center justify-center transition-colors ${key.is_revoked ? "bg-slate-800" : "bg-primary/10 group-hover:bg-primary/20"}`}
                      >
                        <span
                          className={`material-symbols-outlined text-2xl ${key.is_revoked ? "text-slate-500" : "text-primary"}`}
                        >
                          {key.is_revoked ? "lock_reset" : "api"}
                        </span>
                      </div>
                      <div>
                        <h3 className="text-lg font-bold text-white mb-1 flex items-center gap-2">
                          {key.name}
                          {key.is_revoked && (
                            <span className="text-[10px] bg-red-500/10 text-red-500 border border-red-500/20 px-2 py-0.5 rounded-full uppercase tracking-tighter">
                              Revoked
                            </span>
                          )}
                        </h3>
                        <div className="flex items-center gap-4 text-xs font-medium text-slate-500">
                          <code className="bg-black/30 px-2 py-1 rounded-lg text-primary/80 tracking-wider">
                            {key.prefix}...
                          </code>
                          <span className="flex items-center gap-1">
                            <span className="material-symbols-outlined text-sm">
                              calendar_today
                            </span>
                            Created{" "}
                            {new Date(key.created_at).toLocaleDateString()}
                          </span>
                          {key.last_used_at && (
                            <span className="flex items-center gap-1">
                              <span className="material-symbols-outlined text-sm">
                                schedule
                              </span>
                              Last used{" "}
                              {new Date(key.last_used_at).toLocaleDateString()}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>

                    {!key.is_revoked && (
                      <button
                        onClick={() => handleRevokeKey(key.id)}
                        className="opacity-0 group-hover:opacity-100 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white px-4 py-2 rounded-xl text-[10px] font-bold uppercase tracking-widest transition-all"
                      >
                        Revoke Access
                      </button>
                    )}
                  </div>
                </motion.div>
              ))
            )}
          </AnimatePresence>
        </div>
      )}

      {/* Create Key Modal */}
      <AnimatePresence>
        {showCreateModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setShowCreateModal(false)}
              className="absolute inset-0 bg-black/80 backdrop-blur-xl"
            />
            <motion.div
              initial={{ opacity: 0, scale: 0.9, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: 20 }}
              className="relative w-full max-w-lg bg-slate-900 border border-white/10 rounded-[2.5rem] shadow-2xl overflow-hidden"
            >
              <div className="p-10">
                <header className="mb-8">
                  <h2 className="text-2xl font-black text-white mb-2 italic tracking-tight">
                    Issue Programmatic Key
                  </h2>
                  <p className="text-slate-400 text-sm">
                    This key will grant full access to your organization's
                    resources.
                  </p>
                </header>

                {!newKeySecret ? (
                  <div className="space-y-6">
                    <div className="space-y-2">
                      <label className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] ml-2">
                        Key Identifier
                      </label>
                      <input
                        autoFocus
                        type="text"
                        placeholder="e.g. CI/CD Pipeline, Python SDK"
                        className="w-full bg-slate-800/50 border border-white/5 rounded-2xl p-4 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                        value={newKeyName}
                        onChange={(e) => setNewKeyName(e.target.value)}
                      />
                    </div>
                    <button
                      onClick={handleCreateKey}
                      disabled={!newKeyName || isCreating}
                      className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-black py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 uppercase tracking-widest text-xs"
                    >
                      {isCreating
                        ? "Provisioning Storage..."
                        : "Generate New Key"}
                    </button>
                  </div>
                ) : (
                  <div className="space-y-6">
                    <div className="bg-orange-500/10 border border-orange-500/30 rounded-2xl p-6 mb-6">
                      <div className="flex items-center gap-3 text-orange-500 mb-2">
                        <span className="material-symbols-outlined text-xl">
                          warning
                        </span>
                        <span className="text-[10px] font-black uppercase tracking-widest">
                          Security Warning
                        </span>
                      </div>
                      <p className="text-orange-200/70 text-xs leading-relaxed">
                        This is the{" "}
                        <span className="text-white font-bold">only time</span>{" "}
                        the full secret will be displayed. Store it securely in
                        a vault immediately.
                      </p>
                    </div>

                    <div className="space-y-2">
                      <label className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] ml-2">
                        Secret Key
                      </label>
                      <div className="relative group">
                        <div className="w-full bg-black/40 border border-white/10 rounded-2xl p-4 font-mono text-primary break-all pr-12 text-sm select-all">
                          {newKeySecret}
                        </div>
                        <button
                          onClick={() =>
                            navigator.clipboard.writeText(newKeySecret)
                          }
                          className="absolute right-3 top-3 p-2 bg-slate-800 hover:bg-slate-700 rounded-lg text-slate-400 hover:text-white transition-all shadow-lg"
                        >
                          <span className="material-symbols-outlined text-sm">
                            content_copy
                          </span>
                        </button>
                      </div>
                    </div>

                    <button
                      onClick={() => setShowCreateModal(false)}
                      className="w-full bg-slate-800 hover:bg-slate-700 text-white font-bold py-4 rounded-2xl transition-all uppercase tracking-widest text-xs"
                    >
                      I Have Saved The Key
                    </button>
                  </div>
                )}
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default APIKeys;
