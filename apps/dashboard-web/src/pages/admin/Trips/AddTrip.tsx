import {
  ArrowLeftIcon,
  ClockIcon,
  MapPinIcon,
  TruckIcon,
  UserIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import DocumentUploader from "../../../components/trips/DocumentUploader";
import LocationPickerMap, { MapLocation } from "../../../components/trips/LocationPickerMap";
import { useUIStrings } from "../../../hooks/useUIStrings";
import api from "../../../services/api";

interface VehicleOption {
  id: string;
  registration_number?: string;
  plate_number?: string;
  name?: string;
}

interface DriverOption {
  id: string;
  name: string;
  license_number?: string;
}

function calculateDistance(a: MapLocation, b: MapLocation): number {
  const R = 6371;
  const dLat = ((b.lat - a.lat) * Math.PI) / 180;
  const dLon = ((b.lng - a.lng) * Math.PI) / 180;
  const x =
    Math.sin(dLat / 2) ** 2 +
    Math.cos((a.lat * Math.PI) / 180) *
      Math.cos((b.lat * Math.PI) / 180) *
      Math.sin(dLon / 2) ** 2;
  return R * 2 * Math.atan2(Math.sqrt(x), Math.sqrt(1 - x));
}

export default function AddTrip() {
  const navigate = useNavigate();
  const { uiStrings: rawUiStrings } = useUIStrings("tracking");
  const uiStrings = rawUiStrings || {};

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [vehicles, setVehicles] = useState<VehicleOption[]>([]);
  const [drivers, setDrivers] = useState<DriverOption[]>([]);
  const [vehicleSearch, setVehicleSearch] = useState("");
  const [driverSearch, setDriverSearch] = useState("");
  const [showVehicleDropdown, setShowVehicleDropdown] = useState(false);
  const [showDriverDropdown, setShowDriverDropdown] = useState(false);
  const [charCount, setCharCount] = useState(0);

  const [formData, setFormData] = useState({
    vehicle_id: "",
    driver_id: "",
    org_id: "",
    trip_type: "delivery",
    start_location: "",
    end_location: "",
    scheduled_start: "",
    planned_eta: "",
    description: "",
  });

  const [timeError, setTimeError] = useState("");

  // Map locations
  const [startLoc, setStartLoc] = useState<MapLocation | null>(null);
  const [endLoc, setEndLoc] = useState<MapLocation | null>(null);

  // Documents
  const [uploadedDocs, setUploadedDocs] = useState<
    { file: File; preview: string; type: "image" | "pdf" | "other" }[]
  >([]);

  useEffect(() => {
    loadOptions();
  }, []);

  const loadOptions = async () => {
    try {
      const [vRes, dRes] = await Promise.all([
        api.getVehicles({ limit: 500 }),
        api.getDrivers({ limit: 500 }),
      ]);
      setVehicles(vRes.vehicles || vRes.data || []);
      setDrivers(dRes.drivers || dRes.data || []);
    } catch (err) {
      console.error("Failed to load vehicles/drivers:", err);
    }
  };

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));

    if (name === "description") {
      setCharCount(value.length);
    }

    // Validate times when either changes
    if (name === "scheduled_start" || name === "planned_eta") {
      validateTimes(
        name === "scheduled_start" ? value : formData.scheduled_start,
        name === "planned_eta" ? value : formData.planned_eta,
      );
    }
  };

  const validateTimes = (start: string, end: string) => {
    if (start && end) {
      const startDate = new Date(start);
      const endDate = new Date(end);
      if (endDate <= startDate) {
        setTimeError("End time must be after start time.");
        return false;
      }
    }
    setTimeError("");
    return true;
  };

  const selectVehicle = (v: VehicleOption) => {
    setFormData((prev) => ({ ...prev, vehicle_id: v.id }));
    setVehicleSearch(getVehicleLabel(v));
    setShowVehicleDropdown(false);
  };

  const selectDriver = (d: DriverOption) => {
    setFormData((prev) => ({ ...prev, driver_id: d.id }));
    setDriverSearch(d.name);
    setShowDriverDropdown(false);
  };

  const getVehicleLabel = (v: VehicleOption) =>
    v.registration_number || v.plate_number || v.name || v.id.slice(0, 8);

  // Sync map location to text input
  const handleStartLocChange = (loc: MapLocation) => {
    setStartLoc(loc);
    setFormData((p) => ({
      ...p,
      start_location: loc.address || `${loc.lat.toFixed(5)}, ${loc.lng.toFixed(5)}`,
    }));
  };

  const handleEndLocChange = (loc: MapLocation) => {
    setEndLoc(loc);
    setFormData((p) => ({
      ...p,
      end_location: loc.address || `${loc.lat.toFixed(5)}, ${loc.lng.toFixed(5)}`,
    }));
  };

  const filteredVehicles = useMemo(() => {
    if (!vehicleSearch) return vehicles;
    const q = vehicleSearch.toLowerCase();
    return vehicles.filter((v) => getVehicleLabel(v).toLowerCase().includes(q));
  }, [vehicles, vehicleSearch]);

  const filteredDrivers = useMemo(() => {
    if (!driverSearch) return drivers;
    const q = driverSearch.toLowerCase();
    return drivers.filter((d) => d.name.toLowerCase().includes(q));
  }, [drivers, driverSearch]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!formData.vehicle_id) {
      setError("Vehicle is required");
      return;
    }
    if (!formData.start_location || !formData.end_location) {
      setError("Start and end locations are required");
      return;
    }
    if (!formData.scheduled_start) {
      setError("Start time is required");
      return;
    }
    if (formData.planned_eta && !validateTimes(formData.scheduled_start, formData.planned_eta)) {
      setError("End time must be after start time");
      return;
    }

    setLoading(true);

    try {
      const payload: Record<string, unknown> = {
        vehicle_id: formData.vehicle_id,
        driver_id: formData.driver_id || undefined,
        start_location: formData.start_location,
        end_location: formData.end_location,
        scheduled_start: new Date(formData.scheduled_start).toISOString(),
        description: formData.description || undefined,
        planned_eta: formData.planned_eta
          ? new Date(formData.planned_eta).toISOString()
          : undefined,
      };

      if (formData.trip_type) {
        payload.trip_type = formData.trip_type;
      }

      await api.createTrip(payload);
      navigate("/admin/trips");
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        "Failed to create trip";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <Link to="/admin/trips" className="hover:text-white">
              {uiStrings.trips_title || "Trips"}
            </Link>
            <span>/</span>
            <span className="text-primary font-medium">
              {uiStrings.breadcrumb_add || "Add"}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-white">
            {uiStrings.title_add_trip || "Add New Trip"}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/admin/trips"
            className="px-4 py-2 bg-surface-dark hover:bg-surface-darker border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded transition-colors"
          >
            Cancel
          </Link>
          <button
            onClick={() => handleSubmit(new Event("submit") as unknown as React.FormEvent)}
            disabled={loading}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20 disabled:opacity-50"
          >
            <span className="material-symbols-outlined text-[16px]">add</span>
            {loading ? "Creating..." : "Create Trip"}
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-danger/10 border border-danger/20 rounded-lg flex items-center gap-2 text-danger text-sm">
          <XMarkIcon className="w-5 h-5" />
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* ── Trip Essentials ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
              <span className="material-symbols-outlined text-sm">local_shipping</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Trip Essentials</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Basic identification details for the assignment
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Vehicle ID */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Vehicle ID <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">directions_car</span>
                </span>
                <input
                  type="text"
                  value={vehicleSearch}
                  onChange={(e) => {
                    setVehicleSearch(e.target.value);
                    setFormData((p) => ({ ...p, vehicle_id: "" }));
                    setShowVehicleDropdown(true);
                  }}
                  onFocus={() => setShowVehicleDropdown(true)}
                  placeholder="Select or enter vehicle ID"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
                <button
                  type="button"
                  onClick={() => setShowVehicleDropdown(true)}
                  className="absolute inset-y-0 right-0 flex items-center pr-3 text-text-muted hover:text-white"
                >
                  <span className="material-symbols-outlined text-[18px]">search</span>
                </button>
                {showVehicleDropdown && (
                  <div className="absolute top-full left-0 right-0 mt-1 bg-surface-darker border border-border-color rounded-lg shadow-xl z-30 max-h-48 overflow-y-auto">
                    {filteredVehicles.length === 0 ? (
                      <div className="px-4 py-3 text-xs text-text-muted">No vehicles found</div>
                    ) : (
                      filteredVehicles.map((v) => (
                        <button
                          key={v.id}
                          type="button"
                          onClick={() => selectVehicle(v)}
                          className="w-full px-4 py-2.5 text-left text-sm text-white hover:bg-border-color flex items-center gap-2"
                        >
                          <span className="material-symbols-outlined text-text-muted text-sm">
                            directions_car
                          </span>
                          {getVehicleLabel(v)}
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Driver ID */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Driver ID
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">badge</span>
                </span>
                <input
                  type="text"
                  value={driverSearch}
                  onChange={(e) => {
                    setDriverSearch(e.target.value);
                    setFormData((p) => ({ ...p, driver_id: "" }));
                    setShowDriverDropdown(true);
                  }}
                  onFocus={() => setShowDriverDropdown(true)}
                  placeholder="Select assigned driver"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
                {showDriverDropdown && (
                  <div className="absolute top-full left-0 right-0 mt-1 bg-surface-darker border border-border-color rounded-lg shadow-xl z-30 max-h-48 overflow-y-auto">
                    {filteredDrivers.length === 0 ? (
                      <div className="px-4 py-3 text-xs text-text-muted">No drivers found</div>
                    ) : (
                      filteredDrivers.map((d) => (
                        <button
                          key={d.id}
                          type="button"
                          onClick={() => selectDriver(d)}
                          className="w-full px-4 py-2.5 text-left text-sm text-white hover:bg-border-color flex items-center gap-2"
                        >
                          <span className="material-symbols-outlined text-text-muted text-sm">
                            person
                          </span>
                          {d.name}
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Trip Type */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Trip Type
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">category</span>
                </span>
                <select
                  name="trip_type"
                  value={formData.trip_type}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                >
                  <option value="delivery">Freight Delivery</option>
                  <option value="pickup">Client Pickup</option>
                  <option value="maintenance">Maintenance Run</option>
                  <option value="internal">Internal Transfer</option>
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>

            {/* Organization ID */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Organization ID
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">domain</span>
                </span>
                <input
                  type="text"
                  name="org_id"
                  value={formData.org_id}
                  onChange={handleChange}
                  placeholder="Auto-assigned from your tenant context"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-text-muted outline-none"
                  disabled
                />
              </div>
            </div>
          </div>
        </div>

        {/* ── Route & Schedule ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-warning/10 flex items-center justify-center text-warning border border-warning/20">
              <span className="material-symbols-outlined text-sm">map</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Route & Schedule</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Locations and timing constraints
              </p>
            </div>
          </div>
          <div className="p-6">
            {/* Locations */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  Start Location <span className="text-danger">*</span>
                </label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">trip_origin</span>
                  </span>
                  <input
                    type="text"
                    name="start_location"
                    value={formData.start_location}
                    onChange={handleChange}
                    required
                    placeholder="Enter origin address"
                    className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                  />
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  End Location <span className="text-danger">*</span>
                </label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">location_on</span>
                  </span>
                  <input
                    type="text"
                    name="end_location"
                    value={formData.end_location}
                    onChange={handleChange}
                    required
                    placeholder="Enter destination address"
                    className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                  />
                </div>
              </div>
            </div>

            {/* Times */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 relative">
              <div className="hidden md:block absolute left-1/2 top-0 bottom-0 w-px border-l border-dashed border-border-color -translate-x-1/2" />
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  Start Time <span className="text-danger">*</span>
                </label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">schedule</span>
                  </span>
                  <input
                    type="datetime-local"
                    name="scheduled_start"
                    value={formData.scheduled_start}
                    onChange={handleChange}
                    required
                    className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none [color-scheme:dark]"
                  />
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  End Time
                </label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">event_available</span>
                  </span>
                  <input
                    type="datetime-local"
                    name="planned_eta"
                    value={formData.planned_eta}
                    onChange={handleChange}
                    className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none [color-scheme:dark]"
                  />
                </div>
                {timeError && (
                  <p className="text-[10px] text-danger flex items-center gap-1 mt-1">
                    <span className="material-symbols-outlined text-[12px]">error</span>
                    {timeError}
                  </p>
                )}
                {!timeError && (
                  <p className="text-[10px] text-text-muted flex items-center gap-1 mt-1">
                    <span className="material-symbols-outlined text-[12px] text-warning">info</span>
                    End time must be after start time.
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* ── Route on Map ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
              <span className="material-symbols-outlined text-sm">map</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Route on Map</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Click the map to pin start and end locations
              </p>
            </div>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-primary" />
                  <span className="text-xs font-medium text-text-muted uppercase tracking-wide">
                    Start Location
                  </span>
                </div>
                <LocationPickerMap
                  label="Pick Start"
                  color="primary"
                  value={startLoc}
                  onChange={handleStartLocChange}
                  height="h-56"
                />
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-warning" />
                  <span className="text-xs font-medium text-text-muted uppercase tracking-wide">
                    End Location
                  </span>
                </div>
                <LocationPickerMap
                  label="Pick End"
                  color="warning"
                  value={endLoc}
                  onChange={handleEndLocChange}
                  height="h-56"
                />
              </div>
            </div>

            {/* Route Preview */}
            {startLoc && endLoc && (
              <div className="mt-4 bg-surface-darker border border-border-color rounded-lg p-4">
                <div className="flex items-center gap-3">
                  <div className="flex flex-col items-center">
                    <span className="w-3 h-3 rounded-full bg-primary" />
                    <div className="w-px h-6 bg-border-color" />
                    <span className="w-3 h-3 rounded-full bg-warning" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs text-white truncate">{startLoc.address || `${startLoc.lat.toFixed(4)}, ${startLoc.lng.toFixed(4)}`}</p>
                    <p className="text-[10px] text-text-muted mt-1">↓ {calculateDistance(startLoc, endLoc).toFixed(1)} km (straight line)</p>
                    <p className="text-xs text-white truncate mt-1">{endLoc.address || `${endLoc.lat.toFixed(4)}, ${endLoc.lng.toFixed(4)}`}</p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* ── Additional Info ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-success/10 flex items-center justify-center text-success border border-success/20">
              <span className="material-symbols-outlined text-sm">description</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Additional Info</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Optional details and special instructions
              </p>
            </div>
          </div>
          <div className="p-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Remarks / Notes
              </label>
              <textarea
                name="description"
                value={formData.description}
                onChange={handleChange}
                placeholder="Add any special handling instructions, contact numbers, or cargo details..."
                rows={4}
                maxLength={500}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-3 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none resize-none placeholder-text-muted/50"
              />
              <div className="flex justify-between items-center mt-2">
                <span className="text-[10px] text-text-muted">Maximum 500 characters</span>
                <span className="text-[10px] text-text-muted">
                  {charCount}/500
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* ── Documents ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-purple-500/10 flex items-center justify-center text-purple-400 border border-purple-500/20">
              <span className="material-symbols-outlined text-sm">attach_file</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Trip Documents</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Upload invoices, e-way bills, loading slips, or photos
              </p>
            </div>
          </div>
          <div className="p-6">
            <DocumentUploader maxFiles={5} maxSizeMB={10} onChange={setUploadedDocs} />
          </div>
        </div>
      </form>
    </div>
  );
}
