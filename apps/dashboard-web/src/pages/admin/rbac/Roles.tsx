import {
    ShieldCheckIcon
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { DataGrid } from "../../../components/common/DataGrid";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

import { useNavigate } from "react-router-dom";

interface Role {
  id: string;
  name: string;
  description: string;
  is_system: boolean;
  created_at: string;
}

export const Roles: React.FC = () => {
  const navigate = useNavigate();
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [roles, setRoles] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("admin_rbac");

  const handleDelete = async (id: string, isSystem: boolean) => {
    if (isSystem) {
      alert(uiStrings?.err_system_role || "System roles cannot be deleted");
      return;
    }
    if (
      !window.confirm(
        uiStrings?.confirm_delete ||
          "Are you sure you want to delete this role?",
      )
    )
      return;
    try {
      await apiClient.delete(`/rbac/roles/${id}`);
      setRefreshTrigger((prev) => prev + 1); // Keep refreshTrigger for DataGrid
    } catch (err) {
      console.error(err);
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
      <div className="mb-6 flex justify-between items-end">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">
            {uiStrings?.title || "Role Management"}
          </h1>
          <p className="text-slate-500 mt-1">
            {uiStrings?.subtitle ||
              "Define and manage access levels for your organization"}
          </p>
        </div>
        <button
          className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-semibold transition-all shadow-sm flex items-center"
          onClick={() => {
            /* Open create modal */
          }}
        >
          <ShieldCheckIcon className="h-5 w-5 mr-2" />
          {uiStrings?.btn_create || "Create New Role"}
        </button>
      </div>

      <DataGrid<Role>
        title={uiStrings?.title || "Roles"}
        endpoint="/rbac/roles"
        columns={columns}
        refreshTrigger={refreshTrigger}
      />
    </div>
  );
};

export default Roles;
