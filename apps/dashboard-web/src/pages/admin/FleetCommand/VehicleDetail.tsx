import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import apiClient from "../../../services/apiClient";
import { STATUS_COLORS, FUEL_LABELS, TYPE_ICONS } from "./constants";
import { fromKm, fromTonnes, fromLitres, getUnitLabel } from "./utils";

export const VehicleDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [vehicle, setVehicle] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"overview" | "documents" | "route">("overview");
  const [systemStatus, setSystemStatus] = useState<"ok" | "error">("ok");
  const [currentDriver, setCurrentDriver] = useState<any>(null);
  const [displayUnits, setDisplayUnits] = useState({
    odometer: "km",
    capacity: "tonnes",
    fuel: "litres",
  });

  // CNG uses kg, not litres
  const isCNG = vehicle?.fuel_type === "CNG";
  const fuelUnitLabel = isCNG ? "kg" : getUnitLabel("fuel", displayUnits.fuel);
  const displayFuel = isCNG
    ? (vehicle?.fuel_tank_capacity || "—").toString()
    : (vehicle?.fuel_tank_capacity ? fromLitres(vehicle.fuel_tank_capacity, displayUnits.fuel) : "—");

  useEffect(() => {
    fetchVehicle();
    checkSystemHealth();
  }, [id]);

  const fetchVehicle = async () => {
    try {
      const res = await apiClient.get(`/admin/vehicles/${id}`);
      setVehicle(res.data);
      
      // Fetch current driver info if assigned
      if (res.data?.current_driver_id) {
        console.log("[VehicleDetail] Found current_driver_id:", res.data.current_driver_id);
        try {
          const driversRes = await apiClient.get("/admin/drivers?limit=500");
          console.log("[VehicleDetail] Drivers API response:", driversRes);
          const ensureArray = (resp: any, key?: string) => {
            if (!resp) return [];
            if (Array.isArray(resp)) return resp;
            if (key && Array.isArray(resp[key])) return resp[key];
            if (Array.isArray(resp.data)) return resp.data;
            if (resp.data && key && Array.isArray(resp.data[key])) return resp.data[key];
            return [];
          };
          const allDrivers = ensureArray(driversRes, "drivers");
          console.log("[VehicleDetail] All drivers array:", allDrivers);
          
          const targetId = String(res.data.current_driver_id).toLowerCase();
          const driver = allDrivers.find((d: any) => String(d.id).toLowerCase() === targetId);
          
          if (driver) {
            console.log("[VehicleDetail] Matched driver:", driver);
            // Normalize name
            const fullName = `${driver.first_name || ""} ${driver.last_name || ""}`.trim() || driver.name || "Unknown Driver";
            setCurrentDriver({ ...driver, name: fullName });
          } else {
            console.warn("[VehicleDetail] Could not find driver with ID in list:", targetId);
            setCurrentDriver(null);
          }
        } catch (err) {
          console.error("[VehicleDetail] Failed to fetch drivers list:", err);
          setCurrentDriver(null);
        }
      } else {
        console.log("[VehicleDetail] No current_driver_id assigned to this vehicle");
      }
    } catch (error) {
      console.error("Failed to fetch vehicle:", error);
    } finally {
      setLoading(false);
    }
  };

  const checkSystemHealth = async () => {
    try {
      const res = await fetch("/health");
      const data = await res.json();
      setSystemStatus(data.status === "ok" ? "ok" : "error");
    } catch {
      setSystemStatus("error");
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-background-dark">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (!vehicle) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center bg-background-dark text-center">
        <span className="material-symbols-outlined text-6xl text-text-muted opacity-30 mb-4">
          directions_car_off
        </span>
        <h2 className="text-white text-xl font-bold mb-1">Vehicle Not Found</h2>
        <p className="text-text-muted text-sm mb-4">
          This vehicle may have been deleted or does not exist.
        </p>
        <button
          onClick={() => navigate("/admin/vehicles")}
          className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded transition-colors"
        >
          Back to Vehicles
        </button>
      </div>
    );
  }

  const status = STATUS_COLORS[vehicle.status] || {
    bg: "bg-text-muted/20",
    text: "text-text-muted",
    label: vehicle.status,
  };
  const vehicleIcon = TYPE_ICONS[vehicle.vehicle_type] || "directions_car";
  const fuelLabel = FUEL_LABELS[vehicle.fuel_type] || vehicle.fuel_type || "Unknown";

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden">
      {/* Header */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <div className="flex items-center gap-2 text-xs text-text-muted mb-0.5">
              <span
                className="hover:text-white cursor-pointer"
                onClick={() => navigate("/admin/vehicles")}
              >
                Vehicles
              </span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">{vehicle.registration_number}</span>
            </div>
            <h2 className="text-white text-lg font-bold flex items-center gap-2">
              {vehicle.make} {vehicle.model} {vehicle.year}
              <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${status.bg} ${status.text} border border-${status.text.split("-")[1]}/30`}>
                {status.label}
              </span>
            </h2>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button className="px-3 py-2 bg-surface-dark hover:bg-border-color border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors">
            <span className="material-symbols-outlined text-[16px]">autorenew</span>
            Update Status
          </button>
          <button
            onClick={() => navigate(`/admin/vehicles/${vehicle.id}/edit`)}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <span className="material-symbols-outlined text-[16px]">edit</span>
            Edit Vehicle
          </button>
        </div>
      </header>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {/* Stat Cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          {/* Odometer */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-white">speed</span>
            </div>
            <div className="flex items-center justify-between mb-1">
              <div className="text-text-muted text-xs font-medium uppercase">Odometer</div>
              <select
                value={displayUnits.odometer}
                onChange={(e) => setDisplayUnits((prev) => ({ ...prev, odometer: e.target.value }))}
                className="text-[10px] bg-transparent text-text-muted border border-border-color rounded px-1 py-0.5 outline-none"
              >
                <option value="km">km</option>
                <option value="mi">mi</option>
              </select>
            </div>
            <div className="text-2xl font-bold text-white">
              {vehicle.odometer ? fromKm(vehicle.odometer, displayUnits.odometer) : "0"}{" "}
              <span className="text-sm font-normal text-text-muted">{getUnitLabel("odometer", displayUnits.odometer)}</span>
            </div>
            <div className="text-[10px] text-text-muted flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">schedule</span>
              Registered {new Date(vehicle.created_at).toLocaleDateString()}
            </div>
          </div>

          {/* Fuel Type */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-warning">local_gas_station</span>
            </div>
            <div className="text-text-muted text-xs font-medium uppercase mb-1">Fuel Type</div>
            <div className="text-2xl font-bold text-white">{fuelLabel}</div>
            <div className="text-[10px] text-text-muted mt-1">
              {isCNG
                ? (vehicle.fuel_tank_capacity ? `${vehicle.fuel_tank_capacity} kg cylinder` : "Cylinder size unknown")
                : (vehicle.fuel_tank_capacity ? `${vehicle.fuel_tank_capacity}L tank` : "Tank size unknown")
              }
            </div>
          </div>

          {/* Status */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-success">check_circle</span>
            </div>
            <div className="text-text-muted text-xs font-medium uppercase mb-1">Vehicle Status</div>
            <div className={`text-2xl font-bold ${status.text}`}>{status.label}</div>
            <div className="text-[10px] text-text-muted mt-1">
              {vehicle.ownership_type ? `${vehicle.ownership_type.charAt(0) + vehicle.ownership_type.slice(1).toLowerCase()} vehicle` : ""}
            </div>
          </div>

          {/* Current Driver */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-primary">person</span>
            </div>
            <div className="text-text-muted text-xs font-medium uppercase mb-1">Current Driver</div>
            {currentDriver ? (
              <div className="flex items-center gap-2 mt-1">
                <div className="w-8 h-8 rounded-full bg-surface-darker border border-border-color flex items-center justify-center text-xs font-bold text-white">
                  {currentDriver.name?.split(" ").map((n: string) => n[0]).join("").toUpperCase().substring(0, 2)}
                </div>
                <div>
                  <div className="text-sm font-bold text-white">{currentDriver.name}</div>
                  <div className="text-[10px] text-text-muted">{currentDriver.status === "active" ? "On duty" : currentDriver.status}</div>
                </div>
              </div>
            ) : (
              <div className="text-2xl font-bold text-text-muted mt-1">—</div>
            )}
          </div>
        </div>

        {/* Main Grid */}
        <div className="grid grid-cols-12 gap-6">
          {/* Left: Overview */}
          <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
            <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col">
              {/* Tabs */}
              <div className="flex items-center border-b border-border-color px-4">
                {(["overview", "documents"] as const).map((tab) => (
                  <button
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className={`px-4 py-3 text-sm font-medium transition-colors ${
                      activeTab === tab
                        ? "border-b-2 border-primary text-white"
                        : "border-b-2 border-transparent text-text-muted hover:text-white hover:border-border-color"
                    }`}
                  >
                    {tab === "overview" ? "Overview" : "Documents"}
                  </button>
                ))}
              </div>

              {activeTab === "overview" && (
                <div className="p-6">
                  {/* Vehicle Information */}
                  <h3 className="text-white text-sm font-semibold mb-4 flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary text-[18px]">info</span>
                    Vehicle Information
                  </h3>
                  <div className="grid grid-cols-2 gap-x-8 gap-y-6">
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Make & Model</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">
                        {vehicle.make} {vehicle.model}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Year</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">{vehicle.year || "—"}</div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Registration Number</label>
                      <div className="text-white text-sm font-mono border-b border-border-color pb-1">
                        {vehicle.registration_number}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">VIN</label>
                      <div className="text-white text-sm font-mono border-b border-border-color pb-1">
                        {vehicle.vin_number || "—"}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Fuel Type</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">{fuelLabel}</div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Ownership</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">
                        {vehicle.ownership_type ? vehicle.ownership_type.charAt(0) + vehicle.ownership_type.slice(1).toLowerCase() : "—"}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Color</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">{vehicle.color || "—"}</div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Registered</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">
                        {new Date(vehicle.created_at).toLocaleDateString()}
                      </div>
                    </div>
                  </div>

                  {/* Configuration */}
                  <h3 className="text-white text-sm font-semibold mt-8 mb-4 flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary text-[18px]">settings</span>
                    Configuration
                  </h3>
                  <div className="grid grid-cols-2 gap-x-8 gap-y-6">
                    <div className="space-y-1">
                      <div className="flex items-center justify-between">
                        <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Capacity</label>
                        <select
                          value={displayUnits.capacity}
                          onChange={(e) => setDisplayUnits((prev) => ({ ...prev, capacity: e.target.value }))}
                          className="text-[10px] bg-transparent text-text-muted border border-border-color rounded px-1 py-0.5 outline-none"
                        >
                          <option value="tonnes">tonnes</option>
                          <option value="kg">kg</option>
                          <option value="lbs">lbs</option>
                        </select>
                      </div>
                        {vehicle.capacity_tonnage ? `${fromTonnes(vehicle.capacity_tonnage, displayUnits.capacity)} ${getUnitLabel("capacity", displayUnits.capacity)}` : "—"}
                    </div>
                    <div className="space-y-1">
                      <div className="flex items-center justify-between">
                        <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">
                          {isCNG ? "CNG Cylinder" : "Fuel Tank"}
                        </label>
                        {!isCNG && (
                          <select
                            value={displayUnits.fuel}
                            onChange={(e) => setDisplayUnits((prev) => ({ ...prev, fuel: e.target.value }))}
                            className="text-[10px] bg-transparent text-text-muted border border-border-color rounded px-1 py-0.5 outline-none"
                          >
                            <option value="litres">litres</option>
                            <option value="gallons">gallons</option>
                          </select>
                        )}
                      </div>
                      <div className="text-white text-sm border-b border-border-color pb-1">
                        {vehicle.fuel_tank_capacity ? `${displayFuel} ${fuelUnitLabel}` : "—"}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">Speed Limit</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">
                        {vehicle.speed_limit ? `${vehicle.speed_limit} km/h` : "No limit"}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">GPS Device</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">
                        {vehicle.gps_device_id || "Not installed"}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] uppercase tracking-wider text-text-muted font-bold">AIS-140 Compliant</label>
                      <div className="text-white text-sm border-b border-border-color pb-1">
                        {vehicle.ais140_compliant ? "Yes" : "No"}
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Right: GPS + Hardware */}
          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
            {/* GPS Tracking */}
            <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col h-[320px]">
              <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker rounded-t-xl">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-primary text-[18px]">satellite_alt</span>
                  GPS Device Tracking
                </h3>
                {vehicle.gps_device_id ? (
                  <span className="flex h-2 w-2 relative">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-success"></span>
                  </span>
                ) : (
                  <span className="h-2 w-2 rounded-full bg-text-muted"></span>
                )}
              </div>
              <div className="relative flex-1 bg-surface-darker w-full h-full overflow-hidden">
                <div className="absolute inset-0 opacity-10 bg-[url('https://upload.wikimedia.org/wikipedia/commons/thumb/e/ec/World_map_blank_without_borders.svg/2000px-World_map_blank_without_borders.svg.png')] bg-cover bg-center grayscale invert"></div>
                {vehicle.gps_device_id ? (
                  <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
                    <div className="relative">
                      <div className="w-12 h-12 bg-primary/20 rounded-full animate-pulse absolute -top-3 -left-3"></div>
                      <span className="material-symbols-outlined text-primary text-2xl drop-shadow-lg">navigation</span>
                    </div>
                  </div>
                ) : (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="text-center">
                      <span className="material-symbols-outlined text-4xl text-text-muted opacity-30">signal_wifi_off</span>
                      <p className="text-text-muted text-xs mt-2">No GPS device installed</p>
                    </div>
                  </div>
                )}
                {vehicle.gps_device_id && (
                  <div className="absolute bottom-4 left-4 right-4 bg-surface-dark/90 backdrop-blur-sm border border-border-color p-3 rounded-lg shadow-lg">
                    <div className="flex items-start justify-between">
                      <div>
                        <div className="text-xs text-text-muted">Device ID</div>
                        <div className="text-sm font-bold text-white mt-1 font-mono">{vehicle.gps_device_id}</div>
                        <div className="text-[10px] text-text-muted mt-0.5">Last sync: {vehicle.updated_at ? new Date(vehicle.updated_at).toLocaleString() : "Never"}</div>
                      </div>
                      <button className="p-1.5 bg-primary rounded hover:bg-primary-dark text-white transition-colors">
                        <span className="material-symbols-outlined text-[16px]">near_me</span>
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Hardware / Compliance Info */}
            <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col flex-1">
              <div className="p-4 border-b border-border-color bg-surface-darker rounded-t-xl">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-text-muted text-[18px]">memory</span>
                  Device Details
                </h3>
              </div>
              <div className="p-4 space-y-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-surface-darker rounded border border-border-color">
                      <span className="material-symbols-outlined text-text-muted">router</span>
                    </div>
                    <div>
                      <div className="text-xs font-bold text-white">
                        {vehicle.gps_device_id ? "Telematics Unit" : "No Device"}
                      </div>
                      <div className="text-[10px] text-text-muted">
                        {vehicle.gps_device_id ? "AIS-140 Compliant" : "Install required"}
                      </div>
                    </div>
                  </div>
                  <span className={`text-[10px] px-2 py-0.5 rounded border ${
                    vehicle.gps_device_id
                      ? "bg-success/10 text-success border-success/20"
                      : "bg-text-muted/10 text-text-muted border-text-muted/20"
                  }`}>
                    {vehicle.gps_device_id ? "Online" : "Offline"}
                  </span>
                </div>
                <div className="border-t border-border-color pt-3 mt-2">
                  <div className="flex justify-between items-center text-xs mb-1">
                    <span className="text-text-muted">Immobilizer</span>
                    <span className={`font-medium ${vehicle.immobilizer_enabled ? "text-success" : "text-text-muted"}`}>
                      {vehicle.immobilizer_enabled ? "Enabled" : "Disabled"}
                    </span>
                  </div>
                  <div className={`w-full bg-border-color h-1 rounded-full overflow-hidden`}>
                    <div className={`h-full rounded-full ${vehicle.immobilizer_enabled ? "bg-success" : "bg-text-muted"}`} style={{ width: vehicle.immobilizer_enabled ? "100%" : "0%" }}></div>
                  </div>
                </div>
                <div className="pt-1">
                  <div className="flex justify-between items-center text-xs mb-1">
                    <span className="text-text-muted">Version</span>
                    <span className="text-white font-medium">v{vehicle.version || 1}</span>
                  </div>
                  <div className="flex justify-between items-center text-xs mt-2">
                    <span className="text-text-muted">Last Updated</span>
                    <span className="text-white font-mono text-[10px]">
                      {vehicle.updated_at ? new Date(vehicle.updated_at).toLocaleString() : "—"}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
        <span>© {new Date().getFullYear()} Enterprise Fleet Manager. Vehicle Module V2.4</span>
        <div className="flex items-center gap-4">
          <span>Last Sync: <span className="font-mono text-white">{vehicle.updated_at ? new Date(vehicle.updated_at).toLocaleString() : new Date().toLocaleTimeString()}</span></span>
          <span>System Status: <span className={systemStatus === "ok" ? "text-success" : "text-danger"}>{systemStatus === "ok" ? "Connected" : "Disconnected"}</span></span>
        </div>
      </footer>
    </div>
  );
};

export default VehicleDetail;
