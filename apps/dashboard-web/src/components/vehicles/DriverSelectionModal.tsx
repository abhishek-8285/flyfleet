import React, { useState, useEffect } from "react";
import Modal from "../common/Modal";
import { api } from "../../services/api";
import { LoadingSpinner } from "../common/LoadingSpinner";

interface Driver {
  id: string;
  first_name: string;
  last_name: string;
  name?: string;
  email: string;
  phone: string;
  status: string;
  current_vehicle_id?: string;
  current_vehicle_registration?: string;
}

const ZERO_UUID = '00000000-0000-0000-0000-000000000000';

interface DriverSelectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSelect: (driverId: string) => void;
  currentDriverId?: string;
}

export const DriverSelectionModal: React.FC<DriverSelectionModalProps> = ({
  isOpen,
  onClose,
  onSelect,
  currentDriverId
}) => {
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
    }, 400);
    return () => clearTimeout(timer);
  }, [search]);

  useEffect(() => {
    if (isOpen) {
      loadDrivers(debouncedSearch);
    }
  }, [isOpen, debouncedSearch]);

  const loadDrivers = async (query: string) => {
    setLoading(true);
    try {
      // In this system, getDrivers might accept query params
      const res = await api.getDrivers({ name: query, limit: 50 });
      const data = res.data?.drivers || res.drivers || [];
      setDrivers(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Failed to load drivers:", error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active': return 'text-emerald-400 bg-emerald-400/10 border-emerald-400/20';
      case 'suspended': return 'text-rose-400 bg-rose-400/10 border-rose-400/20';
      case 'inactive': return 'text-slate-400 bg-slate-400/10 border-slate-400/20';
      default: return 'text-slate-400 bg-slate-400/10 border-slate-400/20';
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Link Pilot" size="lg">
      <div className="flex flex-col gap-6">
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
            placeholder="Search Name, Phone, Email, License..."
            className="w-full bg-slate-950/50 border border-white/10 rounded-2xl py-4 pl-12 pr-4 text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all font-medium text-sm"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoFocus
          />
        </div>

        <div className="min-h-[400px] max-h-[50vh] overflow-y-auto custom-scrollbar pr-2">
          {loading && drivers.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 gap-4">
              <LoadingSpinner />
              <p className="text-slate-500 text-sm font-medium animate-pulse">Scanning pilot database...</p>
            </div>
          ) : drivers.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-slate-500 gap-3">
              <span className="material-symbols-outlined text-4xl opacity-20">person_off</span>
              <p className="text-sm font-medium">No pilots found for "{search}"</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-2.5">
              {drivers.map((driver) => {
                const isReserved = driver.current_vehicle_id && driver.current_vehicle_id !== ZERO_UUID;
                
                return (
                <button
                  key={driver.id}
                  onClick={() => {
                    if (isReserved && !window.confirm(`This pilot is RESERVED for asset ${driver.current_vehicle_registration || 'another vehicle'}. Are you sure you want to override?`)) {
                      return;
                    }
                    onSelect(driver.id);
                  }}
                  disabled={driver.id === currentDriverId}
                  className={`flex items-center justify-between p-4 rounded-2xl border transition-all group text-left ${
                    driver.id === currentDriverId 
                      ? 'bg-primary/5 border-primary/20 opacity-60 cursor-not-allowed' 
                      : isReserved
                        ? 'bg-amber-500/5 border-amber-500/10 hover:border-amber-500/30'
                        : 'bg-slate-900/40 border-white/5 hover:border-primary/40 hover:bg-primary/5 active:scale-[0.98]'
                  }`}
                >
                  <div className="flex items-center gap-4">
                    <div className={`w-12 h-12 rounded-xl border flex items-center justify-center transition-all ${
                      driver.id === currentDriverId 
                        ? 'bg-primary text-white' 
                        : isReserved
                          ? 'bg-amber-500/20 text-amber-400 border-amber-500/20'
                          : 'bg-slate-950/50 text-slate-500 group-hover:text-primary border-white/5 group-hover:border-primary/20'
                    }`}>
                      <span className="material-symbols-outlined text-2xl">
                        {isReserved ? 'car_tag' : 'person'}
                      </span>
                    </div>
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-white font-bold text-base tracking-tight leading-none">
                          {driver.name || `${driver.first_name} ${driver.last_name}`}
                        </span>
                        {driver.id === currentDriverId && (
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
                        {driver.phone} • {driver.email}
                      </div>
                      {isReserved && (
                         <div className="mt-1 text-[10px] text-amber-400/60 font-bold uppercase flex items-center gap-1">
                           <span className="material-symbols-outlined text-[12px]">local_shipping</span>
                           Asset: {driver.current_vehicle_registration || 'Unknown Vehicle'}
                         </div>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold uppercase border ${getStatusColor(driver.status)}`}>
                      {driver.status || 'Active'}
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

        <div className="pt-2 border-t border-slate-800/50 flex justify-between items-center text-[10px] text-slate-500 font-bold uppercase tracking-widest px-1">
          <span>{loading ? 'Searching...' : `Found ${drivers.length} Pilots`}</span>
        </div>
      </div>
    </Modal>
  );
};
