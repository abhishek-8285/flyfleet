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
import AddressInput from "../../../components/trips/AddressInput";
import RouteMap, { RouteMapPoint } from "../../../components/trips/RouteMap";
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

function calculateDistance(a: RouteMapPoint, b: RouteMapPoint): number {
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
    title: "",
    start_location: "",
    end_location: "",
    scheduled_start: "",
    planned_eta: "",
    description: "",
    customer_name: "",
    customer_phone: "",
    customer_email: "",
    cargo_type: "general",
    cargo_weight: "",
    cargo_value: "",
    cargo_description: "",
    planned_distance: "",
    estimated_duration_min: "",
  });

  const [timeError, setTimeError] = useState("");

  // Map locations
  const [pickup, setPickup] = useState<RouteMapPoint | null>(null);
  const [drop, setDrop] = useState<RouteMapPoint | null>(null);
  const [roadDistance, setRoadDistance] = useState<number | null>(null);
  const [roadDuration, setRoadDuration] = useState<number | null>(null);

  const handleRouteCalculated = (distanceKm: number, durationMin: number) => {
    setRoadDistance(distanceKm);
    setRoadDuration(durationMin);
    // Auto populate the input fields with the real OSRM road distance and duration if user hasn't explicitly overridden them
    setFormData((prev) => ({
      ...prev,
      planned_distance: prev.planned_distance ? prev.planned_distance : distanceKm.toFixed(1),
      estimated_duration_min: prev.estimated_duration_min ? prev.estimated_duration_min : durationMin.toString(),
    }));
  };

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
  const handlePickupChange = (pt: RouteMapPoint) => {
    setPickup(pt);
    if (!pt || pt.lat === 0) {
      setRoadDistance(null);
      setRoadDuration(null);
    }
    setFormData((p) => ({
      ...p,
      start_location: pt.address || `${pt.lat.toFixed(5)}, ${pt.lng.toFixed(5)}`,
    }));
  };

  const handleDropChange = (pt: RouteMapPoint) => {
    setDrop(pt);
    if (!pt || pt.lat === 0) {
      setRoadDistance(null);
      setRoadDuration(null);
    }
    setFormData((p) => ({
      ...p,
      end_location: pt.address || `${pt.lat.toFixed(5)}, ${pt.lng.toFixed(5)}`,
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
    if (!formData.title) {
      setError("Trip title is required");
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
    if (!formData.customer_name) {
      setError("Customer name is required");
      return;
    }
    if (!formData.customer_phone) {
      setError("Customer phone is required");
      return;
    }
    if (formData.planned_eta && !validateTimes(formData.scheduled_start, formData.planned_eta)) {
      setError("End time must be after start time");
      return;
    }

    setLoading(true);

    try {
      // Calculate distance from map if available, else estimate
      const distance = parseFloat(formData.planned_distance) ||
        (roadDistance !== null ? roadDistance : (pickup && drop ? calculateDistance(pickup, drop) : 100));

      // Calculate duration in minutes from times if available
      let durationMin = parseInt(formData.estimated_duration_min) || roadDuration || 0;
      if (!durationMin && formData.scheduled_start && formData.planned_eta) {
        const startMs = new Date(formData.scheduled_start).getTime();
        const endMs = new Date(formData.planned_eta).getTime();
        durationMin = Math.round((endMs - startMs) / 60000);
      }
      if (!durationMin) {
        // Default: assume 60 km/h average speed
        durationMin = Math.round((distance / 60) * 60);
      }

      // Build waypoints from map locations or text inputs
      const waypoints = [
        {
          name: "Origin",
          type: "pickup",
          address: formData.start_location,
          sequence: 1,
          location: pickup && pickup.lat !== 0
            ? { latitude: pickup.lat, longitude: pickup.lng, address: pickup.address || "" }
            : { latitude: 0, longitude: 0, address: formData.start_location },
        },
        {
          name: "Destination",
          type: "dropoff",
          address: formData.end_location,
          sequence: 2,
          location: drop && drop.lat !== 0
            ? { latitude: drop.lat, longitude: drop.lng, address: drop.address || "" }
            : { latitude: 0, longitude: 0, address: formData.end_location },
        },
      ];

      const payload = {
        vehicle_id: formData.vehicle_id,
        driver_id: formData.driver_id || undefined,
        trip_type: formData.trip_type === "pickup" ? "local" : formData.trip_type === "maintenance" ? "service" : formData.trip_type === "internal" ? "standard" : formData.trip_type,
        title: formData.title,
        description: formData.description || "",
        planned_route: waypoints,
        waypoints: waypoints,
        planned_distance: distance,
        estimated_duration_min: durationMin,
        cargo_details: {
          type: formData.cargo_type || "general",
          weight_kg: parseFloat(formData.cargo_weight) || 0,
          volume_m3: 0,
          value_inr: parseFloat(formData.cargo_value) || 0,
          description: formData.cargo_description || "",
          fragile: formData.cargo_type === "fragile",
          temperature: formData.cargo_type === "perishable" ? "cold" : "ambient",
        },
        scheduled_start: new Date(formData.scheduled_start).toISOString(),
        customer_name: formData.customer_name,
        customer_phone: (formData.customer_phone.startsWith("+") ? formData.customer_phone : `+91${formData.customer_phone}`).replace(/\s+/g, ""),
        customer_email: formData.customer_email || undefined,
        planned_eta: formData.planned_eta
          ? new Date(formData.planned_eta).toISOString()
          : undefined,
        pickup_geo: pickup && pickup.lat !== 0
          ? { latitude: pickup.lat, longitude: pickup.lng, address: pickup.address || "" }
          : undefined,
        drop_geo: drop && drop.lat !== 0
          ? { latitude: drop.lat, longitude: drop.lng, address: drop.address || "" }
          : undefined,
      };

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

            {/* Trip Title */}
            <div className="space-y-1.5 md:col-span-2">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Trip Title <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">edit_note</span>
                </span>
                <input
                  type="text"
                  name="title"
                  value={formData.title}
                  onChange={handleChange}
                  required
                  maxLength={100}
                  placeholder="e.g. Delhi → Mumbai Freight Run"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
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
              <span className="material-symbols-outlined text-sm">route</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Route & Schedule</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Search addresses or pin them on the map
              </p>
            </div>
          </div>
          <div className="p-6">
            {/* Address Search Fields */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <AddressInput
                label="Pickup Address"
                type="pickup"
                value={formData.start_location}
                onLocationSelected={(lat, lng, addr) => {
                  setPickup({ lat, lng, address: addr });
                  setFormData((p) => ({ ...p, start_location: addr }));
                }}
              />
              <AddressInput
                label="Drop Address"
                type="drop"
                value={formData.end_location}
                onLocationSelected={(lat, lng, addr) => {
                  setDrop({ lat, lng, address: addr });
                  setFormData((p) => ({ ...p, end_location: addr }));
                }}
              />
            </div>



            {/* Route Summary */}
            {pickup && pickup.lat !== 0 && drop && drop.lat !== 0 && (
              <div className="mt-4 bg-surface-darker border border-border-color rounded-xl p-4">
                <div className="flex items-center gap-4">
                  {/* Route visual */}
                  <div className="flex flex-col items-center">
                    <span className="w-3 h-3 rounded-full bg-emerald-500 ring-4 ring-emerald-500/20" />
                    <div className="w-0.5 h-8 bg-gradient-to-b from-emerald-500 to-amber-500" />
                    <span className="w-3 h-3 rounded-full bg-amber-500 ring-4 ring-amber-500/20" />
                  </div>
                  {/* Details */}
                  <div className="flex-1 min-w-0">
                    <p className="text-xs text-white truncate font-medium">{pickup.address || `${pickup.lat.toFixed(4)}, ${pickup.lng.toFixed(4)}`}</p>
                    <div className="flex items-center gap-3 mt-1">
                      <span className="text-xs text-text-muted font-mono">
                        📏 {roadDistance !== null ? roadDistance.toFixed(1) : calculateDistance(pickup, drop).toFixed(1)} km
                      </span>
                      {roadDuration !== null ? (
                        <span className="text-xs text-text-muted font-mono">
                          ⏱ {roadDuration} min
                        </span>
                      ) : formData.scheduled_start && formData.planned_eta && (
                        <span className="text-xs text-text-muted">
                          ⏱ {Math.round((new Date(formData.planned_eta).getTime() - new Date(formData.scheduled_start).getTime()) / 60000)} min
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-white truncate font-medium mt-1">{drop.address || `${drop.lat.toFixed(4)}, ${drop.lng.toFixed(4)}`}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Times */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-6">
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
              </div>
            </div>

            {/* Planned Distance & Duration */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-6 pt-6 border-t border-border-color">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  Planned Distance (km)
                </label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">straighten</span>
                  </span>
                  <input
                    type="number"
                    name="planned_distance"
                    value={formData.planned_distance}
                    onChange={handleChange}
                    placeholder={pickup && drop && pickup.lat !== 0 && drop.lat !== 0
                      ? `Auto: ${calculateDistance(pickup, drop).toFixed(1)} km`
                      : "e.g. 450"}
                    min="0"
                    step="0.1"
                    className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                  />
                  {pickup && drop && pickup.lat !== 0 && drop.lat !== 0 && (
                    <span className="absolute right-2 top-2 text-[10px] text-primary bg-primary/10 px-2 py-0.5 rounded-full">
                      Auto
                    </span>
                  )}
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  Est. Duration (minutes)
                </label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">hourglass_top</span>
                  </span>
                  <input
                    type="number"
                    name="estimated_duration_min"
                    value={formData.estimated_duration_min}
                    onChange={handleChange}
                    placeholder="e.g. 360"
                    min="1"
                    className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                  />
                </div>
              </div>
            </div>

            {/* Real-world Adjustments Notice */}
            <div className="mt-6 p-4 rounded-xl border border-warning/15 bg-warning/5 text-text-muted flex items-start gap-3">
              <span className="material-symbols-outlined text-warning text-lg mt-0.5">info</span>
              <p className="text-xs leading-relaxed text-text-muted">
                <span className="font-bold text-white">Note:</span> Although route distance and travel duration are calculated automatically based on geographic coordinates, you can manually adjust them to reflect real-world logistics scenarios (e.g., custom local shortcuts, traffic congestion, or required driver rest stops).
              </p>
            </div>
          </div>
        </div>

        {/* ── Customer / Party ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-cyan-500/10 flex items-center justify-center text-cyan-400 border border-cyan-500/20">
              <span className="material-symbols-outlined text-sm">group</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Customer / Party</h3>
              <p className="text-xs text-text-muted mt-0.5">
                The organization or party giving this job
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Customer Name <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">person</span>
                </span>
                <input
                  type="text"
                  name="customer_name"
                  value={formData.customer_name}
                  onChange={handleChange}
                  required
                  placeholder="e.g. Acme Logistics Pvt Ltd"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Customer Phone <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">phone</span>
                </span>
                <input
                  type="tel"
                  name="customer_phone"
                  value={formData.customer_phone}
                  onChange={handleChange}
                  required
                  placeholder="+91 98765 43210"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Customer Email
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">mail</span>
                </span>
                <input
                  type="email"
                  name="customer_email"
                  value={formData.customer_email}
                  onChange={handleChange}
                  placeholder="dispatch@acme.com"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
          </div>
        </div>

        {/* ── Cargo Details ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-400 border border-amber-500/20">
              <span className="material-symbols-outlined text-sm">inventory_2</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Cargo Details</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Shipment information and handling requirements
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Cargo Type <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">category</span>
                </span>
                <select
                  name="cargo_type"
                  value={formData.cargo_type}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                >
                  <option value="general">General Cargo</option>
                  <option value="fragile">Fragile</option>
                  <option value="hazardous">Hazardous</option>
                  <option value="perishable">Perishable / Cold Chain</option>
                  <option value="oversized">Oversized</option>
                  <option value="liquid">Liquid / Bulk</option>
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Weight (kg)
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">scale</span>
                </span>
                <input
                  type="number"
                  name="cargo_weight"
                  value={formData.cargo_weight}
                  onChange={handleChange}
                  placeholder="e.g. 2500"
                  min="0"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Value (₹)
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">payments</span>
                </span>
                <input
                  type="number"
                  name="cargo_value"
                  value={formData.cargo_value}
                  onChange={handleChange}
                  placeholder="e.g. 50000"
                  min="0"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Cargo Description
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">description</span>
                </span>
                <input
                  type="text"
                  name="cargo_description"
                  value={formData.cargo_description}
                  onChange={handleChange}
                  placeholder="e.g. Electronic components, palletized"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
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
