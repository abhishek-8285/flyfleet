import { motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import AuthButton from "../components/auth/AuthButton";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";
import { AccountLockInfo } from "../types/auth";

const AccountLocked: React.FC = () => {
  const [lockInfo, setLockInfo] = useState<AccountLockInfo | null>(null);
  const [timeLeft, setTimeLeft] = useState(0);
  const { uiStrings } = useUIStrings("auth_extended");

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const resp = await api.getLockStatus();
        setLockInfo(resp.data);
        setTimeLeft(resp.data.remainingSeconds);
      } catch (err) {
        // Mock fallback
        setLockInfo({
          isLocked: true,
          reason: "Too many failed login attempts",
          remainingSeconds: 1800, // 30 mins
        });
        setTimeLeft(1800);
      } finally {
        // loading state removed
      }
    };
    fetchStatus();
  }, []);

  useEffect(() => {
    if (timeLeft <= 0) return;
    const timer = setInterval(() => {
      setTimeLeft((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [timeLeft]);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.locked_title || "Access Restricted"}
      subtitle={uiStrings?.locked_sub || "Security protocol engaged"}
    >
      <div className="text-center space-y-8">
        <div className="relative w-24 h-24 mx-auto">
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            className="absolute inset-0 bg-red-500/10 rounded-full blur-xl"
          />
          <div className="relative w-full h-full bg-slate-900 border border-red-500/50 rounded-full flex items-center justify-center">
            <span className="material-symbols-outlined text-red-500 text-5xl">
              lock_clock
            </span>
          </div>
        </div>

        <div className="space-y-2">
          <h3 className="text-white text-xl font-bold tracking-tight">
            Account Temporarily Locked
          </h3>
          <p className="text-slate-400 text-sm leading-relaxed px-4">
            {lockInfo?.reason ||
              "For your protection, we have temporarily suspended login access due to unusual activity."}
          </p>
        </div>

        <div className="bg-slate-800/30 border border-white/5 p-6 rounded-3xl mx-4">
          <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest mb-2">
            Cooling Period Remaining
          </p>
          <p className="text-4xl text-white font-black tracking-tighter font-mono">
            {formatTime(timeLeft)}
          </p>
        </div>

        <div className="space-y-3 px-4 pt-4">
          <AuthButton
            onClick={() => (window.location.href = "/login")}
            disabled={timeLeft > 0}
          >
            {timeLeft > 0 ? "Blocked" : "Try Again"}
          </AuthButton>
          <button
            onClick={() =>
              (window.location.href = "mailto:security@flyfleet.com")
            }
            className="w-full py-4 text-xs font-bold text-slate-500 hover:text-white transition-colors uppercase tracking-widest"
          >
            Contact Systems Support
          </button>
        </div>

        <p className="text-[10px] text-slate-600 font-medium">
          Incident Ref:{" "}
          <span className="text-slate-500 font-mono">
            SEC-LOCK-{Math.random().toString(36).substr(2, 6).toUpperCase()}
          </span>
        </p>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default AccountLocked;
