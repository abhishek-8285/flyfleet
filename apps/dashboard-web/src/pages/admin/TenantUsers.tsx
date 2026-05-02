import React, { useState } from "react";
import { useParams } from "react-router-dom";
import { UserPlusIcon, TrashIcon, NoSymbolIcon, CheckCircleIcon } from "@heroicons/react/24/outline";
import { DataGrid } from "../../components/common/DataGrid";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";
import { InviteUserModal } from "../../components/admin/InviteUserModal";
import { AddUserModal } from "../../components/admin/AddUserModal";

interface User {
  id: string;
  name: string;
  email: string;
  status: string;
  roles: string[];
  created_at: string;
  last_login_at?: string;
}

export const TenantUsers: React.FC = () => {
  const { id: tenantId } = useParams<{ id: string }>();
  const { uiStrings } = useUIStrings("admin_users");
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);

  const handleToggleStatus = async (user: User) => {
    const newStatus = user.status === "active" ? "disabled" : "active";
    try {
      await apiClient.put(`/admin/users/${user.id}/status`, { status: newStatus });
      setRefreshTrigger((prev) => prev + 1);
    } catch (error) {
      console.error("Failed to update status", error);
    }
  };

  const handleDelete = async (user: User) => {
    if (!window.confirm(`Are you sure you want to remove ${user.name || user.email} from this tenant?`)) {
      return;
    }
    try {
      await apiClient.delete(`/admin/users/${user.id}`);
      setRefreshTrigger((prev) => prev + 1);
    } catch (error) {
      console.error("Failed to delete user", error);
    }
  };

  const columns = [
    { field: "name", header: "User", render: (u: User) => (
      <div className="flex items-center gap-3">
        <div className="h-8 w-8 rounded-full bg-gradient-to-br from-primary/20 to-primary/5 flex items-center justify-center border border-primary/10">
          <span className="text-primary text-xs font-bold uppercase">
            {(u.name || u.email).substring(0, 2)}
          </span>
        </div>
        <div>
          <div className="text-white text-sm font-medium">{u.name || "N/A"}</div>
          <div className="text-text-muted text-[10px]">{u.email}</div>
        </div>
      </div>
    )},
    { field: "status", header: "Status", render: (u: User) => (
      <span className={`px-2 py-1 rounded-full text-[10px] uppercase tracking-wider font-semibold ${
        u.status === "active" 
          ? "bg-success/10 text-success border border-success/20" 
          : "bg-danger/10 text-danger border border-danger/20"
      }`}>
        {u.status}
      </span>
    )},
    { field: "roles", header: "Roles", render: (u: User) => (
      <div className="flex flex-wrap gap-1">
        {u.roles?.map(r => (
          <span key={r} className="px-1.5 py-0.5 bg-surface-darker border border-border-color rounded text-[9px] text-text-muted font-mono">
            {r}
          </span>
        )) || <span className="text-text-muted italic text-[9px]">No roles</span>}
      </div>
    )},
    { field: "created_at", header: "Joined", render: (u: User) => (
      <span className="text-text-muted text-xs font-mono">
        {new Date(u.created_at).toLocaleDateString()}
      </span>
    )},
    { field: "actions", header: "", render: (u: User) => (
      <div className="flex items-center justify-end gap-2">
        <button
          onClick={() => handleToggleStatus(u)}
          title={u.status === "active" ? "Deactivate" : "Activate"}
          className="p-1.5 hover:bg-surface-dark rounded-lg transition-colors group"
        >
          {u.status === "active" ? (
            <NoSymbolIcon className="h-4 w-4 text-text-muted group-hover:text-danger" />
          ) : (
            <CheckCircleIcon className="h-4 w-4 text-text-muted group-hover:text-success" />
          )}
        </button>
        <button
          onClick={() => handleDelete(u)}
          className="p-1.5 hover:bg-danger/10 rounded-lg transition-colors group"
          title="Remove"
        >
          <TrashIcon className="h-4 w-4 text-text-muted group-hover:text-danger" />
        </button>
      </div>
    )},
  ];

  const actions = (
    <div className="flex items-center gap-3">
      <button
        onClick={() => setIsInviteModalOpen(true)}
        className="px-4 py-2 bg-surface-dark hover:bg-surface text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors border border-border-color"
      >
        <span className="material-symbols-outlined text-sm">mail</span>
        Invite
      </button>
      <button
        onClick={() => setIsAddModalOpen(true)}
        className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
      >
        <UserPlusIcon className="h-4 w-4" />
        Add User
      </button>
    </div>
  );

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-white text-3xl font-bold mb-1 tracking-tight">
          {uiStrings?.title || "Tenant Identity"}
        </h1>
        <p className="text-text-muted text-sm max-w-2xl leading-relaxed">
          {uiStrings?.subtitle || "Manage your mission control team. Add, remove, or modify user access across your fleet operations center."}
        </p>
      </div>

      <div className="bg-surface-darker/50 rounded-2xl border border-border-color overflow-hidden shadow-2xl shadow-black/40 backdrop-blur-sm">
        <DataGrid<User>
          title="Team Members"
          endpoint={tenantId ? `/admin/users?tenant_id=${tenantId}` : "/admin/users"}
          columns={columns}
          actions={actions}
          refreshTrigger={refreshTrigger}
        />
      </div>

      <InviteUserModal 
        isOpen={isInviteModalOpen}
        onClose={() => setIsInviteModalOpen(false)}
        onSuccess={() => setRefreshTrigger(prev => prev + 1)}
      />

      <AddUserModal
        isOpen={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        onSuccess={() => setRefreshTrigger(prev => prev + 1)}
      />
    </div>
  );
};
