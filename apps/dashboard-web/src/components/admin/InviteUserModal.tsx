import React, { useState, useEffect } from "react";
import Modal from "../common/Modal";
import { TextInput, SelectInput } from "../common/FormFields";
import apiClient from "../../services/apiClient";

interface Role {
  id: string;
  name: string;
}

interface InviteUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const InviteUserModal: React.FC<InviteUserModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [loading, setLoading] = useState(false);
  const [roles, setRoles] = useState<Role[]>([]);
  const [formData, setFormData] = useState({
    email: "",
    name: "",
    role_id: "",
    message: "Welcome to our organization. Please join us on the FlyFleet platform.",
  });
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      fetchRoles();
    }
  }, [isOpen]);

  const fetchRoles = async () => {
    try {
      const response = await apiClient.get("/admin/rbac/roles");
      // Backend returns { data: [...] } or direct array
      const rawRoles = (response as any).data || response;
      if (Array.isArray(rawRoles)) {
        setRoles(rawRoles.map((r: any) => ({ id: r.id, name: r.name })));
      }
    } catch (err) {
      console.error("Failed to fetch roles", err);
    }
  };

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await apiClient.post("/admin/users/invite", {
        email: formData.email,
        role_ids: [formData.role_id],
        message: formData.message,
      });
      onSuccess();
      onClose();
      // Reset form
      setFormData({ email: "", name: "", role_id: "", message: "" });
    } catch (err: any) {
      setError(err.response?.data?.message || "Failed to send invitation");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Invite New User"
      size="md"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-danger/10 border border-danger/20 rounded text-danger text-xs">
            {error}
          </div>
        )}

        <TextInput
          name="email"
          label="Email Address"
          type="email"
          placeholder="colleague@company.com"
          value={formData.email}
          onChange={handleChange}
          required
          icon="mail"
        />

        <TextInput
          name="name"
          label="Full Name (Optional)"
          placeholder="John Doe"
          value={formData.name}
          onChange={handleChange}
          icon="person"
        />

        <SelectInput
          name="role_id"
          label="Assigned Role"
          placeholder="Select a role..."
          value={formData.role_id}
          onChange={handleChange}
          options={roles.map((r) => ({ value: r.id, label: r.name }))}
          required
          icon="shield_person"
        />

        <div className="space-y-1.5">
          <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
            Personal Message
          </label>
          <textarea
            name="message"
            value={formData.message}
            onChange={(e) => setFormData(prev => ({ ...prev, message: e.target.value }))}
            className="w-full bg-surface-darker border border-border-color rounded-lg px-4 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
            rows={3}
          />
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-border-color">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-xs font-semibold text-text-muted hover:text-white transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading || !formData.email || !formData.role_id}
            className="px-6 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            {loading ? "Sending..." : "Send Invitation"}
          </button>
        </div>
      </form>
    </Modal>
  );
};
