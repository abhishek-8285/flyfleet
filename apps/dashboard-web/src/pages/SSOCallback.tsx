import { motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useDispatch } from "react-redux";
import { AppDispatch } from "../store";
import { setToken, setUser } from "../store/authSlice";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import api from "../services/api";

const SSOCallback: React.FC = () => {
  const [error, setError] = useState("");
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();

  useEffect(() => {
    const handleCallback = async () => {
      const code = searchParams.get("code");
      const state = searchParams.get("state");

      if (!code) {
        setError("Invalid SSO response. Authorization code missing.");
        return;
      }

      try {
        const response = await api.ssoCallback({ code, state });
        if (response.token) {
          // Store token in Redux (memory only) - SECURE
          dispatch(setToken(response.token));
          if (response.user) {
            dispatch(setUser(response.user));
          }
          // Redirect to dashboard
          window.location.href = "/dashboard";
        } else {
          throw new Error("Invalid session token received");
        }
      } catch (err: any) {
        setError(err.message || "SSO authentication failed");
        setTimeout(
          () =>
            navigate(
              "/auth/error?msg=" +
                encodeURIComponent(err.message || "SSO authentication failed"),
            ),
          2000,
        );
      }
    };

    handleCallback();
  }, [searchParams, navigate]);

  return (
    <EnterpriseAuthLayout
      title="Authenticating"
      subtitle="Connecting to Single Sign-On provider"
    >
      <div className="flex flex-col items-center justify-center min-h-[200px] space-y-6">
        {!error ? (
          <>
            <div className="relative">
              <div className="w-16 h-16 border-4 border-primary/10 border-t-primary rounded-full animate-spin" />
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="material-symbols-outlined text-primary animate-pulse">
                  sync
                </span>
              </div>
            </div>
            <p className="text-slate-400 text-sm font-medium animate-pulse uppercase tracking-[0.2em]">
              Establishing Secure Session
            </p>
          </>
        ) : (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="text-center"
          >
            <div className="w-16 h-16 bg-red-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-red-500/20">
              <span className="material-symbols-outlined text-red-500 text-3xl">
                gpp_maybe
              </span>
            </div>
            <p className="text-red-500 text-xs font-bold uppercase tracking-widest mb-2">
              SSO Failure
            </p>
            <p className="text-slate-400 text-sm">{error}</p>
          </motion.div>
        )}
      </div>
    </EnterpriseAuthLayout>
  );
};

export default SSOCallback;
