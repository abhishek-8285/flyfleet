import { PlusIcon, SignalIcon, TruckIcon, UserIcon } from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../services/apiClient';

interface Vehicle {
    id: string;
    name: string;
    license_plate: string;
    status: string;
    make: string;
    model: string;
    org_id: string;
    created_at: string;
}

export const Vehicles: React.FC = () => {
    const navigate = useNavigate();
    const [vehicles, setVehicles] = useState<Vehicle[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
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
        fetchVehicles();
    }, []);

    return (
        <div className="p-8 max-w-7xl mx-auto space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900">Fleet Vehicles</h1>
                    <p className="text-slate-500 mt-2">Manage your transport assets and assignments.</p>
                </div>
                <button 
                    onClick={() => navigate('/admin/vehicles/new')}
                    className="flex items-center space-x-2 bg-blue-600 text-white px-5 py-2.5 rounded-xl hover:bg-blue-700 transition-all shadow-lg shadow-blue-900/20"
                >
                    <PlusIcon className="h-5 w-5" />
                    <span>Add Vehicle</span>
                </button>
            </div>

            {loading ? (
                <div className="text-center py-20 text-slate-400">Loading fleet data...</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {vehicles.map((vehicle) => (
                        <div 
                            key={vehicle.id}
                            onClick={() => navigate(`/admin/vehicles/${vehicle.id}`)}
                            className="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm hover:shadow-md hover:border-blue-200 transition-all cursor-pointer group"
                        >
                            <div className="flex justify-between items-start mb-4">
                                <div className="p-3 bg-blue-50 rounded-xl text-blue-600 group-hover:bg-blue-600 group-hover:text-white transition-colors">
                                    <TruckIcon className="h-6 w-6" />
                                </div>
                                <span className={`px-2 py-1 rounded-full text-xs font-bold uppercase ${
                                    vehicle.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-500'
                                }`}>
                                    {vehicle.status}
                                </span>
                            </div>
                            
                            <h3 className="text-lg font-bold text-slate-900">{vehicle.name || 'Unnamed Vehicle'}</h3>
                            <p className="text-slate-500 text-sm font-mono mt-1">{vehicle.license_plate}</p>
                            
                            <div className="mt-6 pt-4 border-t border-slate-50 flex justify-between text-sm text-slate-400">
                                <span className="flex items-center">
                                    <UserIcon className="h-4 w-4 mr-1" /> No Driver
                                </span>
                                <span className="flex items-center">
                                    <SignalIcon className="h-4 w-4 mr-1" /> No Device
                                </span>
                            </div>
                        </div>
                    ))}
                    
                    {/* Empty State Action */}
                    {vehicles.length === 0 && (
                        <div 
                            onClick={() => navigate('/admin/vehicles/new')}
                            className="border-2 border-dashed border-slate-200 rounded-2xl p-6 flex flex-col items-center justify-center text-slate-400 hover:border-blue-300 hover:text-blue-500 cursor-pointer transition-colors min-h-[200px]"
                        >
                            <PlusIcon className="h-8 w-8 mb-2" />
                            <span className="font-medium">Register First Vehicle</span>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};
