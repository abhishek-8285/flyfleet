import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useAuth } from "../context/AuthProvider";
import { useUIStrings } from "../hooks/useUIStrings";

const MFAVerify: React.FC = () => {
  const [code, setCode] = useState<string[]>(Array(6).fill(""));
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [resendTimer, setResendTimer] = useState(30);
  const [canResend, setCanResend] = useState(false);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();
  const { uiStrings } = useUIStrings("auth_extended");
  const tempToken = location.state?.tempToken;
  const partialUser = location.state?.user;

  useEffect(() => {
    if (!tempToken) {
      navigate("/login");
      return;
    }
    inputRefs.current[0]?.focus();
  }, [tempToken, navigate]);

  useEffect(() => {
    let timer: any;
    if (resendTimer > 0) {
      timer = setInterval(() => setResendTimer((prev) => prev - 1), 1000);
    } else {
      setCanResend(true);
    }
    return () => clearInterval(timer);
  }, [resendTimer]);

  const handleChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return;
    const newCode = [...code];
    newCode[index] = value.slice(-1);
    setCode(newCode);
    setError("");

    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (
    index: number,
    e: React.KeyboardEvent<HTMLInputElement>,
  ) => {
    if (e.key === "Backspace" && !code[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    const fullCode = code.join("");
    if (fullCode.length !== 6) return;

    setLoading(true);
    setError("");

    try {
      const response = await fetch("/api/v1/auth/mfa/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ temp_token: tempToken, code: fullCode }),
      });

      if (!response.ok) {
        throw new Error(
          "Verification failed. The code may be incorrect or expired.",
        );
      }

      const data = await response.json();

      const userData = {
        ...data.user,
        permissions:
          data.permissions?.map((p: any) =>
            typeof p === "string" ? p : `${p.module}:${p.action}`,
          ) || [],
      };

      login(data.access_token, data.refresh_token, userData);
      navigate("/admin");
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleResend = () => {
    if (!canResend) return;
    setResendTimer(30);
    setCanResend(false);
    // Call resend API here
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.verify_title || "Identity Verification"}
      subtitle={uiStrings?.mfa_subtitle || "Enter the 6-digit security code"}
    >
      <AnimatePresence mode="wait">
        {error && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </motion.div>
        )}
      </AnimatePresence>

      <form onSubmit={handleVerify} className="space-y-8">
        <div className="flex justify-between gap-2 max-w-[320px] mx-auto">
          {code.map((digit, index) => (
            <input
              key={index}
              ref={(el) => {
                inputRefs.current[index] = el;
              }}
              type="text"
              inputMode="numeric"
              maxLength={1}
              value={digit}
              onChange={(e) => handleChange(index, e.target.value)}
              onKeyDown={(e) => handleKeyDown(index, e)}
              className="w-10 h-14 sm:w-12 sm:h-16 text-center text-2xl font-bold bg-slate-800/30 border border-white/5 rounded-2xl text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-700"
              placeholder="•"
            />
          ))}
        </div>

        <div className="space-y-4">
          <button
            type="submit"
            disabled={loading || code.join("").length < 6}
            className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest"
          >
            {loading
              ? uiStrings?.mfa_btn_verifying || "Verifying Identity..."
              : uiStrings?.mfa_btn_confirm || "Confirm & Access"}
          </button>

          <div className="text-center">
            <p className="text-slate-500 text-xs font-medium mb-2">
              {uiStrings?.mfa_resend_prompt ||
                "Didn't receive a security code?"}
            </p>
            <button
              type="button"
              disabled={!canResend}
              onClick={handleResend}
              className={`text-[10px] font-bold uppercase tracking-widest transition-all ${
                canResend
                  ? "text-primary hover:text-primary-dark"
                  : "text-slate-700 cursor-not-allowed"
              }`}
            >
              {canResend
                ? uiStrings?.mfa_btn_resend || "Resend New Code"
                : uiStrings?.mfa_resend_wait?.replace(
                    "{timer}",
                    resendTimer.toString(),
                  ) || `Resend available in ${resendTimer}s`}
            </button>
          </div>
        </div>
      </form>

      <div className="mt-10 pt-6 border-t border-slate-800/50">
        <div className="flex flex-col items-center gap-4">
          <p className="text-[10px] font-bold text-slate-600 uppercase tracking-widest">
            {uiStrings?.mfa_alt_methods || "Alternative Methods"}
          </p>
          <div className="flex gap-4">
            <button className="h-10 w-10 bg-slate-800/30 border border-white/5 rounded-xl flex items-center justify-center text-slate-400 hover:text-primary hover:border-primary/30 transition-all group">
              <span className="material-symbols-outlined text-xl">
                smartphone
              </span>
            </button>
            <button className="h-10 w-10 bg-slate-800/30 border border-white/5 rounded-xl flex items-center justify-center text-slate-400 hover:text-primary hover:border-primary/30 transition-all">
              <span className="material-symbols-outlined text-xl">mail</span>
            </button>
            <button className="h-10 w-10 bg-slate-800/30 border border-white/5 rounded-xl flex items-center justify-center text-slate-400 hover:text-primary hover:border-primary/30 transition-all">
              <span className="material-symbols-outlined text-xl">
                shield_person
              </span>
            </button>
          </div>
        </div>
      </div>

      <div className="mt-8 text-center text-[10px] text-slate-500 font-medium uppercase tracking-widest">
        {uiStrings?.mfa_session_ref || "Session Ref: "}
        {tempToken?.slice(0, 8)}
      </div>
    </EnterpriseAuthLayout>
  );
};

export default MFAVerify;
