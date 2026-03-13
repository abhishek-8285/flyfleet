import React from "react";
import { useNavigate } from "react-router-dom";
import AuthButton from "../components/auth/AuthButton";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";

const AccessDenied: React.FC = () => {
  const navigate = useNavigate();

  return (
    <EnterpriseAuthLayout
      title="403 - Restricted Access"
      subtitle="Enterprise IAM: Permission Denied"
    >
      <div className="text-center space-y-8">
        <div className="w-24 h-24 bg-slate-800/40 rounded-full flex items-center justify-center mx-auto border border-white/5 shadow-inner">
          <span className="material-symbols-outlined text-slate-400 text-5xl">
            lock_open_right
          </span>
        </div>

        <div className="space-y-4">
          <p className="text-slate-300 text-sm leading-relaxed">
            Your current security clearance level is insufficient to access this
            terminal. Required role:{" "}
            <span className="text-primary font-bold">SYSTEM_ADMIN</span>
          </p>
          <p className="text-slate-500 text-[10px] uppercase font-bold tracking-widest leading-loose">
            If you believe this is an error, please verify your organizational
            role with your supervisor.
          </p>
        </div>

        <div className="pt-4 space-y-3">
          <AuthButton onClick={() => navigate("/dashboard")}>
            Return to Dashboard
          </AuthButton>
          <AuthButton variant="secondary" onClick={() => navigate("/login")}>
            Switch Identity
          </AuthButton>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default AccessDenied;
