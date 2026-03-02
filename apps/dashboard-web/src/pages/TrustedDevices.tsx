import { AnimatePresence, motion } from "framer-motion";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";

interface Device {
  id: string;
  name: string;
  location: string;
  lastActive: string;
  isCurrent?: boolean;
  type: "desktop" | "mobile" | "tablet";
}

const TrustedDevices: React.FC = () => {
  const [devices, setDevices] = useState<Device[]>([]);

  React.useEffect(() => {
    fetch("/api/v1/ui/misc/trusted_devices")
      .then((res) => res.json())
      .then((data) => setDevices(Array.isArray(data) ? data : []))
      .catch((err) => console.error(err));
  }, []);
  const [removingId, setRemovingId] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleRemove = (id: string) => {
    setRemovingId(id);
    setTimeout(() => {
      setDevices((prev) => prev.filter((d) => d.id !== id));
      setRemovingId(null);
    }, 500);
  };

  return (
    <EnterpriseAuthLayout
      title="Secure Devices"
      subtitle="Manage your trusted enterprise access"
    >
      <div className="space-y-6">
        <div className="space-y-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
          <AnimatePresence>
            {devices.map((device) => (
              <motion.div
                key={device.id}
                layout
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, x: 50 }}
                className={`group p-4 rounded-2xl border transition-all ${
                  device.isCurrent
                    ? "bg-primary/10 border-primary/50 ring-1 ring-primary/20"
                    : "bg-slate-800/20 border-white/5 hover:border-white/10"
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div
                      className={`h-10 w-10 rounded-xl flex items-center justify-center ${
                        device.isCurrent
                          ? "bg-primary text-white"
                          : "bg-slate-700/50 text-slate-400"
                      }`}
                    >
                      <span className="material-symbols-outlined">
                        {device.type === "desktop"
                          ? "desktop_windows"
                          : device.type === "mobile"
                            ? "smartphone"
                            : "tablet_mac"}
                      </span>
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="text-sm font-bold text-white tracking-tight">
                          {device.name}
                        </h3>
                        {device.isCurrent && (
                          <span className="text-[8px] font-bold text-primary uppercase tracking-[0.2em] bg-primary/20 px-1.5 py-0.5 rounded">
                            Active
                          </span>
                        )}
                      </div>
                      <p className="text-[10px] font-medium text-slate-500 mt-0.5">
                        {device.location} • {device.lastActive}
                      </p>
                    </div>
                  </div>

                  {!device.isCurrent && (
                    <button
                      onClick={() => handleRemove(device.id)}
                      disabled={removingId === device.id}
                      className="h-8 w-8 rounded-lg bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-white transition-all flex items-center justify-center group/btn"
                    >
                      <span className="material-symbols-outlined text-sm">
                        {removingId === device.id ? "sync" : "delete"}
                      </span>
                    </button>
                  )}
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>

        <div className="pt-2">
          <button
            onClick={() => navigate("/dashboard")}
            className="w-full bg-slate-800/50 hover:bg-slate-700 text-white font-bold py-4 rounded-2xl transition-all border border-white/5 flex items-center justify-center gap-2 text-sm uppercase tracking-widest"
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default TrustedDevices;
