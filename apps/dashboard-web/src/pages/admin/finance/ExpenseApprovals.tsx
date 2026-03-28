import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";
import { useAuth } from "../../../context/AuthProvider";

interface ExpenseReport {
  id: string;
  submitter_id: string;
  total_amount: number;
  currency: string;
  status: string;
  created_at: string;
  metadata?: Record<string, any>;
}

const ExpenseApprovals: React.FC = () => {
  const [reports, setReports] = useState<ExpenseReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { uiStrings } = useUIStrings("finance");
  const { user } = useAuth();

  const loadReports = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = (await apiClient.get("/admin/finance/expenses/reports", {
        params: { limit: 100 },
      })) as any;
      const list: ExpenseReport[] = Array.isArray(response?.reports)
        ? response.reports
        : Array.isArray(response?.data)
          ? response.data
          : [];
      setReports(list);
    } catch (err: any) {
      console.error("Failed to load expense reports", err);
      setError(err?.message || "Failed to load expense reports");
      setReports([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadReports();
  }, []);

  const submitAction = async (
    reportId: string,
    action: "approve" | "reject",
  ) => {
    try {
      // Get authenticated user ID from auth context
      const approverId = user?.user_id || user?.id;
      if (!approverId) {
        throw new Error("User not authenticated");
      }
      await apiClient.post(
        `/admin/finance/expenses/reports/${reportId}/${action}`,
        {
          approver_id: approverId,
          reason: action === "reject" ? "Rejected by admin" : undefined,
        },
      );
      await loadReports();
    } catch (err: any) {
      setError(err?.message || `Failed to ${action} report`);
    }
  };

  return (
    <div className="p-6">
      <header className="mb-6">
        <h1 className="text-2xl font-bold text-white">
          {uiStrings?.card_expenses || "Expense Approvals"}
        </h1>
        <p className="text-text-muted">
          {uiStrings?.subtitle || "Review and approve driver expense reports."}
        </p>
      </header>

      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        {error && (
          <div className="px-4 py-3 text-sm text-red-300 border-b border-red-500/30 bg-red-500/10">
            {error}
          </div>
        )}
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-surface-darker border-b border-border-color text-text-muted text-xs uppercase tracking-wider">
              <th className="p-4 font-semibold">
                {uiStrings?.th_report_id || "Report ID"}
              </th>
              <th className="p-4 font-semibold">
                {uiStrings?.th_user || "Submitter"}
              </th>
              <th className="p-4 font-semibold">
                {uiStrings?.th_category || "Category"}
              </th>
              <th className="p-4 font-semibold">
                {uiStrings?.th_date || "Date"}
              </th>
              <th className="p-4 font-semibold">
                {uiStrings?.th_amount || "Amount"}
              </th>
              <th className="p-4 font-semibold">
                {uiStrings?.th_status || "Status"}
              </th>
              <th className="p-4 font-semibold text-right">
                {uiStrings?.th_actions || "Actions"}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-color">
            {loading && (
              <tr>
                <td className="p-4 text-sm text-text-muted" colSpan={7}>
                  {uiStrings?.loading || "Loading reports..."}
                </td>
              </tr>
            )}
            {!loading && reports.length === 0 && (
              <tr>
                <td className="p-4 text-sm text-text-muted" colSpan={7}>
                  {uiStrings?.msg_no_expenses ||
                    "No expense reports available."}
                </td>
              </tr>
            )}
            {!loading &&
              reports.map((report) => (
                <tr
                  key={report.id}
                  className="hover:bg-surface-darker/50 transition-colors group"
                >
                  <td className="p-4 text-sm font-medium text-white">
                    {report.id}
                  </td>
                  <td className="p-4 text-sm text-text-muted">
                    {report.submitter_id}
                  </td>
                  <td className="p-4 text-sm text-text-muted">
                    <span className="px-2 py-1 rounded bg-surface-darker border border-border-color text-xs">
                      {report.metadata?.category || "GENERAL"}
                    </span>
                  </td>
                  <td className="p-4 text-sm text-text-muted">
                    {new Date(report.created_at).toLocaleDateString()}
                  </td>
                  <td className="p-4 text-sm font-bold text-white">
                    {report.currency}{" "}
                    {Number(report.total_amount || 0).toFixed(2)}
                  </td>
                  <td className="p-4">
                    <span
                      className={`px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wide 
                                        ${
                                          report.status === "APPROVED"
                                            ? "bg-green-500/10 text-green-500"
                                            : report.status === "SUBMITTED"
                                              ? "bg-yellow-500/10 text-yellow-500"
                                              : "bg-red-500/10 text-red-500"
                                        }`}
                    >
                      {report.status}
                    </span>
                  </td>
                  <td className="p-4 text-right">
                    {report.status === "SUBMITTED" && (
                      <div className="flex justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                          onClick={() => submitAction(report.id, "approve")}
                          className="p-1.5 rounded bg-green-500/10 text-green-500 hover:bg-green-500/20 transition-colors"
                          title={uiStrings?.btn_approve || "Approve"}
                        >
                          <span className="material-symbols-outlined text-[18px]">
                            check
                          </span>
                        </button>
                        <button
                          onClick={() => submitAction(report.id, "reject")}
                          className="p-1.5 rounded bg-red-500/10 text-red-500 hover:bg-red-500/20 transition-colors"
                          title={uiStrings?.btn_reject || "Reject"}
                        >
                          <span className="material-symbols-outlined text-[18px]">
                            close
                          </span>
                        </button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default ExpenseApprovals;
