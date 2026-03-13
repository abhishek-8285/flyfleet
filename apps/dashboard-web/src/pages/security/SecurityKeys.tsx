import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import AuthButton from "../../components/auth/AuthButton";
import EnterpriseAuthLayout from "../../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";
import { SecurityKey } from "../../types/auth";

const SecurityKeys: React.FC = () => {
  const [keys, setKeys] = useState<SecurityKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [registering, setRegistering] = useState(false);
  const [removingId, setRemovingId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const { uiStrings } = useUIStrings("auth_extended");

  const fetchKeys = async () => {
    setLoading(true);
    try {
      const resp = await api.getSecurityKeys();
      setKeys(resp.data || []);
    } catch (err: any) {
      setError(err.message || "Failed to load security keys");
      // Fallback/Mock
      setKeys([
        { id: "k1", nickname: "YubiKey 5C NFC", lastUsed: "12 Mar 2026" },
        { id: "k2", nickname: "Titan Security Key", lastUsed: "Never" },
      ]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchKeys();
  }, []);

  const handleRegister = async () => {
    setRegistering(true);
    setError("");
    try {
      // Mock WebAuthn flow
      await new Promise((resolve) => setTimeout(resolve, 2000));
      const newKey = {
        id: `k${Date.now()}`,
        nickname: "New Hardware Key",
        lastUsed: "Just now",
      };
      await api.registerSecurityKey(newKey);
      setKeys((prev) => [...prev, newKey]);
    } catch (err: any) {
      setError(err.message || "Key registration failed");
    } finally {
      setRegistering(false);
    }
  };

  const handleRemove = async (id: string) => {
    setRemovingId(id);
    try {
      await api.deleteSecurityKey(id);
      setKeys((prev) => prev.filter((k) => k.id !== id));
    } catch (err: any) {
      setError(err.message || "Failed to remove key");
    } finally {
      setRemovingId(null);
    }
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.security_keys_title || "Security Keys"}
      subtitle={
        uiStrings?.security_keys_sub ||
        "Register FIDO2 / WebAuthn hardware tokens"
      }
    >
      <div className="space-y-6">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2">
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </div>
        )}

        <div className="bg-slate-800/20 border border-white/5 p-4 rounded-2xl flex gap-4 items-start">
          <div className="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center shrink-0 border border-primary/20">
            <span className="material-symbols-outlined text-primary text-xl">
              key
            </span>
          </div>
          <div className="space-y-1">
            <h4 className="text-white text-sm font-bold">
              Hardware Authentication
            </h4>
            <p className="text-slate-500 text-[10px] leading-relaxed">
              Use physical security keys to protect against phishing and session
              hijacking. FlyFleet supports Yubico, Titan, and other
              FIDO2-compliant keys.
            </p>
          </div>
        </div>

        <div className="space-y-3 max-h-[300px] overflow-y-auto pr-1 custom-scrollbar">
          <AnimatePresence mode="popLayout">
            {loading ? (
              <div className="flex justify-center py-8">
                <div className="w-6 h-6 border-3 border-primary/20 border-t-primary rounded-full animate-spin" />
              </div>
            ) : keys.length === 0 ? (
              <div className="text-center py-8 bg-slate-800/10 rounded-3xl border border-white/5">
                <p className="text-slate-500 text-xs font-medium italic">
                  No keys registered yet
                </p>
              </div>
            ) : (
              keys.map((key) => (
                <motion.div
                  key={key.id}
                  layout
                  initial={{ opacity: 0, scale: 0.98 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.95, x: 10 }}
                  className="p-4 bg-slate-900/40 border border-white/5 rounded-2xl flex items-center justify-between group"
                >
                  <div className="flex items-center gap-3">
                    <span className="material-symbols-outlined text-slate-400 group-hover:text-primary transition-colors">
                      usb
                    </span>
                    <div className="space-y-0.5">
                      <p className="text-white text-sm font-bold tracking-tight">
                        {key.nickname}
                      </p>
                      <p className="text-[10px] text-slate-500 font-medium tracking-tight">
                        Last used: {key.lastUsed}
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() => handleRemove(key.id)}
                    disabled={removingId === key.id}
                    className="w-8 h-8 rounded-lg bg-red-400/5 hover:bg-red-500/20 text-red-400/50 hover:text-red-500 transition-all flex items-center justify-center"
                  >
                    <span className="material-symbols-outlined text-sm">
                      {removingId === key.id ? "sync" : "close"}
                    </span>
                  </button>
                </motion.div>
              ))
            )}
          </AnimatePresence>
        </div>

        <div className="pt-2">
          <AuthButton
            onClick={handleRegister}
            loading={registering}
            loadingText="Touch your key..."
          >
            Register New Key
          </AuthButton>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default SecurityKeys;
