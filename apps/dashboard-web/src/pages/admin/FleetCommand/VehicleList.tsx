import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Sidebar } from '../../../components/layout/Sidebar'; // Assuming standard layout components exist
import apiClient from '../../../services/apiClient';
import { Vehicle } from './types';
// If not, we might need to adapt. relying on previous file analysis which showed Sidebar usage.

export const VehicleList: React.FC = () => {
    const navigate = useNavigate();
    const [vehicles, setVehicles] = useState<Vehicle[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [filterStatus, setFilterStatus] = useState<string>('all');

    useEffect(() => {
        fetchVehicles();
    }, []);

    const fetchVehicles = async () => {
        try {
            const res = await apiClient.get('/admin/vehicles?page=1&size=100') as any;
            setVehicles(res.data || []);
        } catch (error) {
            console.error('Failed to fetch vehicles:', error);
        } finally {
            setLoading(false);
        }
    };

    const filteredVehicles = vehicles.filter(v => {
        const matchesSearch = 
            v.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            v.license_plate?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            v.vin?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            v.vehicle_no?.toLowerCase().includes(searchTerm.toLowerCase());
        
        const matchesStatus = filterStatus === 'all' || v.status === filterStatus;
        
        return matchesSearch && matchesStatus;
    });

    return (
        <div className="flex h-screen w-full bg-[#0f111a]"> {/* Premium Dark Background */}
            <Sidebar />
            
            <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
                {/* Header */}
                <header className="h-20 flex items-center justify-between px-8 border-b border-white/5 bg-[#161b22]">
                    <div>
                        <h1 className="text-2xl font-bold text-white tracking-tight">Fleet Command</h1>
                        <p className="text-slate-400 text-sm mt-1">Real-time asset visibility and control.</p>
                    </div>
                    <div className="flex items-center gap-4">
                        <button 
                            onClick={() => navigate('/admin/vehicles/new')}
                            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-5 py-2.5 rounded-lg font-medium transition-all shadow-lg shadow-blue-500/20"
                        >
                            <span className="material-symbols-outlined text-[20px]">add</span>
                            Add Vehicle
                        </button>
                    </div>
                </header>

                {/* Filters & Search */}
                <div className="p-8 pb-4">
                    <div className="flex justify-between items-center bg-[#1c222b] p-4 rounded-xl border border-white/5">
                        <div className="flex items-center gap-4 flex-1">
                            <div className="relative flex-1 max-w-md">
                                <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">search</span>
                                <input 
                                    type="text" 
                                    placeholder="Search VIN, Plate, Name..." 
                                    className="w-full bg-[#161b22] border border-white/10 text-white pl-10 pr-4 py-2.5 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all placeholder:text-slate-600"
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                />
                            </div>
                            <div className="h-8 w-px bg-white/10 mx-2"></div>
                            <select 
                                value={filterStatus}
                                onChange={(e) => setFilterStatus(e.target.value)}
                                className="bg-[#161b22] border border-white/10 text-slate-300 px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                            >
                                <option value="all">All Status</option>
                                <option value="active">Active</option>
                                <option value="maintenance">Maintenance</option>
                                <option value="inactive">Inactive</option>
                            </select>
                        </div>
                        <div className="flex items-center gap-3 text-sm text-slate-400">
                            <span className="material-symbols-outlined">dns</span>
                            <span>{filteredVehicles.length} Assets</span>
                        </div>
                    </div>
                </div>

                {/* Table */}
                <div className="flex-1 overflow-auto px-8 pb-8">
                    {loading ? (
                        <div className="flex items-center justify-center h-64 text-slate-500">
                            <span className="animate-spin material-symbols-outlined mr-2">progress_activity</span>
                            Loading Fleet Data...
                        </div>
                    ) : (
                        <div className="bg-[#1c222b] rounded-xl border border-white/5 overflow-hidden">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr className="bg-white/5 border-b border-white/5 text-slate-400 text-sm uppercase tracking-wider">
                                        <th className="p-4 font-medium pl-6">Vehicle</th>
                                        <th className="p-4 font-medium">VIN / SN</th>
                                        <th className="p-4 font-medium">Odometer</th>
                                        <th className="p-4 font-medium">Service Due</th>
                                        <th className="p-4 font-medium">Status</th>
                                        <th className="p-4 font-medium text-right pr-6">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="text-sm divide-y divide-white/5">
                                    {filteredVehicles.map((vehicle) => (
                                        <tr 
                                            key={vehicle.id} 
                                            className="hover:bg-white/[0.02] transition-colors group cursor-pointer"
                                            onClick={() => navigate(`/admin/vehicles/${vehicle.id}`)}
                                        >
                                            <td className="p-4 pl-6">
                                                <div className="flex items-center gap-3">
                                                    <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-400">
                                                        <span className="material-symbols-outlined">directions_car</span>
                                                    </div>
                                                    <div>
                                                        <div className="font-medium text-white">{vehicle.name || 'Unnamed'}</div>
                                                        <div className="text-xs text-slate-500">{vehicle.license_plate}</div>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="p-4 text-slate-300 font-mono text-xs">
                                                <div>{vehicle.vin || '-'}</div>
                                                <div className="text-slate-500">{vehicle.vehicle_no && `#${vehicle.vehicle_no}`}</div>
                                            </td>
                                            <td className="p-4 text-slate-300">
                                                {vehicle.odometer ? `${vehicle.odometer.toLocaleString()} km` : '-'}
                                            </td>
                                            <td className="p-4">
                                                {vehicle.service_due ? (
                                                    <span className="text-amber-400 flex items-center gap-1">
                                                        <span className="material-symbols-outlined text-sm">build</span>
                                                        {new Date(vehicle.service_due).toLocaleDateString()}
                                                    </span>
                                                ) : (
                                                    <span className="text-slate-600">-</span>
                                                )}
                                            </td>
                                            <td className="p-4">
                                                <span className={`px-2.5 py-1 rounded-full text-xs font-medium border ${
                                                    vehicle.status === 'active' 
                                                        ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' 
                                                        : vehicle.status === 'maintenance'
                                                        ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                                                        : 'bg-slate-500/10 text-slate-400 border-slate-500/20'
                                                }`}>
                                                    {vehicle.status?.toUpperCase() || 'UNKNOWN'}
                                                </span>
                                            </td>
                                            <td className="p-4 text-right pr-6">
                                                <button className="text-slate-400 hover:text-white p-2 hover:bg-white/10 rounded-full transition-colors"
                                                    onClick={(e) => { e.stopPropagation(); navigate(`/admin/vehicles/${vehicle.id}/edit`); }}
                                                >
                                                    <span className="material-symbols-outlined text-[20px]">edit</span>
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    {filteredVehicles.length === 0 && (
                                        <tr>
                                            <td colSpan={6} className="p-12 text-center text-slate-500">
                                                <span className="material-symbols-outlined text-4xl mb-2 opacity-20">local_shipping</span>
                                                <p>No vehicles found matching your criteria.</p>
                                            </td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
};
