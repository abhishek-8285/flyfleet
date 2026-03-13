import React, { useEffect, useState } from "react";
import AuthButton from "../../components/auth/AuthButton";
import EnterpriseAuthLayout from "../../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";
import { PasswordPolicy as IPasswordPolicy } from "../../types/auth";

const PasswordPolicy: React.FC = () => {
  const [policy, setPolicy] = useState<IPasswordPolicy | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { uiStrings } = useUIStrings("auth_extended");

  useEffect(() => {
    const fetchPolicy = async () => {
      try {
        const resp = await api.getPasswordPolicy();
        setPolicy(resp.data);
      } catch (err: any) {
        setError(err.message || "Failed to load password policy");
        // Mock fallback
        setPolicy({
          minLength: 12,
          requireUppercase: true,
          requireLowercase: true,
          requireNumbers: true,
          requireSpecial: true,
          historyCount: 5,
          expiryDays: 90,
          breachDetection: true,
        });
      } finally {
        setLoading(false);
      }
    };
    fetchPolicy();
  }, []);

  const Requirement: React.FC<{
    label: string;
    value: string | number | boolean;
    icon: string;
  }> = ({ label, value, icon }) => (
    <div className="bg-slate-800/30 border border-white/5 p-4 rounded-2xl flex items-center justify-between">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 bg-slate-700/50 rounded-lg flex items-center justify-center border border-white/5">
          <span className="material-symbols-outlined text-slate-400 text-lg">
            {icon}
          </span>
        </div>
        <span className="text-sm font-medium text-slate-300">{label}</span>
      </div>
      <div className="text-white font-bold text-sm tracking-tight capitalize">
        {typeof value === "boolean" ? (value ? "Enabled" : "Disabled") : value}
      </div>
    </div>
  );

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.policy_title || "Password Policy"}
      subtitle={
        uiStrings?.policy_sub || "Organizational identity compliance rules"
      }
    >
      <div className="space-y-6">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2">
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 gap-3">
          {loading ? (
            <div className="py-20 flex flex-col items-center gap-4">
              <div className="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
              <p className="text-slate-500 text-xs font-medium animate-pulse">
                Loading compliance data...
              </p>
            </div>
          ) : (
            policy && (
              <>
                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div className="bg-primary/10 border border-primary/30 p-4 rounded-2xl text-center space-y-1">
                    <p className="text-[10px] text-primary font-bold uppercase tracking-widest leading-none">
                      Min. Length
                    </p>
                    <p className="text-3xl text-white font-black">
                      {policy.minLength}
                    </p>
                    <p className="text-[9px] text-slate-500 font-medium">
                      Characters required
                    </p>
                  </div>
                  <div className="bg-blue-500/10 border border-blue-500/30 p-4 rounded-2xl text-center space-y-1">
                    <p className="text-[10px] text-blue-400 font-bold uppercase tracking-widest leading-none">
                      Rotation
                    </p>
                    <p className="text-3xl text-white font-black">
                      {policy.expiryDays}
                    </p>
                    <p className="text-[9px] text-slate-500 font-medium">
                      Days until expiry
                    </p>
                  </div>
                </div>
                <Requirement
                  label="Upper & Lowercase"
                  value={policy.requireUppercase && policy.requireLowercase}
                  icon="text_fields"
                />
                <Requirement
                  label="Numeric Digits"
                  value={policy.requireNumbers}
                  icon="numbers"
                />
                <Requirement
                  label="Special Characters"
                  value={policy.requireSpecial}
                  icon="terminal"
                />
                <Requirement
                  label="Password History"
                  value={`Last ${policy.historyCount} passwords`}
                  icon="history"
                />
                <Requirement
                  label="Breach Detection"
                  value={policy.breachDetection}
                  icon="shield_locked"
                />

                <div className="pt-2">
                  <p className="text-[10px] text-slate-500 text-center uppercase font-bold tracking-[0.15em] mb-4">
                    Compliance Status:{" "}
                    <span className="text-green-500">Certified</span>
                  </p>
                  <AuthButton
                    variant="secondary"
                    onClick={() =>
                      (window.location.href = "/security/change-password")
                    }
                  >
                    Update My Password
                  </AuthButton>
                </div>
              </>
            )
          )}
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default PasswordPolicy;
