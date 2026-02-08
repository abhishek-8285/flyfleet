import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import FuelService from '../services/FuelService';
import { FuelSettings } from '../types/fuel';

const FuelSettingsView: React.FC = () => {
    const navigate = useNavigate();
    const [settings, setSettings] = useState<FuelSettings | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        const fetchSettings = async () => {
            try {
                const data = await FuelService.getSettings();
                setSettings(data);
            } catch (error) {
                console.error('Failed to fetch fuel settings:', error);
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, []);

    const handleUpdate = async (update: Partial<FuelSettings>) => {
        if (!settings) return;
        setSaving(true);
        try {
            const updated = await FuelService.updateSettings(update);
            setSettings(updated);
        } catch (error) {
            console.error('Failed to update settings:', error);
        } finally {
            setSaving(false);
        }
    };

    if (loading) return null;

    return (
        <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
            <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex items-center justify-between flex-shrink-0 z-10 font-poppins">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/fuel')} className="p-1 hover:bg-surface-dark rounded text-text-muted transition-colors">
                        <span className="material-symbols-outlined">arrow_back</span>
                    </button>
                    <h2 className="text-xl font-semibold text-white tracking-tight">Fuel Management Settings</h2>
                </div>
                {saving && <div className="text-xs text-primary animate-pulse flex items-center gap-2">
                    <div className="w-3 h-3 border border-primary border-t-transparent rounded-full animate-spin"></div>
                    Saving changes...
                </div>}
            </div>

            <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
                <div className="max-w-3xl mx-auto space-y-8 animate-fade-in">
                    {/* General Settings */}
                    <section className="bg-surface-dark/30 backdrop-blur-md border border-border-color rounded-2xl overflow-hidden shadow-lg p-6 space-y-6">
                        <div className="flex items-center gap-3">
                            <span className="material-symbols-outlined text-primary">currency_exchange</span>
                            <h3 className="text-white font-medium">Regional & Currency</h3>
                        </div>
                        
                        <div className="flex flex-col gap-2">
                            <label className="text-text-muted text-xs font-medium uppercase tracking-wider">Default Currency</label>
                            <select 
                                value={settings?.default_currency}
                                onChange={(e) => handleUpdate({ default_currency: e.target.value })}
                                className="bg-background-dark border border-border-color text-white text-sm rounded-xl px-4 py-2.5 outline-none focus:ring-1 focus:ring-primary transition-all w-48"
                            >
                                <option value="USD">USD - US Dollar</option>
                                <option value="EUR">EUR - Euro</option>
                                <option value="GBP">GBP - British Pound</option>
                                <option value="INR">INR - Indian Rupee</option>
                            </select>
                        </div>
                    </section>

                    {/* Validation Rules */}
                    <section className="bg-surface-dark/30 backdrop-blur-md border border-border-color rounded-2xl overflow-hidden shadow-lg p-6 space-y-6">
                        <div className="flex items-center gap-3">
                            <span className="material-symbols-outlined text-accent-yellow">verified_user</span>
                            <h3 className="text-white font-medium">Entry Validation Rules</h3>
                        </div>
                        
                        <div className="space-y-4">
                            <div className="flex items-center justify-between p-4 bg-background-dark border border-border-color/50 rounded-xl group hover:border-primary/30 transition-colors">
                                <div>
                                    <div className="text-white text-sm font-medium">Previous Odometer Check</div>
                                    <div className="text-text-muted text-xs mt-0.5">Prevent fuel entries with odometer readings lower than preceding logs.</div>
                                </div>
                                <label className="relative inline-flex items-center cursor-pointer">
                                    <input 
                                        type="checkbox" 
                                        className="sr-only peer"
                                        checked={settings?.validation_rules.odometer_check}
                                        onChange={(e) => handleUpdate({ validation_rules: { ...settings!.validation_rules, odometer_check: e.target.checked }})}
                                    />
                                    <div className="w-11 h-6 bg-surface-dark peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                                </label>
                            </div>

                            <div className="flex items-center justify-between p-4 bg-background-dark border border-border-color/50 rounded-xl group hover:border-primary/30 transition-colors">
                                <div>
                                    <div className="text-white text-sm font-medium">Cost Verification</div>
                                    <div className="text-text-muted text-xs mt-0.5">Automatically verify that Total Cost = Quantity × Unit Price.</div>
                                </div>
                                <label className="relative inline-flex items-center cursor-pointer">
                                    <input 
                                        type="checkbox" 
                                        className="sr-only peer"
                                        checked={settings?.validation_rules.cost_calculation_verify}
                                        onChange={(e) => handleUpdate({ validation_rules: { ...settings!.validation_rules, cost_calculation_verify: e.target.checked }})}
                                    />
                                    <div className="w-11 h-6 bg-surface-dark peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                                </label>
                            </div>
                        </div>
                    </section>

                    {/* Fuel Types */}
                    <section className="bg-surface-dark/30 backdrop-blur-md border border-border-color rounded-2xl overflow-hidden shadow-lg p-6 space-y-6">
                        <div className="flex items-center gap-3">
                            <span className="material-symbols-outlined text-orange-400">local_fire_department</span>
                            <h3 className="text-white font-medium">Allowed Fuel Categories</h3>
                        </div>
                        
                        <div className="flex flex-wrap gap-2">
                            {['Diesel', 'Petrol 95', 'Petrol 98', 'AdBlue', 'EV Charge', 'LPG'].map(type => (
                                <button 
                                    key={type}
                                    className={`px-4 py-2 rounded-full text-xs font-medium border transition-all ${
                                        (settings?.allowed_fuel_types || []).includes(type)
                                        ? 'bg-primary/20 border-primary text-primary'
                                        : 'bg-surface-dark border-border-color text-text-muted hover:border-text-muted/50'
                                    }`}
                                >
                                    {type}
                                </button>
                            ))}
                            <button className="w-8 h-8 rounded-full bg-surface-dark border border-dashed border-border-color flex items-center justify-center text-text-muted hover:text-white hover:border-white transition-all">
                                <span className="material-symbols-outlined text-[18px]">add</span>
                            </button>
                        </div>
                    </section>
                </div>
            </div>
        </div>
    );
};

export default FuelSettingsView;
