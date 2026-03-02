import {
    ArrowLeftIcon,
    SignalIcon as SignalSolidIcon,
    TruckIcon,
    UserCircleIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import apiClient from "../../services/apiClient";

interface Vehicle {
  id: string;
  name: string;
  license_plate: string;
  status: string;
  org_id: string;
  current_driver_id?: string;
  current_driver_name?: string;
  current_driver_since?: string;
  device_id?: string;
}

export const VehicleDetail: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchVehicle = async () => {
      try {
        const direct = (await apiClient.get(`/admin/vehicles?id=${id}`)) as any;
        const directVehicle = direct?.data ?? direct;
        if (directVehicle?.id) {
          setVehicle(directVehicle);
          return;
        }

        // Fallback for older API responses.
        const res = (await apiClient.get(
          "/admin/vehicles?page=1&size=100",
        )) as any;
        const list = res?.data ?? [];
        const v = list.find((item: any) => item.id === id);
        if (v) setVehicle(v);
      } catch (error) {
        console.error("Failed to fetch vehicle:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchVehicle();
  }, [id]);

  if (loading)
    return (
      <div className="p-12 text-center text-slate-400">
        Loading asset profile...
      </div>
    );
  if (!vehicle)
    return (
      <div className="p-12 text-center text-red-400">Vehicle not found</div>
    );

  return (
    <div className="p-8 max-w-5xl mx-auto space-y-8">
      <button
        onClick={() => navigate("/admin/vehicles")}
        className="flex items-center space-x-2 text-slate-500 hover:text-blue-600 transition-colors mb-4 group"
      >
        <ArrowLeftIcon className="h-4 w-4 group-hover:-translate-x-1 transition-transform" />
        <span>Back to Fleet</span>
      </button>

      <div className="bg-white rounded-3xl p-8 shadow-sm border border-slate-100 flex justify-between items-start">
        <div className="flex space-x-6">
          <div className="h-24 w-24 bg-blue-50 rounded-2xl flex items-center justify-center text-blue-500">
            <TruckIcon className="h-10 w-10" />
          </div>
          <div>
            <h1 className="text-3xl font-bold text-slate-900">
              {vehicle.name}
            </h1>
            <p className="text-slate-500 mt-1 font-mono text-lg">
              {vehicle.license_plate}
            </p>
            <div className="flex items-center space-x-3 mt-4">
              <span
                className={`px-3 py-1 rounded-full text-xs font-bold uppercase ${
                  vehicle.status === "active"
                    ? "bg-green-100 text-green-700"
                    : "bg-slate-100 text-slate-500"
                }`}
              >
                {vehicle.status}
              </span>
              <span className="text-slate-300">|</span>
              <span className="text-sm text-slate-400">ID: {vehicle.id}</span>
            </div>
          </div>
        </div>
        <div className="text-right">
          <button className="bg-white border border-slate-200 text-slate-600 px-4 py-2 rounded-lg text-sm font-medium hover:bg-slate-50 shadow-sm transition-all">
            Edit Details
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {/* Driver Assignment Card */}
        <div className="bg-white p-8 rounded-3xl border border-slate-100 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex justify-between items-center mb-6">
            <h3 className="font-bold text-slate-800 flex items-center text-lg">
              <div className="p-2 bg-purple-50 rounded-lg mr-3 text-purple-500">
                <UserCircleIcon className="h-6 w-6" />
              </div>
              Assigned Driver
            </h3>
            {vehicle.current_driver_id && (
              <button className="text-sm text-blue-600 font-semibold hover:underline">
                Change
              </button>
            )}
          </div>

          {vehicle.current_driver_id ? (
            <div className="flex items-center space-x-4 p-4 bg-slate-50 rounded-xl border border-slate-100">
              <div className="h-12 w-12 bg-purple-100 rounded-full flex items-center justify-center text-purple-600 font-bold">
                JD
              </div>
              <div>
                <p className="font-bold text-slate-900">
                  {vehicle.current_driver_name || "---"}
                </p>
                <p className="text-xs text-slate-500">
                  Active since {vehicle.current_driver_since || "---"}
                </p>
              </div>
            </div>
          ) : (
            <div className="text-center py-10 bg-slate-50/50 rounded-2xl border border-dashed border-slate-200">
              <p className="text-slate-400 text-sm mb-4">
                No driver currently assigned
              </p>
              <button className="bg-white border border-slate-200 text-slate-600 px-5 py-2.5 rounded-xl text-sm font-semibold hover:border-purple-200 hover:text-purple-600 hover:shadow-sm transition-all">
                Assign Driver
              </button>
            </div>
          )}
        </div>

        {/* Device Assignment Card */}
        <div className="bg-white p-8 rounded-3xl border border-slate-100 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex justify-between items-center mb-6">
            <h3 className="font-bold text-slate-800 flex items-center text-lg">
              <div className="p-2 bg-emerald-50 rounded-lg mr-3 text-emerald-500">
                <SignalSolidIcon className="h-6 w-6" />
              </div>
              Connected Device
            </h3>
            <button className="text-sm text-blue-600 font-semibold hover:underline">
              Manage
            </button>
          </div>

          <div className="text-center py-10 bg-slate-50/50 rounded-2xl border border-dashed border-slate-200">
            <p className="text-slate-400 text-sm mb-4">
              No telemetry device linked
            </p>
            <button className="bg-white border border-slate-200 text-slate-600 px-5 py-2.5 rounded-xl text-sm font-semibold hover:border-emerald-200 hover:text-emerald-600 hover:shadow-sm transition-all">
              Link Device
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
