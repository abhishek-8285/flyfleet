interface ObserverProps {
  metrics: {
    fps: number;
    dropCount: number;
    activeH3: number;
    telemetryLag: number;
  };
}

export const ObserverPanel: React.FC<ObserverProps> = ({ metrics }) => {
  return (
    <div className="absolute top-4 right-4 z-50 bg-slate-900/80 backdrop-blur-md border border-slate-700/50 p-3 rounded-lg text-xs font-mono text-slate-300 w-48 shadow-2xl">
      <div className="flex justify-between items-center mb-2 border-b border-slate-700 pb-1">
        <span className="text-emerald-400 font-bold uppercase tracking-wider">
          Observer
        </span>
        <div className="flex gap-1">
          <div
            className={`w-1.5 h-1.5 rounded-full ${metrics.dropCount > 0 ? "bg-red-500" : "bg-emerald-500"} animate-pulse`}
          />
        </div>
      </div>

      <div className="space-y-1.5">
        <div className="flex justify-between">
          <span>Display:</span>
          <span
            className={metrics.fps < 50 ? "text-amber-400" : "text-emerald-400"}
          >
            {metrics.fps} FPS
          </span>
        </div>
        <div className="flex justify-between">
          <span>Dropped:</span>
          <span
            className={
              metrics.dropCount > 0 ? "text-red-400" : "text-emerald-400"
            }
          >
            {metrics.dropCount}
          </span>
        </div>
        <div className="flex justify-between">
          <span>H3 Tiles:</span>
          <span className="text-sky-400">{metrics.activeH3}</span>
        </div>
        <div className="flex justify-between text-[10px] text-slate-500 pt-1 border-t border-slate-800">
          <span>Lat: {metrics.telemetryLag}ms</span>
          <span>Buffer: ACTIVE</span>
        </div>
      </div>
    </div>
  );
};
