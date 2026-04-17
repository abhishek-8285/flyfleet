import axios from "axios";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";

const PortalDashboard: React.FC = () => {
  const [stats, setStats] = useState({
    activeShipments: 0,
    pendingInvoices: 0,
    totalSpent: 0,
  });
  const [recentShipments, setRecentShipments] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [uiStrings, setUiStrings] = useState<any>(null);

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.portal || data.system_metadata?.ui_strings?.portal,
        ),
      )
      .catch((err) => console.error("Failed to fetch UI strings", err));
  }, []);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const token = localStorage.getItem("customer_token");
        if (!token) return;

        const config = { headers: { Authorization: `Bearer ${token}` } };

        // Fetch shipments
        const shipmentsRes = await axios.get(
          "/api/v1/portal/shipments",
          config,
        );
        setRecentShipments(shipmentsRes.data.slice(0, 5)); // Top 5
        setStats((prev) => ({
          ...prev,
          activeShipments: shipmentsRes.data.filter(
            (s: any) => s.status !== "DELIVERED",
          ).length,
        }));

        // Fetch invoices
        const invoicesRes = await axios.get("/api/v1/portal/invoices", config);
        setStats((prev) => ({
          ...prev,
          pendingInvoices: invoicesRes.data.filter(
            (i: any) => i.status === "PENDING",
          ).length,
          totalSpent: invoicesRes.data.reduce(
            (acc: number, curr: any) => acc + (curr.amount || 0),
            0,
          ),
        }));
      } catch (err) {
        console.error("Failed to fetch dashboard data", err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  return (
    <div className="space-y-8 bg-surface-dim min-h-screen p-8 -mx-6 -mt-6">
      <div className="flex justify-between items-end bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl">
        <div>
          <h1 className="text-4xl font-bold text-white tracking-tighter uppercase mb-2">
            {uiStrings?.dashboard_title || "Client Operations Command"}
          </h1>
          <p className="text-on-surface-variant text-[11px] font-black uppercase tracking-[0.2em] flex items-center gap-2">
            <span className="material-symbols-outlined text-sm">person_pin</span>
            {uiStrings?.welcome_prefix || "Authenticated Organization:"} <span className="text-white">Acme Global Logistics</span>
          </p>
        </div>
        <Link
          to="/portal/bookings"
          className="bg-primary hover:bg-primary-dark text-white px-8 py-3.5 rounded-2xl text-[10px] font-black uppercase tracking-widest transition-all shadow-lg shadow-primary/20 flex items-center gap-3 active:scale-95"
        >
          <span className="material-symbols-outlined text-base">add_circle</span>
          {uiStrings?.new_booking_btn || "Initiate Shipment"}
        </Link>
      </div>

      {/* Stats Cards / Intelligence Tiles */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl hover:border-primary/30 transition-all group">
          <div className="flex justify-between items-start mb-6">
            <div className="p-3 bg-primary/10 rounded-2xl text-primary border border-primary/20 group-hover:scale-110 transition-transform">
              <span className="material-symbols-outlined text-2xl">
                local_shipping
              </span>
            </div>
            <span className="text-[9px] font-black tracking-widest uppercase text-emerald-400 bg-emerald-400/10 px-3 py-1 rounded-full border border-emerald-400/20">
              In Transit
            </span>
          </div>
          <div className="text-4xl font-bold text-white mb-2 tracking-tighter">
            {stats.activeShipments}
          </div>
          <div className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]">
            {uiStrings?.active_shipments_label || "Active Manifests"}
          </div>
        </div>

        <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl hover:border-amber-500/30 transition-all group">
          <div className="flex justify-between items-start mb-6">
            <div className="p-3 bg-amber-500/10 rounded-2xl text-amber-500 border border-amber-500/20 group-hover:scale-110 transition-transform">
              <span className="material-symbols-outlined text-2xl">
                pending_actions
              </span>
            </div>
          </div>
          <div className="text-4xl font-bold text-white mb-2 tracking-tighter">
            {stats.pendingInvoices}
          </div>
          <div className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]">
            {uiStrings?.pending_invoices_label || "Open Payables"}
          </div>
        </div>

        <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl hover:border-emerald-500/30 transition-all group">
          <div className="flex justify-between items-start mb-6">
            <div className="p-3 bg-emerald-500/10 rounded-2xl text-emerald-500 border border-emerald-500/20 group-hover:scale-110 transition-transform">
              <span className="material-symbols-outlined text-2xl">
                account_balance_wallet
              </span>
            </div>
          </div>
          <div className="text-4xl font-bold text-white mb-2 tracking-tighter">
            ₹{stats.totalSpent.toLocaleString('en-IN')}
          </div>
          <div className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]">
            {uiStrings?.total_spent_label || "Net Expenditure (YTD)"}
          </div>
        </div>
      </div>

      {/* Recent Shipments / Operational Stream */}
      <div className="bg-surface-container border border-outline rounded-[2.5rem] overflow-hidden shadow-2xl">
        <div className="px-8 py-6 border-b border-outline flex justify-between items-center bg-surface-container/50">
          <h3 className="text-[11px] font-black text-white uppercase tracking-[0.25em] flex items-center gap-3">
            <span className="material-symbols-outlined text-primary text-xl">conversion_path</span>
            {uiStrings?.recent_shipments_title || "Operational Pipeline Status"}
          </h3>
          <Link
            to="/portal/shipments"
            className="text-[10px] font-black uppercase tracking-widest text-primary hover:text-white transition-all bg-primary/10 px-4 py-2 rounded-xl"
          >
            {uiStrings?.view_all_link || "Full Directory"}
          </Link>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead className="bg-surface-dim border-b border-outline text-[9px] uppercase font-black text-on-surface-variant tracking-[0.2em]">
              <tr>
                <th className="px-8 py-5">
                  {uiStrings?.th_tracking_id || "Manifest ID"}
                </th>
                <th className="px-8 py-5">
                  {uiStrings?.th_origin || "Loading Point"}
                </th>
                <th className="px-8 py-5">
                  {uiStrings?.th_destination || "Discharge Point"}
                </th>
                <th className="px-8 py-5">
                  {uiStrings?.th_status || "Logistics Status"}
                </th>
                <th className="px-8 py-5">{uiStrings?.th_eta || "ETA (System Estimate)"}</th>
                <th className="px-8 py-5 text-right">
                  {uiStrings?.th_actions || "Operations"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-outline/30">
              {loading ? (
                <tr>
                  <td colSpan={6} className="p-4 text-center">
                    {uiStrings?.loading_label || "Loading..."}
                  </td>
                </tr>
              ) : recentShipments.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-4 text-center">
                    {uiStrings?.no_shipments_found ||
                      "No active shipments found."}
                  </td>
                </tr>
              ) : (
                recentShipments.map((shipment) => (
                  <tr
                    key={shipment.id}
                    className="hover:bg-surface-light/5 transition-colors"
                  >
                    <td className="px-8 py-5 font-bold text-white font-mono text-xs">
                      #{shipment.tracking_number}
                    </td>
                    <td className="px-8 py-5 text-[11px] font-bold text-on-surface-variant uppercase tracking-tighter">
                      {shipment.origin?.address || "Pending Registry"}
                    </td>
                    <td className="px-8 py-5 text-[11px] font-bold text-on-surface-variant uppercase tracking-tighter">
                      {shipment.destination?.address || "In Analysis"}
                    </td>
                    <td className="px-8 py-5">
                      <span
                        className={`px-3 py-1 rounded-lg text-[9px] font-black uppercase tracking-[0.15em] 
                        ${
                          shipment.status === "DELIVERED"
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : shipment.status === "IN_TRANSIT"
                              ? "bg-primary/10 text-primary border border-primary/20"
                              : "bg-surface-dim text-on-surface-variant border border-outline"
                        }`}
                      >
                        {shipment.status}
                      </span>
                    </td>
                    <td className="px-8 py-5 text-[11px] font-mono font-bold text-on-surface-variant">
                      {shipment.estimated_delivery
                        ? new Date(
                            shipment.estimated_delivery,
                          ).toLocaleDateString('en-GB')
                        : "-- -- ----"}
                    </td>
                    <td className="px-8 py-5 text-right">
                      <Link
                        to={`/portal/track/${shipment.tracking_number}`}
                        className="bg-surface-dim hover:bg-surface-container-highest text-primary hover:text-white text-[10px] font-black uppercase tracking-widest py-2 px-4 rounded-xl border border-outline transition-all inline-flex items-center gap-2 active:scale-95"
                      >
                        <span className="material-symbols-outlined text-sm">near_me</span>
                        Locate
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default PortalDashboard;
