import {
    GlobeAltIcon,
    PlusIcon,
    ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import SecurityFooter from "../../../components/SecurityFooter";
import SecuritySidebar from "../../../components/SecuritySidebar";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

const IPControls: React.FC = () => {
  const { uiStrings: rawUiStrings } = useUIStrings("admin_security");
  const uiStrings = rawUiStrings || {};
  const [rules, setRules] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const orgId = localStorage.getItem("org_id") || "default";

  useEffect(() => {
    fetchRules();
  }, [orgId]);

  const fetchRules = async () => {
    try {
      setLoading(true);
      const response = (await apiClient.get("/security/ip-rules", {
        params: { org_id: orgId },
      })) as any;
      const data = response?.data ?? response;
      setRules(Array.isArray(data) ? data : data?.items || []);
    } catch (error) {
      console.error("Failed to fetch IP rules", error);
      setRules([]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex h-screen bg-[#0b1117] text-slate-200 overflow-hidden font-inter">
      <SecuritySidebar />

      <main className="flex-1 flex flex-col min-w-0 bg-grid-pattern relative">
        <div className="glow-effect" />

        <header className="px-8 py-6 border-b border-white/5 bg-slate-900/40 backdrop-blur-md z-10 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-white flex items-center gap-3">
              <span className="material-symbols-outlined text-primary text-3xl">
                public
              </span>
              {uiStrings.title_ip || "IP Access Governance"}
            </h2>
            <p className="text-slate-400 text-sm mt-1">
              {uiStrings.subtitle_ip ||
                "Granular network control list and geofencing configuration."}
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-hover text-white rounded-lg font-bold text-sm transition-all shadow-lg shadow-primary/20">
              <PlusIcon className="w-5 h-5" />
              {uiStrings.btn_new_rule || "NEW SECURITY RULE"}
            </button>
          </div>
        </header>

        <div className="p-8 z-10 flex flex-col gap-6 overflow-auto custom-scrollbar flex-1">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 glass-card rounded-2xl flex flex-col overflow-hidden">
              <div className="p-4 border-b border-white/5 bg-white/5 flex items-center justify-between">
                <h3 className="text-sm font-bold text-white flex items-center gap-2">
                  <ShieldCheckIcon className="w-5 h-5 text-primary" />
                  {uiStrings.table_title || "Active Network Rules"}
                </h3>
                <span className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">
                  {uiStrings.table_sync_note ||
                    "Auto-synced with Fleet Firewall nodes"}
                </span>
              </div>

              <div className="flex-1 overflow-auto">
                <table className="w-full text-left">
                  <thead>
                    <tr className="border-b border-white/5 bg-white/[0.02]">
                      <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                        {uiStrings.th_rule_id || "Rule Identification"}
                      </th>
                      <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                        {uiStrings.th_network || "Network / CIDR"}
                      </th>
                      <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                        {uiStrings.th_action || "Action"}
                      </th>
                      <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                        {uiStrings.th_context || "Context"}
                      </th>
                      <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500 text-right">
                        {uiStrings.th_management || "Management"}
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {loading ? (
                      <tr>
                        <td
                          colSpan={5}
                          className="px-6 py-10 text-center text-slate-500 italic"
                        >
                          {uiStrings.msg_loading_rules ||
                            "Reading firewall rule tables..."}
                        </td>
                      </tr>
                    ) : rules.length === 0 ? (
                      <tr>
                        <td
                          colSpan={5}
                          className="px-6 py-10 text-center text-slate-500 italic text-sm"
                        >
                          {uiStrings.msg_no_rules ||
                            "No custom IP rules defined."}
                        </td>
                      </tr>
                    ) : (
                      rules.map((rule) => (
                        <tr
                          key={rule.id || rule._id}
                          className="hover:bg-white/[0.02] transition-colors"
                        >
                          <td className="px-6 py-4 font-bold text-white text-sm">
                            {rule.name ||
                              uiStrings.rule_unnamed ||
                              "Unnamed Rule"}
                          </td>
                          <td className="px-6 py-4 font-mono text-xs text-primary">
                            {rule.cidr || rule.network}
                          </td>
                          <td className="px-6 py-4">
                            {rule.action === "allow" ? (
                              <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 text-[10px] font-black border border-emerald-500/20">
                                {uiStrings.badge_allow || "ALLOW"}
                              </span>
                            ) : (
                              <span className="px-2 py-0.5 rounded bg-vibrant-red/10 text-vibrant-red text-[10px] font-black border border-vibrant-red/20">
                                {uiStrings.badge_block || "BLOCK"}
                              </span>
                            )}
                          </td>
                          <td className="px-6 py-4 text-xs text-slate-400">
                            {rule.description || "--"}
                          </td>
                          <td className="px-6 py-4 text-right">
                            <button className="text-[10px] font-black text-slate-500 hover:text-white uppercase tracking-widest transition-colors">
                              {uiStrings.btn_modify || "Modify"}
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          {/* Security Advice */}
          <div className="grid grid-cols-2 gap-6">
            <div className="glass-card p-6 rounded-2xl bg-blue-500/5">
              <h4 className="flex items-center gap-2 text-white font-bold">
                <GlobeAltIcon className="w-5 h-5 text-blue-400" />
                {uiStrings.title_cidr_practices || "CIDR Best Practices"}
              </h4>
              <p className="text-sm text-slate-400 mt-2">
                {uiStrings.desc_cidr_practices ||
                  "Avoid using broad /16 ranges unless necessary. Use /32 for single hosts and /24 for local networks to minimize the attack surface."}
              </p>
            </div>
            <div className="glass-card p-6 rounded-2xl bg-emerald-500/5">
              <h4 className="flex items-center gap-2 text-white font-bold">
                <ShieldCheckIcon className="w-5 h-5 text-emerald-400" />
                {uiStrings.title_auto_prop || "Auto-Propagation"}
              </h4>
              <p className="text-sm text-slate-400 mt-2">
                {uiStrings.desc_auto_prop ||
                  "Rules registered here are distributed to all edge proxies (BFF Layer) within 30 seconds for immediate enforcement."}
              </p>
            </div>
          </div>
        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default IPControls;
