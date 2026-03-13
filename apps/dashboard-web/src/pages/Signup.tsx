import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import AuthButton from "../components/auth/AuthButton";
import AuthInput from "../components/auth/AuthInput";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import PasswordStrengthMeter from "../components/auth/PasswordStrengthMeter";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";

const Signup: React.FC = () => {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [agreeTerms, setAgreeTerms] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const { uiStrings } = useUIStrings("auth_extended");
  const navigate = useNavigate();

  useEffect(() => {
    // Basic password strength logic
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (/[A-Z]/.test(password)) strength += 25;
    if (/[0-9]/.test(password)) strength += 25;
    if (/[^A-Za-z0-9]/.test(password)) strength += 25;
    setPasswordStrength(strength);
  }, [password]);

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError(uiStrings?.error_mismatch || "Passwords do not match");
      return;
    }
    if (!agreeTerms) {
      setError(
        uiStrings?.agree_error ||
          "You must agree to the Terms and Privacy Policy",
      );
      return;
    }
    setLoading(true);
    setError("");

    try {
      await api.signup({ email, password, name });
      setSuccess(true);
      setTimeout(() => navigate("/login"), 5000);
    } catch (err: any) {
      setError(err.message || "Registration failed");
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <EnterpriseAuthLayout
        title={uiStrings?.verify_sent || "Verification Sent"}
        subtitle={uiStrings?.verify_sub || "Check your corporate inbox"}
      >
        <div className="text-center space-y-6">
          <div className="w-20 h-20 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4 border border-green-500/30">
            <span className="material-symbols-outlined text-green-500 text-4xl">
              mark_email_read
            </span>
          </div>
          <p className="text-slate-300 text-sm leading-relaxed">
            {uiStrings?.verification_msg_prefix ||
              "We've sent a secure verification link to"}{" "}
            <span className="text-white font-bold">{email}</span>.{" "}
            {uiStrings?.verification_msg_suffix ||
              "Please follow the link to activate your enterprise account."}
          </p>
          <div className="pt-4">
            <button
              onClick={() => navigate("/login")}
              className="text-primary hover:text-primary-dark font-bold uppercase tracking-widest text-xs flex items-center justify-center gap-2 mx-auto"
            >
              {uiStrings?.back_to_login || "Back to Login"}{" "}
              <span className="material-symbols-outlined text-sm">
                arrow_forward
              </span>
            </button>
          </div>
        </div>
      </EnterpriseAuthLayout>
    );
  }

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.signup_title || "Create Account"}
      subtitle={uiStrings?.signup_sub || "Join your fleet organization"}
    >
      <AnimatePresence mode="wait">
        {error && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            className="mb-6 p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </motion.div>
        )}
      </AnimatePresence>

      <form onSubmit={handleSignup} className="space-y-4">
        <AuthInput
          label={uiStrings?.fullname_label || "Full Name"}
          icon="person"
          type="text"
          required
          placeholder={uiStrings?.fullname_placeholder || "Full Name"}
          value={name}
          onChange={(e) => setName(e.target.value)}
        />

        <AuthInput
          label={uiStrings?.email_label || "Work Email"}
          icon="mail"
          type="email"
          required
          placeholder={uiStrings?.email_placeholder || "Work Email"}
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />

        <AuthInput
          label={uiStrings?.password_label || "Set Password"}
          icon="lock"
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
          label={uiStrings?.confirm_password_label || "Confirm Password"}
          icon="password"
          type="password"
          required
          placeholder="••••••••"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
        />

        <div className="pt-2 px-1">
          <label className="flex items-start gap-3 cursor-pointer group">
            <input
              type="checkbox"
              required
              className="mt-1 w-4 h-4 rounded border-white/10 bg-slate-800 text-primary focus:ring-primary/50 transition-all cursor-pointer"
              checked={agreeTerms}
              onChange={(e) => setAgreeTerms(e.target.checked)}
            />
            <span className="text-xs text-slate-400 font-medium leading-relaxed group-hover:text-slate-300 transition-colors">
              {uiStrings?.agree_text ||
                "I certify that I have read and agree to the"}
              <a
                href="#"
                className="text-primary hover:underline mx-1"
                onClick={(e) => e.preventDefault()}
              >
                {uiStrings?.terms_link || "Terms of Service"}
              </a>
              and
              <a
                href="#"
                className="text-primary hover:underline ml-1"
                onClick={(e) => e.preventDefault()}
              >
                {uiStrings?.privacy_link || "Privacy Policy"}
              </a>
              .
            </span>
          </label>
        </div>

        <AuthButton
          type="submit"
          loading={loading}
          loadingText={uiStrings?.btn_creating || "Creating Account..."}
        >
          {uiStrings?.register_btn || "Register Now"}
        </AuthButton>
      </form>

      <div className="mt-8 text-center">
        <p className="text-slate-500 text-xs font-medium">
          {uiStrings?.already_member || "Already a fleet member?"}{" "}
          <Link
            to="/login"
            className="text-primary hover:text-primary-dark font-bold uppercase tracking-wider ml-1"
          >
            {uiStrings?.sign_in_link || "Sign In"}
          </Link>
        </p>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default Signup;
