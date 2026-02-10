import { PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { DataGrid } from '../../components/common/DataGrid';
import apiClient from '../../services/apiClient';

interface Organization {
  id: string;
  name: string;
  domain: string;
  status: string;
  default_theme: string;
  created_at: string;
}

export const Organizations: React.FC = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this organization?')) {
      try {
        await apiClient.delete(`/admin/organizations?id=${id}`);
        setRefreshTrigger(prev => prev + 1);
      } catch (error) {
        alert('Failed to delete organization');
      }
    }
  };

  const columns = [
    { field: 'name', header: 'Name', width: '25%' },
    { field: 'domain', header: 'Domain', width: '20%' },
    { 
        field: 'status', 
        header: 'Status', 
        render: (row: Organization) => (
            <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                row.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
            }`}>
                {row.status.toUpperCase()}
            </span>
        ) 
    },
    { field: 'created_at', header: 'Created', render: (row: Organization) => new Date(row.created_at).toLocaleDateString() },
    {
      field: 'actions',
      header: '',
      width: '10%',
      render: (row: Organization) => (
        <button 
          onClick={() => handleDelete(row.id)}
          className="text-red-500 hover:text-red-700 transition-colors"
          title="Delete"
        >
          <TrashIcon className="h-4 w-4" />
        </button>
      )
    }
  ];

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      {/* Sub-Header - Optimized for AdminLayout */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-slate-200 bg-white flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-slate-900 text-lg font-bold">Organization Management</h2>
            <p className="text-[11px] text-slate-500 mt-0.5">Manage tenant organizations and their configurations</p>
          </div>
          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-emerald-100 text-emerald-700 border border-emerald-200">
            Platform Online
          </span>
        </div>
        <div className="flex items-center gap-3">
          <button 
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-blue-600/20"
          >
            <PlusIcon className="h-4 w-4" />
            <span>New Organization</span>
          </button>
        </div>
      </header>
      
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <DataGrid<Organization>
          title="Active Tenants"
          endpoint="/admin/organizations"
          columns={columns}
          refreshTrigger={refreshTrigger}
        />
      </div>
    </div>
  );
};
