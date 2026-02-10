
import React, { useState } from 'react';
import { DataGrid } from '../../components/common/DataGrid';
import apiClient from '../../services/apiClient';

interface Driver {
  id: string;
  name: string;
  phone: string;
  status: string;
  created_at: string;
}

export const Drivers: React.FC = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this driver?')) {
      try {
        await apiClient.delete(`/admin/drivers?id=${id}`);
        setRefreshTrigger(prev => prev + 1);
      } catch (error) {
        alert('Failed to delete driver');
      }
    }
  };

  const columns = [
    { field: 'name', header: 'Name', width: '30%' },
    { field: 'phone', header: 'Phone', width: '25%' },
    { 
        field: 'status', 
        header: 'Status', 
        render: (row: Driver) => (
            <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                row.status === 'active' ? 'bg-success/20 text-success' : 'bg-text-muted/20 text-text-muted'
            }`}>
                {row.status.toUpperCase()}
            </span>
        ) 
    },
    { field: 'created_at', header: 'Joined', render: (row: Driver) => new Date(row.created_at).toLocaleDateString() },
    {
      field: 'actions',
      header: '',
      width: '10%',
      render: (row: Driver) => (
        <button 
          onClick={() => handleDelete(row.id)}
          className="text-error hover:text-error-dark transition-colors"
          title="Delete"
        >
          <span className="material-symbols-outlined text-[16px]">delete</span>
        </button>
      )
    }
  ];

  return (
    <div className="w-full">
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">Drivers</h1>
        <p className="text-text-muted text-sm">Manage driver profiles and assignments</p>
      </div>
      
      <DataGrid<Driver>
        title="Fleet Drivers"
        endpoint="/admin/drivers"
        columns={columns}
        refreshTrigger={refreshTrigger}
        actions={
          <button className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20">
            <span className="material-symbols-outlined text-[16px]">add</span>
            Add Driver
          </button>
        }
      />
    </div>
  );
};
