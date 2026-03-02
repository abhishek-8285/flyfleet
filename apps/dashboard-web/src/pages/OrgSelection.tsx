import { AnimatePresence, motion } from "framer-motion";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import EnterpriseAuthLayout from "../components/auth/EnterpriseAuthLayout";
import { useUIStrings } from "../hooks/useUIStrings";

interface Organization {
  id: string;
  name: string;
  role: string;
  isLocked?: boolean;
}

const OrgSelection: React.FC = () => {
  const [selectedOrgId, setSelectedOrgId] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("auth_extended");

  const [orgs, setOrgs] = React.useState<any[]>([]);
  const [loading_orgs, setLoading_orgs] = React.useState(true);

  React.useEffect(() => {
    const fetchOrgs = async () => {
      try {
        const res = await fetch(
          `{process.env.REACT_APP_API_URL || '/api/v1'}/api/v1/generic`,
        );
        const payload = await res.json();
        setOrgs(payload.data || payload || []);
      } catch (error) {
        console.error("Failed to fetch orgs", error);
      } finally {
        setLoading_orgs(false);
      }
    };
    fetchOrgs();
  }, []);

  const filteredOrgs = orgs.filter((org) =>
    org.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const handleOrgSelect = async () => {
    if (!selectedOrgId) return;
    setLoading(true);

    try {
      // API call to select organization context
      const response = await fetch("/api/v1/auth/select-organization", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ org_id: selectedOrgId }),
      });

      if (!response.ok) throw new Error("Could not switch context");

      navigate("/dashboard");
    } catch (err) {
      console.error(err);
      // Fallback for demo
      setTimeout(() => navigate("/dashboard"), 1000);
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout
      title={uiStrings?.org_title || "Select Organization"}
      subtitle={uiStrings?.org_subtitle || "Choose an active fleet context"}
    >
      <div className="space-y-6">
        <div className="relative group">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
            search
          </span>
          <input
            type="text"
            className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600"
            placeholder={
              uiStrings?.org_search_placeholder || "Search fleets..."
            }
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div className="space-y-3 max-h-[300px] overflow-y-auto pr-2 custom-scrollbar">
          <AnimatePresence>
            {filteredOrgs.map((org, index) => (
              <motion.div
                key={org.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.05 }}
              >
                <div
                  onClick={() => !org.isLocked && setSelectedOrgId(org.id)}
                  className={`group relative p-4 rounded-2xl border transition-all cursor-pointer ${
                    selectedOrgId === org.id
                      ? "bg-primary/20 border-primary shadow-[0_0_20px_rgba(59,130,246,0.1)]"
                      : "bg-slate-800/20 border-white/5 hover:bg-slate-800/40 hover:border-white/10"
                  } ${org.isLocked ? "opacity-50 cursor-not-allowed grayscale" : ""}`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div
                        className={`h-10 w-10 rounded-xl flex items-center justify-center transition-colors shadow-sm ${
                          selectedOrgId === org.id
                            ? "bg-primary text-white"
                            : "bg-slate-700/50 text-slate-400"
                        }`}
                      >
                        <span className="material-symbols-outlined">
                          domain
                        </span>
                      </div>
                      <div>
                        <h3 className="text-sm font-bold text-white tracking-tight">
                          {org.name}
                        </h3>
                        <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-0.5">
                          {org.role}
                        </p>
                      </div>
                    </div>
                    {org.isLocked ? (
                      <span className="material-symbols-outlined text-slate-600">
                        lock
                      </span>
                    ) : (
                      <div
                        className={`w-5 h-5 rounded-full border-2 flex items-center justify-center transition-all ${
                          selectedOrgId === org.id
                            ? "border-primary bg-primary"
                            : "border-slate-700"
                        }`}
                      >
                        {selectedOrgId === org.id && (
                          <span className="material-symbols-outlined text-[10px] text-white font-bold">
                            check
                          </span>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>

        <div className="pt-2">
          <button
            onClick={handleOrgSelect}
            disabled={loading || !selectedOrgId}
            className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center gap-2 text-sm uppercase tracking-widest"
          >
            {loading
              ? uiStrings?.org_switching || "Switching Context..."
              : uiStrings?.org_enter_btn || "Enter Dashboard"}
            {!loading && (
              <span className="material-symbols-outlined text-sm">
                arrow_forward
              </span>
            )}
          </button>
        </div>

        <div className="text-center pt-2">
          <p className="text-slate-500 text-[10px] font-bold uppercase tracking-widest">
            {uiStrings?.org_showing_label || "Showing"} {filteredOrgs.length}{" "}
            {uiStrings?.org_nodes_label || " Enterprise Nodes"}
          </p>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default OrgSelection;
