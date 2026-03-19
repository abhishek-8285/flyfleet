import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";

enum RegisterStep {
  BASIC_INFO,
  ORG_HANDLING,
}

const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const [step, setStep] = useState<RegisterStep>(RegisterStep.BASIC_INFO);
  
  // Basic Info Form
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  
  // Org Handling Form
  const [mode, setMode] = useState<"create" | "invite">("create");
  const [orgName, setOrgName] = useState("");
  const [inviteCode, setInviteCode] = useState("");
  
  // JIT Provisioning State
  const [inviteValid, setInviteValid] = useState<boolean | null>(null);
  const [inviteDetails, setInviteDetails] = useState<{org_name: string, role: string} | null>(null);
  const [inviteError, setInviteError] = useState("");
  const [validatingInvite, setValidatingInvite] = useState(false);
  
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  React.useEffect(() => {
    if (step !== RegisterStep.ORG_HANDLING || mode !== "invite") return;

    // Reset validations on change
    setInviteValid(null);
    setInviteDetails(null);
    setInviteError("");

    // UUIDv4 length is exactly 36 characters. Let's start validation if it reaches that length.
    // For slightly shorter formats or copy-paste we can trim it. 
    const code = inviteCode.trim();
    if (code.length < 10) return; // Prevent early spams

    const validate = async () => {
      setValidatingInvite(true);
      try {
        const res = await fetch("/api/v1/auth/invites/validate", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ code }),
        });
        
        const data = await res.json();
        
        if (res.status === 429) {
          setInviteError("Too many validation attempts. Rate limited.");
          setInviteValid(false);
          return;
        }

        if (res.ok && data.valid) {
          setInviteValid(true);
          setInviteDetails({
            org_name: data.org_name,
            role: data.role,
          });
        } else {
          setInviteValid(false);
          setInviteError("Invalid or expired invite code");
        }
      } catch (err) {
        setInviteError("Failed to validate invite");
      } finally {
        setValidatingInvite(false);
      }
    };

    const timer = setTimeout(validate, 600); // 600ms debounce to prevent rate limit blows while typing
    return () => clearTimeout(timer);
  }, [inviteCode, step, mode]);

  const handleBasicInfoSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name || !email || !password || !confirmPassword) {
      setError("All fields are required");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    const emailValid = z.string().email().safeParse(email).success;
    if (!emailValid) {
      setError("Please enter a valid email address");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters long");
      return;
    }

    setStep(RegisterStep.ORG_HANDLING);
  };

  const handleOrgSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (mode === "create" && !orgName) {
      setError("Organization name is required");
      return;
    }

    if (mode === "invite" && !inviteValid) {
      setError("Please enter a valid invite code before continuing");
      return;
    }

    setLoading(true);
    try {
      const payload: any = {
        name,
        email,
        password,
      };

      if (mode === "create") {
        payload.org_name = orgName;
      } else {
        payload.invite_code = inviteCode;
      }

      const response = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Registration failed");
      }

      // Automatically store tokens
      localStorage.setItem("token", data.access_token);
      localStorage.setItem("refreshToken", data.refresh_token);
      
      const user = data.user;
      
      // Redirect Logic Enforcement
      if (user.role === "org_admin") {
        navigate("/admin");
      } else if (user.role === "fleet_manager") {
        navigate("/dashboard");
      } else if (user.role === "dispatcher") {
        navigate("/dispatch");
      } else if (user.role === "driver") {
        navigate("/driver");
      } else {
        navigate("/");
      }

    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col justify-center py-12 sm:px-6 lg:px-8 relative overflow-hidden">
        {/* Background Decorative */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[400px] bg-primary/20 blur-[120px] rounded-full pointer-events-none" />
        
        <div className="sm:mx-auto sm:w-full sm:max-w-md relative z-10">
          <div className="flex justify-center flex-col items-center">
             <div className="w-16 h-16 bg-gradient-to-tr from-primary to-blue-500 rounded-2xl flex items-center justify-center shadow-lg shadow-primary/20">
               <span className="material-symbols-outlined text-white text-3xl">flight_takeoff</span>
             </div>
             <h2 className="mt-6 text-center text-3xl font-extrabold text-white tracking-tight">
               Enterprise Registration
             </h2>
             <p className="mt-2 text-center text-sm text-slate-400">
                Create a FlyFleet account
                <Link to="/login" className="ml-1 font-medium text-primary hover:text-primary/80 transition-colors">
                  or sign in to an existing one
                </Link>
             </p>
          </div>
        </div>

        <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md relative z-10">
          <div className="bg-slate-900 border border-slate-800 py-8 px-4 shadow sm:rounded-2xl sm:px-10">
             
             {error && (
               <div className="mb-4 bg-red-500/10 border border-red-500/50 rounded-lg p-3 flex items-start gap-3">
                 <span className="material-symbols-outlined text-red-500">error</span>
                 <p className="text-red-500 text-sm mt-0.5">{error}</p>
               </div>
             )}

             {step === RegisterStep.BASIC_INFO && (
                 <form className="space-y-4" onSubmit={handleBasicInfoSubmit}>
                     <div>
                       <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1 mb-1">Full Name</label>
                       <div className="relative">
                         <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500">person</span>
                         <input type="text" required value={name} onChange={e => setName(e.target.value)}
                           className="w-full bg-slate-800/50 border border-white/10 rounded-xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50" 
                           placeholder="John Doe" />
                       </div>
                     </div>
                     <div>
                       <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1 mb-1">Email Address</label>
                       <div className="relative">
                         <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500">mail</span>
                         <input type="email" required value={email} onChange={e => setEmail(e.target.value)}
                           className="w-full bg-slate-800/50 border border-white/10 rounded-xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50" 
                           placeholder="name@company.com" />
                       </div>
                     </div>
                     <div>
                       <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1 mb-1">Password</label>
                       <div className="relative">
                         <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500">lock</span>
                         <input type="password" required value={password} onChange={e => setPassword(e.target.value)}
                           className="w-full bg-slate-800/50 border border-white/10 rounded-xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50" 
                           placeholder="••••••••" />
                       </div>
                     </div>
                     <div>
                       <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1 mb-1">Confirm Password</label>
                       <div className="relative">
                         <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500">lock_reset</span>
                         <input type="password" required value={confirmPassword} onChange={e => setConfirmPassword(e.target.value)}
                           className="w-full bg-slate-800/50 border border-white/10 rounded-xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50" 
                           placeholder="••••••••" />
                       </div>
                     </div>

                     <button type="submit" className="w-full flex justify-center py-3 px-4 border border-transparent rounded-xl shadow-sm text-sm font-medium text-white bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary focus:ring-offset-slate-900 transition-colors mt-6">
                        Continue to Organization
                     </button>
                 </form>
             )}

             {step === RegisterStep.ORG_HANDLING && (
                 <form className="space-y-4" onSubmit={handleOrgSubmit}>
                    
                    <div className="flex bg-slate-800/50 p-1 rounded-xl mb-6">
                       <button type="button" onClick={() => setMode("create")} className={`flex-1 py-2 text-sm font-medium rounded-lg transition-all ${mode === "create" ? "bg-primary text-white shadow" : "text-slate-400 hover:text-white"}`}>
                          Create New
                       </button>
                       <button type="button" onClick={() => setMode("invite")} className={`flex-1 py-2 text-sm font-medium rounded-lg transition-all ${mode === "invite" ? "bg-primary text-white shadow" : "text-slate-400 hover:text-white"}`}>
                          Join via Invite
                       </button>
                    </div>

                    {mode === "create" ? (
                      <div className="animate-in fade-in slide-in-from-right-4 duration-300">
                         <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1 mb-1">Organization Name</label>
                         <div className="relative">
                           <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500">domain</span>
                           <input type="text" required value={orgName} onChange={e => setOrgName(e.target.value)}
                             className="w-full bg-slate-800/50 border border-white/10 rounded-xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50" 
                             placeholder="Acme Logistics Inc." />
                         </div>
                         <p className="text-xs text-slate-500 mt-2 ml-1">You will be designated as the <strong>Organization Admin</strong> for this new tenant.</p>
                      </div>
                    ) : (
                      <div className="animate-in fade-in slide-in-from-right-4 duration-300">
                         <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1 mb-1 flex justify-between">
                            <span>Invite Code</span>
                            {validatingInvite && <span className="text-primary/70 animate-pulse lowercase text-[10px]">Validating...</span>}
                         </label>
                         <div className="relative">
                           <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500">confirmation_number</span>
                           <input type="text" required value={inviteCode} onChange={e => setInviteCode(e.target.value)}
                             className={`w-full bg-slate-800/50 border rounded-xl pl-12 pr-10 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 font-mono tracking-widest uppercase transition-colors ${inviteValid === true ? 'border-green-500/50' : inviteValid === false ? 'border-red-500/50' : 'border-white/10'}`} 
                             placeholder="XXXXXXXX-XXXX-4XXX-YXXX-XXXXXXXXXXXX" maxLength={36} />
                             
                             {inviteValid === true && (
                                <span className="absolute right-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-green-500">check_circle</span>
                             )}
                             {inviteValid === false && (
                                <span className="absolute right-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-red-500">cancel</span>
                             )}
                         </div>
                         
                         {inviteValid === true && inviteDetails ? (
                           <div className="mt-3 p-3 bg-green-500/10 border border-green-500/20 rounded-lg flex items-start gap-3 animate-in fade-in slide-in-from-top-2">
                             <span className="material-symbols-outlined text-green-500 text-lg mt-0.5">verified_user</span>
                             <div>
                               <p className="text-sm text-green-400 font-medium">Invite Verified</p>
                               <p className="text-xs text-green-500/80 mt-0.5">Joining <strong>{inviteDetails.org_name}</strong> as <strong>{inviteDetails.role}</strong>.</p>
                             </div>
                           </div>
                         ) : inviteError ? (
                           <p className="text-xs text-red-400 mt-2 ml-1 animate-in fade-in">{inviteError}</p>
                         ) : (
                           <p className="text-xs text-slate-500 mt-2 ml-1">Paste your secure 36-character invite code to join.</p>
                         )}
                      </div>
                    )}

                    <div className="flex gap-3 mt-8">
                       <button type="button" onClick={() => setStep(RegisterStep.BASIC_INFO)} className="px-4 py-3 border border-slate-700 rounded-xl shadow-sm text-sm font-medium text-slate-300 bg-slate-800 hover:bg-slate-700 focus:outline-none transition-colors">
                          Back
                       </button>
                       <button type="submit" disabled={loading || (mode === "invite" && inviteValid !== true)} className="flex-1 flex justify-center items-center gap-2 py-3 px-4 border border-transparent rounded-xl shadow-sm text-sm font-medium text-white bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary focus:ring-offset-slate-900 transition-colors disabled:opacity-70 disabled:cursor-not-allowed">
                          {loading || validatingInvite ? (
                            <span className="material-symbols-outlined animate-spin">progress_activity</span>
                          ) : (
                            <span className="material-symbols-outlined">how_to_reg</span>
                          )}
                          Complete Registration
                       </button>
                    </div>

                 </form>
             )}

          </div>
        </div>
    </div>
  );
};

export default RegisterPage;
