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
    <div className="p-8">
      <div className="mb-6 flex justify-between items-center">
        <div>
           <h1 className="text-2xl font-bold text-slate-800">Organization Management</h1>
           <p className="text-slate-500 mt-1">Manage tenant organizations and their configurations</p>
        </div>
      </div>
      
      <DataGrid<Organization>
        title="All Organizations"
        endpoint="/admin/organizations"
        columns={columns}
        refreshTrigger={refreshTrigger}
        actions={
          <button className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors shadow-sm shadow-blue-900/20">
            <PlusIcon className="h-5 w-5" />
            <span>New Organization</span>
          </button>
        }
      />
    </div>
  );
};
