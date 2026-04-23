import { WorkerManager } from "../../workers/utils/WorkerManager";
import { Vehicle } from "./types";
import apiClient from "../../services/apiClient";

/**
 * FleetService: Bridge between UI and OperationalWorker.
 * Handles hyperscale search and filtering requests.
 * Falls back to REST API when the worker is not available.
 */
class FleetService {
  private worker: WorkerManager | Worker | null = null;
  private pendingSearchResolver: ((results: any[]) => void) | null = null;

  constructor() {
    if (typeof window !== "undefined") {
      window.addEventListener("VEHICLE_SEARCH_RESULTS", (e: any) => {
        if (this.pendingSearchResolver) {
          this.pendingSearchResolver(e.detail);
          this.pendingSearchResolver = null;
        }
      });
    }
  }

  setWorker(worker: WorkerManager | Worker) {
    this.worker = worker;
  }

  /**
   * Performs a high-performance search using the operational worker.
   * Falls back to REST API when worker is not available OR returns empty.
   */
  async searchVehicles(
    query: string,
    filters?: Record<string, string>,
  ): Promise<any[]> {
    // If worker is available, try it first for fast client-side search
    if (this.worker) {
      const workerResults = await new Promise<any[]>((resolve) => {
        this.pendingSearchResolver = resolve;
        // Timeout after 2 seconds to avoid hanging
        const timeout = setTimeout(() => {
          this.pendingSearchResolver = null;
          resolve([]);
        }, 2000);

        const origResolver = this.pendingSearchResolver;
        this.pendingSearchResolver = (results: any[]) => {
          clearTimeout(timeout);
          origResolver?.(results);
        };

        this.worker?.postMessage(
          {
            type: "SEARCH",
            payload: { query, filters },
          },
          [],
        );
      });

      // If worker returned results, use them
      if (workerResults.length > 0) {
        return workerResults;
      }
      // Worker had no results — fall through to REST API
    }

    // Fallback: fetch vehicles from REST API
    try {
      const isSearch = query || filters?.status || filters?.org || filters?.gps_enabled;
      const endpoint = isSearch ? "/admin/vehicles/search" : "/admin/vehicles";

      const response = await apiClient.get(endpoint, {
        params: {
          page: 1,
          size: 100,
          ...(query ? { q: query } : {}),
          ...(filters?.status ? { status: filters.status } : {}),
          ...((filters?.tenant_id || filters?.org) ? { tenant_id: filters.tenant_id || filters.org } : {}),
          ...(filters?.gps_enabled ? { gps_enabled: filters.gps_enabled } : {}),
        },
      }) as any;

      // Support multiple response shapes:
      // 1. Stub handler: { data: [...] }
      // 2. Vehicle domain: { vehicles: [...], total, page, limit }
      // 3. Direct array: [...]
      const raw = response?.data ?? response;
      const vehicles = Array.isArray(raw)
        ? raw
        : raw?.vehicles
        || raw?.data
        || [];

      return vehicles.map((v: any) => {
        const registration = v.registration_number || v.registration || v.license_plate || v.plate_number;
        
        return {
          id: v.id,
          registrationNumber: registration || `Asset ${v.id?.substring(0, 8)}`,
          name: v.name || registration || `Asset ${v.id?.substring(0, 8)}`,
          status: (v.status || "ACTIVE").toUpperCase(),
          derived_status: (v.status || "ACTIVE").toLowerCase(),
          lastUpdate: v.updated_at || v.last_update || v.last_seen || new Date().toISOString(),
          latitude: v.latitude,
          longitude: v.longitude,
          metadata: {
            capacity: v.capacity_tonnage || v.capacity || v.metadata?.capacity || 0,
          },
          ...v,
        };
      });
    } catch (error) {
      console.error("[FleetService] REST API fallback failed:", error);
      return [];
    }
  }

  /**
   * Injects a batch of vehicles into the worker's index.
   */
  indexVehicles(vehicles: Partial<Vehicle>[]) {
    this.worker?.postMessage(
      {
        type: "INDEX_VEHICLES",
        payload: vehicles,
      },
      [],
    );
  }

  async getFilterMetadata(): Promise<{ orgs: string[]; statuses: string[] }> {
    // If worker is available, use it
    if (this.worker) {
      return new Promise((resolve) => {
        const timeout = setTimeout(() => {
          window.removeEventListener("VEHICLE_FILTER_METADATA", handler);
          resolve({ orgs: [], statuses: [] });
        }, 2000);

        function handler(e: any) {
          clearTimeout(timeout);
          window.removeEventListener("VEHICLE_FILTER_METADATA", handler);
          resolve(e.detail);
        };
        window.addEventListener("VEHICLE_FILTER_METADATA", handler);
        this.worker?.postMessage({ type: "GET_FILTER_METADATA" }, []);
      });
    }

    // Fallback: derive metadata from API response
    try {
      const response = await apiClient.get("/admin/vehicles", {
        params: { page: 1, size: 100 },
      }) as any;

      // Support multiple response shapes
      const raw = response?.data ?? response;
      const vehicles = Array.isArray(raw)
        ? raw
        : raw?.vehicles || raw?.data || [];

      const orgSet = new Set<string>();
      const statusSet = new Set<string>();
      vehicles.forEach((v: any) => {
        const tid = v.tenant_id || v.tenantId || v.org_id || v.org;
        if (tid) orgSet.add(String(tid));
        if (v.org_name) orgSet.add(v.org_name); // Legacy name field
        if (v.status) statusSet.add(v.status);
      });
      return { orgs: Array.from(orgSet), statuses: Array.from(statusSet) };
    } catch (error) {
      console.error("[FleetService] getFilterMetadata REST fallback failed:", error);
      return { orgs: [], statuses: [] };
    }
  }
}

export const fleetService = new FleetService();
