
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
    const navigate = useNavigate();
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
                  row.status === 'active' ? 'bg-success/20 text-success' : 'bg-text-muted/20 text-text-muted'
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
          <h1 className="text-white text-2xl font-bold mb-1">User Management</h1>
          <p className="text-text-muted text-sm">Manage users within your organization</p>
        </div>
        
        <div className="mb-6 flex justify-end">
          <button 
            onClick={() => navigate('/admin/users/new')}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <span className="material-symbols-outlined text-[16px]">add</span>
            Add User
          </button>
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
