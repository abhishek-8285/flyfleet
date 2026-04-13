import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import AddDocumentModal from "../../../components/documents/AddDocumentModal";
import DocumentList from "../../../components/documents/DocumentList";
import { Sidebar } from "../../../components/layout/Sidebar";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";
import { Vehicle } from "./types";

export const VehicleDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("fleet_vehicles");
  const { uiStrings: commonStrings } = useUIStrings("common");
  const uiStrings = rawUiStrings || {};
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"overview" | "documents">(
    "overview",
  );
  const [isAddDocOpen, setIsAddDocOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    fetchVehicle();
  }, [id]);

  const fetchVehicle = async () => {
    try {
      const res = (await apiClient.get(`/admin/vehicles/${id}`)) as any;
      setVehicle(res.data);
    } catch (error) {
      console.error("Failed to fetch vehicle:", error);
    } finally {
      setLoading(false);
    }
  };

  if (loading)
    return (
      <div className="text-white p-10">
        {uiStrings.msg_loading || "Loading..."}
      </div>
    );
  if (!vehicle)
    return (
      <div className="text-white p-10">
        {uiStrings.msg_not_found || "Vehicle not found"}
      </div>
    );

  return (
    <div className="flex h-screen w-full bg-[#0f111a]">
      <Sidebar />
      <main className="flex-1 flex flex-col min-w-0 overflow-y-auto">
        {/* Header */}
        <header className="h-20 flex items-center justify-between px-8 border-b border-white/5 bg-[#161b22] sticky top-0 z-10">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate("/admin/vehicles")}
              className="text-slate-400 hover:text-white transition-colors"
            >
              <span className="material-symbols-outlined">arrow_back</span>
            </button>
            <div>
              <h1 className="text-xl font-bold text-white">{vehicle.name}</h1>
              <div className="flex items-center gap-2 text-sm text-slate-400">
                <span
                  className={`w-2 h-2 rounded-full ${vehicle.status === "active" ? "bg-emerald-500" : "bg-slate-500"}`}
                ></span>
                {vehicle.status.toUpperCase()}
                <span className="text-slate-600">•</span>
                {vehicle.license_plate}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate(`/admin/vehicles/${vehicle.id}/edit`)}
              className="bg-[#1c222b] hover:bg-[#252b36] text-white px-4 py-2 rounded-lg border border-white/10 transition-colors flex items-center gap-2"
            >
              <span className="material-symbols-outlined text-sm">edit</span>
              {commonStrings?.btn_edit || uiStrings.btn_edit || "Edit"}
            </button>
          </div>
        </header>

        <div className="p-8 max-w-6xl mx-auto w-full space-y-6">
          {/* Key Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="bg-[#1c222b] p-5 rounded-xl border border-white/5">
              <div className="text-slate-400 text-xs uppercase font-medium mb-1">
                {uiStrings.th_odometer || "Odometer"}
              </div>
              <div className="text-2xl font-bold text-white">
                {vehicle.odometer?.toLocaleString() || 0}{" "}
                <span className="text-sm text-slate-500 font-normal">
                  {uiStrings.miles_label || "km"}
                </span>
              </div>
            </div>
            <div className="bg-[#1c222b] p-5 rounded-xl border border-white/5">
              <div className="text-slate-400 text-xs uppercase font-medium mb-1">
                {uiStrings.label_fuel_type || "Fuel Type"}
              </div>
              <div className="text-2xl font-bold text-white capitalize">
                {vehicle.fuel_type}
              </div>
            </div>
            <div className="bg-[#1c222b] p-5 rounded-xl border border-white/5">
              <div className="text-slate-400 text-xs uppercase font-medium mb-1">
                Service Due
              </div>
              <div className="text-2xl font-bold text-white">
                {vehicle.service_due
                  ? new Date(vehicle.service_due).toLocaleDateString()
                  : "N/A"}
              </div>
            </div>
            <div className="bg-[#1c222b] p-5 rounded-xl border border-white/5">
              <div className="text-slate-400 text-xs uppercase font-medium mb-1">
                {uiStrings.th_assigned_driver || "Driver"}
              </div>
              <div className="text-2xl font-bold text-white text-slate-500">
                {uiStrings.unassigned || "Unassigned"}
              </div>
            </div>
          </div>

          {/* Detailed Info */}
          {/* Tabs */}
          <div className="flex items-center gap-6 border-b border-white/5 px-1">
            <button
              onClick={() => setActiveTab("overview")}
              className={`pb-3 text-sm font-medium transition-colors relative ${activeTab === "overview" ? "text-emerald-500" : "text-slate-400 hover:text-white"}`}
            >
              {uiStrings.tab_overview || "Overview"}
              {activeTab === "overview" && (
                <span className="absolute bottom-0 left-0 w-full h-0.5 bg-emerald-500 rounded-t-full"></span>
              )}
            </button>
            <button
              onClick={() => setActiveTab("documents")}
              className={`pb-3 text-sm font-medium transition-colors relative ${activeTab === "documents" ? "text-emerald-500" : "text-slate-400 hover:text-white"}`}
            >
              {uiStrings.tab_documents || "Documents"}
              {activeTab === "documents" && (
                <span className="absolute bottom-0 left-0 w-full h-0.5 bg-emerald-500 rounded-t-full"></span>
              )}
            </button>
          </div>

          {activeTab === "overview" ? (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <div className="lg:col-span-2 space-y-6">
                <div className="bg-[#1c222b] rounded-xl border border-white/5 overflow-hidden">
                  <div className="p-4 border-b border-white/5 font-medium text-white">
                    {uiStrings.label_specifications || "Vehicle Specifications"}
                  </div>
                  <div className="p-6 grid grid-cols-2 gap-6">
                    <div>
                      <div className="text-slate-500 text-xs mb-1">
                        {uiStrings.label_vin || "VIN"}
                      </div>
                      <div className="text-white font-mono">
                        {vehicle.vin || "-"}
                      </div>
                    </div>
                    <div>
                      <div className="text-slate-500 text-xs mb-1">
                        {uiStrings.label_vehicle_no || "Vehicle No"}
                      </div>
                      <div className="text-white">
                        {vehicle.vehicle_no || "-"}
                      </div>
                    </div>
                    <div>
                      <div className="text-slate-500 text-xs mb-1">
                        {uiStrings.label_make_model || "Make / Model"}
                      </div>
                      <div className="text-white">
                        {vehicle.make} {vehicle.model}
                      </div>
                    </div>
                    <div>
                      <div className="text-slate-500 text-xs mb-1">
                        {uiStrings.label_year || "Year"}
                      </div>
                      <div className="text-white">{vehicle.year}</div>
                    </div>
                  </div>
                </div>

                <div className="bg-[#1c222b] rounded-xl border border-white/5 overflow-hidden">
                  <div className="p-4 border-b border-white/5 font-medium text-white">
                    {uiStrings.label_maintenance_notes || "Maintenance Notes"}
                  </div>
                  <div className="p-6 text-slate-300 text-sm whitespace-pre-wrap">
                    {vehicle.notes ||
                      uiStrings.no_notes_available ||
                      "No notes available."}
                  </div>
                </div>
              </div>

              <div className="space-y-6">
                {/* Map or device status could go here */}
                <div className="bg-[#1c222b] rounded-xl border border-white/5 overflow-hidden h-64 flex items-center justify-center text-slate-500">
                  <div className="text-center">
                    <span className="material-symbols-outlined text-4xl mb-2 opacity-50">
                      map
                    </span>
                    <p>
                      {uiStrings.live_tracking_unavailable ||
                        "Live Tracking Unavailable"}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <DocumentList
              key={refreshKey}
              vehicleId={vehicle.id}
              onAdd={() => setIsAddDocOpen(true)}
            />
          )}
        </div>
      </main>

      <AddDocumentModal
        vehicleId={vehicle.id}
        isOpen={isAddDocOpen}
        onClose={() => setIsAddDocOpen(false)}
        onSuccess={() => {
          // Refresh documents list? DocumentList automatically fetches on mount/id change.
          // But we might want to trigger a refresh.
          // For now, simpler to just close. DocumentList will need a refresh trigger if we want instant update.
          // Or we can pass a refresh key to DocumentList.
          // Let's modify DocumentList later to accept a refresh trigger or similar.
          // Actually, re-mounting DocumentList by toggling tab works, but generic `onAdd` doesn't propogate refresh.
          // The simplest way: Pass a refreshKey to DocumentList.
          setIsAddDocOpen(false);
          // Force refresh trick: toggle key or just reload page (bad).
          // Better: Add `key={refreshKey}` to DocumentList.
          setRefreshKey((prev) => prev + 1);
        }}
      />
    </div>
  );
};
