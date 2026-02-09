import { ClockIcon, ShieldExclamationIcon } from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { DataGrid } from '../../../components/common/DataGrid';

interface AuditLog {
  id: string;
  user_id: string;
  action: string;
  resource: string;
  status: string;
  ip_address: string;
  created_at: string;
  details: string;
}

export const SecurityLogs: React.FC = () => {
    const [refreshTrigger, setRefreshTrigger] = useState(0);
  
    const columns = [
      { 
          field: 'action', 
          header: 'Action', 
          width: '20%',
          render: (row: AuditLog) => (
              <span className="font-medium text-slate-800">{row.action}</span>
          )
      },
      { field: 'resource', header: 'Resource', width: '15%' },
      { 
          field: 'status', 
          header: 'Status', 
          render: (row: AuditLog) => (
              <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                  row.status === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
              }`}>
                  {row.status.toUpperCase()}
              </span>
          ) 
      },
      { field: 'ip_address', header: 'IP Address', width: '15%' },
      { 
          field: 'created_at', 
          header: 'Time', 
          render: (row: AuditLog) => (
              <div className="flex items-center text-slate-500">
                  <ClockIcon className="h-4 w-4 mr-1" />
                  {new Date(row.created_at).toLocaleString()}
              </div>
          ) 
      },
      {
        field: 'details',
        header: 'Metadata',
        width: '20%',
        render: (row: AuditLog) => (
          <div className="max-w-xs truncate text-xs text-slate-400 font-mono" title={row.details}>
            {row.details}
          </div>
        )
      }
    ];
  
    return (
      <div className="p-8">
        <div className="mb-6 flex justify-between items-end">
            <div>
                <h1 className="text-2xl font-bold text-slate-800 flex items-center">
                    <ShieldExclamationIcon className="h-8 w-8 text-blue-600 mr-3" />
                    Security Audit Logs
                </h1>
                <p className="text-slate-500 mt-1">Immutable record of all security-sensitive operations</p>
            </div>
            <div className="flex space-x-2">
                <button 
                    className="bg-white border border-slate-200 text-slate-600 px-4 py-2 rounded-lg text-sm font-semibold hover:bg-slate-50 transition-all shadow-sm"
                    onClick={() => { /* Export logic */ }}
                >
                    Export CSV
                </button>
            </div>
        </div>
        
        <DataGrid<AuditLog>
          title="Audit Trail"
          endpoint="/admin/audit/logs"
          columns={columns}
          refreshTrigger={refreshTrigger}
        />
      </div>
    );
};
