import { motion } from "framer-motion";
import React, { useState } from "react";
import AuthButton from "../components/auth/AuthButton";
import AuthInput from "../components/auth/AuthInput";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";
import { MFASetupResponse } from "../types/auth";

const MFASetup: React.FC = () => {
  const [step, setStep] = useState<
    "intro" | "scan" | "verify" | "backup" | "success"
  >("intro");
  const [loading, setLoading] = useState(false);
  const [setupData, setSetupData] = useState<MFASetupResponse | null>(null);
  const [otpCode, setOtpCode] = useState("");
  const [error, setError] = useState("");
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const { uiStrings } = useUIStrings("auth_extended");

  const startSetup = async () => {
    setLoading(true);
    setError("");
    try {
      const resp = await api.setupMFA();
      setSetupData(resp);
      setStep("scan");
    } catch (err: any) {
      setError(err.message || "Failed to initialize MFA setup");
    } finally {
      setLoading(false);
    }
  };

  const verifyMFA = async () => {
    if (otpCode.length !== 6) return;
    setLoading(true);
    setError("");
    try {
      await api.verifyMFA(otpCode);
      // Mock generating backup codes
      setBackupCodes([
        "X78A-B92C",
        "K12M-P88L",
        "Q99X-Z12W",
        "L44P-J00K",
        "H55T-G33F",
        "R22S-D11A",
        "B66V-M09N",
        "Y88U-I77O",
      ]);
      setStep("backup");
    } catch (err: any) {
      setError(err.message || "Invalid verification code");
    } finally {
      setLoading(false);
    }
  };

  const containerVariants = {
    hidden: { opacity: 0, x: 20 },
    visible: { opacity: 1, x: 0 },
    exit: { opacity: 0, x: -20 },
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.mfa_setup_title || "MFA Enrollment"}
      subtitle={uiStrings?.mfa_setup_sub || "Strengthen your account security"}
    >
      <div className="min-h-[300px] flex flex-col justify-center">
        <motion.div
          key={step}
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          exit="exit"
          transition={{ duration: 0.3 }}
          className="space-y-6"
        >
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2">
              <span className="material-symbols-outlined text-sm">error</span>
              {error}
            </div>
          )}

          {step === "intro" && (
            <div className="text-center space-y-6">
              <div className="w-20 h-20 bg-primary/10 rounded-3xl flex items-center justify-center mx-auto border border-primary/20">
                <span className="material-symbols-outlined text-4xl text-primary">
                  verified_user
                </span>
              </div>
              <p className="text-slate-400 text-sm leading-relaxed">
                Multi-Factor Authentication adds an extra layer of protection to
                your platform identity. We recommend using Google Authenticator
                or Microsoft Authenticator.
              </p>
              <AuthButton onClick={startSetup} loading={loading}>
                Begin Setup
              </AuthButton>
            </div>
          )}

          {step === "scan" && setupData && (
            <div className="space-y-6 text-center">
              <p className="text-white text-sm font-semibold">
                1. Scan the QR Code
              </p>
              <div className="bg-white p-4 rounded-3xl w-48 h-48 mx-auto flex items-center justify-center border-4 border-slate-800">
                {/* In a real app, render setupData.qrCode here */}
                <div className="w-full h-full bg-slate-100 flex items-center justify-center rounded-xl">
                  <span className="material-symbols-outlined text-slate-800 text-6xl">
                    qr_code_2
                  </span>
                </div>
              </div>
              <div className="bg-slate-800/30 p-3 rounded-xl border border-white/5">
                <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest mb-1">
                  Manual Entry Key
                </p>
                <p className="text-white font-mono text-sm tracking-wider select-all cursor-pointer">
                  {setupData.secret}
                </p>
              </div>
              <AuthButton onClick={() => setStep("verify")}>
                I've Scanned It
              </AuthButton>
            </div>
          )}

          {step === "verify" && (
            <div className="space-y-6 text-center">
              <p className="text-white text-sm font-semibold">
                2. Enter Verification Code
              </p>
              <p className="text-slate-400 text-xs">
                Enter the 6-digit code currently displayed in your authenticator
                app.
              </p>
              <AuthInput
                label="OTP Code"
                icon="password"
                placeholder="000 000"
                value={otpCode}
                onChange={(e) =>
                  setOtpCode(e.target.value.replace(/\D/g, "").slice(0, 6))
                }
                className="text-center text-2xl tracking-[0.5em] font-mono"
              />
              <AuthButton
                onClick={verifyMFA}
                loading={loading}
                disabled={otpCode.length !== 6}
              >
                Verify & Activate
              </AuthButton>
              <button
                onClick={() => setStep("scan")}
                className="text-slate-500 text-xs font-semibold hover:text-white transition-colors"
              >
                Go Back
              </button>
            </div>
          )}

          {step === "backup" && (
            <div className="space-y-6">
              <div className="bg-amber-500/10 border border-amber-500/50 p-4 rounded-2xl flex gap-3">
                <span className="material-symbols-outlined text-amber-500">
                  warning
                </span>
                <p className="text-amber-200 text-xs leading-relaxed">
                  Store these recovery codes in a secure vault. They are the
                  only way to regain access if you lose your device.
                </p>
              </div>
              <div className="grid grid-cols-2 gap-2 font-mono text-xs">
                {backupCodes.map((code, idx) => (
                  <div
                    key={idx}
                    className="bg-slate-800/50 p-2 rounded-lg border border-white/5 text-slate-300 text-center"
                  >
                    {code}
                  </div>
                ))}
              </div>
              <AuthButton onClick={() => setStep("success")}>
                I have saved these codes
              </AuthButton>
            </div>
          )}

          {step === "success" && (
            <div className="text-center space-y-6 pt-4">
              <div className="w-20 h-20 bg-green-500/20 text-green-500 rounded-full flex items-center justify-center mx-auto border border-green-500/30 animate-bounce">
                <span className="material-symbols-outlined text-4xl">
                  check_circle
                </span>
              </div>
              <h3 className="text-white text-xl font-bold">MFA Activated</h3>
              <p className="text-slate-400 text-sm leading-relaxed">
                Your account is now protected with operations-grade Multi-Factor
                Authentication.
              </p>
              <AuthButton
                variant="secondary"
                onClick={() => (window.location.href = "/security/users")}
              >
                Go to Security Settings
              </AuthButton>
            </div>
          )}
        </motion.div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default MFASetup;
