import React, { useEffect, useState } from "react";
import billingService from "../../../services/billingService";
import { Zap, Activity, Info, AlertTriangle } from "lucide-react";
import { useUIStrings } from "../../../hooks/useUIStrings";

const UsageOverview: React.FC = () => {
  const [usage, setUsage] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const { uiStrings } = useUIStrings("admin_billing");

  useEffect(() => {
    fetchUsage();
  }, []);

  const fetchUsage = async () => {
    try {
      const data = await billingService.getUsageSummary();
      setUsage(data || []);
    } catch (error) {
      console.error("Failed to fetch usage", error);
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateInvoice = async () => {
    setGenerating(true);
    try {
      await billingService.generateDraftInvoice();
      // Wait a bit then refresh usage (it should be empty now as it's processed into a draft)
      setTimeout(() => {
        fetchUsage();
        window.location.reload(); // Refresh to show new draft in overview
      }, 1500);
    } catch (error) {
      console.error("Failed to generate draft", error);
    } finally {
      setGenerating(false);
    }
  };

  if (loading) return (
    <div className="p-12 flex justify-center">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>
  );

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-surface-dark border border-border-color rounded-2xl overflow-hidden shadow-xl">
            <div className="p-6 border-b border-border-color flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Activity className="w-5 h-5 text-primary" />
                <h3 className="text-white font-bold">
                  {uiStrings?.usage_title || "Current Consumption"}
                </h3>
              </div>
              {usage?.length > 0 && (
                <button
                  onClick={handleGenerateInvoice}
                  disabled={generating}
                  className="bg-primary text-white px-4 py-2 rounded-lg text-sm font-bold hover:bg-primary/90 transition-all flex items-center gap-2 disabled:opacity-50"
                >
                  {generating ? (
                    <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
                  ) : (
                    <Zap className="w-4 h-4" />
                  )}
                  {uiStrings?.btn_generate_draft || "Generate Invoice Now"}
                </button>
              )}
            </div>

            <div className="p-0">
              <table className="w-full text-left">
                <thead className="bg-white/5 border-b border-border-color">
                  <tr>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">Metric</th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest text-right">Usage</th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest text-right">Amount</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {usage?.map((item, idx) => (
                    <tr key={idx} className="hover:bg-white/5 transition-colors">
                      <td className="px-6 py-4">
                        <p className="text-white font-medium">{item.name}</p>
                        <p className="text-[10px] text-text-muted font-mono uppercase mt-0.5">{item.meter_code}</p>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <span className="text-white font-bold">{item.total.toLocaleString()}</span>
                        <span className="text-[10px] text-text-muted ml-1 uppercase">{item.unit}</span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <span className="text-white font-bold">₹{item.amount.toLocaleString()}</span>
                      </td>
                    </tr>
                  ))}
                  {(usage?.length === 0 || !usage) && (
                    <tr>
                      <td colSpan={3} className="px-6 py-12 text-center text-text-muted">
                        <div className="flex flex-col items-center gap-3">
                          <Info className="w-8 h-8 opacity-20" />
                          <p>{uiStrings?.no_usage || "No unbilled usage found for this period."}</p>
                        </div>
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          <div className="bg-primary/5 border border-primary/20 rounded-2xl p-6 flex items-start gap-4">
            <div className="bg-primary/20 p-2 rounded-lg text-primary">
              <Info className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-primary font-bold text-sm uppercase tracking-wider mb-1">How Metering Works</h4>
              <p className="text-sm text-text-muted leading-relaxed">
                Usage is tracked in real-time based on active GPS pings, completed trips, and daily vehicle availability. 
                Invoices are typically generated automatically on the 1st of every month, but you can trigger a draft 
                anytime for mid-cycle billing.
              </p>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-surface-dark border border-border-color rounded-2xl p-6 shadow-xl">
            <h4 className="text-white font-bold mb-4 flex items-center gap-2">
              <Zap className="w-4 h-4 text-primary" />
              Usage Caps
            </h4>
            <div className="space-y-4">
              <div className="space-y-2">
                <div className="flex justify-between text-xs">
                  <span className="text-text-muted">Active Vehicles</span>
                  <span className="text-white">12 / 20</span>
                </div>
                <div className="w-full bg-white/5 h-1.5 rounded-full overflow-hidden">
                  <div className="bg-primary h-full rounded-full" style={{ width: '60%' }}></div>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-xs">
                  <span className="text-text-muted">API Queries</span>
                  <span className="text-white">8.2k / 50k</span>
                </div>
                <div className="w-full bg-white/5 h-1.5 rounded-full overflow-hidden">
                  <div className="bg-green-500 h-full rounded-full" style={{ width: '16%' }}></div>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-xs">
                  <span className="text-text-muted">Monthly Trips</span>
                  <span className="text-white">1,402 / Unlimited</span>
                </div>
                <div className="w-full bg-white/5 h-1.5 rounded-full overflow-hidden">
                  <div className="bg-blue-500 h-full rounded-full" style={{ width: '45%' }}></div>
                </div>
              </div>
            </div>
            <button className="w-full mt-6 py-2 text-xs font-bold text-primary border border-primary/30 rounded-lg hover:bg-primary/10 transition-all">
              Upgrade Limits
            </button>
          </div>

          <div className="bg-amber-500/5 border border-amber-500/20 rounded-2xl p-6 space-y-3">
             <div className="flex items-center gap-2 text-amber-500">
               <AlertTriangle className="w-4 h-4" />
               <span className="text-xs font-bold uppercase tracking-wider">Quota Warning</span>
             </div>
             <p className="text-xs text-amber-500/80 leading-relaxed">
               Your "Active Vehicles" quota is at 60%. Consider upgrading to the PRO tier to avoid tracking interruptions.
             </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UsageOverview;
