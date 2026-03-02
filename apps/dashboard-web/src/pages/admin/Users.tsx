import { UserPlusIcon } from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { DataGrid } from "../../components/common/DataGrid";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  status: string;
  role_id: string; // Ideally map to role name
  created_at: string;
}

export const Users: React.FC = () => {
  const navigate = useNavigate();
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const { uiStrings } = useUIStrings("admin_users");

  const handleDelete = async (id: string) => {
    if (
      window.confirm(
        uiStrings?.confirm_delete ||
          "Are you sure you want to delete this user?",
      )
    ) {
      try {
        await apiClient.delete(`/admin/users?id=${id}`);
        setRefreshTrigger((prev) => prev + 1);
      } catch (error) {
        alert(uiStrings?.err_delete || "Failed to delete user");
      }
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
    <div className="w-full">
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">
          {uiStrings?.title || "User Management"}
        </h1>
        <p className="text-text-muted text-sm">
          {uiStrings?.subtitle || "Manage users within your organization"}
        </p>
      </div>

      <div className="mb-6 flex justify-end">
        <button
          onClick={() => navigate("/admin/users/new")}
          className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
        >
          <UserPlusIcon className="h-4 w-4" />
          {uiStrings?.btn_add || "Add User"}
        </button>
      </div>

      <DataGrid<User>
        title="Users"
        endpoint="/admin/users"
        columns={columns}
        refreshTrigger={refreshTrigger}
      />
    </div>
  );
};
