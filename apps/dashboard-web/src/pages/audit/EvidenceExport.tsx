import {
    ArrowDownTrayIcon,
    ArrowPathIcon,
    CalendarIcon,
    CheckCircleIcon,
    ClockIcon,
    CloudArrowDownIcon,
    DocumentChartBarIcon,
    ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import type { RootState } from "../../store";
import { api } from "../../services/api";

const EvidenceExport: React.FC = () => {
  const user = useSelector((state: RootState) => state.auth.user);
  const [exports, setExports] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [requesting, setRequesting] = useState(false);
  const [uiStrings, setUiStrings] = useState<any>(null);
  const [form, setForm] = useState({
    module: "ALL",
    format: "json",
    date_range_start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
      .toISOString()
      .split("T")[0],
    date_range_end: new Date().toISOString().split("T")[0],
  });

  const fetchExports = async () => {
    if (!user?.org_id) return;
    setLoading(true);
    try {
      const data = await api.getEvidenceExports(user.org_id);
      setExports(data || []);
    } catch (error) {
      console.error("Failed to fetch exports", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      try {
        const [exportsData, configRes] = await Promise.all([
          api.getEvidenceExports(user?.org_id || ""),
          fetch("/api/v1/ui/global_config").then((res) => res.json()),
        ]);
        setExports(exportsData || []);
        setUiStrings(
          configRes.ui_strings?.governance ||
            configRes.system_metadata?.ui_strings?.governance,
        );
      } catch (error) {
        console.error("Failed to load evidence data", error);
      } finally {
        setLoading(false);
      }
    };
    if (user?.org_id) {
      loadData();
    }
  }, [user?.org_id]);

  const handleRequest = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user?.org_id) return;
    setRequesting(true);
    try {
      await api.createEvidenceExport({
        org_id: user.org_id,
        module: form.module,
        date_range_start: new Date(form.date_range_start).toISOString(),
        date_range_end: new Date(form.date_range_end).toISOString(),
        format: form.format,
      });
      fetchExports();
    } catch (error) {
      console.error("Failed to request export", error);
    } finally {
      setRequesting(false);
    }
  };

  const getStatusStyles = (status: string) => {
    switch (status) {
      case "completed":
        return "text-green-400 bg-green-400/10 border-green-400/20";
      case "processing":
        return "text-blue-400 bg-blue-400/10 border-blue-400/20 animate-pulse";
      case "failed":
        return "text-red-400 bg-red-400/10 border-red-400/20";
      default:
        return "text-slate-400 bg-slate-400/10 border-slate-400/20";
    }
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
      {/* Request Form */}
      <div className="lg:col-span-1 space-y-6">
        <div className="bg-slate-900/60 backdrop-blur-xl p-8 rounded-3xl border border-white/5 shadow-2xl space-y-6 sticky top-24">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-500/20 rounded-xl">
              <DocumentChartBarIcon className="w-6 h-6 text-blue-400" />
            </div>
            <h2 className="text-xl font-semibold text-white">
              {uiStrings?.generate_archive || "Generate Archive"}
            </h2>
          </div>

          <form onSubmit={handleRequest} className="space-y-4">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-500 uppercase tracking-widest px-1">
                {uiStrings?.source_module || "Source Module"}
              </label>
              <select
                value={form.module}
                onChange={(e) => setForm({ ...form, module: e.target.value })}
                className="w-full px-4 py-3 bg-slate-950 border border-slate-800 rounded-2xl text-white appearance-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light"
              >
                <option value="ALL">
                  {uiStrings?.entire_ecosystem || "Entire Ecosystem"}
                </option>
                <option value="RBAC">
                  {uiStrings?.mod_rbac || "RBAC & Identity"}
                </option>
                <option value="FLEET">
                  {uiStrings?.mod_fleet || "Fleet Operations"}
                </option>
                <option value="BILLING">
                  {uiStrings?.mod_billing || "Financial Audit"}
                </option>
                <option value="SECURITY">
                  {uiStrings?.mod_security || "Security Events"}
                </option>
              </select>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-widest px-1 text-center block">
                  {uiStrings?.format_label || "Format"}
                </label>
                <select
                  value={form.format}
                  onChange={(e) => setForm({ ...form, format: e.target.value })}
                  className="w-full px-4 py-3 bg-slate-950 border border-slate-800 rounded-2xl text-white appearance-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light text-center"
                >
                  <option value="json">JSON</option>
                  <option value="csv">CSV</option>
                  <option value="pdf">PDF</option>
                </select>
              </div>
              <div className="flex flex-col justify-end">
                <span className="text-[10px] text-slate-500 text-center mb-2 italic">
                  {uiStrings?.crypto_signed || "Cryptographically Signed"}
                </span>
              </div>
            </div>

            <div className="space-y-4 pt-2">
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-widest px-1">
                  {uiStrings?.observation_start || "Observation Period (Start)"}
                </label>
                <div className="relative">
                  <CalendarIcon className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                  <input
                    type="date"
                    value={form.date_range_start}
                    onChange={(e) =>
                      setForm({ ...form, date_range_start: e.target.value })
                    }
                    className="w-full pl-12 pr-4 py-3 bg-slate-950 border border-slate-800 rounded-2xl text-white focus:ring-2 focus:ring-blue-500/50 transition-all font-light"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-widest px-1">
                  {uiStrings?.observation_end || "Observation Period (End)"}
                </label>
                <div className="relative">
                  <CalendarIcon className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                  <input
                    type="date"
                    value={form.date_range_end}
                    onChange={(e) =>
                      setForm({ ...form, date_range_end: e.target.value })
                    }
                    className="w-full pl-12 pr-4 py-3 bg-slate-950 border border-slate-800 rounded-2xl text-white focus:ring-2 focus:ring-blue-500/50 transition-all font-light"
                  />
                </div>
              </div>
            </div>

            <button
              type="submit"
              disabled={requesting}
              className="w-full py-4 mt-4 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl font-semibold transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
            >
              {requesting ? (
                <ArrowPathIcon className="w-5 h-5 animate-spin" />
              ) : (
                <CloudArrowDownIcon className="w-5 h-5" />
              )}
              {uiStrings?.initiate_btn || "Initiate Extraction"}
            </button>
          </form>

          <div className="p-4 bg-slate-950/50 rounded-2xl border border-white/5 space-y-2">
            <p className="text-[10px] font-bold text-slate-600 uppercase tracking-widest">
              {uiStrings?.compliance_note || "Compliance Note"}
            </p>
            <p className="text-[10px] text-slate-500 leading-relaxed">
              Exports are immutable and retainable for 7 years per regulatory
              standards. Large datasets may take up to 15 minutes to generate.
            </p>
          </div>
        </div>
      </div>

      {/* History List */}
      <div className="lg:col-span-2 space-y-6">
        <div className="flex items-center justify-between mb-4 px-2">
          <h2 className="text-xl font-semibold text-white">
            {uiStrings?.extraction_repo || "Extraction Repository"}
          </h2>
          <button
            onClick={fetchExports}
            className="p-2 text-slate-400 hover:text-white transition-all"
          >
            <ArrowPathIcon
              className={`w-5 h-5 ${loading ? "animate-spin" : ""}`}
            />
          </button>
        </div>

        <div className="space-y-4">
          {exports.length === 0 && !loading ? (
            <div className="py-20 text-center bg-slate-900/20 rounded-3xl border border-dashed border-white/5">
              <ArrowDownTrayIcon className="w-12 h-12 text-slate-700 mx-auto mb-4" />
              <p className="text-slate-500 font-light">
                {uiStrings?.no_archives ||
                  "No archive extractions recorded yet."}
              </p>
            </div>
          ) : (
            exports.map((exp) => (
              <div
                key={exp.id}
                className="bg-slate-900/40 backdrop-blur-xl p-6 rounded-3xl border border-white/5 hover:border-white/10 transition-all group flex flex-col md:flex-row md:items-center justify-between gap-6 shadow-xl"
              >
                <div className="flex items-start gap-4">
                  <div
                    className={`p-3 rounded-2xl border ${getStatusStyles(exp.status)}`}
                  >
                    {exp.status === "completed" ? (
                      <CheckCircleIcon className="w-6 h-6" />
                    ) : exp.status === "processing" ? (
                      <ClockIcon className="w-6 h-6" />
                    ) : (
                      <ExclamationCircleIcon className="w-6 h-6" />
                    )}
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <h3 className="text-white font-medium">
                        Compliance Archive - {exp.module}
                      </h3>
                      <span className="text-[10px] font-bold text-blue-500 uppercase tracking-widest">
                        {exp.format}
                      </span>
                    </div>
                    <p className="text-xs text-slate-400">
                      Range:{" "}
                      {new Date(exp.date_range_start).toLocaleDateString()} -{" "}
                      {new Date(exp.date_range_end).toLocaleDateString()}
                    </p>
                    <p className="text-[10px] text-slate-600 font-mono">
                      ID: {exp.id}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-3 self-end md:self-auto">
                  {exp.status === "completed" ? (
                    <a
                      href={exp.file_url}
                      download
                      className="px-6 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-sm font-medium transition-all border border-slate-700 hover:border-blue-500/50 flex items-center gap-2"
                    >
                      <CloudArrowDownIcon className="w-4 h-4" />
                      {uiStrings?.download_btn || "Download"}
                    </a>
                  ) : exp.status === "failed" ? (
                    <div className="text-xs text-red-400 bg-red-400/5 px-3 py-1.5 rounded-lg border border-red-400/10 italic">
                      {exp.error_message ||
                        "Unexpected failure during extraction"}
                    </div>
                  ) : (
                    <span className="text-xs text-blue-400 bg-blue-400/5 px-4 py-2 rounded-xl border border-blue-400/10 flex items-center gap-2">
                      <ArrowPathIcon className="w-4 h-4 animate-spin" />
                      {uiStrings?.assembling_label || "Assembling Data..."}
                    </span>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default EvidenceExport;
