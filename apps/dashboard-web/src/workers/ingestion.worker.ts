/* eslint-disable no-restricted-globals */
import { RingBuffer } from "../modules/telemetry/RingBuffer";

let ringBuffer: RingBuffer | null = null;
let socket: WebSocket | null = null;
let reconnectAttempts = 0;
let reconnectTimer: any = null;

// Fixed-size record layout (128 bytes)
const RECORD_SIZE = 128;
let dropCount = 0;

let activeH3Subscriptions = new Set<string>();

let pressureArray: Int32Array | null = null;
let isPressureShared = false;

self.onmessage = (event: MessageEvent) => {
  const { type, payload } = event.data;

  if (type === "INIT_SAB") {
    const { sharedBuffer, offset, size, pressureSab } = payload;
    ringBuffer = new RingBuffer(sharedBuffer, offset, size);
    if (pressureSab) {
      pressureArray = new Int32Array(pressureSab);
      isPressureShared = typeof SharedArrayBuffer !== "undefined" && pressureSab instanceof SharedArrayBuffer;
    }
    console.log(
      "[IngestionWorker] Initialized SharedArrayBuffer + Pressure Hub",
    );
  }

  if (type === "CONNECT_WEBSOCKET") {
    const { url } = payload;
    connect(url);
  }

  let traceCount = 0;

  if (type === "TELEMETRY_UPDATE") {
    if (!ringBuffer) return;
    const chunk = serializeTelemetry(payload);
    ringBuffer.push(chunk);
  }

  if (type === "TELEMETRY_BATCH") {
    if (!ringBuffer) return;
    const updates = payload as any[];
    updates.forEach((update) => {
      const chunk = serializeTelemetry(update);
      ringBuffer?.push(chunk);
    });
  }

  if (type === "PING") {
    self.postMessage({ type: "PONG" });
    return;
  }

  if (type === "UPDATE_H3_SUBSCRIPTIONS") {
    const { toSubscribe, toUnsubscribe } = payload;
    toSubscribe.forEach((h3Index: string) =>
      activeH3Subscriptions.add(h3Index),
    );
    toUnsubscribe.forEach((h3Index: string) =>
      activeH3Subscriptions.delete(h3Index),
    );

    // In a real production system, we'd send these to the backend via WebSocket
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: "H3_SUBSCRIBE", cells: toSubscribe }));
      socket.send(
        JSON.stringify({ type: "H3_UNSUBSCRIBE", cells: toUnsubscribe }),
      );
    }

    console.log(
      `[IngestionWorker] Active H3 Subscriptions: ${activeH3Subscriptions.size}`,
    );
  }
};

// Reuse buffers to avoid GC pressure
const serializationBuffer = new ArrayBuffer(RECORD_SIZE);
const serializationView = new DataView(serializationBuffer);
const serializationUint8 = new Uint8Array(serializationBuffer);
const encoder = new TextEncoder();

function connect(url: string) {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }

  if (socket) {
    console.log("[IngestionWorker] Closing existing socket before reconnecting...");
    socket.onopen = null;
    socket.onmessage = null;
    socket.onclose = null;
    socket.onerror = null;
    socket.close();
    socket = null;
  }

  console.log(`[IngestionWorker] Connecting to ${url} (Attempt ${reconnectAttempts + 1})`);
  socket = new WebSocket(url);
  socket.binaryType = "arraybuffer";

  socket.onopen = () => {
    console.log("[IngestionWorker] WebSocket Connected");
    reconnectAttempts = 0; // Reset attempts on success
  };

  socket.onmessage = (event: MessageEvent) => {
    if (!ringBuffer) return;

    // Check Pressure Signal from Telemetry Processor
    const pressure = pressureArray ? (isPressureShared ? Atomics.load(pressureArray, 0) : pressureArray[0]) : 0;
    if (pressure > 0) {
      // 50% probability of dropping if pressure is high
      if (Math.random() > 0.5) {
        dropCount++;
        performance.mark("FF_INGEST_DROP");
        return;
      }
    }

    let chunk: Uint8Array;

    if (event.data instanceof ArrayBuffer) {
      // PROMPT: Real hyperscale systems receive binary directly
      // If it's already a 128-byte chunk, copy directly
      if (event.data.byteLength === RECORD_SIZE) {
        chunk = new Uint8Array(event.data);
      } else {
        // Handle custom binary format decoding here if needed
        // For now, we assume if it's binary, it matches our internal format
        chunk = new Uint8Array(event.data).subarray(0, RECORD_SIZE);
      }
    } else {
      // Legacy JSON path
      let data =
        typeof event.data === "string" ? JSON.parse(event.data) : event.data;
      
      try {
        // Handle message types
        if (data && data.type === "connection_ack") {
          console.log("[IngestionWorker] Connection Ack received:", data.data);
          return;
        }
        if (data && data.type === "ping") {
          socket?.send(JSON.stringify({ type: "pong", data: { timestamp: new Date().toISOString() } }));
          return;
        }
        
        // Unwrap WSMessage envelope if present
        if (data && data.type === "position_update" && data.data) {
          data = data.data;
          const vehicleId = data.vehicle_id || data.id || "unknown";
          console.log(`[IngestionWorker] Position Update for ${vehicleId}:`, data);
        } else if (data && data.type) {
          // Ignore other message types for now
          return;
        }
        
        chunk = serializeTelemetry(data);
      } catch (err) {
        console.error("[IngestionWorker] Error processing message:", err);
        return;
      }
    }

    const pushed = ringBuffer.push(chunk);

    if (!pushed) {
      dropCount++;
      // Load Shedding: If buffer is full, we stop processing until cleared
      if (dropCount % 1000 === 0) {
        self.postMessage({
          type: "INGESTION_METRICS",
          payload: { dropCount, status: "CRITICAL_SHEDDING", utilization: 1.0 },
        });
      }
    } else if (dropCount > 0) {
      dropCount = 0;
      self.postMessage({
        type: "INGESTION_METRICS",
        payload: { dropCount: 0, status: "HEALTHY" },
      });
    }
  };

  socket.onclose = () => {
    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000) + Math.random() * 1000;
    console.log(`[IngestionWorker] WebSocket Disconnected. Reconnecting in ${Math.round(delay)}ms...`);
    
    reconnectAttempts++;
    reconnectTimer = setTimeout(() => connect(url), delay);
  };

  socket.onerror = (error) => {
    console.error("[IngestionWorker] WebSocket Error:", error);
    // onclose will handle the reconnection
  };
}

function serializeTelemetry(data: any): Uint8Array {
  // Direct-to-buffer serialization to avoid allocations
  serializationView.setFloat64(0, Date.now());
  serializationView.setFloat64(8, data.latitude || 0);
  serializationView.setFloat64(16, data.longitude || 0);
  serializationView.setFloat32(24, data.speed_kmh || data.speed || 0);
  serializationView.setFloat32(28, data.heading || 0);

  const idStr = data.vehicle_id || data.id || "";
  const idBytes = encoder.encode(idStr);
  const len = Math.min(idBytes.length, 95);

  serializationView.setUint8(32, len);
  serializationUint8.fill(0, 33, 128); // Zero out old ID data
  serializationUint8.set(idBytes.subarray(0, len), 33);

  return serializationUint8;
}

export { };

