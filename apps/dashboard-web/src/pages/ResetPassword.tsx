import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import AuthButton from "../components/auth/AuthButton";
import AuthInput from "../components/auth/AuthInput";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import PasswordStrengthMeter from "../components/auth/PasswordStrengthMeter";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";

const ResetPassword: React.FC = () => {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("auth_extended");
  const token = searchParams.get("token");

  useEffect(() => {
    if (!token) {
      setError(
        uiStrings?.error_token_expired ||
          "Invalid or expired reset token. Please request a new link.",
      );
    }
  }, [token, uiStrings]);

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
          "Password does not meet enterprise security requirements",
      );
      return;
    }
    setLoading(true);
    setError("");

    try {
      await api.resetPassword({ token, newPassword: password });
      setSuccess(true);
      setTimeout(() => navigate("/login"), 3000);
    } catch (err: any) {
      setError(err.message || "Failed to reset password");
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.reset_title || "Reset Password"}
      subtitle={
        uiStrings?.reset_subtitle || "Update your corporate credentials"
      }
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
            {uiStrings?.msg_success_redirect ||
              "Password reset successful. Redirecting to login..."}
          </motion.div>
        )}
      </AnimatePresence>

      {!success && !error.includes("expired reset token") && (
        <form onSubmit={handleReset} className="space-y-4">
          <AuthInput
            label={uiStrings?.label_new_password || "New Password"}
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
            label={uiStrings?.label_confirm_password || "Confirm Password"}
            icon="password"
            type="password"
            required
            placeholder="••••••••"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
          />

          <AuthButton
            type="submit"
            disabled={!token}
            loading={loading}
            loadingText={uiStrings?.btn_updating || "Updating Credentials..."}
          >
            {uiStrings?.btn_update || "Update Password"}
          </AuthButton>
        </form>
      )}

      {(error.includes("expired reset token") || success) && (
        <div className="mt-8 pt-4">
          <AuthButton variant="secondary" onClick={() => navigate("/login")}>
            {uiStrings?.btn_back_to_signin || "Return to Login"}
          </AuthButton>
        </div>
      )}
    </EnterpriseAuthLayout>
  );
};

export default ResetPassword;
