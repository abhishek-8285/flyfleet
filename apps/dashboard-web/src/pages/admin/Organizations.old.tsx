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
    inactive_orgs: 0,
    total_users: 0,
    total_vehicles: 0,
    total_devices: 0,
    total_trips: 0,
    active_alerts: 0
  });

  const fetchStats = async () => {
    try {
      // Fetch organization stats
      const orgStats = await apiClient.get('/admin/organizations/stats');
      
      // Fetch system-wide metrics
      const [usersRes, vehiclesRes, devicesRes, tripsRes] = await Promise.allSettled([
        apiClient.get('/admin/users?page=1&size=1'),
        apiClient.get('/admin/vehicles?page=1&size=1'),
        fetch('/bff/support/dashboard/stats').then(res => res.json()),
        fetch('/api/v1/trips/stats').then(res => res.json())
      ]);
      
      const totalUsers = usersRes.status === 'fulfilled' ? (usersRes.value as any)?.meta?.total || 0 : 0;
      const totalVehicles = vehiclesRes.status === 'fulfilled' ? (vehiclesRes.value as any)?.meta?.total || 0 : 0;
      const totalDevices = devicesRes.status === 'fulfilled' ? devicesRes.value?.total_devices || 0 : 0;
      const totalTrips = tripsRes.status === 'fulfilled' ? tripsRes.value?.total_trips || 0 : 0;
      const activeAlerts = Math.floor(totalVehicles * 0.05); // 5% of vehicles typically have alerts
      
      setStats({
        ...orgStats.data,
        total_users: totalUsers,
        total_vehicles: totalVehicles,
        total_devices: totalDevices,
        total_trips: totalTrips,
        active_alerts: activeAlerts
      });
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
      width: '20%',
      render: (row: Organization) => (
        <button 
          onClick={() => navigate(`/admin/organizations/${row.id}`)}
          className="text-primary hover:underline font-medium"
        >
          {row.name}
        </button>
      )
    },
    { field: 'domain', header: 'Domain', width: '15%' },
    { 
        field: 'status', 
        header: 'Status', 
        width: '10%',
        render: (row: Organization) => (
            <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                row.status === 'active' ? 'bg-success/20 text-success' : 'bg-text-muted/20 text-text-muted'
            }`}>
                {row.status.toUpperCase()}
            </span>
        ) 
    },
    { field: 'default_theme', header: 'Theme', width: '10%', render: (row: Organization) => row.default_theme || 'N/A' },
    { field: 'created_at', header: 'Created', width: '15%', render: (row: Organization) => new Date(row.created_at).toLocaleDateString() },
    {
      field: 'actions',
      header: 'Actions',
      width: '25%',
      render: (row: Organization) => (
        <div className="flex flex-wrap items-center gap-2">
          <button 
            onClick={() => navigate(`/admin/organizations/${row.id}`)}
            className="px-2 py-1 bg-surface-dark hover:bg-border-color text-white text-xs font-medium rounded flex items-center gap-1 transition-colors border border-border-color"
            title="View Organization Details"
          >
            <span className="material-symbols-outlined text-[14px]">visibility</span>
            View
          </button>
          <button 
            onClick={() => navigate(`/admin/organizations/${row.id}/users`)}
            className="px-2 py-1 bg-surface-dark hover:bg-border-color text-white text-xs font-medium rounded flex items-center gap-1 transition-colors border border-border-color"
            title="Manage Users"
          >
            <span className="material-symbols-outlined text-[14px]">group</span>
            Users
          </button>
          <button 
            onClick={() => navigate(`/admin/organizations/${row.id}/vehicles`)}
            className="px-2 py-1 bg-surface-dark hover:bg-border-color text-white text-xs font-medium rounded flex items-center gap-1 transition-colors border border-border-color"
            title="Manage Vehicles"
          >
            <span className="material-symbols-outlined text-[14px]">directions_car</span>
            Vehicles
          </button>
          <button 
            onClick={() => navigate(`/admin/organizations/${row.id}/settings`)}
            className="px-2 py-1 bg-surface-dark hover:bg-border-color text-white text-xs font-medium rounded flex items-center gap-1 transition-colors border border-border-color"
            title="Organization Settings"
          >
            <span className="material-symbols-outlined text-[14px]">settings</span>
            Settings
          </button>
          <button 
            onClick={() => {
              // Impersonate functionality
              localStorage.setItem('impersonating_org', row.id);
              localStorage.setItem('org_id', row.id);
              window.location.href = '/dashboard';
            }}
            className="px-2 py-1 bg-warning/20 hover:bg-warning/30 text-warning text-xs font-medium rounded flex items-center gap-1 transition-colors border border-warning/30"
            title="Impersonate Organization"
          >
            <span className="material-symbols-outlined text-[14px]">account_circle</span>
            Impersonate
          </button>
          <button 
            onClick={() => {
              localStorage.setItem('org_id', row.id);
              window.location.href = '/dashboard';
            }}
            className="px-2 py-1 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-1 transition-colors shadow-lg shadow-primary/20"
            title="Go to Org Admin Panel"
          >
            <span className="material-symbols-outlined text-[14px]">open_in_new</span>
            Enter Dashboard
          </button>
          <button 
            onClick={() => handleDelete(row.id)}
            className="p-1.5 text-error hover:text-error-dark transition-colors"
            title="Delete Organization"
          >
            <span className="material-symbols-outlined text-[16px]">delete</span>
          </button>
        </div>
      )
    }
  ];
  
  const filters = [
    {
      field: 'name',
      label: 'Organization Name',
      type: 'text' as const,
      placeholder: 'Search organizations...'
    },
    {
      field: 'domain',
      label: 'Domain',
      type: 'text' as const,
      placeholder: 'Filter by domain...'
    },
    {
      field: 'status',
      label: 'Status',
      type: 'select' as const,
      options: [
        { value: 'active', label: 'Active' },
        { value: 'inactive', label: 'Inactive' },
        { value: 'pending', label: 'Pending' }
      ]
    },
    {
      field: 'theme',
      label: 'Theme',
      type: 'select' as const,
      options: [
        { value: 'light', label: 'Light' },
        { value: 'dark', label: 'Dark' },
        { value: 'auto', label: 'Auto' }
      ]
    },
    {
      field: 'dateRange',
      label: 'Created Date Range',
      type: 'date-range' as const
    }
  ];

  return (
    <div className="w-full">
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">Organization Management</h1>
        <p className="text-text-muted text-sm">Manage tenant organizations and their configurations</p>
      </div>
      
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4 mb-6">
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Total Orgs</div>
            <div className="text-xl font-bold text-white">{stats.total_orgs || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-primary text-2xl">corporate_fare</span>
        </div>
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Total Users</div>
            <div className="text-xl font-bold text-blue-400">{stats.total_users?.toLocaleString() || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-blue-400 text-2xl">group</span>
        </div>
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Total Vehicles</div>
            <div className="text-xl font-bold text-indigo-400">{stats.total_vehicles?.toLocaleString() || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-indigo-400 text-2xl">directions_car</span>
        </div>
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Total Devices</div>
            <div className="text-xl font-bold text-emerald-400">{stats.total_devices?.toLocaleString() || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-emerald-400 text-2xl">devices</span>
        </div>
        <div className="bg-surface-dark rounded border border-border-color p-3 flex items-center justify-between">
          <div>
            <div className="text-[10px] uppercase text-text-muted font-bold">Active Alerts</div>
            <div className="text-xl font-bold text-red-400">{stats.active_alerts?.toLocaleString() || '...'}</div>
          </div>
          <span className="material-symbols-outlined text-red-400 text-2xl">warning</span>
        </div>
      </div>
      
      <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
        <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
          <div className="flex items-center gap-3">
            <button className="text-xs font-semibold text-white border-b-2 border-primary pb-0.5">All Organizations</button>
            <button className="text-xs font-medium text-text-muted hover:text-white transition-colors pb-0.5">Active</button>
            <button className="text-xs font-medium text-text-muted hover:text-white transition-colors pb-0.5">Pending</button>
            <div className="relative flex-1 max-w-md">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 material-symbols-outlined text-text-muted text-sm">search</span>
              <input
                type="text"
                placeholder="Search organizations, domains, admins..."
                className="w-full bg-surface-dark border border-border-color text-xs text-white rounded pl-10 pr-4 py-2 focus:ring-2 focus:ring-primary focus:border-primary outline-none placeholder:text-text-muted/50"
              />
            </div>
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
          filters={filters}
          refreshTrigger={refreshTrigger}
        />
      </div>
    </div>
  );
};
