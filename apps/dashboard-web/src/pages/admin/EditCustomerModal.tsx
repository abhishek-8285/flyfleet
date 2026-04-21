import React, { useState, useEffect } from "react";
import Modal from "../../components/common/Modal";
import { customerV2Service, Customer } from "../../services/customerV2Service";
import { Building2, Mail, Phone, MapPin, Hash, CheckCircle2 } from "lucide-react";

interface EditCustomerModalProps {
  isOpen: boolean;
  onClose: () => void;
  customer: Customer;
  onSuccess: (updatedCustomer: Customer) => void;
}

export const EditCustomerModal: React.FC<EditCustomerModalProps> = ({
  isOpen,
  onClose,
  customer,
  onSuccess,
}) => {
  const [formData, setFormData] = useState({
    name: customer.name || "",
    email: customer.email || "",
    phone: customer.phone || "",
    address: customer.address || "",
    gstin: customer.gstin || "",
    status: customer.status || "ACTIVE",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (customer) {
      setFormData({
        name: customer.name || "",
        email: customer.email || "",
        phone: customer.phone || "",
        address: customer.address || "",
        gstin: customer.gstin || "",
        status: customer.status || "ACTIVE",
      });
    }
  }, [customer]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const updated = await customerV2Service.updateCustomer(customer.id, formData);
      onSuccess(updated);
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to update customer");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Edit Customer Profile">
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Organization Name */}
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">
              Organization Name
            </label>
            <div className="relative group">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-blue-400 transition-colors">
                <Building2 className="w-4 h-4" />
              </div>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
                className="w-full bg-slate-900/50 border border-white/5 rounded-xl py-3 pl-11 pr-4 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500/50 transition-all font-medium"
                placeholder="e.g. Acme Corp"
              />
            </div>
          </div>

          {/* Status Selection */}
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">
              Account Status
            </label>
            <select
              name="status"
              value={formData.status}
              onChange={handleChange}
              className="w-full bg-slate-900/50 border border-white/5 rounded-xl py-3 px-4 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500/50 transition-all font-medium appearance-none"
            >
              <option value="ACTIVE">Active</option>
              <option value="PENDING">Pending</option>
              <option value="SUSPENDED">Suspended</option>
            </select>
          </div>

          {/* Email Address */}
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">
              Contact Email
            </label>
            <div className="relative group">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-blue-400 transition-colors">
                <Mail className="w-4 h-4" />
              </div>
              <input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                className="w-full bg-slate-900/50 border border-white/5 rounded-xl py-3 pl-11 pr-4 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500/50 transition-all font-medium"
                placeholder="billing@acme.com"
              />
            </div>
          </div>

          {/* Phone Number */}
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">
              Phone Number
            </label>
            <div className="relative group">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-blue-400 transition-colors">
                <Phone className="w-4 h-4" />
              </div>
              <input
                type="text"
                name="phone"
                value={formData.phone}
                onChange={handleChange}
                className="w-full bg-slate-900/50 border border-white/5 rounded-xl py-3 pl-11 pr-4 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500/50 transition-all font-medium"
                placeholder="+91 98765 43210"
              />
            </div>
          </div>

          {/* GSTIN */}
          <div className="space-y-2 md:col-span-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">
              Registration (GSTIN)
            </label>
            <div className="relative group">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-blue-400 transition-colors">
                <Hash className="w-4 h-4" />
              </div>
              <input
                type="text"
                name="gstin"
                value={formData.gstin}
                onChange={handleChange}
                className="w-full bg-slate-900/50 border border-white/5 rounded-xl py-3 pl-11 pr-4 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500/50 transition-all font-mono font-bold tracking-widest"
                placeholder="27AAAAA0000A1Z5"
              />
            </div>
          </div>

          {/* Billing Address */}
          <div className="space-y-2 md:col-span-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">
              Full Billing Address
            </label>
            <div className="relative group">
              <div className="absolute top-3 left-0 pl-4 flex items-start pointer-events-none text-slate-500 group-focus-within:text-blue-400 transition-colors">
                <MapPin className="w-4 h-4" />
              </div>
              <textarea
                name="address"
                value={formData.address}
                onChange={handleChange}
                rows={4}
                className="w-full bg-slate-900/50 border border-white/5 rounded-xl py-3 pl-11 pr-4 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500/50 transition-all font-medium resize-none"
                placeholder="Street Address, City, State, ZIP..."
              />
            </div>
          </div>
        </div>

        {error && (
          <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-xl text-xs text-rose-400 font-bold">
            {error}
          </div>
        )}

        <div className="flex gap-4 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 py-3 px-6 bg-white/5 hover:bg-white/10 text-slate-400 font-bold rounded-xl transition-all border border-transparent hover:border-white/10"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="flex-1 py-3 px-6 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-black rounded-xl transition-all shadow-lg shadow-blue-900/20 flex items-center justify-center gap-2"
          >
            {loading ? (
              <div className="w-5 h-5 border-2 border-white/20 border-t-white rounded-full animate-spin" />
            ) : (
              <>
                <CheckCircle2 className="w-5 h-5" />
                Update Profile
              </>
            )}
          </button>
        </div>
      </form>
    </Modal>
  );
};
