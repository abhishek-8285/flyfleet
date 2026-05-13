import React, { useEffect, useState, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { 
  Truck, 
  ChevronLeft, 
  Edit3, 
  Signal, 
  ShieldCheck, 
  AlertCircle,
  Clock,
  Activity,
  User,
  History,
  FileText
} from "lucide-react";
import { api } from "../../services/api";
import { DriverSelectionModal } from "../../components/vehicles/DriverSelectionModal";

interface Vehicle {
  id: string;
  registration_number: string;
  make: string;
  model: string;
  status: string;
  current_driver_id?: string;
  current_driver_name?: string;
  current_driver?: {
    id: string;
    name: string;
  };
  summary?: {
    compliance_status?: string;
    gps_status?: string;
    last_seen?: string;
    risk_score?: number;
    risk_category?: string;
  };
  metadata?: any;
}

export const VehicleDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const loadVehicle = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await api.getVehicle(id);
      setVehicle(res?.data || res);
    } catch (error) {
      console.error("Failed to fetch vehicle:", error);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadVehicle();
  }, [loadVehicle]);

  const handleAssignDriver = async (driverId: string) => {
    if (!id) return;
    setIsProcessing(true);
    try {
      await api.assignVehicle(driverId, id);
      setToast({ message: "Pilot linked successfully", type: 'success' });
      setIsModalOpen(false);
      loadVehicle();
    } catch (error: any) {
      console.error("Failed to assign driver:", error);
      setToast({ message: error.response?.data?.error || "Failed to link pilot", type: 'error' });
    } finally {
      setIsProcessing(false);
      setTimeout(() => setToast(null), 3000);
    }
  };

  const handleUnassignDriver = async () => {
    if (!id || !vehicle?.current_driver_id) return;
    if (!window.confirm("Are you sure you want to unlink this pilot?")) return;

    setIsProcessing(true);
    try {
      await api.unassignVehicle(vehicle.current_driver_id);
      setToast({ message: "Pilot unlinked successfully", type: 'success' });
      loadVehicle();
    } catch (error: any) {
      console.error("Failed to unassign driver:", error);
      setToast({ message: error.response?.data?.error || "Failed to unlink pilot", type: 'error' });
    } finally {
      setIsProcessing(false);
      setTimeout(() => setToast(null), 3000);
    }
  };

  if (loading) return (
    <div className="flex h-screen items-center justify-center bg-surface-dark text-white">
      <div className="flex flex-col items-center gap-4">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary" />
        <p className="text-gray-400 text-xs font-bold uppercase tracking-widest">Resolving Asset Profile...</p>
      </div>
    </div>
  );

  if (!vehicle) return (
    <div className="flex h-screen items-center justify-center bg-surface-dark">
      <div className="text-center">
        <AlertCircle className="w-12 h-12 text-rose-500 mx-auto mb-4" />
        <h2 className="text-xl font-bold text-white">Asset Not Found</h2>
        <p className="text-gray-500 mt-2">The requested vehicle record does not exist.</p>
        <button 
          onClick={() => navigate("/admin/vehicles")}
          className="mt-6 px-6 py-2 bg-white/5 border border-white/10 text-white rounded-lg hover:bg-white/10 transition-all text-sm font-semibold"
        >
          Return to Registry
        </button>
      </div>
    </div>
  );

  const summary = vehicle.summary || {};
  const driverName = vehicle.current_driver?.name || vehicle.current_driver_name;
  const compliance = summary.compliance_status || "UNKNOWN";
  const gpsStatus = summary.gps_status || "OFFLINE";
  const lastSeen = summary.last_seen;
  const riskScore = summary.risk_score || 0;
  const status = vehicle.status?.toLowerCase();

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto pb-20">
      {/* Navigation & Actions */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate("/admin/vehicles")}
          className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors group text-sm font-semibold"
        >
          <ChevronLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
          Asset Registry
        </button>
        <div className="flex gap-3">
          <button className="flex items-center gap-2 px-4 py-2 bg-white/5 border border-white/10 text-white rounded-lg hover:bg-white/10 transition-all text-sm font-semibold">
            <History className="w-4 h-4" />
            Audit Log
          </button>
          <button 
            onClick={() => navigate(`/admin/vehicles/${id}/edit`)}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all text-sm font-bold shadow-lg shadow-primary/20"
          >
            <Edit3 className="w-4 h-4" />
            Modify Asset
          </button>
        </div>
      </div>

      {/* Header Info - Sober Style */}
      <div className="bg-surface-dark border border-border-color rounded-2xl p-8 flex items-center gap-8 shadow-2xl">
        <div className="w-24 h-24 rounded-2xl bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
          <Truck className="w-12 h-12" />
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-4">
            <h1 className="text-3xl font-bold text-white tracking-tight">
              {vehicle.registration_number}
            </h1>
            <span className={`px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-widest border ${
              status === 'active' ? 'bg-success/10 text-success border-success/20' : 'bg-white/5 text-gray-500 border-white/10'
            }`}>
              {vehicle.status}
            </span>
          </div>
          <p className="text-gray-400 font-semibold mt-1 uppercase tracking-wider text-sm">
            {vehicle.make} {vehicle.model} • <span className="font-mono text-gray-500">{vehicle.id.slice(0, 8)}</span>
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Compliance & Health */}
        <div className="md:col-span-2 space-y-6">
          <div className="bg-surface-dark border border-border-color p-6 rounded-2xl space-y-6">
            <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-2 pb-4 border-b border-border-color">
              <ShieldCheck className="w-4 h-4 text-primary" /> Operational Health
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
               <div className="space-y-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">Compliance Status</p>
                  <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${compliance === 'GREEN' ? 'bg-success' : compliance === 'YELLOW' ? 'bg-amber-500' : 'bg-rose-500'}`} />
                    <span className={`text-lg font-bold ${compliance === 'GREEN' ? 'text-success' : 'text-rose-500'}`}>
                      {compliance}
                    </span>
                  </div>
               </div>
               <div className="space-y-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">GPS Connectivity</p>
                  <div className="flex items-center gap-2 text-white font-bold text-lg">
                    <Signal className={`w-5 h-5 ${gpsStatus === 'ONLINE' ? 'text-success' : 'text-gray-600'}`} />
                    {gpsStatus}
                  </div>
               </div>
               <div className="space-y-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">Safety Index</p>
                  <div className="flex items-end gap-1">
                    <span className="text-2xl font-bold text-white">{riskScore}%</span>
                    <span className="text-[10px] text-gray-500 font-bold mb-1 uppercase">Risk</span>
                  </div>
               </div>
            </div>
          </div>

          {/* Detailed Specs / Documents */}
          <div className="bg-surface-dark border border-border-color p-6 rounded-2xl space-y-4">
             <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-2">
                <FileText className="w-4 h-4 text-primary" /> Registered Documents
             </h3>
             <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex items-center justify-between p-4 bg-white/5 rounded-xl border border-border-color">
                   <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center text-gray-400">RC</div>
                      <span className="text-sm font-semibold text-gray-200">Registration Certificate</span>
                   </div>
                   <span className="text-[10px] font-bold text-success uppercase">Valid</span>
                </div>
                <div className="flex items-center justify-between p-4 bg-white/5 rounded-xl border border-border-color">
                   <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center text-gray-400">INS</div>
                      <span className="text-sm font-semibold text-gray-200">Insurance Policy</span>
                   </div>
                   <span className="text-[10px] font-bold text-amber-500 uppercase">Expiring in 12d</span>
                </div>
             </div>
          </div>
        </div>

        {/* Assignments & Telemetry Sidebar */}
        <div className="space-y-6">
           <div className="bg-surface-dark border border-border-color p-6 rounded-2xl space-y-6">
              <h3 className="text-xs font-bold text-white uppercase tracking-widest">Active Assignment</h3>
              {driverName ? (
                <div className="space-y-4">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold">
                      {driverName[0]}
                    </div>
                    <div>
                      <div className="font-bold text-white">{driverName}</div>
                      <p className="text-[10px] text-gray-500 uppercase font-bold tracking-widest mt-0.5">Assigned Pilot</p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button 
                      onClick={() => navigate(`/admin/drivers/${vehicle.current_driver?.id || vehicle.current_driver_id}`)}
                      className="flex-1 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg text-[10px] font-bold text-white uppercase tracking-widest transition-all"
                    >
                      View Profile
                    </button>
                    <button 
                      onClick={handleUnassignDriver}
                      disabled={isProcessing}
                      className="flex-1 py-2 bg-rose-500/10 hover:bg-rose-500/20 border border-rose-500/20 rounded-lg text-[10px] font-bold text-rose-500 uppercase tracking-widest transition-all disabled:opacity-50"
                    >
                      {isProcessing ? 'Releasing...' : 'Unlink Pilot'}
                    </button>
                  </div>
                </div>
              ) : (
                <div className="text-center py-6 bg-white/5 rounded-xl border border-dashed border-border-color">
                   <User className="w-8 h-8 text-gray-700 mx-auto mb-2" />
                   <p className="text-[10px] font-bold text-gray-600 uppercase">No Pilot Assigned</p>
                   <button 
                    onClick={() => setIsModalOpen(true)}
                    className="mt-3 text-[10px] font-bold text-primary uppercase hover:underline"
                   >
                    Link Pilot
                   </button>
                </div>
              )}
           </div>

           <div className="bg-surface-dark border border-border-color p-6 rounded-2xl space-y-4">
              <div className="flex items-center justify-between">
                 <h3 className="text-xs font-bold text-white uppercase tracking-widest">Last Activity</h3>
                 <Clock className="w-4 h-4 text-gray-600" />
              </div>
              <div className="space-y-4">
                 <div className="flex justify-between items-center">
                    <span className="text-[10px] font-bold text-gray-500 uppercase">Telemetry Signal</span>
                    <span className="text-xs font-semibold text-gray-300">
                      {lastSeen ? new Date(lastSeen).toLocaleTimeString() : 'NEVER'}
                    </span>
                 </div>
                 <div className="flex justify-between items-center">
                    <span className="text-[10px] font-bold text-gray-500 uppercase">System Status</span>
                    <span className={`text-xs font-semibold uppercase ${gpsStatus === 'ONLINE' ? 'text-success' : 'text-gray-500'}`}>
                      {gpsStatus}
                    </span>
                 </div>
              </div>
           </div>
        </div>
      </div>

      <DriverSelectionModal 
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSelect={handleAssignDriver}
        currentDriverId={vehicle.current_driver?.id || vehicle.current_driver_id}
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
