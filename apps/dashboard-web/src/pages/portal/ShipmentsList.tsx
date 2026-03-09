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
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white">
          {uiStrings?.title || "Your Shipments"}
        </h1>
      </div>

      <div className="bg-surface-dark border border-border-color rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-text-muted">
            <thead className="bg-[#111a22] text-xs uppercase font-semibold text-text-muted/70">
              <tr>
                <th className="px-5 py-3">
                  {uiStrings?.th_tracking || "Tracking ID"}
                </th>
                <th className="px-5 py-3">
                  {uiStrings?.th_origin || "Origin"}
                </th>
                <th className="px-5 py-3">
                  {uiStrings?.th_destination || "Destination"}
                </th>
                <th className="px-5 py-3">
                  {uiStrings?.th_status || "Status"}
                </th>
                <th className="px-5 py-3">{uiStrings?.th_eta || "ETA"}</th>
                <th className="px-5 py-3">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {loading ? (
                <tr>
                  <td colSpan={6} className="p-10 text-center">
                    <div className="flex flex-col items-center gap-2">
                      <span className="material-symbols-outlined animate-spin text-primary">
                        progress_activity
                      </span>
                      <span>
                        {uiStrings?.loading || "Loading shipments..."}
                      </span>
                    </div>
                  </td>
                </tr>
              ) : shipments.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-10 text-center text-text-muted">
                    {uiStrings?.empty_state || "No shipments found."}
                  </td>
                </tr>
              ) : (
                shipments.map((s) => (
                  <tr
                    key={s.id}
                    className="hover:bg-surface-light/5 transition-colors"
                  >
                    <td className="px-5 py-4 font-medium text-white">
                      {s.tracking_number}
                    </td>
                    <td className="px-5 py-4">{s.origin?.address || "N/A"}</td>
                    <td className="px-5 py-4">
                      {s.destination?.address || "N/A"}
                    </td>
                    <td className="px-5 py-4">
                      <span
                        className={`px-2 py-1 rounded text-[10px] font-bold uppercase ${
                          s.status === "DELIVERED"
                            ? "bg-green-500/10 text-green-500"
                            : s.status === "IN_TRANSIT"
                              ? "bg-blue-500/10 text-blue-500"
                              : s.status === "PICKED_UP"
                                ? "bg-yellow-500/10 text-yellow-500"
                                : "bg-gray-500/10 text-gray-500"
                        }`}
                      >
                        {s.status}
                      </span>
                    </td>
                    <td className="px-5 py-4">
                      {s.estimated_delivery
                        ? new Date(s.estimated_delivery).toLocaleDateString()
                        : "-"}
                    </td>
                    <td className="px-5 py-4">
                      <Link
                        to={`/portal/track/${s.tracking_number}`}
                        className="text-primary hover:text-white transition-colors flex items-center gap-1"
                      >
                        <span className="material-symbols-outlined text-[18px]">
                          location_on
                        </span>
                        {uiStrings?.action_track || "Track"}
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
