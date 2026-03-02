import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import { notificationService } from "../../../services/notificationService";
import { NotificationLog } from "../../../types/notification";

const NotificationDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [log, setLog] = useState<NotificationLog | null>(null);
  const [loading, setLoading] = useState(true);
  const [resending, setResending] = useState(false);
  const [copied, setCopied] = useState(false);
  const { uiStrings: rawUiStrings } = useUIStrings("notifications");
  const uiStrings = rawUiStrings || {};

  useEffect(() => {
    if (id) {
      notificationService
        .getLogDetail(id)
        .then(setLog)
        .catch((err) => console.error("Failed to fetch detail", err))
        .finally(() => setLoading(false));
    }
  }, [id]);

  const handleResend = async () => {
    if (!id || !log) return;
    setResending(true);
    try {
      const result = await notificationService.resendNotification(id, {
        reason: "Manual resend by admin",
        high_priority: false,
      });
      alert(
        uiStrings?.msg_resend_initiated?.replace(
          "{id}",
          result.new_notification_id,
        ) || `Resend initiated! New ID: ${result.new_notification_id}`,
      );
    } catch (err) {
      alert(uiStrings?.msg_resend_failed || "Failed to resend");
    } finally {
      setResending(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (loading)
    return (
      <div className="p-8 flex items-center justify-center animate-pulse">
        <div className="w-12 h-12 rounded-full border-4 border-slate-700 border-t-blue-500 animate-spin"></div>
        <span className="ml-3 text-slate-400">
          {uiStrings?.msg_loading || "Loading..."}
        </span>
      </div>
    );

  if (!log)
    return (
      <div className="p-8 text-center text-slate-500">
        {uiStrings?.msg_not_found || "Notification not found."}
      </div>
    );

  return (
    <div className="p-6 space-y-6 animate-in slide-in-from-bottom-4 duration-500">
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate("..")}
          className="flex items-center gap-2 text-slate-400 hover:text-white transition-all group"
        >
          <span className="material-symbols-outlined group-hover:-translate-x-1 transition-transform">
            arrow_back
          </span>
          {uiStrings?.link_back_to_audit || "Back to Audit"}
        </button>
        <div className="flex gap-3">
          <button
            onClick={() => copyToClipboard(JSON.stringify(log, null, 2))}
            className="px-4 py-2 rounded-xl bg-slate-800 border border-slate-700 hover:bg-slate-700 text-slate-300 transition-all flex items-center gap-2 text-sm"
          >
            <span
              className={`material-symbols-outlined text-[18px] ${copied ? "text-green-400" : ""}`}
            >
              {copied ? "check_circle" : "content_copy"}
            </span>
            {copied
              ? uiStrings?.btn_copied || "Copied"
              : uiStrings?.btn_copy_json || "Copy JSON"}
          </button>
          <button
            onClick={handleResend}
            disabled={resending}
            className="px-6 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-medium transition-all shadow-lg shadow-blue-900/40 flex items-center gap-2 disabled:opacity-50"
          >
            <span className="material-symbols-outlined text-[18px]">send</span>
            {resending
              ? uiStrings?.btn_sending || "Sending..."
              : uiStrings?.btn_issue_resend || "Issue Resend"}
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Log Details */}
        <div className="lg:col-span-2 space-y-6">
          <div className="p-6 rounded-2xl bg-slate-900/50 border border-slate-800 backdrop-blur-xl">
            <h2 className="text-xl font-bold text-slate-100 flex items-center gap-2 mb-4">
              <span className="material-symbols-outlined text-blue-400">
                info
              </span>
              {uiStrings?.title_general_info || "General Information"}
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-y-6 gap-x-8">
              <div>
                <label className="text-xs uppercase tracking-widest text-slate-500 font-bold">
                  {uiStrings?.label_correlation_id || "Correlation ID"}
                </label>
                <div className="text-slate-200 mt-1 font-mono text-sm break-all">
                  {log.correlation_id}
                </div>
              </div>
              <div>
                <label className="text-xs uppercase tracking-widest text-slate-500 font-bold">
                  {uiStrings?.label_channel_provider || "Channel & Provider"}
                </label>
                <div className="text-slate-200 mt-1 flex items-center gap-2 font-medium">
                  {log.channel} <span className="text-slate-500">via</span>{" "}
                  {log.provider || "N/A"}
                </div>
              </div>
              <div className="md:col-span-2">
                <label className="text-xs uppercase tracking-widest text-slate-500 font-bold">
                  {uiStrings?.label_subject || "Subject"}
                </label>
                <div className="text-lg font-semibold text-white mt-1">
                  {log.subject}
                </div>
              </div>
              <div className="md:col-span-2">
                <label className="text-xs uppercase tracking-widest text-slate-500 font-bold">
                  {uiStrings?.label_message_content || "Message Content"}
                </label>
                <div className="mt-2 p-4 rounded-xl bg-slate-950 border border-slate-800 text-slate-300 leading-relaxed max-h-64 overflow-y-auto whitespace-pre-wrap">
                  {log.body}
                </div>
              </div>
              {log.provider_response && (
                <div className="md:col-span-2">
                  <label className="text-xs uppercase tracking-widest text-slate-500 font-bold text-red-400">
                    {uiStrings?.label_provider_response ||
                      "Provider Response (Error)"}
                  </label>
                  <div className="mt-2 p-4 rounded-xl bg-red-900/10 border border-red-500/20 text-red-300 font-mono text-xs overflow-x-auto">
                    {log.provider_response}
                  </div>
                </div>
              )}
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-slate-900/50 border border-slate-800 backdrop-blur-xl">
            <h2 className="text-xl font-bold text-slate-100 flex items-center gap-2 mb-4">
              <span className="material-symbols-outlined text-purple-400">
                code
              </span>
              {uiStrings?.title_metadata || "Metadata Context"}
            </h2>
            <pre className="p-4 rounded-xl bg-slate-950 border border-slate-800 text-purple-300 font-mono text-sm overflow-x-auto">
              {JSON.stringify(log.metadata, null, 2)}
            </pre>
          </div>
        </div>

        {/* Right Column: Timeline & Stats */}
        <div className="space-y-6">
          <div className="p-6 rounded-2xl bg-slate-900/50 border border-slate-800 backdrop-blur-xl">
            <h2 className="text-xl font-bold text-slate-100 flex items-center gap-2 mb-6">
              <span className="material-symbols-outlined text-indigo-400">
                timeline
              </span>
              {uiStrings?.title_timeline || "Delivery Timeline"}
            </h2>
            <div className="space-y-6 relative before:absolute before:left-[11px] before:top-2 before:bottom-2 before:w-[2px] before:bg-slate-800">
              {log.items?.map((event, idx) => (
                <div key={event.id} className="relative pl-10">
                  <div
                    className={`absolute left-0 top-1 w-6 h-6 rounded-full border-4 border-slate-900 z-10 
                    ${
                      event.event_type === "SENT"
                        ? "bg-green-500"
                        : event.event_type === "FAILED"
                          ? "bg-red-500"
                          : event.event_type === "RETRY"
                            ? "bg-yellow-500"
                            : "bg-blue-500"
                    }`}
                  />
                  <div className="flex flex-col">
                    <span className="text-sm font-bold text-slate-200">
                      {event.event_type}
                    </span>
                    <span className="text-sm text-slate-400 mt-0.5">
                      {event.message}
                    </span>
                    <span className="text-[11px] text-slate-600 font-mono mt-1">
                      {new Date(event.created_at).toLocaleString()}
                    </span>
                  </div>
                </div>
              ))}
              {(!log.items || log.items.length === 0) && (
                <div className="pl-10 text-slate-600 italic text-sm">
                  {uiStrings?.msg_no_timeline || "No timeline events recorded"}
                </div>
              )}
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-slate-900/50 border border-slate-800 backdrop-blur-xl">
            <h2 className="text-xl font-bold text-slate-100 flex items-center gap-2 mb-4">
              {uiStrings?.title_stats_retries || "Stats & Retries"}
            </h2>
            <div className="space-y-4">
              <div className="flex justify-between items-center py-2 border-b border-slate-800">
                <span className="text-sm text-slate-400 uppercase tracking-wider">
                  {uiStrings?.label_retry_count || "Retry Count"}
                </span>
                <span className="text-lg font-bold text-white">
                  {log.retry_count} / {log.max_retries}
                </span>
              </div>
              <div className="flex justify-between items-center py-2 border-b border-slate-800">
                <span className="text-sm text-slate-400 uppercase tracking-wider">
                  {uiStrings?.label_priority || "Priority"}
                </span>
                <span className="px-2 py-0.5 rounded bg-blue-900/40 text-blue-400 text-xs font-bold border border-blue-500/30">
                  {log.metadata?.priority || "NORMAL"}
                </span>
              </div>
              <div className="flex justify-between items-center py-2">
                <span className="text-sm text-slate-400 uppercase tracking-wider">
                  {uiStrings?.label_retention || "Retention"}
                </span>
                <span className="text-sm text-slate-300">
                  {uiStrings?.msg_retention_days || "180 Days"}
                </span>
              </div>
            </div>

            {log.status === "FAILED" && (
              <div className="mt-6 p-4 rounded-xl bg-orange-900/10 border border-orange-500/20 flex items-start gap-3">
                <span className="material-symbols-outlined text-orange-400 text-[20px] mt-0.5">
                  report_problem
                </span>
                <p className="text-xs text-orange-200 leading-relaxed">
                  {uiStrings?.msg_exhausted ||
                    "Automatic retries have been exhausted. Manual intervention required if delivery is still necessary."}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default NotificationDetail;
