import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

const CreateReport: React.FC = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        report_type: 'SECURITY_AUDIT',
        format: 'PDF',
        frequency: 'WEEKLY',
        cron_expression: '0 8 * * 1', // Default
        delivery_time: '08:00',
        timezone: 'UTC',
        recipients: '',
        filters: JSON.stringify({ severity: ["high", "critical"] }, null, 2)
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            const recipientsList = formData.recipients.split(',').map(email => email.trim()).filter(Boolean);
            let filtersJson = {};
            try {
                filtersJson = JSON.parse(formData.filters);
            } catch (err) {
                alert("Invalid JSON format in filters");
                setLoading(false);
                return;
            }

            const payload = {
                ...formData,
                recipients: recipientsList,
                filters: filtersJson
            };

            await apiClient.post('/admin/reports/schedules', payload);
            navigate('/admin/reports');
        } catch (error: any) {
            console.error('Submission failed', error);
            alert(`Failed to submit: ${error?.message || 'Unknown error'}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex h-screen w-full bg-background-dark text-white">
            <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                    <div className="flex flex-col">
                        <nav className="flex items-center gap-2 text-[10px] uppercase tracking-wider font-bold text-text-muted mb-1">
                            <span className="hover:text-primary transition-colors cursor-pointer" onClick={() => navigate('/admin/reports')}>Scheduled Reports</span>
                            <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                            <span className="text-primary">New Schedule</span>
                        </nav>
                        <h2 className="text-white text-lg font-bold leading-tight">Create New Scheduled Report</h2>
                    </div>
                    <div className="flex items-center gap-3">
                        <button onClick={() => navigate('/admin/reports')} className="px-4 py-2 text-text-muted hover:text-white text-xs font-semibold rounded transition-colors">
                            Cancel
                        </button>
                        <button onClick={handleSubmit} disabled={loading} className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20 disabled:opacity-50">
                            <span className="material-symbols-outlined text-[16px]">save</span>
                            {loading ? 'Creating...' : 'Create Schedule'}
                        </button>
                    </div>
                </header>
                <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
                    <div className="max-w-4xl mx-auto space-y-8">
                        <section className="space-y-4">
                            <div className="flex items-center gap-2 pb-2 border-b border-border-color">
                                <span className="material-symbols-outlined text-primary text-[20px]">assignment</span>
                                <h3 className="text-white font-semibold">Report Details</h3>
                            </div>
                            <div className="grid grid-cols-2 gap-6">
                                <div className="col-span-2">
                                    <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">Report Name</label>
                                    <input name="name" onChange={handleChange} value={formData.name} className="w-full bg-surface-dark border border-border-color rounded-lg px-4 py-2.5 text-white text-sm focus:ring-1 focus:ring-primary focus:border-primary outline-none input-dark" placeholder="e.g., Weekly MFA Compliance Audit" type="text" required/>
                                </div>
                                <div>
                                    <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">Report Type</label>
                                    <select name="report_type" onChange={handleChange} value={formData.report_type} className="w-full bg-surface-dark border border-border-color rounded-lg px-4 py-2.5 text-white text-sm focus:ring-1 focus:ring-primary focus:border-primary outline-none">
                                        <option value="SECURITY_AUDIT">Security Audit</option>
                                        <option value="MFA_ADOPTION">MFA Adoption Metrics</option>
                                        <option value="COMPLIANCE_STATUS">Compliance Status</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">Export Format</label>
                                    <div className="flex gap-2">
                                        {['PDF', 'CSV', 'XLSX'].map(fmt => (
                                            <label key={fmt} className="flex-1">
                                                <input className="sr-only peer" name="format" type="radio" value={fmt} checked={formData.format === fmt} onChange={handleChange} />
                                                <div className="text-center py-2 text-xs font-medium rounded border border-border-color bg-surface-dark text-text-muted peer-checked:bg-primary/10 peer-checked:text-primary peer-checked:border-primary cursor-pointer transition-all">{fmt}</div>
                                            </label>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </section>
                        <section className="space-y-4">
                            <div className="flex items-center gap-2 pb-2 border-b border-border-color">
                                <span className="material-symbols-outlined text-primary text-[20px]">calendar_month</span>
                                <h3 className="text-white font-semibold">Schedule Configuration</h3>
                            </div>
                            <div className="grid grid-cols-3 gap-6">
                                <div>
                                    <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">Frequency</label>
                                    <select name="frequency" onChange={handleChange} value={formData.frequency} className="w-full bg-surface-dark border border-border-color rounded-lg px-4 py-2.5 text-white text-sm focus:ring-1 focus:ring-primary focus:border-primary outline-none">
                                        <option value="DAILY">Daily</option>
                                        <option value="WEEKLY">Weekly</option>
                                        <option value="MONTHLY">Monthly</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">Delivery Time</label>
                                    <input name="delivery_time" onChange={handleChange} value={formData.delivery_time} className="w-full bg-surface-dark border border-border-color rounded-lg px-4 py-2.5 text-white text-sm focus:ring-1 focus:ring-primary focus:border-primary outline-none" type="time"/>
                                </div>
                                <div>
                                    <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">Timezone</label>
                                    <select name="timezone" onChange={handleChange} value={formData.timezone} className="w-full bg-surface-dark border border-border-color rounded-lg px-4 py-2.5 text-white text-sm focus:ring-1 focus:ring-primary focus:border-primary outline-none">
                                        <option value="UTC">UTC</option>
                                        <option value="EST">Eastern Time</option>
                                        <option value="CET">Central European Time</option>
                                    </select>
                                </div>
                            </div>
                        </section>
                        <section className="space-y-4">
                            <div className="flex items-center gap-2 pb-2 border-b border-border-color">
                                <span className="material-symbols-outlined text-primary text-[20px]">mail</span>
                                <h3 className="text-white font-semibold">Recipients</h3>
                            </div>
                            <div>
                                <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">Email List</label>
                                <textarea name="recipients" onChange={handleChange} value={formData.recipients} className="w-full bg-surface-dark border border-border-color rounded-lg px-4 py-2.5 text-white text-sm focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder:text-slate-600" placeholder="Enter emails separated by commas..." rows={3}></textarea>
                            </div>
                        </section>
                        <section className="space-y-4">
                            <div className="flex items-center gap-2 pb-2 border-b border-border-color">
                                <span className="material-symbols-outlined text-primary text-[20px]">filter_alt</span>
                                <h3 className="text-white font-semibold">Data Filters (JSON)</h3>
                            </div>
                            <div className="space-y-4">
                                <div className="p-4 bg-surface-darker border border-border-color rounded-lg">
                                    <div className="flex items-center justify-between mb-2">
                                        <span className="text-[10px] font-bold text-text-muted uppercase">Query Definition</span>
                                        <span className="text-[10px] text-success font-bold flex items-center gap-1">
                                            <span className="material-symbols-outlined text-[12px]">check_circle</span> Use valid JSON
                                        </span>
                                    </div>
                                    <textarea name="filters" onChange={handleChange} value={formData.filters} className="w-full bg-[#0d151c] border border-border-color rounded font-mono text-xs text-emerald-500 p-4 focus:ring-1 focus:ring-primary focus:border-primary outline-none" rows={6}></textarea>
                                </div>
                            </div>
                        </section>
                    </div>
                </div>
            </main>
        </div>
    );
};

export default CreateReport;
