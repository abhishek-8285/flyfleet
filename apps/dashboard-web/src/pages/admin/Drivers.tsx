import React, { useEffect, useState } from "react";
import { DataGrid } from "../../components/common/DataGrid";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

interface Driver {
  id: string;
  name: string;
  phone: string;
  status: string;
  created_at: string;
}

export const Drivers: React.FC = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const { uiStrings } = useUIStrings("admin_drivers");

  const handleDelete = async (id: string) => {
    if (
      window.confirm(
        uiStrings?.delete_confirm ||
          "Are you sure you want to delete this driver?",
      )
    ) {
      try {
        await apiClient.delete(`/admin/drivers?id=${id}`);
        setRefreshTrigger((prev) => prev + 1);
      } catch (error) {
        alert(uiStrings?.delete_error || "Failed to delete driver");
      }
    }
  };

  const [columns, setColumns] = useState<any[]>([]);
  useEffect(() => {
    apiClient
      .get("/ui/settings/columns")
      .then((res: any) => {
        const cols = res?.drivers || [];
        setColumns(cols);
      })
      .catch((err) => console.error(err));
  }, []);

  return (
    <div className="w-full">
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">
          {uiStrings?.title || "Drivers"}
        </h1>
        <p className="text-text-muted text-sm">
          {uiStrings?.subtitle || "Manage driver profiles and assignments"}
        </p>
      </div>

      <DataGrid<Driver>
        title={uiStrings?.list_title || "Fleet Drivers"}
        endpoint="/admin/drivers"
        columns={columns}
        refreshTrigger={refreshTrigger}
        actions={
          <button className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20">
            <span className="material-symbols-outlined text-[16px]">add</span>
            {uiStrings?.add_btn || "Add Driver"}
          </button>
        }
      />
    </div>
  );
};
