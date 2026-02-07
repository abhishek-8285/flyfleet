import { PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { DataGrid } from '../../components/common/DataGrid';
import apiClient from '../../services/apiClient';

interface Vehicle {
  id: string;
  name: string;
  license_plate: string;
  status: string;
  created_at: string;
}

export const Vehicles: React.FC = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this vehicle?')) {
      try {
        await apiClient.delete(`/admin/vehicles?id=${id}`);
        setRefreshTrigger(prev => prev + 1);
      } catch (error) {
        alert('Failed to delete vehicle');
      }
    }
  };

  const columns = [
    { field: 'name', header: 'Vehicle Name', width: '30%' },
    { field: 'license_plate', header: 'License Plate', width: '30%' },
    { 
        field: 'status', 
        header: 'Status', 
        render: (row: Vehicle) => (
            <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                row.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
            }`}>
                {row.status.toUpperCase()}
            </span>
        ) 
    },
    { field: 'created_at', header: 'Added On', render: (row: Vehicle) => new Date(row.created_at).toLocaleDateString() },
    {
      field: 'actions',
      header: '',
      width: '10%',
      render: (row: Vehicle) => (
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
           <h1 className="text-2xl font-bold text-slate-800">Vehicle Management</h1>
           <p className="text-slate-500 mt-1">Manage fleet vehicles and asset status</p>
        </div>
      </div>
      
      <DataGrid<Vehicle>
        title="Fleet Vehicles"
        endpoint="/admin/vehicles"
        columns={columns}
        refreshTrigger={refreshTrigger}
        actions={
          <button className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors shadow-sm shadow-blue-900/20">
            <PlusIcon className="h-5 w-5" />
            <span>Add Vehicle</span>
          </button>
        }
      />
    </div>
  );
};
