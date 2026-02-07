import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import { AppDispatch, RootState } from '../store';
import { fetchVehicles } from '../store/vehiclesSlice';
import { Vehicle } from '../types';

const Vehicles: React.FC = () => {
    const dispatch = useDispatch<AppDispatch>();
    const { vehicles, loading } = useSelector((state: RootState) => state.vehicles);

    useEffect(() => {
        dispatch(fetchVehicles({}) as any);
    }, [dispatch]);

    return (
        <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col shadow-lg">
                <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
                    <h3 className="text-white font-semibold">Fleet Management</h3>
                    <div className="flex gap-2">
                        <button className="px-3 py-1.5 bg-surface-dark border border-border-color hover:bg-border-color text-text-muted hover:text-white text-xs font-medium rounded transition-colors">
                            Download Report
                        </button>
                        <button className="px-3 py-1.5 bg-primary text-white text-xs font-medium rounded hover:bg-blue-600 transition-colors">
                            Add Vehicle
                        </button>
                    </div>
                </div>
                <div className="overflow-x-auto">
                    <table className="w-full text-left text-sm text-text-muted">
                        <thead className="bg-[#1d2936] text-xs uppercase font-medium">
                            <tr>
                                <th className="px-6 py-4">Vehicle ID</th>
                                <th className="px-6 py-4">Driver</th>
                                <th className="px-6 py-4">Status</th>
                                <th className="px-6 py-4 text-right">Action</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border-color">
                            {loading ? (
                                <tr>
                                    <td colSpan={4} className="px-6 py-8 text-center text-text-muted">Loading fleet data...</td>
                                </tr>
                            ) : vehicles.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="px-6 py-8 text-center text-text-muted">No vehicles found.</td>
                                </tr>
                            ) : (
                                vehicles.map((vehicle: Vehicle) => (
                                    <tr key={vehicle.id} className="hover:bg-[#1d2936] transition-colors">
                                        <td className="px-6 py-4 font-medium text-white">
                                            <div className="flex items-center gap-3">
                                                <div className="w-8 h-8 rounded bg-border-color flex items-center justify-center text-text-muted">
                                                    <span className="material-symbols-outlined text-[18px]">local_shipping</span>
                                                </div>
                                                <div>
                                                    <div className="text-white font-medium">#{vehicle.id.slice(-4)}</div>
                                                    <div className="text-xs text-text-muted">{vehicle.registration_number}</div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            {vehicle.current_driver ? (
                                                <div className="flex items-center gap-2">
                                                    <div className="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center text-[10px] text-primary font-bold">
                                                        {vehicle.current_driver.name.charAt(0)}
                                                    </div>
                                                    <span>{vehicle.current_driver.name}</span>
                                                </div>
                                            ) : (
                                                <span className="text-text-muted/50 italic text-xs">Unassigned</span>
                                            )}
                                        </td>
                                        <td className="px-6 py-4">
                                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${
                                                vehicle.status === 'active' 
                                                ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' 
                                                : 'bg-orange-500/10 text-orange-500 border-orange-500/20'
                                            }`}>
                                                {vehicle.status.charAt(0).toUpperCase() + vehicle.status.slice(1)}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 text-right">
                                            <Link to={`/dashboard/vehicles/${vehicle.id}`} className="text-primary hover:text-white text-xs font-medium">
                                                Detail
                                            </Link>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
            <footer className="mt-8 text-center text-xs text-[#64748b] pb-6 flex flex-col items-center gap-1">
                <span>© 2026 Enterprise Fleet Manager. All rights reserved.</span>
                <span className="text-[10px]">v2.4.1 | Server Time: <span className="font-mono">14:02 UTC</span></span>
            </footer>
        </div>
    );
};

export default Vehicles;
