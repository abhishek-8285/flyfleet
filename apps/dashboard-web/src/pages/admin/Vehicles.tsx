import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import apiClient from "../../services/apiClient";

interface Vehicle {
  id: string;
  name: string;
  registration_number: string;
  license_plate: string;
  status: string;
  make: string;
  model: string;
  org_id: string;
  created_at: string;
}

export const Vehicles: React.FC = () => {
  const navigate = useNavigate();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [uiStrings, setUiStrings] = useState<any>(null);

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.admin || data.system_metadata?.ui_strings?.admin,
        ),
      )
      .catch((err) => console.error(err));
  }, []);

  useEffect(() => {
    const fetchVehicles = async () => {
      try {
        const res = (await apiClient.get(
          "/admin/vehicles?page=1&size=100",
        )) as any;
        setVehicles(res.data || []);
      } catch (error) {
        console.error("Failed to fetch vehicles:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchVehicles();
  }, []);

  return (
    <div className="w-full">
      <div className="mb-6">
        <h1 className="text-white text-2xl font-bold mb-1">
          {uiStrings?.vehicles_title || "Fleet Vehicles"}
        </h1>
        <p className="text-text-muted text-sm">
          {uiStrings?.vehicles_sub ||
            "Manage your transport assets and assignments."}
        </p>
      </div>

      <div className="mb-6 flex justify-end">
        <button
          onClick={() => navigate("/admin/vehicles/new")}
          className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
        >
          <span className="material-symbols-outlined text-[16px]">add</span>
          {uiStrings?.add_vehicle_btn || "Add Vehicle"}
        </button>
      </div>

      {loading ? (
        <div className="text-center py-20 text-text-muted">
          {uiStrings?.loading_label || "Loading fleet data..."}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {vehicles.map((vehicle) => (
            <div
              key={vehicle.id}
              onClick={() => navigate(`/admin/vehicles/${vehicle.id}`)}
              className="bg-surface-dark p-6 rounded-xl border border-border-color hover:border-primary/30 transition-all cursor-pointer group"
            >
              <div className="flex justify-between items-start mb-4">
                <div className="p-3 bg-primary/10 rounded-xl text-primary group-hover:bg-primary group-hover:text-white transition-colors">
                  <span className="material-symbols-outlined text-[24px]">
                    local_shipping
                  </span>
                </div>
                <span
                  className={`px-2 py-1 rounded-full text-xs font-bold uppercase ${
                    vehicle.status === "active"
                      ? "bg-success/20 text-success"
                      : "bg-text-muted/20 text-text-muted"
                  }`}
                >
                  {vehicle.status}
                </span>
              </div>

              <h3 className="text-lg font-bold text-white">
                {vehicle.registration_number || vehicle.license_plate || vehicle.name || "Unnamed Asset"}
              </h3>
              <p className="text-text-muted text-[10px] font-mono mt-1 opacity-50 uppercase tracking-wider">
                {vehicle.id}
              </p>

              <div className="mt-6 pt-4 border-t border-border-color flex justify-between text-sm text-text-muted">
                <span className="flex items-center">
                  <span className="material-symbols-outlined text-[16px] mr-1">
                    person
                  </span>{" "}
                  {uiStrings?.no_driver || "No Driver"}
                </span>
                <span className="flex items-center">
                  <span className="material-symbols-outlined text-[16px] mr-1">
                    devices
                  </span>{" "}
                  {uiStrings?.no_device || "No Device"}
                </span>
              </div>
            </div>
          ))}

          {/* Empty State Action */}
          {vehicles.length === 0 && (
            <div
              onClick={() => navigate("/admin/vehicles/new")}
              className="border-2 border-dashed border-border-color rounded-xl p-6 flex flex-col items-center justify-center text-text-muted hover:border-primary/50 hover:text-primary cursor-pointer transition-colors min-h-[200px]"
            >
              <span className="material-symbols-outlined text-[32px] mb-2">
                add
              </span>
              <span className="font-medium">
                {uiStrings?.register_first_btn || "Register First Vehicle"}
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
