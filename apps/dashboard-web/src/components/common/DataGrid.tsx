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
    <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
        <div>
          {title && <h2 className="text-lg font-semibold text-white">{title}</h2>}
          {meta && <p className="text-xs text-text-muted mt-1">Total: {meta.total} records</p>}
        </div>
        <div className="flex space-x-3">
          <button 
             onClick={fetchData} 
             data-testid="refresh-btn"
             className="p-2 text-text-muted hover:text-primary transition-colors"
             title="Refresh"
          >
            <span className="material-symbols-outlined text-[18px] ${loading ? 'animate-spin' : ''}">refresh</span>
          </button>
          
          {/* External Actions (Create Button, etc) */}
          {actions}
        </div>
      </div>

      {/* Loading / Error States */}
      {loading && data.length === 0 && (
         <div className="p-8 text-center text-text-muted">Loading data...</div>
      )}
      
      {error && (
         <div className="p-8 text-center text-error bg-error/10">{error}</div>
      )}

      {/* Table */}
      {!loading && !error && (
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-surface-darker text-text-muted text-xs uppercase tracking-wider font-semibold border-b border-border-color">
                {columns.map((col, idx) => (
                  <th key={idx} className="px-6 py-3" style={{ width: col.width }}>
                    {col.header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {data.map((row) => (
                <tr key={row.id} className="hover:bg-surface-light/10 transition-colors">
                  {columns.map((col, idx) => (
                    <td key={idx} className="px-6 py-4 text-sm text-text-muted whitespace-nowrap">
                      {col.render ? col.render(row) : (row[col.field as keyof T] as React.ReactNode)}
                    </td>
                  ))}
                </tr>
              ))}
              {data.length === 0 && (
                 <tr>
                    <td colSpan={columns.length} className="px-6 py-8 text-center text-text-muted italic">
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
        <div className="px-6 py-3 border-t border-border-color flex items-center justify-between text-sm text-text-muted bg-surface-darker">
          <div>
             Page {meta.page} of {meta.total_page}
          </div>
          <div className="flex items-center space-x-2">
            <select 
               value={pageSize} 
               onChange={(e) => { setPageSize(Number(e.target.value)); setPage(1); }}
               className="border border-border-color rounded px-2 py-1 text-xs bg-surface-light focus:ring-2 focus:ring-primary outline-none text-text-muted"
            >
               <option value={10}>10 / page</option>
               <option value={20}>20 / page</option>
               <option value={50}>50 / page</option>
            </select>
            
            <div className="flex rounded-md">
                <button
                    onClick={handlePrevPage}
                    disabled={page === 1}
                    className="px-3 py-1 border border-border-color rounded-l-md hover:bg-border-color disabled:opacity-50 disabled:cursor-not-allowed text-text-muted hover:text-white"
                >
                    <span className="material-symbols-outlined text-[16px]">chevron_left</span>
                </button>
                <button
                    onClick={handleNextPage}
                    disabled={page >= meta.total_page}
                    className="px-3 py-1 border-l-0 border border-border-color rounded-r-md hover:bg-border-color disabled:opacity-50 disabled:cursor-not-allowed text-text-muted hover:text-white"
                >
                    <span className="material-symbols-outlined text-[16px]">chevron_right</span>
                </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
