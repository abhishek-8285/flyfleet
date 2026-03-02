import branding from "@flyfleet/config";
import React from "react";
import { useUIStrings } from "../hooks/useUIStrings";

const Reports: React.FC = () => {
  const [reportCategories, setReportCategories] = React.useState<any[]>([]);
  const [systemMeta, setSystemMeta] = React.useState<any>(null);
  const { uiStrings } = useUIStrings("reports");

  React.useEffect(() => {
    const fetchReports = async () => {
      try {
        const res = await fetch(
          `${process.env.REACT_APP_API_URL || "/api/v1"}/reports/categories`,
        );
        if (res.ok) {
          const data = await res.json();
          setReportCategories(data.data || []);
        }
      } catch (err) {
        console.error("Failed to fetch abstract report categories", err);
      }
    };
    fetchReports();
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) => {
        setSystemMeta(data.system_metadata);
      })
      .catch((err) => console.error("Failed to fetch system config", err));
  }, []);

  return (
    <div className="flex-1 overflow-y-auto p-6 custom-scrollbar bg-background-dark">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-[1600px] mx-auto">
        {reportCategories.map((report, idx) => (
          <div
            key={idx}
            className="bg-surface-dark p-6 rounded-xl border border-border-color hover:border-primary/50 transition-all group cursor-pointer shadow-lg hover:shadow-primary/5 shadow-black/20"
          >
            <div className="flex items-center gap-4 mb-4">
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary group-hover:bg-primary group-hover:text-white transition-all transform group-hover:scale-110 duration-300">
                <span className="material-symbols-outlined text-2xl">
                  {report.icon}
                </span>
              </div>
              <h3 className="text-white font-bold tracking-tight">
                {report.title}
              </h3>
            </div>
            <p className="text-sm text-text-muted mb-6 leading-relaxed">
              {report.desc}
            </p>
            <button className="text-xs font-semibold text-primary flex items-center gap-1 group-hover:gap-2 transition-all">
              {uiStrings?.generate_btn || "Generate Report"}
              <span className="material-symbols-outlined text-sm">
                arrow_forward
              </span>
            </button>
          </div>
        ))}
      </div>
      <footer className="mt-12 text-center text-xs text-[#64748b] pb-6 flex flex-col items-center gap-1">
        <span>
          © {new Date().getFullYear()} {branding.companyName} Enterprise. All
          rights reserved.
        </span>
        <span className="text-[10px]">
          {systemMeta?.version || "v--"} | Server Time:{" "}
          <span className="font-mono">{systemMeta?.server_time || "---"}</span>
        </span>
      </footer>
    </div>
  );
};

export default Reports;
