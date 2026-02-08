import { ArrowLeftIcon, BuildingOffice2Icon, CheckCircleIcon, GlobeAltIcon, PaintBrushIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../services/apiClient';

export const AddOrganization: React.FC = () => {
    const navigate = useNavigate();
    const [step, setStep] = useState(1);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        domain: '',
        default_theme: 'dark',
        contact_email: '',
        subscription_tier: 'enterprise'
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            await apiClient.post('/admin/organizations', formData);
            setStep(3); // Success step
        } catch (error) {
            alert('Failed to create organization');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="p-8 max-w-4xl mx-auto">
            <button 
                onClick={() => navigate('/admin/organizations')}
                className="flex items-center space-x-2 text-slate-500 hover:text-blue-600 transition-colors mb-8 group"
            >
                <ArrowLeftIcon className="h-4 w-4 group-hover:-translate-x-1 transition-transform" />
                <span>Back to Organizations</span>
            </button>

            <div className="bg-white/70 backdrop-blur-xl border border-white/40 shadow-2xl rounded-3xl overflow-hidden">
                {/* Stepper Header */}
                <div className="bg-slate-900/5 px-8 py-6 border-b border-slate-100 flex justify-between items-center">
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">Onboard New Organization</h1>
                        <p className="text-slate-500 text-sm mt-1">Configure workspace and branding for the new tenant</p>
                    </div>
                    <div className="flex items-center space-x-2">
                        {[1, 2, 3].map((s) => (
                            <div 
                                key={s}
                                className={`h-2 w-8 rounded-full transition-all duration-300 ${
                                    step === s ? 'bg-blue-600 w-12' : step > s ? 'bg-green-500' : 'bg-slate-200'
                                }`}
                            />
                        ))}
                    </div>
                </div>

                <div className="p-12">
                    {step === 1 && (
                        <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
                            <div className="grid grid-cols-2 gap-8">
                                <div className="space-y-4">
                                    <label className="text-sm font-semibold text-slate-700 flex items-center">
                                        <BuildingOffice2Icon className="h-4 w-4 mr-2 text-blue-500" />
                                        Organization Name
                                    </label>
                                    <input 
                                        type="text"
                                        placeholder="e.g. Acme Corp"
                                        className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                                        value={formData.name}
                                        onChange={(e) => setFormData({...formData, name: e.target.value})}
                                    />
                                </div>
                                <div className="space-y-4">
                                    <label className="text-sm font-semibold text-slate-700 flex items-center">
                                        <GlobeAltIcon className="h-4 w-4 mr-2 text-blue-500" />
                                        Custom Domain
                                    </label>
                                    <input 
                                        type="text"
                                        placeholder="acme.flyfleet.io"
                                        className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                                        value={formData.domain}
                                        onChange={(e) => setFormData({...formData, domain: e.target.value})}
                                    />
                                </div>
                            </div>
                            <div className="flex justify-end pt-8">
                                <button 
                                    onClick={() => setStep(2)}
                                    disabled={!formData.name}
                                    className="bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white px-8 py-3 rounded-xl font-semibold shadow-lg shadow-blue-900/20 transition-all"
                                >
                                    Continue to Settings
                                </button>
                            </div>
                        </div>
                    )}

                    {step === 2 && (
                        <form onSubmit={handleSubmit} className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500">
                             <div className="space-y-6">
                                <label className="text-sm font-semibold text-slate-700 flex items-center">
                                    <PaintBrushIcon className="h-4 w-4 mr-2 text-blue-500" />
                                    Default Workspace Theme
                                </label>
                                <div className="grid grid-cols-3 gap-4">
                                    {['light', 'dark', 'system'].map((t) => (
                                        <button
                                            key={t}
                                            type="button"
                                            onClick={() => setFormData({...formData, default_theme: t})}
                                            className={`p-4 rounded-xl border-2 transition-all ${
                                                formData.default_theme === t 
                                                ? 'border-blue-500 bg-blue-50 text-blue-700' 
                                                : 'border-slate-100 bg-white text-slate-500 hover:border-slate-200'
                                            }`}
                                        >
                                            <span className="capitalize font-medium">{t}</span>
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div className="flex justify-between pt-8 border-t border-slate-100">
                                <button 
                                    type="button"
                                    onClick={() => setStep(1)}
                                    className="text-slate-500 hover:text-slate-800 font-medium px-6 py-3"
                                >
                                    Back
                                </button>
                                <button 
                                    type="submit"
                                    disabled={loading}
                                    className="bg-slate-900 hover:bg-black text-white px-10 py-3 rounded-xl font-semibold shadow-xl transition-all flex items-center"
                                >
                                    {loading ? 'Creating...' : 'Initialize Organization'}
                                </button>
                            </div>
                        </form>
                    )}

                    {step === 3 && (
                        <div className="py-12 text-center space-y-6 animate-in zoom-in duration-500">
                            <div className="inline-flex items-center justify-center p-4 bg-green-100 rounded-full mb-4">
                                <CheckCircleIcon className="h-12 w-12 text-green-600" />
                            </div>
                            <h2 className="text-3xl font-bold text-slate-900">Organization Created!</h2>
                            <p className="text-slate-500 max-w-md mx-auto">
                                The workspace is now ready. You can now start adding users and configuring fleet assets.
                            </p>
                            <div className="pt-8">
                                <button 
                                    onClick={() => navigate('/admin/organizations')}
                                    className="bg-blue-600 hover:bg-blue-700 text-white px-8 py-3 rounded-xl font-semibold shadow-lg shadow-blue-900/20 transition-all"
                                >
                                    Back to Dashboard
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
