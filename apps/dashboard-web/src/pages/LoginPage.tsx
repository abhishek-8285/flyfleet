import { AnimatePresence, motion } from "framer-motion";
import React, { useState, useEffect, createContext, useContext, useRef } from "react";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useAuth } from "../context/AuthProvider";
import { useUIStrings } from "../hooks/useUIStrings";

enum LoginState {
  IDENTIFY,
  SELECT_ORG,
  AUTHENTICATE,
  MFA,
}

interface Org {
  org_id: string;
  org_name: string;
}

interface GlobalError {
  message: string;
  isRateLimit?: boolean;
  cooldownSeconds?: number;
}

interface LoginContextType {
  state: LoginState;
  setState: (s: LoginState) => void;
  emailOrPhone: string;
  setEmailOrPhone: (val: string) => void;
  organizations: Org[];
  setOrganizations: (orgs: Org[]) => void;
  selectedOrgId: string;
  setSelectedOrgId: (id: string) => void;
  globalError: GlobalError | null;
  setGlobalError: (err: GlobalError | null) => void;
  loading: boolean;
  setLoading: (l: boolean) => void;
  tempToken: string;
  setTempToken: (t: string) => void;
  partialUser: any;
  setPartialUser: (u: any) => void;
  rememberMe: boolean;
  setRememberMe: (val: boolean) => void;
  uiStrings: any;
}

const LoginContext = createContext<LoginContextType | undefined>(undefined);

const useLoginStore = () => {
  const ctx = useContext(LoginContext);
  if (!ctx) throw new Error("Must be used within LoginProvider");
  return ctx;
};

const slideVariants = {
  hidden: { opacity: 0, x: 20 },
  visible: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: -20 },
};

const IdentifyStep: React.FC = () => {
  const { emailOrPhone, setEmailOrPhone, setState, setOrganizations, setSelectedOrgId, setGlobalError, setLoading, uiStrings } = useLoginStore();
  const [localError, setLocalError] = useState("");

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let val = e.target.value;
    if (val.length === 1 && /[0-9]/.test(val)) {
      val = "+91" + val;
    }
    setEmailOrPhone(val);
  };

  const handleIdentify = async (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError("");
    setGlobalError(null);

    const isEmail = z.string().email().safeParse(emailOrPhone).success;
    const isPhone = /^\+?91[\d\s-]{10,}$/.test(emailOrPhone);
    if (!isEmail && !isPhone) {
      setLocalError("Please enter a valid Operator Email or Indian Phone (+91)");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("/api/v1/auth/resolve-orgs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: emailOrPhone }),
      });

      if (response.status === 429) {
         setGlobalError({ 
           message: "Too many attempts. Please wait.", 
           isRateLimit: true, 
           cooldownSeconds: parseInt(response.headers.get("Retry-After") || "60") 
         });
         return;
      }

      if (!response.ok) {
        setState(LoginState.AUTHENTICATE);
        return;
      }
      
      const data = await response.json();
      const orgs = data.organizations || [];
      setOrganizations(orgs);
      
      if (orgs.length === 1) {
        setSelectedOrgId(orgs[0].org_id);
        setState(LoginState.AUTHENTICATE);
      } else if (orgs.length > 1) {
        setState(LoginState.SELECT_ORG);
      } else {
        setState(LoginState.AUTHENTICATE);
      }
    } catch (err: any) {
      setGlobalError({ message: "Network error occurred." });
    } finally {
      setLoading(false);
    }
  };

  return (
    <motion.form variants={slideVariants} initial="hidden" animate="visible" exit="exit" onSubmit={handleIdentify} className="space-y-4">
      <div className="space-y-1">
        <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
          {uiStrings?.label_email || "Operator Email or Phone"}
        </label>
        <div className="relative group">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
            badge
          </span>
          <input
            type="text"
            required
            autoFocus
            className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-slate-600"
            placeholder={uiStrings?.placeholder_email || "name@operator.com or +91..."}
            value={emailOrPhone}
            onChange={handleChange}
          />
        </div>
        {localError && <p className="text-red-500 text-xs ml-1 mt-1">{localError}</p>}
      </div>

      <button type="submit" className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-3 rounded-2xl transition-all shadow-xl shadow-primary/20 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest mt-2">
        <span className="material-symbols-outlined text-sm">arrow_forward</span>
        <span>Continue</span>
      </button>
    </motion.form>
  );
};

const OrgSelectStep: React.FC = () => {
  const { organizations, setSelectedOrgId, setState } = useLoginStore();
  
  return (
    <motion.div variants={slideVariants} initial="hidden" animate="visible" exit="exit" className="space-y-4">
      <h3 className="text-sm font-bold text-slate-300 uppercase tracking-widest text-center mb-4">Select Workspace</h3>
      <div className="grid grid-cols-1 gap-3">
        {organizations.map(org => (
          <button
            key={org.org_id}
            onClick={() => { setSelectedOrgId(org.org_id); setState(LoginState.AUTHENTICATE); }}
            className="p-4 bg-slate-800/30 border border-white/5 rounded-2xl hover:border-primary/50 hover:bg-slate-800/50 transition-all text-left flex items-center gap-3 group"
          >
            <div className="w-10 h-10 rounded-xl bg-slate-900 flex items-center justify-center text-primary group-hover:bg-primary group-hover:text-white transition-colors">
              <span className="material-symbols-outlined">corporate_fare</span>
            </div>
            <div>
              <div className="text-sm font-bold text-white">{org.org_name}</div>
              <div className="text-xs text-slate-500 font-mono mt-0.5">{org.org_id.slice(0, 8)}</div>
            </div>
          </button>
        ))}
      </div>
      <button onClick={() => setState(LoginState.IDENTIFY)} className="text-[10px] text-slate-500 hover:text-primary transition-colors block text-center w-full uppercase tracking-widest pt-2">
        Back
      </button>
    </motion.div>
  );
};

const AuthenticateStep: React.FC = () => {
  const { emailOrPhone, selectedOrgId, setState, setGlobalError, setLoading, setTempToken, setPartialUser, rememberMe, setRememberMe, uiStrings } = useLoginStore();
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const { login } = useAuth();
  
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setGlobalError(null);

    try {
      const response = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: emailOrPhone, password, org_id: selectedOrgId }),
      });

      if (response.status === 429) {
         setGlobalError({ 
           message: "Too many attempts. Please wait.", 
           isRateLimit: true, 
           cooldownSeconds: parseInt(response.headers.get("Retry-After") || "60") 
         });
         return;
      }

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || data.message || "Invalid credentials");
      }

      const data = await response.json();

      if (data.mfa_required) {
        setTempToken(data.temp_token);
        setPartialUser(data.user);
        setState(LoginState.MFA);
        return;
      }

      const userData = {
        ...data.user,
        permissions: data.permissions?.map((p: any) => typeof p === "string" ? p : `${p.module}:${p.action}`) || [],
      };

      login(data.token || data.access_token, data.refresh_token || "", userData);

      if (rememberMe) {
        localStorage.setItem("remembered_email", emailOrPhone);
      }
      if (selectedOrgId) {
        localStorage.setItem("org_id", selectedOrgId);
      }
      
      const role = (data.user?.role_id || data.user?.role || "").toLowerCase();
      if (role.includes("driver") || role === "staff") {
        window.location.href = "/dashboard";
      } else if (role.includes("customer") || role.includes("portal") || role === "user") {
        window.location.href = "/portal";
      } else if (role === "super_admin" || role === "admin" || role === "support" || role === "org_admin" || role === "fleet_manager" || role === "dispatcher") {
        window.location.href = "/admin";
      } else {
        // Fallback for unrecognized roles (default to portal to prevent unauthorized dashboard access)
        window.location.href = "/portal";
      }
    } catch (err: any) {
      setGlobalError({ message: err.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <motion.form variants={slideVariants} initial="hidden" animate="visible" exit="exit" onSubmit={handleLogin} className="space-y-4">
      <div className="space-y-1">
        <div className="flex justify-between items-center ml-1">
          <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest">
            {uiStrings?.label_password || "Password or OTP"}
          </label>
          <button type="button" onClick={() => setState(LoginState.IDENTIFY)} className="text-[10px] text-primary hover:underline">
            Change User
          </button>
        </div>
        <div className="relative group">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
            lock
          </span>
          <input
            type={showPassword ? "text" : "password"}
            required
            autoFocus
            className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-12 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-slate-600"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <button type="button" onClick={() => setShowPassword(!showPassword)} className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300 transition-colors">
            <span className="material-symbols-outlined text-xl">
              {showPassword ? "visibility_off" : "visibility"}
            </span>
          </button>
        </div>
      </div>

      <div className="flex justify-end -mt-0.5">
        <Link to="/forgot-password" title="sm" className="text-primary hover:text-primary-dark transition-colors text-[10px] font-bold uppercase tracking-wider">
          {uiStrings?.link_forgot || "Forgot your password?"}
        </Link>
      </div>

      <div className="flex items-center px-1">
        <label className="flex items-center gap-2 cursor-pointer group">
          <input
            type="checkbox"
            className="w-4 h-4 rounded border-white/10 bg-slate-800 text-primary focus:ring-primary/50 transition-all cursor-pointer"
            checked={rememberMe}
            onChange={(e) => setRememberMe(e.target.checked)}
          />
          <span className="text-xs text-slate-400 font-medium group-hover:text-slate-300 transition-colors">
            {uiStrings?.label_remember || "Remember this device"}
          </span>
        </label>
      </div>

      <button type="submit" className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-3 rounded-2xl transition-all shadow-xl shadow-primary/20 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest mt-2">
        <span className="material-symbols-outlined text-sm">login</span>
        <span>{uiStrings?.btn_sign_in || "Sign In"}</span>
      </button>
    </motion.form>
  );
};

const MFAStep: React.FC = () => {
  const { tempToken, partialUser, setGlobalError, setLoading, uiStrings } = useLoginStore();
  const [code, setCode] = useState<string[]>(Array(6).fill(""));
  const [resendTimer, setResendTimer] = useState(60);
  const [canResend, setCanResend] = useState(false);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);
  const { login } = useAuth();
  
  useEffect(() => {
    inputRefs.current[0]?.focus();
  }, []);

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
    setGlobalError(null);

    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Backspace" && !code[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    const fullCode = code.join("");
    if (fullCode.length !== 6) return;

    setLoading(true);
    setGlobalError(null);

    try {
      const response = await fetch("/api/v1/auth/mfa/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ temp_token: tempToken, code: fullCode }),
      });

      if (response.status === 429) {
         setGlobalError({ 
           message: "Too many attempts. Please wait.", 
           isRateLimit: true, 
           cooldownSeconds: parseInt(response.headers.get("Retry-After") || "60") 
         });
         return;
      }

      if (!response.ok) {
        throw new Error("Verification failed. The code may be incorrect or expired.");
      }

      const data = await response.json();
      const userData = {
        ...data.user,
        permissions: data.permissions?.map((p: any) => typeof p === "string" ? p : `${p.module}:${p.action}`) || [],
      };

      login(data.access_token, data.refresh_token, userData);
      window.location.href = "/admin"; // Simple routing map for MFA
    } catch (err: any) {
      setGlobalError({ message: err.message });
    } finally {
      setLoading(false);
    }
  };

  const handleResend = () => {
    if (!canResend) return;
    setResendTimer(60);
    setCanResend(false);
    // Call resend API here if it existed
  };

  return (
    <motion.form variants={slideVariants} initial="hidden" animate="visible" exit="exit" onSubmit={handleVerify} className="space-y-6">
      <div className="text-center mb-6">
        <h3 className="text-sm font-bold text-white uppercase tracking-widest">Two-Factor Auth</h3>
        <p className="text-xs text-slate-500 mt-1">Enter the 6-digit code sent to your device</p>
      </div>
      
      <div className="flex justify-between gap-2 max-w-[320px] mx-auto">
        {code.map((digit, index) => (
          <input
            key={index}
            ref={(el) => { inputRefs.current[index] = el; }}
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
        <button type="submit" disabled={code.join("").length < 6} className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-3 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest">
          {uiStrings?.mfa_btn_confirm || "Confirm & Access"}
        </button>

        <div className="text-center">
          <button type="button" disabled={!canResend} onClick={handleResend} className={`text-[10px] font-bold uppercase tracking-widest transition-all ${canResend ? "text-primary hover:text-primary-dark" : "text-slate-700 cursor-not-allowed"}`}>
            {canResend ? (uiStrings?.mfa_btn_resend || "Resend New Code") : (uiStrings?.mfa_resend_wait?.replace("{timer}", resendTimer.toString()) || `Resend available in ${resendTimer}s`)}
          </button>
        </div>
      </div>
    </motion.form>
  );
};

// Global Error Banner
const GlobalErrorBanner: React.FC = () => {
  const { globalError, setGlobalError } = useLoginStore();
  
  if (!globalError) return null;

  return (
    <motion.div initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: "auto" }} exit={{ opacity: 0, height: 0 }} className={`mb-6 p-3 rounded-xl text-xs font-medium flex items-center gap-2 ${globalError.isRateLimit ? "bg-orange-500/10 border border-orange-500/50 text-orange-500" : "bg-red-500/10 border border-red-500/50 text-red-500"}`}>
      <span className="material-symbols-outlined text-sm">
        {globalError.isRateLimit ? "timer" : "error"}
      </span>
      <div className="flex-1">
        {globalError.message}
        {globalError.isRateLimit && globalError.cooldownSeconds !== undefined && globalError.cooldownSeconds > 0 && (
          <span className="ml-1 font-bold">Try again in {globalError.cooldownSeconds}s.</span>
        )}
      </div>
      <button type="button" onClick={() => setGlobalError(null)} className="text-current opacity-70 hover:opacity-100">
        <span className="material-symbols-outlined text-sm">close</span>
      </button>
    </motion.div>
  );
};

const Login: React.FC = () => {
  const [state, setState] = useState<LoginState>(LoginState.IDENTIFY);
  const [emailOrPhone, setEmailOrPhone] = useState("");
  const [organizations, setOrganizations] = useState<Org[]>([]);
  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [globalError, setGlobalError] = useState<GlobalError | null>(null);
  const [loading, setLoading] = useState(false);
  const [tempToken, setTempToken] = useState("");
  const [partialUser, setPartialUser] = useState<any>(null);
  const [rememberMe, setRememberMe] = useState(false);
  const { uiStrings } = useUIStrings("auth_extended");

  // Handle global rate limit countdown
  useEffect(() => {
    let timer: any;
    if (globalError?.isRateLimit && globalError.cooldownSeconds && globalError.cooldownSeconds > 0) {
      timer = setInterval(() => {
        setGlobalError(prev => {
          if (!prev || prev.cooldownSeconds === undefined) return prev;
          if (prev.cooldownSeconds <= 1) return null; // clear error when done
          return { ...prev, cooldownSeconds: prev.cooldownSeconds - 1 };
        });
      }, 1000);
    }
    return () => clearInterval(timer);
  }, [globalError]);

  return (
    <LoginContext.Provider value={{
      state, setState, emailOrPhone, setEmailOrPhone,
      organizations, setOrganizations, selectedOrgId, setSelectedOrgId,
      globalError, setGlobalError, loading, setLoading,
      tempToken, setTempToken, partialUser, setPartialUser,
      rememberMe, setRememberMe, uiStrings
    }}>
      <EnterpriseAuthLayout
        title={uiStrings?.login_title || "Welcome Back"}
        subtitle={uiStrings?.login_subtitle || "Sign in to your operations center"}
      >
        <AnimatePresence mode="wait">
          {globalError && <GlobalErrorBanner key="error" />}
        </AnimatePresence>

        <div className="relative overflow-hidden">
          <AnimatePresence mode="wait">
            {state === LoginState.IDENTIFY && <IdentifyStep key="identify" />}
            {state === LoginState.SELECT_ORG && <OrgSelectStep key="select-org" />}
            {state === LoginState.AUTHENTICATE && <AuthenticateStep key="auth" />}
            {state === LoginState.MFA && <MFAStep key="mfa" />}
          </AnimatePresence>
        </div>

        {/* Loading Overlay */}
        <AnimatePresence>
          {loading && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="absolute inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm rounded-3xl">
               <svg className="animate-spin h-8 w-8 text-primary" fill="none" viewBox="0 0 24 24">
                 <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                 <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
               </svg>
            </motion.div>
          )}
        </AnimatePresence>

        {state !== LoginState.MFA && (
          <>
            <div className="mt-6">
              <div className="relative mb-5">
                <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-slate-800"></div></div>
                <div className="relative flex justify-center text-[10px] uppercase font-bold tracking-[0.2em]">
                  <span className="bg-[#111827] px-4 text-slate-600">{uiStrings?.label_sso || "Platform SSO"}</span>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <button className="flex items-center justify-center gap-2 bg-slate-800/30 border border-white/5 hover:bg-slate-800/50 py-2.5 rounded-2xl transition-all group">
                  <svg className="w-4 h-4" viewBox="0 0 24 24"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"></path><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"></path><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"></path><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"></path></svg>
                  <span className="text-xs font-bold text-slate-300 group-hover:text-white transition-colors uppercase tracking-tight">Google</span>
                </button>
                <button className="flex items-center justify-center gap-2 bg-slate-800/30 border border-white/5 hover:bg-slate-800/50 py-2.5 rounded-2xl transition-all group">
                  <svg className="w-4 h-4" viewBox="0 0 21 21"><rect fill="#F25022" height="9" width="9" x="1" y="1"></rect><rect fill="#7FBA00" height="9" width="9" x="11" y="1"></rect><rect fill="#00A4EF" height="9" width="9" x="1" y="11"></rect><rect fill="#FFB900" height="9" width="9" x="11" y="11"></rect></svg>
                  <span className="text-xs font-bold text-slate-300 group-hover:text-white transition-colors uppercase tracking-tight">Azure AD</span>
                </button>
              </div>
            </div>

            <div className="mt-6 pt-4 border-t border-slate-800/50 text-center">
              <p className="text-slate-500 text-xs font-medium mt-4">
                {uiStrings?.label_new_officer || "New fleet officer?"}{" "}
                <Link to="/signup" className="text-primary hover:text-primary-dark font-bold uppercase tracking-wider ml-1">
                  {uiStrings?.link_create_account || "Create Account"}
                </Link>
              </p>
            </div>
          </>
        )}
      </EnterpriseAuthLayout>
    </LoginContext.Provider>
  );
};

export default Login;
