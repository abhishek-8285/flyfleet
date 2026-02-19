import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

const CreateBackupView: React.FC = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [formData, setFormData] = useState({
        backup_type: 'FULL',
        location: '',
        retention_days: 30,
        remarks: '',
        initiated_by: 'Alex Morgan (alex.m@fleet.com)' // Hardcoded for now per UI spec
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async () => {
        if (!formData.location) {
            alert("Please select a storage location");
            return;
        }

        setLoading(true);
        setError(null);
        try {
            await apiClient.post('/admin/dr/backups', {
                backup_type: formData.backup_type,
                location: formData.location,
                retention_days: Number(formData.retention_days),
                remarks: formData.remarks,
                initiated_by: formData.initiated_by
            });
            navigate('/admin/dr');
        } catch (err) {
            console.error(err);
            setError('Failed to start backup');
            alert("Failed to start backup");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
             <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex flex-col">
                    <nav className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-wider font-semibold mb-0.5">
                        <button onClick={() => navigate('/admin/dr')} className="hover:text-primary transition-colors">Disaster Recovery</button>
                        <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                        <button onClick={() => navigate('/admin/dr')} className="hover:text-primary transition-colors">Backups</button>
                        <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                        <span className="text-white">Create Job</span>
                    </nav>
                    <h2 className="text-white text-lg font-bold leading-tight">Create Backup Job</h2>
                </div>
                <div className="flex items-center gap-3">
                    <button onClick={() => navigate('/admin/dr')} className="px-4 py-2 border border-border-color text-text-muted hover:text-white hover:bg-surface-dark text-xs font-semibold rounded transition-colors">
                        Cancel
                    </button>
                    <button 
                        onClick={handleSubmit} 
                        disabled={loading}
                        className={`px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20 ${loading ? 'opacity-50 cursor-not-allowed' : ''}`}>
                        <span className="material-symbols-outlined text-[16px]">play_circle</span>
                        {loading ? 'Starting...' : 'Start Backup Job'}
                    </button>
                </div>
            </header>
            <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
                <div className="max-w-4xl mx-auto space-y-8">
                    <section className="space-y-4">
                        <div className="flex items-center gap-2 mb-2">
                            <span className="material-symbols-outlined text-primary">settings_backup_restore</span>
                            <h3 className="text-white font-semibold text-base">Backup Configuration</h3>
                        </div>
                        <div className="bg-surface-dark border border-border-color rounded-xl p-6">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Backup Type</label>
                                    <div className="grid grid-cols-2 gap-2 p-1 bg-surface-darker border border-border-color rounded-lg">
                                        <label className="cursor-pointer">
                                            <input 
                                                className="sr-only peer" 
                                                name="backup_type" 
                                                type="radio" 
                                                value="FULL"
                                                checked={formData.backup_type === 'FULL'}
                                                onChange={handleChange}
                                            />
                                            <div className="flex items-center justify-center py-2 px-3 rounded-md text-xs font-medium text-text-muted peer-checked:bg-primary peer-checked:text-white transition-all">
                                                FULL
                                            </div>
                                        </label>
                                        <label className="cursor-pointer">
                                            <input 
                                                className="sr-only peer" 
                                                name="backup_type" 
                                                type="radio" 
                                                value="INCREMENTAL"
                                                checked={formData.backup_type === 'INCREMENTAL'}
                                                onChange={handleChange}
                                            />
                                            <div className="flex items-center justify-center py-2 px-3 rounded-md text-xs font-medium text-text-muted peer-checked:bg-primary peer-checked:text-white transition-all">
                                                INCREMENTAL
                                            </div>
                                        </label>
                                    </div>
                                    <p className="text-[10px] text-text-muted">Full: complete snapshot of all data. Incremental: only changes since last job.</p>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Storage Location</label>
                                    <div className="relative">
                                        <select 
                                            name="location"
                                            value={formData.location}
                                            onChange={handleChange}
                                            className={`w-full bg-surface-darker border text-sm text-white rounded-lg px-4 py-2.5 outline-none focus:ring-1 focus:ring-danger appearance-none ${!formData.location ? 'border-danger' : 'border-border-color focus:border-primary focus:ring-primary'}`}>
                                            <option value="">Select location...</option>
                                            <option value="AWS-US-EAST-1-HOT">HOT (Amazon S3 - Standard)</option>
                                            <option value="AWS-US-EAST-1-COLD">COLD (Amazon S3 Glacier)</option>
                                            <option value="AZURE-WEST-EUROPE">Azure Blob Storage (West Europe)</option>
                                        </select>
                                        <div className="absolute right-3 top-2.5 pointer-events-none text-text-muted">
                                            <span className="material-symbols-outlined text-[18px]">expand_more</span>
                                        </div>
                                    </div>
                                    {!formData.location && (
                                        <p className="text-[10px] text-danger flex items-center gap-1">
                                            <span className="material-symbols-outlined text-[12px]">error</span>
                                            Validation Error: Please select a storage location
                                        </p>
                                    )}
                                </div>
                                <div className="md:col-span-2 space-y-2">
                                    <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Retention Policy (Days)</label>
                                    <div className="flex flex-col md:flex-row items-center gap-4">
                                        <div className="flex-1 w-full">
                                            <input 
                                                name="retention_days"
                                                value={formData.retention_days}
                                                onChange={handleChange}
                                                className="w-full bg-surface-darker border border-border-color focus:border-primary focus:ring-primary text-sm text-white rounded-lg px-4 py-2.5 outline-none" 
                                                type="number" 
                                            />
                                        </div>
                                        {Number(formData.retention_days) > 90 && (
                                            <div className="flex-1 w-full flex items-center gap-4 px-4 py-2 bg-warning/10 border border-warning/30 rounded-lg">
                                                <span className="material-symbols-outlined text-warning text-[20px]">warning</span>
                                                <p className="text-[11px] text-warning font-medium leading-tight">Retention periods over 90 days may incur additional cold storage costs.</p>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </section>

                    <section className="space-y-4">
                        <div className="flex items-center gap-2 mb-2">
                            <span className="material-symbols-outlined text-primary">info</span>
                            <h3 className="text-white font-semibold text-base">Metadata</h3>
                        </div>
                        <div className="bg-surface-dark border border-border-color rounded-xl p-6">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="md:col-span-2 space-y-2">
                                    <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Remarks / Description</label>
                                    <textarea 
                                        name="remarks"
                                        value={formData.remarks}
                                        onChange={handleChange}
                                        className="w-full bg-surface-darker border border-border-color focus:border-primary focus:ring-primary text-sm text-white rounded-lg px-4 py-2.5 outline-none resize-none" 
                                        placeholder="Describe the purpose of this backup job..." 
                                        rows={3}
                                    ></textarea>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Initiated By</label>
                                    <div className="flex items-center gap-3 px-4 py-2.5 bg-surface-darker border border-border-color rounded-lg text-sm text-text-muted">
                                        <span className="material-symbols-outlined text-[18px]">person</span>
                                        <span>{formData.initiated_by}</span>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Job Identifier</label>
                                    <div className="px-4 py-2.5 bg-surface-darker border border-border-color rounded-lg text-sm font-mono text-text-muted">
                                        DR-BACKUP-{new Date().toISOString().slice(0,10).replace(/-/g,'')}-AUTO
                                    </div>
                                </div>
                            </div>
                        </div>
                    </section>

                    {error && (
                        <div className="flex items-start gap-3 p-4 bg-danger/10 border border-danger/20 rounded-lg text-danger text-xs">
                            <span className="material-symbols-outlined text-[18px]">error</span>
                            <span>{error}</span>
                        </div>
                    )}

                    <div className="flex items-start gap-3 p-4 bg-primary/5 border border-primary/20 rounded-lg">
                        <span className="material-symbols-outlined text-primary text-[20px]">info</span>
                        <div>
                            <p className="text-xs text-white font-medium">API Endpoint: <code className="bg-surface-darker px-1 rounded text-primary">POST /dr/backups</code></p>
                            <p className="text-[11px] text-text-muted mt-1">Submit this form to trigger a Disaster Recovery backup via the Web BFF. All operations are logged for compliance audits.</p>
                        </div>
                    </div>
                </div>
            </div>
            <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20">
                <span>© 2023 Enterprise Fleet Manager. DR Module V1.2</span>
                <div className="flex items-center gap-4">
                    <span>Data Source: <span className="text-primary">POST /dr/backups</span></span>
                </div>
            </footer>
        </div>
    );
};

export default CreateBackupView;
