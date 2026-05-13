import React, { useState, useEffect } from "react";
import Modal from "../common/Modal";
import { api } from "../../services/api";
import { LoadingSpinner } from "../common/LoadingSpinner";

interface Vehicle {
  id: string;
  registration_number: string;
  make: string;
  model: string;
  status: string;
  vin_number?: string;
  current_driver_id?: string;
  current_driver_name?: string;
  metadata?: {
    licensePlate?: string;
  };
}

const ZERO_UUID = '00000000-0000-0000-0000-000000000000';

interface VehicleSelectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSelect: (vehicleId: string) => void;
  currentVehicleId?: string;
}

export const VehicleSelectionModal: React.FC<VehicleSelectionModalProps> = ({
  isOpen,
  onClose,
  onSelect,
  currentVehicleId
}) => {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");

  // Debouncing logic
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
    }, 400); // 400ms debounce
    return () => clearTimeout(timer);
  }, [search]);

  // Load vehicles when modal opens or debounced search changes
  useEffect(() => {
    if (isOpen) {
      loadVehicles(debouncedSearch);
    }
  }, [isOpen, debouncedSearch]);

  const loadVehicles = async (query: string) => {
    setLoading(true);
    try {
      let res;
      if (query) {
        res = await api.searchVehicles(query, { limit: 50 });
      } else {
        res = await api.getVehicles({ limit: 50 });
      }
      
      // Axios interceptor already unwraps response.data, so res is the VehicleListResponse body
      // models.VehicleListResponse has {"vehicles": [...]}
      const data = res.vehicles || [];
      setVehicles(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Failed to load vehicles:", error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active': return 'text-emerald-400 bg-emerald-400/10 border-emerald-400/20';
      case 'idle': return 'text-blue-400 bg-blue-400/10 border-blue-400/20';
      case 'maintenance': return 'text-orange-400 bg-orange-400/10 border-orange-400/20';
      default: return 'text-slate-400 bg-slate-400/10 border-slate-400/20';
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Assign Vehicle" size="lg">
      <div className="flex flex-col gap-6">
        {/* Search Bar */}
        <div className="relative">
          <div className={`absolute left-4 top-1/2 -translate-y-1/2 transition-all ${loading ? 'opacity-0' : 'opacity-100'}`}>
            <span className="material-symbols-outlined text-slate-500">search</span>
          </div>
          {loading && (
            <div className="absolute left-4 top-1/2 -translate-y-1/2">
              <span className="material-symbols-outlined text-primary text-sm animate-spin">progress_activity</span>
            </div>
          )}
          <input
            type="text"
            placeholder="Search Registration, VIN, Make, Model..."
            className="w-full bg-slate-950/50 border border-white/10 rounded-2xl py-4 pl-12 pr-4 text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all font-medium text-sm"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoFocus
          />
        </div>

        {/* List Content */}
        <div className="min-h-[400px] max-h-[50vh] overflow-y-auto custom-scrollbar pr-2">
          {loading && vehicles.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 gap-4">
              <LoadingSpinner />
              <p className="text-slate-500 text-sm font-medium animate-pulse">Searching fleet database...</p>
            </div>
          ) : vehicles.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-slate-500 gap-3">
              <span className="material-symbols-outlined text-4xl opacity-20">minor_crash</span>
              <p className="text-sm font-medium">No vehicles found for "{search}"</p>
              <button 
                onClick={() => setSearch('')}
                className="text-primary text-xs font-bold uppercase hover:underline mt-2"
              >
                Reset Search
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-2.5">
              {vehicles.map((vehicle) => {
                const isReserved = vehicle.current_driver_id && vehicle.current_driver_id !== ZERO_UUID;
                
                return (
                <button
                  key={vehicle.id}
                  onClick={() => {
                    if (isReserved && !window.confirm(`This vehicle is RESERVED for ${vehicle.current_driver_name || 'another driver'}. Are you sure you want to override?`)) {
                      return;
                    }
                    onSelect(vehicle.id);
                  }}
                  disabled={vehicle.id === currentVehicleId}
                  className={`flex items-center justify-between p-4 rounded-2xl border transition-all group text-left ${
                    vehicle.id === currentVehicleId 
                      ? 'bg-primary/5 border-primary/20 opacity-60 cursor-not-allowed' 
                      : isReserved
                        ? 'bg-amber-500/5 border-amber-500/10 hover:border-amber-500/30'
                        : 'bg-slate-900/40 border-white/5 hover:border-primary/40 hover:bg-primary/5 active:scale-[0.98]'
                  }`}
                >
                  <div className="flex items-center gap-4">
                    <div className={`p-3 rounded-xl border ${
                      vehicle.id === currentVehicleId 
                        ? 'bg-primary text-white' 
                        : isReserved
                          ? 'bg-amber-500/20 text-amber-400 border-amber-500/20'
                          : 'bg-slate-950/50 text-slate-500 group-hover:text-primary transition-colors border-white/5 group-hover:border-primary/20'
                    }`}>
                      <span className="material-symbols-outlined text-2xl">
                        {isReserved ? 'lock_person' : 'local_shipping'}
                      </span>
                    </div>
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-white font-bold text-base tracking-tight leading-none">
                          {vehicle.registration_number || vehicle.metadata?.licensePlate || "No plate"}
                        </span>
                        {vehicle.id === currentVehicleId && (
                            <span className="text-[10px] font-bold uppercase text-primary bg-primary/10 px-1.5 py-0.5 rounded leading-none">Current</span>
                        )}
                        {isReserved && (
                            <span className="text-[10px] font-bold uppercase text-amber-400 bg-amber-400/10 px-1.5 py-0.5 rounded leading-none flex items-center gap-1 border border-amber-500/20">
                              <span className="material-symbols-outlined text-[10px]">lock</span>
                              Reserved
                            </span>
                        )}
                      </div>
                      <div className="text-[11px] text-slate-400 font-medium">
                        {vehicle.make} {vehicle.model} 
                        {vehicle.vin_number && <span className="text-slate-600 font-mono ml-2">VIN: {vehicle.vin_number}</span>}
                      </div>
                      {isReserved && (
                         <div className="mt-1 text-[10px] text-amber-400/60 font-bold uppercase flex items-center gap-1">
                           <span className="material-symbols-outlined text-[12px]">person</span>
                           Occupied by: {vehicle.current_driver_name || 'Anonymous Pilot'}
                         </div>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold uppercase border ${getStatusColor(vehicle.status)}`}>
                      {vehicle.status || 'Active'}
                    </span>
                    <span className={`material-symbols-outlined transition-colors ${isReserved ? 'text-amber-900 group-hover:text-amber-400' : 'text-slate-700 group-hover:text-primary'}`}>
                      chevron_right
                    </span>
                  </div>
                </button>
                );
              })}
            </div>
          )}
        </div>

        {/* Footer info */}
        <div className="pt-2 border-t border-slate-800/50 flex justify-between items-center text-[10px] text-slate-500 font-bold uppercase tracking-widest px-1">
          <span>{loading ? 'Searching...' : `Found ${vehicles.length} Results`}</span>
          <span className="text-primary/50">Full Database Search Enabled</span>
        </div>
      </div>
    </Modal>
  );
};
