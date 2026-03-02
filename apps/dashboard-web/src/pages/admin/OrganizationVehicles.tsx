import React from "react";
import { useParams } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

export const OrganizationVehicles: React.FC = () => {
  const { id } = useParams();
  const { uiStrings } = useUIStrings("admin_organizations");

  return (
    <div className="p-8">
      <h1 className="text-white text-2xl font-bold mb-4">
        {uiStrings?.title_vehicles || "Organization Vehicles"}
      </h1>
      <p className="text-text-muted">
        {uiStrings?.subtitle_vehicles?.replace("{id}", id || "") ||
          `Manage vehicles for organization ID: ${id}`}
      </p>
      <div className="mt-8 bg-surface-dark rounded-xl border border-border-color p-6">
        <p className="text-text-muted text-center">
          {uiStrings?.coming_soon ||
            "Vehicle management interface coming soon..."}
        </p>
      </div>
    </div>
  );
};
