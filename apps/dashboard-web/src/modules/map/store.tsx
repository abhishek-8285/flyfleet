import React, { createContext, useContext, useState, useCallback } from 'react';

// Keep the worker instance completely outside React's reactive scope
let workerInstance: Worker | null = null;

// Non-reactive GeoJSON reference
let currentGeoJSON: any = { type: "FeatureCollection", features: [] };

interface MapContextType {
  pulse: number;
  triggerPulse: () => void;
  initializeWorker: (orgId: string, token: string, wsUrl: string) => void;
  destroyWorker: () => void;
  // Expose non-reactive ref for direct access if needed
  getCurrentGeoJSON: () => any;
}

const MapContext = createContext<MapContextType | undefined>(undefined);

export const MapProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [pulse, setPulse] = useState(0);

  const triggerPulse = useCallback(() => {
    setPulse(p => p + 1);
  }, []);

  const initializeWorker = useCallback((orgId: string, token: string, wsUrl: string) => {
    if (workerInstance) {
      workerInstance.terminate();
    }

    workerInstance = new Worker(new URL("./telemetry.worker.ts", import.meta.url));

    workerInstance.onmessage = (event) => {
      const msg = event.data;
      if (msg.type === "GEOJSON_BATCH") {
        // Silently update the non-reactive reference
        currentGeoJSON = msg.payload;
        // Trigger React pulse to run rAF flush
        triggerPulse();
      }
    };

    workerInstance.postMessage({ type: "CONNECT", orgId, token, wsUrl });
  }, [triggerPulse]);

  const destroyWorker = useCallback(() => {
    if (workerInstance) {
      workerInstance.postMessage({ type: "DISCONNECT" });
      workerInstance.terminate();
      workerInstance = null;
    }
  }, []);

  const getCurrentGeoJSON = useCallback(() => currentGeoJSON, []);

  return (
    <MapContext.Provider value={{
      pulse,
      triggerPulse,
      initializeWorker,
      destroyWorker,
      getCurrentGeoJSON,
    }}>
      {children}
    </MapContext.Provider>
  );
};

export const useMapStore = () => {
  const context = useContext(MapContext);
  if (!context) {
    throw new Error("useMapStore must be used within a MapProvider");
  }
  return context;
};
