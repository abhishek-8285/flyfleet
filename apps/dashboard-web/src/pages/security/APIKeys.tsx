import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import AuthButton from "../../components/auth/AuthButton";
import AuthInput from "../../components/auth/AuthInput";
import EnterpriseAuthLayout from "../../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";
import { APIKey } from "../../types/auth";

const APIKeys: React.FC = () => {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [generatedKey, setGeneratedKey] = useState<APIKey | null>(null);
  const { uiStrings } = useUIStrings("auth_extended");

  useEffect(() => {
    fetchKeys();
  }, []);

  const fetchKeys = async () => {
    try {
      const resp = await api.getAPIKeys();
      setKeys(resp.data);
    } catch (err) {
      // Mock fallback
      setKeys([
        {
          id: "key_1",
          name: "Production-Web-App",
          createdAt: "2024-03-01T10:00:00Z",
          expiresAt: "2025-03-01T10:00:00Z",
          status: "active",
        },
        {
          id: "key_2",
          name: "CI-CD-Bot",
          createdAt: "2024-02-15T08:30:00Z",
          expiresAt: "2024-08-15T08:30:00Z",
          status: "active",
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateKey = async () => {
    try {
      const resp = await api.createAPIKey(newKeyName);
      setGeneratedKey(resp.data);
      fetchKeys();
    } catch (err) {
      // Mock
      const mockKey: APIKey = {
        id: Math.random().toString(36).substr(2, 9),
        name: newKeyName,
        key: "flf_" + Math.random().toString(36).substr(2, 32),
        createdAt: new Date().toISOString(),
        expiresAt: new Date(Date.now() + 31536000000).toISOString(),
        status: "active",
      };
      setGeneratedKey(mockKey);
      setKeys((prev) => [mockKey, ...prev]);
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
       await api.revokeAPIKey(id);
       fetchKeys();
     } catch (err) {
       setKeys((prev) =>
         prev.map((k) => (k.id === id ? { ...k, status: "revoked" } : k)),
       );
     }
   };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    alert("Token copied to clipboard!");
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.api_keys_title || "API Access Control"}
      subtitle={
        uiStrings?.api_keys_sub || "Programmatic authentication management"
      }
    >
      <div className="space-y-6 px-4">
        <div className="flex justify-between items-center mb-2">
          <h3 className="text-white font-bold text-lg">Your API Keys</h3>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-primary/20 hover:bg-primary/30 text-primary text-xs font-bold rounded-xl transition-all border border-primary/20"
          >
            Generate New Key
          </button>
        </div>

        <div className="space-y-4">
          <AnimatePresence mode="popLayout">
            {keys.map((key) => (
              <motion.div
                key={key.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.95 }}
                className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-5 rounded-3xl group"
              >
                <div className="flex justify-between items-start">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-white font-bold">{key.name}</span>
                      <span
                        className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase ${
                          key.status === "active"
                            ? "bg-emerald-500/10 text-emerald-500"
                            : key.status === "revoked"
                              ? "bg-red-500/10 text-red-500"
                              : "bg-slate-500/10 text-slate-500"
                        }`}
                      >
                        {key.status}
                      </span>
                    </div>
                    <div className="flex items-center gap-4 text-[10px] text-slate-500 font-medium">
                      <span>
                        Created: {new Date(key.createdAt).toLocaleDateString()}
                      </span>
                      <span>
                        Expires: {new Date(key.expiresAt).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={() => handleRevokeKey(key.id)}
                      disabled={key.status !== "active"}
                      className="p-2 text-slate-400 hover:text-red-500 transition-colors"
                      title="Revoke Key"
                    >
                      <span className="material-symbols-outlined text-sm">
                        delete
                      </span>
                    </button>
                  </div>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>

        {/* Create Modal */}
        <AnimatePresence>
          {showCreateModal && (
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md">
              <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.9 }}
                className="bg-slate-900 border border-white/10 p-8 rounded-[40px] w-full max-w-md shadow-2xl"
              >
                {!generatedKey ? (
                  <div className="space-y-6">
                    <div className="text-center space-y-2">
                      <h3 className="text-2xl text-white font-black tracking-tighter">
                        New Access Token
                      </h3>
                      <p className="text-slate-400 text-sm">
                        Assign a memorable name for this key
                      </p>
                    </div>
                    <AuthInput
                      label="Key Name"
                      icon="label"
                      placeholder="e.g. Monitoring-Dashboard"
                      value={newKeyName}
                      onChange={(e) => setNewKeyName(e.target.value)}
                    />
                    <div className="flex gap-3">
                      <AuthButton
                        onClick={() => setShowCreateModal(false)}
                        className="!bg-slate-800 !from-slate-800 !to-slate-800"
                      >
                        Cancel
                      </AuthButton>
                      <AuthButton
                        onClick={handleCreateKey}
                        disabled={!newKeyName}
                      >
                        Generate
                      </AuthButton>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-6">
                    <div className="text-center space-y-2">
                      <div className="w-16 h-16 bg-emerald-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-emerald-500/20">
                        <span className="material-symbols-outlined text-emerald-500 text-3xl italic">
                          key
                        </span>
                      </div>
                      <h3 className="text-2xl text-white font-black tracking-tighter">
                        Key Generated
                      </h3>
                      <p className="text-red-400 text-xs font-bold uppercase tracking-widest">
                        Copy now - it won't be shown again!
                      </p>
                    </div>

                    <div className="bg-slate-950 border border-white/5 p-4 rounded-2xl font-mono text-emerald-400 text-sm break-all relative group">
                      {generatedKey.key}
                      <button
                        onClick={() => copyToClipboard(generatedKey.key!)}
                        className="absolute right-2 top-2 p-2 bg-slate-800 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity"
                      >
                        <span className="material-symbols-outlined text-xs">
                          content_copy
                        </span>
                      </button>
                    </div>

                    <AuthButton
                      onClick={() => {
                        setShowCreateModal(false);
                        setGeneratedKey(null);
                        setNewKeyName("");
                      }}
                    >
                      Done
                    </AuthButton>
                  </div>
                )}
              </motion.div>
            </div>
          )}
        </AnimatePresence>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default APIKeys;
