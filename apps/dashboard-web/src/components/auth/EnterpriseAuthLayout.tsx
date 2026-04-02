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
      <div className="min-h-screen flex flex-col md:flex-row bg-slate-50 md:bg-white dark:bg-[#121212] selection:bg-blue-600/30">
        {/* Top/Left Background */}
        <div className="flex flex-col w-full h-[40vh] md:h-auto md:w-1/2 relative bg-[#1e3a8a] overflow-hidden items-center md:justify-center">
          <div className="absolute inset-0">
            <img 
              src="https://images.unsplash.com/photo-1601584115197-04ecc0da31d7?auto=format&fit=crop&q=80&w=2000" 
              className="object-cover w-full h-full mix-blend-overlay opacity-80" 
              alt="Fleet Highway Background" 
            />
          </div>
          <div className="mt-[12vh] md:mt-0 relative z-10 flex items-center gap-2 md:gap-3 p-3 md:p-6 drop-shadow-lg">
            <span className="material-symbols-outlined text-white text-[36px] md:text-[56px] drop-shadow-md">local_shipping</span>
            <span className="text-white text-3xl md:text-5xl font-bold tracking-tight drop-shadow-md">{branding.companyName}</span>
          </div>
        </div>

        {/* Bottom/Right Form Panel */}
        <div className="w-full md:w-1/2 flex flex-col justify-start md:justify-center items-center px-4 md:p-8 lg:p-16 relative bg-slate-50 dark:bg-[#121212] md:bg-white md:dark:bg-[#121212] pt-0 md:pt-0 min-h-[60vh] md:min-h-0">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: "easeOut" }}
            className="w-full max-w-sm bg-white dark:bg-[#1c1c1e] rounded-2xl shadow-xl dark:shadow-2xl md:shadow-none md:bg-transparent md:dark:bg-transparent border border-transparent dark:border-white/5 md:dark:border-transparent p-6 md:p-0 -mt-16 md:mt-0 relative z-20"
          >
            <div className="mb-6 md:mb-10 text-center md:text-left">
              <h1 className="text-2xl md:text-3xl font-bold md:font-medium text-slate-900 dark:text-white mb-1">{title}</h1>
              {subtitle && <p className="text-slate-500 dark:text-slate-400 text-sm mt-1">{subtitle}</p>}
            </div>

            {children}
          </motion.div>
          
          {/* Footer */}
          <div className="mt-8 md:mt-16 text-center text-slate-500 pb-8 md:pb-0 w-full md:max-w-sm relative z-20">
            <div className="flex justify-center gap-3 text-[11px] md:text-xs font-semibold pt-1">
              <button onClick={() => setActiveModal("privacy")} className="hover:text-blue-600 transition-colors">Privacy</button>
              <span className="text-slate-300">•</span>
              <button onClick={() => setActiveModal("terms")} className="hover:text-blue-600 transition-colors">Terms</button>
              <span className="text-slate-300">•</span>
              <button onClick={() => setActiveModal("support")} className="hover:text-blue-600 transition-colors">Support</button>
            </div>
          </div>
        </div>
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
              {legalDocs?.privacy_policy?.last_updated || "Loading..."}
            </h4>
            <p className="text-slate-400 text-sm">
              This Privacy Policy describes how FlyFleet Global Operations
              collects, uses, and protects your personal information.
            </p>
          </div>

          <div className="space-y-8">
            {legalDocs?.privacy_policy?.sections.map((section, idx) => (
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
              {legalDocs?.terms_of_service?.effective_date || "Loading..."}
            </h4>
            <p className="text-slate-400 text-sm">
              These Terms of Service govern your access to and use of FlyFleet's
              fleet management platform and services.
            </p>
          </div>

          <div className="space-y-8">
            {legalDocs?.terms_of_service?.sections.map((section, idx) => (
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
              {uiStrings?.support_247 || "24/7 Operations Support"}
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
