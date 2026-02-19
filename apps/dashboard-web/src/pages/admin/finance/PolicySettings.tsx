import {
    BanknotesIcon,
    PlusIcon,
    ShieldCheckIcon,
    TrashIcon
} from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import apiClient from '../../../services/apiClient';

interface Policy {
    id: string;
    category: string;
    limit_amount: number;
    requires_approval: boolean;
    receipt_required: boolean;
    currency?: string;
}

const PolicySettings: React.FC = () => {
    const [policies, setPolicies] = useState<Policy[]>([]);
    const [savingId, setSavingId] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

    React.useEffect(() => {
        const load = async () => {
            try {
                const response = await apiClient.get('/admin/finance/expenses/policies') as any;
                const rows = Array.isArray(response?.policies) ? response.policies : [];
                setPolicies(rows.map((p: any) => ({
                    id: p.id,
                    category: p.category,
                    limit_amount: Number(p.limit_amount || 0),
                    requires_approval: !!p.requires_approval,
                    receipt_required: !!p.receipt_required,
                    currency: p.currency || 'USD',
                })));
            } catch (err: any) {
                setError(err?.message || 'Failed to load policies');
                setPolicies([]);
            }
        };
        load();
    }, []);

    const handleLimitChange = (id: string, newLimit: number) => {
        setPolicies(policies.map(p => p.id === id ? { ...p, limit_amount: newLimit } : p));
    };

    const toggleApproval = (id: string) => {
        setPolicies(policies.map(p => p.id === id ? { ...p, requires_approval: !p.requires_approval } : p));
    };

    const savePolicy = async (policy: Policy) => {
        setSavingId(policy.id);
        setError(null);
        try {
            await apiClient.put('/admin/finance/expenses/policies', {
                id: policy.id,
                category: policy.category,
                limit_amount: policy.limit_amount,
                currency: policy.currency || 'USD',
                requires_approval: policy.requires_approval,
                receipt_required: policy.receipt_required,
            });
        } catch (err: any) {
            setError(err?.message || `Failed to save ${policy.category}`);
        } finally {
            setSavingId(null);
        }
    };

    return (
        <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-white">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-2">
                        <ShieldCheckIcon className="h-8 w-8 text-green-500" />
                        Expense Policies
                    </h1>
                    <p className="text-slate-400">Configure spending limits and approval rules per category.</p>
                </div>
                <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg flex items-center gap-2 transition">
                    <PlusIcon className="h-5 w-5" />
                    Add Category
                </button>
            </div>
            {error && <div className="text-sm text-red-300 bg-red-500/10 border border-red-500/30 rounded px-3 py-2">{error}</div>}

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {policies.map((policy) => (
                    <div key={policy.id} className="bg-slate-800 p-6 rounded-xl border border-slate-700">
                        <div className="flex justify-between items-start mb-4">
                            <div className="flex items-center gap-3">
                                <div className="p-3 bg-slate-700/50 rounded-lg">
                                    <BanknotesIcon className="h-6 w-6 text-green-400" />
                                </div>
                                <div>
                                    <h3 className="font-bold text-lg">{policy.category}</h3>
                                    <p className="text-xs text-slate-400">Policy ID: POL-{policy.id.padStart(3, '0')}</p>
                                </div>
                            </div>
                            <button className="text-slate-500 hover:text-red-400 transition">
                                <TrashIcon className="h-5 w-5" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm text-slate-400 mb-1">Daily Limit (INR)</label>
                                <input
                                    type="number"
                                    value={policy.limit_amount}
                                    onChange={(e) => handleLimitChange(policy.id, parseInt(e.target.value))}
                                    className="w-full bg-slate-900 border border-slate-600 rounded px-3 py-2 text-white focus:border-green-500 focus:outline-none"
                                />
                            </div>

                            <div className="flex items-center justify-between py-2 border-t border-slate-700/50">
                                <span className="text-sm text-slate-300">Requires Manager Approval</span>
                                <button
                                    onClick={() => toggleApproval(policy.id)}
                                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${policy.requires_approval ? 'bg-green-600' : 'bg-slate-700'}`}
                                >
                                    <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition ${policy.requires_approval ? 'translate-x-6' : 'translate-x-1'}`} />
                                </button>
                            </div>

                            <div className="flex items-center justify-between py-2 border-t border-slate-700/50">
                                <span className="text-sm text-slate-300">Receipt Mandatory</span>
                                <div className={`px-2 py-1 rounded text-xs font-bold ${policy.receipt_required ? 'bg-blue-500/20 text-blue-300' : 'bg-slate-700 text-slate-400'}`}>
                                    {policy.receipt_required ? 'YES' : 'OPTIONAL'}
                                </div>
                            </div>
                            <button
                                onClick={() => savePolicy(policy)}
                                disabled={savingId === policy.id}
                                className="w-full mt-2 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white rounded px-3 py-2 text-sm font-medium"
                            >
                                {savingId === policy.id ? 'Saving...' : 'Save Policy'}
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default PolicySettings;
