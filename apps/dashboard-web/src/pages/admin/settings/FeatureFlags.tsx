import React, { useEffect, useState } from 'react';
import apiClient from '../../../services/apiClient';

interface FeatureFlag {
    id: string;
    key: string;
    description: string;
    scope: string;
    enabled: boolean;
    tags: string[];
    updated_at: string;
}

const FeatureFlags: React.FC = () => {
    const [flags, setFlags] = useState<FeatureFlag[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState('');

    useEffect(() => {
        fetchFlags();
    }, []);

    const fetchFlags = async () => {
        setError(null);
        try {
            const response = await apiClient.get('/settings/flags') as any;
            const items = Array.isArray(response)
                ? response
                : Array.isArray(response?.data)
                    ? response.data
                    : [];
            setFlags(items);
        } catch (error) {
            console.error('Failed to fetch flags', error);
            setError('Failed to load feature flags');
        } finally {
            setLoading(false);
        }
    };

    const handleToggle = async (id: string) => {
        // Optimistic update
        const updatedFlags = flags.map(f => 
            f.id === id ? { ...f, enabled: !f.enabled } : f
        );
        setFlags(updatedFlags);

        try {
            await apiClient.patch(`/settings/flags/${id}/toggle`);
        } catch (error) {
            console.error('Failed to toggle flag', error);
            // Revert on failure
            fetchFlags();
        }
    };

    const filteredFlags = flags.filter(flag => 
        flag.key.toLowerCase().includes(searchTerm.toLowerCase()) ||
        flag.description.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const activeCount = flags.filter(f => f.enabled).length;
    const globalCount = flags.filter(f => f.scope === 'GLOBAL').length;

    if (loading) return <div className="p-6 text-white">Loading...</div>;

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
            <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <h2 className="text-white text-lg font-bold">Feature Flags Control Hub</h2>
                        <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5 uppercase tracking-wide font-medium">
                            <span>System Settings</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-white">Feature Flags</span>
                        </div>
                    </div>
                </div>
                <div className="flex items-center gap-4">
                    <div className="relative">
                        <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[#92adc9] text-[18px]">search</span>
                        <input 
                            className="bg-[#16202a] border border-[#233648] text-xs text-white rounded pl-9 pr-3 py-1.5 w-64 focus:ring-1 focus:ring-[#137fec] focus:border-[#137fec] outline-none transition-all placeholder:text-[#92adc9]/50" 
                            placeholder="Search flag keys..." 
                            type="text"
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                        />
                    </div>
                    <button className="px-4 py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-[#137fec]/20">
                        <span className="material-symbols-outlined text-[16px]">add</span>
                        Create Flag
                    </button>
                </div>
            </header>
            <div className="flex flex-1 overflow-hidden">
                <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                    {error && (
                        <div className="mb-4 rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">
                            {error}
                        </div>
                    )}
                    <div className="grid grid-cols-4 gap-4 mb-6">
                        <div className="bg-[#16202a] rounded border border-[#233648] p-3 flex items-center justify-between">
                            <div>
                                <div className="text-[10px] uppercase text-[#92adc9] font-bold">Active Flags</div>
                                <div className="text-xl font-bold text-white">{activeCount}</div>
                            </div>
                            <span className="material-symbols-outlined text-[#137fec] text-2xl">toggle_on</span>
                        </div>
                        <div className="bg-[#16202a] rounded border border-[#233648] p-3 flex items-center justify-between">
                            <div>
                                <div className="text-[10px] uppercase text-[#92adc9] font-bold">Total Flags</div>
                                <div className="text-xl font-bold text-white">{flags.length}</div>
                            </div>
                            <span className="material-symbols-outlined text-[#f59e0b] text-2xl">history_toggle_off</span>
                        </div>
                        <div className="bg-[#16202a] rounded border border-[#233648] p-3 flex items-center justify-between">
                            <div>
                                <div className="text-[10px] uppercase text-[#92adc9] font-bold">Global Scope</div>
                                <div className="text-xl font-bold text-white">{globalCount}</div>
                            </div>
                            <span className="material-symbols-outlined text-[#92adc9] text-2xl">public</span>
                        </div>
                    </div>
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden h-[calc(100vh-230px)]">
                        <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                            <div className="flex items-center gap-3">
                                <button className="text-xs font-semibold text-white border-b-2 border-[#137fec] pb-0.5">All Flags</button>
                            </div>
                            <div className="flex items-center gap-2">
                                <button onClick={fetchFlags} className="p-1.5 hover:bg-[#233648] rounded text-[#92adc9] hover:text-white">
                                    <span className="material-symbols-outlined text-[18px]">refresh</span>
                                </button>
                            </div>
                        </div>
                        <div className="flex-1 overflow-auto custom-scrollbar">
                            {filteredFlags.length === 0 ? (
                                <div className="px-4 py-10 text-sm text-[#92adc9]">
                                    No feature flags available.
                                </div>
                            ) : (
                            <table className="w-full text-left border-collapse">
                                <thead className="sticky top-0 z-10">
                                    <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22] shadow-sm">
                                        <th className="px-4 py-3 font-semibold w-1/4">Flag Key</th>
                                        <th className="px-4 py-3 font-semibold w-1/3">Description</th>
                                        <th className="px-4 py-3 font-semibold w-1/6">Scope</th>
                                        <th className="px-4 py-3 font-semibold w-1/6">Tags</th>
                                        <th className="px-4 py-3 font-semibold text-right w-1/12">Status</th>
                                    </tr>
                                </thead>
                                <tbody className="text-sm divide-y divide-[#233648]">
                                    {filteredFlags.map((flag) => (
                                        <tr key={flag.id} className="hover:bg-[#111a22]/50 transition-colors group">
                                            <td className="px-4 py-3">
                                                <div className="flex flex-col">
                                                    <span className="font-mono text-xs text-[#137fec] font-medium tracking-tight">{flag.key}</span>
                                                    <span className="text-[10px] text-[#92adc9] mt-0.5">Updated {new Date(flag.updated_at).toLocaleDateString()}</span>
                                                </div>
                                            </td>
                                            <td className="px-4 py-3 text-[#92adc9] text-xs">{flag.description}</td>
                                            <td className="px-4 py-3">
                                                <div className="flex items-center gap-1.5 text-xs text-white">
                                                    <span className="material-symbols-outlined text-[14px] text-indigo-400">domain</span>
                                                    {flag.scope}
                                                </div>
                                            </td>
                                            <td className="px-4 py-3">
                                                {flag.tags && flag.tags.map(tag => (
                                                    <span key={tag} className="inline-flex items-center px-1.5 py-0.5 rounded text-[9px] font-bold bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 uppercase tracking-wide mr-1">
                                                        {tag}
                                                    </span>
                                                ))}
                                            </td>
                                            <td className="px-4 py-3 text-right">
                                                <label className="relative inline-flex items-center cursor-pointer">
                                                    <input 
                                                        type="checkbox" 
                                                        checked={flag.enabled} 
                                                        onChange={() => handleToggle(flag.id)}
                                                        className="sr-only peer" 
                                                    />
                                                    <div className="w-8 h-4 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-3 after:w-3 after:transition-all peer-checked:bg-[#10b981]"></div>
                                                </label>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </main>
    );
};

export default FeatureFlags;
