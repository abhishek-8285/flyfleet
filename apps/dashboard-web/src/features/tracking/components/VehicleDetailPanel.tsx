// components/VehicleDetailPanel.tsx

import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '../../../store';
import { setSelectedVehicle } from '../../../store/telemetrySlice';
import { statusColor, statusLabel } from '../utils/statusColor';

export const VehicleDetailPanel: React.FC = () => {
    const dispatch = useDispatch();
    const selectedVehicleId = useSelector((state: RootState) => state.telemetry.selectedVehicleId);
    const vehicles = useSelector((state: RootState) => state.telemetry.vehicles);

    if (!selectedVehicleId || !vehicles[selectedVehicleId]) {
        return null;
    }

    const vehicle = vehicles[selectedVehicleId];
    const color = statusColor((vehicle.derived_status || vehicle.status || 'offline') as any);

    return (
        <div className="fixed right-0 top-0 h-full w-[450px] bg-surface-container border-l border-outline shadow-2xl z-20 flex flex-col overflow-hidden animate-in slide-in-from-right duration-500">
            {/* Intel Panel Header */}
            <div className="p-8 border-b border-outline bg-surface-dim">
                <div className="flex justify-between items-start">
                    <div>
                        <nav className="text-[9px] font-black uppercase tracking-[0.3em] text-on-surface-variant mb-4 flex items-center gap-2">
                            <span className="material-symbols-outlined text-sm">satellite_alt</span>
                            Asset Telemetry Data
                        </nav>
                        <h2 className="text-4xl font-bold text-white tracking-tighter uppercase">
                            {vehicle.registrationNumber || vehicle.registration || `Asset ${String(vehicle.id).substring(0, 8)}`}
                        </h2>
                    </div>
                    <button
                        onClick={() => dispatch(setSelectedVehicle(null))}
                        className="h-10 w-10 flex items-center justify-center rounded-xl bg-surface-container-highest border border-outline text-white hover:bg-rose-500 transition-all active:scale-95"
                    >
                        <span className="material-symbols-outlined">close</span>
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-8 space-y-8 custom-scrollbar">
                    {/* Logistics Status Indicator */}
                    <div className="bg-surface-dim border border-outline p-6 rounded-[2rem] flex items-center justify-between">
                        <div className="flex items-center gap-4">
                            <div className="relative">
                                <div 
                                    className="w-3 h-3 rounded-full z-10"
                                    style={{ backgroundColor: color }}
                                />
                                <div 
                                    className="absolute inset-0 w-3 h-3 rounded-full animate-ping opacity-30"
                                    style={{ backgroundColor: color }}
                                />
                            </div>
                            <span className="text-[11px] font-black uppercase tracking-[0.2em] text-white">
                                {statusLabel(vehicle.derived_status || (vehicle as any).status || "offline")}
                            </span>
                        </div>
                        <span className="text-[9px] font-black uppercase tracking-widest text-on-surface-variant opacity-50">
                            Real-time Status
                        </span>
                    </div>

                    {/* Asset Specifications */}
                    <div className="bg-surface-dim border border-outline p-6 rounded-[2rem]">
                        <div className="flex items-center gap-3 mb-6 opacity-50">
                            <span className="material-symbols-outlined text-lg text-primary">precision_manufacturing</span>
                            <h3 className="text-[10px] font-black text-white uppercase tracking-widest">Asset Specifications</h3>
                        </div>
                        <div className="grid grid-cols-2 gap-6">
                            <div className="space-y-1">
                                <span className="text-[9px] font-black uppercase tracking-widest text-on-surface-variant">Make</span>
                                <p className="text-sm font-bold text-white tracking-tight">{vehicle.make}</p>
                            </div>
                            <div className="space-y-1">
                                <span className="text-[9px] font-black uppercase tracking-widest text-on-surface-variant">Architecture</span>
                                <p className="text-sm font-bold text-white tracking-tight">{vehicle.model}</p>
                            </div>
                            <div className="space-y-1">
                                <span className="text-[9px] font-black uppercase tracking-widest text-on-surface-variant">Operational Year</span>
                                <p className="text-sm font-bold text-white tracking-tight font-mono">{vehicle.year}</p>
                            </div>
                            <div className="space-y-1">
                                <span className="text-[9px] font-black uppercase tracking-widest text-on-surface-variant">Telemetry Update</span>
                                <p className="text-sm font-bold text-primary tracking-tight font-mono">
                                    {new Date(vehicle.lastUpdate || vehicle.last_update || 0).toLocaleTimeString()}
                                </p>
                            </div>
                        </div>
                    </div>

                    {/* Positional Intelligence */}
                    <div className="bg-surface-dim border border-outline p-6 rounded-[2rem]">
                        <div className="flex items-center gap-3 mb-6 opacity-50">
                            <span className="material-symbols-outlined text-lg text-primary">explore</span>
                            <h3 className="text-[10px] font-black text-white uppercase tracking-widest">Positional Intelligence</h3>
                        </div>
                        <div className="space-y-6">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="bg-surface-container rounded-2xl p-4 border border-outline/50">
                                    <span className="text-[8px] font-black uppercase tracking-widest text-on-surface-variant block mb-1">LATITUDE</span>
                                    <p className="font-mono text-xs font-bold text-white tracking-tighter">{(Number(vehicle.latitude) || 0).toFixed(6)}</p>
                                </div>
                                <div className="bg-surface-container rounded-2xl p-4 border border-outline/50">
                                    <span className="text-[8px] font-black uppercase tracking-widest text-on-surface-variant block mb-1">LONGITUDE</span>
                                    <p className="font-mono text-xs font-bold text-white tracking-tighter">{(Number(vehicle.longitude) || 0).toFixed(6)}</p>
                                </div>
                            </div>
                            {vehicle.speed_kmh !== undefined && (
                                <div className="flex items-center justify-between px-2">
                                    <div className="flex items-center gap-3">
                                        <span className="material-symbols-outlined text-xl text-amber-500">speed</span>
                                        <span className="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">Velocity Spectrum</span>
                                    </div>
                                    <p className="text-xl font-bold text-white tracking-tighter">
                                        <span className="text-2xl">{(vehicle.speed_kmh ?? vehicle.speed ?? 0).toFixed(1)}</span> 
                                        <span className="text-[10px] ml-1 opacity-50">KM/H</span>
                                    </p>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Personnel Operations */}
                    {vehicle.driver_name && (
                        <div className="bg-surface-dim border border-outline p-6 rounded-[2rem]">
                            <div className="flex items-center gap-3 mb-6 opacity-50">
                                <span className="material-symbols-outlined text-lg text-primary">badge</span>
                                <h3 className="text-[10px] font-black text-white uppercase tracking-widest">Personnel Operations</h3>
                            </div>
                            <div className="flex items-center gap-4">
                                <div className="h-12 w-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary font-black uppercase">
                                    {vehicle.driver_name.charAt(0)}
                                </div>
                                <div>
                                    <p className="text-sm font-bold text-white tracking-tight">{vehicle.driver_name}</p>
                                    {vehicle.driver_phone && (
                                        <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest font-mono mt-1">{vehicle.driver_phone}</p>
                                    )}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Consumable Analytics */}
                    {(vehicle.fuel_level !== undefined && vehicle.fuel_level !== null) && (
                        <div className="bg-surface-dim border border-outline p-6 rounded-[2rem]">
                            <div className="flex items-center gap-3 mb-6 opacity-50">
                                <span className="material-symbols-outlined text-lg text-primary">ev_station</span>
                                <h3 className="text-[10px] font-black text-white uppercase tracking-widest">Consumable Analytics</h3>
                            </div>
                            <div className="space-y-4">
                                <div className="flex justify-between items-end">
                                    <span className="text-[9px] font-black uppercase tracking-widest text-on-surface-variant">Fuel Reservoir</span>
                                    <span className="text-sm font-bold text-white font-mono">{(Number(vehicle.fuel_level) || 0).toFixed(1)}%</span>
                                </div>
                                <div className="w-full bg-surface-container rounded-full h-3 overflow-hidden border border-outline/50">
                                    <div
                                        className={`h-full transition-all duration-1000 ${
                                            (vehicle.fuel_level || 0) < 20 ? 'bg-rose-500 shadow-[0_0_15px_rgba(244,63,94,0.5)]' : 'bg-primary'
                                        }`}
                                        style={{ width: `${vehicle.fuel_level || 0}%` }}
                                    />
                                </div>
                            </div>
                        </div>
                    )}

            </div>

            {/* Operational Directives */}
            <div className="p-8 border-t border-outline bg-surface-dim grid grid-cols-2 gap-4">
                <button className="bg-primary hover:bg-primary-dark text-white py-4 px-6 rounded-2xl text-[10px] font-black uppercase tracking-widest transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-3 active:scale-95">
                    <span className="material-symbols-outlined text-base">history</span>
                    Audit Trail
                </button>
                <button className="bg-surface-container hover:bg-surface-container-highest border border-outline text-white py-4 px-6 rounded-2xl text-[10px] font-black uppercase tracking-widest transition-all flex items-center justify-center gap-3 active:scale-95">
                    <span className="material-symbols-outlined text-base">call</span>
                    Direct Link
                </button>
            </div>
        </div>
    );
};
