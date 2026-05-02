import {
    ArrowLeftIcon,
    CheckCircleIcon,
    SparklesIcon,
} from "@heroicons/react/24/outline";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

export const AddOrganization: React.FC = () => {
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [tenantName, setTenantName] = useState("");
  const [tenantDomain, setTenantDomain] = useState("");
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const { uiStrings } = useUIStrings("admin_orgs");
  const [formData, setFormData] = useState({
    default_theme: "dark",
    contact_email: "",
    subscription_tier: "enterprise",
  });

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await apiClient.post("/admin/tenants", {
        name: tenantName,
        domain: tenantDomain,
        default_theme: formData.default_theme,
        contact_email: formData.contact_email,
        subscription_tier: formData.subscription_tier,
      });
      setSuccess(true); // Success step
      setStep(3);
    } catch (error) {
      alert("Failed to create tenant");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <button
        onClick={() => navigate("/admin/tenants")}
        className="flex items-center text-slate-500 hover:text-white transition-colors group mb-8"
      >
        <ArrowLeftIcon className="h-4 w-4 mr-2 group-hover:-translate-x-1 transition-transform" />
        {uiStrings?.add_back || "Back to Tenants"}
      </button>

      <div className="bg-surface-dark border border-primary shadow-2xl rounded-3xl overflow-hidden">
        {/* Stepper Header */}
        <div className="bg-surface-darker px-8 py-6 border-b border-border-color flex justify-between items-center">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <SparklesIcon className="h-8 w-8 text-primary" />
              <h1 className="text-4xl font-black text-white">
                {uiStrings?.add_title || "Onboard New Tenant"}
              </h1>
            </div>
            <p className="text-slate-400 text-lg">
              {uiStrings?.add_subtitle ||
                "Configure workspace and branding for the new tenant"}
            </p>
          </div>
          <div className="flex items-center space-x-2">
            {[1, 2, 3].map((s) => (
              <div
                key={s}
                className={`h-2 w-8 rounded-full transition-all duration-300 ${
                  step === s
                    ? "bg-primary w-12"
                    : step > s
                      ? "bg-green-500"
                      : "bg-border-color"
                }`}
              />
            ))}
          </div>
        </div>

        <div className="p-12">
          {step === 1 && (
            <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
              <div className="grid grid-cols-2 gap-8">
                <div className="space-y-2">
                  <label className="text-sm font-bold text-slate-400 uppercase tracking-widest">
                    {uiStrings?.label_org_name || "Tenant Name"}
                  </label>
                  <input
                    type="text"
                    placeholder={
                      uiStrings?.placeholder_org_name || "e.g. Acme Corp"
                    }
                    value={tenantName}
                    onChange={(e) => setTenantName(e.target.value)}
                    className="w-full bg-slate-900 border border-white/5 rounded-2xl px-6 py-4 text-white placeholder:text-slate-700 outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-bold text-slate-400 uppercase tracking-widest">
                    {uiStrings?.label_domain || "Custom Domain"}
                  </label>
                  <input
                    type="text"
                    placeholder={
                      uiStrings?.placeholder_domain || "acme.flyfleet.io"
                    }
                    value={tenantDomain}
                    onChange={(e) => setTenantDomain(e.target.value)}
                    className="w-full bg-slate-900 border border-white/5 rounded-2xl px-6 py-4 text-white placeholder:text-slate-700 outline-none focus:ring-2 focus:ring-primary/50 transition-all font-mono"
                  />
                </div>
              </div>
              <div className="flex justify-end pt-8">
                <button
                  onClick={() => setStep(2)}
                  disabled={!tenantName}
                  className="w-full bg-primary text-white py-4 rounded-2xl font-bold text-lg shadow-xl shadow-primary/20 hover:bg-primary-hover disabled:opacity-50 transition-all"
                >
                  {uiStrings?.btn_continue || "Continue to Settings"}
                </button>
              </div>
            </div>
          )}

          {step === 2 && (
            <form
              onSubmit={handleCreate}
              className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500"
            >
              <div className="space-y-6">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-4 block">
                  {uiStrings?.label_theme || "Default Workspace Theme"}
                </label>
                <div className="grid grid-cols-3 gap-4">
                  {["light", "dark", "system"].map((t) => (
                    <button
                      key={t}
                      type="button"
                      onClick={() =>
                        setFormData({ ...formData, default_theme: t })
                      }
                      className={`p-4 rounded-xl border-2 transition-all ${
                        formData.default_theme === t
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-border-color bg-surface-dark text-text-muted hover:border-border-color/60"
                      }`}
                    >
                      <span className="capitalize font-medium">{t}</span>
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex justify-between pt-8 border-t border-border-color">
                <button
                  type="button"
                  onClick={() => setStep(1)}
                  className="bg-primary hover:bg-primary-dark text-white font-medium px-6 py-3 rounded-xl transition-all border border-primary"
                >
                  Back
                </button>
                <button
                  type="submit"
                  onClick={handleCreate}
                  disabled={loading}
                  className="w-full bg-primary text-white py-4 rounded-2xl font-bold text-lg shadow-xl shadow-primary/20 hover:bg-primary-hover disabled:opacity-50 transition-all"
                >
                  {loading
                    ? uiStrings?.btn_creating || "Creating..."
                    : uiStrings?.btn_initialize || "Initialize Tenant"}
                </button>
              </div>
            </form>
          )}

          {step === 3 && (
            <div className="py-12 text-center space-y-6 animate-in zoom-in duration-500">
              <div className="max-w-2xl mx-auto text-center space-y-6">
                <div className="flex justify-center">
                  <div className="h-24 w-24 bg-primary/10 rounded-full flex items-center justify-center">
                    <CheckCircleIcon className="h-12 w-12 text-primary" />
                  </div>
                </div>
                <h1 className="text-5xl font-black text-white">
                  {uiStrings?.success_title || "Tenant Created!"}
                </h1>
                <p className="text-slate-400 text-xl">
                  {uiStrings?.success_desc ||
                    "The workspace is now ready. You can now start adding users and configuring fleet assets."}
                </p>
              </div>
              <div className="pt-8">
                <button
                  onClick={() => navigate("/admin/tenants")}
                  className="bg-primary hover:bg-primary-dark text-white px-8 py-3 rounded-xl font-semibold shadow-lg transition-all border border-primary"
                >
                  Back to Dashboard
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
