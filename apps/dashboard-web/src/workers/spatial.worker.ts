/* eslint-disable no-restricted-globals */
import * as h3 from "h3-js";

/**
 * SpatialWorker: Orchestrates H3 tile subscriptions based on viewport.
 * This worker ensures we ONLY ingest telemetry for vehicles in or near the viewport.
 */

let activeH3Cells = new Map<string, number>(); // cell -> lastSeenTimestamp
const H3_RESOLUTION = 7;
const EVICTION_TIMEOUT_MS = 60000; // 60s

// Global Error Handler for diagnostics
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

// GC Loop
setInterval(() => {
  const now = Date.now();
  const toUnsubscribe: string[] = [];
  activeH3Cells.forEach((lastSeen, cell) => {
    if (now - lastSeen > EVICTION_TIMEOUT_MS) {
      toUnsubscribe.push(cell);
    }
  });

  if (toUnsubscribe.length > 0) {
    toUnsubscribe.forEach((cell) => activeH3Cells.delete(cell));
    self.postMessage({
      type: "SYNC_H3_SUBSCRIPTIONS",
      payload: {
        toSubscribe: [],
        toUnsubscribe,
        activeCount: activeH3Cells.size,
      },
    });
  }
}, 10000);

self.onmessage = (event: MessageEvent) => {
  const { type, payload } = event.data;

  if (type === "UPDATE_VIEWPORT") {
    const { minLng, maxLng, minLat, maxLat } = payload;
    updateSubscriptions(minLng, maxLng, minLat, maxLat);
  } else if (type === "PING") {
    // console.log("[SpatialWorker] Received PING, responding with PONG");
    self.postMessage({ type: "PONG" });
  }
};

function updateSubscriptions(
  minLng: number,
  maxLng: number,
  minLat: number,
  maxLat: number,
) {
  const polygon: [number, number][] = [
    [minLat, minLng],
    [maxLat, minLng],
    [maxLat, maxLng],
    [minLat, maxLng],
    [minLat, minLng],
  ];

  const lngDiff = Math.abs(maxLng - minLng);
  const latDiff = Math.abs(maxLat - minLat);
  const maxDiff = Math.max(lngDiff, latDiff);

  let resolution = 7;
  if (maxDiff > 60) resolution = 3;
  else if (maxDiff > 20) resolution = 4;
  else if (maxDiff > 5) resolution = 5;
  else if (maxDiff > 1) resolution = 6;

  const newCellsList = h3.polygonToCells(polygon, resolution);
  const now = Date.now();
  const toSubscribe: string[] = [];

  newCellsList.forEach((cell) => {
    if (!activeH3Cells.has(cell)) {
      toSubscribe.push(cell);
    }
    activeH3Cells.set(cell, now); // Mark as recently seen
  });

  const activeCount = activeH3Cells.size;

  if (toSubscribe.length > 0) {
    self.postMessage({
      type: "SYNC_H3_SUBSCRIPTIONS",
      payload: { toSubscribe, toUnsubscribe: [], activeCount, resolution },
    });
  }
}
