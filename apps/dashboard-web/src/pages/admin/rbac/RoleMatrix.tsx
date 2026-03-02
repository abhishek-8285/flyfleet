import {
    CheckIcon,
    ShieldExclamationIcon,
    XMarkIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface Permission {
  id: string;
  module: string;
  action: string;
  description: string;
}

interface Role {
  id: string;
  name: string;
  is_system: boolean;
}

export const RoleMatrix: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [role, setRole] = useState<Role | null>(null);
  const [allPermissions, setAllPermissions] = useState<Permission[]>([]);
  const [rolePermissions, setRolePermissions] = useState<string[]>([]); // Array of IDs
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { uiStrings } = useUIStrings("admin_rbac");

  const fetchData = async () => {
    try {
      const [roleRes, allPermsRes, rolePermsRes] = await Promise.all([
        apiClient.get(`/rbac/roles/${id}`) as any,
        apiClient.get("/rbac/permissions") as any,
        apiClient.get(`/rbac/roles/${id}/permissions`) as any,
      ]);
      setRole(roleRes.data);
      setAllPermissions(allPermsRes.data);
      setRolePermissions(rolePermsRes.data.map((p: any) => p.id));
    } catch (error) {
      console.error("Failed to fetch matrix data", error);
      setError(uiStrings?.fetch_error || "Failed to load permission matrix.");
    }
  };

  useEffect(() => {
    if (id) fetchData();
  }, [id]);

  const handleToggle = (permId: string) => {
    if (role?.is_system) return;
    setRolePermissions((prev) =>
      prev.includes(permId)
        ? prev.filter((p) => p !== permId)
        : [...prev, permId],
    );
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      await apiClient.put(`/rbac/roles/${id}/permissions`, {
        permission_ids: rolePermissions,
      });
      // alert('Permissions updated successfully'); // Removed as per diff
    } catch (error) {
      console.error("Failed to update permissions", error);
      setError(uiStrings?.save_error || "Failed to update permissions.");
    } finally {
      setSaving(false);
    }
  };

  if (!role && !error)
    return (
      <div className="flex items-center justify-center h-64 text-text-muted">
        {uiStrings?.loading_matrix || "Loading permission matrix..."}
      </div>
    );

  if (error) {
    return <div className="p-8 text-center text-red-500">{error}</div>;
  }

  // Group permissions by module
  const modules = Array.from(new Set(allPermissions.map((p) => p.module)));

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-white text-xl font-bold">
            {uiStrings?.matrix_title || "Permission Matrix:"} {role?.name}
          </h2>
          <p className="text-text-muted text-sm">
            {uiStrings?.matrix_subtitle ||
              "Configure granular access controls for this role"}
          </p>
        </div>
        {!role?.is_system && (
          <button
            onClick={handleSave}
            disabled={saving}
            className="bg-primary text-white px-6 py-2 rounded-lg font-bold hover:bg-primary-hover disabled:opacity-50"
          >
            {saving
              ? uiStrings?.btn_saving || "Saving..."
              : uiStrings?.btn_save || "Save Changes"}
          </button>
        )}
      </div>

      {role?.is_system && (
        <div className="mb-6 bg-amber-500/10 border border-amber-500/20 rounded-xl p-4 flex items-center gap-3">
          <ShieldExclamationIcon className="h-5 w-5 text-amber-500" />
          <p className="text-amber-500 text-sm font-medium">
            {uiStrings?.system_locked_msg ||
              "System roles have locked permissions that cannot be modified."}
          </p>
        </div>
      )}

      <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <table className="w-full">
          <thead className="bg-surface-dark border-b border-border-color">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                {uiStrings?.th_module || "Module / Permission"}
              </th>
              <th className="px-6 py-3 text-center text-xs font-medium text-text-muted uppercase tracking-wider">
                {uiStrings?.th_status || "Status"}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                {uiStrings?.th_desc || "Description"}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {modules.map((module) => (
              <React.Fragment key={module}>
                <tr className="bg-slate-50/50">
                  <td
                    colSpan={3}
                    className="px-6 py-2 text-xs font-bold text-slate-400 uppercase tracking-widest bg-slate-100/30"
                  >
                    {module}
                  </td>
                </tr>
                {allPermissions
                  .filter((p) => p.module === module)
                  .map((perm) => (
                    <tr
                      key={perm.id}
                      className={`hover:bg-slate-50/80 transition-colors cursor-pointer ${role?.is_system ? "cursor-default" : ""}`}
                      onClick={() => handleToggle(perm.id)}
                    >
                      <td className="px-6 py-4 text-sm font-medium text-slate-700">
                        {perm.action}
                      </td>
                      <td className="px-6 py-4 text-center">
                        <div
                          className={`mx-auto h-6 w-6 rounded flex items-center justify-center transition-all ${
                            rolePermissions.includes(perm.id)
                              ? "bg-blue-100 text-blue-600"
                              : "bg-slate-100 text-slate-300"
                          }`}
                        >
                          {rolePermissions.includes(perm.id) ? (
                            <CheckIcon className="h-4 w-4 stroke-[3]" />
                          ) : (
                            <XMarkIcon className="h-4 w-4" />
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-500 italic">
                        {perm.description ||
                          `Grants access to ${perm.action} in ${perm.module}`}
                      </td>
                    </tr>
                  ))}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
