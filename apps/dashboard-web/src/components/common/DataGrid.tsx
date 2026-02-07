import { ArrowPathIcon, ChevronLeftIcon, ChevronRightIcon } from '@heroicons/react/24/outline'; // Using generic icons or heroicons
import React, { useEffect, useState } from 'react';
import apiClient from '../../services/apiClient';

interface Column<T> {
  field: keyof T | string;
  header: string;
  width?: string;
  render?: (row: T) => React.ReactNode;
}

interface DataGridProps<T> {
  endpoint: string;
  columns: Column<T>[];
  title?: string;
  actions?: React.ReactNode;
  refreshTrigger?: number; // Increment to force refresh
}

interface PaginationMeta {
  page: number;
  size: number;
  total: number;
  total_page: number;
}

export function DataGrid<T extends { id: string | number }>({ endpoint, columns, title, actions, refreshTrigger }: DataGridProps<T>) {
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Pagination State
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get(endpoint, {
        params: { page, size: pageSize }
      }) as any;
      // Response structure matches our AdminMiddleware: { data: [], meta: {} }
      setData(response.data || []);
      setMeta(response.meta);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize, endpoint, refreshTrigger]);

  const handleNextPage = () => {
    if (meta && page < meta.total_page) setPage(p => p + 1);
  };

  const handlePrevPage = () => {
    if (page > 1) setPage(p => p - 1);
  };

  return (
    <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
        <div>
          {title && <h2 className="text-lg font-semibold text-slate-800">{title}</h2>}
          {meta && <p className="text-xs text-slate-500 mt-1">Total: {meta.total} records</p>}
        </div>
        <div className="flex space-x-3">
          <button 
             onClick={fetchData} 
             data-testid="refresh-btn"
             className="p-2 text-slate-400 hover:text-blue-600 transition-colors"
             title="Refresh"
          >
            <ArrowPathIcon className={`h-5 w-5 ${loading ? 'animate-spin' : ''}`} />
          </button>
          
          {/* External Actions (Create Button, etc) */}
          {actions}
        </div>
      </div>

      {/* Loading / Error States */}
      {loading && data.length === 0 && (
         <div className="p-8 text-center text-slate-500">Loading data...</div>
      )}
      
      {error && (
         <div className="p-8 text-center text-red-500 bg-red-50">{error}</div>
      )}

      {/* Table */}
      {!loading && !error && (
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider font-semibold border-b border-slate-100">
                {columns.map((col, idx) => (
                  <th key={idx} className="px-6 py-3" style={{ width: col.width }}>
                    {col.header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {data.map((row) => (
                <tr key={row.id} className="hover:bg-slate-50/80 transition-colors">
                  {columns.map((col, idx) => (
                    <td key={idx} className="px-6 py-4 text-sm text-slate-600 whitespace-nowrap">
                      {col.render ? col.render(row) : (row[col.field as keyof T] as React.ReactNode)}
                    </td>
                  ))}
                </tr>
              ))}
              {data.length === 0 && (
                 <tr>
                    <td colSpan={columns.length} className="px-6 py-8 text-center text-slate-400 italic">
                        No records found
                    </td>
                 </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination Footer */}
      {meta && (
        <div className="px-6 py-3 border-t border-slate-100 flex items-center justify-between text-sm text-slate-500 bg-slate-50/30">
          <div>
             Page {meta.page} of {meta.total_page}
          </div>
          <div className="flex items-center space-x-2">
            <select 
               value={pageSize} 
               onChange={(e) => { setPageSize(Number(e.target.value)); setPage(1); }}
               className="border border-slate-200 rounded px-2 py-1 text-xs bg-white focus:ring-2 focus:ring-blue-500 outline-none"
            >
               <option value={10}>10 / page</option>
               <option value={20}>20 / page</option>
               <option value={50}>50 / page</option>
            </select>
            
            <div className="flex rounded-md shadow-sm">
                <button
                    onClick={handlePrevPage}
                    disabled={page === 1}
                    className="px-3 py-1 border border-slate-200 rounded-l-md hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    <ChevronLeftIcon className="h-4 w-4" />
                </button>
                <button
                    onClick={handleNextPage}
                    disabled={page >= meta.total_page}
                    className="px-3 py-1 border-l-0 border border-slate-200 rounded-r-md hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    <ChevronRightIcon className="h-4 w-4" />
                </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
