import { TrashIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { DataGrid } from '../../components/common/DataGrid';
import apiClient from '../../services/apiClient';

interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  status: string;
  role_id: string; // Ideally map to role name
  created_at: string;
}

export const Users: React.FC = () => {
    const [refreshTrigger, setRefreshTrigger] = useState(0);
  
    const handleDelete = async (id: string) => {
      if (window.confirm('Are you sure you want to delete this user?')) {
        try {
          await apiClient.delete(`/admin/users?id=${id}`);
          setRefreshTrigger(prev => prev + 1);
        } catch (error) {
          alert('Failed to delete user');
        }
      }
    };
  
    const columns = [
      { field: 'name', header: 'Name', width: '20%' },
      { field: 'email', header: 'Email', width: '25%' },
      { field: 'phone', header: 'Phone', width: '15%' },
      { 
          field: 'status', 
          header: 'Status', 
          render: (row: User) => (
              <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                  row.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
              }`}>
                  {row.status.toUpperCase()}
              </span>
          ) 
      },
      { field: 'created_at', header: 'Joined', render: (row: User) => new Date(row.created_at).toLocaleDateString() },
      {
        field: 'actions',
        header: '',
        width: '5%',
        render: (row: User) => (
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
        <div className="mb-6">
             <h1 className="text-2xl font-bold text-slate-800">User Management</h1>
             <p className="text-slate-500 mt-1">Manage users within your organization</p>
        </div>
        
        <DataGrid<User>
          title="Users"
          endpoint="/admin/users"
          columns={columns}
          refreshTrigger={refreshTrigger}
        />
      </div>
    );
  };
