import React, { useEffect, useState } from "react";
import { DataGrid } from "../../../components/common/DataGrid";
import { useUIStrings } from "../../../hooks/useUIStrings";

interface AuditLog {
  id: string;
  user_id: string;
  action: string;
  resource: string;
  status: string;
  ip_address: string;
  created_at: string;
  details: string;
}

export const SecurityLogs: React.FC = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const { uiStrings } = useUIStrings("admin_rbac_logs");

  const [columns, setColumns] = useState<any[]>([]);
  useEffect(() => {
    fetch("/api/v1/ui/settings/columns")
      .then((res) => res.json())
      .then((data) => setColumns(data))
      .catch((err) => console.error(err));
  }, []);

  return (
    <div className="p-8">
      <div className="mb-6 flex justify-between items-end">
        <div>
          <h1 className="text-2xl font-bold text-slate-800 flex items-center">
            <span className="material-symbols-outlined text-blue-600 text-[32px] mr-3">
              shield_exclamation
            </span>
            {uiStrings?.title || "Security Audit Logs"}
          </h1>
          <p className="text-slate-500 mt-1">
            {uiStrings?.subtitle ||
              "Immutable record of all security-sensitive operations"}
          </p>
        </div>
        <div className="flex space-x-2">
          <button
            className="bg-white border border-slate-200 text-slate-600 px-4 py-2 rounded-lg text-sm font-semibold hover:bg-slate-50 transition-all shadow-sm"
            onClick={() => {
              /* Export logic */
            }}
          >
            {uiStrings?.btn_export || "Export CSV"}
          </button>
        </div>
      </div>

      <DataGrid<AuditLog>
        title={uiStrings?.grid_title || "Audit Trail"}
        endpoint="/admin/audit/logs"
        columns={columns}
        refreshTrigger={refreshTrigger}
      />
    </div>
  );
};
