import api from "@/services/api";
import { Vehicle, VehicleSearchParams, VehicleListResponse } from "./types";

class FleetService {
  async searchVehicles(params: VehicleSearchParams): Promise<VehicleListResponse> {
    const response = await api.get("/admin/vehicles/search", {
      params: {
        ...params,
        ...(params.tenant_id ? { tenant_id: params.tenant_id } : {}),
      },
    });

    const data = response || {};
    const rawVehicles = data.vehicles || [];

    console.log('[FleetService] searchVehicles response:', {
      url: "/admin/vehicles/search",
      params,
      data,
      vehicleCount: rawVehicles.length
    });

    return {
      vehicles: rawVehicles.map((v: any) => this.mapVehicle(v)),
      total: data.total || 0,
      page: data.page || 1,
      limit: data.limit || 50,
    };
  }

  async getVehicle(id: string): Promise<Vehicle> {
    const response = await api.get(`/v1/vehicles/${id}`);
    return this.mapVehicle(response.data);
  }

  private mapVehicle(v: any): Vehicle {
    return {
      id: v.id,
      registrationNumber: v.registration_number,
      maskedRegistration: v.masked_registration,
      type: v.vehicle_type,
      status: v.status,
      tenantId: v.tenant_id,
      driverId: v.current_driver_id,
      lastUpdate: v.updated_at,
      make: v.make,
      model: v.model,
      year: v.year,
      fuelType: v.fuel_type,
      capacityTonnage: v.capacity_tonnage,
      fuelTankCapacity: v.fuel_tank_capacity,
      ownershipType: v.ownership_type,
      isBlacklisted: v.is_blacklisted,
      isHardBlocked: v.is_hard_blocked,
      riskScore: v.risk_score,
      gpsEnabled: v.gps_enabled,
      gpsDeviceId: v.gps_device_id,
      version: v.version,
      odometer: v.odometer,
      createdAt: v.created_at,
      documents: v.documents?.map((d: any) => ({
        id: d.id,
        type: d.type,
        documentNumber: d.document_number,
        expiryDate: d.expiry_date,
        status: d.status,
        version: d.version,
        documentUrl: d.document_url,
        verifiedAt: d.verified_at,
      })),
      summary: v.summary ? {
        complianceStatus: v.summary.compliance_status,
        complianceScore: v.summary.compliance_score || 0,
        complianceBreachCount: v.summary.compliance_breach_count || 0,
        gpsStatus: v.summary.gps_status,
        deviceHealthStatus: v.summary.device_health_status,
        ignitionStatus: v.summary.ignition_status,
        lastSeen: v.summary.last_seen,
        alertCount: v.summary.alert_count,
        criticalAlertCount: v.summary.critical_alert_count || 0,
        overSpeedCount: v.summary.over_speed_count,
        geoFenceViolation: v.summary.geofence_violation,
        activeTrip: v.summary.active_trip,
        lastTripEndTime: v.summary.last_trip_end_time,
        dataFreshnessScore: v.summary.data_freshness_score || 0,
        nextExpiryDate: v.summary.next_expiry_date,
        nextExpiryType: v.summary.next_expiry_type,
        packetLossPercentage: v.summary.packet_loss_percentage,
        telemetryReliability: v.summary.telemetry_reliability,
        maintenanceStatus: v.summary.maintenance_status,
        utilizationScore: v.summary.utilization_score || 0,
        fuelEfficiency30d: v.summary.fuel_efficiency_30d || 0,
        driverAssigned: v.summary.driver_assigned,
        driverLicenseStatus: v.summary.driver_license_status,
        riskScore: v.summary.risk_score,
        riskCategory: v.summary.risk_category,
        lastLocationLat: v.summary.last_location_lat,
        lastLocationLng: v.summary.last_location_lng,
        lastPingIP: v.summary.last_ping_ip,
        networkType: v.summary.network_type,
        isHardBlocked: v.summary.is_hard_blocked || false,
      } : undefined,
      currentDriver: v.current_driver,
    };
  }
}

export default new FleetService();
