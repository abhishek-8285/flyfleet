import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useDispatch } from "react-redux";
import { AppDispatch } from "../store";
import { setToken, setUser } from "../store/authSlice";
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
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [agreeTerms, setAgreeTerms] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const [step, setStep] = useState<1 | 2>(1);
  const [mode, setMode] = useState<"create" | "invite">("create");
  const [orgName, setOrgName] = useState("");
  const [inviteCode, setInviteCode] = useState("");
  const { uiStrings } = useUIStrings("auth_extended");
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();

  useEffect(() => {
    // Basic password strength logic
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (/[A-Z]/.test(password)) strength += 25;
    if (/[0-9]/.test(password)) strength += 25;
    if (/[^A-Za-z0-9]/.test(password)) strength += 25;
    setPasswordStrength(strength);
  }, [password]);

  const handleNextStep = (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setError(uiStrings?.invalid_email || "Please enter a valid email address");
      return;
    }
    if (password.length < 8) {
      setError(uiStrings?.password_too_short || "Password must be at least 8 characters long");
      return;
    }
    if (!/[A-Z]/.test(password)) {
      setError("Password must contain at least one uppercase letter");
      return;
    }
    if (!/[a-z]/.test(password)) {
      setError("Password must contain at least one lowercase letter");
      return;
    }
    if (!/[0-9]/.test(password)) {
       setError("Password must contain at least one number");
       return;
    }
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
    setError("");
    setStep(2);
  };

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (mode === "create" && !orgName) {
      setError("Organization name is required");
      return;
    }
    if (mode === "invite" && !inviteCode) {
      setError("Invite code is required");
      return;
    }

    setLoading(true);
    setError("");

    try {
      const payload: any = { 
        email, 
        password, 
        name,
        agree_terms: agreeTerms 
      };
      if (mode === "create") payload.org_name = orgName;
      else payload.invite_code = inviteCode;

      const res = await api.signup(payload);

      if (res && res.access_token) {
        // Store tokens in Redux (memory only) - SECURE
        dispatch(setToken(res.access_token));
        if (res.refresh_token) {
          // Note: refresh token may also need storage if used; currently in Redux memory
          // If refresh is handled via httpOnly cookie, no need to store
        }
        if (res.user) {
          dispatch(setUser(res.user));
        }
      }
      
      const user = res.user;
      if (user?.role === "org_admin") navigate("/admin");
      else if (user?.role === "fleet_manager") navigate("/dashboard");
      else if (user?.role === "dispatcher") navigate("/dispatch");
      else if (user?.role === "driver") navigate("/driver");
      else {
         setSuccess(true);
         setTimeout(() => navigate("/login"), 5000);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || "Registration failed");
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <EnterpriseAuthLayout
        title={uiStrings?.verify_sent || "Verification Sent"}
        subtitle={uiStrings?.verify_sub || "Check your operator inbox"}
      >
        <div className="text-center space-y-6">
          <div className="w-20 h-20 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4 border border-green-500/30">
            <span className="material-symbols-outlined text-green-500 text-4xl">
              mark_email_read
            </span>
          </div>
          <p className="text-slate-600 text-sm leading-relaxed">
            {uiStrings?.verification_msg_prefix ||
              "We've sent a secure verification link to"}{" "}
            <span className="text-slate-900 font-bold">{email}</span>.{" "}
            {uiStrings?.verification_msg_suffix ||
              "Please follow the link to activate your platform account."}
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
      title={step === 1 ? (uiStrings?.signup_title || "Create Account") : "Organization Setup"}
      subtitle={step === 1 ? (uiStrings?.signup_sub || "Join your fleet organization") : "Create new or join via invite"}
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

      <form onSubmit={step === 1 ? handleNextStep : handleSignup} className="space-y-4">
        {step === 1 && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-4">
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
              type={showPassword ? "text" : "password"}
              required
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              rightElement={
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="text-slate-500 hover:text-primary transition-colors focus:outline-none flex items-center justify-center p-2 rounded-lg"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  <span className="material-symbols-outlined text-[20px]">
                    {showPassword ? "visibility_off" : "visibility"}
                  </span>
                </button>
              }
            />

            {password && password.length > 0 && (
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
            )}

            <AuthInput
              label={uiStrings?.confirm_password_label || "Confirm Password"}
              icon="password"
              type={showConfirmPassword ? "text" : "password"}
              required
              placeholder="••••••••"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              rightElement={
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="text-slate-500 dark:text-slate-400 hover:text-primary dark:hover:text-white transition-colors focus:outline-none flex items-center justify-center p-2 rounded-lg"
                  aria-label={showConfirmPassword ? "Hide password" : "Show password"}
                >
                  <span className="material-symbols-outlined text-[20px]">
                    {showConfirmPassword ? "visibility_off" : "visibility"}
                  </span>
                </button>
              }
            />

            <div className="pt-2 px-1">
              <label className="flex items-start gap-3 cursor-pointer group">
                <input
                  type="checkbox"
                  required
                  className="mt-1 w-4 h-4 rounded border-slate-300 dark:border-slate-700 dark:bg-transparent bg-white text-blue-600 focus:ring-blue-600/50 transition-all cursor-pointer"
                  checked={agreeTerms}
                  onChange={(e) => setAgreeTerms(e.target.checked)}
                />
                <span className="text-xs text-slate-600 dark:text-slate-300 font-medium leading-relaxed group-hover:text-slate-800 dark:group-hover:text-white transition-colors">
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
              loadingText="Validating..."
            >
              Continue
            </AuthButton>
          </motion.div>
        )}

        {step === 2 && (
          <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-4">
             <div className="flex bg-slate-100 border border-slate-200 p-1 rounded-md mb-4">
               <button type="button" onClick={() => setMode("create")} className={`flex-1 py-2 text-sm font-medium rounded-md transition-all ${mode === "create" ? "bg-white text-slate-900 shadow-sm border border-slate-200/60" : "text-slate-500 hover:text-slate-700"}`}>
                  Create New
               </button>
               <button type="button" onClick={() => setMode("invite")} className={`flex-1 py-2 text-sm font-medium rounded-md transition-all ${mode === "invite" ? "bg-white text-slate-900 shadow-sm border border-slate-200/60" : "text-slate-500 hover:text-slate-700"}`}>
                  Join via Invite
               </button>
             </div>

             {mode === "create" ? (
               <div className="space-y-2">
                 <AuthInput
                   label="Organization Name"
                   icon="domain"
                   type="text"
                   required
                   placeholder="Acme Logistics Inc."
                   value={orgName}
                   onChange={(e) => setOrgName(e.target.value)}
                 />
                 <p className="text-xs text-slate-500 ml-1">You will be designated as the <strong>Organization Admin</strong>.</p>
               </div>
             ) : (
               <div className="space-y-2">
                 <AuthInput
                   label="Invite Code"
                   icon="confirmation_number"
                   type="text"
                   required
                   placeholder="XY79-BK2M"
                   value={inviteCode}
                   onChange={(e) => setInviteCode(e.target.value)}
                 />
                 <p className="text-xs text-slate-500 ml-1">Your role will be assigned automatically based on the secure invite code.</p>
               </div>
             )}

             <div className="flex gap-3 pt-4">
                <button type="button" onClick={() => setStep(1)} className="px-5 py-3 border border-slate-200 rounded-md shadow-sm text-sm font-semibold text-slate-600 bg-white hover:bg-slate-50 focus:outline-none transition-colors">
                   Back
                </button>
                <div className="flex-1">
                  <AuthButton
                    type="submit"
                    loading={loading}
                    loadingText={uiStrings?.btn_creating || "Creating Account..."}
                  >
                    {uiStrings?.register_btn || "Complete Registration"}
                  </AuthButton>
                </div>
             </div>
          </motion.div>
        )}
      </form>

      <div className="mt-8 text-center">
        <p className="text-slate-500 text-xs font-medium">
          {uiStrings?.already_member || "Already a fleet member?"}{" "}
          <Link
            to="/login"
            className="text-blue-600 hover:text-blue-800 font-semibold uppercase tracking-wider ml-1"
          >
            {uiStrings?.sign_in_link || "Sign In"}
          </Link>
        </p>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default Signup;
