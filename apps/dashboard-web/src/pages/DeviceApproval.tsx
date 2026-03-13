import { AnimatePresence, motion } from "framer-motion";
import React, { useState } from "react";
import AuthButton from "../components/auth/AuthButton";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";

const DeviceApproval: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState<"approved" | "denied" | null>(null);
  const { uiStrings } = useUIStrings("auth_extended");

  // Mock device data - in a real app, this would come from a secure temporary token
  const deviceData = {
    fingerprint: "DF-8821-X99",
    location: "Mumbai, India",
    ip: "49.36.12.184",
    browser: "Chrome 122.0.0 (MacOS)",
    timestamp: new Date().toISOString(),
  };

  const handleAction = async (approve: boolean) => {
    setLoading(true);
    setError("");

    try {
      await api.approveDevice(deviceData.fingerprint, approve);
      setSuccess(approve ? "approved" : "denied");
    } catch (err: any) {
      setError(err.message || "Action failed");
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <EnterpriseAuthLayout
        title={success === "approved" ? "Device Approved" : "Access Denied"}
        subtitle="Device security protocol updated"
      >
        <div className="text-center space-y-6">
          <div
            className={`w-20 h-20 ${success === "approved" ? "bg-green-500/20 text-green-500 border-green-500/30" : "bg-red-500/20 text-red-500 border-red-500/30"} rounded-full flex items-center justify-center mx-auto mb-4 border`}
          >
            <span className="material-symbols-outlined text-4xl">
              {success === "approved" ? "verified_user" : "block"}
            </span>
          </div>
          <p className="text-slate-300 text-sm leading-relaxed">
            {success === "approved"
              ? "This device has been added to your trusted vault. You can now proceed with the login."
              : "Access from this device has been permanently blocked. If this was a mistake, contact your IT administrator."}
          </p>
          <AuthButton
            variant="secondary"
            onClick={() => (window.location.href = "/login")}
          >
            Return to Login
          </AuthButton>
        </div>
      </EnterpriseAuthLayout>
    );
  }

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.device_approval_title || "Security Alert"}
      subtitle={uiStrings?.device_approval_sub || "Unknown device detection"}
    >
      <AnimatePresence mode="wait">
        {error && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </motion.div>
        )}
      </AnimatePresence>
      <div className="space-y-6">
        <div className="bg-slate-800/30 border border-white/5 rounded-2xl p-5 space-y-4">
          <div className="flex items-start gap-4">
            <div className="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center border border-primary/20 shrink-0">
              <span className="material-symbols-outlined text-primary text-xl">
                devices
              </span>
            </div>
            <div className="space-y-1">
              <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                Device Fingerprint
              </p>
              <p className="text-white font-mono text-sm">
                {deviceData.fingerprint}
              </p>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                Location
              </p>
              <div className="flex items-center gap-1.5 text-slate-200">
                <span className="material-symbols-outlined text-xs">
                  location_on
                </span>
                <span className="text-xs font-semibold">
                  {deviceData.location}
                </span>
              </div>
            </div>
            <div className="space-y-1">
              <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                IP Address
              </p>
              <div className="flex items-center gap-1.5 text-slate-200">
                <span className="material-symbols-outlined text-xs">
                  public
                </span>
                <span className="text-xs font-semibold">{deviceData.ip}</span>
              </div>
            </div>
          </div>

          <div className="space-y-1">
            <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
              Browser Agent
            </p>
            <div className="flex items-center gap-1.5 text-slate-200">
              <span className="material-symbols-outlined text-xs">
                language
              </span>
              <span className="text-xs font-semibold">
                {deviceData.browser}
              </span>
            </div>
          </div>
        </div>

        <div className="flex gap-3">
          <AuthButton
            variant="secondary"
            className="border-red-500/20 text-red-500 hover:bg-red-500/10"
            disabled={loading}
            onClick={() => handleAction(false)}
          >
            Deny
          </AuthButton>
          <AuthButton
            loading={loading}
            loadingText="Approving..."
            onClick={() => handleAction(true)}
          >
            Approve
          </AuthButton>
        </div>

        <p className="text-center text-[10px] text-slate-500 font-medium leading-relaxed uppercase tracking-wider px-4">
          Approving this device will bypass secondary verification for 30 days
          from this IP range.
        </p>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default DeviceApproval;
