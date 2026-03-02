import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

interface Vendor {
  id: string;
  name: string;
  contact_person: string;
  email: string;
  phone: string;
  address: string;
  category: string;
  is_active: boolean;
  rating: number;
}

const VendorDirectory: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("maintenance");
  const uiStrings = rawUiStrings || {};

  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");

  const fetchVendors = useCallback(async () => {
    setLoading(true);
    try {
      const data = (await apiClient.get("/admin/vendors")) as any;
      setVendors(data?.vendors || data?.data || []);
    } catch (error) {
      console.error("Failed to fetch vendors:", error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchVendors();
  }, [fetchVendors]);

  const filteredVendors = vendors.filter(
    (v) =>
      v.name.toLowerCase().includes(search.toLowerCase()) ||
      v.contact_person.toLowerCase().includes(search.toLowerCase()) ||
      v.category.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white">
              {uiStrings.vendors_title || "Vendor Directory"}
            </h1>
            <p className="text-sm text-text-muted mt-1">
              {uiStrings.vendors_sub || "Manage maintenance service providers"}
            </p>
          </div>
          <button
            onClick={() => navigate("/maintenance/vendors/new")}
            className="bg-primary hover:bg-primary/90 text-white text-sm font-medium py-2 px-4 rounded-lg flex items-center gap-2 shadow-[0_0_15px_rgba(19,127,236,0.3)] transition-all"
          >
            <span className="material-symbols-outlined text-lg">add</span>
            {uiStrings.btn_add_vendor || "Add Vendor"}
          </button>
        </div>
      </div>

      {/* Search */}
      <div className="px-6 py-3">
        <div className="relative max-w-md">
          <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-lg">
            search
          </span>
          <input
            type="text"
            placeholder={
              uiStrings.search_vendors_placeholder || "Search vendors..."
            }
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg pl-10 pr-4 py-2 focus:ring-1 focus:ring-primary focus:border-primary"
          />
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto px-6 pb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {loading ? (
            Array.from({ length: 6 }).map((_, i) => (
              <div
                key={i}
                className="bg-surface-dark border border-border-color rounded-xl p-6 animate-pulse"
              >
                <div className="h-4 bg-surface-darker rounded w-3/4 mb-4"></div>
                <div className="h-3 bg-surface-darker rounded w-1/2 mb-2"></div>
                <div className="h-3 bg-surface-darker rounded w-1/3"></div>
              </div>
            ))
          ) : filteredVendors.length === 0 ? (
            <div className="col-span-full flex flex-col items-center justify-center py-20 text-text-muted">
              <span className="material-symbols-outlined text-4xl mb-3">
                storefront
              </span>
              <p>{uiStrings.msg_no_vendors_found || "No vendors found"}</p>
            </div>
          ) : (
            filteredVendors.map((vendor) => (
              <div
                key={vendor.id}
                className="bg-surface-dark border border-border-color rounded-xl p-5 hover:border-primary/50 transition-colors cursor-pointer group"
                onClick={() => navigate(`/maintenance/vendors/${vendor.id}`)}
              >
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <h3 className="text-white font-medium text-lg">
                      {vendor.name}
                    </h3>
                    <p className="text-text-muted text-sm capitalize">
                      {vendor.category}
                    </p>
                  </div>
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded ${
                      vendor.is_active
                        ? "bg-emerald-500/20 text-emerald-400"
                        : "bg-red-500/20 text-red-400"
                    }`}
                  >
                    {vendor.is_active
                      ? uiStrings.label_active || "Active"
                      : uiStrings.label_inactive || "Inactive"}
                  </span>
                </div>

                <div className="space-y-2 mt-4">
                  <div className="flex items-center gap-2 text-sm text-text-muted">
                    <span className="material-symbols-outlined text-base">
                      person
                    </span>
                    {vendor.contact_person}
                  </div>
                  <div className="flex items-center gap-2 text-sm text-text-muted group-hover:text-primary transition-colors">
                    <span className="material-symbols-outlined text-base">
                      call
                    </span>
                    {vendor.phone}
                  </div>
                  <div className="flex items-center gap-2 text-sm text-text-muted group-hover:text-primary transition-colors">
                    <span className="material-symbols-outlined text-base">
                      mail
                    </span>
                    {vendor.email}
                  </div>
                </div>

                <div className="mt-4 pt-4 border-t border-border-color flex justify-between items-center">
                  <div className="flex items-center text-amber-400">
                    <span className="material-symbols-outlined text-lg">
                      star
                    </span>
                    <span className="text-white ml-1 font-medium">
                      {vendor.rating.toFixed(1)}
                    </span>
                  </div>
                  <button className="text-text-muted hover:text-white transition-colors">
                    <span className="material-symbols-outlined">
                      arrow_forward
                    </span>
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default VendorDirectory;
