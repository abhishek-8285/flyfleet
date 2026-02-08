import { PlusIcon, SignalIcon, TrashIcon } from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import apiClient from '../../services/apiClient';

interface Device {
    id: string;
    serial_number: string;
    type: string;
    status: string;
    vehicle_id?: string;
    last_ping_at?: string;
    created_at: string;
}

export const Devices: React.FC = () => {
    const [devices, setDevices] = useState<Device[]>([]);
    const [loading, setLoading] = useState(true);
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);
    const [newDevice, setNewDevice] = useState({ serial: '', type: 'OBD-II' });

    const fetchDevices = async () => {
        try {
            const res = await apiClient.get('/admin/devices?page=1&size=100') as any;
            setDevices(res.data || []);
        } catch (error) {
            console.error('Failed to fetch devices:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchDevices();
    }, []);

    const handleAdd = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await apiClient.post('/admin/devices', {
                serial_number: newDevice.serial,
                type: newDevice.type
            });
            setIsAddModalOpen(false);
            setNewDevice({ serial: '', type: 'OBD-II' });
            fetchDevices();
        } catch (error) {
            alert('Failed to register device');
        }
    };

    const handleDelete = async (id: string) => {
        if (!window.confirm('Are you sure you want to remove this device?')) return;
        try {
            await apiClient.delete(`/admin/devices?id=${id}`);
            fetchDevices();
        } catch (error) {
            alert('Failed to delete device');
        }
    };

    return (
        <div className="p-8 max-w-7xl mx-auto space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900">Hardware Devices</h1>
                    <p className="text-slate-500 mt-2">Manage telemetry units and sensors.</p>
                </div>
                <button 
                    onClick={() => setIsAddModalOpen(true)}
                    className="flex items-center space-x-2 bg-blue-600 text-white px-5 py-2.5 rounded-xl hover:bg-blue-700 transition-all shadow-lg shadow-blue-900/20"
                >
                    <PlusIcon className="h-5 w-5" />
                    <span>Register Device</span>
                </button>
            </div>

            {loading ? (
                <div className="text-center py-20 text-slate-400">Loading hardware inventory...</div>
            ) : (
                <div className="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
                    <table className="w-full text-left">
                        <thead className="bg-slate-50 border-b border-slate-100 text-slate-500 font-semibold text-sm uppercase">
                            <tr>
                                <th className="px-6 py-4">Serial Number</th>
                                <th className="px-6 py-4">Type</th>
                                <th className="px-6 py-4">Status</th>
                                <th className="px-6 py-4">Linked Vehicle</th>
                                <th className="px-6 py-4">Last Ping</th>
                                <th className="px-6 py-4 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-50">
                            {devices.map((device) => (
                                <tr key={device.id} className="hover:bg-slate-50/50 transition-colors">
                                    <td className="px-6 py-4 font-mono font-medium text-slate-700">{device.serial_number}</td>
                                    <td className="px-6 py-4 text-slate-600">{device.type}</td>
                                    <td className="px-6 py-4">
                                        <span className={`px-2 py-1 rounded-full text-xs font-bold uppercase ${
                                            device.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-500'
                                        }`}>
                                            {device.status}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4">
                                        {device.vehicle_id ? (
                                            <span className="flex items-center text-blue-600 bg-blue-50 px-3 py-1 rounded-lg w-fit text-sm">
                                                <SignalIcon className="h-4 w-4 mr-1" />
                                                Assigned
                                            </span>
                                        ) : (
                                            <span className="text-slate-400 text-sm">Unassigned</span>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 text-slate-400 text-sm">
                                        {device.last_ping_at ? new Date(device.last_ping_at).toLocaleString() : 'Never'}
                                    </td>
                                    <td className="px-6 py-4 text-right">
                                        <button 
                                            onClick={() => handleDelete(device.id)}
                                            className="text-slate-400 hover:text-red-500 p-2 hover:bg-red-50 rounded-lg transition-colors"
                                        >
                                            <TrashIcon className="h-5 w-5" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                            {devices.length === 0 && (
                                <tr>
                                    <td colSpan={6} className="text-center py-12 text-slate-400">
                                        No devices found. Register your first hardware unit above.
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Add Device Modal */}
            {isAddModalOpen && (
                <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-50">
                    <div className="bg-white rounded-3xl p-8 w-full max-w-md shadow-2xl animate-fade-in-up">
                        <h2 className="text-2xl font-bold text-slate-900 mb-6">Register Device</h2>
                        <form onSubmit={handleAdd} className="space-y-4">
                            <div>
                                <label className="block text-sm font-semibold text-slate-700 mb-1">Serial Number</label>
                                <input 
                                    type="text" 
                                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 outline-none focus:ring-2 focus:ring-blue-500/20"
                                    placeholder="Enter S/N"
                                    value={newDevice.serial}
                                    onChange={(e) => setNewDevice({...newDevice, serial: e.target.value})}
                                    required
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-semibold text-slate-700 mb-1">Device Type</label>
                                <select 
                                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 outline-none focus:ring-2 focus:ring-blue-500/20"
                                    value={newDevice.type}
                                    onChange={(e) => setNewDevice({...newDevice, type: e.target.value})}
                                >
                                    <option value="OBD-II">OBD-II Dongle</option>
                                    <option value="GPS-Tracker">GPS Tracker</option>
                                    <option value="Dashcam">Dashcam</option>
                                </select>
                            </div>
                            <div className="flex space-x-3 mt-8">
                                <button 
                                    type="button"
                                    onClick={() => setIsAddModalOpen(false)}
                                    className="flex-1 py-3 text-slate-600 font-semibold hover:bg-slate-50 rounded-xl transition-colors"
                                >
                                    Cancel
                                </button>
                                <button 
                                    type="submit"
                                    className="flex-1 bg-slate-900 text-white py-3 rounded-xl font-semibold hover:bg-black transition-colors"
                                >
                                    Register
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};
