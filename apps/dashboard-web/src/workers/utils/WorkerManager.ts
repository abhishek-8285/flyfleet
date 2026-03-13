/**
 * WorkerManager: A robust wrapper for Web Workers with auto-restart and state re-hydration.
 * Ensures the system remains operational even if a worker crashes.
 */

export interface WorkerOptions {
  name: string;
  createWorker: () => Worker;
  onRestart?: () => void;
  onMessage?: (event: MessageEvent) => void;
}

export class WorkerManager {
  private worker: Worker | null = null;
  private messageQueue: any[] = [];
  private isInitializing: boolean = false;
  private config: any = null;

  constructor(private options: WorkerOptions) {
    this.spawn();
  }

  private spawn() {
    console.log(`[WorkerManager] Spawning worker: ${this.options.name}`);
    this.worker = this.options.createWorker();

    this.worker.onmessage = (e) => {
      if (this.options.onMessage) {
        this.options.onMessage(e);
      }
    };

    this.worker.onerror = (err) => {
      console.error(`[WorkerManager] Error in ${this.options.name}:`, err);
      this.handleCrash();
    };

    // If we have a stored config, re-apply it immediately (e.g. SABs)
    if (this.config) {
      this.rehydrate();
    }

    this.flushQueue();
  }

  private handleCrash() {
    console.warn(
      `[WorkerManager] Worker ${this.options.name} crashed. Restarting...`,
    );
    this.worker?.terminate();
    this.worker = null;

    // Exponential backoff or simple delay could be added here
    setTimeout(() => {
      this.spawn();
      if (this.options.onRestart) {
        this.options.onRestart();
      }
    }, 1000);
  }

  /**
   * Initial configuration that should survive restarts (e.g., SharedArrayBuffers).
   */
  public setInitialConfig(config: any) {
    this.config = config;
    this.postMessage(config);
  }

  private rehydrate() {
    if (this.config && this.worker) {
      console.log(
        `[WorkerManager] Re-hydrating ${this.options.name} with initial config.`,
      );
      this.worker.postMessage(this.config);
    }
  }

  public postMessage(message: any, transfer?: Transferable[]) {
    if (this.worker) {
      this.worker.postMessage(message, transfer as any);
    } else {
      console.log(
        `[WorkerManager] Worker ${this.options.name} is down. Queuing message.`,
      );
      this.messageQueue.push({ message, transfer });
    }
  }

  private flushQueue() {
    if (!this.worker) return;
    while (this.messageQueue.length > 0) {
      const { message, transfer } = this.messageQueue.shift();
      this.worker.postMessage(message, transfer);
    }
  }

  public terminate() {
    this.worker?.terminate();
    this.worker = null;
  }

  public forceCrash() {
    console.warn(`[WorkerManager] Force-crashing ${this.options.name}`);
    this.handleCrash();
  }
}
