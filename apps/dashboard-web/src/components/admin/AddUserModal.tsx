import React, { useState, useEffect } from "react";
import Modal from "../common/Modal";
import { TextInput, SelectInput } from "../common/FormFields";
import api from "../../services/api";

interface Role {
  id: string;
  name: string;
}

interface AddUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const AddUserModal: React.FC<AddUserModalProps> = ({
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
    password: "",
  });
  const [error, setError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (isOpen) {
      fetchRoles();
      // Generate a random temporary password for convenience
      const tempPass = Math.random().toString(36).slice(-8) + "A1!";
      setFormData(prev => ({ ...prev, password: tempPass }));
    }
  }, [isOpen]);

  const fetchRoles = async () => {
    try {
      const response = await api.get("/admin/rbac/roles");
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
      await api.adminCreateUser({
        email: formData.email,
        name: formData.name,
        role_id: formData.role_id,
        password: formData.password,
      });
      onSuccess();
      onClose();
      // Reset form
      setFormData({ email: "", name: "", role_id: "", password: "" });
    } catch (err: any) {
      setError(err.response?.data?.message || "Failed to create user");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create New User"
      size="md"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <p className="text-[10px] text-text-muted mt-[-10px] mb-4">
          Directly create a user account. They will be forced to change their password upon first login.
        </p>

        {error && (
          <div className="p-3 bg-danger/10 border border-danger/20 rounded text-danger text-xs">
            {error}
          </div>
        )}

        <TextInput
          name="name"
          label="Full Name"
          placeholder="John Doe"
          value={formData.name}
          onChange={handleChange}
          required
          icon="person"
        />

        <TextInput
          name="email"
          label="Email Address"
          type="email"
          placeholder="user@example.com"
          value={formData.email}
          onChange={handleChange}
          required
          icon="mail"
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

        <div className="relative">
          <TextInput
            name="password"
            label="Temporary Password"
            type={showPassword ? "text" : "password"}
            placeholder="••••••••"
            value={formData.password}
            onChange={handleChange}
            required
            icon="key"
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-[32px] text-text-muted hover:text-white transition-colors"
          >
            <span className="material-symbols-outlined text-sm">
              {showPassword ? "visibility_off" : "visibility"}
            </span>
          </button>
        </div>

        <div className="p-3 bg-primary/5 border border-primary/10 rounded-lg">
          <div className="flex items-start gap-2">
            <span className="material-symbols-outlined text-primary text-sm mt-0.5">info</span>
            <div className="text-[10px] text-text-muted leading-relaxed">
              <span className="text-primary font-bold">Important:</span> Securely share these credentials with the user.
              The <span className="text-white font-medium">Must Reset Password</span> flag will be automatically set for this account.
            </div>
          </div>
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
            disabled={loading || !formData.email || !formData.role_id || !formData.password}
            className="px-6 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            {loading ? "Creating..." : "Create User"}
          </button>
        </div>
      </form>
    </Modal>
  );
};
