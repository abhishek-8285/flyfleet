import { store } from "../store";
import { addTimelineEvent } from "../store/timelineSlice";

export interface TelemetryPayload {
  vehicleId: string;
  driverId: string;
  tripId: string;
  lat: number;
  lng: number;
  speed: number;
  heading: number;
  status: "moving" | "idling" | "alert" | "offline";
  timestamp: string;
}

class TelemetryService {
  private socket: WebSocket | null = null;
  private url: string = process.env.REACT_APP_WS_URL || "ws://localhost:3001";
  private reconnectAttempts: number = 0;
  private buffer: any[] = [];
  private batchInterval: NodeJS.Timeout | null = null;

  connect() {
    if (this.socket) return;

    this.socket = new WebSocket(this.url);

    this.socket.onopen = () => {
      console.log("Telemetry WebSocket connected");
      this.reconnectAttempts = 0; // Reset on success
      this.startBatchProcessing();
    };

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === "TELEMETRY_PUSH") {
          this.buffer.push(data.payload);
        } else if (data.type === "EVENT_PUSH") {
          store.dispatch(addTimelineEvent(data.payload));
          if (data.payload.type === "alert") {
            store.dispatch({
              type: "alerts/addAlertEvent",
              payload: data.payload,
            });
            store.dispatch({ type: "dashboard/incrementAlertCount" });
          } else if (data.payload.type === "trip_started") {
            store.dispatch({
              type: "dashboard/updateKpi",
              payload: {
                active_trips:
                  (store.getState() as any).dashboard.stats?.active_trips + 1,
              },
            });
          }
        }
      } catch (err) {
        console.error("Failed to parse WebSocket message", err);
      }
    };

    this.socket.onclose = () => {
      console.log("Telemetry WebSocket disconnected. Reconnecting...");
      this.socket = null;
      this.stopBatchProcessing();

      this.reconnectAttempts++;
      const backoff = Math.min(
        1000 * Math.pow(2, this.reconnectAttempts),
        30000,
      ); // Exponential backoff capped at 30s

      setTimeout(() => this.connect(), backoff);
    };

    this.socket.onerror = (error) => {
      console.error("WebSocket Error:", error);
    };
  }

  private startBatchProcessing() {
    if (this.batchInterval) return;
    this.batchInterval = setInterval(() => {
      if (this.buffer.length > 0) {
        const batch = [...this.buffer];
        this.buffer = [];
        store.dispatch({
          type: "telemetry/updateTelemetryBatch",
          payload: batch,
        });
      }
    }, 100); // 10Hz update rate for the UI state
  }

  private stopBatchProcessing() {
    if (this.batchInterval) {
      clearInterval(this.batchInterval);
      this.batchInterval = null;
    }
  }

  disconnect() {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.stopBatchProcessing();
  }

  send(data: any) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(data));
    }
  }
}

export const telemetryService = new TelemetryService();
export default telemetryService;
