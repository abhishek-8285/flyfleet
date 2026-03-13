import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import AuthButton from "../../components/auth/AuthButton";
import EnterpriseAuthLayout from "../../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";
import { TrustedDevice } from "../../types/auth";

const Devices: React.FC = () => {
  const [devices, setDevices] = useState<TrustedDevice[]>([]);
  const [loading, setLoading] = useState(true);
  const [revokingId, setRevokingId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const { uiStrings } = useUIStrings("auth_extended");

  const fetchDevices = async () => {
    setLoading(true);
    try {
      const resp = await api.getDevices();
      setDevices(resp.data || []);
    } catch (err: any) {
      setError(err.message || "Failed to load devices");
      // Fallback for demo/dev if API fails
      setDevices([
        {
          id: "1",
          name: "MacBook Pro 14",
          ip: "192.168.1.10",
          location: "Mumbai, India",
          lastUsed: "2 mins ago",
          isCurrent: true,
        },
        {
          id: "2",
          name: "iPhone 15 Pro",
          ip: "49.36.12.184",
          location: "Bangalore, India",
          lastUsed: "1 hour ago",
          isCurrent: false,
        },
        {
          id: "3",
          name: "iPad Air",
          ip: "10.0.0.5",
          location: "Delhi, India",
          lastUsed: "昨天",
          isCurrent: false,
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDevices();
  }, []);

  const handleRevoke = async (id: string) => {
    setRevokingId(id);
    try {
      await api.revokeDevice(id);
      setDevices((prev) => prev.filter((d) => d.id !== id));
    } catch (err: any) {
      setError(err.message || "Failed to revoke device");
    } finally {
      setRevokingId(null);
    }
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.devices_title || "Trusted Devices"}
      subtitle={
        uiStrings?.devices_sub || "Manage active instances of your identity"
      }
    >
      <div className="space-y-6">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2">
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </div>
        )}

        <div className="space-y-3 max-h-[460px] overflow-y-auto pr-1 custom-scrollbar">
          <AnimatePresence mode="popLayout">
            {loading ? (
              <div className="flex justify-center py-12">
                <div className="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
              </div>
            ) : devices.length === 0 ? (
              <div className="text-center py-12 bg-slate-800/20 rounded-3xl border border-white/5">
                <span className="material-symbols-outlined text-4xl text-slate-600 mb-2">
                  devices_off
                </span>
                <p className="text-slate-500 text-sm">
                  No other devices registered
                </p>
              </div>
            ) : (
              devices.map((device) => (
                <motion.div
                  key={device.id}
                  layout
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.9, x: -20 }}
                  className={`p-4 rounded-2xl border transition-all ${
                    device.isCurrent
                      ? "bg-primary/10 border-primary/30 ring-1 ring-primary/10"
                      : "bg-slate-800/30 border-white/5 hover:border-white/10"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div
                        className={`w-10 h-10 rounded-xl flex items-center justify-center ${device.isCurrent ? "bg-primary text-white" : "bg-slate-700/50 text-slate-400 border border-white/5"}`}
                      >
                        <span className="material-symbols-outlined">
                          {device.name.toLowerCase().includes("mac") ||
                          device.name.toLowerCase().includes("windows")
                            ? "desktop_windows"
                            : "smartphone"}
                        </span>
                      </div>
                      <div className="space-y-0.5">
                        <div className="flex items-center gap-2">
                          <h4 className="text-white text-sm font-bold">
                            {device.name}
                          </h4>
                          {device.isCurrent && (
                            <span className="text-[8px] font-bold text-primary uppercase tracking-widest bg-primary/20 px-1.5 py-0.5 rounded leading-none">
                              Current
                            </span>
                          )}
                        </div>
                        <p className="text-[10px] text-slate-500 font-medium tracking-tight">
                          {device.location} • {device.ip}
                        </p>
                        <p className="text-[10px] text-slate-600">
                          Last used: {device.lastUsed}
                        </p>
                      </div>
                    </div>

                    {!device.isCurrent && (
                      <button
                        onClick={() => handleRevoke(device.id)}
                        disabled={revokingId === device.id}
                        className="w-8 h-8 rounded-lg bg-red-400/10 text-red-500 hover:bg-red-500 hover:text-white transition-all border border-red-500/20 flex items-center justify-center disabled:opacity-50"
                      >
                        <span className="material-symbols-outlined text-sm">
                          {revokingId === device.id ? "sync" : "logout"}
                        </span>
                      </button>
                    )}
                  </div>
                </motion.div>
              ))
            )}
          </AnimatePresence>
        </div>

        <div className="pt-2">
          <AuthButton
            variant="secondary"
            onClick={() => window.location.reload()}
            disabled={loading}
          >
            Trust Current Device
          </AuthButton>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default Devices;
