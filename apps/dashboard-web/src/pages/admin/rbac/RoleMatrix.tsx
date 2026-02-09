import { CheckIcon, XMarkIcon } from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

interface Permission {
  id: string;
  module: string;
  action: string;
  description: string;
}

interface Role {
    id: string;
    name: string;
    is_system: boolean;
}

export const RoleMatrix: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const [role, setRole] = useState<Role | null>(null);
    const [allPermissions, setAllPermissions] = useState<Permission[]>([]);
    const [rolePermissions, setRolePermissions] = useState<string[]>([]); // Array of IDs
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    const fetchData = async () => {
        try {
            const [roleRes, allPermsRes, rolePermsRes] = await Promise.all([
                apiClient.get(`/rbac/roles/${id}`) as any,
                apiClient.get('/rbac/permissions') as any,
                apiClient.get(`/rbac/roles/${id}/permissions`) as any
            ]);
            setRole(roleRes.data);
            setAllPermissions(allPermsRes.data);
            setRolePermissions(rolePermsRes.data.map((p: any) => p.id));
        } catch (error) {
            console.error('Failed to fetch matrix data', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (id) fetchData();
    }, [id]);

    const handleToggle = (permId: string) => {
        if (role?.is_system) return;
        setRolePermissions(prev => 
            prev.includes(permId) ? prev.filter(p => p !== permId) : [...prev, permId]
        );
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            await apiClient.put(`/rbac/roles/${id}/permissions`, {
                permission_ids: rolePermissions
            });
            alert('Permissions updated successfully');
        } catch (error) {
            alert('Failed to update permissions');
        } finally {
            setSaving(false);
        }
    };

    if (loading) return <div className="p-8 text-center text-slate-500">Loading permission matrix...</div>;

    // Group permissions by module
    const modules = Array.from(new Set(allPermissions.map(p => p.module)));

    return (
        <div className="p-8">
            <div className="mb-6 flex justify-between items-end">
                <div>
                    <h1 className="text-2xl font-bold text-slate-800">Permission Matrix: {role?.name}</h1>
                    <p className="text-slate-500 mt-1">Configure granular access controls for this role</p>
                </div>
                {!role?.is_system && (
                    <button 
                        className={`bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg text-sm font-semibold transition-all shadow-sm ${saving ? 'opacity-50 cursor-not-allowed' : ''}`}
                        onClick={handleSave}
                        disabled={saving}
                    >
                        {saving ? 'Saving...' : 'Save Changes'}
                    </button>
                )}
            </div>

            {role?.is_system && (
                <div className="mb-6 bg-amber-50 border border-amber-200 p-4 rounded-lg flex items-center text-amber-700 text-sm">
                    <ShieldCheckIcon className="h-5 w-5 mr-3" />
                    System roles have locked permissions that cannot be modified.
                </div>
            )}

            <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider font-semibold border-b border-slate-100">
                            <th className="px-6 py-3 w-1/3">Module / Permission</th>
                            <th className="px-6 py-3 text-center">Status</th>
                            <th className="px-6 py-3">Description</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                        {modules.map(module => (
                            <React.Fragment key={module}>
                                <tr className="bg-slate-50/50">
                                    <td colSpan={3} className="px-6 py-2 text-xs font-bold text-slate-400 uppercase tracking-widest bg-slate-100/30">
                                        {module}
                                    </td>
                                </tr>
                                {allPermissions.filter(p => p.module === module).map(perm => (
                                    <tr 
                                        key={perm.id} 
                                        className={`hover:bg-slate-50/80 transition-colors cursor-pointer ${role?.is_system ? 'cursor-default' : ''}`}
                                        onClick={() => handleToggle(perm.id)}
                                    >
                                        <td className="px-6 py-4 text-sm font-medium text-slate-700">
                                            {perm.action}
                                        </td>
                                        <td className="px-6 py-4 text-center">
                                            <div className={`mx-auto h-6 w-6 rounded flex items-center justify-center transition-all ${
                                                rolePermissions.includes(perm.id) 
                                                    ? 'bg-blue-100 text-blue-600' 
                                                    : 'bg-slate-100 text-slate-300'
                                            }`}>
                                                {rolePermissions.includes(perm.id) ? (
                                                    <CheckIcon className="h-4 w-4 stroke-[3]" />
                                                ) : (
                                                    <XMarkIcon className="h-4 w-4" />
                                                )}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 text-sm text-slate-500 italic">
                                            {perm.description || `Grants access to ${perm.action} in ${perm.module}`}
                                        </td>
                                    </tr>
                                ))}
                            </React.Fragment>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

import { ShieldCheckIcon } from '@heroicons/react/24/outline';
