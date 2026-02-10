import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
  const navigate = useNavigate();
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [stats, setStats] = useState({
    total_orgs: 0,
    active_orgs: 0,
    pending_orgs: 0,
    inactive_orgs: 0
  });

  const fetchStats = async () => {
    try {
      const data = await apiClient.get('/admin/organizations/stats');
      setStats(data.data);
    } catch (error) {
       console.error("Failed to fetch organization stats", error);
    }
  };

  useEffect(() => {
    fetchStats();
  }, [refreshTrigger]);

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
    { 
      field: 'name', 
      header: 'Name', 
      width: '25%',
      render: (row: Organization) => (
        <button 
          onClick={() => navigate(`/admin/organizations/${row.id}`)}
          className="text-primary hover:underline font-medium"
        >
          {row.name}
        </button>
      )
    },
    { field: 'domain', header: 'Domain', width: '20%' },
    { 
        field: 'status', 
        header: 'Status', 
        render: (row: Organization) => (
            <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                row.status === 'active' ? 'bg-success/20 text-success' : 'bg-text-muted/20 text-text-muted'
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
        <h1 className="text-white text-2xl font-bold mb-1">Organization Management</h1>
        <p className="text-text-muted text-sm">Manage tenant organizations and their configurations</p>
      </div>
      
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Total Orgs</div>
            <div className="text-xl font-bold text-white">{stats.total_orgs || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-primary text-2xl">corporate_fare</span>
        </div>
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Active</div>
            <div className="text-xl font-bold text-success">{stats.active_orgs || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-success text-2xl">check_circle</span>
        </div>
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Pending</div>
            <div className="text-xl font-bold text-warning">{stats.pending_orgs || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-warning text-2xl">schedule</span>
        </div>
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Inactive</div>
            <div className="text-xl font-bold text-text-muted">{stats.inactive_orgs || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-text-muted text-2xl">block</span>
        </div>
      </div>
      
      <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
        <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
          <div className="flex items-center gap-3">
            <button className="text-xs font-semibold text-white border-b-2 border-primary pb-0.5">All Organizations</button>
            <button className="text-xs font-medium text-text-muted hover:text-white transition-colors pb-0.5">Active</button>
            <button className="text-xs font-medium text-text-muted hover:text-white transition-colors pb-0.5">Pending</button>
          </div>
          <div className="flex items-center gap-2">
            <button className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white flex items-center gap-1 text-[10px] border border-border-color">
              <span className="material-symbols-outlined text-[14px]">filter_list</span>
              Filter
            </button>
            <button className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white">
              <span className="material-symbols-outlined text-[18px]">refresh</span>
            </button>
            <button 
              onClick={() => navigate('/admin/organizations/new')}
              className="px-3 py-1.5 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-1 transition-colors shadow-lg shadow-primary/20"
            >
              <span className="material-symbols-outlined text-[16px]">add</span>
              New Organization
            </button>
          </div>
        </div>
      </div>
      
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
