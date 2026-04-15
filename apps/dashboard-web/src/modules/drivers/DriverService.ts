import { api } from "../../services/api";
import {
  Driver,
  DriverListResponse,
  DriverDocument,
  BehaviorEvent,
  DriverAnalytics,
} from "./types";

/**
 * DriverService: Real API integration for driver management.
 */
class DriverService {
  /**
   * Fetch paginated list of drivers with filters.
   */
  async getDrivers(params?: {
    page?: number;
    limit?: number;
    status?: string;
    name?: string;
    vehicle_id?: string;
    order_by?: string;
    order_dir?: string;
  }): Promise<DriverListResponse> {
    const response = await api.getDrivers(params);
    return response;
  }

  /**
   * Get a single driver by ID.
   */
  async getDriver(id: string): Promise<Driver> {
    const response = await api.getDriver(id);
    return response.data;
  }

  /**
   * Create a new driver.
   */
  async createDriver(data: {
    first_name: string;
    last_name: string;
    email: string;
    phone: string;
    license_number: string;
    license_type: string;
    license_expiry: string;
    license_authority: string;
    date_of_birth: string;
    employment_type: string;
    address: {
      street: string;
      city: string;
      state: string;
      pincode: string;
      country: string;
    };
    emergency_contact: {
      name: string;
      phone: string;
      relation: string;
    };
    employee_code?: string;
    aadhaar_number?: string;
    pan_number?: string;
    license_issue_date?: string;
    experience_years?: number;
    metadata?: Record<string, unknown>;
  }): Promise<Driver> {
    const response = await api.createDriver(data);
    return response.data;
  }

  /**
   * Update an existing driver.
   */
  async updateDriver(
    id: string,
    data: Partial<{
      first_name: string;
      last_name: string;
      email: string;
      phone: string;
      license_number: string;
      license_type: string;
      license_expiry: string;
      license_authority: string;
      date_of_birth: string;
      employment_type: string;
      address: Record<string, unknown>;
      emergency_contact: Record<string, unknown>;
      employee_code: string;
      experience_years: number;
      metadata: Record<string, unknown>;
    }>,
  ): Promise<Driver> {
    const response = await api.updateDriver(id, data);
    return response.data;
  }

  /**
   * Delete a driver (soft delete).
   */
  async deleteDriver(id: string): Promise<void> {
    await api.deleteDriver(id);
  }

  /**
   * Update driver status.
   */
  async updateStatus(id: string, status: string): Promise<void> {
    await api.updateDriverStatus(id, status);
  }

  /**
   * Assign a vehicle to a driver.
   */
  async assignVehicle(driverId: string, vehicleId: string): Promise<void> {
    await api.assignVehicle(driverId, vehicleId);
  }

  /**
   * Unassign vehicle from a driver.
   */
  async unassignVehicle(driverId: string): Promise<void> {
    await api.unassignVehicle(driverId);
  }

  /**
   * Get driver analytics and behavior data.
   */
  async getAnalytics(id: string): Promise<DriverAnalytics> {
    const response = await api.getDriverAnalytics(id);
    return response;
  }

  /**
   * Get driver behavior score.
   */
  async getBehaviorScore(id: string): Promise<{
    driver_id: string;
    behavior_score: number;
    max_score: number;
  }> {
    const response = await api.getDriverBehaviorScore(id);
    return response;
  }

  /**
   * List driver documents.
   */
  async getDocuments(id: string): Promise<DriverDocument[]> {
    const response = await api.getDriverDocuments(id);
    return response.data || [];
  }

  /**
   * Upload a driver document.
   */
  async uploadDocument(
    driverId: string,
    data: {
      document_type: string;
      document_number: string;
      issuing_authority: string;
      issue_date: string;
      expiry_date: string;
      file_url: string;
    },
  ): Promise<DriverDocument> {
    const response = await api.uploadDriverDocument(driverId, data);
    return response.data;
  }

  /**
   * Delete a driver document.
   */
  async deleteDocument(driverId: string, docId: string): Promise<void> {
    await api.deleteDriverDocument(driverId, docId);
  }
}

export const driverService = new DriverService();
