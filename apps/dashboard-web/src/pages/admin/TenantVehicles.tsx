import React from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import { DataGrid } from "../../components/common/DataGrid";

interface Vehicle {
  id: string;
  registration_number: string;
  type: string;
  make: string;
  model: string;
  status: string;
}

export const TenantVehicles: React.FC = () => {
  const { id: tenantId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("admin_vehicles");

  const columns = [
    { 
      field: "registration_number", 
      header: "Registration",
      render: (v: Vehicle) => (
        <span className="font-mono font-bold text-white uppercase tracking-tight">
          {v.registration_number || "N/A"}
        </span>
      )
    },
    { 
      field: "type", 
      header: "Asset Type",
      render: (v: Vehicle) => (
        <div className="flex items-center gap-2">
           <span className="material-symbols-outlined text-[16px] text-primary">
             {v.type?.toLowerCase() === 'truck' ? 'local_shipping' : 'directions_car'}
           </span>
           <span className="capitalize">{v.type || "Unknown"}</span>
        </div>
      )
    },
    { 
      field: "make", 
      header: "Make / Model",
      render: (v: Vehicle) => (
        <span className="text-text-muted">
          {v.make} {v.model}
        </span>
      )
    },
    { 
      field: "status", 
      header: "Status",
      render: (v: Vehicle) => (
        <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase border ${
          v.status?.toLowerCase() === 'active' 
            ? 'bg-success/10 text-success border-success/20' 
            : 'bg-danger/10 text-danger border-danger/20'
        }`}>
          {v.status || "Unknown"}
        </span>
      )
    },
    {
      field: "actions",
      header: "",
      render: (v: Vehicle) => (
        <div className="flex justify-end">
          <button
            onClick={() => navigate(`/admin/vehicles/${v.id}`)}
            className="p-1.5 hover:bg-surface-dark rounded text-text-muted hover:text-white transition-colors"
          >
            <span className="material-symbols-outlined text-[18px]">visibility</span>
          </button>
        </div>
      )
    }
  ];

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-white text-3xl font-bold mb-1 tracking-tight">
          {uiStrings?.title || "Tenant Assets"}
        </h1>
        <p className="text-text-muted text-sm max-w-2xl leading-relaxed">
          {uiStrings?.subtitle || `Reviewing all registered fleet assets for tenant ID: ${tenantId}`}
        </p>
      </div>

      <div className="bg-surface-darker/50 rounded-2xl border border-border-color overflow-hidden shadow-2xl shadow-black/40 backdrop-blur-sm">
        <DataGrid<Vehicle>
          title="Vehicle Inventory"
          endpoint={tenantId ? `/admin/vehicles?tenant_id=${tenantId}` : "/admin/vehicles"}
          columns={columns}
        />
      </div>
    </div>
  );
};
