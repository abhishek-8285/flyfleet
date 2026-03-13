import { motion } from "framer-motion";
import React, { useState } from "react";
import AuthButton from "../components/auth/AuthButton";
import AuthInput from "../components/auth/AuthInput";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";

const RiskChallenge: React.FC = () => {
  const [step, setStep] = useState<"warning" | "verify" | "success">("warning");
  const [loading, setLoading] = useState(false);
  const [otp, setOtp] = useState("");
  const [error, setError] = useState("");
  const { uiStrings } = useUIStrings("auth_extended");

  const handleStartVerification = () => {
    setStep("verify");
  };

  const handleVerify = async () => {
    if (otp.length !== 6) return;
    setLoading(true);
    setError("");
    try {
      await api.resolveRiskChallenge("mock-id", { otp });
      setStep("success");
      setTimeout(() => (window.location.href = "/dashboard"), 2000);
    } catch (err: any) {
      setError(err.message || "Verification failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.challenge_title || "Step-up Verification"}
      subtitle={
        uiStrings?.challenge_sub || "Additional identity confirmation required"
      }
    >
      <div className="min-h-[340px] flex flex-col justify-center">
        <motion.div
          key={step}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="space-y-8 text-center"
        >
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2 text-left">
              <span className="material-symbols-outlined text-sm">error</span>
              {error}
            </div>
          )}

          {step === "warning" && (
            <div className="space-y-6">
              <div className="w-20 h-20 bg-amber-500/10 rounded-3xl flex items-center justify-center mx-auto border border-amber-500/20">
                <span className="material-symbols-outlined text-4xl text-amber-500">
                  security_update_warning
                </span>
              </div>
              <div className="space-y-2">
                <h3 className="text-white text-lg font-bold">
                  Unusual Login Pattern
                </h3>
                <p className="text-slate-400 text-sm leading-relaxed">
                  We've detected a sign-in attempt that differs from your usual
                  behavior (New location: San Francisco, US). Please confirm
                  your identity to continue.
                </p>
              </div>
              <AuthButton onClick={handleStartVerification}>
                Verify My Identity
              </AuthButton>
            </div>
          )}

          {step === "verify" && (
            <div className="space-y-6">
              <div className="space-y-2">
                <p className="text-white text-sm font-semibold">
                  Enter Security Code
                </p>
                <p className="text-slate-400 text-xs">
                  A verification code has been sent to your primary corporate
                  email.
                </p>
              </div>
              <AuthInput
                label="Verification Code"
                icon="shield_person"
                placeholder="000 000"
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/\D/g, ""))}
                className="text-center text-2xl tracking-[0.5em] font-mono"
              />
              <AuthButton
                onClick={handleVerify}
                loading={loading}
                disabled={otp.length !== 6}
              >
                Continue to Terminal
              </AuthButton>
            </div>
          )}

          {step === "success" && (
            <div className="space-y-6">
              <div className="w-20 h-20 bg-green-500/20 text-green-500 rounded-full flex items-center justify-center mx-auto border border-green-500/30">
                <span className="material-symbols-outlined text-4xl">
                  verified_user
                </span>
              </div>
              <div className="space-y-2">
                <h3 className="text-white text-xl font-bold">
                  Identity Confirmed
                </h3>
                <p className="text-slate-400 text-sm">
                  Access granted. Redirecting to command center...
                </p>
              </div>
            </div>
          )}
        </motion.div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default RiskChallenge;
