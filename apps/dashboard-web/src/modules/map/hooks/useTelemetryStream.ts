import React, { useEffect, useRef } from "react";
import { useDispatch } from "react-redux";
import {
  updateBreadcrumbs,
  updateTelemetryBatch,
} from "../../../store/telemetrySlice";
import { WorkerManager } from "../../../workers/utils/WorkerManager";

interface TelemetryMetrics {
  fps: number;
  dropCount: number;
  activeH3: number;
  telemetryLag: number;
  vramMb: number;
  activeEntities: number;
}

export const useTelemetryStream = () => {
  const dispatch = useDispatch();
  const telemetryWorkerRef = useRef<WorkerManager | null>(null);
  const ingestionWorkerRef = useRef<WorkerManager | null>(null);
  const spatialWorkerRef = useRef<WorkerManager | null>(null);
  const searchWorkerRef = useRef<WorkerManager | null>(null);

  const [metrics, setMetrics] = React.useState<TelemetryMetrics>({
    fps: 0,
    dropCount: 0,
    activeH3: 0,
    telemetryLag: 0,
    vramMb: 0,
    activeEntities: 0,
  });

  const SAB_CONSTRUCTOR = typeof SharedArrayBuffer !== "undefined" ? SharedArrayBuffer : ArrayBuffer;

  const [renderBufferSabs] = React.useState(() =>
    Array.from({ length: 10 }, () => new SAB_CONSTRUCTOR(100000 * 8 * 4)),
  ); 

  useEffect(() => {
    // 1. Core Production Constants
    const RECORD_SIZE = 128;
    const RING_SIZE = RECORD_SIZE * 10000;
    const telemetrySab = new SAB_CONSTRUCTOR(RING_SIZE + 8); // +8 for head/tail pointers
    const pressureSab = new SAB_CONSTRUCTOR(4); // 1 int32 slot for backpressure
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const pressureArray = new Int32Array(pressureSab);

    // 2. Initialize WorkerManagers
    const telemetryWorker = new WorkerManager({
      name: "telemetry",
      createWorker: () =>
        new Worker(
          new URL("../../../workers/telemetry.worker.ts", import.meta.url),
          { type: "module" },
        ),
      onMessage: (e) => handleTelemetryMessage(e),
    });

    const ingestionWorker = new WorkerManager({
      name: "ingestion",
      createWorker: () =>
        new Worker(
          new URL("../../../workers/ingestion.worker.ts", import.meta.url),
          { type: "module" },
        ),
      onMessage: (e) => handleIngestionMessage(e),
    });

    const spatialWorker = new WorkerManager({
      name: "spatial",
      createWorker: () =>
        new Worker(
          new URL("../../../workers/spatial.worker.ts", import.meta.url),
          { type: "module" },
        ),
      onMessage: (e) => handleSpatialMessage(e),
    });

    const operationalWorker = new WorkerManager({
      name: "operational",
      createWorker: () =>
        new Worker(
          new URL("../../../workers/operational.worker.ts", import.meta.url),
          { type: "module" },
        ),
      onMessage: (e) => handleOperationalMessage(e),
    });

    // Connect FleetService for enterprise registry support
    import("../../../modules/fleet/FleetService").then(({ fleetService }) => {
      fleetService.setWorker(operationalWorker);
    });

    telemetryWorkerRef.current = telemetryWorker;
    ingestionWorkerRef.current = ingestionWorker;
    spatialWorkerRef.current = spatialWorker;
    searchWorkerRef.current = operationalWorker;

    // 3. Setup Payloads & Initialize Grid
    const initPayload = {
      sharedBuffer: telemetrySab,
      offset: 0,
      size: RING_SIZE,
      pressureSab,
    };

    ingestionWorker.setInitialConfig({
      type: "INIT_SAB",
      payload: initPayload,
    });
    telemetryWorker.setInitialConfig({
      type: "INIT_SAB",
      payload: { ...initPayload, renderSabs: renderBufferSabs },
    });
    spatialWorker.setInitialConfig({ type: "INIT_SAB", payload: initPayload });
    operationalWorker.setInitialConfig({
      type: "INIT_SAB",
      payload: initPayload,
    });

    if (typeof window !== "undefined") {
      (window as any).__DEBUG_WORKER_MANAGERS__ = {
        ingestion: ingestionWorker,
        telemetry: telemetryWorker,
        spatial: spatialWorker,
        operational: operationalWorker,
      };
    }

    // 4. Health & FPS Monitoring
    const heartbeats = {
      telemetry: Date.now(),
      ingestion: Date.now(),
      spatial: Date.now(),
      search: Date.now(),
    };

    const hbInterval = setInterval(() => {
      const now = Date.now();
      const workers = [
        { ref: telemetryWorkerRef, name: "telemetry" as const },
        { ref: ingestionWorkerRef, name: "ingestion" as const },
        { ref: spatialWorkerRef, name: "spatial" as const },
        { ref: searchWorkerRef, name: "search" as const },
      ];

      workers.forEach(({ ref, name }) => {
        if (now - heartbeats[name] > 15000) {
          console.warn(`[WorkerSupervisor] ${name} worker hung.`);
        }
        ref.current?.postMessage({ type: "PING" });
      });
    }, 2000);

    let frameCount = 0;
    let lastFPSUpdate = performance.now();
    let fpsHandle: number;

    const animate = () => {
      frameCount++;
      const now = performance.now();
      if (now - lastFPSUpdate >= 1000) {
        setMetrics((m) => ({ ...m, fps: frameCount }));
        frameCount = 0;
        lastFPSUpdate = now;
      }
      fpsHandle = requestAnimationFrame(animate);
    };
    fpsHandle = requestAnimationFrame(animate);

    // 5. Global Message Handlers
    function handleTelemetryMessage(e: MessageEvent) {
      const { type, payload } = e.data;
      if (type === "PONG") {
        heartbeats.telemetry = Date.now();
        window.dispatchEvent(new CustomEvent("TELEMETRY_PONG"));
        return;
      }

      if (type === "TELEMETRY_BATCH") {
        // Only update Redux for small batches to avoid main-thread hangs
        // Massive batches (>1000) are intended for GPU/SAB path only
        if (payload.length < 1000) {
          dispatch(updateTelemetryBatch(payload));
          operationalWorker.postMessage({ type: "INDEX_VEHICLES", payload });
        }

        if (e.data.metrics) {
          setMetrics((m) => ({
            ...m,
            telemetryLag: e.data.metrics.avgLag,
            vramMb: e.data.metrics.vramMb,
            activeEntities: e.data.metrics.activeEntities,
          }));
        }
        if (e.data.dirtySlots) {
          window.dispatchEvent(
            new CustomEvent("TELEMETRY_DIRTY_SLOTS", {
              detail: e.data.dirtySlots,
            }),
          );
        }
      }

      if (type === "H3_DENSITY_UPDATE") {
        window.dispatchEvent(
          new CustomEvent("TELEMETRY_H3_DENSITY", { detail: payload }),
        );
      }

      if (type === "BREADCRUMBS_RESPONSE") {
        dispatch(updateBreadcrumbs(payload));
      }
    }

    function handleIngestionMessage(e: MessageEvent) {
      const { type, payload } = e.data;
      if (type === "PONG") {
        heartbeats.ingestion = Date.now();
        window.dispatchEvent(new CustomEvent("INGESTION_PONG"));
        return;
      }
      if (type === "INGESTION_METRICS") {
        setMetrics((m) => ({ ...m, activeEntities: payload.dropCount }));
      }
    }

    function handleSpatialMessage(e: MessageEvent) {
      const { type, payload } = e.data;
      if (type === "PONG") {
        heartbeats.spatial = Date.now();
        window.dispatchEvent(new CustomEvent("SPATIAL_PONG"));
        return;
      }
      if (type === "SYNC_H3_SUBSCRIPTIONS") {
        setMetrics((m) => ({ ...m, activeH3: payload.activeCount }));
        ingestionWorker.postMessage({
          type: "UPDATE_H3_SUBSCRIPTIONS",
          payload,
        });
        telemetryWorker.postMessage({ type: "SYNC_H3_SUBSCRIPTIONS", payload });
        window.dispatchEvent(
          new CustomEvent("SYNC_H3_SUBSCRIPTIONS", { detail: payload }),
        );
      }
    }

    function handleOperationalMessage(e: MessageEvent) {
      const { type, payload } = e.data;
      if (type === "PONG") {
        heartbeats.search = Date.now();
        window.dispatchEvent(new CustomEvent("OPERATIONAL_PONG"));
        return;
      }
      if (type === "SEARCH_RESULTS") {
        window.dispatchEvent(
          new CustomEvent("VEHICLE_SEARCH_RESULTS", { detail: payload }),
        );
      }
      if (type === "FILTER_METADATA") {
        window.dispatchEvent(
          new CustomEvent("VEHICLE_FILTER_METADATA", { detail: payload }),
        );
      }
      if (type === "REPLAY_DATA") {
        window.dispatchEvent(
          new CustomEvent("VEHICLE_REPLAY_DATA", { detail: payload }),
        );
      }
    }

    const handleForceCrash = (e: any) => {
      const type = e.detail;
      console.warn(`[useTelemetryStream] Received FORCE_CRASH for: ${type}`);
      if (type === "telemetry" && telemetryWorkerRef.current) {
        telemetryWorkerRef.current.forceCrash();
      } else if (type === "operational" && searchWorkerRef.current) {
        searchWorkerRef.current.forceCrash();
      }
    };
    window.addEventListener("FORCE_CRASH_WORKER", handleForceCrash);

    return () => {
      clearInterval(hbInterval);
      cancelAnimationFrame(fpsHandle);
      window.removeEventListener("FORCE_CRASH_WORKER", handleForceCrash);
      ingestionWorker.terminate();
      telemetryWorker.terminate();
      spatialWorker.terminate();
      operationalWorker.terminate();
      ingestionWorkerRef.current = null;
      telemetryWorkerRef.current = null;
      spatialWorkerRef.current = null;
      searchWorkerRef.current = null;
    };
  }, [dispatch, renderBufferSabs]);

  /**
   * Updates all workers with latest viewport bounds.
   */
  const updateViewport = React.useCallback((bounds: {
    minLng: number;
    maxLng: number;
    minLat: number;
    maxLat: number;
  }) => {
    // Spatial worker handles tile subscriptions
    spatialWorkerRef.current?.postMessage({
      type: "UPDATE_VIEWPORT",
      payload: bounds,
    });
    // Telemetry worker handles visual filtering (hiding from GPU)
    telemetryWorkerRef.current?.postMessage({
      type: "UPDATE_VIEWPORT",
      payload: bounds,
    });
  }, []);

  const sendToWorker = React.useCallback((type: string, payload: any) => {
    if (ingestionWorkerRef.current) {
      ingestionWorkerRef.current.postMessage({ type, payload });
    }
  }, []);

  const connectTelemetry = React.useCallback((url: string) => {
    if (ingestionWorkerRef.current) {
      ingestionWorkerRef.current.postMessage({
        type: "CONNECT_WEBSOCKET",
        payload: { url },
      });
    }
  }, []);

  const searchVehicle = React.useCallback((query: string, filters?: any) => {
    searchWorkerRef.current?.postMessage({
      type: "SEARCH",
      payload: { query, filters },
    });
  }, []);

  const getFilterMetadata = React.useCallback(() => {
    searchWorkerRef.current?.postMessage({ type: "GET_FILTER_METADATA" });
  }, []);

  const requestReplayData = React.useCallback((start: number, end: number) => {
    searchWorkerRef.current?.postMessage({
      type: "GET_REPLAY_DATA",
      payload: { start, end },
    });
  }, []);

  const setReplayState = React.useCallback((vehicles: any[]) => {
    telemetryWorkerRef.current?.postMessage({
      type: "SET_REPLAY_STATE",
      payload: vehicles,
    });
  }, []);

  const setWorkerMode = React.useCallback((mode: "LIVE" | "REPLAY") => {
    telemetryWorkerRef.current?.postMessage({
      type: "SET_WORKER_MODE",
      payload: { mode },
    });
  }, []);

  const requestBreadcrumbs = React.useCallback((vehicleId: string) => {
    telemetryWorkerRef.current?.postMessage({
      type: "GET_BREADCRUMBS",
      payload: { vehicleId },
    });
  }, []);

  return {
    sendToWorker,
    connectTelemetry,
    updateViewport,
    metrics,
    searchVehicle,
    getFilterMetadata,
    requestReplayData,
    setReplayState,
    setWorkerMode,
    requestBreadcrumbs,
    renderBufferSabs,
  };
};
