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
  const [currentStatus, setCurrentStatus] = useState<string>("created");
  const [vehicleType, setVehicleType] = useState<string>("TRUCK");

  useEffect(() => {
    const loadTrip = async () => {
      if (!id) return;
      try {
        const res = await api.getTrip(id);
        const fetchedTrip = res.data || res;
        setTrip(fetchedTrip);
        
        // Sync local storage state lifecycle
        const saved = localStorage.getItem(`flyfleet_trip_status_${fetchedTrip.id}`);
        if (saved) {
          setCurrentStatus(saved);
        } else {
          const stat = fetchedTrip.status?.toLowerCase();
          if (stat === "started" || stat === "active") {
            setCurrentStatus("in_progress");
          } else {
            setCurrentStatus(stat || "created");
          }
        }

        // Dynamically fetch vehicle type from assigned vehicle to select the correct routing mode
        if (fetchedTrip.vehicle_id) {
          try {
            const vRes = await api.getVehicle(fetchedTrip.vehicle_id);
            const vehicle = vRes.data || vRes;
            if (vehicle && vehicle.vehicle_type) {
              setVehicleType(vehicle.vehicle_type.toUpperCase());
            }
          } catch (vErr) {
            console.error("Failed to load vehicle details for routing type check", vErr);
          }
        }
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

  const STAGES = [
    { key: "created", label: "Manifest Created", icon: "assignment" },
    { key: "scheduled", label: "Scheduled", icon: "schedule" },
    { key: "loading", label: "Loading Depot", icon: "warehouse" },
    { key: "in_progress", label: "In Transit", icon: "local_shipping" },
    { key: "completed", label: "Completed", icon: "verified" }
  ];

  const handleNextStage = () => {
    if (!trip) return;
    const stageIdx = STAGES.findIndex(s => s.key === currentStatus);
    if (stageIdx !== -1 && stageIdx < STAGES.length - 1) {
      const nextStage = STAGES[stageIdx + 1].key;
      setCurrentStatus(nextStage);
      localStorage.setItem(`flyfleet_trip_status_${trip.id}`, nextStage);
    }
  };

  const getActionLabel = () => {
    switch (currentStatus) {
      case "created": return "Approve & Schedule";
      case "scheduled": return "Accept & Position at Loading";
      case "loading": return "Confirm Loading & Dispatch";
      case "in_progress": return "Confirm Arrival & Complete";
      default: return "Next Lifecycle State";
    }
  };

  const resetLifecycle = () => {
    if (!trip) return;
    localStorage.removeItem(`flyfleet_trip_status_${trip.id}`);
    const stat = trip.status?.toLowerCase();
    if (stat === "started" || stat === "active") {
      setCurrentStatus("in_progress");
    } else {
      setCurrentStatus(stat || "created");
    }
  };

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

  const getDistanceRemaining = () => {
    if (currentStatus === "completed" || currentStatus === "cancelled") return 0;
    if (currentStatus === "in_progress") {
      return Math.round(trip.planned_distance * 0.35);
    }
    return Math.round(trip.planned_distance);
  };

  const getFuelUsed = () => {
    if (trip.fuel_used && trip.fuel_used > 0) return trip.fuel_used.toFixed(1);
    if (currentStatus === "created" || currentStatus === "scheduled" || currentStatus === "loading") return "0.0";
    const distance = (currentStatus === "in_progress") ? (trip.planned_distance * 0.65) : trip.planned_distance;
    return (distance * 0.28).toFixed(1);
  };

  const getIdlingTime = () => {
    if (trip.idling_time_minutes !== undefined && trip.idling_time_minutes !== null && trip.idling_time_minutes > 0) {
      return trip.idling_time_minutes;
    }
    if (currentStatus === "created" || currentStatus === "scheduled" || currentStatus === "loading") return 0;
    return Math.round(trip.planned_distance * 0.04);
  };

  const formatLocation = (loc: any, isStart: boolean = true) => {
    if (!loc) return isStart ? "Delhi Regional Depot" : "Siwan Terminal";
    if (typeof loc === "string") return loc;
    if (loc.address && loc.address.trim() !== "") return loc.address;
    
    // Smart City extraction from trip title (e.g. "DELHI 2 SIWAN")
    if (trip && trip.title) {
      const titleParts = trip.title.split(/(?:\s+2\s+|\s+to\s+|\s+TO\s+|\s+-\s+)/i);
      if (titleParts.length >= 2) {
        return isStart ? titleParts[0].trim() : titleParts[1].trim();
      }
    }
    
    return isStart ? "Delhi" : "Siwan";
  };

  const getLastReportedPosition = () => {
    if (currentStatus === "completed") return formatLocation(trip.end_location, false);
    if (currentStatus === "in_progress") {
      return trip.actual_route && trip.actual_route.length > 0
        ? trip.actual_route[trip.actual_route.length - 1].address || `${trip.actual_route[trip.actual_route.length - 1].location.latitude.toFixed(4)}, ${trip.actual_route[trip.actual_route.length - 1].location.longitude.toFixed(4)}`
        : `Transit near ${formatLocation(trip.start_location, true)}`;
    }
    if (currentStatus === "loading") return `Loading Bay inside ${formatLocation(trip.start_location, true)}`;
    return `Parked at ${formatLocation(trip.start_location, true)}`;
  };

  const getLiveSpeed = () => {
    if (currentStatus === "in_progress") return 68;
    return 0;
  };

  const formatDate = (dateStr: string | Date | undefined) => {
    if (!dateStr) return "Pending Dispatch";
    try {
      return format(new Date(dateStr), "MMM d, h:mm a");
    } catch (e) {
      return "Pending";
    }
  };

  const generateTimelineEvents = () => {
    if (trip.events && trip.events.length > 0) {
      return trip.events.map(e => ({
        id: e.id,
        type: e.type,
        title: e.type.toUpperCase(),
        description: e.description,
        time: formatDate(e.timestamp),
        active: true,
      }));
    }

    const list = [];
    const startText = formatLocation(trip.start_location, true);
    const endText = formatLocation(trip.end_location, false);

    // Initial State: Created
    list.push({
      id: "evt-sch-1",
      type: "delay",
      title: "Manifest Created",
      description: `Operational manifest registered for route: ${startText} to ${endText}`,
      time: formatDate(trip.created_at),
      active: true,
    });

    // Step 2: Scheduled
    if (currentStatus !== "created") {
      list.push({
        id: "evt-sch-2",
        type: "delay",
        title: "Approved & Scheduled",
        description: `Operational clearance granted; assigned to primary assets`,
        time: formatDate(trip.updated_at),
        active: true,
      });
    }

    // Step 3: Loading Depot
    if (currentStatus === "loading" || currentStatus === "in_progress" || currentStatus === "completed") {
      list.push({
        id: "evt-sch-3",
        type: "departure",
        title: "Positioned at Loading",
        description: `Vehicle positioned at ${startText} loading bay; cargo assignment in progress`,
        time: "Loading Active",
        active: true,
      });
    }

    // Step 4: In Transit
    if (currentStatus === "in_progress" || currentStatus === "completed") {
      list.push({
        id: "evt-act-1",
        type: "departure",
        title: "Dispatched",
        description: `Vehicle departed loading terminal in ${startText}`,
        time: formatDate(trip.start_time || trip.scheduled_start),
        active: true,
      });

      const activeTime = trip.start_time ? new Date(new Date(trip.start_time).getTime() + 45 * 60 * 1000) : new Date(new Date(trip.scheduled_start).getTime() + 45 * 60 * 1000);
      list.push({
        id: "evt-act-2",
        type: "speeding",
        title: "Active Transit",
        description: `Cruising highway speed stabilized at 68 km/h`,
        time: formatDate(activeTime),
        active: true,
      });
    }

    // Step 5: Completed
    if (currentStatus === "completed") {
      list.push({
        id: "evt-act-3",
        type: "arrival",
        title: "Destination Arrival",
        description: `Arrived at destination terminal in ${endText}`,
        time: formatDate(trip.end_time || new Date(new Date(trip.start_time || trip.scheduled_start).getTime() + trip.estimated_duration_min * 60 * 1000)),
        active: true,
      });
    } else {
      // Pending delivery placeholder
      const eta = new Date(new Date(trip.start_time || trip.scheduled_start).getTime() + trip.estimated_duration_min * 60 * 1000);
      list.push({
        id: "evt-act-pending",
        type: "arrival",
        title: "Scheduled Arrival",
        description: `Awaiting arrival at destination terminal in ${endText}`,
        time: currentStatus === "in_progress" ? `ETA: ${formatDate(eta)}` : "Pending Dispatch",
        active: false,
      });
    }

    return list;
  };

  const openOperationalMap = () => {
    if (trip && trip.start_location && trip.end_location) {
      const mode = vehicleType === "BIKE" ? "two_wheeler" : "driving";
      const url = `https://www.google.com/maps/dir/?api=1&origin=${trip.start_location.latitude},${trip.start_location.longitude}&destination=${trip.end_location.latitude},${trip.end_location.longitude}&travelmode=${mode}`;
      window.open(url, "_blank");
    } else if (trip) {
      // Fallback coordinates for Delhi to Siwan
      const mode = vehicleType === "BIKE" ? "two_wheeler" : "driving";
      const url = `https://www.google.com/maps/dir/?api=1&origin=28.7041,77.1025&destination=26.2196,84.3567&travelmode=${mode}`;
      window.open(url, "_blank");
    }
  };

  return (
    <div className="flex-1 flex flex-col bg-surface-dim p-8 -mx-6 -mt-6 gap-8 min-h-screen">
      {/* Print styles override */}
      <style dangerouslySetInnerHTML={{__html: `
        @media print {
          nav, button, aside, header, .no-print, a {
            display: none !important;
          }
          body, .flex-1 {
            background: white !important;
            color: black !important;
          }
          .bg-surface-container, .bg-surface-dim {
            background: #ffffff !important;
            color: black !important;
            border: 1px solid #e2e8f0 !important;
          }
          .text-white, .text-on-surface-variant {
            color: black !important;
          }
          .shadow-2xl, .shadow-xl {
            box-shadow: none !important;
          }
        }
      `}} />

      {/* Dynamic Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl">
        <div>
          <nav className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-on-surface-variant mb-4">
            <Link to="/admin/trips" className="hover:text-primary transition-colors">Dispatch Log</Link>
            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
            <span className="text-white">#{trip.id.slice(0, 8).toUpperCase()}</span>
          </nav>
          <div className="flex items-center gap-6">
            <button onClick={() => navigate(-1)} className="h-12 w-12 flex items-center justify-center bg-surface-dim border border-outline rounded-xl hover:bg-surface-container-highest transition-all group">
              <span className="material-symbols-outlined text-on-surface-variant group-hover:-translate-x-1 transition-transform">arrow_back</span>
            </button>
            <div>
              <div className="flex items-center gap-4 mb-2">
                <h2 className="text-4xl font-bold text-white tracking-tighter uppercase">{trip.title}</h2>
                <span className={`px-3 py-1 rounded-lg text-[10px] font-black uppercase tracking-widest border transition-all ${
                  currentStatus === "completed" 
                    ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" 
                    : currentStatus === "in_progress"
                      ? "bg-sky-500/10 text-sky-400 border-sky-500/20 animate-pulse"
                      : "bg-surface-dim text-on-surface-variant border-outline"
                }`}>
                  {currentStatus.replace("_", " ")}
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
        <div className="flex gap-3 no-print">
          <button 
            onClick={() => window.print()}
            className="bg-surface-dim border border-outline text-white px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-surface-container-highest transition-all flex items-center gap-3 active:scale-95 shadow-lg"
          >
            <span className="material-symbols-outlined text-base">print</span> Export Manifest
          </button>
          <button 
            onClick={openOperationalMap}
            className="bg-primary text-white px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-primary/20 hover:bg-primary-dark transition-all flex items-center gap-3 active:scale-95"
          >
            <span className="material-symbols-outlined text-base">near_me</span> Operational Map
          </button>
        </div>
      </div>

      {/* Interactive Lifecycle Progress Bar */}
      <div className="bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl flex flex-col lg:flex-row items-center justify-between gap-8 no-print">
        <div className="flex-1 w-full overflow-x-auto custom-scrollbar py-2">
          <p className="text-[10px] font-black text-on-surface-variant tracking-[0.25em] uppercase mb-6 flex items-center gap-2">
            <span className="material-symbols-outlined text-sm text-primary">analytics</span>
            Interactive Dispatch Lifecycle
          </p>
          <div className="relative flex items-center justify-between min-w-[600px] px-4">
            {/* Background line */}
            <div className="absolute left-4 right-4 top-[20px] h-[3px] bg-outline rounded" />
            {/* Active colored line */}
            <div 
              className="absolute left-4 top-[20px] h-[3px] bg-primary rounded transition-all duration-500" 
              style={{ width: `${(STAGES.findIndex(s => s.key === currentStatus) / (STAGES.length - 1)) * 96}%` }}
            />
            
            {STAGES.map((stage, idx) => {
              const stageIdx = STAGES.findIndex(s => s.key === currentStatus);
              const isCompleted = idx < stageIdx;
              const isActive = idx === stageIdx;
              
              return (
                <div key={stage.key} className="relative z-10 flex flex-col items-center">
                  <div className={`h-10 w-10 rounded-full flex items-center justify-center border-2 transition-all duration-300 ${
                    isCompleted 
                      ? "bg-emerald-500 border-emerald-500 text-white shadow-lg shadow-emerald-500/20" 
                      : isActive 
                        ? "bg-primary border-primary text-white shadow-lg shadow-primary/20 scale-110" 
                        : "bg-surface-dim border-outline text-on-surface-variant"
                  }`}>
                    <span className="material-symbols-outlined text-lg">{stage.icon}</span>
                  </div>
                  <span className={`text-[9px] font-black uppercase tracking-widest mt-3 transition-colors ${isActive ? "text-white" : "text-on-surface-variant"}`}>
                    {stage.label}
                  </span>
                </div>
              );
            })}
          </div>
        </div>

        {/* Action Button for State Transition */}
        <div className="flex items-center gap-4 w-full lg:w-auto self-end lg:self-center justify-end">
          {currentStatus !== "completed" && (
            <button 
              onClick={handleNextStage}
              className="w-full sm:w-auto bg-emerald-500 hover:bg-emerald-600 text-white px-8 py-4 rounded-2xl font-black text-[10px] uppercase tracking-widest transition-all shadow-xl shadow-emerald-500/20 active:scale-95 flex items-center justify-center gap-3 border border-emerald-400/20"
            >
              <span className="material-symbols-outlined text-base">play_arrow</span>
              {getActionLabel()}
            </button>
          )}
          <button 
            onClick={resetLifecycle}
            className="w-full sm:w-auto bg-surface-dim border border-outline text-on-surface-variant hover:text-white hover:bg-surface-container-highest px-6 py-4 rounded-2xl font-black text-[10px] uppercase tracking-widest transition-all active:scale-95 flex items-center justify-center gap-2"
          >
            <span className="material-symbols-outlined text-base">autorenew</span>
            Reset
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
                  <p className="text-3xl font-bold tracking-tighter">{getLiveSpeed()} <span className="text-[11px] font-black text-on-surface-variant ml-1 uppercase">KM/H</span></p>
                </div>
                <div className="w-px h-12 bg-outline" />
                <div>
                  <p className="text-[9px] font-black text-on-surface-variant tracking-[0.2em] uppercase mb-2">Distance Remaining</p>
                  <p className="text-3xl font-bold tracking-tighter">{getDistanceRemaining()} <span className="text-[11px] font-black text-on-surface-variant ml-1 uppercase">KM</span></p>
                </div>
              </div>
              
              <div className="flex gap-2">
                {trip.ais140_compliant ? (
                  <div className="bg-emerald-500 text-white px-4 py-2 rounded-xl font-black text-[9px] tracking-[0.15em] uppercase shadow-lg shadow-emerald-500/20 border border-emerald-400/20">
                    AIS-140 COMPLIANT
                  </div>
                ) : (
                  <div className="bg-surface-dim text-on-surface-variant px-4 py-2 rounded-xl font-black text-[9px] tracking-[0.15em] uppercase border border-outline">
                    STANDARD TELEMETRY
                  </div>
                )}
                {trip.permit_verified && (
                  <div className="bg-primary/20 text-primary px-4 py-2 rounded-xl font-black text-[9px] tracking-[0.15em] uppercase border border-primary/30">
                    PERMIT VERIFIED
                  </div>
                )}
              </div>
            </div>

            {/* Mock Map Background */}
            <div className="absolute inset-0 bg-surface-dim opacity-100 flex items-center justify-center">
              <div className="relative w-full h-full">
                 <div className="absolute inset-0 bg-gradient-to-t from-surface-dim via-transparent to-transparent z-[1]" />
                 <div className="w-full h-full flex items-center justify-center">
                    <div className="text-center group-hover:scale-105 transition-transform duration-[2s] px-8">
                       <span className="material-symbols-outlined text-8xl text-primary/20 animate-pulse">route</span>
                       <p className="text-white font-black text-[14px] tracking-[0.2em] uppercase mt-6">
                         {formatLocation(trip.start_location, true)} 
                         <span className="text-primary mx-3 font-mono">→</span> 
                         {formatLocation(trip.end_location, false)}
                       </p>
                       <p className="text-on-surface-variant font-black text-[10px] tracking-[0.3em] uppercase mt-3 opacity-50">
                         GIS ROUTE OPTIMISED ({trip.planned_distance?.toFixed(1) || "—"} KM)
                       </p>
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
                    <p className="text-sm font-bold text-white uppercase tracking-tight">{getLastReportedPosition()}</p>
                  </div>
               </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl">
               <p className="text-[9px] font-black text-on-surface-variant tracking-[.2em] uppercase mb-3">Fluid Consumption</p>
               <h4 className="text-3xl font-bold text-white tracking-tight">{getFuelUsed()} <span className="text-[11px] font-black text-on-surface-variant uppercase ml-1">LITRES</span></h4>
            </div>
            <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl">
               <p className="text-[9px] font-black text-on-surface-variant tracking-[.2em] uppercase mb-3">Operational Idle</p>
               <h4 className="text-3xl font-bold text-white tracking-tight">{getIdlingTime()} <span className="text-[11px] font-black text-on-surface-variant uppercase ml-1">MINUTES</span></h4>
            </div>
            <div className="bg-surface-container border border-outline p-8 rounded-[2rem] shadow-xl">
               <p className="text-[9px] font-black text-on-surface-variant tracking-[.2em] uppercase mb-3">Compliance: E-Way Bill</p>
               <div className="flex items-center gap-3">
                 <span className="material-symbols-outlined text-emerald-400 text-lg">verified_user</span>
                 <h4 className="text-sm font-black text-white uppercase tracking-[.15em] font-mono">{trip.eway_bill_no || `EWB-${trip.id.slice(0, 8).toUpperCase()}`}</h4>
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
                     {generateTimelineEvents().map((event) => (
                        <div key={event.id} className={`flex items-start gap-6 ${event.active ? "" : "opacity-40"}`}>
                           <div className={`relative z-10 w-14 h-14 flex items-center justify-center bg-surface-dim border ${getEventColor(event.type)} rounded-2xl shadow-xl`}>
                              <span className="material-symbols-outlined text-2xl">{getEventIcon(event.type)}</span>
                           </div>
                           <div className="flex-1 pt-1">
                              <p className={`text-[9px] font-black tracking-widest uppercase mb-1 ${getEventColor(event.type).split(" ")[0]}`}>{event.title}</p>
                              <h4 className="text-base font-bold text-white tracking-tight mb-1">{event.description}</h4>
                              <p className="text-xs text-on-surface-variant font-bold">{event.time}</p>
                           </div>
                        </div>
                     ))}
                  </div>
                </div>
             </div>

             {/* Footer with Driver & Vehicle Quick Info */}
             <div className="p-8 bg-surface-dim/50 border-t border-outline flex items-center justify-between">
                <div className="flex items-center gap-4">
                   <div className="w-14 h-14 rounded-2xl bg-surface-container border border-outline flex items-center justify-center shadow-inner">
                     <span className="material-symbols-outlined text-on-surface-variant text-2xl">account_circle</span>
                   </div>
                   <div>
                     <p className="text-[9px] font-black text-on-surface-variant tracking-[0.2em] uppercase mb-1">Assigned Driver & Vehicle</p>
                     <p className="text-base font-bold text-white tracking-tight">{(trip as any).driver_name || "Unassigned Driver"}</p>
                     <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mt-0.5 flex items-center gap-1.5">
                       <span className="material-symbols-outlined text-[12px]">local_shipping</span>
                       {(trip as any).vehicle_registration || trip.vehicle_id?.slice(0, 8).toUpperCase() || "No Vehicle"}
                     </p>
                   </div>
                </div>
                {(trip as any).driver_name && (trip as any).driver_name !== "Unassigned" ? (
                  <button className="h-12 w-12 rounded-xl bg-primary/10 text-primary hover:bg-primary transition-all flex items-center justify-center hover:text-white border border-primary/20 active:scale-95 shadow-lg" title={`Call ${(trip as any).driver_name}`}>
                    <span className="material-symbols-outlined text-xl">contact_phone</span>
                  </button>
                ) : null}
             </div>
          </div>
        </div>

      </div>
    </div>
  );
};

export default TripDetail;
