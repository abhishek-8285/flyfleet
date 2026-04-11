import { store } from "../store";
import { addTimelineEvent } from "../store/timelineSlice";
import { updateKpi, incrementAlertCount, fetchStatsSuccess } from "../store/dashboardSlice";
import type { RootState } from "../store";

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

export interface StatsUpdatePayload {
  stats: {
    total_vehicles: number;
    active_trips: number;
    active_drivers: number;
    alerts_count: number;
    fleet_health_score: number;
    vehicles_offline: number;
    sla_violations: number;
  };
}

export interface OverviewUpdatePayload {
  stats: {
    total_vehicles: number;
    active_trips: number;
    active_drivers: number;
    alerts_count: number;
    fleet_health_score: number;
    vehicles_offline: number;
    sla_violations: number;
  };
  timeline: any[];
  alerts: any[];
}

export const getWsUrl = (): string => {
  let wsUrl = process.env.REACT_APP_WS_URL || "";
  if (!wsUrl) {
    const apiUrl = import.meta.env.VITE_API_URL || "/api/v1";
    if (apiUrl.startsWith("http")) {
      wsUrl = apiUrl.replace(/^http/, "ws") + "/ws";
    } else {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      wsUrl = `${protocol}//${window.location.host}${apiUrl}/ws`;
    }
  }

  const state = store.getState() as RootState;
  const token = state.auth.token || "";
  let orgId = "";
  if (token) {
    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
          return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
      }).join(''));
      const payload = JSON.parse(jsonPayload);
      orgId = payload.org_id || "";
    } catch (e) {}
  }

  return `${wsUrl}?token=${token}&org_id=${orgId}`;
};

class TelemetryService {
  private socket: WebSocket | null = null;
  private reconnectAttempts: number = 0;
  private buffer: any[] = [];
  private batchInterval: NodeJS.Timeout | null = null;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private shouldReconnect: boolean = true;

  connect() {
    if (this.socket) return;

    const state = store.getState() as RootState;
    const token = state.auth.token || "";
    if (!token) {
      console.warn("TelemetryService: No token found, skipping connection");
      return;
    }

    this.shouldReconnect = true;
    const url = getWsUrl();
    this.socket = new WebSocket(url);

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
        } else if (data.type === "STATS_UPDATE") {
          // Update dashboard statistics in real-time
          store.dispatch(fetchStatsSuccess(data.payload.stats));
        } else if (data.type === "OVERVIEW_UPDATE") {
          // Update full dashboard overview
          store.dispatch(fetchStatsSuccess(data.payload.stats));
        }
      } catch (err) {
        console.error("Failed to parse WebSocket message", err);
      }
    };

    this.socket.onclose = () => {
      this.socket = null;
      this.stopBatchProcessing();

      if (!this.shouldReconnect) {
        console.log("Telemetry WebSocket closed intentionally.");
        return;
      }

      console.log("Telemetry WebSocket disconnected. Reconnecting...");
      this.reconnectAttempts++;
      const backoff = Math.min(
        1000 * Math.pow(2, this.reconnectAttempts),
        30000,
      ); // Exponential backoff capped at 30s

      if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = setTimeout(() => this.connect(), backoff);
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
    this.shouldReconnect = false;
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
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
