import React, { useEffect, useState, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useSelector } from "react-redux";
import { RootState } from "../../../store";
import apiClient from "../../../services/apiClient";
import LiveOperationsMap from "../../../modules/dashboard/components/LiveOperationsMap";
import { STATUS_COLORS } from "./constants";
import { 
  ShieldCheck, 
  Zap, 
  User, 
  Calendar, 
  Activity, 
  Settings, 
  Truck, 
  Signal, 
  Toolbox, 
  Lock, 
  Unlock, 
  Plus,
  RefreshCw,
  Ban,
  ArrowLeft,
  UserPlus,
  FileText,
  MapPin,
  ChevronRight,
  Eye,
  Download,
  Upload,
  History,
  CheckCircle2,
  AlertTriangle,
  History as AuditIcon,
  Wifi,
  HardDrive,
  Clock,
  Terminal,
  Cpu,
  X,
  ShieldAlert,
  Search,
  Loader2,
  ChevronDown,
  Info,
  UserMinus
} from "lucide-react";

const ZERO_UUID = '00000000-0000-0000-0000-000000000000';

// --- Types ---
interface ComplianceDoc {
  id: string;
  type: string;
  status: 'VALID' | 'EXPIRING' | 'EXPIRED' | 'MISSING' | 'REJECTED' | 'PENDING_VERIFICATION';
  pulse: string;
  impact: string;
  meta: string;
  color: 'emerald' | 'amber' | 'rose' | 'gray';
  mandatory: boolean;
  version: string;
}

interface AuditLog {
  id: string;
  time: string;
  event: string;
  actor: string;
  desc: string;
  severity: 'normal' | 'critical' | 'warn';
  nodeId: string;
  sessionId: string;
}

type UserRole = 'FLYFLEET_CENTRAL_OPS' | 'FLEET_ADMIN' | 'FLEET_OPERATOR';

type CommandType = 'assign' | 'maintenance' | 'block' | 'upload' | 'provision' | 'builder' | 'governance' | 'none';

export const VehicleDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  
  // --- Operational States ---
  const [vehicle, setVehicle] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  const [activeTab, setActiveTab] = useState<"overview" | "documents">("overview");
  const [currentDriver, setCurrentDriver] = useState<any>(null);
  const [command, setCommand] = useState<CommandType>('none');
  const [commandData, setCommandData] = useState<any>(null);
  const [userRole, setUserRole] = useState<UserRole>('FLEET_ADMIN'); // Simulated RBAC context
  const [conflictData, setConflictData] = useState<any>(null);
  const [policyViolation, setPolicyViolation] = useState<any>(null);
  
  // --- Escalation & Recovery State ---
  const [escalations, setEscalations] = useState<any[]>([
    { id: 'E1', type: 'CRITICAL_BLOCKADE', severity: 'critical', desc: 'Fitness Certificate v2.1.4 Expired. Automatic asset suspension active.', timestamp: '18:51:04', workflow: 'compliance' },
    { id: 'E2', type: 'HARDWARE_OFFLINE', severity: 'warn', desc: 'Rolling reliability dropped below 90%. Manual signal check required.', timestamp: '19:12:22', workflow: 'hardware' }
  ]);
  const [showEscalations, setShowEscalations] = useState(false);
  
  // --- Registry & Audit State ---
  const [registryDocs, setRegistryDocs] = useState<ComplianceDoc[]>([
    { id: '1', type: 'Insurance Policy', status: 'VALID', pulse: '420d remaining', impact: 'Dispatch Allowed', meta: 'Verified by HQ Delhi', color: 'emerald', mandatory: true, version: 'v2.4.0' },
    { id: '2', type: 'Pollution (PUC)', status: 'EXPIRING', pulse: '8d remaining', impact: 'Manual Review', meta: 'Last Audit 2d ago', color: 'amber', mandatory: true, version: 'v1.2.1' },
    { id: '3', type: 'Commercial Permit', status: 'VALID', pulse: '124d remaining', impact: 'Dispatch Allowed', meta: 'Verified by Compliance AI', color: 'emerald', mandatory: true, version: 'v3.0.0' },
    { id: '4', type: 'Fitness Certificate', status: 'EXPIRED', pulse: '12d overdue', impact: 'Dispatch Blocked', meta: 'Verification Failed', color: 'rose', mandatory: true, version: 'v2.1.4' },
  ]);
  
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([
    { id: '101', time: '18:51:04', event: 'COMPLIANCE_HOLD', actor: 'SYSTEM', desc: 'Critical document "Fitness Certificate" hit EXPIRED state. Asset suspended.', severity: 'critical', nodeId: 'node-01', sessionId: 'session-01' },
    { id: '102', time: '16:22:12', event: 'DOC_UPLOAD', actor: 'Admin', desc: 'Insurance Policy v2.4 uploaded via Portal.', severity: 'normal', nodeId: 'node-01', sessionId: 'session-01' },
  ]);

  // --- Search & Selection (Pilot Assignment) ---
  const [allDrivers, setAllDrivers] = useState<any[]>([]);
  console.log("[VehicleDetail] Render - allDrivers count:", allDrivers.length);
  const [searchQuery, setSearchQuery] = useState("");
  const filteredDrivers = useMemo(() => {
    console.log("[VehicleDetail] Filtering drivers. Source count:", allDrivers.length, "Query:", searchQuery);
    return allDrivers.filter(d => 
      d.name?.toLowerCase().includes(searchQuery.toLowerCase()) || 
      d.phone?.includes(searchQuery)
    );
  }, [allDrivers, searchQuery]);

  const telemetry = useSelector((state: RootState) => state.telemetry.vehicles[id || ""]);

  // --- Computed Stats ---
  const stats = useMemo(() => {
    const blockedDocs = registryDocs.filter(d => ['EXPIRED', 'REJECTED', 'MISSING'].includes(d.status) && d.mandatory);
    const blockedCount = blockedDocs.length;
    const blockReason = blockedCount > 0 ? blockedDocs[0].type : (vehicle?.status === 'blocked' ? 'Manual Security Hold' : null);
    
    return {
      compliant: registryDocs.filter(d => d.status === 'VALID').length,
      expiring: registryDocs.filter(d => d.status === 'EXPIRING').length,
      blocked: blockedCount,
      isRestricted: blockedCount > 0 || vehicle?.status === 'blocked',
      blockReason
    };
  }, [registryDocs, vehicle?.status]);

  const fetchAuditLogs = async () => {
    try {
      const res: any = await apiClient.get(`/admin/vehicles/${id}/audit`);
      const data = res.data || res;
      if (Array.isArray(data)) {
        const logs: AuditLog[] = data.map((l: any) => ({
          id: l.id,
          time: new Date(l.timestamp).toLocaleTimeString('en-GB', { hour12: false }),
          event: l.action,
          actor: l.user_id ? 'Admin' : 'System',
          desc: l.new_value ? `${l.field_name} updated to ${l.new_value}` : l.action,
          severity: 'normal' as const,
          nodeId: l.id.substring(0, 8).toUpperCase(),
          sessionId: 'BACKEND'
        }));
        setAuditLogs(logs);
      }
    } catch (e) {
      console.error("Failed to fetch audit logs:", e);
    }
  };

  const hasGps = !!vehicle?.gps_device_id;
  const isGpsOnline = hasGps && telemetry?.status !== 'offline';

  const addRegistryRequirement = (config: any) => {
    const newDoc: ComplianceDoc = {
      id: Math.random().toString(36).substr(2, 9),
      type: config.type,
      status: 'MISSING',
      pulse: 'Awaiting Upload',
      impact: config.impact,
      meta: `${config.jurisdiction || 'All'} | ${config.vClass || 'All'}`,
      color: 'gray',
      mandatory: config.mandatory,
      version: 'v1.0.0'
    };
    setRegistryDocs(prev => [...prev, newDoc]);
    addAudit('REQUIREMENT_ADDED', `New ${config.mandatory ? 'Mandatory' : 'Optional'} requirement "${config.type}" defined for ${config.jurisdiction || 'Global'}.`, 'normal');
  };

  useEffect(() => {
    fetchVehicle();
    fetchDrivers();
    fetchAuditLogs();
  }, [id]);

  // Refresh drivers when opening the assign modal
  useEffect(() => {
    if (command === 'assign') {
      fetchDrivers();
    }
  }, [command]);

  const fetchVehicle = async () => {
    try {
      const res = await apiClient.get(`/admin/vehicles/${id}`);
      // Robust extraction: backend returns {"data": vehicle}
      const vData = res.data || res;
      setVehicle(vData);
    } catch (error) { 
      console.error("[VehicleDetail] fetchVehicle error:", error); 
    } finally { 
      setLoading(false); 
      setIsSyncing(false); 
    }
  };

  const fetchDrivers = async () => {
    try {
      console.log("[VehicleDetail] Fetching drivers...");
      const response: any = await apiClient.get("/admin/drivers?limit=500");
      console.log("[VehicleDetail] fetchDrivers raw response:", response);
      
      // Highly robust extraction: handle wrapped, unwrapped, or named property structures
      let ds = [];
      if (response && response.data && Array.isArray(response.data)) {
        ds = response.data;
      } else if (Array.isArray(response)) {
        ds = response;
      } else if (response && response.drivers && Array.isArray(response.drivers)) {
        ds = response.drivers;
      }
      
      console.log("[VehicleDetail] setAllDrivers with count:", ds.length);
      setAllDrivers(ds);
    } catch (e) { 
      console.error("[VehicleDetail] fetchDrivers error:", e); 
    }
  };

  // Sync current driver when either vehicle or drivers list updates
  useEffect(() => {
    if (vehicle?.current_driver_id && allDrivers.length > 0) {
      const d = allDrivers.find((driver: any) => String(driver.id) === String(vehicle.current_driver_id));
      if (d) {
        setCurrentDriver(d);
      }
    }
  }, [vehicle?.current_driver_id, allDrivers]);

  // --- Automated Escalation Surveillance ---
  useEffect(() => {
    const criticalDocs = registryDocs.filter(d => d.status === 'EXPIRED' && d.mandatory);
    if (criticalDocs.length > 0) {
      const doc = criticalDocs[0];
      if (!escalations.find(e => e.id === `DOC_${doc.id}`)) {
        setEscalations(prev => [{
          id: `DOC_${doc.id}`,
          type: 'COMPLIANCE_FAILURE',
          severity: 'critical',
          desc: `${doc.type} hit EXPIRED state. Asset blockade enforced.`,
          timestamp: new Date().toLocaleTimeString('en-GB', { hour12: false }),
          workflow: 'compliance'
        }, ...prev]);
        addAudit('ESCALATION_TRIGGERED', `Automated blockade: ${doc.type} compliance failure.`, 'critical', 'Surveillance_Node');
      }
    } else {
       setEscalations(prev => prev.filter(e => e.type !== 'COMPLIANCE_FAILURE'));
    }
  }, [registryDocs]);

  useEffect(() => {
    if (!isGpsOnline && hasGps && !escalations.find(e => e.type === 'HARDWARE_OFFLINE')) {
      setEscalations(prev => [{
        id: 'H1',
        type: 'HARDWARE_OFFLINE',
        severity: 'warn',
        desc: 'GPS signal lost. Rolling reliability below threshold.',
        timestamp: new Date().toLocaleTimeString('en-GB', { hour12: false }),
        workflow: 'hardware'
      }, ...prev]);
    } else if (isGpsOnline) {
      setEscalations(prev => prev.filter(e => e.type !== 'HARDWARE_OFFLINE'));
    }
  }, [isGpsOnline, hasGps]);

  // --- Command Handlers ---
  const addAudit = (event: string, desc: string, severity: AuditLog['severity'] = 'normal', actor: string = 'Admin') => {
    const newLog: AuditLog = { 
      id: `log-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`, 
      time: new Date().toLocaleTimeString('en-GB', { hour12: false }), 
      event, 
      actor, 
      desc, 
      severity,
      nodeId: `NODE-${Math.floor(Math.random() * 900) + 100}`,
      sessionId: `SESS-${Math.random().toString(36).substr(2, 6).toUpperCase()}`
    };
    setAuditLogs(prev => [newLog, ...prev]);
  };


  const handleAssignDriver = async (driver: any, bypass: boolean = false, bypassReason: string = "") => {
    if (!driver?.id) {
      console.error("Cannot assign driver without a valid identifier");
      return;
    }
    try {
      setIsSyncing(true);
      const idempotencyKey = `assign_${id}_${driver.id}_${Date.now()}`;
      
      const payload = { 
        vehicle_id: id,
        vehicle_version: vehicle.version || 0,
        driver_id: driver.id,
        driver_version: driver.version || 0,
        remarks: bypass ? `Bypass: ${bypassReason}` : "Standard operational assignment",
        bypass_policy: bypass,
        bypass_reason: bypassReason
      };

      await apiClient.post(`/admin/assignments/assign`, payload, {
        headers: { 'Idempotency-Key': idempotencyKey }
      });
      
      addAudit('Assignment.Created', `Asset control transferred to ${driver.name}. Version validation complete.`, 'normal', 'CentralOps');
      setCommand('none');
      setPolicyViolation(null);
      setConflictData(null);
      
      await fetchVehicle();
      await fetchDrivers();
      await fetchAuditLogs();
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;

      if (status === 409) {
        setConflictData({ driver, serverState: data });
        addAudit('Assignment.ConflictDetected', "OCC version mismatch detected.", 'critical', 'System');
      } else if (status === 400 && data?.error?.toLowerCase().includes("policy")) {
        setPolicyViolation({ driver, message: data.error });
      } else {
        alert(data?.error || "Assignment failed");
      }
    } finally {
      setIsSyncing(false);
    }
  };

  // --- Modal Components ---
  const ConflictResolutionModal = () => (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-md p-4">
      <div className="w-full max-w-lg bg-zinc-900 border border-rose-500/30 rounded-3xl overflow-hidden shadow-2xl shadow-rose-500/10">
        <div className="p-6 border-b border-white/5 bg-rose-500/10">
          <div className="flex items-center gap-3 text-rose-400 mb-2">
            <ShieldAlert className="w-6 h-6" />
            <h3 className="text-xl font-bold">Divergence Detected</h3>
          </div>
          <p className="text-zinc-400 text-sm">The asset state has changed since you opened this view. To prevent corruption, please reconcile.</p>
        </div>
        <div className="p-6 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 rounded-2xl bg-white/5 border border-white/10">
              <span className="text-[10px] uppercase tracking-widest text-zinc-500 block mb-1">Your View</span>
              <div className="font-mono text-xs text-white">v{vehicle.version}</div>
            </div>
            <div className="p-4 rounded-2xl bg-rose-500/5 border border-rose-500/20">
              <span className="text-[10px] uppercase tracking-widest text-rose-500 block mb-1">Live Server</span>
              <div className="font-mono text-xs text-rose-400">v{conflictData.serverState.CurrentVersions?.vehicle}</div>
            </div>
          </div>
          <div className="p-4 rounded-2xl bg-white/5 border border-white/10 text-xs text-zinc-400 leading-relaxed">
            Re-syncing will refresh the versions and allow you to re-attempt assignment safely.
          </div>
        </div>
        <div className="p-6 bg-white/[0.02] border-t border-white/5 flex gap-3">
          <button onClick={() => setConflictData(null)} className="flex-1 py-3 rounded-xl bg-white/5 hover:bg-white/10 text-white font-medium transition-all">Cancel</button>
          <button 
            onClick={() => {
              fetchVehicle();
              fetchDrivers();
              setConflictData(null);
            }} 
            className="flex-1 py-3 rounded-xl bg-rose-500 hover:bg-rose-400 text-white font-bold transition-all shadow-lg shadow-rose-500/20"
          >
            Sync & Refresh
          </button>
        </div>
      </div>
    </div>
  );

  const PolicyBypassModal = () => {
    const [reason, setReason] = useState("");
    return (
      <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-md p-4">
        <div className="w-full max-w-lg bg-zinc-900 border border-amber-500/30 rounded-3xl overflow-hidden shadow-2xl shadow-amber-500/10">
          <div className="p-6 border-b border-white/5 bg-amber-500/10">
            <div className="flex items-center gap-3 text-amber-400 mb-2">
              <AlertTriangle className="w-6 h-6" />
              <h3 className="text-xl font-bold">Compliance Override</h3>
            </div>
            <p className="text-zinc-400 text-sm">Policy engine blocked assignment: {policyViolation.message}</p>
          </div>
          <div className="p-6 space-y-4">
            <label className="block">
              <span className="text-xs text-zinc-500 uppercase tracking-widest mb-2 block">Bypass Rationale (Required)</span>
              <textarea 
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="Explain why this safety block is being bypassed..."
                className="w-full h-32 bg-white/5 border border-white/10 rounded-2xl p-4 text-white placeholder:text-zinc-600 focus:outline-none focus:border-amber-500/50 transition-all resize-none"
              />
            </label>
          </div>
          <div className="p-6 bg-white/[0.02] border-t border-white/5 flex gap-3">
            <button onClick={() => setPolicyViolation(null)} className="flex-1 py-3 rounded-xl bg-white/5 hover:bg-white/10 text-white font-medium transition-all">Abort</button>
            <button 
              disabled={!reason.trim()}
              onClick={() => handleAssignDriver(policyViolation.driver, true, reason)} 
              className="flex-1 py-3 rounded-xl bg-amber-500 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-amber-400 text-black font-bold transition-all shadow-lg shadow-amber-500/20"
            >
              Force Assign
            </button>
          </div>
        </div>
      </div>
    );
  };

  const handleUnassignDriver = async () => {
    try {
      setIsSyncing(true);
      await apiClient.post(`/admin/assignments/unassign`, { 
        vehicle_id: id,
        remarks: "Manual operational unassignment"
      });
      
      addAudit('Assignment.Released', `Asset control revoked from ${currentDriver?.name || 'Pilot'}.`, 'warn', 'CentralOps');
      
      await fetchVehicle();
      await fetchDrivers();
      await fetchAuditLogs();
    } catch (err: any) {
      console.error("Unassignment failed:", err);
      const msg = err.response?.data?.error || err.message || "Unassignment failed";
      addAudit('UNASSIGN_FAILURE', msg, 'critical', 'System');
    } finally {
      setIsSyncing(false);
    }
  };

  const handleDocumentUpload = (docId?: string) => {
    setCommand('upload');
    setCommandData({ docId, progress: 0, state: 'idle' });
  };

  const simulateOCR = (docName: string) => {
    setCommandData({ ...commandData, state: 'processing', progress: 10 });
    let p = 10;
    const interval = setInterval(() => {
      p += Math.random() * 20;
      if (p >= 100) {
        clearInterval(interval);
        setCommandData({ ...commandData, state: 'success', progress: 100 });
        setRegistryDocs(prev => prev.map(d => d.type === docName ? { ...d, status: 'VALID', color: 'emerald', pulse: 'Verified', impact: 'Dispatch Allowed' } : d));
        addAudit('OCR_COMPLETE', `Extraction successful for ${docName}. Metadata pulse: 420d remaining.`, 'normal', 'ComplianceAI');
        setTimeout(() => setCommand('none'), 1500);
      } else {
        setCommandData((prev: any) => ({ ...prev, progress: Math.min(p, 95) }));
      }
    }, 400);
  };

  if (loading || !vehicle) return null;

  const isRestricted = stats.isRestricted || vehicle.status === 'blocked';
  const markers = [{ 
    id: vehicle.id, 
    registration_number: vehicle.registration_number, 
    latitude: telemetry?.latitude || 28.6139, 
    longitude: telemetry?.longitude || 77.2090, 
    status: vehicle.status,
    lat: telemetry?.latitude || 28.6139,
    lng: telemetry?.longitude || 77.2090,
    label: vehicle.registration_number
  }];

  return (
    <div className="flex-1 flex flex-col bg-background-dark overflow-hidden font-sans relative">
      
      {/* --- Full Command Center Modals --- */}
      {command !== 'none' && (
        <div className="absolute inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-6">
           <div className="w-full max-w-lg bg-[#0d1117] border border-white/10 rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
              <div className="px-6 py-4 border-b border-white/5 flex items-center justify-between bg-white/[0.02]">
                 <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white flex items-center gap-2">
                    {command === 'assign' && <UserPlus className="w-4 h-4 text-primary" />}
                    {command === 'upload' && <Upload className="w-4 h-4 text-emerald-500" />}
                    {command === 'block' && <Ban className="w-4 h-4 text-rose-500" />}
                    {command === 'maintenance' && <Toolbox className="w-4 h-4 text-amber-500" />}
                    {command === 'provision' && <Zap className="w-4 h-4 text-primary" />}
                    {command === 'builder' && <Plus className="w-4 h-4 text-primary" />}
                    {command === 'assign' ? 'Assign Driver' : `${command.toUpperCase()} COMMAND CENTER`}
                 </h3>
                 <button onClick={() => setCommand('none')} className="p-2 hover:bg-white/5 rounded-full transition-all text-gray-500 hover:text-white">
                    <X className="w-4 h-4" />
                 </button>
              </div>

               <div className="p-8 space-y-6">
                  {command === 'assign' && (
                    <div className="space-y-6">
                       {/* Unassign Option if pilot currently exists */}
                       {(vehicle?.current_driver_id && vehicle?.current_driver_id !== ZERO_UUID) && (
                         <div className="space-y-3 pb-6 border-b border-white/5">
                            <label className="text-[10px] font-black text-rose-500 uppercase tracking-widest block">Release Control</label>
                            <button 
                              onClick={handleUnassignDriver}
                              className="w-full flex items-center justify-between p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 hover:bg-rose-500/20 transition-all group"
                            >
                               <div className="flex items-center gap-4">
                                  <div className="w-10 h-10 rounded-full bg-rose-500/20 flex items-center justify-center text-rose-500">
                                     <UserMinus className="w-5 h-5" />
                                  </div>
                                  <div className="text-left">
                                     <div className="text-xs font-black text-white uppercase tracking-tight">Unlink Current Pilot</div>
                                     <div className="text-[9px] text-rose-500/60 uppercase font-bold tracking-widest">Release asset for reassignment</div>
                                  </div>
                               </div>
                               <ChevronRight className="w-4 h-4 text-rose-900 group-hover:text-rose-500 transition-all" />
                            </button>
                         </div>
                       )}

                       <div className="space-y-4">
                          <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Recommended Drivers</label>
                          <div className="grid grid-cols-1 gap-2">
                             {allDrivers.slice(0, 2).map((rec) => {
                                const isReserved = (rec.current_vehicle?.id && rec.current_vehicle.id !== ZERO_UUID) || 
                                                 (rec.current_vehicle_id && rec.current_vehicle_id !== ZERO_UUID);
                                const assetReg = rec.current_vehicle?.registration_number || 'Reserved';
                                
                                return (
                                <div key={rec.id || `rec-${rec.name}`} onClick={() => {
                                   if (isReserved && !window.confirm(`This pilot is RESERVED for asset ${assetReg}. Are you sure you want to override?`)) return;
                                   handleAssignDriver(rec);
                                }} className={`flex items-center justify-between p-4 rounded-xl border transition-all group cursor-pointer text-left ${
                                   isReserved ? 'bg-amber-500/5 border-amber-500/20 hover:bg-amber-500/10' : 'bg-primary/5 border-primary/20 hover:bg-primary/10'
                                }`}>
                                   <div className="flex items-center gap-4">
                                      <div className={`w-10 h-10 rounded-full flex items-center justify-center font-black uppercase text-xs ${
                                         isReserved ? 'bg-amber-500/20 text-amber-500' : 'bg-primary/10 text-primary'
                                      }`}>
                                         {rec.name?.charAt(0) || 'D'}
                                      </div>
                                      <div>
                                         <div className="text-xs font-black text-white uppercase">{rec.name || 'Unknown Driver'}</div>
                                         <div className={`text-[9px] uppercase font-bold tracking-widest ${isReserved ? 'text-amber-500' : 'text-primary'}`}>
                                            {rec.id === vehicle?.current_driver_id ? 'Current Pilot' : isReserved ? `Reserved • ${assetReg}` : 'Recommended'}
                                         </div>
                                      </div>
                                   </div>
                                   <span className={`text-[8px] font-black uppercase border px-1.5 py-0.5 rounded ${
                                      isReserved ? 'text-amber-500 border-amber-500/20' : 'text-gray-700 border-white/5'
                                   }`}>{rec.status || 'Available'}</span>
                                </div>
                                )})}
                             {allDrivers.length === 0 && (
                               <div className="text-[10px] text-gray-600 uppercase font-bold text-center py-4">No available drivers found</div>
                             )}
                          </div>
                       </div>

                       <div className="relative">
                          <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-3">Driver Registry Search</label>
                          <div className="relative">
                             <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                             <input 
                               type="text" 
                               placeholder="Search Driver Registry..." 
                               className="w-full bg-black/40 border border-white/5 rounded-xl pl-10 pr-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all"
                               value={searchQuery}
                               onChange={(e) => setSearchQuery(e.target.value)}
                             />
                          </div>
                       </div>
                       
                       <div className="max-h-[200px] overflow-y-auto custom-scrollbar space-y-2">
                          {filteredDrivers.filter(d => !allDrivers.slice(0, 2).map(rd => rd.id).includes(d.id)).map(d => {
                             const isReserved = (d.current_vehicle?.id && d.current_vehicle.id !== ZERO_UUID) || 
                                              (d.current_vehicle_id && d.current_vehicle_id !== ZERO_UUID);
                             const assetReg = d.current_vehicle?.registration_number || 'Reserved';

                             return (
                             <div key={d.id} onClick={() => {
                                if (isReserved && !window.confirm(`This pilot is RESERVED for asset ${assetReg}. Are you sure you want to override?`)) return;
                                handleAssignDriver(d);
                             }} className={`flex items-center justify-between p-3 rounded-xl border cursor-pointer transition-all group ${
                                isReserved ? 'bg-amber-500/5 border-amber-500/10 hover:border-amber-500/30' : 'bg-white/[0.02] border-white/5 hover:border-primary/40 hover:bg-white/[0.05]'
                             }`}>
                               <div className="flex items-center gap-4">
                                  <div className={`w-8 h-8 rounded-full flex items-center justify-center font-black uppercase text-[10px] ${
                                      isReserved ? 'bg-amber-500/20 text-amber-500' : 'bg-white/5 text-gray-500'
                                   }`}>
                                     {d.name?.charAt(0)}
                                  </div>
                                  <div>
                                     <div className="text-xs font-black text-white uppercase">{d.name}</div>
                                      <div className={`text-[9px] uppercase font-bold tracking-widest ${isReserved ? 'text-amber-500' : 'text-gray-700'}`}>
                                         {isReserved ? `Reserved • ${assetReg}` : (d.phone || 'NO_VOICE_LINK')}
                                      </div>
                                   </div>
                                </div>
                                <ChevronRight className={`w-4 h-4 transition-all ${isReserved ? 'text-amber-900 group-hover:text-amber-500' : 'text-gray-800 group-hover:text-primary'}`} />
                             </div>
                           )})}
                        </div>
                     </div>
                   )}

                   {command === 'provision' && (
                    <div className="space-y-8">
                       <div className="flex items-center justify-between mb-8">
                          {[1, 2, 3].map(s => (
                             <div key={s} className="flex items-center gap-2">
                                <div className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-black ${commandData?.step >= s ? 'bg-primary text-white' : 'bg-white/5 text-gray-700'}`}>{s}</div>
                                {s < 3 && <div className={`h-0.5 w-12 rounded ${commandData?.step >= s ? 'bg-primary/40' : 'bg-white/5'}`} />}
                             </div>
                          ))}
                       </div>

                       {(!commandData?.step || commandData.step === 1) && (
                         <div className="space-y-6 animate-in slide-in-from-bottom-4 duration-300">
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2">Device Serial Mapping</label>
                               <input type="text" placeholder="Scan or Enter Serial Number..." className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all" />
                            </div>
                            <button onClick={() => setCommandData({ ...commandData, step: 2 })} className="w-full bg-primary text-white py-3.5 rounded-xl text-xs font-black uppercase tracking-widest">VERIFY MAPPING</button>
                         </div>
                       )}

                       {commandData?.step === 2 && (
                         <div className="space-y-6 animate-in slide-in-from-bottom-4 duration-300">
                            <div className="p-8 bg-black/40 rounded-2xl border border-white/5 flex flex-col items-center justify-center gap-4">
                               <div className="w-12 h-12 rounded-full border-2 border-primary border-t-transparent animate-spin" />
                               <div className="text-[10px] font-black text-white uppercase tracking-widest">Establishing Operational Uplink...</div>
                            </div>
                            <button onClick={() => {
                               setCommandData({ ...commandData, step: 3 });
                               addAudit('PROVISION_SYNC', 'Device handshake successful. Syncing telemetry state.', 'normal');
                            }} className="w-full bg-primary text-white py-3.5 rounded-xl text-xs font-black uppercase tracking-widest">NEXT: HANDSHAKE</button>
                         </div>
                       )}

                       {commandData?.step === 3 && (
                         <div className="space-y-6 animate-in slide-in-from-bottom-4 duration-300">
                            <div className="p-6 bg-emerald-500/5 border border-emerald-500/20 rounded-2xl text-center">
                               <div className="text-emerald-500 font-black uppercase text-xs mb-2">Setup Integrity Verified</div>
                               <div className="text-[9px] text-emerald-500/60 uppercase font-bold tracking-widest">AIS-140 COMPLIANT SIGNAL DETECTED</div>
                            </div>
                            <button onClick={() => {
                               addAudit('PROVISION_COMPLETE', 'Hardware setup finalized. Asset is now traceable.', 'normal');
                               setCommand('none');
                            }} className="w-full bg-emerald-600 text-white py-3.5 rounded-xl text-xs font-black uppercase tracking-widest">FINALIZE SETUP</button>
                         </div>
                       )}
                    </div>
                  )}

                 {command === 'builder' && (
                   <div className="space-y-6 max-h-[70vh] overflow-y-auto custom-scrollbar pr-2">
                      <div className="space-y-6">
                         {/* Basic Definition */}
                         <div className="grid grid-cols-2 gap-4">
                            <div className="col-span-2">
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Requirement Name (Plain Language)</label>
                               <input id="req-type" type="text" placeholder="e.g. National Permit" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all" />
                            </div>
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Jurisdiction/State</label>
                               <input id="req-jurisdiction" type="text" placeholder="e.g. Maharashtra" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all" />
                            </div>
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Vehicle Class</label>
                               <select id="req-class" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer">
                                  <option value="All">All Classes</option>
                                  <option value="HCV">HCV (Heavy)</option>
                                  <option value="LCV">LCV (Light)</option>
                                  <option value="Trailer">Trailer</option>
                                  <option value="Tanker">Tanker</option>
                               </select>
                            </div>
                         </div>

                         {/* Durations & Thresholds */}
                         <div className="grid grid-cols-2 gap-4">
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Expiry Duration (Days)</label>
                               <input id="req-duration" type="number" placeholder="365" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all" />
                            </div>
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Renewal Reminder (Days)</label>
                               <input id="req-reminder" type="number" placeholder="15" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all" />
                            </div>
                         </div>
                         
                         {/* Operational Logic */}
                         <div className="p-4 bg-white/[0.02] border border-white/5 rounded-xl flex items-center justify-between">
                            <div>
                               <div className="text-xs font-black text-white uppercase mb-1">Mandatory Enforcement</div>
                               <div className="text-[9px] text-gray-600 uppercase font-bold tracking-widest ">Blocks asset if non-compliant</div>
                            </div>
                            <input id="req-mandatory" type="checkbox" className="w-5 h-5 rounded border-white/10 bg-black/40 text-primary focus:ring-primary" defaultChecked />
                         </div>

                         <div className="grid grid-cols-2 gap-4">
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Operational Consequence</label>
                               <select id="req-impact" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer">
                                  <option value="Dispatch Blocked">Block Dispatch (Critical)</option>
                                  <option value="Manual Review">Manual Review Required</option>
                                  <option value="Route Restricted">Route Restricted</option>
                               </select>
                            </div>
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Requirement Class</label>
                               <select id="req-class-type" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer" disabled={userRole !== 'FLYFLEET_CENTRAL_OPS'}>
                                  <option value="Tenant">Tenant Policy (Local)</option>
                                  <option value="Global" selected={userRole === 'FLYFLEET_CENTRAL_OPS'}>Global Standard (Platform)</option>
                               </select>
                            </div>
                         </div>
                         
                         <div className="grid grid-cols-1">
                            <div>
                               <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2 ">Verification Mode</label>
                               <select id="req-verify" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer">
                                  <option value="AI_OCR">AI Extraction (OCR)</option>
                                  <option value="Manual">Manual HQ Approval</option>
                                  <option value="ThirdParty">Third-Party Verification</option>
                               </select>
                            </div>
                         </div>
                      </div>

                      <button 
                        onClick={() => {
                          const type = (document.getElementById('req-type') as HTMLInputElement).value;
                          const jurisdiction = (document.getElementById('req-jurisdiction') as HTMLInputElement).value;
                          const vClass = (document.getElementById('req-class') as HTMLSelectElement).value;
                          const mandatory = (document.getElementById('req-mandatory') as HTMLInputElement).checked;
                          const impact = (document.getElementById('req-impact') as HTMLSelectElement).value;
                          if (type) {
                            addRegistryRequirement({ type, jurisdiction, vClass, mandatory, impact });
                            setCommand('none');
                          }
                        }}
                        className="w-full bg-primary hover:bg-primary-dark text-white py-4 rounded-xl text-xs font-black uppercase tracking-[0.2em] shadow-xl shadow-primary/10 transition-all"
                      >
                         SAVE & INGEST REQUIREMENT
                      </button>
                   </div>
                 )}

                 {command === 'upload' && (
                   <div className="space-y-8 py-4 text-center">
                      {commandData?.state === 'idle' ? (
                        <>
                          <div className="w-20 h-20 rounded-full bg-emerald-500/10 flex items-center justify-center mx-auto mb-6">
                             <Upload className="w-10 h-10 text-emerald-500" />
                          </div>
                          <div>
                             <h4 className="text-white font-black uppercase tracking-widest text-sm mb-2">Registry Ingestion Required</h4>
                             <p className="text-xs text-gray-600 leading-relaxed max-w-[280px] mx-auto">Upload the required regulatory entity. OCR Node will automatically extract and verify integrity.</p>
                          </div>
                          <button 
                            onClick={() => simulateOCR(registryDocs.find(d => d.id === commandData?.docId)?.type || 'Insurance Policy')}
                            className="w-full bg-emerald-600 hover:bg-emerald-500 text-white py-3.5 rounded-xl text-xs font-black uppercase tracking-[0.2em] shadow-xl shadow-emerald-500/10 transition-all"
                          >
                             Select & Ingest Node
                          </button>
                        </>
                      ) : (
                        <div className="space-y-6">
                           <div className="relative w-24 h-24 mx-auto">
                              <svg className="w-full h-full -rotate-90">
                                 <circle cx="48" cy="48" r="40" fill="transparent" stroke="rgba(255,255,255,0.05)" strokeWidth="8" />
                                 <circle cx="48" cy="48" r="40" fill="transparent" stroke="#10b981" strokeWidth="8" strokeDasharray="251.2" strokeDashoffset={251.2 - (251.2 * commandData.progress) / 100} className="transition-all duration-300" />
                              </svg>
                              <div className="absolute inset-0 flex items-center justify-center text-[10px] font-black text-emerald-500">
                                 {Math.round(commandData.progress)}%
                              </div>
                           </div>
                           <div>
                              <h4 className="text-white font-black uppercase tracking-widest text-sm mb-1">
                                 {commandData.state === 'processing' ? 'OCR EXTRACTION ACTIVE' : 'INTEGRITY VERIFIED'}
                              </h4>
                              <p className="text-[10px] text-gray-600 uppercase font-black tracking-widest">
                                 {commandData.state === 'processing' ? 'PULSE: READING METADATA...' : 'PULSE: REGISTRY UPDATED'}
                              </p>
                           </div>
                        </div>
                      )}
                   </div>
                 )}

                 {command === 'block' && (
                   <div className="space-y-6">
                      <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-xl flex gap-4">
                         <AlertTriangle className="w-6 h-6 text-rose-500 shrink-0" />
                         <div className="text-[10px] text-rose-500 font-bold leading-relaxed">
                            <span className="block font-black uppercase mb-1 underline">SECURITY PROTOCOL INITIATED</span>
                            Blocking this asset will restrict all dispatch operations and notify HQ immediately.
                         </div>
                      </div>
                      <div className="space-y-2">
                         <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest">Reason for Blockade</label>
                         <textarea className="w-full bg-black/40 border border-white/5 rounded-xl p-4 text-sm text-white focus:outline-none focus:border-rose-500 transition-all min-h-[100px]" placeholder="State security or compliance failure..." />
                      </div>
                      <button onClick={() => { addAudit('ASSET_BLOCKED', 'Manual security block initiated by Admin.', 'critical'); setCommand('none'); }} className="w-full bg-rose-600 hover:bg-rose-500 text-white py-3.5 rounded-xl text-xs font-black uppercase tracking-[0.2em] transition-all">
                         CONFIRM ASSET BLOCKADE
                      </button>
                   </div>
                 )}

                 {command === 'maintenance' && (
                    <div className="space-y-6">
                       <div className="grid grid-cols-2 gap-4">
                          <div className="col-span-2">
                             <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2">Service Type</label>
                             <select className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer">
                                <option>Routine Maintenance</option>
                                <option>Hardware Replacement</option>
                                <option>Accident Repair</option>
                             </select>
                          </div>
                          <div>
                             <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2">Estimated Downtime</label>
                             <input type="text" placeholder="e.g. 4 Hours" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white" />
                          </div>
                          <div>
                             <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2">Technician ID</label>
                             <input type="text" placeholder="Node #24" className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white" />
                          </div>
                       </div>
                       <button onClick={() => { addAudit('MAINTENANCE_LOGGED', 'Routine maintenance cycle initiated.', 'normal'); setCommand('none'); }} className="w-full bg-amber-600 hover:bg-amber-500 text-white py-3.5 rounded-xl text-xs font-black uppercase tracking-widest transition-all">
                          AUTHORIZE MAINTENANCE
                       </button>
                    </div>
                 )}

                 {command === 'governance' && (
                    <div className="space-y-8">
                       <div className="p-6 bg-rose-500/5 border border-rose-500/20 rounded-2xl">
                          <div className="flex items-center gap-3 mb-4 text-rose-500">
                             <Lock className="w-5 h-5" />
                             <span className="text-xs font-black uppercase tracking-[0.2em]">Platform Override Lock</span>
                          </div>
                          <p className="text-[10px] text-gray-500 font-bold leading-relaxed">
                             You are requesting a manual bypass of a mission-critical blockade: <span className="text-white">"{commandData?.desc}"</span>.
                             This action is strictly restricted to <span className="text-rose-500 underline">FlyFleet Central Operations</span> and requires a multi-factor handshake.
                          </p>
                       </div>

                       <div className="space-y-4">
                          <div>
                             <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2">Governance Key</label>
                             <input type="password" placeholder="Enter Platform Authorization Key..." className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-rose-500 transition-all" />
                          </div>
                          <div>
                             <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block mb-2">Override Justification (Audit Required)</label>
                             <textarea placeholder="State reason for manual bypass..." className="w-full bg-black/40 border border-white/5 rounded-xl p-4 text-sm text-white focus:outline-none focus:border-rose-500 transition-all min-h-[80px]" />
                          </div>
                       </div>

                       <button 
                         onClick={() => {
                           addAudit('PLATFORM_OVERRIDE', `Manual blockade bypass authorized for ${commandData?.type}. Registry lock suspended.`, 'critical', 'FlyFleet_HQ');
                           setEscalations(prev => prev.filter(e => e.id !== commandData?.id));
                           setCommand('none');
                         }}
                         className="w-full bg-rose-600 hover:bg-rose-500 text-white py-4 rounded-xl text-xs font-black uppercase tracking-[0.2em] shadow-xl shadow-rose-500/20 transition-all"
                       >
                          BYPASS SECURITY LOCK
                       </button>
                    </div>
                 )}
              </div>
           </div>
        </div>
      )}

      {/* --- Escalation Overlay Panel --- */}
      {showEscalations && (
        <div className="absolute top-16 right-6 w-96 bg-[#0d1117] border border-white/10 rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] z-[60] overflow-hidden animate-in slide-in-from-top-4 duration-300">
           <div className="p-4 border-b border-white/5 bg-white/[0.02] flex items-center justify-between">
              <div className="flex items-center gap-2">
                 <ShieldAlert className="w-4 h-4 text-rose-500" />
                 <span className="text-[10px] font-black text-white uppercase tracking-widest">Active Escalations</span>
              </div>
              <button onClick={() => setShowEscalations(false)} className="text-gray-500 hover:text-white transition-all"><X className="w-4 h-4" /></button>
           </div>
           <div className="max-h-[400px] overflow-y-auto custom-scrollbar p-2">
              {escalations.length === 0 ? (
                <div className="p-12 text-center">
                   <ShieldCheck className="w-8 h-8 text-emerald-500/20 mx-auto mb-3" />
                   <div className="text-[10px] font-black text-gray-700 uppercase tracking-widest">No Active Operational Blockades</div>
                </div>
              ) : (
                <div className="space-y-1">
                   {escalations.map((esc) => (
                     <div key={esc.id} className="p-4 rounded-xl bg-white/[0.02] border border-white/5 hover:bg-white/[0.04] transition-all">
                        <div className="flex items-start justify-between mb-2">
                           <div className={`px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-widest ${esc.severity === 'critical' ? 'bg-rose-500/20 text-rose-500' : 'bg-amber-500/20 text-amber-500'}`}>
                              {esc.type}
                           </div>
                           <span className="text-[8px] font-black text-gray-800 uppercase">{esc.timestamp}</span>
                        </div>
                        <p className="text-[10px] text-gray-400 font-bold leading-relaxed mb-4">{esc.desc}</p>
                        <div className="flex gap-2">
                           <button 
                             onClick={() => {
                               if (esc.workflow === 'compliance') setActiveTab('documents');
                               if (esc.workflow === 'hardware') setCommand('provision');
                               setShowEscalations(false);
                             }}
                             className="flex-1 bg-white/5 hover:bg-white/10 text-white py-2 rounded-lg text-[9px] font-black uppercase tracking-widest transition-all"
                           >
                             SEARCH
                           </button>
                           <button 
                             onClick={() => { setCommand('governance'); setCommandData(esc); setShowEscalations(false); }}
                             disabled={userRole !== 'FLYFLEET_CENTRAL_OPS'}
                             className={`px-3 rounded-lg transition-all ${
                               userRole === 'FLYFLEET_CENTRAL_OPS' 
                                 ? 'bg-rose-500/10 hover:bg-rose-500/20 text-rose-500' 
                                 : 'bg-gray-800 text-gray-900 cursor-not-allowed opacity-50'
                             }`}
                             title={userRole === 'FLYFLEET_CENTRAL_OPS' ? "Request Override" : "Restricted: Central Ops Only"}
                           >
                              <Lock className="w-3.5 h-3.5" />
                           </button>
                        </div>
                     </div>
                   ))}
                </div>
              )}
           </div>
           <div className="p-4 bg-white/[0.01] border-t border-white/5 text-center">
              <div className="text-[8px] font-black text-gray-800 uppercase tracking-widest">
                 Awaiting HQ Governance Clearance
              </div>
           </div>
        </div>
      )}

      {/* --- Header Section --- */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-30">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate("/admin/vehicles")} className="p-2 hover:bg-white/5 rounded-full transition-colors text-text-muted">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <div className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-[0.2em] font-black mb-0.5">
              <span>Fleet Operations</span>
              <ChevronRight className="w-3 h-3 opacity-30" />
              <span className="text-white">{vehicle.registration_number}</span>
            </div>
            <h2 className="text-white text-lg font-bold flex items-center gap-2 uppercase tracking-tight ">
              {vehicle.make} {vehicle.model}
              <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-wider border transition-all ${
                isRestricted ? 'bg-rose-500/10 text-rose-500 border-rose-500/20 shadow-[0_0_10px_rgba(244,63,94,0.1)]' : 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20'
              }`}>
                {isRestricted 
                  ? (stats.blocked > 1 ? `Blocked: ${stats.blocked} Critical Failures` : `Blocked: ${stats.blockReason}`) 
                  : 'Operational'}
              </span>
            </h2>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button 
            onClick={() => setShowEscalations(!showEscalations)}
            className={`relative px-4 py-2 border rounded-lg flex items-center gap-2 transition-all ${
              escalations.length > 0 
                ? 'bg-rose-500/10 border-rose-500/20 text-rose-500 animate-pulse' 
                : 'bg-surface-dark border-border-color text-text-muted'
            }`}
          >
            <ShieldAlert className="w-3.5 h-3.5" />
            <span className="text-[10px] font-black uppercase tracking-widest">Escalations</span>
            {escalations.length > 0 && (
              <span className="absolute -top-1.5 -right-1.5 w-4 h-4 bg-rose-600 text-white text-[8px] font-black flex items-center justify-center rounded-full shadow-lg shadow-rose-500/40">
                {escalations.length}
              </span>
            )}
          </button>
          <button onClick={() => { setIsSyncing(true); fetchVehicle(); }} className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded-lg flex items-center gap-2 transition-all">
            <RefreshCw className={`w-3.5 h-3.5 ${isSyncing ? 'animate-spin' : ''}`} />
            {isSyncing ? 'Recalculating...' : 'Sync Data'}
          </button>
          <button onClick={() => navigate(`/admin/vehicles/${vehicle.id}/edit`)} className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 shadow-lg shadow-primary/20 transition-all">
            <Settings className="w-3.5 h-3.5" />
            Edit Profile
          </button>
        </div>
      </header>

      {/* --- Main Dashboard Area --- */}
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        
        {/* Signal Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4 cursor-pointer hover:bg-surface-darker transition-all" onClick={() => setActiveTab('documents')}>
            <div className={`p-3 rounded-xl ${isRestricted ? 'bg-rose-500/10 text-rose-500' : 'bg-emerald-500/10 text-emerald-500'}`}>
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div>
              <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Vehicle Status</div>
              <div className={`text-xl font-bold ${isRestricted ? 'text-rose-500' : 'text-emerald-500'}`}>{isRestricted ? 'RESTRICTED' : 'OPERATIONAL'}</div>
            </div>
          </div>

          <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
            <div className={`p-3 rounded-xl ${isGpsOnline ? 'bg-emerald-500/10 text-emerald-500' : 'bg-rose-500/10 text-rose-500'}`}>
              <Zap className="w-6 h-6" />
            </div>
            <div>
              <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Vehicle Tracking</div>
              <div className={`text-xl font-bold ${isGpsOnline ? 'text-emerald-500' : 'text-rose-500'}`}>{hasGps ? (isGpsOnline ? 'LIVE' : 'OFFLINE') : 'MISSING'}</div>
            </div>
          </div>

          <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4 cursor-pointer hover:bg-surface-darker transition-all group" onClick={() => setCommand('assign')}>
            <div className={`p-3 rounded-xl transition-all ${currentDriver ? 'bg-primary/10 text-primary group-hover:bg-primary/20' : 'bg-amber-500/10 text-amber-500 group-hover:bg-amber-500/20'}`}>
              <User className="w-6 h-6" />
            </div>
            <div className="flex-1">
              <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Assigned Driver</div>
              <div className="flex items-center justify-between">
                <div className="text-xl font-bold text-white uppercase ">{currentDriver ? currentDriver.name?.split(' ')[0] : 'VACANT'}</div>
                {currentDriver && (
                  <button 
                    onClick={(e) => { e.stopPropagation(); handleUnassignDriver(); }}
                    className="p-1.5 hover:bg-rose-500/10 text-gray-700 hover:text-rose-500 rounded transition-all"
                    title="Unassign Driver"
                  >
                    <Ban className="w-3.5 h-3.5" />
                  </button>
                )}
              </div>
            </div>
          </div>

          <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
            <div className="p-3 rounded-xl bg-white/5 text-white">
              <Signal className="w-6 h-6" />
            </div>
            <div>
              <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Utilization</div>
              <div className="text-xl font-bold text-white uppercase ">{vehicle.odometer || 0} <span className="text-sm font-normal text-text-muted uppercase tracking-widest">KM</span></div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-6">
          {/* Information & Documents Panel */}
          <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden shadow-xl flex flex-col min-h-[600px]">
              
              <div className="flex items-center border-b border-border-color px-2 bg-surface-darker/30">
                <button onClick={() => setActiveTab("overview")} className={`px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] transition-all ${activeTab === "overview" ? "text-primary border-b-2 border-primary" : "text-gray-600 hover:text-white"}`}>
                  Vehicle Information
                </button>
                <button onClick={() => setActiveTab("documents")} className={`px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] transition-all ${activeTab === "documents" ? "text-primary border-b-2 border-primary" : "text-gray-600 hover:text-white"}`}>
                  Documents & Compliance
                </button>
              </div>

              <div className="p-6">
                {activeTab === "overview" && (
                  <div className="space-y-12">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-12">
                      <div>
                        <h3 className="text-[10px] font-black text-gray-500 uppercase tracking-[0.2em] mb-6 flex items-center gap-2 ">
                          <Truck className="w-4 h-4 text-primary opacity-50" />
                          Vehicle Identification
                        </h3>
                        <div className="space-y-4">
                          {[
                            { label: 'Vehicle ID', val: vehicle.registration_number },
                            { label: 'VIN Number', val: vehicle.vin_number || '—' },
                            { label: 'Manufacturer', val: vehicle.make },
                            { label: 'Performance Details', val: `${vehicle.model} (${vehicle.year})` },
                          ].map((it, i) => (
                            <div key={i} className="flex justify-between items-center py-2.5 border-b border-white/[0.03]">
                              <span className="text-xs text-text-muted font-bold uppercase tracking-tight">{it.label}</span>
                              <span className="text-sm font-black text-white uppercase ">{it.val}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                      <div>
                        <h3 className="text-[10px] font-black text-gray-500 uppercase tracking-[0.2em] mb-6 flex items-center gap-2 ">
                          <Activity className="w-4 h-4 text-emerald-500 opacity-50" />
                          Technical Specifications
                        </h3>
                        <div className="space-y-4">
                          {[
                            { label: 'Fuel Type', val: vehicle.fuel_type || 'Diesel' },
                            { label: 'Fuel Tank Capacity', val: `${vehicle.fuel_tank_capacity || '—'}L` },
                            { label: 'Speed Limit', val: vehicle.speed_limit ? `${vehicle.speed_limit} km/h` : 'OFF' },
                            { label: 'Safety Device (AIS-140)', val: vehicle.ais140_compliant ? 'Verified' : 'Pending' },
                          ].map((it, i) => (
                            <div key={i} className="flex justify-between items-center py-2.5 border-b border-white/[0.03]">
                              <span className="text-xs text-text-muted font-bold uppercase tracking-tight">{it.label}</span>
                              <span className="text-sm font-black text-white uppercase ">{it.val}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                    
                    {/* Operational Commands */}
                    <div className="pt-10 border-t border-border-color grid grid-cols-2 md:grid-cols-4 gap-3">
                       <button onClick={() => setCommand('assign')} className="px-4 py-2.5 bg-white/5 hover:bg-white/10 text-white text-[10px] font-black uppercase tracking-widest rounded-lg flex items-center justify-center gap-2 transition-all border border-white/5">
                          <UserPlus className="w-4 h-4" />
                          Assign Driver
                       </button>
                       <button onClick={() => setCommand('maintenance')} className="px-4 py-2.5 bg-white/5 hover:bg-white/10 text-white text-[10px] font-black uppercase tracking-widest rounded-lg flex items-center justify-center gap-2 transition-all border border-white/5">
                          <Toolbox className="w-4 h-4" />
                          Maintenance
                       </button>
                       <button onClick={() => setActiveTab('documents')} className="px-4 py-2.5 bg-white/5 hover:bg-white/10 text-white text-[10px] font-black uppercase tracking-widest rounded-lg flex items-center justify-center gap-2 transition-all border border-white/5">
                          <FileText className="w-4 h-4" />
                          Compliance
                       </button>
                       <button onClick={() => setCommand('block')} className="px-4 py-2.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-500 text-[10px] font-black uppercase tracking-widest rounded-lg flex items-center justify-center gap-2 transition-all border border-rose-500/20">
                          <Ban className="w-4 h-4" />
                          Block Asset
                       </button>
                    </div>
                  </div>
                )}

                {activeTab === "documents" && (
                  <div className="flex flex-col h-full animate-in fade-in duration-500">
                    <div className="px-4 py-3 bg-white/[0.02] border-b border-border-color flex items-center gap-6">
                       <div className="flex items-center gap-2">
                          <div className="w-2 h-2 rounded-full bg-emerald-500" />
                          <span className="text-[9px] font-black uppercase tracking-widest text-emerald-500">{stats.compliant} Compliant</span>
                       </div>
                       <div className="flex items-center gap-2">
                          <div className="w-2 h-2 rounded-full bg-amber-500" />
                          <span className="text-[9px] font-black uppercase tracking-widest text-amber-500">{stats.expiring} Expiring Soon</span>
                       </div>
                       <div className="flex items-center gap-2">
                          <div className="w-2 h-2 rounded-full bg-rose-500" />
                          <span className="text-[9px] font-black uppercase tracking-widest text-rose-500">{stats.blocked} Blocked</span>
                       </div>
                       <div className="ml-auto flex items-center gap-3">
                          <span className="text-[9px] font-black uppercase tracking-widest text-gray-700">Operational Eligibility:</span>
                          <span className={`text-[9px] font-black uppercase tracking-widest  ${isRestricted ? 'text-rose-500' : 'text-emerald-500'}`}>
                            {isRestricted ? 'Dispatch Restricted' : 'Verified Ready'}
                          </span>
                       </div>
                    </div>

                    <div className="overflow-x-auto">
                      <table className="w-full text-left">
                        <thead>
                          <tr className="bg-white/[0.01] border-b border-border-color">
                            <th className="px-6 py-4 text-[9px] font-black text-text-muted uppercase tracking-widest">Document Registry</th>
                            <th className="px-6 py-4 text-[9px] font-black text-text-muted uppercase tracking-widest">Status</th>
                            <th className="px-6 py-4 text-[9px] font-black text-text-muted uppercase tracking-widest">Expiry Pulse</th>
                            <th className="px-6 py-4 text-[9px] font-black text-text-muted uppercase tracking-widest">Operational Impact</th>
                            <th className="px-6 py-4 text-[9px] font-black text-text-muted uppercase tracking-widest">Verification Source</th>
                            <th className="px-6 py-4 text-right"></th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-white/[0.02]">
                          {registryDocs.map((doc) => (
                            <tr key={doc.id} id={`doc-${doc.id}`} className="hover:bg-white/[0.01] group transition-all">
                              <td className="px-6 py-4">
                                 <div className="flex items-center gap-3">
                                    <div>
                                       <div className="text-sm font-black text-white uppercase  tracking-tighter leading-none mb-1">{doc.type}</div>
                                       <div className="flex gap-1.5">
                                          {['v1', 'v2', 'v3'].map(v => (
                                            <button key={v} className={`text-[7px] font-black uppercase px-1 rounded border ${v === doc.version?.split('.')[0] ? 'bg-primary/20 border-primary/40 text-primary' : 'bg-white/5 border-white/5 text-gray-700 hover:text-white transition-all'}`}>
                                              {v}
                                            </button>
                                          ))}
                                       </div>
                                    </div>
                                 </div>
                              </td>
                              <td className="px-6 py-4">
                                 <span className={`px-2 py-0.5 rounded text-[8px] font-black uppercase tracking-widest bg-${doc.color}-500/10 text-${doc.color}-500 border border-${doc.color}-500/20`}>{doc.status}</span>
                              </td>
                              <td className="px-6 py-4 text-[10px] font-black text-white  tracking-tighter uppercase">{doc.pulse}</td>
                              <td className="px-6 py-4 text-[10px] font-black uppercase  tracking-tighter text-gray-400">{doc.impact}</td>
                              <td className="px-6 py-4">
                                 <div className="text-[9px] font-black text-white uppercase tracking-tighter leading-none mb-1">{doc.meta}</div>
                                 <div className="text-[7px] font-black text-gray-800 uppercase tracking-widest">Source: Portal Device</div>
                              </td>
                              <td className="px-6 py-4 text-right">
                                 <div className="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-all translate-x-2 group-hover:translate-x-0">
                                    <button onClick={() => handleDocumentUpload(doc.id)} className="p-2 hover:bg-white/5 rounded text-gray-600 hover:text-emerald-500 transition-all" title="Upload Replacement"><Upload className="w-3.5 h-3.5" /></button>
                                    <button className="p-2 hover:bg-white/5 rounded text-gray-600 hover:text-primary transition-all"><Eye className="w-3.5 h-3.5" /></button>
                                    <button className="p-2 hover:bg-white/5 rounded text-gray-600 hover:text-white transition-all"><History className="w-3.5 h-3.5" /></button>
                                 </div>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    
                    <div className="p-8 flex justify-center border-t border-border-color bg-surface-darker/10">
                       <button 
                         onClick={() => setCommand('builder')}
                         className="text-[10px] font-black text-primary uppercase tracking-widest flex items-center gap-2 hover:underline group"
                       >
                          <Plus className="w-4 h-4 group-hover:rotate-90 transition-transform duration-300" />
                          Add Requirement
                       </button>
                    </div>

                    {/* Vitals Mini-Bar */}
                    <div className="px-6 py-3 bg-black/40 border-t border-white/5 flex items-center justify-between text-[9px] font-black uppercase tracking-[0.2em] text-gray-600">
                       <div className="flex gap-8">
                          <div className="flex items-center gap-2"><Truck className="w-3 h-3" /> <span>ODO: {vehicle.odometer} KM</span></div>
                          <div className="flex items-center gap-2"><Zap className="w-3 h-3" /> <span>FUEL: {vehicle.fuel_tank_capacity ? 'STABLE' : 'N/A'}</span></div>
                          <div className="flex items-center gap-2"><User className="w-3 h-3" /> <span>DRIVER: {currentDriver?.name || 'VACANT'}</span></div>
                       </div>
                       <div className="flex items-center gap-2 text-emerald-500/50 ">
                          <Activity className="w-3 h-3" />
                          <span>REGISTRY_SYNCED</span>
                       </div>
                    </div>
                  </div>
                )}
              </div>

              {/* Bottom Operational Registry */}
              <div className="bg-surface-darker/50 border-t border-border-color p-5">
                 <div className="flex items-center gap-2 mb-4 border-b border-white/[0.03] pb-4">
                    <Terminal className="w-4 h-4 text-gray-700" />
                    <h3 className="text-[10px] font-black text-gray-600 uppercase tracking-[0.2em] ">Operational Audit Registry</h3>
                 </div>
                 <div className="space-y-4 max-h-[160px] overflow-y-auto custom-scrollbar">
                    {auditLogs.map((log) => (
                      <div 
                        key={log.id} 
                        className="flex gap-4 text-[10px] font-mono group animate-in slide-in-from-left-2 duration-300 cursor-pointer hover:bg-white/[0.02] p-1 rounded"
                      >
                         <div className="w-16 shrink-0 font-bold text-gray-800 leading-none pt-0.5">{log.time}</div>
                         <div className="flex-1">
                            <div className="flex items-center gap-2 mb-1">
                               <div className={`w-1.5 h-1.5 rounded-full ${log.severity === 'critical' ? 'bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.3)]' : log.severity === 'warn' ? 'bg-amber-500' : 'bg-emerald-500/20'}`} />
                               <span className="font-black text-white/80 uppercase tracking-tighter ">{log.event}</span>
                               <span className="text-[8px] text-gray-900 font-black">[{log.actor}]</span>
                               <span className="text-[7px] text-gray-700 font-black tracking-widest ml-auto">{log.nodeId} | {log.sessionId}</span>
                            </div>
                            <p className="text-gray-700 leading-tight pr-10">{log.desc}</p>
                         </div>
                      </div>
                    ))}
                 </div>
              </div>
            </div>
          </div>

          {/* Right Panels: Real-time Telemetry Uplink */}
          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden h-[400px] flex flex-col shadow-xl">
              <div className="p-4 border-b border-border-color bg-surface-darker flex justify-between items-center">
                 <h3 className="text-[10px] font-black text-text-muted uppercase tracking-[0.2em] flex items-center gap-2 ">
                    <MapPin className="w-4 h-4" />
                    Signal Connection
                 </h3>
                 {isGpsOnline && (
                   <div className="flex items-center gap-2">
                      <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.4)]" />
                      <span className="text-[10px] font-black text-emerald-500 uppercase tracking-widest ">Live Signal</span>
                   </div>
                 )}
              </div>
              <div className="flex-1 relative bg-surface-darker">
                {hasGps ? (
                  <LiveOperationsMap vehicles={markers} selectedId={id} height="100%" showLegend={false} />
                ) : (
                  <div className="absolute inset-0 flex flex-col items-center justify-center p-12 text-center">
                    <Cpu className="w-10 h-10 text-text-muted opacity-10 mb-4" />
                    <div className="space-y-4 mb-8">
                       <h4 className="text-white font-black  uppercase tracking-tighter text-[11px]">Hardware Lifecycle Management</h4>
                       <div className="grid grid-cols-2 gap-3 text-[8px] font-black uppercase tracking-widest text-gray-800 text-left px-4">
                          <span>Device ID: UNMAPPED</span>
                          <span>AIS140 Req: MANDATORY</span>
                          <span className="col-span-2">Setup Progress: 0%</span>
                       </div>
                    </div>
                    <button onClick={() => { setCommand('provision'); setCommandData({ step: 1 }); addAudit('PROVISION_INIT', 'GPS hardware provisioning sequence started.', 'warn'); }} className="px-6 py-2.5 bg-primary/10 border border-primary/20 text-primary text-[10px] font-black uppercase tracking-widest rounded-lg hover:bg-primary/20 transition-all">
                       Initiate Provisioning
                    </button>
                  </div>
                )}
              </div>
            </div>

            {/* Hardware Status & Integrity Node */}
            <div className="bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl">
               <div className="flex justify-between items-center mb-6">
                  <h3 className="text-[10px] font-black text-text-muted uppercase tracking-[0.2em] flex items-center gap-2 ">
                     <Signal className="w-4 h-4" />
                     Device Health Node
                  </h3>
                  <div className="flex items-center gap-2">
                     <div className="text-[8px] font-black text-emerald-500 uppercase px-1.5 py-0.5 bg-emerald-500/10 border border-emerald-500/20 rounded">
                        Reliability: {hasGps ? '98.4%' : '0%'}
                     </div>
                  </div>
               </div>
               
               <div className="space-y-5">
                  {[
                    { label: 'Current Device State', val: hasGps ? (isGpsOnline ? 'In Transit / Active' : 'Ignition Off / Parked') : 'Device Not Attached', s: hasGps ? 'ok' : 'warn' },
                    { label: 'Connection Quality', val: hasGps ? (isGpsOnline ? 'Stable Connection' : 'Signal Lost - Untraceable') : 'N/A', s: isGpsOnline ? 'ok' : 'rose' },
                    { label: 'Data Freshness', val: telemetry?.lastUpdate ? `${Math.floor((Date.now() - new Date(telemetry.lastUpdate).getTime()) / 1000)}s Latency` : 'No Signal', s: 'ok' },
                    { label: 'Power Integrity', val: hasGps ? 'Main Battery (Active)' : 'N/A', s: 'ok' },
                    { label: 'AIS-140 Device', val: vehicle.ais140_compliant ? 'Verified Signal' : 'Pending Verification', s: vehicle.ais140_compliant ? 'ok' : 'warn' },
                  ].map((h, i) => (
                    <div key={i} className="flex justify-between items-center">
                       <span className="text-xs text-text-muted font-bold uppercase tracking-tight">{h.label}</span>
                       <span className={`text-[10px] font-black uppercase  tracking-tighter ${h.s === 'ok' ? 'text-emerald-500' : h.s === 'rose' ? 'text-rose-500' : h.s === 'warn' ? 'text-amber-500' : 'text-white'}`}>{h.val}</span>
                    </div>
                  ))}
                  
                  <div className="pt-5 border-t border-white/5 space-y-3">
                     <div className="flex justify-between items-center">
                        <span className="text-[8px] text-text-muted uppercase font-black tracking-widest">Last Signal Pulse</span>
                        <span className="text-[10px] text-white font-mono">{telemetry?.lastUpdate ? new Date(telemetry.lastUpdate).toLocaleTimeString() : 'SIGNAL_LOST'}</span>
                     </div>
                     <div className="h-1.5 w-full bg-white/5 rounded-full overflow-hidden flex gap-1">
                        {[80, 100, 90, 85, 95, 100, 90, 100].map((v, i) => (
                           <div key={i} className="h-full bg-emerald-500/30 flex-1 rounded-sm" style={{ opacity: v / 100 }} />
                        ))}
                     </div>
                     <div className="text-[7px] text-gray-800 font-black uppercase tracking-widest text-center">Rolling Signal Reliability (Last 24h)</div>
                  </div>
               </div>
            </div>
          </div>
        </div>
      </div>

      <footer className="h-10 px-6 bg-surface-darker border-t border-border-color flex items-center justify-between text-[8px] font-black text-text-muted uppercase tracking-[0.4em]">
         <span>© {new Date().getFullYear()} FlyFleet Intelligence Systems</span>
         <div className="flex items-center gap-8">
            <div className="flex items-center gap-2">
               <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_5px_rgba(16,185,129,0.4)]" />
               <span>Operational State: Verified</span>
            </div>
         </div>
      </footer>

      {conflictData && <ConflictResolutionModal />}
      {policyViolation && <PolicyBypassModal />}
    </div>
  );
};

export default VehicleDetail;
