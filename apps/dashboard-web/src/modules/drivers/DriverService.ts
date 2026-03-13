import { Driver, DriverSession, SafetyScore } from "./types";

/**
 * DriverService: Orchestrates driver identity and session lifecycle.
 */
class DriverService {
  private activeSessions: Map<string, DriverSession> = new Map();

  /**
   * Binds a driver to a specific vehicle for a telemetry session.
   */
  async startSession(
    driverId: string,
    vehicleId: string,
  ): Promise<DriverSession> {
    const session: DriverSession = {
      id: `sess_${Date.now()}`,
      driverId,
      vehicleId,
      startTime: new Date().toISOString(),
      startLocation: { lat: 0, lng: 0 }, // In production, this comes from the IoT Gateway
    };

    this.activeSessions.set(driverId, session);
    // In production, we would call the Backend Driver Service here
    return session;
  }

  async endSession(driverId: string): Promise<void> {
    const session = this.activeSessions.get(driverId);
    if (session) {
      session.endTime = new Date().toISOString();
      this.activeSessions.delete(driverId);
    }
  }

  /**
   * Calculates real-time safety metrics by aggregating harsh events.
   */
  async getSafetyScore(driverId: string): Promise<SafetyScore> {
    // Placeholder for telemetry aggregation logic
    return {
      driverId,
      score: 95.5,
      events: {
        harshBraking: 2,
        harshTurning: 0,
        overspeeding: 1,
      },
      lastUpdated: new Date().toISOString(),
    };
  }

  async getDriver(id: string): Promise<Driver | null> {
    // Mock for registry lookup
    return {
      id,
      name: "John Doe",
      orgId: "flyfleet-internal",
      status: "ON_DUTY",
      licenseNumber: "CDL-987234",
      contactNumber: "+1 (555) 902-1234",
      rating: 4.8,
    };
  }
}

export const driverService = new DriverService();
