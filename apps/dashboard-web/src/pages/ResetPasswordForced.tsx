import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import AuthButton from "../components/auth/AuthButton";
import AuthInput from "../components/auth/AuthInput";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import PasswordStrengthMeter from "../components/auth/PasswordStrengthMeter";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";
import { useAuth } from "../context/AuthProvider";

const ResetPasswordForced: React.FC = () => {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const navigate = useNavigate();
  const { logout } = useAuth();
  const { uiStrings } = useUIStrings("auth_extended");

  useEffect(() => {
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (/[A-Z]/.test(password)) strength += 25;
    if (/[0-9]/.test(password)) strength += 25;
    if (/[^A-Za-z0-9]/.test(password)) strength += 25;
    setPasswordStrength(strength);
  }, [password]);

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError(uiStrings?.error_mismatch || "Passwords do not match");
      return;
    }
    if (passwordStrength < 75) {
      setError(
        uiStrings?.error_security_req ||
          "Password does not meet platform security requirements"
      );
      return;
    }
    setLoading(true);
    setError("");

    try {
      await api.resetPasswordForced({ new_password: password });
      setSuccess(true);
      // Logout and redirect to login to ensure fresh session with cleared flag
      setTimeout(() => logout(), 3000);
    } catch (err: any) {
      setError(err.message || "Failed to update password");
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout
      title="Secure Account Activation"
      subtitle="To protect your account, please set a permanent password before proceeding."
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

        {success && (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="mb-6 p-4 bg-green-500/10 border border-green-500/50 rounded-xl text-green-500 text-xs font-medium flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">
              check_circle
            </span>
            Password updated successfully. For security, please sign in again with your new password.
          </motion.div>
        )}
      </AnimatePresence>

      {!success && (
        <form onSubmit={handleReset} className="space-y-4">
          <AuthInput
            label="Create New Password"
            icon="key"
            type="password"
            required
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />

          <PasswordStrengthMeter
            strength={passwordStrength}
            label={uiStrings?.strength_label}
            strengthLabels={{
              weak: uiStrings?.strength_weak || "Weak",
              fair: uiStrings?.strength_fair || "Fair",
              good: uiStrings?.strength_good || "Good",
              strong: uiStrings?.strength_strong || "Strong",
            }}
          />

          <AuthInput
            label="Confirm New Password"
            icon="password"
            type="password"
            required
            placeholder="••••••••"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
          />

          <div className="bg-slate-900/50 p-4 rounded-xl border border-white/5 mb-6">
            <h4 className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-2">
              <span className="material-symbols-outlined text-xs">info</span>
              Security Policy
            </h4>
            <ul className="space-y-1">
              <li className={`text-[10px] flex items-center gap-1.5 ${password.length >= 8 ? 'text-green-500' : 'text-slate-500'}`}>
                <span className="material-symbols-outlined text-[12px]">{password.length >= 8 ? 'check' : 'close'}</span>
                Minimum 8 characters
              </li>
              <li className={`text-[10px] flex items-center gap-1.5 ${/[A-Z]/.test(password) ? 'text-green-500' : 'text-slate-500'}`}>
                <span className="material-symbols-outlined text-[12px]">{/[A-Z]/.test(password) ? 'check' : 'close'}</span>
                Contains uppercase letter
              </li>
              <li className={`text-[10px] flex items-center gap-1.5 ${/[0-9]/.test(password) ? 'text-green-500' : 'text-slate-500'}`}>
                <span className="material-symbols-outlined text-[12px]">{/[0-9]/.test(password) ? 'check' : 'close'}</span>
                Contains numeric digit
              </li>
              <li className={`text-[10px] flex items-center gap-1.5 ${/[^A-Za-z0-9]/.test(password) ? 'text-green-500' : 'text-slate-500'}`}>
                <span className="material-symbols-outlined text-[12px]">{/[^A-Za-z0-9]/.test(password) ? 'check' : 'close'}</span>
                Contains special character
              </li>
            </ul>
          </div>

          <AuthButton
            type="submit"
            loading={loading}
            loadingText="Enforcing Security..."
          >
            Activate Account
          </AuthButton>
        </form>
      )}

      {success && (
        <div className="mt-8 pt-4">
          <AuthButton variant="secondary" onClick={() => logout()}>
            Proceed to Login
          </AuthButton>
        </div>
      )}
    </EnterpriseAuthLayout>
  );
};

export default ResetPasswordForced;
