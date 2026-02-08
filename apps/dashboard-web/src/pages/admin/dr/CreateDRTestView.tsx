import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const CreateDRTestView: React.FC = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        test_type: 'FAILOVER',
        target_env: 'STAGING',
        success_criteria: 'RTO: 30 min, RPO: 5 min',
        scheduled_at: ''
    });
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            const res = await fetch('/api/v1/admin/dr/tests', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData)
            });
            if (res.ok) {
                const data = await res.json();
                navigate(`/admin/dr/tests/${data.data.id}`);
            }
        } catch (error) {
            console.error('Failed to create DR test:', error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex flex-col h-full w-full bg-[#0a0a0f] text-white font-sans overflow-hidden">
            {/* Header */}
            <header className="h-16 flex items-center justify-between px-6 border-b border-white/10 bg-[#12121a] flex-shrink-0">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/admin/dr/tests')} className="text-white/60 hover:text-white">
                        ← Back
                    </button>
                    <h1 className="text-lg font-semibold">Schedule DR Test</h1>
                </div>
            </header>

            {/* Form */}
            <div className="flex-1 overflow-auto p-6">
                <form onSubmit={handleSubmit} className="max-w-2xl mx-auto space-y-6">
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-6 space-y-6">
                        <div>
                            <label className="block text-sm text-white/70 mb-2">Test Name</label>
                            <input
                                type="text"
                                value={formData.name}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                                placeholder="Monthly Failover Drill"
                                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-cyan-500"
                                required
                            />
                        </div>

                        <div>
                            <label className="block text-sm text-white/70 mb-2">Description</label>
                            <textarea
                                value={formData.description}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                                placeholder="Describe the test scenario..."
                                rows={3}
                                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-cyan-500"
                            />
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm text-white/70 mb-2">Test Type</label>
                                <select
                                    value={formData.test_type}
                                    onChange={(e) => setFormData({...formData, test_type: e.target.value})}
                                    className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-cyan-500"
                                >
                                    <option value="FAILOVER">🔄 Failover Simulation</option>
                                    <option value="BACKUP">💾 Backup Verification</option>
                                    <option value="RESTORE">📥 Restore Drill</option>
                                </select>
                            </div>

                            <div>
                                <label className="block text-sm text-white/70 mb-2">Target Environment</label>
                                <select
                                    value={formData.target_env}
                                    onChange={(e) => setFormData({...formData, target_env: e.target.value})}
                                    className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-cyan-500"
                                >
                                    <option value="STAGING">Staging</option>
                                    <option value="UAT">UAT</option>
                                    <option value="PROD">Production</option>
                                </select>
                            </div>
                        </div>

                        <div>
                            <label className="block text-sm text-white/70 mb-2">Success Criteria</label>
                            <input
                                type="text"
                                value={formData.success_criteria}
                                onChange={(e) => setFormData({...formData, success_criteria: e.target.value})}
                                placeholder="RTO: 30 min, RPO: 5 min"
                                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-cyan-500"
                            />
                        </div>

                        <div>
                            <label className="block text-sm text-white/70 mb-2">Schedule (Optional)</label>
                            <input
                                type="datetime-local"
                                value={formData.scheduled_at}
                                onChange={(e) => setFormData({...formData, scheduled_at: e.target.value ? new Date(e.target.value).toISOString() : ''})}
                                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-cyan-500"
                            />
                            <p className="text-xs text-white/40 mt-1">Leave empty to run immediately</p>
                        </div>
                    </div>

                    {formData.target_env === 'PROD' && (
                        <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 flex items-center gap-3">
                            <span className="text-2xl">⚠️</span>
                            <div>
                                <p className="text-red-400 font-medium">Production Environment Selected</p>
                                <p className="text-sm text-red-300/70">This test will execute against production systems. Ensure proper approvals are in place.</p>
                            </div>
                        </div>
                    )}

                    <div className="flex justify-end gap-3">
                        <button
                            type="button"
                            onClick={() => navigate('/admin/dr/tests')}
                            className="px-6 py-3 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={loading || !formData.name}
                            className="px-6 py-3 bg-gradient-to-r from-emerald-500 to-cyan-500 hover:opacity-90 rounded-lg font-medium disabled:opacity-50"
                        >
                            {loading ? 'Scheduling...' : 'Schedule Test'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default CreateDRTestView;
