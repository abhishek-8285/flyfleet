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

export const DriverDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [driver, setDriver] = useState<any>(null);
  const [loading, setLoading] = useState(true);

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
      <div className="bg-surface-dark border border-border-color rounded-[2rem] p-10 flex flex-col md:flex-row items-center gap-10 shadow-2xl relative overflow-hidden">
        <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full -mr-32 -mt-32 blur-3xl" />
        
        <div className="w-40 h-40 rounded-[2.5rem] bg-gradient-to-br from-primary/20 to-primary/5 border border-primary/20 flex items-center justify-center text-5xl font-black text-primary shadow-inner">
          {(driver.name || driver.first_name)?.[0]}{(driver.last_name || "")?.[0]}
        </div>

        <div className="flex-1 text-center md:text-left space-y-4">
          <div>
            <div className="flex items-center justify-center md:justify-start gap-4 mb-2">
                <h1 className="text-4xl font-black text-white tracking-tight">{driver.name || `${driver.first_name} ${driver.last_name}`}</h1>
                <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest ${driver.status === 'active' ? 'bg-success/10 text-success' : 'bg-primary/10 text-primary'}`}>
                  {driver.status || 'ACTIVE'}
                </span>
            </div>
            <p className="text-text-muted font-medium flex items-center justify-center md:justify-start gap-2">
                <Briefcase className="w-4 h-4" />
                {driver.employment_type || 'Full Time'} • Personnel Node {driver.employee_code || '#FF-001'}
            </p>
          </div>

          <div className="flex flex-wrap items-center justify-center md:justify-start gap-6 pt-2">
             <div className="flex items-center gap-2 text-sm text-white/70">
                <Phone className="w-4 h-4 text-primary" />
                {driver.phone}
             </div>
             <div className="flex items-center gap-2 text-sm text-white/70">
                <Mail className="w-4 h-4 text-primary" />
                {driver.email || 'No email contact'}
             </div>
             <div className="flex items-center gap-2 text-sm text-white/70">
                <MapPin className="w-4 h-4 text-primary" />
                {driver.address?.city || 'No Address Signal'}
             </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="md:col-span-2 space-y-8">
            <div className="bg-surface-dark border border-border-color p-8 rounded-3xl space-y-8">
               <h3 className="text-xs font-bold text-white uppercase tracking-[0.2em] flex items-center gap-2">
                  <ShieldCheck className="w-4 h-4 text-primary" /> Compliance & Credentials
               </h3>
               <div className="grid grid-cols-2 gap-10">
                  <div>
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest mb-1">License ID</p>
                    <p className="text-lg font-bold text-white font-mono">{driver.license_number || 'N/A'}</p>
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest mb-1">Registry Class</p>
                    <p className="text-lg font-bold text-white uppercase">{driver.license_type || 'LMV'}</p>
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest mb-1">Aadhaar Signature</p>
                    <p className="text-lg font-bold text-white font-mono">{driver.aadhaar_number || 'N/A'}</p>
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest mb-1">Taxation Index (PAN)</p>
                    <p className="text-lg font-bold text-white font-mono">{driver.pan_number || 'N/A'}</p>
                  </div>
               </div>
            </div>

            <div className="grid grid-cols-2 gap-6">
                <div className="bg-surface-dark border border-border-color p-6 rounded-3xl space-y-1">
                   <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Tenure Points</p>
                   <h4 className="text-3xl font-black text-white">{driver.experience_years}<span className="text-sm font-medium text-text-muted ml-2">Years</span></h4>
                </div>
                <div className="bg-surface-dark border border-border-color p-6 rounded-3xl space-y-1">
                   <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Experience</p>
                   <h4 className="text-3xl font-black text-success">98.4<span className="text-sm font-medium text-text-muted ml-2">PHI</span></h4>
                </div>
            </div>
        </div>

        <div className="space-y-8">
            <div className="bg-surface-dark border border-border-color p-8 rounded-3xl space-y-6">
               <h3 className="text-xs font-bold text-white uppercase tracking-[0.2em]">Bio Telemetry</h3>
               <div className="space-y-4">
                  <div className="flex items-center justify-between">
                     <span className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Blood Type</span>
                     <span className="text-sm font-black text-primary">{driver.metadata?.blood_group || 'O+'}</span>
                  </div>
                  <div className="flex items-center justify-between">
                     <span className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Shift Protocol</span>
                     <span className="text-sm font-black text-white uppercase">{driver.metadata?.shift_type || 'Morning'}</span>
                  </div>
                  <div className="flex items-center justify-between">
                     <span className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Neural Link</span>
                     <span className={`text-[10px] font-bold uppercase ${driver.metadata?.behavior_monitoring ? 'text-success' : 'text-primary'}`}>
                        {driver.metadata?.behavior_monitoring ? 'ACTIVE' : 'OFFLINE'}
                     </span>
                  </div>
               </div>
            </div>

            <div className="bg-primary/5 border border-primary/20 p-8 rounded-3xl space-y-4">
               <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center">
                  <Truck className="w-5 h-5 text-primary" />
               </div>
               <h4 className="text-xs font-bold text-white uppercase tracking-widest">Assigned Asset</h4>
               {driver.current_vehicle_id ? (
                    <p className="text-sm text-white font-bold">{driver.current_vehicle_id}</p>
               ) : (
                    <p className="text-sm text-text-muted">No vehicle currently linked.</p>
               )}
            </div>
        </div>
    </div>
    </div>
  );
};

export default DriverDetail;
