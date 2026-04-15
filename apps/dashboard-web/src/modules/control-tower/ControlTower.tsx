import {
    Activity,
    Filter,
    Navigation,
    Play,
    ShieldAlert,
    Square,
    Truck,
} from "lucide-react";
import React, { useCallback, useEffect, useRef } from "react";
import { useSelector } from "react-redux";
import { useMultiMonitor } from "../../modules/collaboration/useMultiMonitor";
import MapEngine from "../../modules/map/MapEngine";
import { MapErrorBoundary } from "../../modules/map/components/MapErrorBoundary";
import { useTelemetryStream } from "../../modules/map/hooks/useTelemetryStream";
import { PointCloudLayer } from "../../modules/map/layers/PointCloudLayer";
import { InstancedVehicleLayer } from "../../modules/map/layers/InstancedVehicleLayer";
import { ClusterLayer } from "../../modules/map/layers/ClusterLayer";
import {
    addRouteLayers,
    clearRoutes,
    updateBreadcrumbs,
    updatePlannedRoute,
} from "../../modules/map/layers/RouteLayer";
import {
    addVehicleLayers,
    updateVehicleSource,
} from "../../modules/map/layers/VehicleSymbolLayer";
import { RootState } from "../../store";
import { WorkerMonitor } from "../../workers/utils/WorkerMonitor";
import { H3StreamingSimulator } from "../telemetry/H3StreamingSimulator";
import TimeScrubber from "./components/TimeScrubber";

const debounce = (fn: Function, ms: number) => {
  let timeoutId: ReturnType<typeof setTimeout>;
  return function (this: any, ...args: any[]) {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn.apply(this, args), ms);
  };
};

const ControlTower: React.FC = () => {
  console.log('[ControlTower] Component rendering...');
  
  const telemetryVehicles = useSelector(
    (state: RootState) => state.telemetry.vehicles,
  );
  const breadcrumbs = useSelector(
    (state: RootState) => state.telemetry.breadcrumbs,
  );
  
  console.log('[ControlTower] State loaded:', {
    vehicleCount: telemetryVehicles?.length || 0,
    breadcrumbs: Object.keys(breadcrumbs || {}).length
  });
  
  const [filterMetadata, setFilterMetadata] = React.useState<{
    orgs: string[];
    statuses: string[];
  }>({ orgs: [], statuses: [] });
  const [activeFilters, setActiveFilters] = React.useState<{
    org?: string;
    status?: string;
  }>({});
  const mapRef = useRef<maplibregl.Map | null>(null);
  const {
    sendToWorker,
    updateViewport,
    metrics,
    searchVehicle,
    getFilterMetadata,
    requestReplayData,
    setReplayState,
    setWorkerMode,
    requestBreadcrumbs,
    renderBufferSabs,
  } = useTelemetryStream();

  // Multi-Monitor Synchronization
  const [isSyncEnabled] = React.useState(false);
  const { broadcastViewport, broadcastSimControl } = useMultiMonitor({
    onViewportSync: (bounds: {
      minLng: number;
      maxLng: number;
      minLat: number;
      maxLat: number;
    }) => {
      if (isSyncEnabled && mapRef.current) {
        mapRef.current.fitBounds(
          [
            [bounds.minLng, bounds.minLat],
            [bounds.maxLng, bounds.maxLat],
          ],
          { animate: false },
        );
      }
    },
    onSimControl: (isActive: boolean) => {
      if (isSyncEnabled) {
        if (isActive) startSimulation();
        else stopSimulation();
      }
    },
  });

  const [showBreadcrumbs, setShowBreadcrumbs] = React.useState(true);
  const [isLivePlayback, setIsLivePlayback] = React.useState(true);
  const [replayRange] = React.useState(() => {
    const end = Date.now();
    const start = end - 30 * 60 * 1000; // 30 mins back
    return { start, end };
  });
  const instancedLayerRef = useRef<any>(null);
  const clusterLayerRef = useRef<any>(null);
  const pointCloudLayerRef = useRef<PointCloudLayer | null>(null);

  // Metrics Ref for external access (Certification)
  const metricsRef = useRef(metrics);
  metricsRef.current = metrics;

  // Worker Health State
  const [workerHealth, setWorkerHealth] = React.useState<any[]>([]);
  const monitorRef = useRef<WorkerMonitor | null>(null);

  // Initialize the layers dynamically when buffers are ready and map is loaded
  useEffect(() => {
    if (renderBufferSabs.length > 0 && mapRef.current && !(mapRef.current as any)._destroyed) {
      if (!instancedLayerRef.current) {
        const shards = renderBufferSabs.map((sab) => new Float32Array(sab));
        instancedLayerRef.current = new InstancedVehicleLayer(shards);
        
        // Check if map already has the layer to prevent double-adding
        if (!mapRef.current.getLayer(instancedLayerRef.current.id)) {
           mapRef.current.addLayer(instancedLayerRef.current);
        }
      }
      if (!clusterLayerRef.current) {
        const clusterLayer = new ClusterLayer();
        clusterLayerRef.current = clusterLayer;
        clusterLayerRef.current.onAdd(mapRef.current);
      }
    }
  }, [renderBufferSabs]);

  // Simulation state
  const [isSimulating, setIsSimulating] = React.useState(false);
  const simInterval = useRef<NodeJS.Timeout | null>(null);
  const simulatorRef = useRef<H3StreamingSimulator | null>(null);
  const [selectedVehicle, setSelectedVehicle] = React.useState<any>(null);

  const startSimulation = useCallback(() => {
    // Check if simulating via simInterval ref instead of state to keep callback stable
    if (simInterval.current) return;
    setIsSimulating(true);
    broadcastSimControl(true);

    // FAILSAFE: Force spatial worker to subscribe to the entire globe grid
    updateViewport({
      minLng: 68,
      maxLng: 98,
      minLat: 8,
      maxLat: 38,
    });

    if (!simulatorRef.current) {
      simulatorRef.current = new H3StreamingSimulator((updates: any[]) => {
        // Batch updates to avoid blocking the main thread
        const batch = updates.map((v: any) => ({
          id: v.id,
          latitude: v.lat,
          longitude: v.lng,
          org: v.org,
          status: v.status,
          speed: 50,
          heading: Math.random() * 360,
          lastUpdate: new Date().toISOString(),
        }));

        // Send in smaller chunks if the batch is too large
        const CHUNK_SIZE = 5000;
        for (let i = 0; i < batch.length; i += CHUNK_SIZE) {
          sendToWorker("TELEMETRY_BATCH", batch.slice(i, i + CHUNK_SIZE));
        }
      });
    }

    simInterval.current = setInterval(() => {
      simulatorRef.current?.tick();
    }, 1000);
  }, [updateViewport, sendToWorker, broadcastSimControl]);

  const stopSimulation = useCallback(() => {
    if (simInterval.current) {
      clearInterval(simInterval.current);
      simInterval.current = null;
    }
    setIsSimulating(false);
    broadcastSimControl(false);
  }, [broadcastSimControl]);

  useEffect(() => {
    const monitor = new WorkerMonitor((type) => {
      console.error(
        `[ControlTower] Critical failure detected in ${type} worker.`,
      );
      // Recovery logic could go here (e.g., restartSimulation or force re-sync)
    });

    monitor.register("telemetry");
    monitor.register("operational");
    monitor.start();
    monitorRef.current = monitor;

    const healthInterval = setInterval(() => {
      setWorkerHealth(monitor.getStatus());
      // Send heartbeats to workers (they need to respond)
      sendToWorker("PING", {});
    }, 2000);

    return () => {
      monitor.stop();
      clearInterval(healthInterval);
      stopSimulation();
    };
  }, []);

  // MapHeartbeat handler
  useEffect(() => {
    const handlePong = (e: any) => {
      // In a real system, workers would send specific PONG types
      monitorRef.current?.heartbeat("telemetry");
      monitorRef.current?.heartbeat("operational");
    };
    window.addEventListener("TELEMETRY_PONG", handlePong);
    window.addEventListener("OPERATIONAL_PONG", handlePong);
    return () => {
      window.removeEventListener("TELEMETRY_PONG", handlePong);
      window.removeEventListener("OPERATIONAL_PONG", handlePong);
    };
  }, []);

  useEffect(() => {
    const handleFilterMetadata = (e: any) => {
      setFilterMetadata(e.detail);
    };
    window.addEventListener("VEHICLE_FILTER_METADATA", handleFilterMetadata);
    getFilterMetadata(); // Initial fetch

    return () => {
      window.removeEventListener(
        "VEHICLE_FILTER_METADATA",
        handleFilterMetadata,
      );
    };
  }, [getFilterMetadata]);

  useEffect(() => {
    const handleReplayData = (e: any) => {
      setReplayState(e.detail);
    };
    window.addEventListener("VEHICLE_REPLAY_DATA", handleReplayData);
    return () =>
      window.removeEventListener("VEHICLE_REPLAY_DATA", handleReplayData);
  }, [setReplayState]);

  const handleLiveToggle = (live: boolean) => {
    setIsLivePlayback(live);
    setWorkerMode(live ? "LIVE" : "REPLAY");
    if (live) {
      if (mapRef.current) clearRoutes(mapRef.current);
    }
  };

  const handleTimeChange = (ts: number) => {
    // Request snapshots for a small window around the scrubbed time
    requestReplayData(ts - 2000, ts + 2000);
  };

  const handleMapLoad = useCallback((map: maplibregl.Map) => {
    mapRef.current = map;
    addVehicleLayers(map);
    addRouteLayers(map);

    // Always create fresh layer instances for a new map to prevent WebGL context sharing errors
    if (renderBufferSabs.length > 0) {
      const shards = renderBufferSabs.map((sab) => new Float32Array(sab));
      instancedLayerRef.current = new InstancedVehicleLayer(shards);
      map.addLayer(instancedLayerRef.current);
    }
    
    clusterLayerRef.current = new ClusterLayer();
    clusterLayerRef.current.onAdd(map);

    map.on('remove', () => {
       // Clean up refs when the map is destroyed so they don't hold stale WebGL resources
       mapRef.current = null;
       instancedLayerRef.current = null;
       clusterLayerRef.current = null;
    });
  }, [renderBufferSabs]);

  // Update map markers when telemetry state changes (Markers only)
  useEffect(() => {
    if (mapRef.current) {
      const vehicleList = Object.values(telemetryVehicles);
      updateVehicleSource(mapRef.current, vehicleList);

      // Periodically request breadcrumbs for selected vehicle
      if (selectedVehicle && showBreadcrumbs) {
        requestBreadcrumbs(selectedVehicle.id);
      }
    }
  }, [telemetryVehicles, selectedVehicle, showBreadcrumbs]);

  // Update Breadcrumbs on map when Redux state changes
  useEffect(() => {
    if (mapRef.current && selectedVehicle && showBreadcrumbs) {
      const trail = breadcrumbs[selectedVehicle.id];
      if (trail) {
        updateBreadcrumbs(mapRef.current, trail as [number, number][]);
      }
    } else if (mapRef.current && (!selectedVehicle || !showBreadcrumbs)) {
      clearRoutes(mapRef.current);
    }
  }, [breadcrumbs, selectedVehicle, showBreadcrumbs]);

  // Event-driven Instanced Layer Updates (Zero-Churn)
  useEffect(() => {
    const handleDirtySlots = (e: any) => {
      if (instancedLayerRef.current && mapRef.current && !(mapRef.current as any)._destroyed) {
        const shards = renderBufferSabs.map((sab) => new Float32Array(sab));
        instancedLayerRef.current.updateData(shards, e.detail);
        mapRef.current.triggerRepaint();
      }
    };

    window.addEventListener("TELEMETRY_DIRTY_SLOTS", handleDirtySlots);
    return () =>
      window.removeEventListener("TELEMETRY_DIRTY_SLOTS", handleDirtySlots);
  }, [renderBufferSabs]);

  // Handle Global Worker Events
  useEffect(() => {
    const handleDensityUpdate = (e: any) => {
      if (clusterLayerRef.current) {
        clusterLayerRef.current.updateData(e.detail);
      }
    };
    window.addEventListener("TELEMETRY_H3_DENSITY", handleDensityUpdate);

    const handleH3Sync = (e: any) => {
      const { toSubscribe, toUnsubscribe, resolution } = e.detail;
      simulatorRef.current?.updateSubscriptions(
        toSubscribe,
        toUnsubscribe,
        resolution,
      );
    };
    window.addEventListener("SYNC_H3_SUBSCRIPTIONS", handleH3Sync);

    const handleFocusVehicle = (e: any) => {
      const { pickingId, lng, lat } = e.detail;
      if (instancedLayerRef.current) {
        instancedLayerRef.current.setFocusedVehicle(pickingId);
      }
      if (pointCloudLayerRef.current) {
        pointCloudLayerRef.current.setFocusedVehicle(pickingId);
      }
      if (mapRef.current && !(mapRef.current as any)._destroyed) {
        mapRef.current.flyTo({ center: [lng, lat], zoom: 16, duration: 2000 });
      }
    };

    window.addEventListener("VEHICLE_FOCUS_REQUEST", handleFocusVehicle);

    const handleZoom = () => {
      if (!mapRef.current) return;
      const zoom = mapRef.current.getZoom();

      // LOD 0: Heatmap/Cluster (Zoom 0-8)
      mapRef.current.setLayoutProperty(
        "cluster-layer",
        "visibility",
        zoom <= 8 ? "visible" : "none",
      );

      // LOD 1: PointCloud (Zoom 8-12)
      const pcLayer = mapRef.current.getLayer("point-cloud-layer");
      if (pcLayer) {
        mapRef.current.setLayoutProperty(
          "point-cloud-layer",
          "visibility",
          zoom > 8 && zoom <= 12 ? "visible" : "none",
        );
      }

      // LOD 2: High-Fidelity Icons (Zoom 12+)
      const instanceLayer = mapRef.current.getLayer("instanced-vehicles");
      if (instanceLayer) {
        mapRef.current.setLayoutProperty(
          "instanced-vehicles",
          "visibility",
          zoom > 12 ? "visible" : "none",
        );
      }
    };

    if (mapRef.current) {
      mapRef.current.on("zoom", handleZoom);
      handleZoom();
    }

    return () => {
      window.removeEventListener("TELEMETRY_H3_DENSITY", handleDensityUpdate);
      window.removeEventListener("SYNC_H3_SUBSCRIPTIONS", handleH3Sync);
      if (mapRef.current) {
        mapRef.current.off("zoom", handleZoom);
      }
    };
  }, [mapRef.current]);

  // Sync viewport to all workers and other tabs

  // Sync viewport to all workers and other tabs
  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    const syncViewport = () => {
      const bounds = map.getBounds();
      const viewportData = {
        minLng: bounds.getWest(),
        maxLng: bounds.getEast(),
        minLat: bounds.getSouth(),
        maxLat: bounds.getNorth(),
      };

      updateViewport(viewportData);

      if (isSyncEnabled) {
        broadcastViewport(viewportData);
      }
    };

    map.on("moveend", syncViewport);
    syncViewport(); // Initial sync

    return () => {
      map.off("moveend", syncViewport);
    };
  }, [updateViewport, isSyncEnabled, broadcastViewport]);

  // Command Palette State
  const [isSearchOpen, setIsSearchOpen] = React.useState(false);
  const [searchQuery, setSearchQuery] = React.useState("");
  const [searchResults, setSearchResults] = React.useState<any[]>([]);

  useEffect(() => {
    const handleSearchResults = (e: any) => {
      setSearchResults(e.detail);
    };
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setIsSearchOpen((prev) => !prev);
      }
      if (e.key === "Escape") setIsSearchOpen(false);
    };

    window.addEventListener("VEHICLE_SEARCH_RESULTS", handleSearchResults);
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("VEHICLE_SEARCH_RESULTS", handleSearchResults);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const searchDebounced = useCallback(
    debounce((query: string, filters: any) => {
      searchVehicle(query, filters);
    }, 300),
    [searchVehicle],
  );

  useEffect(() => {
    searchDebounced(searchQuery, activeFilters);
  }, [searchQuery, activeFilters, searchDebounced]);

  const onSearchQueryChange = (q: string) => {
    setSearchQuery(q);
  };

  const jumpToVehicle = (vehicle: any) => {
    setSelectedVehicle(vehicle);
    if (mapRef.current && vehicle.lng && vehicle.lat) {
      mapRef.current.flyTo({ center: [vehicle.lng, vehicle.lat], zoom: 14 });

      const pickingId = vehicle.index ?? -1;
      if (instancedLayerRef.current)
        instancedLayerRef.current.setFocusedVehicle(pickingId);
      if (pointCloudLayerRef.current)
        pointCloudLayerRef.current.setFocusedVehicle(pickingId);

      // Show mock planned route
      const mockRoute: [number, number][] = [
        [vehicle.lng, vehicle.lat],
        [vehicle.lng + 0.05, vehicle.lat + 0.05],
        [vehicle.lng + 0.1, vehicle.lat + 0.08],
      ];
      updatePlannedRoute(mapRef.current, mockRoute);
      requestBreadcrumbs(vehicle.id);
    }
    setIsSearchOpen(false);
  };

  // Keyboard Shortcuts (Cmd+K)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        document.getElementById("fleet-search")?.focus();
        setIsSearchOpen(true);
      }
      if (e.key === "Escape") {
        setIsSearchOpen(false);
        document.getElementById("fleet-search")?.blur();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []); // Mount only for shortcuts

  // EXPOSE HYPERSCALE CERTIFICATION CONTROLLER (Stable Mount)
  useEffect(() => {
    (window as any).__CERTIFICATION_CONTROLLER__ = {
      startStorm: (tps: number) => {
        startSimulation();
        simulatorRef.current?.setTps(tps);
      },
      triggerBurst: (count: number) => {
        simulatorRef.current?.triggerBurst(count);
      },
      crashWorker: (type: "telemetry" | "operational") => {
        window.dispatchEvent(
          new CustomEvent("FORCE_CRASH_WORKER", { detail: type }),
        );
      },
      setTileOutage: (active: boolean) => {
        simulatorRef.current?.setLatency(active ? 10000 : 0);
      },
      getMetrics: () => metricsRef.current,
      getWorkerHealth: () => monitorRef.current?.getStatus(),
    };

    return () => {
      delete (window as any).__CERTIFICATION_CONTROLLER__;
    };
  }, [startSimulation]); // startSimulation is memoized

  console.log('[ControlTower] Rendering JSX, mapRef:', !!mapRef.current);
  
  return (
    <div className="relative w-full h-screen bg-slate-950 flex overflow-hidden">
      {/* Search Overlay (Cmd+K) */}
      {isSearchOpen && (
        <div className="absolute inset-0 z-[100] bg-slate-950/60 backdrop-blur-sm flex items-start justify-center pt-24">
          <div className="bg-slate-900 w-full max-w-xl border border-slate-700 rounded-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="p-4 border-b border-slate-800 flex items-center gap-3">
              <Activity className="text-primary w-5 h-5" />
              <input
                autoFocus
                className="bg-transparent border-none outline-none text-white w-full text-lg placeholder:text-slate-600"
                placeholder="Search fleet (Vehicle ID, Type, Status)..."
                value={searchQuery}
                onChange={(e) => onSearchQueryChange(e.target.value)}
              />
            </div>
            <div className="max-h-[60vh] overflow-y-auto p-2">
              {searchResults.length > 0 ? (
                searchResults.map((v) => (
                  <div
                    key={v.id}
                    onClick={() => jumpToVehicle(v)}
                    className="p-3 rounded-lg hover:bg-slate-800 cursor-pointer flex justify-between items-center group transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <Truck className="text-slate-500 group-hover:text-primary w-4 h-4" />
                      <span className="text-slate-200 font-medium">{v.id}</span>
                    </div>
                    <span className="text-[10px] text-slate-500 uppercase tracking-widest">
                      {v.status || "active"}
                    </span>
                  </div>
                ))
              ) : searchQuery ? (
                <div className="p-8 text-center text-slate-500">
                  No vehicles matching "{searchQuery}"
                </div>
              ) : (
                <div className="p-8 text-center text-slate-500">
                  Type to search 1M+ vehicles...
                </div>
              )}
            </div>
            <div className="p-3 bg-slate-950/50 border-t border-slate-800 text-[10px] text-slate-600 flex justify-between">
              <span>ESC to close</span>
              <span>Worker-indexed fuzzy search</span>
            </div>
          </div>
        </div>
      )}

      {/* Sidebar Toolrail */}
      <div className="w-16 h-full bg-slate-900 border-r border-slate-800 flex flex-col items-center py-6 gap-6 z-20">
        <div className="p-2 bg-primary/20 rounded-lg">
          <Activity className="text-primary h-6 w-6" />
        </div>
        <Truck className="text-slate-400 h-6 w-6 cursor-pointer hover:text-white transition-colors" />
        <Navigation className="text-slate-400 h-6 w-6 cursor-pointer hover:text-white transition-colors" />
        <ShieldAlert className="text-red-500 h-6 w-6 cursor-pointer animate-pulse" />

        {/* Filter Panel (Simplified for Sidebar) */}
        <div className="mt-auto mb-6 flex flex-col items-center gap-4 border-t border-slate-800 pt-6">
          <div className="group relative">
            <Filter
              className={`h-5 w-5 cursor-pointer transition-colors ${Object.keys(activeFilters).length > 0 ? "text-primary" : "text-slate-500 hover:text-white"}`}
            />
            {/* Tooltip-style Simple Filter UI */}
            <div className="absolute left-12 bottom-0 w-48 bg-slate-900 border border-slate-700 rounded-xl shadow-2xl p-4 opacity-0 group-hover:opacity-100 pointer-events-none group-hover:pointer-events-auto transition-all scale-95 group-hover:scale-100 origin-left z-[110]">
              <h4 className="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-3">
                Segment Fleet
              </h4>
              <div className="space-y-3">
                <select
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg py-1.5 px-2 text-xs text-slate-300 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  value={activeFilters.org || ""}
                  onChange={(e) =>
                    setActiveFilters({
                      ...activeFilters,
                      org: e.target.value || undefined,
                    })
                  }
                >
                  <option value="">All Regions</option>
                  {filterMetadata.orgs.map((org) => (
                    <option key={org} value={org}>
                      {org}
                    </option>
                  ))}
                </select>
                <select
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg py-1.5 px-2 text-xs text-slate-300 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  value={activeFilters.status || ""}
                  onChange={(e) =>
                    setActiveFilters({
                      ...activeFilters,
                      status: e.target.value || undefined,
                    })
                  }
                >
                  <option value="">All Statuses</option>
                  {filterMetadata.statuses.map((status) => (
                    <option key={status} value={status}>
                      {status}
                    </option>
                  ))}
                </select>
                <button
                  onClick={() => setActiveFilters({})}
                  className="w-full text-[10px] text-slate-500 hover:text-white uppercase tracking-tighter text-left"
                >
                  Reset Filters
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Workspace */}
      <div className="flex-1 relative w-full h-full">
        <MapErrorBoundary>
          <MapEngine onMapLoad={handleMapLoad} />
        </MapErrorBoundary>

        {/* Top KPI Cluster */}
        <div className="absolute top-6 left-6 flex gap-4 pointer-events-none">
          <div className="bg-slate-900/80 backdrop-blur-md border border-slate-700/50 p-4 rounded-xl shadow-2xl pointer-events-auto min-w-[120px]">
            <p className="text-[10px] text-slate-500 uppercase font-black tracking-widest mb-1">
              Active Fleet
            </p>
            <p className="text-2xl font-black text-white">
              {metrics.activeEntities?.toLocaleString() || "0"}
            </p>
          </div>

          <div className="bg-slate-900/80 backdrop-blur-md border border-slate-700/50 p-4 rounded-xl shadow-2xl pointer-events-auto flex flex-col justify-center">
            <button
              onClick={isSimulating ? stopSimulation : startSimulation}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-wider transition-all active:scale-95 ${
                isSimulating
                  ? "bg-red-500 shadow-[0_0_15px_rgba(239,68,68,0.4)] text-white"
                  : "bg-primary shadow-[0_0_15px_rgba(37,99,235,0.4)] text-white hover:bg-primary-dark"
              }`}
            >
              {isSimulating ? (
                <>
                  <Square className="h-3 w-3 fill-current" /> Stop Sim
                </>
              ) : (
                <>
                  <Play className="h-3 w-3 fill-current" /> Start Sim
                </>
              )}
            </button>
          </div>
        </div>

        {/* System Integrity Panel */}
        <div className="absolute bottom-12 right-6 z-50 p-6 rounded-2xl bg-slate-900/80 backdrop-blur-xl border border-white/10 shadow-2xl space-y-4 w-64">
          <h3 className="text-xs font-bold tracking-[0.2em] text-blue-400 uppercase">
            System Integrity
          </h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-[10px] text-slate-400 uppercase tracking-tighter">
                GPU VRAM
              </p>
              <p className="text-2xl font-mono text-slate-100">
                {Math.round(metrics.vramMb || 0)}
                <span className="text-xs ml-0.5">MB</span>
              </p>
            </div>
            <div>
              <p className="text-[10px] text-slate-400 uppercase tracking-tighter">
                Active H3
              </p>
              <p className="text-2xl font-mono text-slate-100">
                {metrics.activeH3}
              </p>
            </div>
            <div>
              <p className="text-[10px] text-slate-400 uppercase tracking-tighter">
                Render FPS
              </p>
              <p
                className={`text-2xl font-mono ${metrics.fps > 55 ? "text-emerald-400" : "text-amber-400"}`}
              >
                {metrics.fps}
              </p>
            </div>
            <div>
              <p className="text-[10px] text-slate-400 uppercase tracking-tighter">
                E2E Lag
              </p>
              <p
                className={`text-2xl font-mono ${metrics.telemetryLag < 200 ? "text-emerald-400" : "text-rose-400"}`}
              >
                {Math.round(metrics.telemetryLag)}
                <span className="text-xs ml-0.5">ms</span>
              </p>
            </div>
          </div>
          <div className="pt-2 border-t border-white/5">
            <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
              <div
                className={`h-full transition-all duration-300 ${metrics.telemetryLag > 300 ? "bg-rose-500" : "bg-blue-500"}`}
                style={{
                  width: `${Math.min(100, (metrics.telemetryLag / 500) * 100)}%`,
                }}
              />
            </div>
          </div>
        </div>

        {/* Worker Health HUD */}
        <div className="absolute top-6 right-96 z-50 flex gap-2">
          {workerHealth.map((worker) => (
            <div
              key={worker.type}
              className={`px-3 py-1.5 rounded-lg border backdrop-blur-md flex items-center gap-2 transition-all duration-500 animate-entrance ${
                worker.isHealthy
                  ? "bg-slate-900/60 border-emerald-500/20 text-emerald-400"
                  : "bg-rose-500/20 border-rose-500/50 text-rose-500 animate-pulse"
              }`}
            >
              <div
                className={`w-1.5 h-1.5 rounded-full ${worker.isHealthy ? "bg-emerald-500" : "bg-rose-500"}`}
              />
              <span className="text-[9px] font-black uppercase tracking-widest">
                {worker.type}
              </span>
              <span className="text-[10px] font-mono opacity-50">
                {Math.round(worker.latency)}ms
              </span>
            </div>
          ))}
        </div>

        {/* Vehicle Details Panel (LOD 2 Focus) */}
        {selectedVehicle && (
          <div className="absolute left-6 bottom-12 z-50 w-80 bg-slate-900/90 backdrop-blur-2xl border border-white/10 p-6 rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] animate-in slide-in-from-left-5 duration-300">
            <div className="flex justify-between items-start mb-4">
              <div>
                <h2 className="text-white font-black text-xl tracking-tight">
                  {selectedVehicle.id}
                </h2>
                <span className="text-[10px] text-emerald-400 font-bold uppercase tracking-widest px-2 py-0.5 bg-emerald-500/10 rounded-full">
                  Active Operational
                </span>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setShowBreadcrumbs(!showBreadcrumbs)}
                  className={`p-1.5 rounded-lg border transition-all ${
                    showBreadcrumbs
                      ? "bg-emerald-500/20 border-emerald-500/50 text-emerald-400"
                      : "bg-slate-800 border-slate-700 text-slate-500"
                  }`}
                  title="Toggle Breadcrumbs"
                >
                  <Activity className="w-4 h-4" />
                </button>
                <button
                  onClick={() => {
                    setSelectedVehicle(null);
                    if (mapRef.current) clearRoutes(mapRef.current);
                  }}
                  className="p-1.5 bg-slate-800 border border-slate-700 rounded-lg text-slate-500 hover:text-white transition-colors"
                  title="Clear Selection"
                >
                  <Square className="w-4 h-4" />
                </button>
              </div>
            </div>

            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-white/5 p-3 rounded-xl border border-white/5">
                  <p className="text-[9px] text-slate-500 uppercase font-black mb-1">
                    Current Speed
                  </p>
                  <p className="text-lg font-mono text-white tracking-tighter">
                    {Math.round(selectedVehicle.speed || 0)}{" "}
                    <span className="text-[10px] text-slate-400">KM/H</span>
                  </p>
                </div>
                <div className="bg-white/5 p-3 rounded-xl border border-white/5">
                  <p className="text-[9px] text-slate-500 uppercase font-black mb-1">
                    H3 Cell (Res 7)
                  </p>
                  <p className="text-sm font-mono text-white truncate">
                    872830828ffffff
                  </p>
                </div>
              </div>

              <div className="bg-primary/10 border border-primary/20 p-3 rounded-xl flex items-center gap-3">
                <Navigation className="text-primary w-5 h-5" />
                <div>
                  <p className="text-[10px] text-primary font-bold uppercase tracking-wider">
                    Navigating To
                  </p>
                  <p className="text-xs text-white">
                    Central Logistics Hub - Sector 4
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Bottom Rail / Timeline Area */}
        <div className="absolute bottom-12 left-1/2 -translate-x-1/2 z-50 w-[600px] flex flex-col items-center gap-4">
          <TimeScrubber
            isLive={isLivePlayback}
            onLiveToggle={handleLiveToggle}
            onTimeChange={handleTimeChange}
            startTime={replayRange.start}
            endTime={replayRange.end}
          />

          {/* Global Alert Rail (Moved up slightly if needed, or kept) */}
          <div className="w-[500px] h-6 bg-rose-500/10 backdrop-blur-md border border-rose-500/20 rounded-full flex items-center px-4 gap-4 overflow-hidden shadow-2xl">
            <div className="flex items-center gap-2 shrink-0">
              <div className="w-2 h-2 bg-rose-500 rounded-full animate-pulse" />
              <span className="text-[9px] font-black text-rose-500 uppercase tracking-widest whitespace-nowrap">
                Operational Alerts:
              </span>
            </div>
            <div className="flex gap-8 animate-marquee whitespace-nowrap">
              <span className="text-[9px] text-rose-100 font-bold">
                [SIM-042] OVERSPEED DETECTED (112 km/h)
              </span>
              <span className="text-[9px] text-rose-100 font-bold">
                [GEO-92] GEOFENCE EXIT: MUMBAY_WAREHOUSE
              </span>
              <span className="text-[9px] text-rose-100 font-bold">
                [SIM-128] BATTERY CRITICAL: 4%
              </span>
            </div>
          </div>
        </div>

        {/* Quick Search */}
        <div className="absolute top-6 right-6 w-80 z-50">
          <div className="bg-slate-900/80 backdrop-blur-md border border-slate-700/50 p-2 rounded-xl shadow-2xl">
            <input
              id="fleet-search"
              type="text"
              className="w-full bg-slate-800/50 border border-slate-700 rounded-lg px-4 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
              placeholder="Search 1M vehicles... (Cmd+K)"
              autoComplete="off"
              onChange={(e) => searchVehicle(e.target.value)}
              onFocus={() => setIsSearchOpen(true)}
            />
          </div>

          {/* Search Results Overlay */}
          {isSearchOpen && searchResults.length > 0 && (
            <div className="mt-2 bg-slate-900/95 backdrop-blur-xl border border-white/10 rounded-xl shadow-2xl overflow-hidden max-h-[300px] overflow-y-auto custom-scrollbar animate-in fade-in zoom-in-95 duration-200">
              {searchResults.map((v: any) => (
                <div
                  key={v.id}
                  onClick={() => jumpToVehicle(v)}
                  className="px-4 py-3 hover:bg-white/5 cursor-pointer border-b border-white/5 last:border-0 transition-colors group flex justify-between items-center"
                >
                  <div>
                    <p className="text-sm font-bold text-white group-hover:text-primary transition-colors">
                      {v.id}
                    </p>
                    <p className="text-[10px] text-slate-500 uppercase tracking-tighter">
                      {v.type || "Commercial Fleet"}
                    </p>
                  </div>
                  <Navigation className="w-3 h-3 text-slate-600 group-hover:text-primary opacity-0 group-hover:opacity-100 transition-all" />
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Bottom Rail */}
        <div className="absolute bottom-0 inset-x-0 h-8 bg-slate-900/90 border-t border-slate-800 backdrop-blur flex items-center justify-between px-6 text-[10px] font-mono text-slate-500 z-10">
          <div className="flex gap-4">
            <span className="flex items-center gap-1.5">
              <span
                className={`h-1.5 w-1.5 rounded-full ${isSimulating ? "bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.8)]" : "bg-slate-600"}`}
              />
              WS: {isSimulating ? "SIMULATING" : "CONNECTED"}
            </span>
            <span>TELEMETRY: {isSimulating ? "100k/sec" : "IDLE"}</span>
          </div>
          <div className="flex gap-4 uppercase">
            <span>MAPLIBRE V5.x</span>
            <span>HYPERSCALE MODE</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ControlTower;
