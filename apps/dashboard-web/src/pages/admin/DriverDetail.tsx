import React, { useEffect, useState, useCallback, useMemo } from "react";
import { useNavigate, useParams, Link } from "react-router-dom";
import { api } from "../../services/api";
import { Driver, DriverDocument, DriverAnalytics } from "../../modules/drivers/types";
import { RecentTrips } from "../../modules/drivers/components/RecentTrips";
import { DriverPerformance } from "../../modules/drivers/components/DriverPerformance";
import { LoadingSpinner } from "../../components/common/LoadingSpinner";
import { format, formatDistanceToNow } from "date-fns";

type TabType = "overview" | "behavior" | "documents";

export const DriverDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [driver, setDriver] = useState<Driver | null>(null);
  const [analytics, setAnalytics] = useState<DriverAnalytics | null>(null);
  const [documents, setDocuments] = useState<DriverDocument[]>([]);
  const [currentVehicle, setCurrentVehicle] = useState<any>(null);
  const [activeTab, setActiveTab] = useState<TabType>("overview");
  const [loading, setLoading] = useState(true);
  
  const [isReassignModalOpen, setIsReassignModalOpen] = useState(false);
  const [availableVehicles, setAvailableVehicles] = useState<any[]>([]);
  const [assigning, setAssigning] = useState(false);

  const loadData = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [driverRes, analyticsRes, docsRes] = await Promise.all([
        api.getDriver(id),
        api.getDriverAnalytics(id),
        api.getDriverDocuments(id),
      ]);

      const rawDriver = driverRes.data || driverRes;
      const rawAnalytics = analyticsRes.data || analyticsRes;
      
      setDriver(rawDriver);
      setAnalytics(rawAnalytics);
      setDocuments(docsRes?.documents || docsRes?.data || docsRes || []);

      if (rawDriver.current_vehicle_id) {
        const vehicleRes = await api.getVehicle(rawDriver.current_vehicle_id);
        setCurrentVehicle(vehicleRes.data || vehicleRes);
      }
    } catch (error) {
      console.error("Failed to load driver data:", error);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const compliance = useMemo(() => {
    const checkDoc = (type: string) => {
      const doc = documents.find(d => d.document_type.toLowerCase().includes(type.toLowerCase()));
      if (!doc) return { status: 'MISSING', color: 'text-error', icon: 'error', label: 'Missing' };
      if (!doc.verified) return { status: 'PENDING', color: 'text-amber-400', icon: 'pending', label: 'Pending' };
      return { status: 'VALID', color: 'text-emerald-400', icon: 'check_circle', label: 'Verified' };
    };
    return {
      license: checkDoc('license'),
      medical: checkDoc('medical'),
      vetting: driver?.status === 'active' ? { status: 'VALID', color: 'text-emerald-400', icon: 'check_circle', label: 'Verified' } : { status: 'PENDING', color: 'text-amber-400', icon: 'pending', label: 'Pending' }
    };
  }, [documents, driver]);

  const handleReassignOpen = async () => {
    setIsReassignModalOpen(true);
    try {
      const res = await api.getVehicles({ status: 'ACTIVE', limit: 100 });
      setAvailableVehicles(res.data || res.vehicles || []);
    } catch (err) {
      console.error("Failed to fetch vehicles:", err);
    }
  };

  const handleAssignVehicle = async (vehicleId: string) => {
    if (!id) return;
    setAssigning(true);
    try {
      await api.assignVehicle(id, vehicleId);
      await loadData();
      setIsReassignModalOpen(false);
    } catch (err) {
      alert("Failed to assign vehicle.");
    } finally {
      setAssigning(false);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-4 bg-background-dark h-full">
        <LoadingSpinner />
        <p className="text-on-surface-variant font-bold uppercase tracking-wider text-xs">Loading Driver Data...</p>
      </div>
    );
  }

  if (!driver) return null;

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark -m-6 overflow-hidden">
      
      {/* Contextual Navigation Sub-Header (Docks below Global Header) */}
      <div className="z-[30] w-full border-b border-outline bg-surface-dim/80 backdrop-blur-md flex items-center px-8 h-12 flex-shrink-0">
        <nav className="flex gap-8 h-full items-center font-body text-sm font-medium">
          <button 
            onClick={() => setActiveTab('overview')} 
            className={`h-full flex items-center ${activeTab === 'overview' ? 'text-primary border-b-2 border-primary' : 'text-on-surface-variant hover:text-white'} transition-all px-1`}
          >
            Overview
          </button>
          <button 
            onClick={() => setActiveTab('behavior')} 
            className={`h-full flex items-center ${activeTab === 'behavior' ? 'text-primary border-b-2 border-primary' : 'text-on-surface-variant hover:text-white'} transition-all px-1`}
          >
            Behavior
          </button>
          <button 
            onClick={() => setActiveTab('documents')} 
            className={`h-full flex items-center ${activeTab === 'documents' ? 'text-primary border-b-2 border-primary' : 'text-on-surface-variant hover:text-white'} transition-all px-1`}
          >
            Documents
          </button>
        </nav>
      </div>

      <main className="flex-1 overflow-y-auto p-8 custom-scrollbar">
        {/* Breadcrumbs & Actions */}
        <div className="flex justify-between items-end mb-8">
          <div>
            <nav className="flex items-center gap-2 text-xs text-on-surface-variant mb-2">
              <Link to="/admin/drivers" className="hover:text-primary transition-colors">Drivers</Link>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white font-medium">{driver.name}</span>
            </nav>
            <div className="flex items-center gap-4">
              <button 
                onClick={() => navigate("/admin/drivers")}
                className="p-1 hover:bg-surface-container rounded transition-colors"
              >
                <span className="material-symbols-outlined text-on-surface-variant">arrow_back</span>
              </button>
              <h2 className="text-2xl font-bold tracking-tight text-white">{driver.name}</h2>
              <span className={`px-2 py-0.5 rounded text-[10px] font-bold border ${driver.status === 'active' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border-rose-500/20'}`}>
                {driver.status.toUpperCase()}
              </span>
            </div>
          </div>
          <div className="flex gap-3">
            <button onClick={handleReassignOpen} className="bg-surface-container-highest border border-outline text-on-surface hover:text-white px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-2 active:scale-95 shadow-sm">
              <span className="material-symbols-outlined text-sm">swap_horiz</span>
              Reassign Vehicle
            </button>
            <button className="bg-surface-container-highest border border-outline text-error hover:bg-error/5 px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-2 active:scale-95">
              <span className="material-symbols-outlined text-sm">block</span>
              Suspend
            </button>
            <button 
              onClick={() => navigate(`/admin/drivers/${id}/edit`)}
              className="bg-primary text-white px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-2 shadow-lg shadow-primary/20 active:scale-95"
            >
              <span className="material-symbols-outlined text-sm">edit</span>
              Edit Driver
            </button>
          </div>
        </div>

        {/* Bento Stat Grid */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
          <div className="bg-surface-container rounded-xl p-5 border border-outline relative overflow-hidden group shadow-sm transition-all hover:border-primary/30">
            <div className="flex justify-between items-start mb-4">
              <div className="p-2 bg-primary/10 rounded-lg text-primary">
                <span className="material-symbols-outlined">analytics</span>
              </div>
              <span className="text-[10px] font-bold text-emerald-400">+2.4% vs last month</span>
            </div>
            <p className="text-on-surface-variant text-xs font-bold uppercase tracking-wider mb-1">Safety Score</p>
            <h3 className="text-3xl font-bold text-white">{driver.behavior_score}<span className="text-sm text-on-surface-variant ml-1">/100</span></h3>
            <div className="absolute bottom-0 left-0 h-1 bg-primary transition-all duration-1000" style={{ width: `${driver.behavior_score}%` }}></div>
          </div>

          <div className="bg-surface-container rounded-xl p-5 border border-outline group shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <div className="p-2 bg-secondary-container rounded-lg text-secondary">
                <span className="material-symbols-outlined">route</span>
              </div>
            </div>
            <p className="text-on-surface-variant text-xs font-bold uppercase tracking-wider mb-1">Total Trips</p>
            <h3 className="text-3xl font-bold text-white">{driver.total_trips}</h3>
            <p className="text-[10px] text-on-surface-variant mt-2 font-bold uppercase tracking-wider">LIFETIME METRICS</p>
          </div>

          <div className="bg-surface-container rounded-xl p-5 border border-outline group shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <div className="p-2 bg-tertiary-container/20 rounded-lg text-tertiary">
                <span className="material-symbols-outlined">distance</span>
              </div>
            </div>
            <p className="text-on-surface-variant text-xs font-bold uppercase tracking-wider mb-1">Distance</p>
            <h3 className="text-3xl font-bold text-white">{driver.total_distance_km.toLocaleString()} <span className="text-xs text-on-surface-variant">KM</span></h3>
            <p className="text-[10px] text-on-surface-variant mt-2 font-bold tracking-wider">AVG {(driver.total_distance_km / Math.max(driver.experience_years * 300, 1)).toFixed(0)} KM/DAY</p>
          </div>

          <div className="bg-surface-container rounded-xl p-5 border border-outline group shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <div className="p-2 bg-emerald-500/10 rounded-lg text-emerald-500">
                <span className="material-symbols-outlined">timer</span>
              </div>
            </div>
            <p className="text-on-surface-variant text-xs font-bold uppercase tracking-wider mb-1">On-Time Rate</p>
            <h3 className="text-3xl font-bold text-white">{analytics?.on_time_rate || 92}%</h3>
            <div className="w-full bg-surface-container-highest h-1 rounded-full mt-4">
              <div className="bg-emerald-500 h-1 rounded-full shadow-sm" style={{ width: `${analytics?.on_time_rate || 92}%` }}></div>
            </div>
          </div>
        </div>

        {/* Main Content Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
          
          {/* Left Column (2/3) */}
          <div className="lg:col-span-8 space-y-8">
            {activeTab === 'overview' && (
              <>
                <section className="bg-surface-container rounded-xl border border-outline overflow-hidden shadow-sm">
                  <div className="p-6 border-b border-outline flex justify-between items-center">
                    <h3 className="font-bold text-white flex items-center gap-2">
                      <span className="material-symbols-outlined text-primary">person</span>
                      Driver Details
                    </h3>
                    <button className="text-primary text-xs font-bold hover:underline">UPDATE RECORD</button>
                  </div>
                  <div className="p-6 grid grid-cols-2 gap-y-6 gap-x-12">
                    <div>
                      <label className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider block mb-1">Full Name</label>
                      <p className="text-sm text-white font-medium">{driver.name}</p>
                    </div>
                    <div>
                      <label className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider block mb-1">License Number</label>
                      <p className="text-sm text-white font-mono">{driver.license_number}</p>
                    </div>
                    <div>
                      <label className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider block mb-1">Phone / Email</label>
                      <p className="text-sm text-white font-medium">{driver.phone}</p>
                      <p className="text-xs text-on-surface-variant">{driver.email}</p>
                    </div>
                    <div>
                      <label className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider block mb-1">Join Date</label>
                      <p className="text-sm text-white font-medium">{format(new Date(driver.joining_date), 'MMM d, yyyy')}</p>
                      <p className="text-xs text-on-surface-variant">Tenure: {driver.experience_years} Years</p>
                    </div>
                    <div className="col-span-2">
                      <label className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider block mb-1">Address</label>
                      <p className="text-sm text-white">{driver.address?.street}, {driver.address?.city}, {driver.address?.state} {driver.address?.pincode}</p>
                    </div>
                    <div>
                      <label className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider block mb-1">Emergency Contact</label>
                      <p className="text-sm text-white font-medium">{driver.emergency_contact?.name} ({driver.emergency_contact?.relation})</p>
                      <p className="text-xs text-on-surface-variant">{driver.emergency_contact?.phone}</p>
                    </div>
                  </div>
                </section>
                <RecentTrips driverId={id!} />
              </>
            )}

            {activeTab === 'behavior' && (
              <DriverPerformance driver={driver!} analytics={analytics} />
            )}

            {activeTab === 'documents' && (
              <div className="grid grid-cols-2 gap-4">
                {documents.map((doc) => (
                  <div key={doc.id} className="bg-surface-container p-5 rounded-xl border border-outline flex flex-col justify-between group hover:border-primary/50 transition-all shadow-sm">
                    <div className="flex justify-between items-start mb-4">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-surface-container-highest rounded border border-outline text-primary">
                          <span className="material-symbols-outlined text-[20px]">{doc.document_type === 'license' ? 'badge' : 'file_present'}</span>
                        </div>
                        <div>
                          <p className="text-xs font-bold text-white uppercase tracking-tight">{doc.document_type.replace(/_/g, ' ')}</p>
                          <p className="text-[10px] text-on-surface-variant font-mono">{doc.document_number}</p>
                        </div>
                      </div>
                      <span className={`px-2 py-0.5 rounded text-[8px] font-bold uppercase border ${doc.verified ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border-amber-500/20'}`}>
                        {doc.verified ? 'Verified' : 'Pending'}
                      </span>
                    </div>
                    <div className="flex items-center justify-between mt-4">
                      <p className="text-[10px] text-on-surface-variant font-bold uppercase tracking-wider">Exp: {format(new Date(doc.expiry_date), 'MMM yyyy')}</p>
                      <button onClick={() => window.open(doc.file_url)} className="text-[10px] font-bold text-primary hover:underline uppercase tracking-wider">VIEW DOCUMENT</button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Right Column (1/3) */}
          <div className="lg:col-span-4 space-y-6">
            
            {/* Current Vehicle Assignment */}
            <div className="bg-surface-container rounded-xl border border-outline p-6 relative shadow-sm">
              <div className="flex items-center justify-between mb-6">
                <h4 className="text-xs font-bold text-on-surface-variant uppercase tracking-wider">Vehicle Assignment</h4>
                <span className="material-symbols-outlined text-primary text-lg">local_shipping</span>
              </div>
              <div className="bg-surface-container-high rounded-lg p-4 mb-4 border border-outline group shadow-inner">
                <div className="flex items-center gap-4">
                  <div className="bg-surface p-2 rounded border border-outline group-hover:bg-primary/5 transition-all">
                    <span className="material-symbols-outlined text-on-surface-variant group-hover:text-primary transition-all">airport_shuttle</span>
                  </div>
                  <div>
                    <p className="text-lg font-bold text-white tracking-tight">
                      {currentVehicle?.registration_number || (driver?.current_vehicle_id ? 'IDENTIFYING...' : 'UNASSIGNED')}
                    </p>
                    <p className="text-[10px] text-on-surface-variant uppercase tracking-wider font-bold">
                      {currentVehicle?.metadata?.make || 'UNASSIGNED'} {currentVehicle?.metadata?.model || ''}
                    </p>
                  </div>
                </div>
              </div>
              <div className="space-y-3">
                <div className="flex justify-between text-xs">
                  <span className="text-on-surface-variant font-bold">Fuel Level</span>
                  <span className="text-white font-bold">{currentVehicle?.metadata?.fuelLevel || "84"}%</span>
                </div>
                <div className="w-full bg-surface-container-high h-1 rounded-full overflow-hidden">
                  <div className="bg-primary h-1 rounded-full shadow-sm" style={{ width: `${currentVehicle?.metadata?.fuelLevel || 84}%` }}></div>
                </div>
                <div className="flex justify-between text-xs pt-1 border-t border-outline mt-4">
                  <span className="text-on-surface-variant font-bold">Health Status</span>
                  <span className="text-white font-bold uppercase text-[10px]">
                    {currentVehicle?.telemetry?.lastHeartbeat ? formatDistanceToNow(new Date(currentVehicle.telemetry.lastHeartbeat)) + ' ago' : 'OFFLINE'}
                  </span>
                </div>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="bg-surface-container rounded-xl border border-outline p-6 shadow-sm">
              <h4 className="text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-6">Quick Actions</h4>
              <div className="grid grid-cols-2 gap-4">
                <button className="flex flex-col items-center justify-center p-4 bg-surface-container-high border border-outline rounded-lg hover:border-primary transition-all group active:scale-95 shadow-sm">
                  <span className="material-symbols-outlined text-primary mb-2 group-hover:scale-110 transition-transform">call</span>
                  <span className="text-[10px] font-bold text-white uppercase tracking-wider">Voice Call</span>
                </button>
                <button className="flex flex-col items-center justify-center p-4 bg-surface-container-high border border-outline rounded-lg hover:border-primary transition-all group active:scale-95 shadow-sm">
                  <span className="material-symbols-outlined text-primary mb-2 group-hover:scale-110 transition-transform">chat_bubble</span>
                  <span className="text-[10px] font-bold text-white uppercase tracking-wider">Messages</span>
                </button>
              </div>
            </div>

            {/* Compliance Matrix */}
            <div className="bg-surface-container rounded-xl border border-outline p-6 shadow-sm">
              <div className="flex items-center justify-between mb-6">
                <h4 className="text-xs font-bold text-on-surface-variant uppercase tracking-wider">Compliance Status</h4>
                <span className={`px-2 py-0.5 rounded text-[10px] font-bold border uppercase tracking-wider ${compliance.license.status === 'VALID' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border-rose-500/20'}`}>
                  {driver?.status === 'active' ? 'APPROVED' : 'PENDING'}
                </span>
              </div>
              <ul className="space-y-4">
                <li className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className={`material-symbols-outlined text-sm ${compliance.license.color}`}>{compliance.license.icon}</span>
                    <span className="text-sm text-on-surface-variant">DL Validity</span>
                  </div>
                  <span className="text-xs text-white font-mono uppercase tracking-tighter">{compliance.license.label}</span>
                </li>
                <li className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className={`material-symbols-outlined text-sm ${compliance.medical.color}`}>{compliance.medical.icon}</span>
                    <span className="text-sm text-on-surface-variant">Health Cert</span>
                  </div>
                  <span className="text-xs text-white font-mono uppercase tracking-tighter">{compliance.medical.label}</span>
                </li>
                <li className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className={`material-symbols-outlined text-sm ${compliance.vetting.color}`}>{compliance.vetting.icon}</span>
                    <span className="text-sm text-on-surface-variant">Identity Vetting</span>
                  </div>
                  <span className="text-xs text-white font-mono uppercase tracking-tighter">{compliance.vetting.label}</span>
                </li>
              </ul>
            </div>

            {/* System Metadata */}
            <div className="bg-surface-container rounded-xl border border-outline p-6 bg-gradient-to-br from-surface-container to-surface-dim shadow-sm">
              <h4 className="text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-4">Internal Metadata</h4>
              <div className="space-y-2">
                <div className="flex justify-between items-center">
                  <span className="text-[10px] text-on-surface-variant uppercase font-bold">Driver ID</span>
                  <span className="text-xs font-mono text-white">#{driver.id.slice(0, 8).toUpperCase()}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-[10px] text-on-surface-variant uppercase font-bold">System Ref</span>
                  <span className="text-xs font-mono text-white">FLT-D-{driver.id.slice(0, 5).toUpperCase()}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-[10px] text-on-surface-variant uppercase font-bold">Schema</span>
                  <span className="text-xs font-mono text-white">v4.2.1-prod</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Reassignment Modal */}
      {isReassignModalOpen && (
        <div className="absolute inset-0 z-50 flex items-center justify-center bg-surface/80 backdrop-blur-sm p-6 overflow-hidden">
          <div className="bg-surface-container border border-outline shadow-2xl rounded-2xl w-full max-w-lg max-h-full flex flex-col overflow-hidden animate-entrance">
            <div className="p-6 border-b border-outline flex justify-between items-center bg-surface-container-high">
              <div>
                <h3 className="text-white font-bold text-sm uppercase tracking-wider">Vehicle Reassignment</h3>
              </div>
              <button 
                onClick={() => setIsReassignModalOpen(false)}
                className="p-2 hover:bg-white/5 rounded-lg text-on-surface-variant hover:text-white transition-all"
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>
            <div className="p-4 flex-1 overflow-y-auto custom-scrollbar space-y-2">
              {availableVehicles.map((v) => (
                <button
                  key={v.id}
                  disabled={assigning}
                  onClick={() => handleAssignVehicle(v.id)}
                  className="w-full flex items-center justify-between p-4 bg-surface-container-low border border-outline rounded-xl hover:border-primary/50 transition-all group"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-surface rounded border border-outline">
                      <span className="material-symbols-outlined text-on-surface-variant group-hover:text-primary transition-colors">local_shipping</span>
                    </div>
                    <div className="text-left">
                      <p className="text-sm font-bold text-white group-hover:text-primary transition-colors">{v.registration_number}</p>
                      <p className="text-[10px] text-on-surface-variant font-bold uppercase tracking-wider">{v.metadata?.make} {v.metadata?.model}</p>
                    </div>
                  </div>
                  <span className="material-symbols-outlined text-outline group-hover:text-primary transition-colors">chevron_right</span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
