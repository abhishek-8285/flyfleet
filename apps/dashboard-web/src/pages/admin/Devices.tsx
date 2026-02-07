import { PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { DataGrid } from '../../components/common/DataGrid';
import apiClient from '../../services/apiClient';

interface Device {
  id: string;
  serial_number: string;
  type: string;
  status: string;
  last_ping_at?: string;
  created_at: string;
}

export const Devices: React.FC = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this device?')) {
      try {
        await apiClient.delete(`/admin/devices?id=${id}`);
        setRefreshTrigger(prev => prev + 1);
      } catch (error) {
        alert('Failed to delete device');
      }
    }
  };

  const columns = [
    { field: 'serial_number', header: 'Serial Number', width: '30%' },
    { field: 'type', header: 'Type', width: '15%', render: (row: Device) => row.type.toUpperCase() },
    { 
        field: 'status', 
        header: 'Status', 
        render: (row: Device) => (
            <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                row.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
            }`}>
                {row.status.toUpperCase()}
            </span>
        ) 
    },
    { field: 'last_ping_at', header: 'Last Ping', render: (row: Device) => row.last_ping_at ? new Date(row.last_ping_at).toLocaleString() : 'Never' },
    {
      field: 'actions',
      header: '',
      width: '5%',
      render: (row: Device) => (
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
           <h1 className="text-2xl font-bold text-slate-800">Devices</h1>
           <p className="text-slate-500 mt-1">Manage GPS trackers and IoT devices</p>
        </div>
      </div>
      
      <DataGrid<Device>
        title="Fleet Devices"
        endpoint="/admin/devices"
        columns={columns}
        refreshTrigger={refreshTrigger}
        actions={
          <button className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors shadow-sm shadow-blue-900/20">
            <PlusIcon className="h-5 w-5" />
            <span>Add Device</span>
          </button>
        }
      />
    </div>
  );
};
