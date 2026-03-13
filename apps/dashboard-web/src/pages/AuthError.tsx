import React from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import AuthButton from "../components/auth/AuthButton";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";

const AuthError: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const msg =
    searchParams.get("msg") ||
    "An unexpected error occurred during authentication.";

  return (
    <EnterpriseAuthLayout
      title="Authentication Error"
      subtitle="Security protocol interruption detected"
    >
      <div className="text-center space-y-8">
        <div className="w-24 h-24 bg-red-500/10 rounded-3xl flex items-center justify-center mx-auto border border-red-500/20 rotate-12">
          <span className="material-symbols-outlined text-red-500 text-5xl -rotate-12">
            warning
          </span>
        </div>

        <div className="space-y-4">
          <p className="text-slate-300 text-sm leading-relaxed px-4">{msg}</p>
          <div className="bg-slate-800/50 rounded-2xl p-4 border border-white/5">
            <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest mb-1">
              Error Code
            </p>
            <p className="text-white font-mono text-sm">
              ERR_AUTH_PROTOCOL_VIOLATION
            </p>
          </div>
        </div>

        <div className="space-y-3 pt-4">
          <AuthButton onClick={() => navigate("/login")}>
            Retry Login
          </AuthButton>
          <AuthButton
            variant="secondary"
            onClick={() =>
              (window.location.href = "mailto:support@flyfleet.io")
            }
          >
            Contact Systems Admin
          </AuthButton>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default AuthError;
