import React, { useState, useRef, useEffect } from "react";
import { useSelector } from "react-redux";
import maplibregl from "maplibre-gl";
import { LiveAlertsFeed } from "../../modules/dispatch/components/LiveAlertsFeed";
import { FleetList } from "../../modules/fleet/components/FleetList";
import MapEngine from "../../modules/map/MapEngine";
import { MapErrorBoundary } from "../../modules/map/components/MapErrorBoundary";
import { useTelemetryStream } from "../../modules/map/hooks/useTelemetryStream";
import {
    addVehicleLayers,
    updateVehicleSource,
} from "../../modules/map/layers/VehicleSymbolLayer";
import { VehicleDetailPanel } from "../../features/tracking/components/VehicleDetailPanel";
import { RootState } from "../../store";
import { useAuth } from "../../context/AuthProvider";

/**
 * DispatchCenter: The central operational hub of the FlyFleet platform.
 * Integrates Registry, Alerts, and Map into a high-density split-pane view.
 */
export const DispatchCenter: React.FC = () => {
  const { toggleSidebar } = useAuth();
  const sidebarCollapsed = useSelector((state: RootState) => state.auth.user?.sidebar_collapsed);
  const mapRef = useRef<maplibregl.Map | null>(null);
  
  // Immersive 'Focus Mode': Auto-collapse global navigation on mount
  useEffect(() => {
    if (!sidebarCollapsed) {
      toggleSidebar();
    }
  }, []);

  const [sidebarWidth, setSidebarWidth] = useState(400);
  const telemetryVehicles = useSelector(
    (state: RootState) => state.telemetry.vehicles,
  );
  const selectedVehicleId = useSelector((state: RootState) => state.telemetry.selectedVehicleId);
  const { updateViewport } = useTelemetryStream();

  // Selection Focus: Smoothly fly to selected asset
  useEffect(() => {
    if (mapRef.current && selectedVehicleId && telemetryVehicles[selectedVehicleId]) {
        const v = telemetryVehicles[selectedVehicleId];
        if (v.latitude && v.longitude) {
            mapRef.current.flyTo({
                center: [v.longitude, v.latitude],
                zoom: 15,
                essential: true,
                padding: { right: 450, left: 0, top: 0, bottom: 0 } // Offset for the right drawer
            });
        }
    }
  }, [selectedVehicleId]);

  // Reactive Map Sync: Push data to map source without React re-render loops
  useEffect(() => {
    if (mapRef.current) {
        const vehicleList = Object.values(telemetryVehicles);
        updateVehicleSource(mapRef.current, vehicleList);
    }
  }, [telemetryVehicles]);

  const handleMapLoad = (map: maplibregl.Map) => {
    mapRef.current = map;
    addVehicleLayers(map);

    // Initial Hydration: Force data onto map in case it arrived before map was ready
    const initialVehicles = Object.values(telemetryVehicles);
    if (initialVehicles.length > 0) {
      updateVehicleSource(map, initialVehicles);
    }
    
    // Interactive Map Popups for live telemetry
    map.on("click", "unclustered-point", (e) => {
        const features = map.queryRenderedFeatures(e.point, { layers: ["unclustered-point"] });
        if (!features.length) return;
        const f = features[0];
        const props = f.properties;
        
        new maplibregl.Popup({ className: 'telemetry-popup' })
          .setLngLat((f.geometry as any).coordinates)
          .setHTML(`
            <div class="p-1">
              <h3 class="font-bold text-[11px] mb-1 text-primary-dark">${props.registration}</h3>
              <div class="text-[9px] space-y-1 text-slate-700">
                <p>Velocity: <span class="font-bold">${props.speed} km/h</span></p>
                <p>Status: <span class="uppercase font-bold">${props.status}</span></p>
              </div>
            </div>
          `)
          .addTo(map);
    });

    // Change cursor on hover
    map.on("mouseenter", "unclustered-point", () => map.getCanvas().style.cursor = "pointer");
    map.on("mouseleave", "unclustered-point", () => map.getCanvas().style.cursor = "");
  };

  return (
    <div className="flex h-full w-full bg-background-dark overflow-hidden relative">
      <div className="flex h-full w-full relative z-10">
        {/* Left Sidebar: Registry */}
        <div
          style={{ width: sidebarWidth }}
          className="flex flex-col border-r border-border-color bg-surface shadow-2xl transition-all"
        >
          <div className="p-4 border-b border-border-color bg-surface-dark/50 flex justify-between items-center">
            <h2 className="text-sm font-bold text-white flex items-center gap-2">
              <span className="material-symbols-outlined text-primary text-lg">inventory_2</span>
              Fleet Registry
            </h2>
          </div>
          <div className="flex-1 overflow-hidden">
            <FleetList variant="compact" gpsOnly={true} />
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
                  Math.min(600, startWidth + (moveEvent.pageX - startX)),
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
        <div className="flex-1 relative bg-background-dark">
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

        {/* Right Sidebar: Active Asset Detail Drawer */}
        <VehicleDetailPanel />
      </div>
    </div>
  );
};

export default DispatchCenter;
