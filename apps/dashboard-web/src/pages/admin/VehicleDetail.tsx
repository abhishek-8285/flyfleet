import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import apiClient from "../../services/apiClient";

interface Vehicle {
  id: string;
  name: string;
  license_plate: string;
  status: string;
  org_id: string;
  current_driver_id?: string;
  current_driver_name?: string;
  current_driver_since?: string;
  device_id?: string;
}

export const VehicleDetail: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchVehicle = async () => {
      try {
        const direct = (await apiClient.get(`/admin/vehicles?id=${id}`)) as any;
        const directVehicle = direct?.data ?? direct;
        if (directVehicle?.id) {
          setVehicle(directVehicle);
          return;
        }

        // Fallback for older API responses.
        const res = (await apiClient.get(
          "/admin/vehicles?page=1&size=100",
        )) as any;
        const list = res?.data ?? [];
        const v = list.find((item: any) => item.id === id);
        if (v) setVehicle(v);
      } catch (error) {
        console.error("Failed to fetch vehicle:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchVehicle();
  }, [id]);

  if (loading)
    return (
      <div className="flex items-center justify-center min-h-screen bg-surface-dim">
        <div className="flex flex-col items-center gap-4">
          <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary" />
          <p className="text-on-surface-variant font-bold text-xs uppercase tracking-widest">Initialising Asset Profile...</p>
        </div>
      </div>
    );
  if (!vehicle)
    return (
      <div className="flex items-center justify-center min-h-screen bg-surface-dim">
        <div className="text-center">
          <span className="material-symbols-outlined text-6xl text-rose-500/20 mb-4">error</span>
          <p className="text-white text-xl font-bold">Vehicle Not Found</p>
          <button onClick={() => navigate("/admin/vehicles")} className="mt-4 text-primary hover:underline font-bold text-sm">Return to Directory</button>
        </div>
      </div>
    );

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-8 bg-surface-dim min-h-screen">
      <button
        onClick={() => navigate("/admin/vehicles")}
        className="flex items-center gap-2 text-on-surface-variant hover:text-white transition-all mb-4 group font-bold text-[11px] uppercase tracking-wider"
      >
        <span className="material-symbols-outlined text-xs group-hover:-translate-x-1 transition-transform">arrow_back</span>
        <span>Return to Directory</span>
      </button>

      {/* Header Section */}
      <div className="bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl flex justify-between items-start">
        <div className="flex gap-8">
          <div className="h-28 w-28 bg-primary/10 rounded-2xl flex items-center justify-center text-primary border border-primary/20 shadow-inner">
            <span className="material-symbols-outlined text-5xl">local_shipping</span>
          </div>
          <div className="flex flex-col justify-center">
            <h1 className="text-4xl font-bold text-white tracking-tighter">
              {vehicle.name}
            </h1>
            <p className="text-on-surface-variant mt-2 font-mono text-xl font-bold">
              {vehicle.license_plate}
            </p>
            <div className="flex items-center gap-4 mt-6">
              <span
                className={`px-3 py-1 rounded-lg text-[10px] font-black uppercase tracking-widest border ${
                  vehicle.status === "active"
                    ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                    : "bg-surface-container-highest text-on-surface-variant border-outline"
                }`}
              >
                {vehicle.status}
              </span>
              <div className="w-1 h-1 rounded-full bg-outline" />
              <span className="text-[10px] text-on-surface-variant font-bold uppercase tracking-widest">System ID: {vehicle.id.slice(0, 8)}...</span>
            </div>
          </div>
        </div>
        <div className="flex gap-3">
          <button className="bg-surface-container-highest border border-outline text-white px-5 py-2.5 rounded-xl text-[11px] font-bold uppercase tracking-wider hover:bg-white/5 transition-all active:scale-95">
            Maintenance Log
          </button>
          <button className="bg-primary text-white px-5 py-2.5 rounded-xl text-[11px] font-bold uppercase tracking-wider hover:bg-primary-dark transition-all shadow-lg shadow-primary/20 active:scale-95">
            Edit Asset
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Core Vehicle Stats */}
        <div className="bg-surface-container border border-outline rounded-3xl p-6 space-y-6">
          <h3 className="text-white font-bold text-xs uppercase tracking-widest border-b border-outline pb-4">Operational Status</h3>
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-surface-dim rounded-2xl border border-outline">
              <p className="text-on-surface-variant text-[9px] font-bold uppercase tracking-widest mb-1">Fuel Level</p>
              <p className="text-white font-bold text-lg">78%</p>
            </div>
            <div className="p-4 bg-surface-dim rounded-2xl border border-outline">
              <p className="text-on-surface-variant text-[9px] font-bold uppercase tracking-widest mb-1">Odometer</p>
              <p className="text-white font-bold text-lg">24,521 <span className="text-[10px]">km</span></p>
            </div>
            <div className="p-4 bg-surface-dim rounded-2xl border border-outline">
              <p className="text-on-surface-variant text-[9px] font-bold uppercase tracking-widest mb-1">Battery</p>
              <p className="text-emerald-400 font-bold text-lg">Good</p>
            </div>
            <div className="p-4 bg-surface-dim rounded-2xl border border-outline">
              <p className="text-on-surface-variant text-[9px] font-bold uppercase tracking-widest mb-1">Next Service</p>
              <p className="text-amber-400 font-bold text-lg">1,200 <span className="text-[10px]">km</span></p>
            </div>
          </div>
        </div>

        {/* Driver Assignment Card */}
        <div className="bg-surface-container border border-outline rounded-3xl p-6 flex flex-col">
          <div className="flex justify-between items-center mb-6">
            <h3 className="text-white font-bold text-xs uppercase tracking-widest">
              Current Personnel
            </h3>
            {vehicle.current_driver_id && (
              <button className="text-[10px] text-primary font-black uppercase tracking-widest hover:underline">
                Reassign
              </button>
            )}
          </div>

          <div className="flex-1 flex flex-col justify-center">
            {vehicle.current_driver_id ? (
              <div className="flex items-center gap-4 p-5 bg-surface-dim rounded-2xl border border-outline hover:border-primary/40 transition-all group">
                <div className="h-14 w-14 bg-primary/20 rounded-full flex items-center justify-center text-primary font-black text-lg border border-primary/20">
                  {vehicle.current_driver_name?.split(' ').map(n => n[0]).join('') || "JD"}
                </div>
                <div>
                  <p className="font-bold text-white text-lg tracking-tight group-hover:text-primary transition-colors">
                    {vehicle.current_driver_name || "Unassigned"}
                  </p>
                  <p className="text-[10px] text-on-surface-variant font-bold uppercase tracking-wider mt-1">
                    Active since {vehicle.current_driver_since || "---"}
                  </p>
                </div>
              </div>
            ) : (
              <div className="text-center py-8 bg-surface-container-highest/20 rounded-2xl border border-dashed border-outline">
                <span className="material-symbols-outlined text-4xl text-on-surface-variant/20 mb-3">person_add</span>
                <p className="text-on-surface-variant font-bold text-[10px] uppercase tracking-widest mb-4">
                  No personnel assigned
                </p>
                <button className="bg-white/5 border border-outline text-white px-5 py-2 rounded-xl text-[10px] font-bold uppercase tracking-wider hover:bg-white/10 transition-all active:scale-95">
                  Assign Driver
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Device Intelligence Card */}
        <div className="bg-surface-container border border-outline rounded-3xl p-6 flex flex-col">
          <div className="flex justify-between items-center mb-6">
            <h3 className="text-white font-bold text-xs uppercase tracking-widest text">
              Telemetry Status
            </h3>
            <button className="text-[10px] text-primary font-black uppercase tracking-widest hover:underline">
              Diagnostics
            </button>
          </div>

          <div className="flex-1 flex flex-col justify-center">
            <div className="text-center py-8 bg-surface-container-highest/20 rounded-2xl border border-dashed border-outline">
              <span className="material-symbols-outlined text-4xl text-on-surface-variant/20 mb-3">sensors_off</span>
              <p className="text-on-surface-variant font-bold text-[10px] uppercase tracking-widest mb-4">
                No telemetry link active
              </p>
              <button className="bg-white/5 border border-outline text-white px-5 py-2 rounded-xl text-[10px] font-bold uppercase tracking-wider hover:bg-white/10 transition-all active:scale-95">
                Link Device
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
