import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { useWebSocket } from "../../hooks/useWebSocket";
import { api } from "../../services/api";
import { getWsUrl } from "../../services/telemetryService";
import { Location, Trip, Vehicle } from "../../types";
import LiveMap from "../LiveMap";

interface LiveOperationsMapProps {
  trips: Trip[];
}

export const LiveOperationsMap: React.FC<LiveOperationsMapProps> = ({
  trips,
}) => {
  const { uiStrings } = useUIStrings("dashboard_telemetry");
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const { lastMessage } = useWebSocket(getWsUrl());

  useEffect(() => {
    const fetchVehicles = async () => {
      try {
        const response = await api.getVehicles();
        setVehicles(response.data || []);
      } catch (err) {
        console.error("Failed to fetch vehicles", err);
      }
    };
    fetchVehicles();
  }, []);

  useEffect(() => {
    if (lastMessage?.subject === "TELEMETRY_PUSH") {
      const telemetry = lastMessage.data;
      setVehicles((prev) =>
        prev.map((v) => {
          if (v.id === telemetry.vehicleId) {
            return {
              ...v,
              current_location: {
                latitude: telemetry.lat,
                longitude: telemetry.lng,
                speed: telemetry.speed,
                heading: telemetry.heading || 0,
                timestamp: telemetry.timestamp,
                ignition: telemetry.ignition,
                accuracy: 1,
              } as Location,
            };
          }
          return v;
        }),
      );
    }
  }, [lastMessage]);
  return (
    <div className="bg-surface-dark rounded-xl border border-border-color shadow-sm flex flex-col h-[500px] overflow-hidden relative group/map">
      <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker z-10 relative">
        <h3 className="text-white font-semibold flex items-center gap-2">
          <span className="material-symbols-outlined text-primary">public</span>
          {uiStrings?.title_map || "Live Fleet Map"}
        </h3>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 mr-2">
            <label className="text-xs text-text-muted font-medium cursor-pointer">
              Heatmap
            </label>
            <button
              role="switch"
              aria-checked="false"
              className="w-8 h-4 rounded-full bg-border-color relative transition-colors focus:outline-none hover:bg-primary/50"
            >
              <span className="absolute left-0.5 top-0.5 w-3 h-3 bg-white rounded-full transition-transform"></span>
            </button>
          </div>
          <div className="h-4 w-px bg-border-color"></div>
          <div className="flex gap-1 bg-border-color/50 p-1 rounded">
            <button className="px-3 py-1 text-xs font-medium text-white bg-surface-dark rounded shadow-sm">
              Map
            </button>
            <button className="px-3 py-1 text-xs font-medium text-text-muted hover:text-white transition-colors">
              Satellite
            </button>
          </div>
        </div>
      </div>

      <div className="relative flex-1 bg-[#0f151b] overflow-hidden">
        {/* Real Map Component - Now Primary */}
        <div className="absolute inset-0 z-0 opacity-100">
          <LiveMap trips={trips} vehicles={vehicles} height="100%" />
        </div>

        {/* Overlays */}
        <div className="absolute top-4 left-4 z-20 flex flex-col gap-3">
          <div className="relative shadow-lg">
            <input
              type="text"
              className="bg-surface-darker border border-border-color text-white text-xs rounded pl-8 pr-3 py-2 w-48 focus:ring-1 focus:ring-primary shadow-sm placeholder-text-muted/70"
              placeholder={
                uiStrings?.map_search_placeholder || "Search map area..."
              }
            />
            <span className="material-symbols-outlined absolute left-2 top-1.5 text-base text-text-muted">
              search
            </span>
          </div>
          <div className="bg-surface-darker/90 backdrop-blur-sm border border-border-color rounded p-2 flex flex-col gap-2 w-32 shadow-lg">
            <div className="flex items-center justify-between">
              <span className="text-[10px] text-text-muted font-medium">
                Geofences
              </span>
              <button className="w-6 h-3 rounded-full bg-primary relative transition-colors focus:outline-none">
                <span className="absolute right-0.5 top-0.5 w-2 h-2 bg-white rounded-full transition-transform"></span>
              </button>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[10px] text-text-muted font-medium">
                Traffic
              </span>
              <button className="w-6 h-3 rounded-full bg-border-color relative transition-colors focus:outline-none">
                <span className="absolute left-0.5 top-0.5 w-2 h-2 bg-white rounded-full transition-transform"></span>
              </button>
            </div>
          </div>
        </div>

        <div className="absolute top-4 right-4 bg-surface-darker/90 backdrop-blur-sm border border-border-color rounded p-3 z-20 shadow-lg">
          <div className="text-[10px] font-semibold text-text-muted uppercase mb-2">
            {uiStrings?.label_legend || "Legend"}
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
              <span className="text-xs text-white">
                {uiStrings?.status_active || "Active"}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-primary"></span>
              <span className="text-xs text-white">
                {uiStrings?.status_moving || "Moving"}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-yellow-500"></span>
              <span className="text-xs text-white">
                {uiStrings?.status_idling || "Idling"}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-red-500"></span>
              <span className="text-xs text-white">
                {uiStrings?.status_alert || "Alert"}
              </span>
            </div>
          </div>
        </div>

        <div className="absolute bottom-4 right-4 flex flex-col gap-2 z-20">
          {/* Custom Zoom Controls */}
          <button className="w-8 h-8 bg-surface-dark border border-border-color text-white rounded flex items-center justify-center hover:bg-border-color shadow-lg">
            <span className="material-symbols-outlined text-sm">add</span>
          </button>
          <button className="w-8 h-8 bg-surface-dark border border-border-color text-white rounded flex items-center justify-center hover:bg-border-color shadow-lg">
            <span className="material-symbols-outlined text-sm">remove</span>
          </button>
        </div>
        <div className="absolute bottom-1 left-1 px-2 py-1 bg-surface-darker/80 text-[10px] text-text-muted rounded backdrop-blur">
          © OpenMapTiles | Source: GIS Data
        </div>
      </div>
    </div>
  );
};
