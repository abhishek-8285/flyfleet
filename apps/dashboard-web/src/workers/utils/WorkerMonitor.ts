/**
 * WorkerMonitor: Tracks the health of worker threads via heartbeats.
 * Automatically triggers recovery if a worker becomes non-responsive.
 */

export type WorkerType = "telemetry" | "operational" | "ingestion" | "spatial" | "collaboration";

interface WorkerHealth {
  lastHeartbeat: number;
  isHealthy: boolean;
  restartCount: number;
}

export class WorkerMonitor {
  private healthState: Map<WorkerType, WorkerHealth> = new Map();
  private maxLatency: number = 10000; // 10 seconds before marking unhealthy
  private checkInterval: number = 2000;
  private intervalId: any = null;
  private onFailure?: (type: WorkerType) => void;

  constructor(onFailure?: (type: WorkerType) => void) {
    this.onFailure = onFailure;
  }

  public register(type: WorkerType) {
    this.healthState.set(type, {
      lastHeartbeat: Date.now(),
      isHealthy: true,
      restartCount: 0,
    });
  }

  public heartbeat(type: WorkerType) {
    const state = this.healthState.get(type);
    if (state) {
      state.lastHeartbeat = Date.now();
      state.isHealthy = true;
    }
  }

  public start() {
    if (this.intervalId) return;

    this.intervalId = setInterval(() => {
      const now = Date.now();
      this.healthState.forEach((state, type) => {
        const latency = now - state.lastHeartbeat;
        if (latency > this.maxLatency && state.isHealthy) {
          console.warn(
            `[WorkerMonitor] Worker ${type} is non-responsive. Latency: ${latency}ms`,
          );
          state.isHealthy = false;
          if (this.onFailure) {
            this.onFailure(type);
          }
        }
      });
    }, this.checkInterval);
  }

  public stop() {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  public getStatus() {
    return Array.from(this.healthState.entries()).map(([type, state]) => ({
      type,
      ...state,
      latency: Date.now() - state.lastHeartbeat,
    }));
  }
}
