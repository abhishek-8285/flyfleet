import { PencilSquareIcon, ShieldCheckIcon, TrashIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { DataGrid } from '../../../components/common/DataGrid';
import apiClient from '../../../services/apiClient';

import { useNavigate } from 'react-router-dom';

interface Role {
  id: string;
  name: string;
  description: string;
  is_system: boolean;
  created_at: string;
}

export const Roles: React.FC = () => {
    const navigate = useNavigate();
    const [refreshTrigger, setRefreshTrigger] = useState(0);
  
    const handleDelete = async (id: string, isSystem: boolean) => {
      if (isSystem) {
        alert('System roles cannot be deleted');
        return;
      }
      if (window.confirm('Are you sure you want to delete this role?')) {
        try {
          await apiClient.delete(`/rbac/roles/${id}`);
          setRefreshTrigger(prev => prev + 1);
        } catch (error) {
          alert('Failed to delete role');
        }
      }
    };
  
    const columns = [
      { field: 'name', header: 'Role Name', width: '25%' },
      { field: 'description', header: 'Description', width: '40%' },
      { 
          field: 'is_system', 
          header: 'Type', 
          render: (row: Role) => (
              <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                  row.is_system ? 'bg-amber-100 text-amber-700' : 'bg-blue-100 text-blue-700'
              }`}>
                  {row.is_system ? 'SYSTEM' : 'CUSTOM'}
              </span>
          ) 
      },
      { field: 'created_at', header: 'Created At', render: (row: Role) => new Date(row.created_at).toLocaleDateString() },
      {
        field: 'actions',
        header: '',
        width: '10%',
        render: (row: Role) => (
          <div className="flex space-x-3">
            <button 
              className="text-blue-500 hover:text-blue-700 transition-colors"
              title="Edit Permissions"
              onClick={() => navigate(`/admin/rbac/roles/${row.id}/permissions`)}
            >
              <ShieldCheckIcon className="h-4 w-4" />
            </button>
            {!row.is_system && (
                <>
                    <button 
                        className="text-slate-500 hover:text-slate-700 transition-colors"
                        title="Edit Details"
                    >
                        <PencilSquareIcon className="h-4 w-4" />
                    </button>
                    <button 
                        onClick={() => handleDelete(row.id, row.is_system)}
                        className="text-red-500 hover:text-red-700 transition-colors"
                        title="Delete"
                    >
                        <TrashIcon className="h-4 w-4" />
                    </button>
                </>
            )}
          </div>
        )
      }
    ];
  
    return (
      <div className="p-8">
        <div className="mb-6 flex justify-between items-end">
            <div>
                <h1 className="text-2xl font-bold text-slate-800">Role Management</h1>
                <p className="text-slate-500 mt-1">Define and manage access levels for your organization</p>
            </div>
            <button 
                className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-semibold transition-all shadow-sm flex items-center"
                onClick={() => { /* Open create modal */ }}
            >
                Create New Role
            </button>
        </div>
        
        <DataGrid<Role>
          title="Roles"
          endpoint="/rbac/roles"
          columns={columns}
          refreshTrigger={refreshTrigger}
        />
      </div>
    );
};
