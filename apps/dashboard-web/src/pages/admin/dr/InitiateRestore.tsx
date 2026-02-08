import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface Backup {
    id: string;
    backup_type: string;
    created_at: string;
    size_bytes?: number;
    status: string;
}

const InitiateRestore: React.FC = () => {
    const navigate = useNavigate();
    const [backups, setBackups] = useState<Backup[]>([]);
    const [selectedBackup, setSelectedBackup] = useState<string>('');
    const [targetEnv, setTargetEnv] = useState('STAGING');
    const [integrityCheck, setIntegrityCheck] = useState(true);
    const [confirmText, setConfirmText] = useState('');
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchBackups = async () => {
            try {
                const res = await fetch('/api/v1/admin/dr/backups');
                const data = await res.json();
                // Filter for successfully COMPLETED/VERIFIED backups only
                const validBackups = (data.data || []).filter((b: Backup) => ['COMPLETED', 'VERIFIED'].includes(b.status));
                setBackups(validBackups);
            } catch (err) {
                console.error(err);
            } finally {
                setLoading(false);
            }
        };
        fetchBackups();
    }, []);

    const handleSubmit = async () => {
        if (!selectedBackup) return;
        
        try {
            const res = await fetch('/api/v1/admin/dr/restores', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    backup_id: selectedBackup,
                    target_env: targetEnv,
                    integrity_check: integrityCheck,
                    requested_by: 'Alex Morgan (alex.m@fleet.com)'
                })
            });

            if (res.ok) {
                navigate('/admin/dr/restores');
            } else {
                alert("Failed to initiate restore");
            }
        } catch (err) {
            console.error(err);
        }
    };

    const formatSize = (bytes?: number) => bytes ? `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB` : '---';

    return (
        <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
             <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex flex-col">
                    <nav className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-widest font-semibold mb-1">
                        <button onClick={() => navigate('/admin/dr/restores')} className="hover:text-primary transition-colors">Restore Operations</button>
                        <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                        <span className="text-primary">Initiate</span>
                    </nav>
                    <h2 className="text-white text-lg font-bold leading-tight">Initiate System Restore</h2>
                </div>
                <div className="flex items-center gap-3">
                    <button onClick={() => navigate('/admin/dr/restores')} className="px-4 py-2 border border-border-color hover:bg-surface-dark text-text-muted hover:text-white text-xs font-semibold rounded transition-colors">
                        Cancel
                    </button>
                    <button 
                        disabled={confirmText !== 'RESTORE' || !selectedBackup}
                        onClick={handleSubmit}
                        className={`px-4 py-2 bg-danger hover:bg-red-600 text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-danger/20 ${confirmText !== 'RESTORE' || !selectedBackup ? 'opacity-50 cursor-not-allowed' : ''}`}
                    >
                        <span className="material-symbols-outlined text-[16px]">dangerous</span>
                        Start Full System Restore
                    </button>
                </div>
            </header>

            <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
                <div className="max-w-5xl mx-auto space-y-8">
                    <section className="space-y-4">
                        <div className="flex items-center justify-between border-b border-border-color pb-2">
                            <div className="flex items-center gap-2 text-primary">
                                <span className="material-symbols-outlined text-xl">database</span>
                                <h3 className="font-semibold text-white uppercase tracking-wider text-sm">Restore Point Selection</h3>
                            </div>
                            <div className="relative">
                                <input className="bg-surface-dark border border-border-color text-xs text-white rounded px-3 py-2 w-72 pl-9 focus:ring-1 focus:ring-primary focus:border-primary outline-none" placeholder="Search backups by ID..." type="text"/>
                                <span className="material-symbols-outlined absolute left-2.5 top-2.5 text-text-muted text-[18px]">search</span>
                            </div>
                        </div>
                        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden max-h-[300px] overflow-y-auto custom-scrollbar">
                            <table className="w-full text-left border-collapse">
                                <thead className="sticky top-0 z-10">
                                    <tr className="bg-surface-darker/90 backdrop-blur text-xs text-text-muted uppercase tracking-wider border-b border-border-color">
                                        <th className="px-6 py-4 font-semibold w-12 text-center">Select</th>
                                        <th className="px-6 py-4 font-semibold">Snapshot ID</th>
                                        <th className="px-6 py-4 font-semibold">Creation Date</th>
                                        <th className="px-6 py-4 font-semibold">Size</th>
                                        <th className="px-6 py-4 font-semibold">Type</th>
                                        <th className="px-6 py-4 font-semibold text-right">Status</th>
                                    </tr>
                                </thead>
                                <tbody className="text-sm">
                                    {loading ? (
                                        <tr><td colSpan={6} className="px-6 py-4 text-center text-text-muted">Loading backups...</td></tr>
                                    ) : backups.map((snap) => (
                                        <tr key={snap.id} 
                                            className={`border-b border-border-color transition-colors hover:bg-primary/5 cursor-pointer ${selectedBackup === snap.id ? 'bg-primary/10' : ''}`}
                                            onClick={() => setSelectedBackup(snap.id)}
                                        >
                                            <td className="px-6 py-4 text-center">
                                                <input 
                                                    checked={selectedBackup === snap.id}
                                                    onChange={() => setSelectedBackup(snap.id)}
                                                    className="w-4 h-4 text-primary bg-surface-darker border-border-color focus:ring-primary focus:ring-offset-surface-darker" 
                                                    name="restore_point" 
                                                    type="radio"
                                                />
                                            </td>
                                            <td className="px-6 py-4 font-mono text-xs text-white">{snap.id.substring(0, 8)}...</td>
                                            <td className="px-6 py-4 text-text-muted">{new Date(snap.created_at).toLocaleString()}</td>
                                            <td className="px-6 py-4 text-text-muted">{formatSize(snap.size_bytes)}</td>
                                            <td className="px-6 py-4 text-text-muted">{snap.backup_type}</td>
                                            <td className="px-6 py-4 text-right">
                                                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold bg-success/10 text-success border border-success/20">
                                                    VALID
                                                </span>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </section>
                    
                    {/* ... Configuration and Confirmation Sections same as user provided ... */}
                    <section className="grid grid-cols-1 md:grid-cols-2 gap-8">
                        <div className="space-y-4">
                            <div className="flex items-center gap-2 text-primary border-b border-border-color pb-2">
                                <span className="material-symbols-outlined text-xl">tune</span>
                                <h3 className="font-semibold text-white uppercase tracking-wider text-sm">Configuration</h3>
                            </div>
                            <div className="bg-surface-dark rounded-xl border border-border-color p-6 space-y-6">
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-text-muted uppercase">Target Environment</label>
                                    <div className="grid grid-cols-3 gap-3">
                                        {['PROD', 'STAGING', 'UAT'].map((env) => (
                                            <label key={env} className="cursor-pointer">
                                                <input 
                                                    className="sr-only peer" 
                                                    name="env" 
                                                    type="radio" 
                                                    value={env} 
                                                    checked={targetEnv === env}
                                                    onChange={(e) => setTargetEnv(e.target.value)}
                                                />
                                                <div className={`p-3 text-center rounded-lg border border-border-color bg-surface-darker peer-checked:bg-${env === 'PROD' ? 'danger' : env === 'STAGING' ? 'primary' : 'success'}/10 peer-checked:text-${env === 'PROD' ? 'danger' : env === 'STAGING' ? 'primary' : 'success'} peer-checked:border-${env === 'PROD' ? 'danger' : env === 'STAGING' ? 'primary' : 'success'} transition-all`}>
                                                    <div className="text-xs font-bold">{env}</div>
                                                </div>
                                            </label>
                                        ))}
                                    </div>
                                    <p className="text-[10px] text-text-muted italic">Warning: PROD restore will cause system downtime.</p>
                                </div>
                                <div className="flex items-center justify-between p-4 bg-surface-darker rounded-lg border border-border-color">
                                    <div className="flex flex-col">
                                        <span className="text-sm font-medium text-white">Pre-Restore Integrity Check</span>
                                        <span className="text-[11px] text-text-muted">Validate all blocks before overwriting target</span>
                                    </div>
                                    <label className="relative inline-flex items-center cursor-pointer">
                                        <input 
                                            checked={integrityCheck}
                                            onChange={(e) => setIntegrityCheck(e.target.checked)}
                                            className="sr-only peer" 
                                            type="checkbox"
                                        />
                                        <div className="w-10 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-success"></div>
                                    </label>
                                </div>
                            </div>
                        </div>

                        <div className="space-y-4">
                            <div className="flex items-center gap-2 text-primary border-b border-border-color pb-2">
                                <span className="material-symbols-outlined text-xl">verified_user</span>
                                <h3 className="font-semibold text-white uppercase tracking-wider text-sm">Confirmation</h3>
                            </div>
                            <div className="bg-surface-dark rounded-xl border border-border-color p-6 space-y-6">
                                <div className="bg-danger/10 border border-danger/30 rounded-lg p-4 flex gap-3">
                                    <span className="material-symbols-outlined text-danger">warning</span>
                                    <div>
                                        <h4 className="text-xs font-bold text-danger uppercase mb-1">Critical: System Overwrite Warning</h4>
                                        <p className="text-xs text-text-muted leading-relaxed">
                                            Initiating a restore will permanently overwrite current data on the target environment. This action is <span className="text-white font-bold underline">irreversible</span> once the process starts.
                                        </p>
                                    </div>
                                </div>
                                <div className="space-y-4">
                                    <div className="space-y-2">
                                        <label className="text-xs font-semibold text-text-muted uppercase">Verification Step 1</label>
                                        <div className="flex items-center gap-3">
                                            <input className="w-4 h-4 rounded border-border-color bg-surface-darker text-primary focus:ring-primary" required type="checkbox"/>
                                            <span className="text-xs text-white">I confirm I have selected the correct restore point.</span>
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-xs font-semibold text-text-muted uppercase">Verification Step 2</label>
                                        <input 
                                            value={confirmText}
                                            onChange={(e) => setConfirmText(e.target.value.toUpperCase())}
                                            className="w-full bg-surface-darker border border-border-color text-xs text-white rounded px-3 py-2.5 focus:ring-1 focus:ring-danger focus:border-danger outline-none" 
                                            placeholder="Type RESTORE to confirm" 
                                            type="text"
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </section>
                </div>
            </div>
             <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
                <span>© 2023 Enterprise Fleet Manager. DR Module V2.4</span>
            </footer>
        </div>
    );
};

export default InitiateRestore;
