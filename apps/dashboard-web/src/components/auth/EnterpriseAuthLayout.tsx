import branding from "@flyfleet/config";
import { motion } from "framer-motion";
import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { api } from "../../services/api";
import Modal from "../common/Modal";

interface EnterpriseAuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle?: string;
}

interface LegalDocSection {
  title: string;
  items?: string[];
  content?: string;
}

interface LegalDocs {
  privacy_policy: {
    last_updated: string;
    sections: LegalDocSection[];
  };
  terms_of_service: {
    effective_date: string;
    sections: LegalDocSection[];
  };
}

const EnterpriseAuthLayout: React.FC<EnterpriseAuthLayoutProps> = ({
  children,
  title,
  subtitle,
}) => {
  const { uiStrings } = useUIStrings("auth_layout");
  const [activeModal, setActiveModal] = useState<
    "privacy" | "terms" | "support" | null
  >(null);
  const [legalDocs, setLegalDocs] = useState<LegalDocs | null>(null);

  useEffect(() => {
    const fetchLegal = async () => {
      try {
        const res = (await api.get("/ui/misc/legal_docs")) as any;
        setLegalDocs(res.data);
      } catch (err) {
        console.error("Failed to fetch legal docs", err);
      }
    };
    fetchLegal();
  }, []);

  return (
    <>
      <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 selection:bg-primary/30">
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/10 rounded-full blur-[120px] animate-pulse"></div>
          <div
            className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse"
            style={{ animationDelay: "2s" }}
          ></div>
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: "easeOut" }}
          className="max-w-md w-full relative z-10"
        >
          <div className="bg-slate-900/40 backdrop-blur-2xl border border-white/5 p-6 rounded-3xl shadow-[0_20px_50px_rgba(0,0,0,0.3)] border-t-white/10">
            <div className="text-center mb-6">
              <div className="flex justify-center mb-4">
                <div className="h-12 w-12 bg-gradient-to-br from-primary to-blue-600 rounded-2xl flex items-center justify-center shadow-lg shadow-primary/20 rotate-3 hover:rotate-0 transition-transform duration-300">
                  <span className="material-symbols-outlined text-white text-2xl">
                    local_shipping
                  </span>
                </div>
              </div>
              <h1 className="text-2xl font-bold text-white tracking-tight mb-1">
                {branding.companyName}
              </h1>
              <div className="h-1 w-10 bg-primary mx-auto rounded-full mb-3"></div>
              <h2 className="text-lg font-semibold text-slate-200 mb-1">
                {title}
              </h2>
              {subtitle && (
                <p className="text-slate-400 text-sm font-medium">{subtitle}</p>
              )}
            </div>

            {children}

            <div className="mt-6 pt-4 border-t border-slate-800/50 text-center">
              <p className="text-slate-500 text-xs tracking-wide uppercase font-semibold mb-3">
                {uiStrings?.gateway_label || "Secure Enterprise Gateway"}
              </p>
              <div className="flex justify-center gap-4 text-xs text-slate-500">
                <button
                  className="hover:text-primary transition-colors"
                  onClick={() => setActiveModal("privacy")}
                >
                  {uiStrings?.label_privacy || "Privacy"}
                </button>
                <span className="text-slate-800">•</span>
                <button
                  className="hover:text-primary transition-colors"
                  onClick={() => setActiveModal("terms")}
                >
                  {uiStrings?.label_terms || "Terms"}
                </button>
                <span className="text-slate-800">•</span>
                <button
                  className="hover:text-primary transition-colors"
                  onClick={() => setActiveModal("support")}
                >
                  {uiStrings?.label_support || "Support"}
                </button>
              </div>
            </div>
          </div>

          <p className="text-center mt-6 text-slate-600 text-[10px] uppercase tracking-[0.2em] font-bold">
            {uiStrings?.powered_by || "Powered by FleetCommand Core v4.0"}
          </p>
        </motion.div>
      </div>

      {/* Privacy Modal */}
      <Modal
        isOpen={activeModal === "privacy"}
        onClose={() => setActiveModal(null)}
        title={uiStrings?.title_privacy || "Privacy Policy"}
        size="lg"
      >
        <div className="prose prose-invert max-w-none space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">
                policy
              </span>
              {uiStrings?.last_updated || "Last Updated:"}{" "}
              {legalDocs?.privacy_policy.last_updated || "Loading..."}
            </h4>
            <p className="text-slate-400 text-sm">
              This Privacy Policy describes how FlyFleet Global Operations
              collects, uses, and protects your personal information.
            </p>
          </div>

          <div className="space-y-8">
            {legalDocs?.privacy_policy.sections.map((section, idx) => (
              <div key={idx}>
                <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">
                  {section.title}
                </h4>
                {section.content && (
                  <p className="text-slate-400 text-sm">{section.content}</p>
                )}
                {section.items && (
                  <ul className="text-slate-400 space-y-1 text-sm ml-4 mt-2">
                    {section.items.map((item, i) => (
                      <li key={i} className="list-disc">
                        {item}
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            ))}
          </div>
        </div>
      </Modal>

      {/* Terms Modal */}
      <Modal
        isOpen={activeModal === "terms"}
        onClose={() => setActiveModal(null)}
        title={uiStrings?.title_terms || "Terms of Service"}
        size="lg"
      >
        <div className="prose prose-invert max-w-none space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">
                description
              </span>
              {uiStrings?.effective_date || "Effective Date:"}{" "}
              {legalDocs?.terms_of_service.effective_date || "Loading..."}
            </h4>
            <p className="text-slate-400 text-sm">
              These Terms of Service govern your access to and use of FlyFleet's
              fleet management platform and services.
            </p>
          </div>

          <div className="space-y-8">
            {legalDocs?.terms_of_service.sections.map((section, idx) => (
              <div key={idx}>
                <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">
                  {section.title}
                </h4>
                {section.content && (
                  <p className="text-slate-400 text-sm">{section.content}</p>
                )}
                {section.items && (
                  <ul className="text-slate-400 space-y-1 text-sm ml-4 mt-2">
                    {section.items.map((item, i) => (
                      <li key={i} className="list-disc">
                        {item}
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            ))}
          </div>
        </div>
      </Modal>

      {/* Support Modal (Keep as is since it's mostly links/actions) */}
      <Modal
        isOpen={activeModal === "support"}
        onClose={() => setActiveModal(null)}
        title={uiStrings?.title_support || "Support Center"}
        size="md"
      >
        <div className="space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">
                support_agent
              </span>
              {uiStrings?.support_247 || "24/7 Enterprise Support"}
            </h4>
            <p className="text-slate-400 text-sm">
              {uiStrings?.support_desc ||
                "Our dedicated support team is available around the clock to assist with any technical issues or questions."}
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4">
            <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4 hover:bg-slate-800/50 transition-all cursor-pointer group">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-blue-400 text-xl group-hover:scale-110 transition-transform">
                  phone
                </span>
                <div>
                  <h5 className="text-white font-semibold">
                    {uiStrings?.btn_phone || "Phone Support"}
                  </h5>
                  <p className="text-slate-400 text-sm">+1 (555) FLEET-HELP</p>
                </div>
              </div>
            </div>

            <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4 hover:bg-slate-800/50 transition-all cursor-pointer group">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-green-400 text-xl group-hover:scale-110 transition-transform">
                  mail
                </span>
                <div>
                  <h5 className="text-white font-semibold">
                    {uiStrings?.btn_email || "Email Support"}
                  </h5>
                  <p className="text-slate-400 text-sm">support@flyfleet.com</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Modal>
    </>
  );
};

export default EnterpriseAuthLayout;
