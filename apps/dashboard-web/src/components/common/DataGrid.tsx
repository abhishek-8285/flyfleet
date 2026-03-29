import React, { useEffect, useState, useCallback } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

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
  filters?: FilterConfig[];
  onFiltersChange?: (filters: Record<string, any>) => void;
}

interface FilterConfig {
  field: string;
  label: string;
  type: "text" | "select" | "date" | "date-range" | "multi-select";
  options?: { value: string; label: string }[];
  placeholder?: string;
}

interface PaginationMeta {
  page: number;
  size: number;
  total: number;
  total_page: number;
}

export function DataGrid<T extends { id: string | number }>({
  endpoint,
  columns,
  title,
  actions,
  refreshTrigger,
  filters = [],
  onFiltersChange,
}: DataGridProps<T>) {
  const { uiStrings } = useUIStrings("common_datagrid");
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Pagination State
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);

  // Filter State
  const [filterValues, setFilterValues] = useState<Record<string, any>>({});
  const [showFilters, setShowFilters] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params: Record<string, any> = { page, size: pageSize };

      // Add filter parameters
      Object.entries(filterValues).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== "") {
          if (key === "dateRange") {
            if (value.from) params.start_date = value.from;
            if (value.to) params.end_date = value.to;
          } else {
            params[key] = value;
          }
        }
      });

      const response = (await apiClient.get(endpoint, { params })) as any;
      // Response structure matches our AdminMiddleware: { data: [], meta: {} }
      setData(response.data || []);
      setMeta(response.meta);
    } catch (err: any) {
      setError(err.message || "Failed to fetch data");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, endpoint, refreshTrigger, filterValues]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleFilterChange = (field: string, value: any) => {
    setFilterValues((prev) => ({
      ...prev,
      [field]: value,
    }));

    if (onFiltersChange) {
      onFiltersChange({
        ...filterValues,
        [field]: value,
      });
    }
  };

  const clearFilters = () => {
    setFilterValues({});
    if (onFiltersChange) {
      onFiltersChange({});
    }
  };

  const handleNextPage = () => {
    if (meta && page < meta.total_page) setPage((p) => p + 1);
  };

  const handlePrevPage = () => {
    if (page > 1) setPage((p) => p - 1);
  };

  return (
    <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
        <div>
          {title && (
            <h2 className="text-lg font-semibold text-white">{title}</h2>
          )}
          {meta && (
            <p className="text-xs text-text-muted mt-1">
              {(uiStrings?.total_records || "Total: {{count}} records").replace(
                "{{count}}",
                meta.total.toString(),
              )}
            </p>
          )}
        </div>
        <div className="flex space-x-3">
          {filters.length > 0 && (
            <button
              onClick={() => setShowFilters(!showFilters)}
              className="p-2 text-text-muted hover:text-primary transition-colors flex items-center gap-1 border border-border-color rounded hover:bg-border-color"
              title={uiStrings?.label_filters || "Filters"}
            >
              <span className="material-symbols-outlined text-[18px]">
                filter_list
              </span>
              <span className="text-xs">
                {uiStrings?.label_filters || "Filters"}
              </span>
              {Object.keys(filterValues).filter(
                (k) => filterValues[k] !== undefined && filterValues[k] !== "",
              ).length > 0 && (
                <span className="ml-1 bg-primary text-white text-[10px] rounded-full w-4 h-4 flex items-center justify-center">
                  {
                    Object.keys(filterValues).filter(
                      (k) =>
                        filterValues[k] !== undefined && filterValues[k] !== "",
                    ).length
                  }
                </span>
              )}
            </button>
          )}
          <button
            onClick={fetchData}
            data-testid="refresh-btn"
            className="p-2 text-text-muted hover:text-primary transition-colors"
            title={uiStrings?.btn_refresh || "Refresh"}
          >
            <span className={`material-symbols-outlined text-[18px] ${loading ? 'animate-spin' : ''}`}>
              {uiStrings?.btn_refresh || "refresh"}
            </span>
          </button>

          {/* External Actions (Create Button, etc) */}
          {actions}
        </div>
      </div>

      {/* Filters Panel */}
      {showFilters && filters.length > 0 && (
        <div className="p-4 border-b border-border-color bg-surface-darker">
          <div className="flex justify-between items-center mb-3">
            <h3 className="text-sm font-semibold text-white">
              {uiStrings?.label_filters || "Filters"}
            </h3>
            <div className="flex gap-2">
              <button
                onClick={clearFilters}
                className="text-xs text-text-muted hover:text-white transition-colors"
              >
                {uiStrings?.btn_clear_all || "Clear All"}
              </button>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filters.map((filter) => (
              <div key={filter.field} className="flex flex-col gap-1">
                <label className="text-xs text-text-muted uppercase tracking-wider font-medium">
                  {filter.label}
                </label>
                {filter.type === "text" && (
                  <input
                    type="text"
                    value={filterValues[filter.field] || ""}
                    onChange={(e) =>
                      handleFilterChange(filter.field, e.target.value)
                    }
                    placeholder={
                      filter.placeholder ||
                      (
                        uiStrings?.placeholder_filter ||
                        "Filter by {{label}}..."
                      ).replace("{{label}}", filter.label.toLowerCase())
                    }
                    className="px-3 py-2 bg-surface-dark border border-border-color rounded text-sm text-white focus:ring-2 focus:ring-primary focus:border-primary outline-none"
                  />
                )}
                {filter.type === "select" && (
                  <select
                    value={filterValues[filter.field] || ""}
                    onChange={(e) =>
                      handleFilterChange(filter.field, e.target.value)
                    }
                    className="px-3 py-2 bg-surface-dark border border-border-color rounded text-sm text-white focus:ring-2 focus:ring-primary focus:border-primary outline-none"
                  >
                    <option value="">All {filter.label}</option>
                    {filter.options?.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                )}
                {filter.type === "date" && (
                  <input
                    type="date"
                    value={filterValues[filter.field] || ""}
                    onChange={(e) =>
                      handleFilterChange(filter.field, e.target.value)
                    }
                    className="px-3 py-2 bg-surface-dark border border-border-color rounded text-sm text-white focus:ring-2 focus:ring-primary focus:border-primary outline-none"
                  />
                )}
                {filter.type === "date-range" && (
                  <div className="flex gap-2">
                    <input
                      type="date"
                      value={filterValues[filter.field]?.from || ""}
                      onChange={(e) =>
                        handleFilterChange(filter.field, {
                          ...filterValues[filter.field],
                          from: e.target.value,
                        })
                      }
                      placeholder={uiStrings?.label_from || "From"}
                      className="px-3 py-2 bg-surface-dark border border-border-color rounded text-sm text-white focus:ring-2 focus:ring-primary focus:border-primary outline-none flex-1"
                    />
                    <input
                      type="date"
                      value={filterValues[filter.field]?.to || ""}
                      onChange={(e) =>
                        handleFilterChange(filter.field, {
                          ...filterValues[filter.field],
                          to: e.target.value,
                        })
                      }
                      placeholder={uiStrings?.label_to || "To"}
                      className="px-3 py-2 bg-surface-dark border border-border-color rounded text-sm text-white focus:ring-2 focus:ring-primary focus:border-primary outline-none flex-1"
                    />
                  </div>
                )}
                {filter.type === "multi-select" && (
                  <select
                    multiple
                    value={filterValues[filter.field] || []}
                    onChange={(e) =>
                      handleFilterChange(
                        filter.field,
                        Array.from(
                          e.target.selectedOptions,
                          (option) => option.value,
                        ),
                      )
                    }
                    className="px-3 py-2 bg-surface-dark border border-border-color rounded text-sm text-white focus:ring-2 focus:ring-primary focus:border-primary outline-none h-24"
                  >
                    {filter.options?.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Loading / Error States */}
      {loading && data.length === 0 && (
        <div className="p-8 text-center text-text-muted">
          {uiStrings?.msg_loading || "Loading data..."}
        </div>
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
                  <th
                    key={idx}
                    className="px-6 py-3"
                    style={{ width: col.width }}
                  >
                    {col.header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {data.map((row) => (
                <tr
                  key={row.id}
                  className="hover:bg-surface-light/10 transition-colors"
                >
                  {columns.map((col, idx) => (
                    <td
                      key={idx}
                      className="px-6 py-4 text-sm text-text-muted whitespace-nowrap"
                    >
                      {col.render
                        ? col.render(row)
                        : (row[col.field as keyof T] as React.ReactNode)}
                    </td>
                  ))}
                </tr>
              ))}
              {data.length === 0 && (
                <tr>
                  <td
                    colSpan={columns.length}
                    className="px-6 py-8 text-center text-text-muted italic"
                  >
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
            {(uiStrings?.label_page || "Page {{current}} of {{total}}")
              .replace("{{current}}", meta.page.toString())
              .replace("{{total}}", meta.total_page.toString())}
          </div>
          <div className="flex items-center space-x-2">
            <select
              value={pageSize}
              onChange={(e) => {
                setPageSize(Number(e.target.value));
                setPage(1);
              }}
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
                <span className="material-symbols-outlined text-[16px]">
                  chevron_left
                </span>
              </button>
              <button
                onClick={handleNextPage}
                disabled={page >= meta.total_page}
                className="px-3 py-1 border-l-0 border border-border-color rounded-r-md hover:bg-border-color disabled:opacity-50 disabled:cursor-not-allowed text-text-muted hover:text-white"
              >
                <span className="material-symbols-outlined text-[16px]">
                  chevron_right
                </span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
