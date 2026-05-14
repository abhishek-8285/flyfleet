import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useAuth } from '../../context/AuthProvider';
import { VehicleDetailPanel } from '../../features/tracking/components/VehicleDetailPanel';
import LiveOperationsMap from '../../modules/dashboard/components/LiveOperationsMap';
import { AssetDispatchList } from '../../modules/fleet/components/AssetDispatchList';
import { RootState } from '../../store';
import { useTelemetryStream } from '../../modules/map/hooks/useTelemetryStream';

/**
 * DispatchCenter: The central operational hub of the FlyFleet platform.
 * Integrates Registry, Alerts, and Map into a high-density split-pane view.
 */
export const DispatchCenter: React.FC = () => {
  const { toggleSidebar } = useAuth();
  const [sidebarWidth, setSidebarWidth] = useState(400);
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const [showLegend, setShowLegend] = useState(true);
  const [showLayerPicker, setShowLayerPicker] = useState(false);
  const [currentLayer, setCurrentLayer] = useState('roadmap');

  const mapLayers = {
    roadmap: "https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}",
    satellite: "https://mt1.google.com/vt/lyrs=s&x={x}&y={y}&z={z}",
    hybrid: "https://mt1.google.com/vt/lyrs=y&x={x}&y={y}&z={z}",
    terrain: "https://mt1.google.com/vt/lyrs=p&x={x}&y={y}&z={z}",
    dark: "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png",
  };

  const telemetryVehicles = useSelector((state: RootState) => state.telemetry.vehicles);
  const selectedVehicleId = useSelector((state: RootState) => state.telemetry.selectedVehicleId);
  const selectedId = selectedVehicleId; // Alias for LiveOperationsMap

  const { connectTelemetry } = useTelemetryStream();
  const user = useSelector((state: RootState) => state.auth.user);
  const theme = user?.theme || 'light';

  // Auto-switch to dark layer if theme is dark and layer is roadmap
  const effectiveLayer = (currentLayer === 'roadmap' && theme === 'dark') ? 'dark' : currentLayer;

  useEffect(() => {
    const token = localStorage.getItem('access_token') || '';
    const wsUrl = `ws://localhost:8084/ws/telemetry?token=${token}`;
    connectTelemetry(wsUrl);
  }, [connectTelemetry]);

  return (
    <div className="flex h-screen w-full bg-background-dark dark:bg-black font-sans overflow-hidden transition-colors duration-500">
      <div className="flex h-full w-full relative z-10">
        {/* Left Sidebar: Registry */}
        <div
          style={{ width: isSidebarOpen ? sidebarWidth : 0 }}
          className={`flex flex-col bg-white dark:bg-[#0b1219] shadow-2xl transition-all duration-300 ease-in-out relative ${!isSidebarOpen ? 'overflow-hidden' : 'border-r border-gray-200 dark:border-white/5'}`}
        >
          <div className="flex-1 overflow-hidden">
            <AssetDispatchList />
          </div>
        </div>

        {/* Resize Handle & Toggle Button Container */}
        <div className="relative flex items-center">
          {/* Resize Handle */}
          {isSidebarOpen && (
            <div
              className="w-1 h-full cursor-col-resize hover:bg-primary transition-colors z-30"
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
          )}

          {/* Sidebar Toggle Button (Circular Chevron) */}
          <button
            onClick={() => setIsSidebarOpen(!isSidebarOpen)}
            className={`absolute top-4 z-[1001] w-8 h-8 bg-white dark:bg-[#1a2b3c] border border-gray-200 dark:border-white/10 hover:bg-gray-50 dark:hover:bg-white/5 transition-all flex items-center justify-center rounded-full shadow-lg ${
              isSidebarOpen ? '-right-4' : 'left-4'
            }`}
            title={isSidebarOpen ? "Collapse Sidebar" : "Expand Sidebar"}
          >
            <span className={`material-symbols-outlined text-gray-600 dark:text-gray-300 text-[18px] transition-transform duration-300 ${isSidebarOpen ? '' : 'rotate-180'}`}>
              chevron_left
            </span>
          </button>
        </div>

        {/* Main Map Workspace */}
        <div className="flex-1 relative bg-background-dark dark:bg-black overflow-hidden">
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
            showLegend={showLegend}
            tileLayerUrl={mapLayers[effectiveLayer as keyof typeof mapLayers]}
          />

          {/* Map Controls: Legend & Layers */}
          <div className="absolute top-6 right-6 flex flex-col gap-3 z-[1001]">
            {/* Info Toggle */}
            <button
              onClick={() => setShowLegend(!showLegend)}
              className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all shadow-2xl border ${
                showLegend 
                  ? 'bg-primary text-white border-primary' 
                  : 'bg-white/90 dark:bg-[#1a2b3c]/90 backdrop-blur-md text-gray-700 dark:text-gray-200 border-gray-200 dark:border-white/10 hover:bg-white dark:hover:bg-white/5'
              }`}
              title="Map Legend"
            >
              <span className="material-symbols-outlined text-lg font-bold">info</span>
            </button>

            {/* Layer Picker Toggle */}
            <div className="relative">
              <button
                onClick={() => setShowLayerPicker(!showLayerPicker)}
                className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all shadow-2xl border ${
                  showLayerPicker 
                    ? 'bg-primary text-white border-primary' 
                    : 'bg-white/90 dark:bg-[#1a2b3c]/90 backdrop-blur-md text-gray-700 dark:text-gray-200 border-gray-200 dark:border-white/10 hover:bg-white dark:hover:bg-white/5'
                }`}
                title="Map Layers"
              >
                <span className="material-symbols-outlined text-lg font-bold">layers</span>
              </button>

              {/* Layer Options Menu */}
              {showLayerPicker && (
                <div className="absolute top-0 right-14 bg-white/95 dark:bg-[#1a2b3c]/95 backdrop-blur-xl border border-gray-200 dark:border-white/10 p-2 rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.2)] min-w-[140px] animate-in fade-in slide-in-from-right-4 duration-200">
                  <div className="flex flex-col gap-1">
                    {Object.keys(mapLayers).map((layer) => (
                      <button
                        key={layer}
                        onClick={() => {
                          setCurrentLayer(layer);
                          setShowLayerPicker(false);
                        }}
                        className={`px-4 py-2.5 rounded-xl text-[11px] font-black uppercase tracking-widest text-left transition-all ${
                          currentLayer === layer 
                            ? 'bg-primary text-white' 
                            : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-white/5 hover:text-black dark:hover:text-white'
                        }`}
                      >
                        {layer}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Removed Ops Center overlay per user request */}
        </div>

        {/* Right Sidebar: Active Asset Detail Drawer */}
        <VehicleDetailPanel />
      </div>
    </div>
  );
};

export default DispatchCenter;
