import {
  ArrowLeftIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { customerV2Service } from "../../services/customerV2Service";

/**
 * AddCustomer Component
 * 
 * Re-refactored to strictly match the dark-theme admin style of the Trips module.
 * Uses standard surface tokens, grid density, and operational copy tone.
 */
export const AddCustomer: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [formData, setFormData] = useState({
    name: "",
    email: "",
    phone: "",
    address: "",
    gstin: "",
    status: "ACTIVE",
    credit_limit: "0",
    billing_cycle: "MONTHLY",
    payment_terms: "NET_30",
  });

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setError("");

    if (!formData.name) {
      setError("Company Legal Name is required");
      return;
    }

    setLoading(true);
    try {
      await customerV2Service.createCustomer({
        ...formData,
        credit_limit: parseFloat(formData.credit_limit) || 0,
      });
      navigate("/admin/customers");
    } catch (err: any) {
      const msg = err?.response?.data?.error || "Failed to add customer";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6 animate-in fade-in duration-500">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1 font-medium">
            <Link to="/admin/customers" className="hover:text-white transition-colors">
              Customers
            </Link>
            <span>/</span>
            <span className="text-primary font-bold">Add</span>
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Add New Customer</h1>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/admin/customers"
            className="px-4 py-2 bg-surface-dark hover:bg-surface-darker border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded-lg transition-colors"
          >
            Cancel
          </Link>
          <button
            onClick={() => handleSubmit()}
            disabled={loading}
            className="px-5 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 transition-colors shadow-lg shadow-primary/10 active:scale-95 disabled:opacity-50"
          >
            <span className="material-symbols-outlined text-[18px]">add</span>
            {loading ? "Creating..." : "Create Customer"}
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-danger/10 border border-danger/20 rounded-lg flex items-center gap-3 text-danger text-sm font-medium">
          <XMarkIcon className="h-5 w-5" />
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6 pb-12">
        {/* ── Customer Essentials ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden shadow-sm">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
              <span className="material-symbols-outlined text-[20px]">corporate_fare</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Customer Essentials</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Basic identification details for the corporate entity
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Entity Name */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Company Legal Name <span className="text-danger">*</span>
              </label>
              <div className="relative group">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted group-focus-within:text-primary transition-colors">
                  <span className="material-symbols-outlined text-[18px]">business</span>
                </span>
                <input
                  type="text"
                  name="name"
                  value={formData.name}
                  onChange={handleChange}
                  required
                  placeholder="e.g. Acme Logistics Pvt Ltd"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-text-muted/40 font-medium"
                />
              </div>
            </div>

            {/* Status */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Account Status
              </label>
              <div className="relative group">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted group-focus-within:text-primary transition-colors">
                  <span className="material-symbols-outlined text-[18px]">verified_user</span>
                </span>
                <select
                  name="status"
                  value={formData.status}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none transition-all font-medium"
                >
                  <option value="ACTIVE">Active</option>
                  <option value="PENDING">Pending</option>
                  <option value="SUSPENDED">Suspended</option>
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>

            {/* Phone */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Contact Phone
              </label>
              <div className="relative group">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted group-focus-within:text-primary transition-colors">
                  <span className="material-symbols-outlined text-[18px]">call</span>
                </span>
                <input
                  type="tel"
                  name="phone"
                  value={formData.phone}
                  onChange={handleChange}
                  placeholder="+91 98765 43210"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-text-muted/40 font-medium"
                />
              </div>
            </div>

            {/* Email */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Primary Email
              </label>
              <div className="relative group">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted group-focus-within:text-primary transition-colors">
                  <span className="material-symbols-outlined text-[18px]">mail</span>
                </span>
                <input
                  type="email"
                  name="email"
                  value={formData.email}
                  onChange={handleChange}
                  placeholder="ops@company-logistics.com"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-text-muted/40 font-medium"
                />
              </div>
            </div>
          </div>
        </div>

        {/* ── Legal & Compliance ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden shadow-sm">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
              <span className="material-symbols-outlined text-[20px]">policy</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Legal & Compliance</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Tax identification and registered office details
              </p>
            </div>
          </div>
          <div className="p-6 space-y-6">
            {/* GSTIN */}
            <div className="max-w-md space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                GSTIN Number
              </label>
              <div className="relative group">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted group-focus-within:text-primary transition-colors">
                  <span className="material-symbols-outlined text-[18px]">account_balance</span>
                </span>
                <input
                  type="text"
                  name="gstin"
                  value={formData.gstin}
                  onChange={handleChange}
                  placeholder="27AAAAA0000A1Z5"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white font-mono focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-text-muted/40 uppercase"
                />
              </div>
            </div>

            {/* Address */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Registered Billing Address
              </label>
              <div className="relative group">
                <span className="absolute top-3 left-3 text-text-muted group-focus-within:text-primary transition-colors">
                  <span className="material-symbols-outlined text-[18px]">location_on</span>
                </span>
                <textarea
                  name="address"
                  value={formData.address}
                  onChange={handleChange}
                  rows={4}
                  placeholder="Full legal address for tax invoicing as per GST portal..."
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-3 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-text-muted/40 resize-none font-medium"
                />
              </div>
            </div>
          </div>
        </div>

        {/* ── Commercial Policy ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden shadow-sm">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
              <span className="material-symbols-outlined text-[20px]">payments</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Commercial Policy</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Financial credit constraints and settlement logic
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Credit Limit */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Credit Limit (₹)
              </label>
              <input
                type="number"
                name="credit_limit"
                value={formData.credit_limit}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all font-bold"
              />
            </div>

            {/* Billing Cycle */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Billing Cycle
              </label>
              <div className="relative group">
                <select
                  name="billing_cycle"
                  value={formData.billing_cycle}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none transition-all font-medium"
                >
                  <option value="WEEKLY">Weekly</option>
                  <option value="BIWEEKLY">Bi-Weekly</option>
                  <option value="MONTHLY">Monthly</option>
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>

            {/* Payment Terms */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Payment Terms
              </label>
              <div className="relative group">
                <select
                  name="payment_terms"
                  value={formData.payment_terms}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none transition-all font-medium"
                >
                  <option value="IMMEDIATE">Immediate</option>
                  <option value="NET_15">Net 15 Days</option>
                  <option value="NET_30">Net 30 Days</option>
                  <option value="NET_45">Net 45 Days</option>
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
};
