/* eslint-disable no-restricted-globals */
import * as h3 from "h3-js";
import { RingBuffer } from "../modules/telemetry/RingBuffer";
import { SlotManager } from "../modules/telemetry/SlotManager";

// Global Error Handler for deeper diagnostics
self.onerror = (message, source, lineno, colno, error) => {
  self.postMessage({
    type: "WORKER_CRASH_DIAGNOSTIC",
    payload: {
      message: typeof message === 'string' ? message : (message as any).message,
      lineno,
      colno,
      stack: error?.stack,
    },
  });
};

self.onunhandledrejection = (e) => {
  self.postMessage({
    type: "WORKER_CRASH_DIAGNOSTIC",
    payload: {
      message: `Unhandled Rejection: ${e.reason}`,
      stack: e.reason?.stack,
    },
  });
};

const AGGREGATE_RESOLUTION = 4; // Low resolution for global highlights
const AGGREGATE_INTERVAL_MS = 2000; // Slow update for global heatmap
let lastAggregateTime = 0;
const h3DensityMap = new Map<string, number>();

interface TelemetryUpdate {
  id: string;
  latitude: number;
  longitude: number;
  speed: number;
  heading: number;
  timestamp: number;
  lastUpdate: string;
}

const RECORD_SIZE = 128;
const SHARD_SIZE = 100000;
const NUM_SHARDS = 10;
const MAX_VEHICLES = SHARD_SIZE * NUM_SHARDS;
const VBO_STRIDE = 8; // [prev_lng, prev_lat, prev_hdg, curr_lng, curr_lat, curr_hdg, start_time, picking_id]
const POLL_INTERVAL_MS = 100;
const BATCH_INTERVAL_MS = 600;

let isReplayMode = false;
let ringBuffer: RingBuffer | null = null;
let renderBuffers: Float32Array[] = [];
let slotManager: SlotManager | null = null;

// Mapping of vehicle UUID to index in the renderBuffer
const vehicleIndexMap = new Map<string, number>();
const dirtySlots = new Set<number>();
let pressureArray: Int32Array | null = null;
let isPressureShared = false;

let telemetryBuffer: TelemetryUpdate[] = [];

let H3_RESOLUTION = 7;
let activeH3Cells = new Set<string>();

// Breadcrumb tracking: vehicleId -> Circular buffer of [lng, lat]
const MAX_BREADCRUMBS = 50;
const vehicleBreadcrumbMap = new Map<string, number[][]>();

self.onmessage = (event: MessageEvent) => {
  const { type, payload } = event.data;

  if (type === "INIT_SAB") {
    const { sharedBuffer, offset, size, renderSabs, pressureSab } = payload;
    ringBuffer = new RingBuffer(sharedBuffer, offset, size);
    if (renderSabs && Array.isArray(renderSabs)) {
      renderBuffers = renderSabs.map((sab) => new Float32Array(sab));
      slotManager = new SlotManager(MAX_VEHICLES);
    }
    if (pressureSab) {
      pressureArray = new Int32Array(pressureSab);
      isPressureShared = typeof SharedArrayBuffer !== "undefined" && pressureSab instanceof SharedArrayBuffer;
    }
    console.log(
      `[TelemetryWorker] Initialized with ${NUM_SHARDS} shards (Total: ${MAX_VEHICLES} vehicles)`,
    );
  } else if (type === "SET_WORKER_MODE") {
    isReplayMode = payload.mode === "REPLAY";
    console.log(`[TelemetryWorker] Mode changed to: ${payload.mode}`);
  } else if (type === "SYNC_H3_SUBSCRIPTIONS") {
    const { toSubscribe, toUnsubscribe, resolution } = payload;
    if (resolution !== undefined) {
      H3_RESOLUTION = resolution;
    }
    toSubscribe.forEach((c: string) => activeH3Cells.add(c));
    toUnsubscribe.forEach((c: string) => activeH3Cells.delete(c));
  } else if (type === "GET_BREADCRUMBS") {
    const { vehicleId } = payload;
    const breadcrumbs = vehicleBreadcrumbMap.get(vehicleId) || [];
    self.postMessage({
      type: "BREADCRUMBS_RESPONSE",
      payload: { vehicleId, breadcrumbs },
    });
  } else if (type === "SET_REPLAY_STATE") {
    const vehicles = payload as TelemetryUpdate[];
    const sm = slotManager;
    if (!sm || renderBuffers.length === 0) return;

    vehicles.forEach((update) => {
      let index = vehicleIndexMap.get(update.id);
      if (index === undefined) {
        index = sm.allocate();
        if (index === -1) return;
        vehicleIndexMap.set(update.id, index);
        const { shard, local } = SlotManager.decompose(index, SHARD_SIZE);
        renderBuffers[shard][local * VBO_STRIDE + 7] = index;
      }

      const { shard, local } = SlotManager.decompose(index, SHARD_SIZE);
      const buffer = renderBuffers[shard];
      const offset = local * VBO_STRIDE;

      // In replay mode, we often want snap-to-position, so prev = curr
      buffer[offset] = update.longitude;
      buffer[offset + 1] = update.latitude;
      buffer[offset + 2] = update.heading;

      buffer[offset + 3] = update.longitude;
      buffer[offset + 4] = update.latitude;
      buffer[offset + 5] = update.heading;
      buffer[offset + 6] = performance.now(); // Reset interpolation start

      dirtySlots.add(index);
    });
  } else if (type === "PING") {
    self.postMessage({ type: "PONG" });
  }
};

// Polling & Processing Loop (Load Shedder)
setInterval(() => {
  try {
    if (!ringBuffer || renderBuffers.length === 0 || !slotManager || isReplayMode)
      return;

    const occupied = ringBuffer.getOccupied();
    const recordsToRead = Math.floor(occupied / RECORD_SIZE);

    // Load Shedder: Cap the number of updates per tick to maintain 60FPS render target
    const MAX_UPDATES_PER_TICK = 5000;
    const VISIBLE_ICON_CAP = 5000; // Hard cap on high-fidelity icons (Section 3)
    const toProcess = Math.min(recordsToRead, MAX_UPDATES_PER_TICK);

    let visibleCount = 0;

    for (let i = 0; i < toProcess; i++) {
      const raw = ringBuffer.pop(RECORD_SIZE);
      if (!raw) break;

      const view = new DataView(raw.buffer, raw.byteOffset, raw.byteLength);
      const timestamp = view.getFloat64(0);
      const latitude = view.getFloat64(8);
      const longitude = view.getFloat64(16);

      const idLen = view.getUint8(32);
      // SECURITY FIX: Validate idLength to prevent out-of-bounds read (CWE-125)
      const idStart = 33;
      const maxIDLength = raw.byteLength - idStart;
      // Enforce reasonable maximum (UUID is 36 chars max) to limit memory allocation
      const MAX_VEHICLE_ID_LENGTH = 256;
      if (idLen > maxIDLength || idLen > MAX_VEHICLE_ID_LENGTH) {
        console.error('[TelemetryWorker] Invalid vehicle ID length, skipping record:', idLen);
        continue; // Skip this malformed record
      }
      const id = new TextDecoder().decode(raw.subarray(idStart, idStart + idLen));

      const update: TelemetryUpdate = {
        id,
        latitude,
        longitude,
        speed: view.getFloat32(24),
        heading: view.getFloat32(28),
        timestamp,
        lastUpdate: new Date(timestamp).toISOString(),
      };

      telemetryBuffer.push(update);

      // Viewport Culling using H3 cells
      const cell = h3.latLngToCell(
        update.latitude,
        update.longitude,
        H3_RESOLUTION,
      );
      let isVisible = activeH3Cells.has(cell);

      // Hard Entity Cap: Only allow 50k visible icons per tick
      if (isVisible) {
        if (visibleCount < VISIBLE_ICON_CAP) {
          visibleCount++;
        } else {
          isVisible = false; // Load shed high-fidelity icons if cap is reached
        }
      }

      let index = vehicleIndexMap.get(update.id);
      if (index === undefined) {
        index = slotManager.allocate();
        if (index === -1) continue;
        vehicleIndexMap.set(update.id, index);

        const { shard, local } = SlotManager.decompose(index, SHARD_SIZE);
        const buffer = renderBuffers[shard];
        const offset = local * VBO_STRIDE;
        buffer[offset + 7] = index; // Picking ID
      }

      const { shard, local } = SlotManager.decompose(index, SHARD_SIZE);
      const buffer = renderBuffers[shard];
      const offset = local * VBO_STRIDE;

      // Shift Current -> Previous
      buffer[offset] = buffer[offset + 3];
      buffer[offset + 1] = buffer[offset + 4];
      buffer[offset + 2] = buffer[offset + 5];

      if (isVisible) {
        buffer[offset + 3] = update.longitude;
        buffer[offset + 4] = update.latitude;
        buffer[offset + 5] = update.heading;
        buffer[offset + 6] = performance.now();
      } else {
        buffer[offset + 6] = 0;
        buffer[offset + 3] = -999;
      }

      dirtySlots.add(index);

      // Update Breadcrumbs
      let breadcrumbs = vehicleBreadcrumbMap.get(update.id);
      if (!breadcrumbs) {
        breadcrumbs = [];
        vehicleBreadcrumbMap.set(update.id, breadcrumbs);
      }
      breadcrumbs.push([update.longitude, update.latitude]);
      if (breadcrumbs.length > MAX_BREADCRUMBS) {
        breadcrumbs.shift();
      }

      const h3Cell = h3.latLngToCell(
        update.latitude,
        update.longitude,
        AGGREGATE_RESOLUTION,
      );
      h3DensityMap.set(h3Cell, (h3DensityMap.get(h3Cell) || 0) + 1);
    }

    // Load Shedding Breadcrumbs
    if (recordsToRead > MAX_UPDATES_PER_TICK) {
      performance.mark("FF_LOAD_SHEDDING_START");
      performance.measure("FF_LOAD_SHEDDING_DURATION", "FF_LOAD_SHEDDING_START");
    }

    // Backpressure Signaling
    if (pressureArray) {
      const val = recordsToRead > MAX_UPDATES_PER_TICK * 2 ? 1 : 0;
      if (isPressureShared) {
        Atomics.store(pressureArray, 0, val);
      } else {
        pressureArray[0] = val;
      }
    }

    // Periodic Aggregate Sync
    const now = performance.now();
    if (now - lastAggregateTime > AGGREGATE_INTERVAL_MS) {
      self.postMessage({
        type: "H3_DENSITY_UPDATE",
        payload: Array.from(h3DensityMap.entries()),
      });
      h3DensityMap.clear();
      lastAggregateTime = now;
    }
  } catch (err: any) {
    self.onerror?.(err.message, "", 0, 0, err);
  }
}, POLL_INTERVAL_MS);

// 2. Mapping of vehicle UUID to index in the renderBuffer
const vehicleLastUpdateMap = new Map<string, number>();

const TTL_MS = 10000; // Remove vehicle after 10s of inactivity

// ... (existing code for telemetryBuffer) ...

// Periodically flush deduped updates to the main thread
let lastBatchTime = performance.now();
const MAX_BATCH_PROCESSING_MS = 16; // 1 frame budget

setInterval(() => {
  if (telemetryBuffer.length > 0) {
    const batchStart = performance.now();
    const currentEpoch = Date.now();
    const deduped: Record<string, TelemetryUpdate & { index?: number }> = {};

    let totalLag = 0;
    let count = 0;

    telemetryBuffer.forEach((update) => {
      deduped[update.id] = {
        ...update,
        index: vehicleIndexMap.get(update.id),
      };
      vehicleLastUpdateMap.set(update.id, performance.now());
      if (update.timestamp) {
        totalLag += currentEpoch - update.timestamp;
        count++;
      }
    });

    // VRAM Calculation: Float32Array length * 4 bytes + picking buffers
    const approximateVramBytes = renderBuffers.reduce(
      (acc, b) => acc + b.byteLength,
      0,
    );

    self.postMessage({
      type: "TELEMETRY_BATCH",
      payload: Object.values(deduped),
      dirtySlots: Array.from(dirtySlots),
      metrics: {
        avgLag: count > 0 ? totalLag / count : 0,
        vramMb: approximateVramBytes / (1024 * 1024),
        activeEntities: vehicleIndexMap.size,
      },
    });

    dirtySlots.clear();
    telemetryBuffer = [];

    const processingTime = performance.now() - batchStart;
    if (processingTime > MAX_BATCH_PROCESSING_MS) {
      console.warn(
        `[TelemetryWorker] Batch processing exceeded frame budget: ${processingTime.toFixed(2)}ms`,
      );
    }
  }

  // Slot Recycling (TTL)
  const now = performance.now();
  Array.from(vehicleLastUpdateMap.entries()).forEach(([id, lastUpdate]) => {
    if (now - lastUpdate > TTL_MS) {
      const index = vehicleIndexMap.get(id);
      if (index !== undefined && slotManager && renderBuffers.length > 0) {
        slotManager.release(index);
        vehicleIndexMap.delete(id);
        vehicleLastUpdateMap.delete(id);
        vehicleBreadcrumbMap.delete(id);
        dirtySlots.add(index);

        // Zero out the VBO slot to stop rendering
        const { shard, local } = SlotManager.decompose(index, SHARD_SIZE);
        const buffer = renderBuffers[shard];
        const offset = local * VBO_STRIDE;
        for (let i = 0; i < VBO_STRIDE; i++) buffer[offset + i] = 0;
      }
    }
  });
}, BATCH_INTERVAL_MS);

function deserializeTelemetry(chunk: Uint8Array): TelemetryUpdate {
  const view = new DataView(chunk.buffer, chunk.byteOffset, chunk.byteLength);

  // Layout matching IngestionWorker
  const timestamp = view.getFloat64(0);
  const latitude = view.getFloat64(8);
  const longitude = view.getFloat64(16);
  const speed = view.getFloat32(24);
  const heading = view.getFloat32(28);

  const idLen = view.getUint8(32);
  const idBytes = chunk.subarray(33, 33 + idLen);
  const decoder = new TextDecoder();
  const id = decoder.decode(idBytes);

  return {
    id,
    latitude,
    longitude,
    speed,
    heading,
    timestamp,
    lastUpdate: new Date(timestamp).toISOString(),
  };
}

export { };

