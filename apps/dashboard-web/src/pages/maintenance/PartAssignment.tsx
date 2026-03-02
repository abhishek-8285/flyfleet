import React, { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

interface InventoryPart {
  id: string;
  part_number: string;
  name: string;
  description: string;
  quantity_available: number;
  unit_cost: number;
}

const PartAssignment: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("maintenance");
  const uiStrings = rawUiStrings || {};

  const [parts, setParts] = useState<InventoryPart[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [selectedParts, setSelectedParts] = useState<Record<string, number>>(
    {},
  );
  const [submitting, setSubmitting] = useState(false);

  const fetchParts = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      params.set("available_only", "true");
      params.set("page_size", "100");
      if (search) params.set("search", search);

      const response = await fetch(`/api/v1/admin/inventory?${params}`);
      if (response.ok) {
        const data = await response.json();
        setParts(data.parts || []);
      }
    } catch (error) {
      console.error("Failed to fetch parts:", error);
    } finally {
      setLoading(false);
    }
  }, [search]);

  useEffect(() => {
    fetchParts();
  }, [fetchParts]);

  const addPart = (partId: string) => {
    setSelectedParts((prev) => ({
      ...prev,
      [partId]: (prev[partId] || 0) + 1,
    }));
  };

  const removePart = (partId: string) => {
    setSelectedParts((prev) => {
      const newValue = (prev[partId] || 0) - 1;
      if (newValue <= 0) {
        const { [partId]: _, ...rest } = prev;
        return rest;
      }
      return { ...prev, [partId]: newValue };
    });
  };

  const assignParts = async () => {
    setSubmitting(true);
    try {
      for (const [partId, quantity] of Object.entries(selectedParts)) {
        await fetch(`/api/v1/admin/work-orders/${id}/parts`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ part_id: partId, quantity }),
        });
      }
      navigate(`/maintenance/work-orders/${id}`);
    } catch (error) {
      console.error("Failed to assign parts:", error);
    } finally {
      setSubmitting(false);
    }
  };

  const selectedCount = Object.values(selectedParts).reduce((a, b) => a + b, 0);

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate(`/maintenance/work-orders/${id}`)}
              className="p-2 hover:bg-surface-dark rounded-lg transition-colors"
            >
              <span className="material-symbols-outlined text-text-muted">
                arrow_back
              </span>
            </button>
            <div>
              <h1 className="text-2xl font-bold text-white">
                {uiStrings.assign_parts_title || "Add Parts"}
              </h1>
              <p className="text-sm text-text-muted mt-1">
                {uiStrings.assign_parts_sub ||
                  "Select parts to assign to this work order"}
              </p>
            </div>
          </div>
          <button
            onClick={assignParts}
            disabled={selectedCount === 0 || submitting}
            className="px-4 py-2 text-sm font-medium bg-primary hover:bg-primary/90 text-white rounded-lg shadow-[0_0_15px_rgba(19,127,236,0.3)] transition-all disabled:opacity-50 flex items-center gap-2"
          >
            {submitting && (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
            )}
            Assign {selectedCount} Part{selectedCount !== 1 ? "s" : ""}
          </button>
        </div>
      </div>

      {/* Search */}
      <div className="px-6 py-3 flex-shrink-0">
        <div className="relative max-w-md">
          <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-lg">
            search
          </span>
          <input
            type="text"
            placeholder={
              uiStrings.placeholder_search_parts || "Search parts..."
            }
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-10 pr-4 py-2 focus:ring-1 focus:ring-primary focus:border-primary"
          />
        </div>
      </div>

      {/* Parts Grid */}
      <div className="flex-1 overflow-y-auto px-6 pb-6">
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : parts.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-text-muted">
            <span className="material-symbols-outlined text-4xl mb-3">
              inventory_2
            </span>
            <p>{uiStrings.msg_no_parts_available || "No parts available"}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {parts.map((part) => {
              const qty = selectedParts[part.id] || 0;
              const isSelected = qty > 0;
              return (
                <div
                  key={part.id}
                  className={`bg-surface-dark border rounded-xl p-4 transition-all ${
                    isSelected
                      ? "border-primary shadow-[0_0_10px_rgba(19,127,236,0.2)]"
                      : "border-border-color"
                  }`}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h3 className="text-sm font-medium text-white">
                        {part.name}
                      </h3>
                      <p className="text-xs text-text-muted">
                        {part.part_number}
                      </p>
                    </div>
                    <span
                      className={`text-xs px-2 py-0.5 rounded ${
                        part.quantity_available <= 5
                          ? "bg-red-500/20 text-red-400"
                          : "bg-emerald-500/20 text-emerald-400"
                      }`}
                    >
                      {(uiStrings.label_in_stock || "{count} in stock").replace(
                        "{count}",
                        part.quantity_available.toString(),
                      )}
                    </span>
                  </div>
                  <p className="text-xs text-text-muted line-clamp-2 mb-3">
                    {part.description || "No description"}
                  </p>
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-bold text-primary">
                      ${part.unit_cost.toFixed(2)}
                    </span>
                    <div className="flex items-center gap-2">
                      {isSelected ? (
                        <>
                          <button
                            onClick={() => removePart(part.id)}
                            className="w-7 h-7 rounded bg-surface-darker border border-border-color flex items-center justify-center text-text-muted hover:text-white"
                          >
                            <span className="material-symbols-outlined text-sm">
                              remove
                            </span>
                          </button>
                          <span className="text-sm font-medium text-white w-6 text-center">
                            {qty}
                          </span>
                          <button
                            onClick={() => addPart(part.id)}
                            disabled={qty >= part.quantity_available}
                            className="w-7 h-7 rounded bg-primary flex items-center justify-center text-white disabled:opacity-50"
                          >
                            <span className="material-symbols-outlined text-sm">
                              add
                            </span>
                          </button>
                        </>
                      ) : (
                        <button
                          onClick={() => addPart(part.id)}
                          className="text-xs text-primary hover:text-primary/80 font-medium"
                        >
                          {uiStrings.btn_add_short || "+ Add"}
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default PartAssignment;
