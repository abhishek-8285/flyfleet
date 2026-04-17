import axios from "axios";
import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import { AppDispatch } from "../../store";
import { setCustomerToken } from "../../store/customerAuthSlice";

const CustomerLogin: React.FC = () => {
  const [apiKey, setApiKey] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [uiStrings, setUiStrings] = useState<any>(null);
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.portal || data.system_metadata?.ui_strings?.portal,
        ),
      )
      .catch((err) => console.error("Failed to fetch UI strings", err));
  }, []);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // For now, we are using API Key based login as per backend implementation
      // Ideally, a proper portal would have email/password, but we'll stick to the verified backend auth
      // which uses API Key for now or we can assume there's a way to get token via API Key
      // The backend endpoint is POST /api/v1/customer/auth/token with body { api_key: "..." }

      const response = await axios.post("/api/v1/customer/auth/token", {
        api_key: apiKey,
      });

      const token = response.data.token;
      if (token) {
        // Store token in Redux (memory only) - SECURE
        dispatch(setCustomerToken(token));
        navigate("/portal/dashboard");
      } else {
        setError("Invalid credentials received");
      }
    } catch (err: any) {
      console.error("Login failed", err);
      setError(
        err.response?.data?.error ||
          "Authentication failed. Please check your API Key.",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-surface-dim flex items-center justify-center p-8 relative overflow-hidden font-sans">
      {/* Dynamic Command Background */}
      <div className="absolute top-0 left-0 w-full h-full overflow-hidden z-0">
        <div className="absolute top-[-20%] left-[-10%] w-[60%] h-[60%] rounded-full bg-primary/20 blur-[150px] animate-pulse" />
        <div className="absolute bottom-[-20%] right-[-10%] w-[60%] h-[60%] rounded-full bg-indigo-500/10 blur-[150px] animate-pulse" style={{ animationDelay: '2s' }} />
        
        {/* Grid Pattern Overlay */}
        <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'radial-gradient(circle, #fff 1px, transparent 1px)', backgroundSize: '40px 40px' }} />
      </div>

      <div className="max-w-[480px] w-full bg-surface-container border border-outline rounded-[2.5rem] p-12 shadow-2xl z-10 relative backdrop-blur-3xl">
        <div className="text-center mb-12">
          <div className="w-20 h-20 bg-primary/10 rounded-3xl flex items-center justify-center mx-auto mb-8 text-primary border border-primary/20 shadow-lg shadow-primary/10">
            <span className="material-symbols-outlined text-[40px] font-bold">
              hub
            </span>
          </div>
          <h1 className="text-4xl font-black text-white mb-3 tracking-tighter uppercase">
            {uiStrings?.portal_title || "Client Command"}
          </h1>
          <p className="text-[11px] font-black text-on-surface-variant uppercase tracking-[0.25em] opacity-60">
            {uiStrings?.portal_subtitle || "Secure Logistic Intelligence Access"}
          </p>
        </div>

        {error && (
          <div className="bg-rose-500/10 border border-rose-500/20 text-rose-400 p-4 rounded-2xl text-[11px] font-bold uppercase tracking-wider mb-8 flex items-center gap-3 animate-in fade-in slide-in-from-top-2 duration-300">
            <span className="material-symbols-outlined text-[18px]">gpp_maybe</span>
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-8">
          <div>
            <label className="block text-[9px] font-black text-on-surface-variant uppercase tracking-[0.3em] mb-3 ml-2">
              {uiStrings?.api_key_label || "Secure Access Token (API)"}
            </label>
            <div className="relative group">
              <span className="material-symbols-outlined absolute left-5 top-1/2 -translate-y-1/2 text-on-surface-variant text-[20px] transition-all group-focus-within:text-primary">
                vpn_key
              </span>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="w-full bg-surface-dim border border-outline rounded-[1.25rem] py-5 pl-14 pr-6 text-white text-sm font-bold focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all placeholder:text-on-surface-variant/20 uppercase tracking-tighter"
                placeholder={
                  uiStrings?.api_key_placeholder || "Enter Deployment Key"
                }
                required
              />
            </div>
            <p className="text-[9px] text-on-surface-variant mt-3 ml-2 font-bold uppercase tracking-widest opacity-40">
              {uiStrings?.api_key_hint ||
                "Credentials issued by Fleet Command HQ."}
            </p>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary hover:bg-primary-dark text-white text-[11px] font-black uppercase tracking-[0.2em] py-5 rounded-[1.25rem] transition-all duration-300 shadow-xl shadow-primary/20 hover:shadow-primary/40 disabled:opacity-70 disabled:cursor-not-allowed flex items-center justify-center gap-3 active:scale-[0.98]"
          >
            {loading ? (
              <>
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span>
                  {uiStrings?.authenticating_label || "Establishing Secure Link..."}
                </span>
              </>
            ) : (
              <>
                <span>{uiStrings?.sign_in_btn || "Initialize Session"}</span>
                <span className="material-symbols-outlined text-base">
                  login
                </span>
              </>
            )}
          </button>
        </form>

        <div className="mt-12 pt-8 border-t border-outline text-center">
          <p className="text-[10px] text-on-surface-variant font-black uppercase tracking-widest">
            {uiStrings?.need_access_label || "Deauthorized?"}{" "}
            <button
              type="button"
              onClick={() => navigate("/support")}
              className="text-primary hover:text-white transition-colors"
            >
              {uiStrings?.contact_support_btn || "Contact Support Core"}
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};

export default CustomerLogin;
