import { AnimatePresence, motion } from "framer-motion";
import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";

const ForgotPassword: React.FC = () => {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("auth_extended");

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    setMessage("");

    try {
      // API call placeholder for password reset initiation
      await fetch("/api/v1/auth/forgot-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });

      // Note: Backend might not return 404 for security reasons (don't reveal users)
      setMessage(
        uiStrings?.msg_email_sent_body ||
          "If an operator account exists for this email, you will receive a reset link shortly.",
      );
    } catch (err: any) {
      setError(
        uiStrings?.error_connection ||
          "A secure connection could not be established. Please try again later.",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.recover_title || "Secure Recovery"}
      subtitle={
        uiStrings?.recover_subtitle || "A reset link will be sent to your inbox"
      }
    >
      <AnimatePresence mode="wait">
        {error && (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="mb-6 p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </motion.div>
        )}

        {message && (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="mb-6 p-4 bg-primary/10 border border-primary/50 rounded-xl text-primary text-xs font-medium leading-relaxed"
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="material-symbols-outlined text-sm">
                mark_email_unread
              </span>
              <span className="uppercase tracking-widest font-bold">
                {uiStrings?.msg_email_sent_title || "Email Sent"}
              </span>
            </div>
            {message}
          </motion.div>
        )}
      </AnimatePresence>

      {!message ? (
        <form onSubmit={handleReset} className="space-y-6">
          <div className="space-y-1.5">
            <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
              {uiStrings?.label_email || "Operator Email"}
            </label>
            <div className="relative group">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
                mail
              </span>
              <input
                type="email"
                required
                className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-4 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600"
                placeholder={uiStrings?.placeholder_email || "name@company.com"}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest"
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
                <span>
                  {uiStrings?.btn_sending_link || "Sending Secure Link..."}
                </span>
              </>
            ) : (
              <span>
                {uiStrings?.btn_send_recover || "Send Recovery Email"}
              </span>
            )}
          </button>
        </form>
      ) : (
        <div className="pt-4">
          <button
            onClick={() => navigate("/login")}
            className="w-full bg-slate-800/50 hover:bg-slate-800 text-white font-bold py-4 rounded-2xl transition-all border border-white/5 text-sm uppercase tracking-widest"
          >
            {uiStrings?.btn_back_to_signin || "Back to Sign In"}
          </button>
        </div>
      )}

      {!message && (
        <div className="mt-8 text-center">
          <Link
            to="/login"
            className="text-slate-500 hover:text-primary text-[10px] font-bold uppercase tracking-[0.2em] transition-colors"
          >
            {uiStrings?.link_return_login || "Return to Login Authority"}
          </Link>
        </div>
      )}
    </EnterpriseAuthLayout>
  );
};

export default ForgotPassword;
