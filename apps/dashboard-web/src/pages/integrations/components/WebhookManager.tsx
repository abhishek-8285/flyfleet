import {
    CheckCircleIcon,
    GlobeAltIcon,
    PencilSquareIcon,
    PlusIcon,
    SignalIcon,
    TrashIcon,
} from "@heroicons/react/24/outline";
import React from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";

const WebhookManager: React.FC = () => {
  const [webhooks, setWebhooks] = React.useState<any[]>([]);
  const [loading_webhooks, setLoading_webhooks] = React.useState(true);
  const [systemMeta, setSystemMeta] = React.useState<any>(null);
  const [webhookMeta, setWebhookMeta] = React.useState<any>(null);
  const { uiStrings } = useUIStrings("integrations");

  React.useEffect(() => {
    const fetchWebhooks = async () => {
      try {
        const res = await fetch(
          `{import.meta.env.VITE_API_URL || '/api/v1'}/api/v1/integrations/webhookmanager`,
        );
        const payload = await res.json();
        setWebhooks(payload.data || payload || []);
      } catch (error) {
        console.error("Failed to fetch webhooks", error);
      } finally {
        setLoading_webhooks(false);
      }
    };
    fetchWebhooks();
  }, []);

  return (
    <div className="space-y-6">
      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-slate-900/40 border border-slate-800 p-6 rounded-2xl flex items-center justify-between backdrop-blur-sm group">
          <div>
            <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">
              {uiStrings?.stat_total_deliveries || "Total Deliveries (24h)"}
            </p>
            <p className="text-2xl font-bold text-white">
              {webhookMeta?.total_calls || "--"}
            </p>
            <p className="text-[10px] text-emerald-500 font-bold mt-1 flex items-center gap-1">
              <span className="text-emerald-500 font-bold ml-auto">
                {webhookMeta?.health_increase || "+--%"}{" "}
                {uiStrings?.increase_label || "Increase"}
              </span>
            </p>
          </div>
          <SignalIcon className="w-8 h-8 text-primary opacity-20 group-hover:opacity-40 transition-opacity" />
        </div>

        <div className="bg-slate-900/40 border border-slate-800 p-6 rounded-2xl flex items-center justify-between backdrop-blur-sm group">
          <div>
            <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">
              {uiStrings?.health_score_label || "Webhook Health Score"}
            </p>
            <p className="text-2xl font-bold text-emerald-500">
              {systemMeta?.success_rate || "--%"}
            </p>
            <div className="w-32 h-1 bg-slate-800 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-emerald-500 h-full"
                style={{ width: systemMeta?.success_rate || "0%" }}
              />
            </div>
          </div>
          <CheckCircleIcon className="w-8 h-8 text-emerald-500 opacity-20 group-hover:opacity-40 transition-opacity" />
        </div>
      </div>

      {/* Endpoints List */}
      <div className="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden backdrop-blur-sm">
        <div className="p-6 border-b border-white/5 flex justify-between items-center">
          <h3 className="text-lg font-bold text-white flex items-center gap-2">
            <GlobeAltIcon className="w-5 h-5 text-primary" />
            {uiStrings?.configured_endpoints_title || "Configured Endpoints"}
          </h3>
          <button className="flex items-center gap-2 bg-primary hover:bg-primary-dark text-white px-4 py-2 rounded-xl text-xs font-bold transition-all shadow-lg shadow-primary/20">
            <PlusIcon className="w-4 h-4" />
            {uiStrings?.add_endpoint_btn || "Add Endpoint"}
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-slate-950/50 text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-white/5">
                <th className="px-6 py-4">
                  {uiStrings?.th_endpoint_url || "Endpoint URL"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_event_types || "Event Types"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_health || "Health"}
                </th>
                <th className="px-6 py-4 text-right">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {webhooks.map((wh) => (
                <tr
                  key={wh.id}
                  className="hover:bg-white/[0.02] transition-colors"
                >
                  <td
                    className="px-6 py-4 text-sm font-medium text-white max-w-xs truncate"
                    title={wh.url}
                  >
                    {wh.url}
                    <div className="text-[10px] text-slate-500 font-mono mt-1">
                      ID: {wh.id}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-1 max-w-md">
                      {(wh.events || []).map((e: string) => (
                        <span
                          key={e}
                          className="px-2 py-0.5 rounded bg-slate-950 border border-white/5 text-[9px] font-mono text-slate-400"
                        >
                          {e}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-lg bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 text-[10px] font-bold uppercase">
                      {wh.successRate}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button className="p-2 text-slate-500 hover:text-white hover:bg-slate-800 rounded-xl transition-all">
                        <PencilSquareIcon className="w-4 h-4" />
                      </button>
                      <button className="p-2 text-slate-500 hover:text-rose-500 hover:bg-rose-500/10 rounded-xl transition-all">
                        <TrashIcon className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default WebhookManager;
