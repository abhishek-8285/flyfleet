import { AnimatePresence, motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import AuthButton from "../components/auth/AuthButton";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";
import api from "../services/api";

const VerifyEmail: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("auth_extended");
  const token = searchParams.get("token");

  useEffect(() => {
    const verify = async () => {
      if (!token) {
        setError(
          uiStrings?.error_token_missing ||
            "Missing verification token. Please check your email link.",
        );
        setLoading(false);
        return;
      }

      try {
        await api.verifyEmail(token);
        setSuccess(true);
        setTimeout(() => navigate("/login"), 3000);
      } catch (err: any) {
        setError(err.message || "Verification failed");
      } finally {
        setLoading(false);
      }
    };

    verify();
  }, [token, navigate, uiStrings]);

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.verify_title || "Identity Verification"}
      subtitle={
        uiStrings?.verify_subtitle || "Finalizing your platform registration"
      }
    >
      <div className="flex flex-col items-center justify-center min-h-[200px]">
        <AnimatePresence mode="wait">
          {loading ? (
            <motion.div
              key="loading"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="flex flex-col items-center gap-4"
            >
              <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
              <p className="text-slate-400 text-sm font-medium animate-pulse">
                {uiStrings?.verify_loading || "Verifying credentials..."}
              </p>
            </motion.div>
          ) : error ? (
            <motion.div
              key="error"
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              className="text-center"
            >
              <div className="w-16 h-16 bg-red-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-red-500/20">
                <span className="material-symbols-outlined text-red-500 text-3xl">
                  error
                </span>
              </div>
              <h3 className="text-white font-bold mb-2">
                {uiStrings?.verify_error_title || "Verification Error"}
              </h3>
              <p className="text-slate-400 text-sm mb-6 max-w-[280px] mx-auto">
                {error}
              </p>
              <AuthButton
                variant="secondary"
                onClick={() => navigate("/login")}
              >
                {uiStrings?.btn_back_to_signin || "Back to Login"}
              </AuthButton>
            </motion.div>
          ) : success ? (
            <motion.div
              key="success"
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              className="text-center"
            >
              <div className="w-16 h-16 bg-green-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-green-500/20">
                <span className="material-symbols-outlined text-green-500 text-3xl">
                  verified_user
                </span>
              </div>
              <h3 className="text-white font-bold mb-2">
                {uiStrings?.verify_success_title || "Account Verified"}
              </h3>
              <p className="text-slate-400 text-sm mb-4">
                {uiStrings?.verify_success_body ||
                  "Your platform identity has been confirmed."}
              </p>
              <p className="text-primary text-[10px] font-bold uppercase tracking-widest animate-pulse">
                {uiStrings?.verify_redirecting ||
                  "Redirecting to Access Portal..."}
              </p>
            </motion.div>
          ) : null}
        </AnimatePresence>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default VerifyEmail;
