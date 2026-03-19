import { create } from "zustand";

interface MapStore {
  // Pure, non-reactive GeoJSON payload ready strictly for Mapbox
  currentGeoJSON: any; 
  // Atomic React pulse signal to run the generic requestAnimationFrame bridge
  pulse: number;
  
  // Controls
  initializeWorker: (orgId: string, token: string, wsUrl: string) => void;
  destroyWorker: () => void;
}

// Keep the worker instance completely outside React/Zustand's reactive scope
let workerInstance: Worker | null = null;

export const useMapStore = create<MapStore>((set, get) => ({
  currentGeoJSON: { type: "FeatureCollection", features: [] },
  pulse: 0,

  initializeWorker: (orgId: string, token: string, wsUrl: string) => {
    if (workerInstance) {
      workerInstance.terminate();
    }
    
    workerInstance = new Worker(new URL("./telemetry.worker.ts", import.meta.url));

    workerInstance.onmessage = (event) => {
      const msg = event.data;
      if (msg.type === "GEOJSON_BATCH") {
         // Silently safely update the non-reactive reference
         get().currentGeoJSON = msg.payload;
         
         // Trigger single atomic React pulse to run rAF flush on the canvas
         set((state) => ({ pulse: state.pulse + 1 }));
      }
    };

    workerInstance.postMessage({ type: "CONNECT", orgId, token, wsUrl });
  },

  destroyWorker: () => {
    if (workerInstance) {
      workerInstance.postMessage({ type: "DISCONNECT" });
      workerInstance.terminate();
      workerInstance = null;
    }
  }
}));
