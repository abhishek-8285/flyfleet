import { useEffect } from "react";
import { collaborativeManager, SyncEvent } from "./CollaborativeManager";

export interface MultiMonitorOptions {
  onViewportSync?: (bounds: any) => void;
  onSimControl?: (isActive: boolean) => void;
  onVehicleFocus?: (vehicleId: string) => void;
}

export const useMultiMonitor = (options: MultiMonitorOptions) => {
  useEffect(() => {
    const unsub = collaborativeManager.subscribe((event: SyncEvent) => {
      switch (event.type) {
        case "VIEWPORT_SYNC":
          options.onViewportSync?.(event.payload);
          break;
        case "SIM_CONTROL":
          options.onSimControl?.(event.payload);
          break;
        case "VEHICLE_FOCUS":
          options.onVehicleFocus?.(event.payload);
          break;
      }
    });

    return () => {
      unsub();
    };
  }, [options]);

  const broadcastViewport = (bounds: any) => {
    collaborativeManager.broadcast("VIEWPORT_SYNC", bounds);
  };

  const broadcastSimControl = (isActive: boolean) => {
    collaborativeManager.broadcast("SIM_CONTROL", isActive);
  };

  const broadcastVehicleFocus = (vehicleId: string) => {
    collaborativeManager.broadcast("VEHICLE_FOCUS", vehicleId);
  };

  return {
    broadcastViewport,
    broadcastSimControl,
    broadcastVehicleFocus,
  };
};
