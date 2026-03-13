import React, { useState } from "react";
import { useSelector } from "react-redux";
import { LiveAlertsFeed } from "../../modules/dispatch/components/LiveAlertsFeed";
import { FleetList } from "../../modules/fleet/components/FleetList";
import MapEngine from "../../modules/map/MapEngine";
import { MapErrorBoundary } from "../../modules/map/components/MapErrorBoundary";
import { useTelemetryStream } from "../../modules/map/hooks/useTelemetryStream";
import {
    addVehicleLayers,
    updateVehicleSource,
} from "../../modules/map/layers/VehicleSymbolLayer";
import { RootState } from "../../store";

/**
 * DispatchCenter: The central operational hub of the FlyFleet platform.
 * Integrates Registry, Alerts, and Map into a high-density split-pane view.
 */
const DispatchCenter: React.FC = () => {
  const [sidebarWidth, setSidebarWidth] = useState(400);
  const telemetryVehicles = useSelector(
    (state: RootState) => state.telemetry.vehicles,
  );
  const { updateViewport } = useTelemetryStream();

  const handleMapLoad = (map: maplibregl.Map) => {
    addVehicleLayers(map);
    // Real-time update logic
    window.addEventListener("TELEMETRY_BATCH", () => {
      const vehicleList = Object.values(telemetryVehicles);
      updateVehicleSource(map, vehicleList);
    });
  };

  return (
    <div className="flex h-full w-full bg-background-dark overflow-hidden">
      {/* Left Sidebar: Registry & Alerts */}
      <div
        style={{ width: sidebarWidth }}
        className="flex flex-col border-r border-border-color bg-surface z-20 shadow-2xl transition-all"
      >
        <div className="flex-1 overflow-hidden border-b border-border-color">
          <FleetList />
        </div>
        <div className="h-1/3 min-h-[200px]">
          <LiveAlertsFeed />
        </div>
      </div>

      {/* Resize Handle */}
      <div
        className="w-1 cursor-col-resize hover:bg-primary transition-colors z-30"
        onMouseDown={(e) => {
          const startX = e.pageX;
          const startWidth = sidebarWidth;
          const onMouseMove = (moveEvent: MouseEvent) => {
            setSidebarWidth(
              Math.max(
                300,
                Math.min(800, startWidth + (moveEvent.pageX - startX)),
              ),
            );
          };
          const onMouseUp = () => {
            document.removeEventListener("mousemove", onMouseMove);
            document.removeEventListener("mouseup", onMouseUp);
          };
          document.addEventListener("mousemove", onMouseMove);
          document.addEventListener("mouseup", onMouseUp);
        }}
      />

      {/* Main Map Workspace */}
      <div className="flex-1 relative">
        <MapErrorBoundary>
          <MapEngine onMapLoad={handleMapLoad} />
        </MapErrorBoundary>

        {/* HUD Overlays */}
        <div className="absolute top-6 left-6 pointer-events-none">
          <div className="bg-surface-dark/80 backdrop-blur-md border border-border-color p-4 rounded-xl shadow-2xl pointer-events-auto">
            <h2 className="text-[10px] font-black uppercase tracking-widest text-primary mb-1">
              Fleet Operations
            </h2>
            <p className="text-xl font-bold text-white leading-tight">
              Dispatch Alpha
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DispatchCenter;
