import React, { useState } from 'react';
import { useSelector } from 'react-redux';
import { useAuth } from '../../context/AuthProvider';
import { VehicleDetailPanel } from '../../features/tracking/components/VehicleDetailPanel';
import LiveOperationsMap from '../../modules/dashboard/components/LiveOperationsMap';
import { FleetList } from '../../modules/fleet/components/FleetList';
import { RootState } from '../../store';

/**
 * DispatchCenter: The central operational hub of the FlyFleet platform.
 * Integrates Registry, Alerts, and Map into a high-density split-pane view.
 */
export const DispatchCenter: React.FC = () => {
  const { toggleSidebar } = useAuth();
  const [sidebarWidth, setSidebarWidth] = useState(400);
  const telemetryVehicles = useSelector((state: RootState) => state.telemetry.vehicles);
  const selectedVehicleId = useSelector((state: RootState) => state.telemetry.selectedVehicleId);
  const selectedId = selectedVehicleId; // Alias for LiveOperationsMap

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
            <div className="flex-1 overflow-auto">
              <FleetList variant="compact" gpsOnly={true} />
            </div>
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
                Math.max(300, Math.min(600, startWidth + (moveEvent.pageX - startX)))
              );
            };
            const onMouseUp = () => {
              document.removeEventListener('mousemove', onMouseMove);
              document.removeEventListener('mouseup', onMouseUp);
            };
            document.addEventListener('mousemove', onMouseMove);
            document.addEventListener('mouseup', onMouseUp);
          }}
        />

        {/* Main Map Workspace */}
        <div className="flex-1 relative bg-background-dark overflow-hidden">
          <LiveOperationsMap
            vehicles={Object.values(telemetryVehicles).map((v) => ({
              id: v.id,
              lat: v.latitude,
              lng: v.longitude,
              status: (v.speed ?? 0) > 0 ? 'active' : 'idle',
              label: v.registrationNumber || v.id.substring(0, 8),
            }))}
            height="100%"
            selectedId={selectedId}
            className="rounded-none"
            showLegend={false}
          />
          <div className="absolute top-6 left-6 pointer-events-none">
            <div className="bg-surface-dark/90 backdrop-blur-xl border border-border-color px-6 py-3 rounded-xl shadow-2xl pointer-events-auto flex items-center gap-8">
              <div className="flex flex-col border-r border-white/10 pr-8">
                <h2 className="text-[9px] font-black uppercase tracking-widest text-primary leading-none mb-1">
                  Ops Center
                </h2>
                <p className="text-sm font-bold text-white leading-none">
                  Dispatch Alpha
                </p>
              </div>

              <div className="flex gap-8">
                <div className="flex items-center gap-3">
                  <div className="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
                  <div>
                    <p className="text-[9px] text-text-muted uppercase font-bold tracking-widest leading-none mb-1">Active</p>
                    <p className="text-lg font-black text-white leading-none">
                      {Object.values(telemetryVehicles).filter((v) => (v.speed ?? 0) > 0).length}
                    </p>
                  </div>
                </div>
                
                <div className="flex items-center gap-3">
                  <div className="w-2 h-2 rounded-full bg-yellow-500 shadow-[0_0_8px_rgba(234,179,8,0.5)]" />
                  <div>
                    <p className="text-[9px] text-text-muted uppercase font-bold tracking-widest leading-none mb-1">Idle</p>
                    <p className="text-lg font-black text-white leading-none">
                      {Object.values(telemetryVehicles).filter((v) => (v.speed ?? 0) === 0).length}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <div className="w-2 h-2 rounded-full bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)]" />
                  <div>
                    <p className="text-[9px] text-text-muted uppercase font-bold tracking-widest leading-none mb-1">Alerts</p>
                    <p className="text-lg font-black text-white leading-none">
                      {Object.values(telemetryVehicles).filter((v) => v.status === 'alert').length}
                    </p>
                  </div>
                </div>
              </div>
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
