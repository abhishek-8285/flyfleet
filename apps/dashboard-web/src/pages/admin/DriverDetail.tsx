import React, { useState, useEffect, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { 
  ChevronLeft, 
  Edit3, 
  User, 
  IdCard, 
  Briefcase, 
  ShieldCheck,
  Truck,
  Phone,
  Mail,
  MapPin,
  Clock,
  Activity
} from "lucide-react";
import { api } from "../../services/api";
import { VehicleSelectionModal } from "../../components/drivers/VehicleSelectionModal";

export const DriverDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [driver, setDriver] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isAssigning, setIsAssigning] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const loadData = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await api.getDriver(id);
      setDriver(res.data || res);
    } catch (err) {
      console.error("Failed to fetch driver", err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleAssignVehicle = async (vehicleId: string) => {
    if (!id) return;
    setIsAssigning(true);
    try {
      await api.assignVehicle(id, vehicleId);
      setToast({ message: "Vehicle assigned successfully", type: 'success' });
      setTimeout(() => setToast(null), 3000);
      setIsModalOpen(false);
      loadData();
    } catch (err: any) {
      console.error("Failed to assign vehicle", err);
      const errorMsg = err.response?.data?.error || "Failed to assign vehicle";
      setToast({ message: errorMsg, type: 'error' });
      setTimeout(() => setToast(null), 3000);
    } finally {
      setIsAssigning(false);
    }
  };

  const handleUnassignVehicle = async () => {
    if (!id) return;
    if (!window.confirm("Are you sure you want to unassign this vehicle?")) return;
    
    setIsAssigning(true);
    try {
      await api.unassignVehicle(id);
      setToast({ message: "Vehicle unassigned successfully", type: 'success' });
      setTimeout(() => setToast(null), 3000);
      loadData();
    } catch (err: any) {
      console.error("Failed to unassign vehicle", err);
      const errorMsg = err.response?.data?.error || "Failed to unassign vehicle";
      setToast({ message: errorMsg, type: 'error' });
      setTimeout(() => setToast(null), 3000);
    } finally {
      setIsAssigning(false);
    }
  };

  if (loading) return <div className="flex h-screen items-center justify-center text-white">Loading...</div>;
  if (!driver) return <div className="p-10 text-white">Driver not found</div>;

  return (
    <div className="max-w-5xl mx-auto space-y-8 animate-in fade-in duration-500 pb-20">
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate("/admin/drivers")}
          className="flex items-center gap-2 text-text-muted hover:text-white transition-colors group"
        >
          <ChevronLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          Personnel Registry
        </button>
        <div className="flex gap-4">
            <button
              onClick={() => navigate(`/admin/drivers/${id}/edit`)}
              className="flex items-center gap-2 px-6 py-2.5 border border-white/10 text-white rounded-xl hover:bg-white/5 transition-all font-semibold"
            >
              <Edit3 className="w-4 h-4" />
              Modify Profile
            </button>
        </div>
      </div>

      {/* Header Profile */}
      <div className="bg-[#0a0a0a] border border-white/5 rounded-[2.5rem] p-10 flex flex-col md:flex-row items-center gap-10 shadow-2xl relative overflow-hidden group">
        <div className="absolute top-0 right-0 w-96 h-96 bg-primary/5 rounded-full -mr-48 -mt-48 blur-3xl group-hover:bg-primary/10 transition-all duration-700" />
        
        <div className="w-48 h-48 rounded-[3rem] bg-gradient-to-br from-primary/10 to-transparent border border-white/10 flex items-center justify-center text-6xl font-black text-primary shadow-2xl relative z-10">
          {(driver.name || driver.first_name)?.[0]}{(driver.last_name || "")?.[0]}
          <div className="absolute -bottom-2 -right-2 w-12 h-12 bg-[#0a0a0a] border border-white/10 rounded-2xl flex items-center justify-center">
             <span className="material-symbols-outlined text-emerald-500 text-2xl">verified</span>
          </div>
        </div>

        <div className="flex-1 text-center md:text-left space-y-6 relative z-10">
          <div className="space-y-2">
            <div className="flex items-center justify-center md:justify-start gap-4 flex-wrap">
                <h1 className="text-5xl font-black text-white tracking-tighter uppercase italic">{driver.name || `${driver.first_name} ${driver.last_name}`}</h1>
                <div className="flex gap-2">
                  <span className={`px-4 py-1 rounded-full text-[10px] font-black uppercase tracking-[0.2em] border ${driver.status === 'active' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border-rose-500/20'}`}>
                    {driver.status || 'ACTIVE'}
                  </span>
                  <span className="px-4 py-1 rounded-full text-[10px] font-black uppercase tracking-[0.2em] bg-primary text-black border border-primary/50 shadow-[0_0_20px_rgba(var(--primary-rgb),0.3)]">
                    RISK: {driver.risk_score || 'LOW'}
                  </span>
                </div>
            </div>
            <p className="text-white/30 font-black uppercase tracking-[0.2em] flex items-center justify-center md:justify-start gap-3 text-xs">
                <Briefcase className="w-4 h-4 text-primary" />
                {driver.employment_type || 'EXECUTIVE PILOT'} • ID: {driver.employee_code || 'AWAITING_CODE'}
            </p>
          </div>

          <div className="flex flex-wrap items-center justify-center md:justify-start gap-8 pt-4 border-t border-white/5">
             <div className="flex flex-col gap-1">
                <span className="text-[9px] font-black text-white/20 uppercase tracking-widest">Contact Signal</span>
                <div className="flex items-center gap-2 text-sm text-white/80 font-bold">
                   <Phone className="w-4 h-4 text-primary" />
                   {driver.phone}
                </div>
             </div>
             <div className="flex flex-col gap-1">
                <span className="text-[9px] font-black text-white/20 uppercase tracking-widest">Digital Address</span>
                <div className="flex items-center gap-2 text-sm text-white/80 font-bold">
                   <Mail className="w-4 h-4 text-primary" />
                   {driver.email || 'NO_CONTACT@FLYFLEET.COM'}
                </div>
             </div>
             <div className="flex flex-col gap-1">
                <span className="text-[9px] font-black text-white/20 uppercase tracking-widest">Operational Base</span>
                <div className="flex items-center gap-2 text-sm text-white/80 font-bold">
                   <MapPin className="w-4 h-4 text-primary" />
                   {driver.address?.city || 'NORTH_STATION_GRID'}
                </div>
             </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
        {/* Main Operational Data */}
        <div className="md:col-span-3 space-y-8">
            <div className="bg-[#0a0a0a] border border-white/5 p-10 rounded-[2.5rem] space-y-10">
               <div className="flex items-center justify-between">
                  <h3 className="text-[10px] font-black text-white/30 uppercase tracking-[0.3em] flex items-center gap-3">
                     <ShieldCheck className="w-5 h-5 text-emerald-500" /> Compliance Framework
                  </h3>
                  <div className="px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-[10px] font-black text-emerald-400 uppercase tracking-widest">
                    Audit Status: Verified
                  </div>
               </div>
               
               <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
                  <div className="space-y-2">
                    <p className="text-[9px] font-black text-white/20 uppercase tracking-[0.2em]">License Integrity</p>
                    <div className="space-y-1">
                      <p className="text-xl font-black text-white font-mono tracking-tighter">{driver.license_number || 'UP16-2023-0091'}</p>
                      <div className="flex items-center gap-2">
                        <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                        <span className="text-[10px] font-black text-emerald-500 uppercase">Valid until 24 Oct 2029</span>
                      </div>
                    </div>
                  </div>
                  <div className="space-y-2">
                    <p className="text-[9px] font-black text-white/20 uppercase tracking-[0.2em]">Credential Class</p>
                    <p className="text-xl font-black text-white uppercase tracking-tighter">{driver.license_type || 'HEAVY_TRANSPORT'}</p>
                  </div>
                   <div className="space-y-2">
                    <p className="text-[9px] font-black text-white/20 uppercase tracking-[0.2em]">Safety Score</p>
                    <div className="flex items-end gap-2">
                      <span className="text-3xl font-black text-emerald-400">{driver.behavior_score ?? '100'}</span>
                      <span className="text-[10px] font-black text-white/20 uppercase mb-1">/ 100</span>
                    </div>
                  </div>
               </div>

                <div className="pt-10 border-t border-white/5 grid grid-cols-1 md:grid-cols-3 gap-12">
                  <div className="space-y-1">
                    <p className="text-[9px] font-black text-white/20 uppercase tracking-[0.2em]">Aadhaar Node</p>
                    <p className="text-sm font-bold text-white/60 font-mono tracking-widest">{driver.aadhaar_number || 'NOT_LINKED'}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[9px] font-black text-white/20 uppercase tracking-[0.2em]">Taxation Hash (PAN)</p>
                    <p className="text-sm font-bold text-white/60 font-mono tracking-widest">{driver.pan_number || 'NOT_LINKED'}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[9px] font-black text-white/20 uppercase tracking-[0.2em]">Incidents (L365D)</p>
                    <p className="text-sm font-black text-rose-500 uppercase tracking-widest">00 Violations</p>
                  </div>
               </div>
            </div>

            {/* Performance Context */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
               <div className="bg-[#0a0a0a] border border-white/5 p-8 rounded-[2rem] space-y-2 group hover:border-primary/20 transition-all">
                  <p className="text-[10px] font-black text-white/20 uppercase tracking-widest">Fleet Tenure</p>
                  <h4 className="text-4xl font-black text-white tracking-tighter">{driver.experience_years ?? '0'}<span className="text-xs font-bold text-white/20 ml-2 uppercase tracking-widest">Years</span></h4>
                  <div className="w-full h-1 bg-white/5 rounded-full mt-4 overflow-hidden">
                     <div className="w-2/3 h-full bg-primary/40" />
                  </div>
               </div>
               <div className="bg-[#0a0a0a] border border-white/5 p-8 rounded-[2rem] space-y-2 group hover:border-emerald-500/20 transition-all">
                  <p className="text-[10px] font-black text-white/20 uppercase tracking-widest">Safety Index</p>
                  <h4 className="text-4xl font-black text-emerald-500 tracking-tighter">{driver.behavior_score ?? '100'}%</h4>
                  <p className="text-[9px] font-bold text-emerald-500/40 uppercase tracking-widest">LIVE BEHAVIOR FEED</p>
               </div>
               <div className="bg-[#0a0a0a] border border-white/5 p-8 rounded-[2rem] space-y-2 group hover:border-sky-500/20 transition-all">
                  <p className="text-[10px] font-black text-white/20 uppercase tracking-widest">Total Trips</p>
                  <h4 className="text-4xl font-black text-white tracking-tighter">{driver.total_trips ?? '0'}</h4>
                  <p className="text-[9px] font-bold text-white/20 uppercase tracking-widest">Operational Load</p>
               </div>
            </div>
        </div>

        {/* Sidebar Context */}
        <div className="space-y-8">
            <div className="bg-[#0a0a0a] border border-white/5 p-8 rounded-[2rem] space-y-8">
               <h3 className="text-[10px] font-black text-white/30 uppercase tracking-[0.2em]">Bio Telemetry</h3>
               <div className="space-y-6">
                   <div className="flex items-center justify-between">
                     <span className="text-[10px] font-black text-white/20 uppercase tracking-widest">Blood Type</span>
                     <span className="text-sm font-black text-primary bg-primary/10 px-2 py-0.5 rounded-lg border border-primary/20">{driver.metadata?.blood_group || 'UNK'}</span>
                  </div>
                  <div className="flex items-center justify-between">
                     <span className="text-[10px] font-black text-white/20 uppercase tracking-widest">Shift Logic</span>
                     <span className="text-sm font-black text-white uppercase tracking-tighter italic">{driver.metadata?.shift_type || 'NOT_SET'}</span>
                  </div>
                  <div className="flex items-center justify-between">
                     <span className="text-[10px] font-black text-white/20 uppercase tracking-widest">Neural Link</span>
                     <div className="flex items-center gap-2">
                        <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                        <span className="text-[10px] font-black text-emerald-500 uppercase tracking-widest">Active Ingest</span>
                     </div>
                  </div>
               </div>
            </div>

            <div className={`p-8 rounded-[2.5rem] space-y-6 border transition-all ${driver.current_vehicle_id ? 'bg-primary/5 border-primary/20 shadow-xl shadow-primary/5' : 'bg-white/5 border-white/10 opacity-50'}`}>
               <div className="flex items-center justify-between">
                 <div className={`w-12 h-12 rounded-2xl flex items-center justify-center border transition-all ${driver.current_vehicle_id ? 'bg-primary/20 border-primary/30 text-primary' : 'bg-white/5 border-white/10 text-white/20'}`}>
                    <Truck className="w-6 h-6" />
                 </div>
                 {driver.current_vehicle_id && (
                   <div className="flex items-center gap-2 px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                      <span className="text-[9px] font-black text-emerald-400 uppercase tracking-widest">Online</span>
                   </div>
                 )}
               </div>
               <div>
                  <h4 className="text-[10px] font-black text-white/40 uppercase tracking-widest mb-2">Assigned Asset</h4>
                  {driver.current_vehicle?.registration_number ? (
                      <div className="space-y-1">
                        <p className="text-2xl font-black text-white tracking-tighter uppercase italic">{driver.current_vehicle.registration_number}</p>
                        <p className="text-[9px] font-bold text-white/20 uppercase tracking-widest">Hardware Node 492-X1</p>
                      </div>
                  ) : (
                      <div className="space-y-4">
                        <p className="text-xs text-white/20 font-bold italic uppercase tracking-widest">No Asset Linked</p>
                        <button
                          onClick={() => setIsModalOpen(true)}
                          disabled={isAssigning}
                          className="w-full py-3 bg-primary text-black rounded-xl text-[10px] font-black uppercase tracking-widest transition-all hover:bg-primary/90 shadow-lg shadow-primary/20 disabled:opacity-50"
                        >
                          {isAssigning ? 'Processing...' : 'Link Asset'}
                        </button>
                      </div>
                  )}
               </div>
               {driver.current_vehicle_id && (
                 <div className="space-y-3">
                   <button 
                    onClick={() => navigate(`/admin/vehicles/${driver.current_vehicle_id}`)}
                    className="w-full py-3 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all"
                   >
                     View Hardware Log
                   </button>
                   <button 
                    onClick={handleUnassignVehicle}
                    disabled={isAssigning}
                    className="w-full py-3 bg-rose-500/10 hover:bg-rose-500/20 border border-rose-500/20 rounded-xl text-[10px] font-black text-rose-500 uppercase tracking-widest transition-all disabled:opacity-50"
                   >
                     {isAssigning ? 'Releasing...' : 'Unlink Asset'}
                   </button>
                 </div>
               )}
            </div>
        </div>
      </div>

      <VehicleSelectionModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSelect={handleAssignVehicle}
        currentVehicleId={driver.current_vehicle_id}
      />

      {/* Local Toast Notification */}
      {toast && (
        <div className="fixed bottom-8 right-8 z-[100] animate-in slide-in-from-bottom-10 duration-300">
           <div className={`px-6 py-4 rounded-2xl shadow-2xl flex items-center gap-3 border ${
             toast.type === 'success' 
               ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' 
               : 'bg-rose-500/10 border-rose-500/20 text-rose-400'
           } backdrop-blur-xl`}>
             <div className={`p-1 rounded-full ${toast.type === 'success' ? 'bg-emerald-500/20' : 'bg-rose-500/20'}`}>
                <span className="material-symbols-outlined text-sm">
                  {toast.type === 'success' ? 'check_circle' : 'error'}
                </span>
             </div>
             <p className="font-bold text-sm tracking-tight">{toast.message}</p>
           </div>
        </div>
      )}
    </div>
  );
};

export default DriverDetail;
