import { WorkerManager } from "../../workers/utils/WorkerManager";
import { Vehicle } from "./types";

/**
 * FleetService: Bridge between UI and OperationalWorker.
 * Handles hyperscale search and filtering requests.
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
   */
  async searchVehicles(
    query: string,
    filters?: Record<string, string>,
  ): Promise<any[]> {
    if (!this.worker) return [];

    return new Promise((resolve) => {
      this.pendingSearchResolver = resolve;
      this.worker?.postMessage(
        {
          type: "SEARCH",
          payload: { query, filters },
        },
        [],
      );
    });
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
    if (!this.worker) return { orgs: [], statuses: [] };

    return new Promise((resolve) => {
      const handler = (e: any) => {
        window.removeEventListener("VEHICLE_FILTER_METADATA", handler);
        resolve(e.detail);
      };
      window.addEventListener("VEHICLE_FILTER_METADATA", handler);
      this.worker?.postMessage({ type: "GET_FILTER_METADATA" }, []);
    });
  }
}

export const fleetService = new FleetService();
