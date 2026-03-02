import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";

const Signup: React.FC = () => {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
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
      const response = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, name }),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || "Registration failed");
      }

      setSuccess(true);
      // In a real app, we'd wait for email verification
      setTimeout(() => navigate("/login"), 5000);
    } catch (err: any) {
      setError(err.message);
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
        <div className="space-y-1.5">
          <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
            {uiStrings?.fullname_label || "Full Name"}
          </label>
          <div className="relative group">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
              person
            </span>
            <input
              type="text"
              required
              className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600"
              placeholder={uiStrings?.fullname_placeholder || "Full Name"}
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
        </div>

        <div className="space-y-1.5">
          <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
            {uiStrings?.email_label || "Work Email"}
          </label>
          <div className="relative group">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
              mail
            </span>
            <input
              type="email"
              required
              className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600"
              placeholder={uiStrings?.email_placeholder || "Work Email"}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
        </div>

        <div className="space-y-1.5">
          <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
            {uiStrings?.password_label || "Set Password"}
          </label>
          <div className="relative group">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
              lock
            </span>
            <input
              type="password"
              required
              className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>

          {/* Password Strength Indicator */}
          <div className="mt-2 px-1">
            <div className="flex justify-between items-center mb-1">
              <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                {uiStrings?.strength_label || "Security Strength"}
              </span>
              <span
                className={`text-[10px] font-bold uppercase tracking-widest ${
                  passwordStrength <= 25
                    ? "text-red-500"
                    : passwordStrength <= 50
                      ? "text-yellow-500"
                      : passwordStrength <= 75
                        ? "text-blue-500"
                        : "text-green-500"
                }`}
              >
                {passwordStrength <= 25
                  ? uiStrings?.strength_weak || "Weak"
                  : passwordStrength <= 50
                    ? uiStrings?.strength_fair || "Fair"
                    : passwordStrength <= 75
                      ? uiStrings?.strength_good || "Good"
                      : uiStrings?.strength_strong || "Strong"}
              </span>
            </div>
            <div className="h-1 w-full bg-slate-800 rounded-full overflow-hidden">
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: `${passwordStrength}%` }}
                className={`h-full transition-colors duration-500 ${
                  passwordStrength <= 25
                    ? "bg-red-500"
                    : passwordStrength <= 50
                      ? "bg-yellow-500"
                      : passwordStrength <= 75
                        ? "bg-blue-500"
                        : "bg-green-500"
                }`}
              />
            </div>
          </div>
        </div>

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
              <a href="#" className="text-primary hover:underline mx-1">
                {uiStrings?.terms_link || "Terms of Service"}
              </a>
              and
              <a href="#" className="text-primary hover:underline ml-1">
                {uiStrings?.privacy_link || "Privacy Policy"}
              </a>
              .
            </span>
          </label>
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest mt-4"
        >
          {loading ? (
            <>
              <svg
                className="animate-spin h-5 w-5 text-white"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                ></circle>
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <span>{uiStrings?.btn_creating || "Creating Account..."}</span>
            </>
          ) : (
            <span>{uiStrings?.register_btn || "Register Now"}</span>
          )}
        </button>
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
