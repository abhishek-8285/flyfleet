import React, { useEffect, useState } from 'react';

const toArray = (value: any) => {
    if (Array.isArray(value)) return value;
    if (Array.isArray(value?.items)) return value.items;
    if (Array.isArray(value?.data)) return value.data;
    return [];
};

const UserConsole: React.FC = () => {
    const [profile, setProfile] = useState<any>(null);
    const [security, setSecurity] = useState<any>(null);
    const [sessions, setSessions] = useState<any[]>([]);
    const [logins, setLogins] = useState<any[]>([]);
    const [roles, setRoles] = useState<any>(null);
    const [permissions, setPermissions] = useState<any>(null);
    const [notes, setNotes] = useState<any[]>([]);
    const [newNote, setNewNote] = useState('');
    const [searchId, setSearchId] = useState(localStorage.getItem('support_last_user_id') || '');
    const [activeUserId, setActiveUserId] = useState(localStorage.getItem('support_last_user_id') || '');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [actionMessage, setActionMessage] = useState<string | null>(null);

    const fetchData = async (userId: string) => {
        if (!userId.trim()) return;
        const token = localStorage.getItem('access_token');
        const headers = { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' };

        try {
            setIsLoading(true);
            setError(null);
            setActionMessage(null);
            const [pRes, sRes, sessRes, lRes, rRes, permRes, nRes] = await Promise.all([
                fetch(`/bff/support/users/${userId}/profile`, { headers }),
                fetch(`/bff/support/users/${userId}/security`, { headers }),
                fetch(`/bff/support/users/${userId}/sessions`, { headers }),
                fetch(`/bff/support/users/${userId}/logins`, { headers }),
                fetch(`/bff/support/users/${userId}/roles`, { headers }),
                fetch(`/bff/support/users/${userId}/permissions`, { headers }),
                fetch(`/bff/support/users/${userId}/notes`, { headers })
            ]);

            if (pRes.ok) setProfile(await pRes.json());
            if (sRes.ok) setSecurity(await sRes.json());
            if (sessRes.ok) setSessions(toArray(await sessRes.json()));
            if (lRes.ok) setLogins(toArray(await lRes.json()));
            if (rRes.ok) setRoles(await rRes.json());
            if (permRes.ok) setPermissions(await permRes.json());
            if (nRes.ok) setNotes(toArray(await nRes.json()));
            localStorage.setItem('support_last_user_id', userId);
        } catch (err) {
            console.error('Failed to fetch user console data', err);
            setError('Failed to load user data');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        if (activeUserId) fetchData(activeUserId);
    }, [activeUserId]);

    const runSearch = () => {
        const userId = searchId.trim();
        if (!userId) return;
        setActiveUserId(userId);
    };

    const handleAction = async (action: string) => {
        const token = localStorage.getItem('access_token');
        const headers = { 'Authorization': `Bearer ${token}` };
        try {
            const res = await fetch(`/bff/support/users/${activeUserId}/${action}`, { method: 'POST', headers });
            if (res.ok) {
                setActionMessage(`${action} executed successfully`);
            } else {
                setError(`Failed to execute ${action}`);
            }
        } catch (err) {
            console.error(err);
            setError(`Failed to execute ${action}`);
        }
    };

    const handleRevoke = async (sessionId: string) => {
        const token = localStorage.getItem('access_token');
        const headers = { 'Authorization': `Bearer ${token}` };
        try {
            const res = await fetch(`/bff/support/users/${activeUserId}/sessions/${sessionId}/revoke`, { method: 'POST', headers });
            if (res.ok) {
                setSessions(sessions.filter(s => s.sessionId !== sessionId));
            }
        } catch (err) {
            console.error(err);
        }
    };

    const saveNote = async () => {
        if (!newNote) return;
        const token = localStorage.getItem('access_token');
        const headers = { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' };
        try {
            const res = await fetch(`/bff/support/users/${activeUserId}/notes`, {
                method: 'POST', 
                headers,
                body: JSON.stringify({ note: newNote })
            });
            if (res.ok) {
                setNewNote('');
                fetchData(activeUserId);
            }
        } catch (err) {
            console.error(err);
        }
    };

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden font-display">
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <div className="flex items-center gap-2 text-[10px] text-[#92adc9] uppercase tracking-wider font-semibold mb-0.5">
                            <span className="hover:text-white cursor-pointer uppercase">Support</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-white uppercase">User Console</span>
                        </div>
                        <h2 className="text-white text-lg font-bold">User Troubleshooting: {profile?.name || '--'}</h2>
                    </div>
                    <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-success/20 text-success border border-success/30 flex items-center gap-1">
                        <span className="w-1.5 h-1.5 rounded-full bg-success"></span>
                        Online
                    </span>
                </div>
                <div className="flex items-center gap-4">
                    <div className="relative">
                        <span className="absolute inset-y-0 left-0 flex items-center pl-3">
                            <span className="material-symbols-outlined text-[#92adc9] text-[18px]">search</span>
                        </span>
                        <input 
                            className="bg-[#16202a] border border-[#233648] text-sm text-white rounded-lg pl-9 pr-3 py-1.5 w-64 focus:ring-1 focus:ring-primary outline-none placeholder:text-[#92adc9]/50" 
                            placeholder="Search by User ID, Email..." 
                            type="text"
                            value={searchId}
                            onChange={(e) => setSearchId(e.target.value)}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') runSearch();
                            }}
                        />
                    </div>
                    <button
                        onClick={runSearch}
                        className="px-3 py-1.5 rounded bg-primary text-white text-xs font-semibold hover:bg-primary-dark transition-colors"
                    >
                        Load
                    </button>
                </div>
            </header>

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                {isLoading && (
                    <div className="mb-4 rounded border border-primary/20 bg-primary/10 px-3 py-2 text-xs text-primary">
                        Loading user data...
                    </div>
                )}
                {error && (
                    <div className="mb-4 rounded border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-300">
                        {error}
                    </div>
                )}
                {actionMessage && (
                    <div className="mb-4 rounded border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-xs text-emerald-300">
                        {actionMessage}
                    </div>
                )}
                <div className="grid grid-cols-12 gap-6 h-full">
                    {/* Sidebar: Profile & Sessions */}
                    <div className="col-span-12 lg:col-span-3 flex flex-col gap-6">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
                            <div className="p-5 border-b border-[#233648] bg-[#111a22]/50">
                                <div className="flex items-start justify-between">
                                    <div className="w-16 h-16 rounded-full bg-primary/20 border-2 border-primary text-primary flex items-center justify-center text-2xl font-bold">
                                        {profile?.name?.split(' ').map((n: string) => n[0]).join('') || 'SJ'}
                                    </div>
                                    <div className="flex flex-col items-end">
                                        <span className="text-[10px] text-[#92adc9] uppercase font-bold tracking-wider">User ID</span>
                                        <span className="text-white font-mono text-xs mt-1">{profile?.userId || activeUserId || '--'}</span>
                                    </div>
                                </div>
                                <div className="mt-4">
                                    <h3 className="text-lg font-bold text-white">{profile?.name || '--'}</h3>
                                    <p className="text-sm text-[#92adc9]">{profile?.email || '--'}</p>
                                    <div className="flex items-center gap-2 mt-2">
                                        <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-[10px] font-medium uppercase">{profile?.role || '--'}</span>
                                        <span className="px-2 py-0.5 rounded bg-[#111a22] text-[#92adc9] border border-[#233648] text-[10px] font-medium uppercase">{profile?.region || '--'}</span>
                                    </div>
                                </div>
                            </div>
                            <div className="p-5 space-y-4">
                                <div>
                                    <div className="text-[10px] uppercase text-[#92adc9] font-bold mb-2 tracking-widest">Account Status</div>
                                    <div className="flex items-center gap-2 text-sm text-white">
                                        <span className="material-symbols-outlined text-success text-[18px]">check_circle</span>
                                        {profile?.status || '--'}
                                    </div>
                                    <div className="text-[10px] text-[#92adc9] mt-1 italic">Member since {profile?.memberSince || '--'}</div>
                                </div>
                                <div>
                                    <div className="text-[10px] uppercase text-[#92adc9] font-bold mb-2 tracking-widest">MFA Status</div>
                                    <div className="flex items-center gap-2 text-sm text-white">
                                        <span className="material-symbols-outlined text-success text-[18px]">{security?.mfaEnabled ? 'lock' : 'lock_open'}</span>
                                        {security?.mfaEnabled ? 'Enforced' : 'Disabled'}
                                    </div>
                                    <div className="flex flex-wrap gap-1 mt-2">
                                        {security?.mfaMethods?.map((m: string) => (
                                            <span key={m} className="px-2 py-0.5 rounded bg-[#111a22] text-[#92adc9] text-[9px] border border-[#233648] uppercase font-bold tracking-tight">
                                                {m.replace('_', ' ')}
                                            </span>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col flex-1 overflow-hidden">
                            <div className="p-4 border-b border-[#233648] bg-[#111a22]/50 flex justify-between items-center">
                                <h3 className="text-xs font-bold text-white uppercase tracking-widest">Active Sessions</h3>
                                <span className="text-[10px] bg-success/20 text-success px-1.5 py-0.5 rounded border border-success/20">{sessions.length} Active</span>
                            </div>
                            <div className="divide-y divide-[#233648] overflow-y-auto custom-scrollbar">
                                {sessions.map((session, idx) => (
                                    <div key={idx} className="p-4 hover:bg-[#111a22] transition-colors group relative">
                                        <div className="absolute right-4 top-4 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <button onClick={() => handleRevoke(session.sessionId)} className="text-[10px] text-danger hover:underline font-bold uppercase">Revoke</button>
                                        </div>
                                        <div className="flex items-center gap-3 mb-2">
                                            <span className="material-symbols-outlined text-[#92adc9] text-[20px]">
                                                {session.device?.includes('iPhone') ? 'smartphone' : 'laptop_mac'}
                                            </span>
                                            <span className="text-sm text-white font-medium">{session.device}</span>
                                        </div>
                                        <div className="flex flex-col gap-1 pl-9">
                                            <div className="flex items-center gap-2 text-[10px] text-[#92adc9]">
                                                <span className={`w-1.5 h-1.5 rounded-full ${session.current ? 'bg-success' : 'bg-warning'}`}></span>
                                                <span>{session.location} • {session.ip}</span>
                                            </div>
                                            <div className="text-[9px] text-[#92adc9]/60 uppercase font-bold">{session.browser || 'Unknown'} • {session.current ? 'Current Session' : 'Idle'}</div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Main Content: Login History & Roles */}
                    <div className="col-span-12 lg:col-span-7 flex flex-col gap-6">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col h-[500px] overflow-hidden">
                            <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                                <h3 className="text-white font-bold text-sm flex items-center gap-2 uppercase tracking-wide">
                                    <span className="material-symbols-outlined text-primary text-[20px]">manage_history</span>
                                    Recent Login History
                                </h3>
                            </div>
                            <div className="flex-1 overflow-auto custom-scrollbar">
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="border-b border-[#233648] text-[10px] text-[#92adc9] uppercase tracking-widest bg-[#111a22]/30 sticky top-0 backdrop-blur-sm z-10">
                                            <th className="px-4 py-3 font-bold">Time</th>
                                            <th className="px-4 py-3 font-bold">Event</th>
                                            <th className="px-4 py-3 font-bold">IP/Location</th>
                                            <th className="px-4 py-3 font-bold">Risk</th>
                                            <th className="px-4 py-3 font-bold text-right">Result</th>
                                        </tr>
                                    </thead>
                                    <tbody className="text-sm">
                                        {logins.map((login, idx) => (
                                            <tr key={idx} className={`border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors ${login.result === 'FAILED' ? 'bg-danger/5' : ''}`}>
                                                <td className="px-4 py-3 text-white text-[10px] whitespace-nowrap">
                                                    {new Date(login.time).toLocaleString()}
                                                </td>
                                                <td className="px-4 py-3 text-[#92adc9] text-[10px] uppercase font-bold">
                                                    {login.eventType}
                                                </td>
                                                <td className="px-4 py-3">
                                                    <div className="flex flex-col">
                                                        <span className="text-white text-[10px] font-mono">{login.ip}</span>
                                                        <span className="text-[9px] text-[#92adc9]">{login.location}</span>
                                                    </div>
                                                </td>
                                                <td className="px-4 py-3">
                                                    <div className="w-16 bg-[#111a22] rounded-full h-1 overflow-hidden mb-1">
                                                        <div className={`h-full ${login.riskScore > 50 ? 'bg-danger' : 'bg-success'}`} style={{ width: `${login.riskScore}%` }}></div>
                                                    </div>
                                                    <span className={`text-[9px] uppercase font-bold ${login.riskScore > 50 ? 'text-danger' : 'text-success'}`}>
                                                        {login.riskScore > 50 ? 'High' : 'Low'}
                                                    </span>
                                                </td>
                                                <td className="px-4 py-3 text-right">
                                                    <span className={`px-2 py-0.5 rounded text-[9px] font-bold border ${login.result === 'SUCCESS' ? 'bg-success/10 text-success border-success/20' : 'bg-danger/10 text-danger border-danger/20'}`}>
                                                        {login.result}
                                                    </span>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>

                        <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 flex flex-col h-[200px]">
                            <div className="flex justify-between items-center mb-4 border-b border-[#233648] pb-2">
                                <h3 className="text-white text-xs font-bold uppercase tracking-widest flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[#92adc9] text-[18px]">verified_user</span>
                                    Permission & Role Summary
                                </h3>
                            </div>
                            <div className="flex gap-6 h-full overflow-hidden">
                                <div className="w-1/3 border-r border-[#233648] pr-6 overflow-y-auto custom-scrollbar">
                                    <div className="text-[9px] uppercase text-[#92adc9] font-bold mb-2">Assigned Roles</div>
                                    <div className="space-y-1.5">
                                        {roles?.roles?.map((r: string) => (
                                            <div key={r} className="flex items-center justify-between text-[11px] text-white bg-[#111a22] p-2 rounded border border-[#233648]">
                                                {r}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                                <div className="w-2/3 pl-2 overflow-y-auto custom-scrollbar">
                                    <div className="text-[9px] uppercase text-[#92adc9] font-bold mb-2">Effective Permissions (Privileged)</div>
                                    <div className="grid grid-cols-2 gap-2">
                                        {permissions?.permissions?.map((p: string) => (
                                            <div key={p} className="flex items-center gap-2 text-[10px] text-[#92adc9]">
                                                <span className="material-symbols-outlined text-success text-[14px]">check</span>
                                                {p.replace('_', ' ')}
                                            </div>
                                        ))}
                                        <div className="flex items-center gap-2 text-[10px] text-[#92adc9]">
                                            <span className="material-symbols-outlined text-danger text-[14px]">close</span>
                                            System Configuration
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Quick Support Actions */}
                    <div className="col-span-12 lg:col-span-2 flex flex-col gap-6">
                        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden h-full">
                            <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                                <h3 className="text-white text-xs font-bold uppercase tracking-widest">Admin Actions</h3>
                                <p className="text-[9px] text-[#92adc9] mt-1 font-medium">Privileged Commands</p>
                            </div>
                            <div className="p-4 flex flex-col gap-2 overflow-y-auto custom-scrollbar">
                                <button disabled={!activeUserId} onClick={() => handleAction('reset-password')} className="w-full group p-2.5 rounded border border-[#233648] hover:border-primary/50 hover:bg-[#111a22] transition-all text-left disabled:opacity-50">
                                    <div className="flex items-center gap-2 mb-1">
                                        <div className="p-1.5 rounded bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white transition-colors">
                                            <span className="material-symbols-outlined text-[16px]">lock_reset</span>
                                        </div>
                                        <span className="text-[11px] font-bold text-white uppercase tracking-tighter">Reset Pass</span>
                                    </div>
                                </button>
                                <button disabled={!activeUserId} onClick={() => handleAction('unlock')} className="w-full group p-2.5 rounded border border-[#233648] hover:border-success/50 hover:bg-[#111a22] transition-all text-left disabled:opacity-50">
                                    <div className="flex items-center gap-2 mb-1">
                                        <div className="p-1.5 rounded bg-success/10 text-success group-hover:bg-success group-hover:text-white transition-colors">
                                            <span className="material-symbols-outlined text-[16px]">lock_open</span>
                                        </div>
                                        <span className="text-[11px] font-bold text-white uppercase tracking-tighter">Unlock Acc</span>
                                    </div>
                                </button>
                                <button disabled={!activeUserId} onClick={() => handleAction('mfa-reset')} className="w-full group p-2.5 rounded border border-[#233648] hover:border-warning/50 hover:bg-[#111a22] transition-all text-left disabled:opacity-50">
                                    <div className="flex items-center gap-2 mb-1">
                                        <div className="p-1.5 rounded bg-warning/10 text-warning group-hover:bg-warning group-hover:text-white transition-colors">
                                            <span className="material-symbols-outlined text-[16px]">phonelink_setup</span>
                                        </div>
                                        <span className="text-[11px] font-bold text-white uppercase tracking-tighter">Reset MFA</span>
                                    </div>
                                </button>
                                <button disabled={!activeUserId} onClick={() => handleAction('force-logout')} className="w-full group p-2.5 rounded border border-danger/30 hover:border-danger hover:bg-danger/10 transition-all text-left disabled:opacity-50">
                                    <div className="flex items-center gap-2 mb-1 text-danger">
                                        <span className="material-symbols-outlined text-[16px]">logout</span>
                                        <span className="text-[11px] font-bold uppercase tracking-tighter">Force Logout</span>
                                    </div>
                                </button>
                                
                                <div className="h-px bg-[#233648] my-2"></div>
                                
                                <div className="bg-[#111a22] p-3 rounded border border-[#233648]">
                                    <div className="text-[9px] uppercase text-[#92adc9] font-bold mb-2">Internal Notes</div>
                                    <textarea 
                                        className="w-full bg-[#16202a] border border-[#233648] rounded p-2 text-[10px] text-white placeholder:text-[#92adc9]/30 focus:ring-1 focus:ring-primary outline-none resize-none h-24 mb-2" 
                                        placeholder="Add internal note..."
                                        value={newNote}
                                        onChange={(e) => setNewNote(e.target.value)}
                                    ></textarea>
                                    <button disabled={!activeUserId} onClick={saveNote} className="w-full py-1.5 bg-primary/10 hover:bg-primary text-primary hover:text-white text-[10px] font-bold uppercase rounded border border-primary/20 transition-all disabled:opacity-50">
                                        Save Note
                                    </button>
                                    
                                    <div className="mt-3 space-y-2 max-h-32 overflow-y-auto custom-scrollbar">
                                        {notes.map((n, i) => (
                                            <div key={i} className="text-[9px] text-[#92adc9] border-l-2 border-[#233648] pl-2 py-1">
                                                <p className="text-white mb-0.5">{n.note}</p>
                                                <span>{n.agent} • {new Date(n.time).toLocaleDateString()}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    );
};

export default UserConsole;
