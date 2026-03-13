import * as h3 from "h3-js";

/**
 * H3StreamingSimulator: Simulates the proposed backend protocol for
 * hyperscale telemetry streaming using H3 spatial indexing.
 */

export interface SimulationVehicle {
  id: string;
  lat: number;
  lng: number;
  org: string;
  status: string;
}

export class H3StreamingSimulator {
  private vehicles: SimulationVehicle[] = [];
  private activeCells: Set<string> = new Set();
  private currentResolution: number = 7;
  private tps: number = 10000; // Default 10k/sec
  private burstActive: boolean = false;
  private burstCount: number = 0;
  private latencyMs: number = 0;
  private onUpdate: (updates: SimulationVehicle[]) => void;

  constructor(onUpdate: (updates: SimulationVehicle[]) => void) {
    this.onUpdate = onUpdate;
    this.initializeFleet(10000000); // 10M vehicles in "backend" memory
  }

  private initializeFleet(count: number) {
    const orgs = [
      "North Fleet",
      "South Logistics",
      "West Express",
      "East Cargo",
    ];
    const statuses = ["active", "idle", "delayed", "maintenance"];

    for (let i = 0; i < count; i++) {
      this.vehicles.push({
        id: `REAL-${i}`,
        lat: 20.5937 + (Math.random() - 0.5) * 10, // Wider India-wide spread
        lng: 78.9629 + (Math.random() - 0.5) * 10,
        org: orgs[i % orgs.length],
        status: statuses[i % statuses.length],
      });
    }
  }

  /**
   * Frontend calls this when the viewport changes.
   * Differential Subscription: backend only tracks cells the client cares about.
   */
  public updateSubscriptions(
    toSubscribe: string[],
    toUnsubscribe: string[],
    resolution?: number,
  ) {
    toSubscribe.forEach((cell) => this.activeCells.add(cell));
    toUnsubscribe.forEach((cell) => this.activeCells.delete(cell));
    if (resolution !== undefined) {
      this.currentResolution = resolution;
    }
    console.log(
      `[BackendSimulator] Active H3 Subscriptions: ${this.activeCells.size} @ Res ${this.currentResolution}`,
    );
  }

  public setTps(tps: number) {
    this.tps = tps;
    console.log(`[BackendSimulator] TPS set to: ${tps}`);
  }

  public triggerBurst(count: number) {
    this.burstActive = true;
    this.burstCount = count;
    console.log(`[BackendSimulator] Triggering burst of ${count} vehicles`);
  }

  public setLatency(ms: number) {
    this.latencyMs = ms;
  }

  /**
   * Main simulation loop: only sends updates for vehicles in active cells.
   * This mimics the server-side spatial filtering.
   */
  public tick() {
    if (this.activeCells.size === 0) return;

    if (this.latencyMs > 0) {
      setTimeout(() => this.processTick(), this.latencyMs);
    } else {
      this.processTick();
    }
  }

  private processTick() {
    const visibleUpdates: SimulationVehicle[] = [];
    let sampleSize = this.tps;

    if (this.burstActive) {
      sampleSize += this.burstCount;
      this.burstActive = false; // Single tick burst
    }

    // Optimization: If TPS is massive, we don't random sample everything, we slice
    const startIdx = Math.floor(
      Math.random() * (this.vehicles.length - sampleSize),
    );

    for (let i = 0; i < sampleSize; i++) {
      const v = this.vehicles[startIdx + i];

      // Move vehicle subtly
      v.lat += (Math.random() - 0.5) * 0.0001;
      v.lng += (Math.random() - 0.5) * 0.0001;

      const cell = h3.latLngToCell(v.lat, v.lng, this.currentResolution);
      if (this.activeCells.has(cell)) {
        visibleUpdates.push({ ...v });
      }
    }

    if (visibleUpdates.length > 0) {
      this.onUpdate(visibleUpdates);
    }
  }
}
