import { ArrowLeftIcon, CubeIcon, PlusIcon, TruckIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../services/apiClient';

export const AddVehicle: React.FC = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        license_plate: '',
        make: '',
        model: '',
        year: new Date().getFullYear(),
        fuel_type: 'petrol'
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            await apiClient.post('/admin/vehicles', formData);
            navigate('/admin/vehicles');
        } catch (error) {
            console.error('Failed to create vehicle:', error);
            alert('Failed to create vehicle. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="p-8 max-w-2xl mx-auto">
            <button 
                onClick={() => navigate('/admin/vehicles')}
                className="flex items-center space-x-2 text-slate-500 hover:text-blue-600 transition-colors mb-8 group"
            >
                <ArrowLeftIcon className="h-4 w-4 group-hover:-translate-x-1 transition-transform" />
                <span>Back to Fleet</span>
            </button>

            <div className="bg-white rounded-3xl shadow-xl border border-slate-100 overflow-hidden">
                <div className="px-10 py-8 border-b border-slate-100 bg-slate-50/50 flex items-center space-x-4">
                    <div className="p-3 bg-blue-100 rounded-xl text-blue-600">
                        <TruckIcon className="h-6 w-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-900">Add New Vehicle</h1>
                        <p className="text-slate-500 text-sm mt-1">Register a new transport asset to your fleet.</p>
                    </div>
                </div>

                <form onSubmit={handleSubmit} className="p-10 space-y-6">
                    <div className="grid grid-cols-2 gap-6">
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700">Vehicle Name</label>
                            <input 
                                type="text"
                                className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                                placeholder="e.g. Truck 101"
                                value={formData.name}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700">License Plate</label>
                            <input 
                                type="text"
                                className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                                placeholder="e.g. KA-01-AB-1234"
                                value={formData.license_plate}
                                onChange={(e) => setFormData({...formData, license_plate: e.target.value})}
                                required
                            />
                        </div>
                    </div>

                    <div className="pt-6 border-t border-slate-100 flex justify-end">
                        <button 
                            type="submit"
                            disabled={loading}
                            className="bg-slate-900 hover:bg-black text-white px-8 py-3 rounded-xl font-semibold shadow-lg transition-all disabled:opacity-50 flex items-center space-x-2"
                        >
                            {loading ? <CubeIcon className="h-5 w-5 animate-spin" /> : <PlusIcon className="h-5 w-5" />}
                            <span>{loading ? 'Registering...' : 'Register Vehicle'}</span>
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};


