import React, { useEffect, useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import { api } from "../../../services/api";
import { Trip, TripEvent } from "../../../modules/trips/types";
import { format } from "date-fns";
import { LoadingSpinner } from "../../../components/common/LoadingSpinner";

export const TripDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [trip, setTrip] = useState<Trip | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadTrip = async () => {
      if (!id) return;
      try {
        const res = await api.getTrip(id);
        setTrip(res.data || res);
      } catch (err) {
        console.error("Failed to load trip", err);
      } finally {
        setLoading(false);
      }
    };
    loadTrip();
  }, [id]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-4 bg-surface-dim min-h-screen">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary" />
        <p className="text-on-surface-variant font-black uppercase tracking-[0.2em] text-[10px]">Initialising Manifest Data...</p>
      </div>
    );
  }

  if (!trip) {
    return (
      <div className="flex flex-col items-center justify-center py-24 text-on-surface-variant bg-surface-dim min-h-screen text-center">
        <span className="material-symbols-outlined text-6xl text-rose-500/20 mb-4">route_off</span>
        <h2 className="text-xl font-bold text-white mb-2">Manifest Not Found</h2>
        <p className="mb-6 text-sm">The requested operational manifest could not be retrieved from the central directory.</p>
        <button onClick={() => navigate(-1)} className="px-6 py-2.5 bg-primary text-white rounded-xl hover:bg-primary-dark transition-all font-bold uppercase text-[10px] tracking-widest shadow-lg shadow-primary/20 active:scale-95">
          Return to directory
        </button>
      </div>
    );
  }

  const getEventIcon = (type: string) => {
    switch (type) {
      case "arrival": return "location_on";
      case "departure": return "near_me";
      case "speeding": return "speed";
      case "braking": return "error";
      case "accident": return "collision";
      case "delay": return "schedule";
      default: return "adjust";
    }
  };

  const getEventColor = (type: string) => {
    switch (type) {
      case "speeding":
      case "braking":
      case "accident": return "text-rose-500 bg-rose-500/10 border-rose-500/20";
      case "delay": return "text-amber-500 bg-amber-500/10 border-amber-500/20";
      case "arrival":
      case "departure": return "text-emerald-500 bg-emerald-500/10 border-emerald-500/20";
      default: return "text-primary bg-primary/10 border-primary/20";
    }
  };

  return (
    <div className="flex-1 flex flex-col bg-surface-dim p-8 -mx-6 -mt-6 gap-8 min-h-screen">
      {/* Dynamic Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl">
        <div>
          <nav className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-on-surface-variant mb-4">
            <Link to="/admin/trips" className="hover:text-primary transition-colors">Dispatch Log</Link>
            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
            <span className="text-white">#{trip.id}</span>
          </nav>
          <div className="flex items-center gap-6">
            <button onClick={() => navigate(-1)} className="h-12 w-12 flex items-center justify-center bg-surface-dim border border-outline rounded-xl hover:bg-surface-container-highest transition-all group">
              <span className="material-symbols-outlined text-on-surface-variant group-hover:-translate-x-1 transition-transform">arrow_back</span>
            </button>
            <div>
              <div className="flex items-center gap-4 mb-2">
                <h2 className="text-4xl font-bold text-white tracking-tighter uppercase">{trip.title}</h2>
                <span className={`px-3 py-1 rounded-lg text-[10px] font-black uppercase tracking-widest border ${
                  trip.status === "STARTED"
                    ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                    : "bg-surface-dim text-on-surface-variant border-outline"
                }`}>
                  {trip.status?.replace("_", " ")}
                </span>
              </div>
              <p className="text-on-surface-variant text-[11px] font-bold uppercase tracking-wider flex items-center gap-2">
                <span className="material-symbols-outlined text-sm">inventory_2</span>
                Manifest ID: <span className="text-white font-mono">#{trip.id}</span>
                <span className="w-1 h-1 rounded-full bg-outline mx-2" />
                <span className="material-symbols-outlined text-sm">corporate_fare</span>
                {trip.customer_name || "Internal Operations"}
              </p>
            </div>
          </div>
        </div>
        <div className="flex gap-3">
          <button className="bg-surface-dim border border-outline text-white px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-surface-container-highest transition-all flex items-center gap-3 active:scale-95 shadow-lg">
            <span className="material-symbols-outlined text-base">print</span> Export Manifest
          </button>
          <button className="bg-primary text-white px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-primary/20 hover:bg-primary-dark transition-all flex items-center gap-3 active:scale-95">
            <span className="material-symbols-outlined text-base">near_me</span> Operational Map
          </button>
        </div>
      </div>

      {/* Main Grid: Map & Content */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 flex-1 min-h-0">
        
        {/* Left Column: Map & Route (60%) */}
        <div className="lg:col-span-8 flex flex-col gap-8 h-full">
          <div className="flex-1 bg-surface-container rounded-[2.5rem] border border-outline relative overflow-hidden min-h-[500px] shadow-2xl group">
            {/* Map Overlay Stats */}
            <div className="absolute top-8 left-8 right-8 flex justify-between items-start z-10 pointer-events-none">
              <div className="bg-surface-dim/90 backdrop-blur-2xl p-6 rounded-3xl border border-outline shadow-2xl text-white flex gap-10">
                <div>
                  <p className="text-[9px] font-black text-on-surface-variant tracking-[0.2em] uppercase mb-2">Live Telemetry</p>
                  <p className="text-3xl font-bold tracking-tighter">68 <span className="text-[11px] font-black text-on-surface-variant ml-1 uppercase">KM/H</span></p>
                </div>
                <div className="w-px h-12 bg-outline" />
                <div>
                  <p className="text-[9px] font-black text-on-surface-variant tracking-[0.2em] uppercase mb-2">Distance Remaining</p>
                  <p className="text-3xl font-bold tracking-tighter">142 <span className="text-[11px] font-black text-on-surface-variant ml-1 uppercase">KM</span></p>
                </div>
              </div>
              
              <div className="bg-emerald-500 text-white px-4 py-2 rounded-xl font-black text-[9px] tracking-[0.15em] uppercase shadow-lg shadow-emerald-500/20 border border-emerald-400/20">
                Security Validated
              </div>
            </div>

            {/* Mock Map Background */}
            <div className="absolute inset-0 bg-surface-dim opacity-100 flex items-center justify-center">
              <div className="relative w-full h-full">
                 <div className="absolute inset-0 bg-gradient-to-t from-surface-dim via-transparent to-transparent z-[1]" />
                 <div className="w-full h-full flex items-center justify-center">
                    <div className="text-center group-hover:scale-105 transition-transform duration-[2s]">
                       <span className="material-symbols-outlined text-8xl text-primary/10">explore</span>
                       <p className="text-on-surface-variant font-black text-[10px] tracking-[0.3em] uppercase mt-6 opacity-50">GIS Operational Grid Initialising...</p>
                    </div>
                 </div>
              </div>
            </div>

            {/* Map Footer: Current Pos */}
            <div className="absolute bottom-8 left-8 right-8 z-10 pointer-events-none">
               <div className="bg-surface-dim/90 backdrop-blur-2xl p-5 rounded-[1.5rem] border border-outline shadow-2xl flex items-center gap-5">
                  <div className="h-12 w-12 bg-primary/10 rounded-xl flex items-center justify-center border border-primary/20">
                    <span className="material-symbols-outlined text-primary text-2xl">my_location</span>
                  </div>
                  <div>
                    <p className="text-[9px] font-black text-on-surface-variant tracking-[0.2em] uppercase mb-1">Last Reported Position</p>
                    <p className="text-sm font-bold text-white uppercase tracking-tight">Expressway Sector 4, Mumbai-Pune Hwy</p>
                  </div>
               </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl">
               <p className="text-[9px] font-black text-on-surface-variant tracking-[.2em] uppercase mb-3">Fluid Consumption</p>
               <h4 className="text-3xl font-bold text-white tracking-tight">42.4 <span className="text-[11px] font-black text-on-surface-variant uppercase ml-1">LITRES</span></h4>
            </div>
            <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl">
               <p className="text-[9px] font-black text-on-surface-variant tracking-[.2em] uppercase mb-3">Operational Idle</p>
               <h4 className="text-3xl font-bold text-white tracking-tight">14 <span className="text-[11px] font-black text-on-surface-variant uppercase ml-1">MINUTES</span></h4>
            </div>
            <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl">
               <p className="text-[9px] font-black text-on-surface-variant tracking-[.2em] uppercase mb-3">Compliance: E-Way Bill</p>
               <div className="flex items-center gap-3">
                 <span className="material-symbols-outlined text-emerald-400 text-lg">verified_user</span>
                 <h4 className="text-sm font-black text-white uppercase tracking-[.15em] font-mono">{trip.eway_bill_no || "EWB-2024-9104"}</h4>
               </div>
            </div>
          </div>
        </div>

        {/* Right Column: Timeline (40%) */}
        <div className="lg:col-span-4 flex flex-col gap-8">
          <div className="bg-surface-container rounded-[2.5rem] border border-outline flex flex-col h-full shadow-2xl overflow-hidden">
             <div className="p-8 border-b border-outline bg-surface-container/50 flex items-center justify-between">
                <h3 className="text-[11px] font-black text-white tracking-[0.2em] uppercase flex items-center gap-3">
                  <span className="material-symbols-outlined text-lg text-primary">history_toggle_off</span>
                  Event Tracking
                </h3>
                <span className="px-3 py-1 bg-primary text-white text-[9px] font-black rounded-lg uppercase tracking-widest shadow-md shadow-primary/10">Live Stream</span>
             </div>
             
             <div className="p-8 flex-1 overflow-y-auto custom-scrollbar">
                <div className="relative">
                  {/* Timeline Backbone */}
                  <div className="absolute left-7 top-0 bottom-0 w-px bg-outline" />
                  
                  <div className="space-y-12 relative">
                     {/* Pickup Event */}
                     <div className="flex items-start gap-6">
                        <div className="relative z-10 w-14 h-14 flex items-center justify-center bg-surface-dim border border-emerald-500/30 rounded-2xl shadow-xl">
                           <span className="material-symbols-outlined text-emerald-400 text-2xl">local_shipping</span>
                        </div>
                        <div className="flex-1 pt-1">
                           <p className="text-[9px] font-black text-emerald-400 tracking-widest uppercase mb-1">Operational Start</p>
                           <h4 className="text-base font-bold text-white tracking-tight mb-1">Navi Mumbai Regional Hub</h4>
                           <p className="text-xs text-on-surface-variant font-bold">{format(new Date(trip.scheduled_start), 'MMM d, h:mm a')}</p>
                        </div>
                     </div>

                     {/* Uniform Speed Log */}
                     <div className="flex items-start gap-6 opacity-60">
                        <div className="relative z-10 w-14 h-14 flex items-center justify-center bg-surface-dim border border-outline rounded-2xl">
                           <span className="material-symbols-outlined text-on-surface-variant text-xl">speed</span>
                        </div>
                        <div className="flex-1 pt-2">
                           <p className="text-[9px] font-black text-on-surface-variant tracking-widest uppercase mb-1">Transit Maintenance</p>
                           <p className="text-xs text-on-surface-variant font-bold">Standardized 60km/h velocity</p>
                        </div>
                     </div>

                     {/* Alert Event */}
                     <div className="flex items-start gap-6">
                        <div className="relative z-10 w-14 h-14 flex items-center justify-center bg-surface-dim border border-rose-500/30 rounded-2xl shadow-xl">
                           <span className="material-symbols-outlined text-rose-400 text-2xl">warning</span>
                        </div>
                        <div className="flex-1 pt-1">
                           <p className="text-[9px] font-black text-rose-400 tracking-widest uppercase mb-1">Policy Violation</p>
                           <h4 className="text-base font-bold text-white tracking-tight mb-1 uppercase tracking-wider">Velocity: 82 KM/H</h4>
                           <p className="text-xs text-rose-400/80 font-bold">Automatic intervention warning issued</p>
                        </div>
                     </div>

                     {/* Fuel Event */}
                     <div className="flex items-start gap-6">
                        <div className="relative z-10 w-14 h-14 flex items-center justify-center bg-surface-dim border border-amber-500/30 rounded-2xl shadow-xl">
                           <span className="material-symbols-outlined text-amber-400 text-2xl">ev_station</span>
                        </div>
                        <div className="flex-1 pt-1">
                           <p className="text-[9px] font-black text-amber-400 tracking-widest uppercase mb-1">Resource Infusion</p>
                           <h4 className="text-sm font-bold text-white mb-1 uppercase tracking-widest">22.4L Refuel Authenticated</h4>
                           <p className="text-xs text-on-surface-variant font-bold">BPCL Station, Panvel Expressway</p>
                        </div>
                     </div>
                  </div>
                </div>
             </div>

             {/* Footer with Driver Quick Info */}
             <div className="p-8 bg-surface-dim/50 border-t border-outline flex items-center justify-between">
                <div className="flex items-center gap-4">
                   <div className="w-14 h-14 rounded-2xl bg-surface-container border border-outline flex items-center justify-center shadow-inner">
                     <span className="material-symbols-outlined text-on-surface-variant text-2xl">account_circle</span>
                   </div>
                   <div>
                     <p className="text-[9px] font-black text-on-surface-variant tracking-[0.2em] uppercase mb-1">Assigned Personnel</p>
                     <p className="text-base font-bold text-white tracking-tight">Rajesh Kumar Prajapati</p>
                   </div>
                </div>
                <button className="h-12 w-12 rounded-xl bg-primary/10 text-primary hover:bg-primary transition-all flex items-center justify-center hover:text-white border border-primary/20 active:scale-90 shadow-lg">
                  <span className="material-symbols-outlined text-xl">contact_phone</span>
                </button>
             </div>
          </div>
        </div>

      </div>
    </div>
  );
};

export default TripDetail;
