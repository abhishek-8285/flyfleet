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
  private activeCells: Set<string> = new Set();
  private currentResolution: number = 7;
  private tps: number = 10000; // Default 10k/sec
  private burstActive: boolean = false;
  private burstCount: number = 0;
  private latencyMs: number = 0;
  private onUpdate: (updates: SimulationVehicle[]) => void;
  private sequenceNum: number = 0;

  constructor(onUpdate: (updates: SimulationVehicle[]) => void) {
    this.onUpdate = onUpdate;
    console.log("[BackendSimulator] Hyperscale Dynamic Engine Initialized");
  }

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

  public async tick() {
    if (this.activeCells.size === 0) return;

    if (this.latencyMs > 0) {
      setTimeout(() => this.processTick(), this.latencyMs);
    } else {
      await this.processTick();
    }
  }

  private async processTick() {
    let sampleSize = this.tps;
    if (this.burstActive) {
      sampleSize += this.burstCount;
      this.burstActive = false; // Single tick burst
    }

    const orgs = ["North Fleet", "South Logistics", "West Express", "East Cargo"];
    const statuses = ["active", "idle", "delayed", "maintenance"];
    const activeCellsArray = Array.from(this.activeCells);
    
    // Determine target updates algebraically to bypass O(N) filtering
    const proportionVisible = Math.min(1.0, activeCellsArray.length / 5000);
    const targetUpdates = Math.max(10, Math.floor(sampleSize * proportionVisible));

    // Chunking to prevent main-thread allocation failure (V8 string limits) 
    const MAX_CHUNK_SIZE = 5000;
    
    for (let c = 0; c < targetUpdates; c += MAX_CHUNK_SIZE) {
      const chunkCount = Math.min(MAX_CHUNK_SIZE, targetUpdates - c);
      const visibleUpdates: SimulationVehicle[] = [];
      
      for (let i = 0; i < chunkCount; i++) {
        this.sequenceNum++;
        const targetCell = activeCellsArray[Math.floor(Math.random() * activeCellsArray.length)];
        
        try {
            const [lat, lng] = h3.cellToLatLng(targetCell);
            const varLat = lat + (Math.random() - 0.5) * 0.05;
            const varLng = lng + (Math.random() - 0.5) * 0.05;
            
            visibleUpdates.push({
              id: `HS-${this.sequenceNum % 10000000}`,
              lat: varLat,
              lng: varLng,
              org: orgs[(this.sequenceNum) % orgs.length],
              status: statuses[(this.sequenceNum) % statuses.length],
            });
        } catch (e) {
            // Ignore mapping exceptions safely
        }
      }
      
      if (visibleUpdates.length > 0) {
        this.onUpdate(visibleUpdates);
        // CRUCIAL: Release main thread so WebGL and Playwright UI don't freeze
        await new Promise((resolve) => setTimeout(resolve, 0));
      }
    }
  }
}
