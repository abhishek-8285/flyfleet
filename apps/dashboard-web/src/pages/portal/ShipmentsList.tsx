import axios from "axios";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

const ShipmentsList: React.FC = () => {
  const [shipments, setShipments] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("portal_shipments");

  useEffect(() => {
    const fetchShipments = async () => {
      try {
        const token = localStorage.getItem("customer_token");
        if (!token) return;

        const config = { headers: { Authorization: `Bearer ${token}` } };
        const res = await axios.get("/api/v1/portal/shipments", config);
        setShipments(res.data || []);
      } catch (err) {
        console.error("Failed to fetch shipments", err);
      } finally {
        setLoading(false);
      }
    };

    fetchShipments();
  }, []);

  return (
    <div className="space-y-8 bg-surface-dim min-h-screen p-8 -mx-6 -mt-6">
      <div className="flex justify-between items-end bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl">
        <div>
          <nav className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-on-surface-variant mb-4">
            <Link to="/portal" className="hover:text-primary transition-colors">Command Overview</Link>
            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
            <span className="text-white">Active Pipeline</span>
          </nav>
          <h1 className="text-4xl font-bold text-white tracking-tighter uppercase">
            {uiStrings?.title || "Operational Pipeline Directory"}
          </h1>
        </div>
      </div>

      <div className="bg-surface-container border border-outline rounded-[2.5rem] overflow-hidden shadow-2xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead className="bg-surface-dim border-b border-outline text-[9px] uppercase font-black text-on-surface-variant tracking-[0.2em]">
              <tr>
                <th className="px-8 py-5">
                  {uiStrings?.th_tracking || "Manifest ID"}
                </th>
                <th className="px-8 py-5">
                  {uiStrings?.th_origin || "Loading Hub"}
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
                  <td colSpan={6} className="p-20 text-center">
                    <div className="flex flex-col items-center gap-4">
                      <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary" />
                      <span className="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">
                        {uiStrings?.loading || "Initialising Operational Data..."}
                      </span>
                    </div>
                  </td>
                </tr>
              ) : shipments.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-20 text-center text-on-surface-variant">
                    <span className="material-symbols-outlined text-6xl opacity-20 mb-4 block">inventory_2</span>
                    <p className="text-[11px] font-black uppercase tracking-widest">
                      {uiStrings?.empty_state || "No active manifests detected in pipeline."}
                    </p>
                  </td>
                </tr>
              ) : (
                shipments.map((s) => (
                  <tr
                    key={s.id}
                    className="hover:bg-surface-light/5 transition-colors group"
                  >
                    <td className="px-8 py-5 font-bold text-white font-mono text-xs">
                      #{s.tracking_number}
                    </td>
                    <td className="px-8 py-5 text-[11px] font-bold text-on-surface-variant uppercase tracking-tighter">
                      {s.origin?.address || "Pending Registry"}
                    </td>
                    <td className="px-8 py-5 text-[11px] font-bold text-on-surface-variant uppercase tracking-tighter">
                      {s.destination?.address || "In Analysis"}
                    </td>
                    <td className="px-8 py-5">
                      <span
                        className={`px-3 py-1 rounded-lg text-[9px] font-black uppercase tracking-[0.15em] 
                        ${
                          s.status === "DELIVERED"
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : s.status === "IN_TRANSIT"
                              ? "bg-primary/10 text-primary border border-primary/20"
                              : s.status === "PICKED_UP"
                                ? "bg-amber-500/10 text-amber-500 border border-amber-500/20"
                                : "bg-surface-dim text-on-surface-variant border border-outline"
                        }`}
                      >
                        {s.status}
                      </span>
                    </td>
                    <td className="px-8 py-5 text-[11px] font-mono font-bold text-on-surface-variant">
                      {s.estimated_delivery
                        ? new Date(s.estimated_delivery).toLocaleDateString('en-GB')
                        : "-- -- ----"}
                    </td>
                    <td className="px-8 py-5 text-right">
                      <Link
                        to={`/portal/track/${s.tracking_number}`}
                        className="bg-surface-dim hover:bg-surface-container-highest text-primary hover:text-white text-[10px] font-black uppercase tracking-widest py-2.5 px-6 rounded-xl border border-outline transition-all inline-flex items-center gap-3 active:scale-95 group-hover:border-primary/50"
                      >
                        <span className="material-symbols-outlined text-base">near_me</span>
                        {uiStrings?.action_track || "Locate"}
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

export default ShipmentsList;
