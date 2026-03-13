import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import AuthButton from "../../components/auth/AuthButton";
import AuthInput from "../../components/auth/AuthInput";
import EnterpriseAuthLayout from "../../components/auth/EnterpriseAuthLayout";
import PasswordStrengthMeter from "../../components/auth/PasswordStrengthMeter";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";

const ChangePassword: React.FC = () => {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const { uiStrings } = useUIStrings("auth_extended");

  useEffect(() => {
    let strength = 0;
    if (newPassword.length >= 8) strength += 25;
    if (/[A-Z]/.test(newPassword)) strength += 25;
    if (/[0-9]/.test(newPassword)) strength += 25;
    if (/[^A-Za-z0-9]/.test(newPassword)) strength += 25;
    setPasswordStrength(strength);
  }, [newPassword]);

  const handleChange = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
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
      await api.changePassword({ currentPassword, newPassword });
      setSuccess(true);
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err: any) {
      setError(err.message || "Failed to change password");
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.change_password_title || "Change Password"}
      subtitle={
        uiStrings?.change_password_subtitle ||
        "Update your security credentials"
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
            <span className="material-symbols-outlined text-sm">verified</span>
            {uiStrings?.msg_password_changed ||
              "Your password has been successfully updated."}
          </motion.div>
        )}
      </AnimatePresence>

      <form onSubmit={handleChange} className="space-y-4">
        <AuthInput
          label={uiStrings?.label_current_password || "Current Password"}
          icon="lock"
          type="password"
          required
          placeholder="••••••••"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
        />

        <div className="pt-2 border-t border-white/5 my-2"></div>

        <AuthInput
          label={uiStrings?.label_new_password || "New Password"}
          icon="key"
          type="password"
          required
          placeholder="••••••••"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
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
          loading={loading}
          loadingText={uiStrings?.btn_updating_password || "Updating..."}
        >
          {uiStrings?.btn_change_password || "Update Password"}
        </AuthButton>
      </form>
    </EnterpriseAuthLayout>
  );
};

export default ChangePassword;
