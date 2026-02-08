import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface InventoryPart {
  id: string;
  part_number: string;
  name: string;
  description: string;
  stock_location: string;
  quantity_available: number;
  reorder_level: number;
  unit_cost: number;
}

const Inventory: React.FC = () => {
  const navigate = useNavigate();
  const [parts, setParts] = useState<InventoryPart[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [showLowStock, setShowLowStock] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);

  const fetchParts = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      params.set('page_size', '100');
      if (search) params.set('search', search);
      if (showLowStock) params.set('low_stock_only', 'true');

      const response = await fetch(`/api/v1/admin/inventory?${params}`);
      if (response.ok) {
        const data = await response.json();
        setParts(data.parts || []);
      }
    } catch (error) {
      console.error('Failed to fetch inventory:', error);
    } finally {
      setLoading(false);
    }
  }, [search, showLowStock]);

  useEffect(() => {
    fetchParts();
  }, [fetchParts]);

  const lowStockCount = parts.filter(p => p.quantity_available <= p.reorder_level).length;

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white">Inventory</h1>
            <p className="text-sm text-text-muted mt-1">Manage maintenance parts inventory</p>
          </div>
          <button
            onClick={() => setShowAddModal(true)}
            className="bg-primary hover:bg-primary/90 text-white text-sm font-medium py-2 px-4 rounded-lg flex items-center gap-2 shadow-[0_0_15px_rgba(19,127,236,0.3)] transition-all"
          >
            <span className="material-symbols-outlined text-lg">add</span>
            Add Part
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="px-6 py-4 grid grid-cols-1 md:grid-cols-3 gap-4 flex-shrink-0">
        <div className="bg-surface-dark border border-border-color rounded-xl p-4">
          <div className="flex items-center justify-between">
            <span className="material-symbols-outlined text-blue-400 text-2xl">inventory_2</span>
          </div>
          <p className="text-3xl font-bold text-white mt-3">{parts.length}</p>
          <p className="text-sm text-text-muted">Total Parts</p>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-xl p-4">
          <div className="flex items-center justify-between">
            <span className="material-symbols-outlined text-amber-400 text-2xl">warning</span>
            {lowStockCount > 0 && (
              <span className="text-xs text-amber-400 bg-amber-500/20 px-2 py-0.5 rounded">Alert</span>
            )}
          </div>
          <p className="text-3xl font-bold text-white mt-3">{lowStockCount}</p>
          <p className="text-sm text-text-muted">Low Stock Items</p>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-xl p-4">
          <div className="flex items-center justify-between">
            <span className="material-symbols-outlined text-emerald-400 text-2xl">attach_money</span>
          </div>
          <p className="text-3xl font-bold text-white mt-3">
            ${parts.reduce((sum, p) => sum + p.quantity_available * p.unit_cost, 0).toLocaleString()}
          </p>
          <p className="text-sm text-text-muted">Total Inventory Value</p>
        </div>
      </div>

      {/* Filters */}
      <div className="px-6 py-3 flex flex-wrap items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-lg">search</span>
          <input
            type="text"
            placeholder="Search parts..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-10 pr-4 py-2 focus:ring-1 focus:ring-primary focus:border-primary"
          />
        </div>
        <button
          onClick={() => setShowLowStock(!showLowStock)}
          className={`px-4 py-2 text-xs font-medium rounded-lg border transition-all ${
            showLowStock
              ? 'bg-amber-500/20 text-amber-400 border-amber-500/30'
              : 'bg-surface-dark border-border-color text-text-muted hover:border-primary/50'
          }`}
        >
          Low Stock Only
        </button>
      </div>

      {/* Table */}
      <div className="flex-1 overflow-y-auto px-6 pb-6">
        <div className="bg-surface-dark border border-border-color rounded-xl overflow-hidden">
          {loading ? (
            <div className="flex items-center justify-center py-20">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : parts.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-text-muted">
              <span className="material-symbols-outlined text-4xl mb-3">inventory_2</span>
              <p>No parts found</p>
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-surface-darker">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase">Part Number</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase">Name</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase">Location</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase">In Stock</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase">Unit Cost</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase">Total Value</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-color">
                {parts.map((part) => {
                  const isLowStock = part.quantity_available <= part.reorder_level;
                  return (
                    <tr key={part.id} className="hover:bg-surface-darker/50 transition-colors">
                      <td className="px-4 py-4 text-sm font-mono text-white">{part.part_number}</td>
                      <td className="px-4 py-4">
                        <div className="text-sm font-medium text-white">{part.name}</div>
                        <div className="text-xs text-text-muted line-clamp-1">{part.description}</div>
                      </td>
                      <td className="px-4 py-4 text-sm text-text-muted">{part.stock_location || '-'}</td>
                      <td className="px-4 py-4">
                        <span className={`px-2 py-1 text-xs font-medium rounded ${
                          isLowStock ? 'bg-red-500/20 text-red-400' : 'bg-emerald-500/20 text-emerald-400'
                        }`}>
                          {part.quantity_available}
                        </span>
                        {isLowStock && (
                          <span className="text-xs text-text-muted ml-2">(min: {part.reorder_level})</span>
                        )}
                      </td>
                      <td className="px-4 py-4 text-sm text-text-muted text-right">${part.unit_cost.toFixed(2)}</td>
                      <td className="px-4 py-4 text-sm font-medium text-white text-right">
                        ${(part.quantity_available * part.unit_cost).toFixed(2)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
};

export default Inventory;
