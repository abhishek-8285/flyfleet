import { ArrowLeftIcon, Cog6ToothIcon, PencilIcon, ShieldCheckIcon, SignalIcon, TruckIcon, UsersIcon } from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import apiClient from '../../services/apiClient';

interface Organization {
    id: string;
    name: string;
    domain: string;
    status: string;
    default_theme: string;
    created_at: string;
}

export const OrganizationDetail: React.FC = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const [org, setOrg] = useState<Organization | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchOrg = async () => {
            try {
                const res = await apiClient.get(`/admin/organizations?id=${id}`) as any;
                setOrg(res.data);
            } catch (error) {
                console.error('Failed to fetch org:', error);
            } finally {
                setLoading(false);
            }
        };
        fetchOrg();
    }, [id]);

    if (loading) return <div className="p-12 text-center text-slate-500">Loading workspace intelligence...</div>;
    if (!org) return <div className="p-12 text-center text-red-500">Organization profile not found.</div>;

    return (
        <div className="p-8 max-w-7xl mx-auto space-y-8">
            {/* Action Bar */}
            <div className="flex justify-between items-center bg-white/50 backdrop-blur-md p-4 rounded-2xl border border-white/20 shadow-sm sticky top-4 z-10">
                <button 
                    onClick={() => navigate('/admin/organizations')}
                    className="flex items-center space-x-2 text-slate-500 hover:text-blue-600 transition-colors group"
                >
                    <ArrowLeftIcon className="h-4 w-4 group-hover:-translate-x-1 transition-transform" />
                    <span className="font-medium">Directory</span>
                </button>
                <div className="flex items-center space-x-3">
                    <button 
                        onClick={() => navigate(`/admin/organizations/${id}/settings`)}
                        className="p-2 text-slate-400 hover:text-slate-800 transition-colors"
                        title="Configuration"
                    >
                        <Cog6ToothIcon className="h-6 w-6" />
                    </button>
                    <button 
                        onClick={() => navigate(`/admin/organizations/${id}/edit`)}
                        className="flex items-center space-x-2 bg-slate-900 text-white px-5 py-2 rounded-xl hover:bg-black transition-all shadow-lg"
                    >
                        <PencilIcon className="h-4 w-4" />
                        <span>Edit Profile</span>
                    </button>
                </div>
            </div>

            {/* Banner Section */}
            <div className="grid grid-cols-3 gap-8">
                <div className="col-span-2 space-y-8">
                    <div className="bg-gradient-to-br from-slate-900 to-slate-800 rounded-3xl p-10 text-white relative overflow-hidden shadow-2xl">
                        <div className="relative z-10">
                            <span className="px-3 py-1 bg-blue-500/20 text-blue-300 rounded-full text-xs font-bold uppercase tracking-widest border border-blue-500/30">
                                {org.status}
                            </span>
                            <h1 className="text-5xl font-black mt-4">{org.name}</h1>
                            <p className="text-slate-400 mt-2 flex items-center">
                                <ShieldCheckIcon className="h-4 w-4 mr-2" />
                                Custom Domain: <span className="text-blue-400 ml-1 underline">{org.domain || 'Not configured'}</span>
                            </p>
                        </div>
                        {/* Abstract Background Element */}
                        <div className="absolute top-0 right-0 w-64 h-64 bg-blue-500/10 rounded-full blur-3xl -mr-20 -mt-20" />
                    </div>

                    {/* Quick Stats Grid */}
                    <div className="grid grid-cols-3 gap-6">
                        {[
                            { label: 'Active Users', value: '42', icon: UsersIcon, color: 'blue' },
                            { label: 'Fleet Assets', value: '128', icon: TruckIcon, color: 'indigo' },
                            { label: 'Live Devices', value: '115', icon: SignalIcon, color: 'emerald' },
                        ].map((stat, i) => (
                            <div key={i} className="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm">
                                <stat.icon className={`h-8 w-8 text-${stat.color}-500 mb-3`} />
                                <div className="text-2xl font-bold text-slate-800">{stat.value}</div>
                                <div className="text-sm text-slate-500">{stat.label}</div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Side Panels */}
                <div className="space-y-6">
                    <div className="bg-white p-6 rounded-3xl border border-slate-100 shadow-sm">
                        <h3 className="font-bold text-slate-800 mb-4 px-1">Infrastructure Health</h3>
                        <div className="space-y-4">
                            {['API Gateway', 'Real-time Tracking', 'Auth Sync'].map((sys, i) => (
                                <div key={i} className="flex justify-between items-center text-sm">
                                    <span className="text-slate-500">{sys}</span>
                                    <span className="text-green-600 font-semibold">Operational</span>
                                </div>
                            ))}
                        </div>
                    </div>
                    
                    <div className="bg-blue-600 p-6 rounded-3xl text-white shadow-xl shadow-blue-900/10 active:scale-95 transition-transform cursor-pointer">
                        <h3 className="font-bold mb-1">Subscription: Enterprise</h3>
                        <p className="text-xs text-blue-100 opacity-80">Next billing on March 12, 2026</p>
                        <div className="mt-4 pt-4 border-t border-white/10 flex justify-between items-center">
                            <span className="text-sm font-semibold">View Invoice</span>
                            <ArrowLeftIcon className="h-4 w-4 rotate-180" />
                        </div>
                    </div>
                </div>
            </div>

            {/* Detailed Listing (Placeholder for Users/Vehicles) */}
            <div className="bg-white rounded-3xl border border-slate-100 shadow-sm overflow-hidden">
                <div className="border-b border-slate-100 px-8 py-4 flex items-center space-x-8">
                    {['Users', 'Devices', 'Audit Trail'].map((tab, i) => (
                        <button 
                            key={i}
                            className={`px-2 py-3 text-sm font-bold transition-colors border-b-2 ${
                                i === 0 ? 'border-blue-600 text-slate-900' : 'border-transparent text-slate-400 hover:text-slate-600'
                            }`}
                        >
                            {tab}
                        </button>
                    ))}
                </div>
                <div className="p-12 text-center text-slate-400 italic">
                    Select a tab to view detailed records for this organization.
                </div>
            </div>
        </div>
    );
};
