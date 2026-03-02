import React from "react";
import { useUIStrings } from "../../hooks/useUIStrings";

export const OrganizationUsers: React.FC = () => {
  const { uiStrings } = useUIStrings("admin_users");
  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">
          {uiStrings?.title || "User Management"}
        </h1>
        <p className="text-text-muted text-sm">
          {uiStrings?.subtitle || "Manage users within your organization"}
        </p>
      </div>
      <div className="flex items-center justify-center h-64 border-2 border-dashed border-border-color rounded-xl">
        <p className="text-text-muted italic">
          {uiStrings?.coming_soon || "User management interface coming soon..."}
        </p>
      </div>
    </div>
  );
};
