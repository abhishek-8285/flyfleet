/**
 * CollaborativeManager: Synchronizes multiple browser tabs/windows for multi-monitor setups.
 * Primarily uses BroadcastChannel for low-latency, same-origin communication.
 */

export interface SyncEvent {
  type: "VIEWPORT_SYNC" | "VEHICLE_FOCUS" | "LAYOUT_SYNC" | "SIM_CONTROL";
  payload: any;
  timestamp: number;
}

export class CollaborativeManager {
  private channel: BroadcastChannel;
  private listeners: Set<(event: SyncEvent) => void> = new Set();

  constructor(channelName: string = "flyfleet_control_tower") {
    this.channel = new BroadcastChannel(channelName);
    this.channel.onmessage = (event: MessageEvent<SyncEvent>) => {
      this.listeners.forEach((fn) => fn(event.data));
    };
  }

  /**
   * Broadcasts an event to all other windows.
   */
  broadcast(type: SyncEvent["type"], payload: any) {
    const event: SyncEvent = {
      type,
      payload,
      timestamp: Date.now(),
    };
    this.channel.postMessage(event);
  }

  /**
   * Adds a listener for incoming sync events.
   */
  subscribe(fn: (event: SyncEvent) => void) {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  close() {
    this.channel.close();
    this.listeners.clear();
  }
}

// Global singleton instance for easy access across hooks
export const collaborativeManager = new CollaborativeManager();
