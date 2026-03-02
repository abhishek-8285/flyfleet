import { PlusIcon, XMarkIcon } from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { DataGrid } from "../../../components/common/DataGrid";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface User {
  id: string;
  name: string;
  email: string;
}

interface Role {
  id: string;
  name: string;
}

export const UserAssignments: React.FC = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [userRoles, setUserRoles] = useState<Role[]>([]);
  const [allRoles, setAllRoles] = useState<Role[]>([]);
  const [showModal, setShowModal] = useState(false);
  const { uiStrings: rawUiStrings } = useUIStrings("rbac_assignments");
  const uiStrings = rawUiStrings || {};

  useEffect(() => {
    const fetchRoles = async () => {
      try {
        const res = (await apiClient.get("/rbac/roles")) as any;
        setAllRoles(res.data);
      } catch (error) {
        console.error("Failed to fetch roles", error);
      }
    };
    fetchRoles();
  }, []);

  const fetchUserRoles = async (userId: string) => {
    try {
      const res = (await apiClient.get(`/rbac/users/${userId}/roles`)) as any;
      setUserRoles(res.data);
    } catch (error) {
      console.error("Failed to fetch user roles", error);
    }
  };

  const handleAssign = async (userId: string, roleId: string) => {
    try {
      await apiClient.post("/rbac/users/assign", {
        user_id: userId,
        role_id: roleId,
      });
      fetchUserRoles(userId);
      setRefreshTrigger((prev) => prev + 1);
    } catch (error) {
      alert(uiStrings?.msg_assign_failed || "Failed to assign role");
    }
  };

  const handleUnassign = async (userId: string, roleId: string) => {
    try {
      await apiClient.delete("/rbac/users/unassign", {
        data: { user_id: userId, role_id: roleId },
      });
      fetchUserRoles(userId);
      setRefreshTrigger((prev) => prev + 1);
    } catch (error) {
      alert(uiStrings?.msg_unassign_failed || "Failed to unassign role");
    }
  };

  const [columns, setColumns] = useState<any[]>([]);
  useEffect(() => {
    fetch("/api/v1/ui/settings/columns")
      .then((res) => res.json())
      .then((data) => setColumns(data))
      .catch((err) => console.error(err));
  }, []);

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-800">
          {uiStrings?.title || "User Role Assignments"}
        </h1>
        <p className="text-slate-500 mt-1">
          {uiStrings?.subtitle ||
            "Assign and manage access roles for team members"}
        </p>
      </div>

      <DataGrid<User>
        title={uiStrings?.grid_title || "Users"}
        endpoint="/admin/users"
        columns={columns}
        refreshTrigger={refreshTrigger}
      />

      {showModal && selectedUser && (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50">
              <div>
                <h3 className="font-bold text-slate-800">
                  {uiStrings?.modal_title?.replace(
                    "{name}",
                    selectedUser.name,
                  ) || `Manage Roles: ${selectedUser.name}`}
                </h3>
                <p className="text-xs text-slate-500">{selectedUser.email}</p>
              </div>
              <button
                onClick={() => setShowModal(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <XMarkIcon className="h-6 w-6" />
              </button>
            </div>

            <div className="p-6">
              <div className="mb-6">
                <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">
                  {uiStrings?.label_current_assignments ||
                    "Current Assignments"}
                </label>
                <div className="flex flex-wrap gap-2">
                  {userRoles.length === 0 ? (
                    <p className="text-sm text-slate-400 italic">
                      {uiStrings?.msg_no_roles || "No roles assigned"}
                    </p>
                  ) : (
                    userRoles.map((role) => (
                      <span
                        key={role.id}
                        className="inline-flex items-center bg-blue-50 text-blue-700 px-3 py-1.5 rounded-full text-sm font-medium border border-blue-100"
                      >
                        {role.name}
                        <button
                          onClick={() =>
                            handleUnassign(selectedUser.id, role.id)
                          }
                          className="ml-2 hover:text-blue-900 transition-colors"
                        >
                          <XMarkIcon className="h-4 w-4" />
                        </button>
                      </span>
                    ))
                  )}
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">
                  {uiStrings?.label_available_roles || "Available Roles"}
                </label>
                <div className="grid grid-cols-1 gap-2 max-h-48 overflow-y-auto pr-2">
                  {allRoles
                    .filter((r) => !userRoles.some((ur) => ur.id === r.id))
                    .map((role) => (
                      <div
                        key={role.id}
                        className="flex justify-between items-center p-3 rounded-xl border border-slate-100 hover:border-blue-200 hover:bg-blue-50/30 transition-all group"
                      >
                        <span className="text-sm font-medium text-slate-700">
                          {role.name}
                        </span>
                        <button
                          onClick={() => handleAssign(selectedUser.id, role.id)}
                          className="text-blue-600 opacity-0 group-hover:opacity-100 transition-opacity flex items-center text-xs font-bold"
                        >
                          <PlusIcon className="h-4 w-4 mr-1" />
                          {uiStrings?.btn_assign || "ASSIGN"}
                        </button>
                      </div>
                    ))}
                  {allRoles.filter(
                    (r) => !userRoles.some((ur) => ur.id === r.id),
                  ).length === 0 && (
                    <p className="text-sm text-slate-400 italic">
                      {uiStrings?.msg_all_assigned || "Everything assigned"}
                    </p>
                  )}
                </div>
              </div>
            </div>

            <div className="px-6 py-4 bg-slate-50 border-t border-slate-100 flex justify-end">
              <button
                onClick={() => setShowModal(false)}
                className="px-4 py-2 text-sm font-semibold text-slate-600 hover:text-slate-800 transition-colors"
              >
                {uiStrings?.btn_done || "Done"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
