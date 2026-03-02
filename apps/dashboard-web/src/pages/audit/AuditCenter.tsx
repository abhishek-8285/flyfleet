import React, { useState } from "react";

const AuditCenter: React.FC = () => {
  const [activeTab, setActiveTab] = useState("logs");

  const [tabs, setTabs] = React.useState<any[]>([]);
  const [loading_tabs, setLoading_tabs] = React.useState(true);
  const [uiStrings, setUiStrings] = React.useState<any>(null);

  React.useEffect(() => {
    const fetchConfig = async () => {
      try {
        const [resTabs, resConfig] = await Promise.all([
          fetch(
            `${process.env.REACT_APP_API_URL || "/api/v1"}/audit/auditcenter`,
          ),
          fetch("/api/v1/ui/global_config"),
        ]);
        const [payloadTabs, payloadConfig] = await Promise.all([
          resTabs.json(),
          resConfig.json(),
        ]);
        setTabs(payloadTabs.data || payloadTabs || []);
        setUiStrings(
          payloadConfig.ui_strings?.governance ||
            payloadConfig.system_metadata?.ui_strings?.governance,
        );
      } catch (error) {
        console.error("Failed to fetch audit center data", error);
      } finally {
        setLoading_tabs(false);
      }
    };
    fetchConfig();
  }, []);

  return (
    <div className="min-h-screen p-8 space-y-8 animate-in fade-in duration-700">
      {/* Massive Header */}
      <div className="relative py-12 px-8 rounded-[40px] bg-slate-900 border border-white/5 overflow-hidden shadow-2xl">
        {/* Background Accents */}
        <div className="absolute top-0 right-0 w-96 h-96 bg-blue-500/10 blur-[120px] rounded-full -translate-y-1/2 translate-x-1/3" />
        <div className="absolute bottom-0 left-0 w-72 h-72 bg-purple-500/10 blur-[100px] rounded-full translate-y-1/2 -translate-x-1/3" />

        <div className="relative z-10 flex flex-col md:flex-row md:items-end justify-between gap-8">
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <span className="px-3 py-1 bg-blue-500/10 text-blue-400 text-[10px] font-bold tracking-[0.2em] uppercase rounded-full border border-blue-500/20">
                {uiStrings?.compliance_label_badge || "Enterprise Compliance"}
              </span>
            </div>
            <h1 className="text-5xl md:text-6xl font-extralight tracking-tight text-white leading-none">
              {uiStrings?.audit_title || "Audit"}{" "}
              <span className="font-semibold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-indigo-400">
                {uiStrings?.intel_label || "Intelligence"}
              </span>
            </h1>
            <p className="text-slate-400 text-lg font-light max-w-xl">
              {uiStrings?.compliance_desc ||
                "A centralized command for absolute transparency, regulatory adherence, and cryptographically verified system traceability."}
            </p>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex flex-wrap items-center gap-2 p-1.5 bg-slate-900/60 backdrop-blur-xl border border-white/5 rounded-[24px] w-fit shadow-lg">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2.5 px-6 py-3 rounded-[18px] text-sm font-medium transition-all duration-300 ${
                isActive
                  ? "bg-blue-600 text-white shadow-lg shadow-blue-500/40"
                  : "text-slate-400 hover:text-white hover:bg-white/5"
              }`}
            >
              <Icon
                className={`w-5 h-5 ${isActive ? "scale-110" : "opacity-60"}`}
              />
              {tab.name}
            </button>
          );
        })}
      </div>

      {/* Main Surface */}
      <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
        {tabs.map((tab) => {
          const Component = tab.component;
          return activeTab === tab.id ? <Component key={tab.id} /> : null;
        })}
      </div>
    </div>
  );
};

export default AuditCenter;
